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

	pchannel "perun.network/go-perun/channel"
	pclient "perun.network/go-perun/client"
	psync "perun.network/go-perun/pkg/sync"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/log"
)

const (
	open chLockState = "open"
)

type (
	channel struct {
		log.Logger

		id               string
		pchannel         *pclient.Channel
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
	}
)

// NewChannel sets up a channel object from the passed pchannel.
func newChannel(pch *pclient.Channel, currency string, parts []string, timeoutCfg timeoutConfig,
	challengeDurSecs uint64) *channel {
	ch := &channel{
		id:                 fmt.Sprintf("%x", pch.ID()),
		pchannel:           pch,
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

	if ch.lockState != open {
		ch.Error("Dropping update request as the channel is " + ch.lockState)
		return perun.ErrChNotOpen
	}

	ctx, cancel := context.WithTimeout(pctx, ch.timeoutCfg.chUpdate())
	defer cancel()
	err := ch.pchannel.UpdateBy(ctx, updater)
	if err != nil {
		ch.Error("Sending channel update:", err)
		return perun.GetAPIError(err)
	}
	prevChInfo := ch.getChInfo()
	ch.currState = ch.pchannel.State().Clone()
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

func (ch *channel) RespondChUpdate(ctx context.Context, updateID string, accpet bool) error {
	return nil
}

func (ch *channel) GetInfo() perun.ChannelInfo {
	ch.Debug("Received request: channel.GetInfo")
	ch.Lock()
	defer ch.Unlock()
	return ch.getChInfo()
}

// This function assumes that caller has already locked the channel.
func (ch *channel) getChInfo() perun.ChannelInfo {
	return perun.ChannelInfo{
		ChannelID: ch.id,
		Currency:  ch.currency,
		State:     ch.currState,
		Parts:     ch.parts,
	}
}

func (ch *channel) Close(ctx context.Context) (perun.ChannelInfo, error) {
	return perun.ChannelInfo{}, nil
}
