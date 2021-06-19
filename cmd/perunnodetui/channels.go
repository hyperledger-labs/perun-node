// Copyright (c) 2021 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/hyperledger-labs/perun-node
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mum4k/termdash/widgets/text"
	"github.com/pkg/errors"
)

type (
	// channel represents an entry in the table.
	channel struct {
		sync.Mutex

		phase   chPhase
		peer    string
		current balInfo

		status   responseStatus
		timeout  string
		proposed balInfo

		chID string

		updateID   string // used for responding to incoming updates.
		isIncoming bool

		sNo        int
		row        *text.Text
		proposerFn proposerFn
	}

	responseStatus int
	chPhase        int
	balInfo        struct {
		ours    string
		theirs  string
		version string
	}

	proposerFn func() (bool, string, balInfo, error)
)

const (
	noUpdate responseStatus = iota
	waitingForPeer
	waitingForUser
	expired
	responding
	userRejected
	peerRejected
	accepted
	errorStatus
)

func (r responseStatus) String() string {
	return map[responseStatus]string{
		0: "",
		1: "For Peer",
		2: "For User",
		3: "Expired",
		4: "Responding",
		5: "User Rej",
		6: "Peer Rej",
		7: "Accepted",
		8: "Error",
	}[r]
}

const (
	errorPhase chPhase = iota
	open
	transact
	register
	settle
	closed
)

func (p chPhase) String() string {
	return map[chPhase]string{
		0: "Error",
		1: "Open",
		2: "Transact",
		3: "Register",
		4: "Settle",
		5: "Closed",
	}[p]
}

func newOutgoingChannel(peer, ours, theirs string, fn proposerFn) *channel {
	i, row, ok := T.getRow()
	if !ok {
		logError("table full, no more channels can be added.")
		return nil
	}

	return &channel{
		phase:  open,
		status: waitingForPeer,
		peer:   peer,
		proposed: balInfo{
			ours:    ours,
			theirs:  theirs,
			version: "0",
		},
		isIncoming: false,
		sNo:        i,
		row:        row,
		proposerFn: fn,
	}
}

func (p *channel) propose() {
	p.refreshEntry()

	isAccepted, chID, currentBalInfo, err := p.proposerFn()
	if err != nil {
		logError(err.Error())
		return
	}

	p.Lock()
	defer p.Unlock()
	if !isAccepted {
		p.phase = errorPhase
		p.status = peerRejected
		p.refreshEntry()
		return
	}

	R.setChID(chID, p.sNo)
	p.toTransact(chID, currentBalInfo)
}

func newIncomingChannel(proposalID, peer, ours, theirs, timeout string) (*channel, error) {
	i, row, ok := T.getRow()
	if !ok {
		return nil, errors.New("table full, no more channels can be added")
	}

	return &channel{
		phase:   open,
		status:  waitingForUser,
		timeout: timeout,
		peer:    peer,
		proposed: balInfo{
			ours:    ours,
			theirs:  theirs,
			version: "0",
		},
		updateID:   proposalID,
		isIncoming: true,

		sNo: i,
		row: row,
	}, nil
}

func (p *channel) notifyIncomingChannel(timeout time.Duration) {
	p.refreshEntry()
	time.Sleep(timeout)
	p.Lock()
	defer p.Unlock()
	if p.status == waitingForUser && p.phase == open {
		p.phase = errorPhase
		p.status = expired
		p.refreshEntry()
	}
}

func (p *channel) respond(accept bool) {
	p.Lock()
	defer p.Unlock()

	if p.status != waitingForUser {
		logError("no response expected for the channel")
		return
	}
	switch p.phase {
	case open:
		p.respondToProposal(accept)
	case transact, register:
		p.respondToUpdate(accept)
	}
}

func (p *channel) respondToProposal(accept bool) {
	p.status = responding
	p.refreshEntry()

	chID, currentBalInfo, err := respondToProposal(p.updateID, accept)
	if err != nil {
		logError(errors.WithMessage(err, "responding to proposal"))
		return
	}
	if !accept {
		p.phase = errorPhase
		p.status = userRejected
		p.refreshEntry()
		return
	}
	R.setChID(chID, p.sNo)
	p.toTransact(chID, currentBalInfo)
}

func (p *channel) toTransact(chID string, current balInfo) {
	p.chID = chID
	p.current = current
	p.phase = transact
	p.clearUpdate()
	p.refreshEntry()

	err := subUpdates(chID)
	if err != nil {
		logError(errors.WithMessage(err, "subscribing to updates"))
	}
	logInfo("subscribed to updates")
}

func (p *channel) notifyClosingUpdate(updated balInfo, errorMsg string) {
	p.Lock()
	if errorMsg == "" {
		p.phase = closed
		if p.current.version != updated.version {
			updated.version += " F"
			p.current = updated
		}
	} else {
		p.status = errorStatus
		logErrorf("settling channel %d: %s", p.sNo, errorMsg)
	}

	p.clearUpdate()
	p.refreshEntry()
	p.Unlock()
}

func (p *channel) notifyNonClosingUpdate(updateID string, proposed balInfo, expiryUnix int64, isFinal bool) {
	p.Lock()
	p.updateID = updateID
	p.proposed = proposed
	expiry := time.Unix(expiryUnix, 0)
	p.timeout = expiry.Format("15:04:05")
	p.status = waitingForUser

	if isFinal {
		p.proposed.version += " F"
		p.phase = register
	}
	p.refreshEntry()
	p.Unlock()

	currUpdateID := p.updateID
	time.Sleep(time.Until(expiry))

	p.Lock()
	defer p.Unlock()
	if p.updateID == currUpdateID && p.status == waitingForUser && (p.phase == transact || p.phase == register) {
		p.status = expired
		p.refreshEntry()
	}
}

func (p *channel) update(amount string, isPayeePeer bool) {
	p.Lock()
	defer p.Unlock()
	payee := ""
	if isPayeePeer {
		payee = p.peer
	} else {
		payee = "self"
	}

	if !(p.phase == transact && p.status != waitingForPeer && p.status != waitingForUser && p.status != responding) {
		logError("Channel not read for updates")
		return
	}

	proposed, err := getProposedBalInfo(p.current, amount, isPayeePeer)
	if err != nil {
		logError(err)
		return
	}
	p.proposed = proposed
	p.status = waitingForPeer
	p.refreshEntry()

	isAccepted, updated, err := sendUpdate(p.chID, payee, amount)
	if err != nil {
		p.status = errorStatus
		p.refreshEntry()
		logError(err)
		return
	}

	if !isAccepted {
		p.status = peerRejected
		p.refreshEntry()
		return
	}

	p.status = accepted
	p.current = updated
	p.refreshEntry()
}

func (p *channel) closeCh() {
	p.Lock()
	defer p.Unlock()

	if !(p.phase == transact && p.status != waitingForPeer && p.status != waitingForUser && p.status != responding) {
		logError("cannot close channel: not in transact phase or has pending updates")
		return
	}

	p.phase = register
	p.clearUpdate()
	p.refreshEntry()

	updated, err := closeCh(p.chID)
	if err != nil {
		p.phase = errorPhase
		p.refreshEntry()
		logError(errors.WithMessage(err, "closing channel"))
	}
	if p.current.version != updated.version {
		updated.version += " F"
	}
	p.current = updated
	p.phase = settle
	p.refreshEntry()
}

func (p *channel) respondToUpdate(accept bool) {
	p.status = responding
	p.refreshEntry()

	updatedBalInfo, err := respondToUpdate(p.chID, p.updateID, accept)
	if err != nil {
		logError(errors.WithMessage(err, "responding to update"))
		return
	}

	if !accept {
		p.status = userRejected
		p.refreshEntry()
		return
	}

	p.current = updatedBalInfo
	p.status = accepted
	if p.phase == register {
		p.current.version += " F"
	}
	p.refreshEntry()
}

func (p *channel) clearUpdate() {
	p.status = noUpdate
	p.timeout = ""
	p.proposed = balInfo{}
	p.proposerFn = nil
}

func (p *channel) refreshEntry() {
	// Trim the leading zeros in amount.
	trimZeros(&p.current.ours)
	trimZeros(&p.current.theirs)
	trimZeros(&p.proposed.ours)
	trimZeros(&p.proposed.theirs)

	p.row.Reset()
	var err error

	err = p.row.Write(fmt.Sprintf(T.bodyTemplate[0],
		p.sNo, p.phase, p.peer, p.current.ours, p.current.theirs, p.current.version))
	if err != nil {
		logErrorf("writing row #%d in the table", p.sNo)
	}

	if p.status == waitingForUser {
		err = p.row.Write(fmt.Sprintf(T.bodyTemplate[1], p.status), forUserWriteOpts)
		if err != nil {
			logErrorf("writing row #%d in the table", p.sNo)
		}
	} else {
		err = p.row.Write(fmt.Sprintf(T.bodyTemplate[1], p.status))
		if err != nil {
			logErrorf("writing row #%d in the table", p.sNo)
		}
	}

	err = p.row.Write(fmt.Sprintf(T.bodyTemplate[2],
		p.timeout, p.proposed.ours, p.proposed.theirs, p.proposed.version))
	if err != nil {
		logErrorf("writing proposal #%d to table", p.sNo)
	}
}

func trimZeros(s *string) {
	if strings.Contains(*s, ".") { // Trim only to the right of decimal point.
		*s = strings.TrimRight(*s, "0") // Trim all trailing zeros.
		*s = strings.TrimRight(*s, ".") // If it ends with ".", trim it.
	}
}

// This is a hacky function, that computes the proposed state when sending an update.
// Assumes, the version is an integer and will be incremented.
func getProposedBalInfo(current balInfo, amountStr string, isPayeePeer bool) (balInfo, error) {
	currOurs, _ := ethCurrency.Parse(current.ours)     // nolint: errcheck
	currTheirs, _ := ethCurrency.Parse(current.theirs) // nolint: errcheck
	// Values received from node should parse without errors.

	amount, err := ethCurrency.Parse(amountStr)
	if err != nil {
		return balInfo{}, errors.Wrap(err, "parsing amount")
	}

	proposed := balInfo{}

	if isPayeePeer {
		proposed.ours = truncateAtAmountMaxLen(ethCurrency.Print(currOurs.Sub(currOurs, amount)))
		proposed.theirs = truncateAtAmountMaxLen(ethCurrency.Print(currTheirs.Add(currTheirs, amount)))
	} else {
		proposed.theirs = truncateAtAmountMaxLen(ethCurrency.Print(currTheirs.Sub(currTheirs, amount)))
		proposed.ours = truncateAtAmountMaxLen(ethCurrency.Print(currOurs.Add(currOurs, amount)))
	}

	currVersion, err := strconv.Atoi(current.version)
	if err != nil {
		return balInfo{}, errors.Wrap(err, "parsing version as integer")
	}
	proposed.version = strconv.Itoa(currVersion + 1)
	return proposed, nil
}

func truncateAtAmountMaxLen(str string) string {
	if len(str) > amountMaxLength {
		return str[:amountMaxLength]
	}
	return str
}
