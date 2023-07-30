// Copyright (c) 2022 - for information on the respective copyright owner
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

	pchannel "perun.network/go-perun/channel"
	pwatcher "perun.network/go-perun/watcher"

	"github.com/hyperledger-labs/perun-node"
)

// StartWatchingLedgerChannel provides a wrapper to call the
// StartWatchingLedgerChannel method on session.
func (s *Session) StartWatchingLedgerChannel(
	ctx context.Context,
	signedState pchannel.SignedState,
) (pwatcher.StatesPub, pwatcher.AdjudicatorSub, perun.APIError) {
	s.WithField("method", "StartWatchingLedgerChannel").Infof("\nReceived request with params %+v", signedState)
	statesPub, adjSub, err := s.watcher.StartWatchingLedgerChannel(ctx, signedState)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("StartWatchingLedgerChannel", apiErr)).Error(apiErr.Message())
		return nil, nil, apiErr
	}
	s.WithField("method", "StartWatchingLedgerChannel").Infof("Started watching for ledger channel %+v",
		signedState.State.ID)
	return statesPub, adjSub, nil
}

// StartWatchingSubChannel provides a wrapper to call the
// StartWatchingSubChannel method on session.
func (s *Session) StartWatchingSubChannel(
	ctx context.Context,
	parent pchannel.ID,
	signedState pchannel.SignedState,
) (pwatcher.StatesPub, pwatcher.AdjudicatorSub, perun.APIError) {
	s.WithField("method", "StartWatchingSubChannel").Infof("\nReceived request with params %+v, %+v", parent, signedState)
	statesPub, adjSub, err := s.watcher.StartWatchingSubChannel(ctx, parent, signedState)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("StartWatchingSubChannel", apiErr)).Error(apiErr.Message())
		return nil, nil, apiErr
	}
	s.WithField("method", "StartWatchingSubChannel").Infof("Started watching for sub channel %+v",
		signedState.State.ID)
	return statesPub, adjSub, nil
}

// StopWatching provides a wrapper to call the StopWatching method on session.
func (s *Session) StopWatching(ctx context.Context, chID pchannel.ID) perun.APIError {
	s.WithField("method", "StopWatching").Infof("\nReceived request with params %+v", chID)
	err := s.watcher.StopWatching(ctx, chID)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("UnsubChProposals", apiErr)).Error(apiErr.Message())
		return apiErr
	}
	s.WithField("method", "StopWatching").Infof("Stopped watching for channel %+v", chID)
	return nil
}
