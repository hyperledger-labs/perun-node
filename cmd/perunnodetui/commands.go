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
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

const commandHelp = "Commands: open <peer> <own> <peer's>, send/req <SNo> <amount>, acc/rej <SNo>, close <SNo>"

const (
	cmdOpen  = "open"
	cmdSend  = "send"
	cmdReq   = "req"
	cmdAcc   = "acc"
	cmdRej   = "rej"
	cmdClose = "close"
)

type commandHandlerFn func([]string) error

var commandHandlers = map[string]commandHandlerFn{
	cmdOpen:  openCmdHandler,
	cmdSend:  sendCmdHandler,
	cmdReq:   reqCmdHandler,
	cmdAcc:   accCmdHandler,
	cmdRej:   rejCmdHandler,
	cmdClose: closeCmdHandler,
}

func handleCmd(input string) error {
	words := strings.Split(input, " ")
	if len(words) < 2 {
		return errors.New("command should have at least 2 words")
	}
	command := words[0]
	args := words[1:]

	_, ok := commandHandlers[command]
	if !ok {
		return errors.New("invalid command: " + command)
	}

	go func() {
		err := commandHandlers[command](args)
		if err != nil {
			logError(errors.WithMessage(err, "executing command"))
		}
	}()
	return nil
}

// Usage: open <peer> <our bal> <peer bal>.
func openCmdHandler(args []string) error {
	// open <peer-name> <ours> <peers>
	countReqArgs := 3
	if len(args) != countReqArgs {
		return errors.New("should have only 3 args")
	}
	peer, ours, theirs := args[0], args[1], args[2]

	fn := newOutgoingChannelProposalFn(peer, ours, theirs)
	p := newOutgoingChannel(peer, ours, theirs, fn)
	R.putAtIndex(p.sNo, p)
	p.propose()
	return nil
}

// Usage: acc <S.No>.
func accCmdHandler(args []string) error {
	return respondCmd(args, true)
}

// Usage: rej <S.No>.
func rejCmdHandler(args []string) error {
	return respondCmd(args, false)
}

// not directly exposed, called by acc & rej commands.
// Usage: respond <S.No> true/false.
func respondCmd(args []string, accept bool) error {
	countReqArgs := 1
	if len(args) != countReqArgs {
		return errors.New("should have only 3 args")
	}
	index, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("first arg should a S.No")
	}

	p := R.getByIndex(index)
	if p == nil {
		return errors.New("first arg should a S.No")
	}

	p.respond(accept)
	return nil
}

// Usage: send <S.No> <amount>.
func sendCmdHandler(args []string) error {
	return updateCmdHandler(args, true)
}

// Usage: req <S.No> <amount>.
func reqCmdHandler(args []string) error {
	return updateCmdHandler(args, false)
}

// not directly exposed, called by send & req commands.
// Usage: respond <S.no> <amount> <isPayeePeer>.
func updateCmdHandler(args []string, isPayeePeer bool) error {
	countReqArgs := 2
	if len(args) != countReqArgs {
		return errors.New("should have only 2 args")
	}
	index, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("first arg should a S.No")
	}

	p := R.getByIndex(index)
	if p == nil {
		return errors.New("first arg should a S.No")
	}

	amount := args[1]
	p.update(amount, isPayeePeer)
	return nil
}

// Usage: close <S.No>.
func closeCmdHandler(args []string) error {
	countReqArgs := 1
	if len(args) != countReqArgs {
		return errors.New("should have only 1 args")
	}
	index, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("first arg should a S.No")
	}

	p := R.getByIndex(index)
	if p == nil {
		return errors.New("first arg should a S.No")
	}

	p.closeCh()
	return nil
}

// findPeerIndex assumes, parts has two entries and one of it is "self" or
// perun.OwnAlias.
func findPeerIndex(parts []string) int {
	if len(parts) != 2 {
		return 0
	}
	for idx := range parts {
		if parts[idx] != perun.OwnAlias {
			return idx
		}
	}
	return 0
}

func grpcPayChInfotoBalInfo(grpcPayChInfo *pb.PayChInfo) balInfo {
	b := grpcBalInfotoBalInfo(grpcPayChInfo.BalInfo)
	b.version = grpcPayChInfo.Version
	return b
}

func grpcBalInfotoBalInfo(grpcPayChInfo *pb.BalInfo) balInfo {
	peerIdx := findPeerIndex(grpcPayChInfo.Parts)
	ourIdx := peerIdx ^ 1
	return balInfo{
		ours:   grpcPayChInfo.Bal[ourIdx],
		theirs: grpcPayChInfo.Bal[peerIdx],
	}
}
