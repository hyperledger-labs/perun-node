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
	ethereumtest.SetupContractsT(t, ethereumtest.ChainURL, ethereumtest.OnChainTxTimeout)

	aliceAlias, bobAlias := "alice", "bob"

	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	aliceCfg := sessiontest.NewConfigT(t, prng)
	bobCfg := sessiontest.NewConfigT(t, prng)

	alice, err := session.New(aliceCfg)
	require.NoErrorf(t, err, "initializing alice session")
	t.Logf("alice session id: %s\n", alice.ID())
	t.Logf("alice database dir is: %s\n", aliceCfg.DatabaseDir)

	bob, err := session.New(bobCfg)
	require.NoErrorf(t, err, "initializing bob session")
	t.Logf("bob session id: %s\n", bob.ID())
	t.Logf("bob database dir is: %s\n", bobCfg.DatabaseDir)

	var alicePeerID, bobPeerID perun.PeerID
	t.Run("GetPeerID", func(t *testing.T) {
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

	t.Run("AddPeerID", func(t *testing.T) {
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

	const challengeDurSecs uint64 = 10
	wg := &sync.WaitGroup{}
	ctx := context.Background()

	t.Run("OpenCh_Sub_Unsub_ChProposal_Respond_Accept", func(t *testing.T) {
		// Propose Channel by alice.
		wg.Add(1)
		go func() {
			defer wg.Done()

			openingBalInfo := perun.BalInfo{
				Currency: currency.ETH,
				Parts:    []string{perun.OwnAlias, bobAlias},
				Bal:      []string{"1", "2"},
			}
			app := perun.App{
				Def:  pchannel.NoApp(),
				Data: pchannel.NoData(),
			}
			// nolint: govet	// err does not shadow, using a new var to prevent data race.
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
	})

	t.Run("OpenCh_Sub_Unsub_ChProposal_Respond_Reject", func(t *testing.T) {
		// Propose Channel by bob.
		wg.Add(1)
		go func() {
			defer wg.Done()

			openingBalInfo := perun.BalInfo{
				Currency: currency.ETH,
				Parts:    []string{aliceAlias, perun.OwnAlias},
				Bal:      []string{"1", "2"},
			}
			app := perun.App{
				Def:  pchannel.NoApp(),
				Data: pchannel.NoData(),
			}
			// nolint: govet	// err does not shadow, using a new var to prevent data race.
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

	var aliceCh, bobCh perun.ChAPI
	t.Run("GetChsInfo_GetCh", func(t *testing.T) {
		aliceChInfos := alice.GetChsInfo()
		require.Lenf(t, aliceChInfos, 1, "alice session should have exactly one channel")
		bobChInfos := bob.GetChsInfo()
		require.Lenf(t, bobChInfos, 1, "bob session should have exactly one channel")

		aliceCh, err = alice.GetCh(aliceChInfos[0].ChID)
		require.NoError(t, err, "getting alice ChAPI instance")

		bobCh, err = bob.GetCh(bobChInfos[0].ChID)
		require.NoError(t, err, "getting bob ChAPI instance")
	})

	t.Run("SendUpdate_Sub_Unsub_ChUpdate_Respond_Accept", func(t *testing.T) {
		// Send channel update by bob.
		wg.Add(1)
		go func() {
			defer wg.Done()
			bobChInfo := bobCh.GetChInfo()
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

			// nolint: govet	// err does not shadow, using a new var to prevent data race.
			_, err := bobCh.SendChUpdate(ctx, updater)
			require.NoError(t, err, "bob sending channel update")
		}()
		defer wg.Wait()

		// Accept channel update by alice.
		aliceChUpdateNotif := make(chan perun.ChUpdateNotif)
		aliceChUpdateNotifier := func(notif perun.ChUpdateNotif) {
			aliceChUpdateNotif <- notif
		}
		err = aliceCh.SubChUpdates(aliceChUpdateNotifier)
		require.NoError(t, err, "alice subscribing to channel updates")

		notif := <-aliceChUpdateNotif
		_, err = aliceCh.RespondChUpdate(ctx, notif.UpdateID, true)
		require.NoError(t, err, "alice accepting channel update")

		err = aliceCh.UnsubChUpdates()
		require.NoError(t, err, "alice unsubscribing from channel updates")
	})

	t.Run("SendUpdate_Sub_Unsub_ChUpdate_Respond_Reject", func(t *testing.T) {
		// Send channel update by alice.
		wg.Add(1)
		go func() {
			defer wg.Done()
			aliceChInfo := aliceCh.GetChInfo()
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

			// nolint: govet	// err does not shadow, using a new var to prevent data race.
			_, err := aliceCh.SendChUpdate(ctx, updater)
			require.Error(t, err, "alice sending channel update should be rejected by bob")
			t.Log(err)
		}()
		defer wg.Wait()

		// Reject channel update by bob.
		bobChUpdateNotif := make(chan perun.ChUpdateNotif)
		bobChUpdateNotifier := func(notif perun.ChUpdateNotif) {
			bobChUpdateNotif <- notif
		}
		err = bobCh.SubChUpdates(bobChUpdateNotifier)
		require.NoError(t, err, "bob subscribing to channel updates")

		notif := <-bobChUpdateNotif
		_, err = bobCh.RespondChUpdate(ctx, notif.UpdateID, false)
		require.NoError(t, err, "bob accepting channel update")

		err = bobCh.UnsubChUpdates()
		require.NoError(t, err, "bob unsubscribing from channel updates")
	})

	t.Run("Session_Close_NoForce_Error", func(t *testing.T) {
		var openChsInfo []perun.ChInfo
		openChsInfo, err = alice.Close(false)
		require.Error(t, err)
		t.Log(err)
		require.Len(t, openChsInfo, 1)
		assert.Equal(t, aliceCh.ID(), openChsInfo[0].ChID)
	})

	t.Run("Collaborative channel close", func(t *testing.T) {
		// Subscribe to channel update notifications by Alice.
		aliceChUpdateNotif := make(chan perun.ChUpdateNotif)
		aliceChUpdateNotifier := func(notif perun.ChUpdateNotif) {
			aliceChUpdateNotif <- notif
		}
		err = aliceCh.SubChUpdates(aliceChUpdateNotifier)
		require.NoError(t, err, "alice subscribing to channel updates")

		// Send channel close by alice.
		wg.Add(1)
		go func() {
			defer wg.Done()
			// nolint: govet	// err does not shadow, using a new var to prevent data race.
			closedChInfo, err := aliceCh.Close(ctx)
			require.NoError(t, err, "alice closing channel")
			t.Log("alice", closedChInfo)
		}()
		defer wg.Wait()

		// Accept final channel update by bob (to enable collaborative close).
		bobChUpdateNotif := make(chan perun.ChUpdateNotif)
		bobChUpdateNotifier := func(notif perun.ChUpdateNotif) {
			bobChUpdateNotif <- notif
		}
		err = bobCh.SubChUpdates(bobChUpdateNotifier)
		require.NoError(t, err, "bob subscribing to channel updates")

		notif := <-bobChUpdateNotif
		_, err = bobCh.RespondChUpdate(ctx, notif.UpdateID, true)
		require.NoError(t, err, "bob accepting channel update")

		// Read channel (closing) update for bob.
		notif = <-bobChUpdateNotif
		t.Log("bob", notif)
		assert.Equal(t, perun.ChUpdateTypeClosed, notif.Type)

		// Responding to channel (closing) update.
		_, err = bobCh.RespondChUpdate(ctx, notif.UpdateID, false)
		require.Error(t, err, "bob responding to (closing) channel update should error")

		err = bobCh.UnsubChUpdates()
		assert.Error(t, err, "bob unsubscribing from channel updates after closing notification should error")
		t.Log(err)

		apiError := perun.NewAPIErrV2FailedPreCondition(session.ErrChClosed.Error())
		require.Equal(t, err, apiError)

		// Read channel (closing) update for alice.
		notif = <-aliceChUpdateNotif
		t.Log("alice", notif)
		assert.Equal(t, perun.ChUpdateTypeClosed, notif.Type)
		err = aliceCh.UnsubChUpdates()
		assert.Error(t, err, "alice unsubscribing from channel updates after closing notification should error")
		t.Log(err)

		t.Run("Session_Close_NoForce_Success", func(t *testing.T) {
			var openChsInfo []perun.ChInfo
			openChsInfo, err = alice.Close(false)
			require.NoError(t, err, "alice closing session with no force option")
			require.Len(t, openChsInfo, 0)
		})

		t.Run("Session_Close_Force_Success", func(t *testing.T) {
			var openChsInfo []perun.ChInfo
			openChsInfo, err = bob.Close(true)
			require.NoError(t, err, " bob closing session with force option")
			require.Len(t, openChsInfo, 0)
		})
	})
}
