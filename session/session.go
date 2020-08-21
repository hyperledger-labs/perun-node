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

	pclient "perun.network/go-perun/client"

	"github.com/hyperledger-labs/perun-node"
)

type session struct {
}

func (s *session) ID() string {
	return ""
}

func (s *session) AddContact(peer perun.Peer) error {
	return nil
}

func (s *session) GetContact(alias string) (perun.Peer, error) {
	return perun.Peer{}, nil
}

func (s *session) OpenCh(
	pctx context.Context,
	peerAlias string,
	openingBals perun.BalInfo,
	app perun.App,
	challengeDurSecs uint64) (perun.ChannelInfo, error) {
	return perun.ChannelInfo{}, nil
}

func (s *session) HandleProposal(chProposal *pclient.ChannelProposal, responder *pclient.ProposalResponder) {
}

func (s *session) SubChProposals(notifier perun.ChProposalNotifier) error {
	return nil
}

func (s *session) UnsubChProposals() error {
	return nil
}

func (s *session) RespondChProposal(pctx context.Context, chProposalID string, accept bool) error {
	return nil
}

func (s *session) GetChInfos() []perun.ChannelInfo {
	return nil
}

func (s *session) GetCh(channelID string) (perun.ChannelAPI, error) {
	return nil, nil
}

func (s *session) HandleUpdate(chUpdate pclient.ChannelUpdate, responder *pclient.UpdateResponder) {
}

func (s *session) Close(force bool) error {
	return nil
}

func (s *session) HandleClose(chID string, err error) {
}

func (s *session) SubChCloses(notifier perun.ChCloseNotifier) error {
	return nil
}

func (s *session) UnsubChCloses() error {
	return nil
}
