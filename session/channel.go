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

func (c *channel) ID() string {
	return ""
}

func (c *channel) SendChUpdate(ctx context.Context, updater perun.StateUpdater) error {
	return nil
}

func (c *channel) SubChUpdates(notifier perun.ChUpdateNotifier) error {
	return nil
}

func (c *channel) UnsubChUpdates() error {
	return nil
}

func (c *channel) RespondChUpdate(ctx context.Context, updateID string, accpet bool) error {
	return nil
}

func (c *channel) GetInfo() perun.ChannelInfo {
	return perun.ChannelInfo{}
}

func (c *channel) Close(ctx context.Context) (perun.ChannelInfo, error) {
	return perun.ChannelInfo{}, nil
}
