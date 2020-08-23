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

	"github.com/hyperledger-labs/perun-node"
)

type channel struct{}

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
