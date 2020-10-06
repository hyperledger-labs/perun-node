// Copyright (c) 2020 - for information on the respective copyright owner
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

package session

import (
	"context"
	"fmt"
	"time"

	pchannel "perun.network/go-perun/channel"
	pclient "perun.network/go-perun/client"
	psync "perun.network/go-perun/pkg/sync"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/log"
)

const (
	open      chLockState = "open"
	finalized chLockState = "finalized"
	closed    chLockState = "closed"
)

type (
	channel struct {
		log.Logger

		id               string
		pch              *pclient.Channel
		lockState        chLockState
		currency         string
		parts            []string
		timeoutCfg       timeoutConfig
		challengeDurSecs uint64
		currState        *pchannel.State

		chUpdateNotifier   perun.ChUpdateNotifier
		chUpdateNotifCache []perun.ChUpdateNotif
		chUpdateResponders map[string]chUpdateResponderEntry

		psync.Mutex
	}

	chLockState string

	chUpdateResponderEntry struct {
		responder chUpdateResponder
		expiry    int64
	}

	//go:generate mockery --name ProposalResponder --output ../internal/mocks

	// ChUpdaterResponder represents the methods on channel update responder that will be used the perun node.
	chUpdateResponder interface {
		Accept(ctx context.Context) error
		Reject(ctx context.Context, reason string) error
	}
)

// newCh sets up a channel object from the passed pchannel.
func newCh(pch *pclient.Channel, currency string, parts []string, timeoutCfg timeoutConfig,
	challengeDurSecs uint64) *channel {
	ch := &channel{
		id:                 fmt.Sprintf("%x", pch.ID()),
		pch:                pch,
		lockState:          open,
		currState:          pch.State().Clone(),
		timeoutCfg:         timeoutCfg,
		challengeDurSecs:   challengeDurSecs,
		currency:           currency,
		parts:              parts,
		chUpdateResponders: make(map[string]chUpdateResponderEntry),
	}
	ch.Logger = log.NewLoggerWithField("channel-id", ch.id)
	return ch
}

func (ch *channel) ID() string {
	return ch.id
}

func (ch *channel) SendChUpdate(pctx context.Context, updater perun.StateUpdater) error {
	ch.Debug("Received request: channel.SendChUpdate")
	ch.Lock()
	defer ch.Unlock()

	if ch.lockState == finalized {
		ch.Error("Dropping update request as the channel is " + ch.lockState)
		return perun.ErrChFinalized
	} else if ch.lockState == closed {
		ch.Error("Dropping update request as the channel is " + ch.lockState)
		return perun.ErrChClosed
	}

	ctx, cancel := context.WithTimeout(pctx, ch.timeoutCfg.chUpdate())
	defer cancel()
	err := ch.pch.UpdateBy(ctx, ch.pch.Idx(), updater)
	if err != nil {
		ch.Error("Sending channel update:", err)
		return perun.GetAPIError(err)
	}
	prevChInfo := ch.getChInfo()
	ch.currState = ch.pch.State().Clone()
	ch.Debugf("State upated from %v to %v", prevChInfo, ch.getChInfo())
	return nil
}

func (ch *channel) SubChUpdates(notifier perun.ChUpdateNotifier) error {
	ch.Debug("Received request: channel.SubChUpdates")
	ch.Lock()
	defer ch.Unlock()

	if ch.chUpdateNotifier != nil {
		ch.Error(perun.ErrSubAlreadyExists)
		return perun.ErrSubAlreadyExists
	}
	ch.chUpdateNotifier = notifier

	// Send all cached notifications
	for i := len(ch.chUpdateNotifCache); i > 0; i-- {
		go ch.chUpdateNotifier(ch.chUpdateNotifCache[0])
		ch.chUpdateNotifCache = ch.chUpdateNotifCache[1:i]
	}
	return nil
}

func (ch *channel) UnsubChUpdates() error {
	ch.Debug("Received request: channel.UnsubChUpdates")
	ch.Lock()
	defer ch.Unlock()

	if ch.chUpdateNotifier == nil {
		ch.Error(perun.ErrNoActiveSub)
		return perun.ErrNoActiveSub
	}
	ch.chUpdateNotifier = nil
	return nil
}

func (ch *channel) RespondChUpdate(pctx context.Context, updateID string, accept bool) error {
	ch.Debug("Received request channel.RespondChUpdate")
	ch.Lock()
	defer ch.Unlock()

	entry, ok := ch.chUpdateResponders[updateID]
	if !ok {
		ch.Error(perun.ErrUnknownUpdateID, updateID)
		return perun.ErrUnknownUpdateID
	}
	delete(ch.chUpdateResponders, updateID)
	currTime := time.Now().UTC().Unix()
	if entry.expiry < currTime {
		ch.Error("timeout:", entry.expiry, "received response at:", currTime)
		return perun.ErrRespTimeoutExpired
	}

	switch accept {
	case true:
		ctx, cancel := context.WithTimeout(pctx, ch.timeoutCfg.respChUpdateAccept())
		defer cancel()
		err := entry.responder.Accept(ctx)
		if err != nil {
			ch.Logger.Error("Accepting channel update", err)
			return perun.GetAPIError(err)
		}
		ch.currState = ch.pch.State().Clone()

	case false:
		ctx, cancel := context.WithTimeout(pctx, ch.timeoutCfg.respChUpdateReject())
		defer cancel()
		err := entry.responder.Reject(ctx, "rejected by user")
		if err != nil {
			ch.Logger.Error("Rejecting channel update", err)
			return perun.GetAPIError(err)
		}
	}

	// TODO: (mano) Provide an option for user to config the node to close finalized channels automatically.
	// For now, it is upto the user to close a channel that has been set to finalized state.
	// if ch.lockState == finalized {
	// }
	return nil
}

func (ch *channel) GetChInfo() perun.ChInfo {
	ch.Debug("Received request: channel.GetChInfo")
	ch.Lock()
	defer ch.Unlock()
	return ch.getChInfo()
}

// This function assumes that caller has already locked the channel.
func (ch *channel) getChInfo() perun.ChInfo {
	return perun.ChInfo{
		ChID:     ch.id,
		Currency: ch.currency,
		State:    ch.currState,
		Parts:    ch.parts,
	}
}

func (ch *channel) Close(pctx context.Context) (perun.ChInfo, error) {
	ch.Debug("Received request channel.Close")
	ch.Lock()
	defer ch.Unlock()

	switch ch.lockState {
	case open:
		ch.lockState = closed
		// Try to finalize state, so that channel can be settled directly without waiting for challenge duration
		// to expire. If this fails, channel will still be settled but by registering the state on-chain
		// and waiting for challenge duration to expire.
		chFinalizer := func(state *pchannel.State) {
			state.IsFinal = true
		}
		upCtx, upCancel := context.WithTimeout(pctx, ch.timeoutCfg.chUpdate())
		defer upCancel()
		if err := ch.pch.UpdateBy(upCtx, ch.pch.Idx(), chFinalizer); err != nil {
			ch.Logger.Info("Error when trying to finalize state for closing:", err)
			ch.Logger.Info("Opting for non collaborative close")
		} else {
			ch.currState = ch.pch.State().Clone()
		}
		fallthrough

	case finalized:
		ch.lockState = closed
		clCtx, clCancel := context.WithTimeout(pctx, ch.timeoutCfg.closeCh(ch.challengeDurSecs))
		defer clCancel()
		err := ch.pch.Settle(clCtx)

		if cerr := ch.pch.Close(); err != nil {
			ch.Logger.Error("Settling channel", err)
			return perun.ChInfo{}, perun.GetAPIError(err)
		} else if cerr != nil {
			ch.Logger.Error("Closing channel", cerr)
		}
		return ch.getChInfo(), nil

	case closed:
		return ch.getChInfo(), perun.ErrChClosed
	}
	ch.Error("Program reached unknonwn state")
	return ch.getChInfo(), perun.ErrInternalServer
}
