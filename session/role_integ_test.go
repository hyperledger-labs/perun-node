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

//go:build integration
// +build integration

package session_test

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	pchannel "perun.network/go-perun/channel"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/session"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

// This test includes all methods on SessionAPI and ChAPI.
func Test_Integ_Role(t *testing.T) {
	// Deploy contracts.
	contracts := ethereumtest.SetupContractsT(t,
		ethereumtest.ChainURL, ethereumtest.ChainID, ethereumtest.OnChainTxTimeout, false)
	currencies := currency.NewRegistry()
	_, err := currencies.Register(currency.ETHSymbol, currency.ETHMaxDecimals)
	require.NoError(t, err)

	aliceAlias, bobAlias := "alice", "bob"

	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	aliceCfg := sessiontest.NewConfigT(t, prng)
	bobCfg := sessiontest.NewConfigT(t, prng)

	alice, err := session.New(aliceCfg, currencies, contracts)
	require.NoErrorf(t, err, "initializing alice session")
	t.Logf("alice session id: %s\n", alice.ID())
	t.Logf("alice database dir is: %s\n", aliceCfg.DatabaseDir)

	bob, err := session.New(bobCfg, currencies, contracts)
	require.NoErrorf(t, err, "initializing bob session")
	t.Logf("bob session id: %s\n", bob.ID())
	t.Logf("bob database dir is: %s\n", bobCfg.DatabaseDir)

	var alicePeerID, bobPeerID perun.PeerID
	passed := t.Run("GetPeerID", func(t *testing.T) {
		t.Run("happy", func(t *testing.T) {
			alicePeerID, err = alice.GetPeerID(perun.OwnAlias)
			require.NoErrorf(t, err, "alice reading own peer ID")
			alicePeerID.Alias = aliceAlias

			bobPeerID, err = bob.GetPeerID(perun.OwnAlias)
			require.NoErrorf(t, err, "bob reading own peer ID")
			bobPeerID.Alias = bobAlias
		})
		t.Run("missing", func(t *testing.T) {
			_, err = alice.GetPeerID("random alias")
			assert.Errorf(t, err, "alice reading random peer ID should error")
			t.Log(err)
		})
	})
	require.True(t, passed)

	passed = t.Run("AddPeerID", func(t *testing.T) {
		t.Run("happy", func(t *testing.T) {
			err = alice.AddPeerID(bobPeerID)
			require.NoErrorf(t, err, "alice adding bob's peer ID")

			err = bob.AddPeerID(alicePeerID)
			require.NoErrorf(t, err, "bob adding alice's peer ID")
		})
		t.Run("already_exists", func(t *testing.T) {
			// Try to add bob peer ID again
			err = alice.AddPeerID(bobPeerID)
			assert.Errorf(t, err, "alice adding bob's peer ID when it is already present should error")
			t.Log(err)
		})
	})
	require.True(t, passed)

	const challengeDurSecs uint64 = 10
	wg := &sync.WaitGroup{}
	ctx := context.Background()

	// Alice will propose two channels that will be accepted by bob.
	// 1. One of the channel will be used for send/accept channel update,
	//    send/reject channel update followed by collaborative close.
	// 2. The other channel will be used for non-collaborative close.
	passed = t.Run("OpenCh_Sub_Unsub_ChProposal_Respond_Accept", func(t *testing.T) {
		proposeAcceptCh := func() {
			// Propose Channel by alice.
			wg.Add(1)
			go func() {
				defer wg.Done()

				openingBalInfo := perun.BalInfo{
					Currencies: []string{currency.ETHSymbol},
					Parts:      []string{perun.OwnAlias, bobAlias},
					Bals:       [][]string{{"1", "2"}},
				}
				app := perun.App{
					Def:  pchannel.NoApp(),
					Data: pchannel.NoData(),
				}
				_, err := alice.OpenCh(ctx, openingBalInfo, app, challengeDurSecs)
				require.NoErrorf(t, err, "alice opening channel with bob")
			}()
			defer wg.Wait()

			// Accept channel by bob.
			bobChProposalNotif := make(chan perun.ChProposalNotif)
			bobChProposalNotifier := func(notif perun.ChProposalNotif) {
				bobChProposalNotif <- notif
			}
			err = bob.SubChProposals(bobChProposalNotifier)
			require.NoError(t, err, "bob subscribing to channel proposals")

			notif := <-bobChProposalNotif
			_, err = bob.RespondChProposal(ctx, notif.ProposalID, true)
			require.NoError(t, err, "bob accepting channel proposal")

			err = bob.UnsubChProposals()
			require.NoError(t, err, "bob unsubscribing from channel proposals")
		}

		proposeAcceptCh()
		proposeAcceptCh()
	})
	require.True(t, passed)

	passed = t.Run("OpenCh_Sub_Unsub_ChProposal_Respond_Reject", func(t *testing.T) {
		// Propose Channel by bob.
		wg.Add(1)
		go func() {
			defer wg.Done()

			openingBalInfo := perun.BalInfo{
				Currencies: []string{currency.ETHSymbol},
				Parts:      []string{aliceAlias, perun.OwnAlias},
				Bals:       [][]string{{"1", "2"}},
			}
			app := perun.App{
				Def:  pchannel.NoApp(),
				Data: pchannel.NoData(),
			}
			_, err := bob.OpenCh(ctx, openingBalInfo, app, challengeDurSecs)
			require.Error(t, err, "bob sending channel proposal should be rejected by alice")
			t.Log(err)
		}()
		defer wg.Wait()

		// Reject channel by alice.
		aliceChProposalNotif := make(chan perun.ChProposalNotif)
		aliceChProposalNotifier := func(notif perun.ChProposalNotif) {
			aliceChProposalNotif <- notif
		}
		err = alice.SubChProposals(aliceChProposalNotifier)
		require.NoError(t, err, "alice subscribing to channel proposals")

		notif := <-aliceChProposalNotif
		_, err = alice.RespondChProposal(ctx, notif.ProposalID, false)
		require.NoError(t, err, "alice rejecting channel proposal")

		err = alice.UnsubChProposals()
		require.NoError(t, err, "alice unsubscribing from channel proposals")
	})
	require.True(t, passed)

	aliceChs, bobChs := make([]perun.ChAPI, 2), make([]perun.ChAPI, 2)
	passed = t.Run("GetChsInfo_GetCh", func(t *testing.T) {
		aliceChInfos := alice.GetChsInfo()
		require.Lenf(t, aliceChInfos, 2, "alice session should have exactly two channels")
		bobChInfos := bob.GetChsInfo()
		require.Lenf(t, bobChInfos, 2, "bob session should have exactly two channels")

		aliceChs[0], err = alice.GetCh(aliceChInfos[0].ChID)
		require.NoError(t, err, "getting alice ChAPI instance")
		aliceChs[1], err = alice.GetCh(aliceChInfos[1].ChID)
		require.NoError(t, err, "getting alice ChAPI instance")

		bobChs[0], err = bob.GetCh(bobChInfos[0].ChID)
		require.NoError(t, err, "getting bob ChAPI instance")
		bobChs[1], err = bob.GetCh(bobChInfos[1].ChID)
		require.NoError(t, err, "getting bob ChAPI instance")
	})
	require.True(t, passed)

	passed = t.Run("SendUpdate_Sub_Unsub_ChUpdate_Respond_Accept", func(t *testing.T) {
		// Send channel update by bob.
		wg.Add(1)
		go func() {
			defer wg.Done()
			bobChInfo := bobChs[0].GetChInfo()
			var ownIdx, peerIdx int
			if bobChInfo.BalInfo.Parts[0] == perun.OwnAlias {
				ownIdx = 0
			} else {
				ownIdx = 1
			}
			peerIdx = ownIdx ^ 1
			amountToSend := decimal.NewFromFloat(0.5e18).BigInt()

			updater := func(state *pchannel.State) error {
				bals := state.Allocation.Clone().Balances[0]
				bals[ownIdx].Sub(bals[ownIdx], amountToSend)
				bals[peerIdx].Add(bals[peerIdx], amountToSend)
				state.Allocation.Balances[0] = bals
				return nil
			}

			_, err := bobChs[0].SendChUpdate(ctx, updater)
			require.NoError(t, err, "bob sending channel update")
		}()
		defer wg.Wait()

		// Accept channel update by alice.
		aliceChUpdateNotif := make(chan perun.ChUpdateNotif)
		aliceChUpdateNotifier := func(notif perun.ChUpdateNotif) {
			aliceChUpdateNotif <- notif
		}
		err = aliceChs[0].SubChUpdates(aliceChUpdateNotifier)
		require.NoError(t, err, "alice subscribing to channel updates")

		notif := <-aliceChUpdateNotif
		_, err = aliceChs[0].RespondChUpdate(ctx, notif.UpdateID, true)
		require.NoError(t, err, "alice accepting channel update")

		err = aliceChs[0].UnsubChUpdates()
		require.NoError(t, err, "alice unsubscribing from channel updates")
	})
	require.True(t, passed)

	passed = t.Run("SendUpdate_Sub_Unsub_ChUpdate_Respond_Reject", func(t *testing.T) {
		// Send channel update by alice.
		wg.Add(1)
		go func() {
			defer wg.Done()
			aliceChInfo := aliceChs[0].GetChInfo()
			var ownIdx, peerIdx int
			if aliceChInfo.BalInfo.Parts[0] == perun.OwnAlias {
				ownIdx = 0
			} else {
				ownIdx = 1
			}
			peerIdx = ownIdx ^ 1
			amountToSend := decimal.NewFromFloat(0.5e18).BigInt()

			updater := func(state *pchannel.State) error {
				bals := state.Allocation.Clone().Balances[0]
				bals[ownIdx].Sub(bals[ownIdx], amountToSend)
				bals[peerIdx].Add(bals[peerIdx], amountToSend)
				state.Allocation.Balances[0] = bals
				return nil
			}

			_, err := aliceChs[0].SendChUpdate(ctx, updater)
			require.Error(t, err, "alice sending channel update should be rejected by bob")
			t.Log(err)
		}()
		defer wg.Wait()

		// Reject channel update by bob.
		bobChUpdateNotif := make(chan perun.ChUpdateNotif)
		bobChUpdateNotifier := func(notif perun.ChUpdateNotif) {
			bobChUpdateNotif <- notif
		}
		err = bobChs[0].SubChUpdates(bobChUpdateNotifier)
		require.NoError(t, err, "bob subscribing to channel updates")

		notif := <-bobChUpdateNotif
		_, err = bobChs[0].RespondChUpdate(ctx, notif.UpdateID, false)
		require.NoError(t, err, "bob accepting channel update")

		err = bobChs[0].UnsubChUpdates()
		require.NoError(t, err, "bob unsubscribing from channel updates")
	})
	require.True(t, passed)

	passed = t.Run("Session_Close_NoForce_Error", func(t *testing.T) {
		var openChsInfo []perun.ChInfo
		openChsInfo, err = alice.Close(false)
		require.Error(t, err)
		t.Log(err)
		_ = openChsInfo
		// require.Len(t, openChsInfo, 2)
		// assert.Equal(t, aliceChs[0].ID(), openChsInfo[0].ChID)
	})
	require.True(t, passed)

	closeCh := func(chIndex int, isCollaborative bool) {
		// Subscribe to channel update notifications by Alice.
		aliceChUpdateNotif := make(chan perun.ChUpdateNotif, 1)
		aliceChUpdateNotifier := func(notif perun.ChUpdateNotif) {
			aliceChUpdateNotif <- notif
		}
		err = aliceChs[chIndex].SubChUpdates(aliceChUpdateNotifier)
		require.NoError(t, err, "alice subscribing to channel updates")

		// Send channel close by alice.
		wg.Add(1)
		go func() {
			defer wg.Done()

			closedChInfo, err := aliceChs[chIndex].Close(ctx)
			require.NoError(t, err, "alice closing channel")
			t.Log("alice", closedChInfo)
		}()
		defer wg.Wait()

		// Accept final channel update by bob (to enable collaborative close).
		bobChUpdateNotif := make(chan perun.ChUpdateNotif, 1)
		bobChUpdateNotifier := func(notif perun.ChUpdateNotif) {
			bobChUpdateNotif <- notif
		}
		err = bobChs[chIndex].SubChUpdates(bobChUpdateNotifier)
		require.NoError(t, err, "bob subscribing to channel updates")

		notif := <-bobChUpdateNotif
		// Accept if collaborative close is required, reject otherwise.
		_, err = bobChs[chIndex].RespondChUpdate(ctx, notif.UpdateID, isCollaborative)
		require.NoError(t, err, "bob accepting channel update")

		// Read channel (closing) update for bob.
		notif = <-bobChUpdateNotif
		t.Log("bob", notif)
		assert.Equal(t, perun.ChUpdateTypeClosed, notif.Type)

		// Responding to channel (closing) update.
		_, err = bobChs[chIndex].RespondChUpdate(ctx, notif.UpdateID, true)
		require.Error(t, err, "bob responding to (closing) channel update should error")

		err = bobChs[chIndex].UnsubChUpdates()
		assert.Error(t, err, "bob unsubscribing from channel updates after closing notification should error")
		t.Log(err)

		// Read channel (closing) update for alice.
		notif = <-aliceChUpdateNotif
		t.Log("alice", notif)
		assert.Equal(t, perun.ChUpdateTypeClosed, notif.Type)
		err = aliceChs[chIndex].UnsubChUpdates()
		assert.Error(t, err, "alice unsubscribing from channel updates after closing notification should error")
		t.Log(err)
	}

	passed = t.Run("Collaborative channel close", func(t *testing.T) {
		closeCh(0, true)
	})
	require.True(t, passed)

	passed = t.Run("Non_collaborative channel close", func(t *testing.T) {
		closeCh(1, false)
	})
	require.True(t, passed)

	passed = t.Run("Session_Close_NoForce_Success", func(t *testing.T) {
		var openChsInfo []perun.ChInfo
		openChsInfo, err = alice.Close(false)
		require.NoError(t, err, "alice closing session with no force option")
		require.Len(t, openChsInfo, 0)
	})
	require.True(t, passed)

	t.Run("Session_Close_Force_Success", func(t *testing.T) {
		var openChsInfo []perun.ChInfo
		openChsInfo, err = bob.Close(true)
		require.NoError(t, err, " bob closing session with force option")
		require.Len(t, openChsInfo, 0)
	})
}
