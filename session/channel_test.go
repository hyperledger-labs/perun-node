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

package session_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	pethchannel "perun.network/go-perun/backend/ethereum/channel"
	pchannel "perun.network/go-perun/channel"
	pclient "perun.network/go-perun/client"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/internal/mocks"
	"github.com/hyperledger-labs/perun-node/peruntest"
	"github.com/hyperledger-labs/perun-node/session"
)

const (
	responseTimeout         = 1 * time.Second
	challengeDurSecs uint64 = 10
)

func Test_ChAPI_Interface(t *testing.T) {
	assert.Implements(t, (*perun.ChAPI)(nil), new(session.Channel))
}

func Test_Getters(t *testing.T) {
	peers := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peers[0].Alias},
		Bal:      []string{"1", "2"},
	}

	pch, _ := newMockPCh()
	ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

	assert.Equal(t, ch.ID(), fmt.Sprintf("%x", pch.ID()))
	assert.Equal(t, ch.Currency(), validOpeningBalInfo.Currency)
	assert.Equal(t, ch.Parts(), validOpeningBalInfo.Parts)
	assert.Equal(t, ch.ChallengeDurSecs(), uint64(10))
}

func Test_GetChInfo(t *testing.T) {
	peers := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peers[0].Alias},
		Bal:      []string{"1", "2"},
	}

	t.Run("happy", func(t *testing.T) {
		pch, _ := newMockPCh()
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		chInfo := ch.GetChInfo()
		assert.Equal(t, chInfo.ChID, fmt.Sprintf("%x", pch.ID()))
		assert.Equal(t, chInfo.BalInfo.Parts, validOpeningBalInfo.Parts)
		assert.Equal(t, chInfo.BalInfo.Currency, validOpeningBalInfo.Currency)
	})

	t.Run("nil_state", func(t *testing.T) {
		pch, _ := newMockPCh()
		pch.On("State").Return(nil)

		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		chInfo := ch.GetChInfo()
		assert.Zero(t, chInfo)
	})
}

func Test_SendChUpdate(t *testing.T) {
	peers := newPeerIDs(t, uint(2))
	peerAlias := peers[0].Alias
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peerAlias},
		Bal:      []string{"1", "2"},
	}
	ourIdx := 0
	noopUpdater := func(s *pchannel.State) error {
		return nil
	}

	t.Run("happy", func(t *testing.T) {
		pch, _ := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(1))
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(nil)
		gotChInfo, err := ch.SendChUpdate(context.Background(), noopUpdater)
		require.NoError(t, err)
		assert.NotZero(t, gotChInfo)
	})

	t.Run("channel_closed", func(t *testing.T) {
		pch, _ := newMockPCh()
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)

		_, err := ch.SendChUpdate(context.Background(), noopUpdater)

		wantMessage := session.ErrChClosed.Error()
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrFailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
	})

	t.Run("UpdateBy_PeerRequestTimedOut", func(t *testing.T) {
		timeout := responseTimeout.String()
		peerRequestTimedOutError := pclient.RequestTimedOutError("some-error")
		pch, _ := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(ourIdx))
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(peerRequestTimedOutError)
		_, err := ch.SendChUpdate(context.Background(), noopUpdater)

		peruntest.AssertAPIError(t, err, perun.ParticipantError, perun.ErrPeerRequestTimedOut)
		peruntest.AssertErrInfoPeerRequestTimedOut(t, err.AddInfo(), peerAlias, timeout)
	})

	t.Run("UpdateBy_RejectedByPeer", func(t *testing.T) {
		reason := "some random reason"
		peerRejectedError := pclient.PeerRejectedError{
			ItemType: "channel update",
			Reason:   reason,
		}
		pch, _ := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(ourIdx))
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(peerRejectedError)
		_, err := ch.SendChUpdate(context.Background(), noopUpdater)

		peruntest.AssertAPIError(t, err, perun.ParticipantError, perun.ErrPeerRejected)
		peruntest.AssertErrInfoPeerRejected(t, err.AddInfo(), peerAlias, reason)
	})
}

func Test_HandleUpdate(t *testing.T) {
	peers := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peers[0].Alias},
		Bal:      []string{"1", "2"},
	}
	updatedBalInfo := validOpeningBalInfo
	updatedBalInfo.Bal = []string{"0.5", "2.5"}
	pch, _ := newMockPCh()

	currState := makeState(t, validOpeningBalInfo, false)
	nonFinalState := makeState(t, updatedBalInfo, false)

	t.Run("happy_ignore_chUpdate_when_closed", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: nonFinalState,
		}
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)
		ch.HandleUpdate(currState, *chUpdate, &mocks.ChUpdateResponder{})
	})
	// Tests if handler does not panic on receiving update when channel is closed.
}

func Test_SubUnsubChUpdate(t *testing.T) {
	peers := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peers[0].Alias},
		Bal:      []string{"1", "2"},
	}

	dummyNotifier := func(notif perun.ChUpdateNotif) {}
	pch, _ := newMockPCh()
	ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

	// SubTest 1: Sub successfully ==
	err := ch.SubChUpdates(dummyNotifier)
	require.NoError(t, err)

	// SubTest 2: Sub again, should error ==
	err = ch.SubChUpdates(dummyNotifier)
	require.Error(t, err)

	peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrResourceExists)
	peruntest.AssertErrInfoResourceExists(t, err.AddInfo(), session.ResTypeUpdateSub, ch.ID())

	// SubTest 3: UnSub successfully ==
	err = ch.UnsubChUpdates()
	require.NoError(t, err)

	// SubTest 4: UnSub again, should error ==
	err = ch.UnsubChUpdates()
	require.Error(t, err)

	peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrResourceNotFound)
	peruntest.AssertErrInfoResourceNotFound(t, err.AddInfo(), session.ResTypeUpdateSub, ch.ID())

	t.Run("Sub_channelClosed", func(t *testing.T) {
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)
		err = ch.SubChUpdates(dummyNotifier)

		wantMessage := session.ErrChClosed.Error()
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrFailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
	})
	t.Run("Unsub_channelClosed", func(t *testing.T) {
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)
		err = ch.UnsubChUpdates()

		wantMessage := session.ErrChClosed.Error()
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrFailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
	})
}

func Test_HandleUpdate_Sub(t *testing.T) {
	peers := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peers[0].Alias},
		Bal:      []string{"1", "2"},
	}
	updatedBalInfo := validOpeningBalInfo
	updatedBalInfo.Bal = []string{"0.5", "2.5"}
	pch, _ := newMockPCh()

	currState := makeState(t, validOpeningBalInfo, false)
	nonFinalState := makeState(t, updatedBalInfo, false)
	finalState := makeState(t, updatedBalInfo, true)

	t.Run("happy_HandleSub_nonFinal", func(t *testing.T) { //nolint: dupl
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		chUpdate := &pclient.ChannelUpdate{
			State: nonFinalState,
		}
		ch.HandleUpdate(currState, *chUpdate, &mocks.ChUpdateResponder{})

		notifs := make([]perun.ChUpdateNotif, 0, 2)
		notifier := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		err := ch.SubChUpdates(notifier)
		require.NoError(t, err)
		notifRecieved := func() bool {
			if len(notifs) != 1 {
				return false
			}
			return notifs[0].Type == perun.ChUpdateTypeOpen
		}
		assert.Eventually(t, notifRecieved, 2*time.Second, 100*time.Millisecond)
	})

	t.Run("happy_HandleSub_Final", func(t *testing.T) { //nolint: dupl
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		chUpdate := &pclient.ChannelUpdate{
			State: finalState,
		}
		ch.HandleUpdate(currState, *chUpdate, &mocks.ChUpdateResponder{})

		notifs := make([]perun.ChUpdateNotif, 0, 2)
		notifier := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		err := ch.SubChUpdates(notifier)
		require.NoError(t, err)
		notifRecieved := func() bool {
			if len(notifs) != 1 {
				return false
			}
			return notifs[0].Type == perun.ChUpdateTypeFinal
		}
		assert.Eventually(t, notifRecieved, 2*time.Second, 100*time.Millisecond)
	})

	t.Run("happy_SubHandle", func(t *testing.T) {
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		notifs := make([]perun.ChUpdateNotif, 0, 2)
		notifier := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		err := ch.SubChUpdates(notifier)
		require.NoError(t, err)
		notifRecieved := func() bool {
			if len(notifs) != 1 {
				return false
			}
			return notifs[0].Type == perun.ChUpdateTypeOpen
		}

		chUpdate := &pclient.ChannelUpdate{
			State: nonFinalState,
		}
		ch.HandleUpdate(currState, *chUpdate, &mocks.ChUpdateResponder{})
		assert.Eventually(t, notifRecieved, 2*time.Second, 100*time.Millisecond)
	})
}

func Test_HandleUpdate_Respond(t *testing.T) {
	peers := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peers[0].Alias},
		Bal:      []string{"1", "2"},
	}
	updatedBalInfo := validOpeningBalInfo
	updatedBalInfo.Bal = []string{"0.5", "2.5"}
	pch, _ := newMockPCh()
	pch.On("State").Return(makeState(t, validOpeningBalInfo, false))

	currState := makeState(t, validOpeningBalInfo, false)
	nonFinalState := makeState(t, updatedBalInfo, false)
	nonFinalState.Version++
	finalState := makeState(t, updatedBalInfo, true)
	finalState.Version++

	t.Run("happy_accept", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: nonFinalState,
		}
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Accept", mock.Anything).Return(nil)
		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		chInfo, err := ch.RespondChUpdate(context.Background(), updateID, true)
		require.NoError(t, err)
		assert.NotZero(t, chInfo)
	})

	t.Run("happy_accept_Final", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: finalState,
		}
		pch, _ := newMockPCh() //nolint: govet		// this does not shadow prev decl. declared in same name on purpose.
		pch.On("State").Return(finalState)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Accept", mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(nil)

		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		chInfo, err := ch.RespondChUpdate(context.Background(), updateID, true)
		require.NoError(t, err)
		assert.NotZero(t, chInfo)
	})

	t.Run("happy_reject", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: nonFinalState,
		}
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Reject", mock.Anything, mock.Anything).Return(nil)
		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		chInfo, err := ch.RespondChUpdate(context.Background(), updateID, false)
		require.NoError(t, err)
		assert.NotZero(t, chInfo)
	})

	t.Run("respond_channel_closed", func(t *testing.T) {
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)
		updateID := "any-update-id" // A closed channel returns error irrespective of update id.

		_, err := ch.RespondChUpdate(context.Background(), updateID, true)

		wantMessage := session.ErrChClosed.Error()
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrFailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
	})

	t.Run("unknown_UpdateID", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: nonFinalState,
		}
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Accept", mock.Anything).Return(nil)
		unknownUpdateID := "random-update-id"
		ch.HandleUpdate(currState, *chUpdate, responder)

		_, err := ch.RespondChUpdate(context.Background(), unknownUpdateID, true)

		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrResourceNotFound)
		peruntest.AssertErrInfoResourceNotFound(t, err.AddInfo(), session.ResTypeUpdate, unknownUpdateID)
	})

	t.Run("response_timeout_expired", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: nonFinalState,
		}
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Accept", mock.Anything).Return(nil)
		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		time.Sleep(2 * time.Second)
		_, err := ch.RespondChUpdate(context.Background(), updateID, true)

		peruntest.AssertAPIError(t, err, perun.ParticipantError, perun.ErrUserResponseTimedOut)
		peruntest.AssertErrInfoUserResponseTimedOut(t, err.AddInfo())
	})

	t.Run("respond_accept_Error", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: nonFinalState,
		}
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Accept", mock.Anything).Return(assert.AnError)
		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		_, err := ch.RespondChUpdate(context.Background(), updateID, true)

		peruntest.AssertAPIError(t, err, perun.InternalError, perun.ErrUnknownInternal)
	})

	t.Run("respond_reject_Error", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: nonFinalState,
		}
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Reject", mock.Anything, mock.Anything).Return(assert.AnError)
		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		_, err := ch.RespondChUpdate(context.Background(), updateID, false)

		peruntest.AssertAPIError(t, err, perun.InternalError, perun.ErrUnknownInternal)
	})

	t.Run("Handle_accept_register_AnError", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: finalState,
		}
		pch, _ := newMockPCh()
		pch.On("State").Return(finalState)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Accept", mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(assert.AnError)

		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		_, err := ch.RespondChUpdate(context.Background(), updateID, true)

		peruntest.AssertAPIError(t, err, perun.InternalError, perun.ErrUnknownInternal)
	})

	t.Run("Handle_accept_register_TxTimeoutError", func(t *testing.T) {
		txTimedOutError := pclient.TxTimedoutError{
			TxType: pethchannel.Register.String(),
			TxID:   "0xabcd",
		}
		chUpdate := &pclient.ChannelUpdate{
			State: finalState,
		}
		pch, _ := newMockPCh()
		pch.On("State").Return(finalState)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Accept", mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(txTimedOutError)

		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		_, err := ch.RespondChUpdate(context.Background(), updateID, true)

		txType := txTimedOutError.TxType
		txID := txTimedOutError.TxID
		txTimeout := ethereumtest.OnChainTxTimeout.String()
		peruntest.AssertAPIError(t, err, perun.ProtocolFatalError, perun.ErrTxTimedOut, txTimedOutError.Error())
		peruntest.AssertErrInfoTxTimedOut(t, err.AddInfo(), txType, txID, txTimeout)
	})

	t.Run("Handle_accept_register_ChainNotReachable", func(t *testing.T) {
		chainURL := ethereumtest.ChainURL
		chainNotReachableError := pclient.ChainNotReachableError{}
		chUpdate := &pclient.ChannelUpdate{
			State: finalState,
		}
		pch, _ := newMockPCh()
		pch.On("State").Return(finalState)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Accept", mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(chainNotReachableError)

		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		_, err := ch.RespondChUpdate(context.Background(), updateID, true)

		peruntest.AssertAPIError(t, err, perun.ProtocolFatalError, perun.ErrChainNotReachable)
		peruntest.AssertErrInfoChainNotReachable(t, err.AddInfo(), chainURL)
	})
}

func Test_Close(t *testing.T) {
	peers := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peers[0].Alias},
		Bal:      []string{"1", "2"},
	}
	peerIdx := 1

	assertNotif := func(t *testing.T, notifs []perun.ChUpdateNotif, wantVersion uint64, wantExpiry int64) {
		t.Helper()
		require.Len(t, notifs, 1)
		require.Equal(t, perun.ChUpdateTypeClosed, notifs[0].Type)
		require.Equal(t, fmt.Sprintf("%d", wantVersion), notifs[0].CurrChInfo.Version)
		require.Equal(t, wantExpiry, notifs[0].Expiry)
	}

	t.Run("happy_forInitiator_finalized_settle_notify", func(t *testing.T) {
		pch, watcherSignal := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		var finalizer perun.StateUpdater
		pch.On("Idx").Return(pchannel.Index(peerIdx))
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false)).Once()
		pch.On("UpdateBy", mock.Anything, mock.MatchedBy(func(gotFinalizer perun.StateUpdater) bool {
			finalizer = gotFinalizer
			return true
		})).Return(nil)
		pch.On("Register", mock.Anything).Return(nil)

		// == Part 1: Check if channel close was initialized correctly.
		gotChInfo, err := ch.Close(context.Background())
		require.NoError(t, err)
		assert.NotZero(t, gotChInfo)

		// == Part 2: Check if finalizer sent to UpdateBy marks the channel as final.
		emptyState := pchannel.State{}
		assert.NoError(t, finalizer(&emptyState))
		assert.True(t, emptyState.IsFinal)

		// == Part 3: Simulate concluded event and test check if channel close notification is sent.
		var concludedVersion uint64 = 1 // will have been incremented in finalizing update.
		concludedEvent := &pchannel.ConcludedEvent{
			AdjudicatorEventBase: *pchannel.NewAdjudicatorEventBase(
				pch.ID(), &pchannel.ElapsedTimeout{}, concludedVersion),
		}
		finalizedState := makeState(t, validOpeningBalInfo, true)
		finalizedState.Version++
		pch.On("State").Return(finalizedState)
		pch.On("Settle", mock.Anything, mock.Anything).Return(nil)
		pch.On("Close").Return(nil).Run(func(args mock.Arguments) {
			watcherSignal <- time.Now() // Signal the watcher to return when pch is closed.
		})

		wantExpiry := int64(0)
		notifs := make([]perun.ChUpdateNotif, 0, 1) // Subscribe to channel close notification.
		notifer := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		require.NoError(t, ch.SubChUpdates(notifer))

		ch.HandleAdjudicatorEvent(concludedEvent)
		assertNotif(t, notifs, concludedVersion, wantExpiry)
	})

	t.Run("happy_forNonInitiator_finalized_settle_notify", func(t *testing.T) {
		pch, watcherSignal := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		// Simulate concluded event and test check if channel close notification is sent.
		var concludedVersion uint64 = 1 // will have been incremented in finalizing update.
		concludedEvent := &pchannel.ConcludedEvent{
			AdjudicatorEventBase: *pchannel.NewAdjudicatorEventBase(pch.ID(), &pchannel.ElapsedTimeout{}, concludedVersion),
		}
		finalizedState := makeState(t, validOpeningBalInfo, true)
		finalizedState.Version++
		pch.On("State").Return(finalizedState)
		pch.On("Settle", mock.Anything, mock.Anything).Return(nil)
		pch.On("Close").Return(nil).Run(func(args mock.Arguments) {
			watcherSignal <- time.Now() // Signal the watcher to return when pch is closed.
		})

		wantExpiry := int64(0)
		notifs := make([]perun.ChUpdateNotif, 0, 1) // Subscribe to channel close notification.
		notifer := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		require.NoError(t, ch.SubChUpdates(notifer))

		ch.HandleAdjudicatorEvent(concludedEvent)
		assertNotif(t, notifs, concludedVersion, wantExpiry)
	})

	t.Run("happy_ForInitiator_notFinalized_settle_notify", func(t *testing.T) {
		pch, watcherSignal := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(peerIdx))
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(assert.AnError)
		pch.On("Register", mock.Anything).Return(nil)

		// == Part 1: Check if channel close was initialized correctly.
		gotChInfo, err := ch.Close(context.Background())
		require.NoError(t, err)
		assert.NotZero(t, gotChInfo)

		// == Part 2: Simulate registered event and test check if channel close notification is sent.
		var registeredVersion uint64
		registeredEvent := pchannel.NewRegisteredEvent(pch.ID(), &pchannel.ElapsedTimeout{}, registeredVersion)
		pch.On("Settle", mock.Anything, mock.Anything).Return(nil)
		pch.On("Close").Return(nil).Run(func(args mock.Arguments) {
			watcherSignal <- time.Now() // Signal the watcher to return when pch is closed.
		})

		wantExpiry := int64(0)
		notifs := make([]perun.ChUpdateNotif, 0, 1) // Subscribe to channel close notification.
		notifer := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		require.NoError(t, ch.SubChUpdates(notifer))

		ch.HandleAdjudicatorEvent(registeredEvent)
		assertNotif(t, notifs, registeredVersion, wantExpiry)
	})

	t.Run("happy_forNonInitiator_notFinalized_settle_notify", func(t *testing.T) {
		pch, watcherSignal := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		// Simulate registered event and test check if channel close notification is sent.
		var registeredVersion uint64
		registeredEvent := pchannel.NewRegisteredEvent(pch.ID(), &pchannel.ElapsedTimeout{}, registeredVersion)
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		pch.On("Settle", mock.Anything, mock.Anything).Return(nil)
		pch.On("Close").Return(nil).Run(func(args mock.Arguments) {
			watcherSignal <- time.Now() // Signal the watcher to return when pch is closed.
		})

		wantExpiry := int64(0)
		notifs := make([]perun.ChUpdateNotif, 0, 1) // Subscribe to channel close notification.
		notifer := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		require.NoError(t, ch.SubChUpdates(notifer))

		ch.HandleAdjudicatorEvent(registeredEvent)
		assertNotif(t, notifs, registeredVersion, wantExpiry)
	})

	// Test for errors returned by register are implemented only for
	// nonInitiator, because, for NonInitiator, register is called after
	// accepting the final update in RespondChUpdate API.
	//
	// Also, we test only for the finalized case, because from the perspective
	// of accessing the go-perun API, there is not difference between calling
	// register on a finalized or non-finalized channel.
	t.Run("forInitiator_finalized_register_AnError", func(t *testing.T) {
		pch, _ := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(peerIdx))
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(assert.AnError)

		_, err := ch.Close(context.Background())

		peruntest.AssertAPIError(t, err, perun.InternalError, perun.ErrUnknownInternal)
	})

	t.Run("forInitiator_finalized_register_TxTimeoutError", func(t *testing.T) {
		txTimedOutError := pclient.TxTimedoutError{
			TxType: pethchannel.Register.String(),
			TxID:   "0xabcd",
		}
		pch, _ := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(peerIdx))
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(txTimedOutError)

		_, err := ch.Close(context.Background())

		txType := txTimedOutError.TxType
		txID := txTimedOutError.TxID
		txTimeout := ethereumtest.OnChainTxTimeout.String()
		peruntest.AssertAPIError(t, err, perun.ProtocolFatalError, perun.ErrTxTimedOut)
		peruntest.AssertErrInfoTxTimedOut(t, err.AddInfo(), txType, txID, txTimeout)
	})

	t.Run("forInitiator_finalized_register_ChainNotReachable", func(t *testing.T) {
		chainURL := ethereumtest.ChainURL
		chainNotReachableError := pclient.ChainNotReachableError{}
		pch, _ := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(peerIdx))
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(chainNotReachableError)

		_, err := ch.Close(context.Background())

		peruntest.AssertAPIError(t, err, perun.ProtocolFatalError, perun.ErrChainNotReachable)
		peruntest.AssertErrInfoChainNotReachable(t, err.AddInfo(), chainURL)
	})

	// While on-chain errors can occur on all cases,
	// (finalzed/notFinalized) (forInitiator/forNonInitiator), all paths will
	// use the same "ch.settle" function. So the tests are implemented in the
	// notFinalized, forNonInitiator case, because this requires minimal setup.
	t.Run("notFinalized_settle_AnError", func(t *testing.T) {
		pch, watcherSignal := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		// Simulate registered event and test check if channel close notification with error is sent.
		var registeredVersion uint64
		registeredEvent := pchannel.NewRegisteredEvent(pch.ID(), &pchannel.ElapsedTimeout{}, registeredVersion)
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		pch.On("Settle", mock.Anything, mock.Anything).Return(assert.AnError)
		pch.On("Close").Return(nil).Run(func(args mock.Arguments) {
			watcherSignal <- time.Now() // Signal the watcher to return when pch is closed.
		})

		wantExpiry := int64(0)
		notifs := make([]perun.ChUpdateNotif, 0, 1) // Subscribe to channel close notification.
		notifer := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		require.NoError(t, ch.SubChUpdates(notifer))

		ch.HandleAdjudicatorEvent(registeredEvent)
		assertNotif(t, notifs, registeredVersion, wantExpiry)

		peruntest.AssertAPIError(t, notifs[0].Error, perun.InternalError, perun.ErrUnknownInternal)
	})

	t.Run("notFinalized_settle_TxTimedout", func(t *testing.T) {
		txTimedOutError := pclient.TxTimedoutError{
			TxType: pethchannel.Withdraw.String(),
			TxID:   "0xabcd",
		}
		pch, watcherSignal := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		// Simulate registered event and test check if channel close notification with error is sent.
		var registeredVersion uint64
		registeredEvent := pchannel.NewRegisteredEvent(pch.ID(), &pchannel.ElapsedTimeout{}, registeredVersion)
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		pch.On("Settle", mock.Anything, mock.Anything).Return(txTimedOutError)
		pch.On("Close").Return(nil).Run(func(args mock.Arguments) {
			watcherSignal <- time.Now() // Signal the watcher to return when pch is closed.
		})

		wantExpiry := int64(0)
		notifs := make([]perun.ChUpdateNotif, 0, 1) // Subscribe to channel close notification.
		notifer := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		require.NoError(t, ch.SubChUpdates(notifer))

		ch.HandleAdjudicatorEvent(registeredEvent)
		assertNotif(t, notifs, registeredVersion, wantExpiry)

		txType := txTimedOutError.TxType
		txID := txTimedOutError.TxID
		txTimeout := ethereumtest.OnChainTxTimeout.String()
		peruntest.AssertAPIError(t, notifs[0].Error, perun.ProtocolFatalError, perun.ErrTxTimedOut, txTimedOutError.Error())
		peruntest.AssertErrInfoTxTimedOut(t, notifs[0].Error.AddInfo(), txType, txID, txTimeout)
	})

	t.Run("notFinalized_settle_ChainNotReachable", func(t *testing.T) {
		chainURL := ethereumtest.ChainURL
		chainNotReachableError := pclient.ChainNotReachableError{}
		pch, watcherSignal := newMockPCh()
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		// Simulate registered event and test check if channel close notification with error is sent.
		var registeredVersion uint64
		registeredEvent := pchannel.NewRegisteredEvent(pch.ID(), &pchannel.ElapsedTimeout{}, registeredVersion)
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		pch.On("Settle", mock.Anything, mock.Anything).Return(chainNotReachableError)
		pch.On("Close").Return(nil).Run(func(args mock.Arguments) {
			watcherSignal <- time.Now() // Signal the watcher to return when pch is closed.
		})

		wantExpiry := int64(0)
		notifs := make([]perun.ChUpdateNotif, 0, 1) // Subscribe to channel close notification.
		notifer := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		require.NoError(t, ch.SubChUpdates(notifer))

		ch.HandleAdjudicatorEvent(registeredEvent)
		assertNotif(t, notifs, registeredVersion, wantExpiry)

		peruntest.AssertAPIError(t, notifs[0].Error, perun.ProtocolFatalError, perun.ErrChainNotReachable)
		peruntest.AssertErrInfoChainNotReachable(t, notifs[0].Error.AddInfo(), chainURL)
	})

	t.Run("channel_closed", func(t *testing.T) {
		pch, _ := newMockPCh()
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)

		_, err := ch.Close(context.Background())
		require.Error(t, err)

		wantMessage := session.ErrChClosed.Error()
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrFailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
	})
}
