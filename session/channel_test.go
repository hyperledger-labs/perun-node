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

	pch, _ := newMockPCh(t, validOpeningBalInfo)
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
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		chInfo := ch.GetChInfo()
		assert.Equal(t, chInfo.ChID, fmt.Sprintf("%x", pch.ID()))
		assert.Equal(t, chInfo.BalInfo.Parts, validOpeningBalInfo.Parts)
		assert.Equal(t, chInfo.BalInfo.Currency, validOpeningBalInfo.Currency)
	})

	t.Run("nil_state", func(t *testing.T) {
		pch := &mocks.Channel{}
		pch.On("ID").Return([32]byte{0, 1, 2})
		pch.On("State").Return(nil)
		watcherSignal := make(chan time.Time)
		pch.On("Watch", mock.Anything).WaitUntil(watcherSignal).Return(nil)

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
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(1))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(nil)
		gotChInfo, err := ch.SendChUpdate(context.Background(), noopUpdater)
		require.NoError(t, err)
		assert.NotZero(t, gotChInfo)
	})

	t.Run("channel_closed", func(t *testing.T) {
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)

		_, err := ch.SendChUpdate(context.Background(), noopUpdater)
		require.Error(t, err)

		wantMessage := session.ErrChClosed.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2FailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
	})

	t.Run("UpdateBy_PeerRequestTimedOut", func(t *testing.T) {
		timeout := responseTimeout.String()
		peerRequestTimedOutError := pclient.RequestTimedOutError("some-error")
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(ourIdx))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(peerRequestTimedOutError)
		_, err := ch.SendChUpdate(context.Background(), noopUpdater)

		wantMessage := peerRequestTimedOutError.Error()
		assertAPIError(t, err, perun.ParticipantError, perun.ErrV2PeerRequestTimedOut, wantMessage)
		assertErrV2InfoPeerRequestTimedOut(t, err.AddInfo(), peerAlias, timeout)
	})

	t.Run("UpdateBy_RejectedByPeer", func(t *testing.T) {
		reason := "some random reason"
		peerRejectedError := pclient.PeerRejectedError{
			ItemType: "channel update",
			Reason:   reason,
		}
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(ourIdx))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(peerRejectedError)
		_, err := ch.SendChUpdate(context.Background(), noopUpdater)

		wantMessage := peerRejectedError.Error()
		assertAPIError(t, err, perun.ParticipantError, perun.ErrV2PeerRejected, wantMessage)
		assertErrV2InfoPeerRejected(t, err.AddInfo(), peerAlias, reason)
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
	pch, _ := newMockPCh(t, validOpeningBalInfo)

	currState := makeState(t, validOpeningBalInfo, false)
	nonFinalState := makeState(t, updatedBalInfo, false)
	finalState := makeState(t, updatedBalInfo, true)

	t.Run("happy_nonFinal", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: nonFinalState,
		}
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		ch.HandleUpdate(currState, *chUpdate, &mocks.ChUpdateResponder{})
	})

	t.Run("happy_final", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: finalState,
		}
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		ch.HandleUpdate(currState, *chUpdate, &mocks.ChUpdateResponder{})
	})

	t.Run("happy_unexpected_chUpdate", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: nonFinalState,
		}
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)
		ch.HandleUpdate(currState, *chUpdate, &mocks.ChUpdateResponder{})
	})
}

func Test_SubUnsubChUpdate(t *testing.T) {
	peers := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peers[0].Alias},
		Bal:      []string{"1", "2"},
	}

	dummyNotifier := func(notif perun.ChUpdateNotif) {}
	pch, _ := newMockPCh(t, validOpeningBalInfo)
	ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

	// SubTest 1: Sub successfully ==
	err := ch.SubChUpdates(dummyNotifier)
	require.NoError(t, err)

	// SubTest 2: Sub again, should error ==
	err = ch.SubChUpdates(dummyNotifier)
	require.Error(t, err)

	wantMessage := session.ErrSubAlreadyExists.Error()
	assertAPIError(t, err, perun.ClientError, perun.ErrV2ResourceExists, wantMessage)
	assertErrV2InfoResourceExists(t, err.AddInfo(), "subscription to channel updates", ch.ID())

	// SubTest 3: UnSub successfully ==
	err = ch.UnsubChUpdates()
	require.NoError(t, err)

	// SubTest 4: UnSub again, should error ==
	err = ch.UnsubChUpdates()
	require.Error(t, err)

	wantMessage = session.ErrNoActiveSub.Error()
	assertAPIError(t, err, perun.ClientError, perun.ErrV2ResourceNotFound, wantMessage)
	assertErrV2InfoResourceNotFound(t, err.AddInfo(), "subscription to channel updates", ch.ID())

	t.Run("Sub_channelClosed", func(t *testing.T) {
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)
		err = ch.SubChUpdates(dummyNotifier)
		require.Error(t, err)
		wantMessage := session.ErrChClosed.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2FailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
	})
	t.Run("Unsub_channelClosed", func(t *testing.T) {
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)
		err = ch.UnsubChUpdates()
		require.Error(t, err)
		wantMessage := session.ErrChClosed.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2FailedPreCondition, wantMessage)
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
	pch, _ := newMockPCh(t, validOpeningBalInfo)

	currState := makeState(t, validOpeningBalInfo, false)
	nonFinalState := makeState(t, updatedBalInfo, false)
	t.Run("happy_HandleSub", func(t *testing.T) {
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
			return len(notifs) == 1
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
			return len(notifs) == 1
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
	pch, _ := newMockPCh(t, validOpeningBalInfo)

	currState := makeState(t, validOpeningBalInfo, false)
	nonFinalState := makeState(t, updatedBalInfo, false)
	finalState := makeState(t, updatedBalInfo, true)

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
		require.Error(t, err)

		wantMessage := session.ErrChClosed.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2FailedPreCondition, wantMessage)
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
		require.Error(t, err)

		assertAPIError(t, err, perun.ClientError, perun.ErrV2ResourceNotFound, "update")
		assertErrV2InfoResourceNotFound(t, err.AddInfo(), "update", unknownUpdateID)
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
		require.Error(t, err)

		assertAPIError(t, err, perun.ParticipantError, perun.ErrV2UserResponseTimedOut, "")
		assertErrV2InfoUserResponseTimedout(t, err.AddInfo())
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
		require.Error(t, err)
		assertAPIError(t, err, perun.InternalError, perun.ErrV2UnknownInternal, assert.AnError.Error())
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
		require.Error(t, err)
		assertAPIError(t, err, perun.InternalError, perun.ErrV2UnknownInternal, assert.AnError.Error())
	})

	t.Run("happy_accept_Final", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: finalState,
		}
		pch, _ := newMockPCh(t, validOpeningBalInfo)
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

	t.Run("Handle_accept_register_AnError", func(t *testing.T) {
		chUpdate := &pclient.ChannelUpdate{
			State: finalState,
		}
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Accept", mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(assert.AnError)

		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		_, err := ch.RespondChUpdate(context.Background(), updateID, true)
		require.Error(t, err)
		assertAPIError(t, err, perun.InternalError, perun.ErrV2UnknownInternal, assert.AnError.Error())
	})

	t.Run("Handle_accept_register_TxTimeoutError", func(t *testing.T) {
		txTimedoutError := pclient.TxTimedoutError{
			TxType: pethchannel.Register.String(),
			TxID:   "0xabcd",
		}
		chUpdate := &pclient.ChannelUpdate{
			State: finalState,
		}
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Accept", mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(txTimedoutError)

		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		_, err := ch.RespondChUpdate(context.Background(), updateID, true)
		require.Error(t, err)
		assertAPIError(t, err, perun.ProtocolFatalError, perun.ErrV2TxTimedOut, txTimedoutError.Error())
		txType := txTimedoutError.TxType
		txID := txTimedoutError.TxID
		txTimeout := ethereumtest.OnChainTxTimeout.String()
		assertErrV2InfoTxTimedOut(t, err.AddInfo(), txType, txID, txTimeout)
	})

	t.Run("Handle_accept_register_ChainNotReachable", func(t *testing.T) {
		chainURL := ethereumtest.ChainURL
		chainNotReachableError := pclient.ChainNotReachableError{}
		chUpdate := &pclient.ChannelUpdate{
			State: finalState,
		}
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Accept", mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(chainNotReachableError)

		updateID := fmt.Sprintf("%s_%d", ch.ID(), chUpdate.State.Version)
		ch.HandleUpdate(currState, *chUpdate, responder)

		_, err := ch.RespondChUpdate(context.Background(), updateID, true)
		require.Error(t, err)
		wantMessage := chainNotReachableError.Error()
		assertAPIError(t, err, perun.ProtocolFatalError, perun.ErrV2ChainNotReachable, wantMessage)
		assertErrV2InfoChainNotReachable(t, err.AddInfo(), chainURL)
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

	t.Run("happy_finalizeNoError_settle", func(t *testing.T) {
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		var finalizer perun.StateUpdater
		pch.On("Idx").Return(pchannel.Index(peerIdx))
		pch.On("UpdateBy", mock.Anything, mock.MatchedBy(func(gotFinalizer perun.StateUpdater) bool {
			finalizer = gotFinalizer
			return true
		})).Return(nil)
		pch.On("Register", mock.Anything).Return(nil)

		gotChInfo, err := ch.Close(context.Background())
		require.NoError(t, err)
		assert.NotZero(t, gotChInfo)

		emptyState := pchannel.State{}
		assert.NoError(t, finalizer(&emptyState))
		assert.True(t, emptyState.IsFinal)
	})

	t.Run("happy_finalizeError_settle", func(t *testing.T) {
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(peerIdx))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(assert.AnError)
		pch.On("Register", mock.Anything).Return(nil)

		gotChInfo, err := ch.Close(context.Background())
		require.NoError(t, err)
		assert.NotZero(t, gotChInfo)
	})

	t.Run("happy_closeError", func(t *testing.T) {
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(peerIdx))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(nil)

		gotChInfo, err := ch.Close(context.Background())
		require.NoError(t, err)
		assert.NotZero(t, gotChInfo)
	})

	t.Run("channel_closed", func(t *testing.T) {
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)

		_, err := ch.Close(context.Background())
		require.Error(t, err)

		wantMessage := session.ErrChClosed.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2FailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
	})

	t.Run("finalized_settle_AnError", func(t *testing.T) {
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(peerIdx))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(assert.AnError)

		_, err := ch.Close(context.Background())
		require.Error(t, err)
		assertAPIError(t, err, perun.InternalError, perun.ErrV2UnknownInternal, assert.AnError.Error())
	})

	t.Run("finalized_settle_TxTimeoutError", func(t *testing.T) {
		txTimedoutError := pclient.TxTimedoutError{
			TxType: pethchannel.Register.String(),
			TxID:   "0xabcd",
		}
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(peerIdx))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(txTimedoutError)

		_, err := ch.Close(context.Background())
		require.Error(t, err)
		assertAPIError(t, err, perun.ProtocolFatalError, perun.ErrV2TxTimedOut, txTimedoutError.Error())
		txType := txTimedoutError.TxType
		txID := txTimedoutError.TxID
		txTimeout := ethereumtest.OnChainTxTimeout.String()
		assertErrV2InfoTxTimedOut(t, err.AddInfo(), txType, txID, txTimeout)
	})

	t.Run("finalized_settle_ChainNotReachable", func(t *testing.T) {
		chainURL := ethereumtest.ChainURL
		chainNotReachableError := pclient.ChainNotReachableError{}
		pch, _ := newMockPCh(t, validOpeningBalInfo)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)

		pch.On("Idx").Return(pchannel.Index(peerIdx))
		pch.On("UpdateBy", mock.Anything, mock.Anything).Return(nil)
		pch.On("Register", mock.Anything).Return(chainNotReachableError)

		_, err := ch.Close(context.Background())
		require.Error(t, err)
		wantMessage := chainNotReachableError.Error()
		assertAPIError(t, err, perun.ProtocolFatalError, perun.ErrV2ChainNotReachable, wantMessage)
		assertErrV2InfoChainNotReachable(t, err.AddInfo(), chainURL)
	})
}

func Test_HandleWatcherReturned(t *testing.T) {
	peers := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peers[0].Alias},
		Bal:      []string{"1", "2"},
	}

	t.Run("happy_openCh_dropNotif", func(t *testing.T) {
		pch, watcherSignal := newMockPCh(t, validOpeningBalInfo)
		pch.On("Close").Return(nil)
		_ = session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		watcherSignal <- time.Now()
	})

	t.Run("happy_closedCh_dropNotif", func(t *testing.T) {
		pch, watcherSignal := newMockPCh(t, validOpeningBalInfo)
		pch.On("Close").Return(nil)
		_ = session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, false)
		watcherSignal <- time.Now()
	})

	t.Run("happy_openCh_hasSub_WatchNoError", func(t *testing.T) {
		pch, watcherSignal := newMockPCh(t, validOpeningBalInfo)
		pch.On("Close").Return(nil)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		notifs := make([]perun.ChUpdateNotif, 0, 2)
		notifer := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		require.NoError(t, ch.SubChUpdates(notifer))
		watcherSignal <- time.Now()
	})

	t.Run("happy_openCh_hasSub_WatchError", func(t *testing.T) {
		pch := &mocks.Channel{}
		pch.On("ID").Return([32]byte{0, 1, 2})
		pch.On("State").Return(makeState(t, validOpeningBalInfo, false))
		watcherSignal := make(chan time.Time)
		pch.On("Watch", mock.Anything).WaitUntil(watcherSignal).Return(assert.AnError)

		pch.On("Close").Return(nil)
		ch := session.NewChForTest(pch, currency.ETH, validOpeningBalInfo.Parts, responseTimeout, challengeDurSecs, true)
		notifs := make([]perun.ChUpdateNotif, 0, 2)
		notifer := func(notif perun.ChUpdateNotif) {
			notifs = append(notifs, notif)
		}
		require.NoError(t, ch.SubChUpdates(notifer))
		watcherSignal <- time.Now()
	})
}
