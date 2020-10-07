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
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ppayment "perun.network/go-perun/apps/payment"
	pchannel "perun.network/go-perun/channel"
	pclient "perun.network/go-perun/client"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/app/payment"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/internal/mocks"
)

var (
	peerAlias               = "peer"
	parts                   = []string{perun.OwnAlias, peerAlias}
	version          uint64 = 1
	versionString    string = "1"
	challengeDurSecs uint64 = 10
	app                     = perun.App{
		Def:  ppayment.NewApp(),
		Data: pchannel.NoData(),
	}

	chID             = "channel1"
	proposalID       = "proposal1"
	updateID         = "update1"
	expiry     int64 = 1597946401

	balance    = []string{"1", "2"}
	allocation = pchannel.Allocation{
		Balances: [][]*big.Int{{big.NewInt(1e18), big.NewInt(2e18)}},
	}
	wantBalance = []string{"1.000000", "2.000000"}

	amountToSend      = "0.5"
	updatedAllocation = pchannel.Allocation{
		Balances: [][]*big.Int{{big.NewInt(0.5e18), big.NewInt(2.5e18)}},
	}
	wantUpdatedBalance = []string{"0.500000", "2.500000"}

	balInfo = perun.BalInfo{
		Currency: currency.ETH,
		Parts:    parts,
		Bal:      balance,
	}
	wantBalInfo = perun.BalInfo{
		Currency: currency.ETH,
		Parts:    parts,
		Bal:      wantBalance,
	}
	wantUpdatedBalInfo = perun.BalInfo{
		Currency: currency.ETH,
		Parts:    parts,
		Bal:      wantUpdatedBalance,
	}

	chInfo = perun.ChInfo{
		ChID:     chID,
		Currency: currency.ETH,
		State: &pchannel.State{
			App:        &ppayment.App{},
			Data:       pchannel.NoData(),
			Allocation: allocation,
			Version:    version,
		},
		Parts: []string{perun.OwnAlias, peerAlias},
	}
	updatedChInfo = perun.ChInfo{
		ChID:     chID,
		Currency: currency.ETH,
		State: &pchannel.State{
			App:        &ppayment.App{},
			Data:       pchannel.NoData(),
			Allocation: updatedAllocation,
			Version:    version,
		},
		Parts: []string{perun.OwnAlias, peerAlias},
	}

	chProposalNotif = perun.ChProposalNotif{
		ProposalID: proposalID,
		Currency:   currency.ETH,
		ChProposal: &pclient.LedgerChannelProposal{
			BaseChannelProposal: pclient.BaseChannelProposal{
				ChallengeDuration: challengeDurSecs,
				InitBals:          &allocation,
			},
		},
		Parts:  parts,
		Expiry: expiry,
	}
	chUpdateNotif = perun.ChUpdateNotif{
		UpdateID: updateID,
		Currency: currency.ETH,
		Update: &pclient.ChannelUpdate{
			State: &pchannel.State{
				Allocation: updatedAllocation,
				IsFinal:    true,
				Version:    version,
			},
		},
		Parts:  parts,
		Expiry: expiry,
	}
	chCloseNotif = perun.ChCloseNotif{
		ChID:     chID,
		Currency: currency.ETH,
		ChState: &pchannel.State{
			Allocation: updatedAllocation,
			Version:    version,
		},
		Parts: parts,
		Error: "",
	}
)

func Test_OpenPayCh(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("OpenCh", context.Background(), balInfo, app, challengeDurSecs).Return(chInfo, nil)

		gotPayChInfo, gotErr := payment.OpenPayCh(context.Background(), sessionAPI, balInfo, challengeDurSecs)
		require.NoError(t, gotErr)
		assert.Equal(t, wantBalInfo, gotPayChInfo.BalInfo)
		assert.Equal(t, versionString, gotPayChInfo.Version)
		assert.NotZero(t, gotPayChInfo.ChID)
	})

	t.Run("error", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("OpenCh", context.Background(), balInfo, app, challengeDurSecs).Return(
			perun.ChInfo{}, assert.AnError)

		_, gotErr := payment.OpenPayCh(context.Background(), sessionAPI, balInfo, challengeDurSecs)
		require.Error(t, gotErr)
	})
}

func Test_GetPayChs(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("GetChsInfo").Return([]perun.ChInfo{chInfo})

		gotPayChInfos := payment.GetPayChsInfo(sessionAPI)
		require.Len(t, gotPayChInfos, 1)
		assert.Equal(t, versionString, gotPayChInfos[0].Version)
		assert.Equal(t, wantBalInfo, gotPayChInfos[0].BalInfo)
		assert.NotZero(t, gotPayChInfos[0].ChID)
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
		assert.Equal(t, chProposalNotif.ProposalID, notif.ProposalID)
		assert.Equal(t, chProposalNotif.Currency, notif.Currency)
		assert.Equal(t, wantBalInfo, notif.OpeningBalInfo)
		assert.Equal(t, chProposalNotif.ChProposal.Proposal().ChallengeDuration, notif.ChallengeDurSecs)
		assert.Equal(t, chProposalNotif.Expiry, notif.Expiry)
	})
	t.Run("error", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("SubChProposals", mock.Anything).Return(assert.AnError)

		dummyNotifier := func(notif payment.PayChProposalNotif) {}
		gotErr := payment.SubPayChProposals(sessionAPI, dummyNotifier)
		assert.Error(t, gotErr)
	})
}

// nolint: dupl	// not duplicate of Test_UnsubPayChCloses.
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
	})
}

// nolint: dupl	// not duplicate of Test_RespondPayChUpdate.
func Test_RespondPayChProposal(t *testing.T) {
	proposalID := "proposal-id-1"
	t.Run("happy_accept", func(t *testing.T) {
		accept := true
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("RespondChProposal", context.Background(), proposalID, accept).Return(nil)

		gotErr := payment.RespondPayChProposal(context.Background(), sessionAPI, proposalID, accept)
		assert.NoError(t, gotErr)
	})
	t.Run("happy_reject", func(t *testing.T) {
		accept := false
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("RespondChProposal", context.Background(), proposalID, accept).Return(nil)

		gotErr := payment.RespondPayChProposal(context.Background(), sessionAPI, proposalID, accept)
		assert.NoError(t, gotErr)
	})
	t.Run("error", func(t *testing.T) {
		accept := true
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("RespondChProposal", context.Background(), proposalID, accept).Return(assert.AnError)

		gotErr := payment.RespondPayChProposal(context.Background(), sessionAPI, proposalID, accept)
		assert.Error(t, gotErr)
	})
}

func Test_SubPayChCloses(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		var notifier perun.ChCloseNotifier
		var notif payment.PayChCloseNotif
		dummyNotifier := func(gotNotif payment.PayChCloseNotif) {
			notif = gotNotif
		}
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("SubChCloses", mock.MatchedBy(func(gotNotifier perun.ChCloseNotifier) bool {
			notifier = gotNotifier
			return true
		})).Return(nil)

		gotErr := payment.SubPayChCloses(sessionAPI, dummyNotifier)
		require.NoError(t, gotErr)
		require.NotNil(t, notifier)
		notifier(chCloseNotif)
		require.NotZero(t, notif)
		assert.Equal(t, chCloseNotif.ChID, notif.ClosedPayChInfo.ChID)
		assert.Equal(t, wantUpdatedBalInfo, notif.ClosedPayChInfo.BalInfo)
		assert.Equal(t, versionString, notif.ClosedPayChInfo.Version)
	})
	t.Run("error", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("SubChCloses", mock.Anything).Return(assert.AnError)

		dummyNotifier := func(notif payment.PayChCloseNotif) {}
		gotErr := payment.SubPayChCloses(sessionAPI, dummyNotifier)
		assert.Error(t, gotErr)
	})
}

// nolint: dupl	// not duplicate of Test_UnsubPayChProposals.
func Test_UnsubPayChCloses(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("UnsubChCloses", mock.Anything).Return(nil)

		gotErr := payment.UnsubPayChCloses(sessionAPI)
		assert.NoError(t, gotErr)
	})
	t.Run("error", func(t *testing.T) {
		sessionAPI := &mocks.SessionAPI{}
		sessionAPI.On("UnsubChCloses", mock.Anything).Return(assert.AnError)

		gotErr := payment.UnsubPayChCloses(sessionAPI)
		assert.Error(t, gotErr)
	})
}
