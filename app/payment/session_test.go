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

package payment_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ppayment "perun.network/go-perun/apps/payment"
	pchannel "perun.network/go-perun/channel"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/app/payment"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/internal/mocks"
)

var (
	// Fixed channel data.
	peerAlias = "peer"
	parts     = []string{perun.OwnAlias, peerAlias}
	curr      = currency.ETH

	challengeDurSecs uint64 = 10
	app                     = perun.App{
		Def:  ppayment.NewApp(),
		Data: pchannel.NoData(),
	}

	chID             = "channel1"
	proposalID       = "proposal1"
	updateID         = "update1"
	expiry     int64 = 1597946401

	// Initial channel data.
	openingBalInfoInput = perun.BalInfo{
		Currency: curr,
		Parts:    parts,
		Bal:      []string{"1", "2"},
	}
	openingBalInfo = perun.BalInfo{
		Currency: curr,
		Parts:    parts,
		Bal:      []string{"1.000000", "2.000000"},
	}
	openedChInfo = perun.ChInfo{
		ChID:    chID,
		BalInfo: openingBalInfo,
		App:     app,
		Version: "0",
	}
	wantOpenedPayChInfo = payment.PayChInfo{
		ChID:    chID,
		BalInfo: openingBalInfo,
		Version: "0",
	}

	chProposalNotif = perun.ChProposalNotif{
		ProposalID:       proposalID,
		OpeningBalInfo:   openingBalInfo,
		App:              app,
		ChallengeDurSecs: challengeDurSecs,
		Expiry:           expiry,
	}
	wantPayChProposalNotif = payment.PayChProposalNotif{proposalID, openingBalInfo, challengeDurSecs, expiry}

	// Updated channel data.
	amountToSend   = "0.5"
	updatedBalInfo = perun.BalInfo{
		Currency: curr,
		Parts:    parts,
		Bal:      []string{"0.500000", "2.500000"},
	}
	updatedChInfo = perun.ChInfo{
		ChID:    chID,
		BalInfo: updatedBalInfo,
		App:     app,
		Version: "1",
	}
	wantUpdatedPayChInfo = payment.PayChInfo{
		ChID:    chID,
		BalInfo: updatedBalInfo,
		Version: "1",
	}

	chUpdateNotif = perun.ChUpdateNotif{
		UpdateID:       updateID,
		ProposedChInfo: updatedChInfo,
		Type:           perun.ChUpdateTypeOpen,
		Expiry:         expiry,
		Error:          "",
	}
	wantPayChUpdateNotif = payment.PayChUpdateNotif{
		UpdateID:          updateID,
		ProposedPayChInfo: wantUpdatedPayChInfo,
		Type:              perun.ChUpdateTypeOpen,
		Expiry:            expiry,
		Error:             "",
	}
)

func Test_OpenPayCh(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("OpenCh", context.Background(), openingBalInfoInput, app, challengeDurSecs).Return(openedChInfo, nil)

		gotPayChInfo, gotErr := payment.OpenPayCh(context.Background(), sessionAPI, openingBalInfoInput, challengeDurSecs)
		require.NoError(t, gotErr)
		assert.Equal(t, wantOpenedPayChInfo, gotPayChInfo)
	})

	t.Run("error", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("OpenCh", context.Background(), openingBalInfoInput, app, challengeDurSecs).Return(
			perun.ChInfo{}, assert.AnError)

		_, gotErr := payment.OpenPayCh(context.Background(), sessionAPI, openingBalInfoInput, challengeDurSecs)
		require.Error(t, gotErr)
		t.Log(gotErr)
	})
}

func Test_GetPayChs(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("GetChsInfo").Return([]perun.ChInfo{openedChInfo})

		gotPayChInfos := payment.GetPayChsInfo(sessionAPI)
		require.Len(t, gotPayChInfos, 1)
		assert.Equal(t, wantOpenedPayChInfo, gotPayChInfos[0])
	})
}

func Test_SubPayChProposals(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		var notifier perun.ChProposalNotifier
		var notif payment.PayChProposalNotif
		dummyNotifier := func(gotNotif payment.PayChProposalNotif) {
			notif = gotNotif
		}
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("SubChProposals", mock.MatchedBy(func(gotNotifier perun.ChProposalNotifier) bool {
			notifier = gotNotifier
			return true
		})).Return(nil)

		gotErr := payment.SubPayChProposals(sessionAPI, dummyNotifier)
		require.NoError(t, gotErr)
		require.NotNil(t, notifier)

		notifier(chProposalNotif)
		require.NotZero(t, notif)
		assert.Equal(t, wantPayChProposalNotif, notif)
	})
	t.Run("error", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("SubChProposals", mock.Anything).Return(assert.AnError)

		dummyNotifier := func(notif payment.PayChProposalNotif) {}
		gotErr := payment.SubPayChProposals(sessionAPI, dummyNotifier)
		assert.Error(t, gotErr)
		t.Log(gotErr)
	})
}

func Test_UnsubPayChProposals(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("UnsubChProposals", mock.Anything).Return(nil)

		gotErr := payment.UnsubPayChProposals(sessionAPI)
		assert.NoError(t, gotErr)
	})
	t.Run("error", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("UnsubChProposals", mock.Anything).Return(assert.AnError)

		gotErr := payment.UnsubPayChProposals(sessionAPI)
		assert.Error(t, gotErr)
		t.Log(gotErr)
	})
}

// nolint: dupl	// not duplicate of Test_RespondPayChUpdate.
func Test_RespondPayChProposal(t *testing.T) {
	proposalID := "proposal-id-1"
	t.Run("happy_accept", func(t *testing.T) {
		accept := true
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("RespondChProposal", context.Background(), proposalID, accept).Return(openedChInfo, nil)

		gotOpenedPayChInfo, gotErr := payment.RespondPayChProposal(context.Background(), sessionAPI, proposalID, accept)
		assert.NoError(t, gotErr)
		assert.Equal(t, wantOpenedPayChInfo, gotOpenedPayChInfo)
	})
	t.Run("happy_reject", func(t *testing.T) {
		accept := false
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("RespondChProposal", context.Background(), proposalID, accept).Return(perun.ChInfo{}, nil)

		_, gotErr := payment.RespondPayChProposal(context.Background(), sessionAPI, proposalID, accept)
		assert.NoError(t, gotErr)
	})
	t.Run("error", func(t *testing.T) {
		accept := true
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("RespondChProposal", context.Background(), proposalID, accept).Return(perun.ChInfo{}, assert.AnError)

		_, gotErr := payment.RespondPayChProposal(context.Background(), sessionAPI, proposalID, accept)
		assert.Error(t, gotErr)
		t.Log(gotErr)
	})
}

func Test_CloseSession(t *testing.T) {
	t.Run("happy_noForce", func(t *testing.T) {
		force := false
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("Close", force).Return(nil, nil)
	})
	t.Run("happy_Force", func(t *testing.T) {
		force := true
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("Close", force).Return(nil, nil)
	})
	t.Run("ErrorOpenChs_noForce", func(t *testing.T) {
		force := false
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("Close", force).Return([]perun.ChInfo{updatedChInfo}, nil)
	})
}
