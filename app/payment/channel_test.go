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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/app/payment"
	"github.com/hyperledger-labs/perun-node/internal/mocks"
)

func Test_SendPayChUpdate(t *testing.T) {
	t.Run("happy_sendPayment", func(t *testing.T) {
		var updater perun.StateUpdater
		chAPI := &mocks.ChAPI{}
		chAPI.On("GetChInfo").Return(openedChInfo)
		chAPI.On("SendChUpdate", context.Background(), mock.MatchedBy(func(gotUpdater perun.StateUpdater) bool {
			updater = gotUpdater
			return true
		})).Return(nil)

		gotErr := payment.SendPayChUpdate(context.Background(), chAPI, peerAlias, amountToSend)
		require.NoError(t, gotErr)
		require.NotNil(t, updater)

		// TODO: Now that State is not available, how to test the updater function ?
		// stateCopy := chInfo.State.Clone()
		// updater(stateCopy)
		// assert.Equal(t, chUpdateNotif.ProposedChInfo.State.Balances, stateCopy.Allocation.Balances)
	})

	t.Run("happy_requestPayment", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("GetChInfo").Return(openedChInfo)
		chAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(nil)

		gotErr := payment.SendPayChUpdate(context.Background(), chAPI, perun.OwnAlias, amountToSend)
		require.NoError(t, gotErr)
	})

	t.Run("error_InvalidAmount", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("GetChInfo").Return(openedChInfo)
		chAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(nil)

		invalidAmount := "abc"
		gotErr := payment.SendPayChUpdate(context.Background(), chAPI, peerAlias, invalidAmount)
		require.True(t, errors.Is(gotErr, perun.ErrInvalidAmount))
	})

	t.Run("error_InvalidPayee", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("GetChInfo").Return(openedChInfo)
		chAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(nil)

		invalidPayee := "invalid-payee"
		gotErr := payment.SendPayChUpdate(context.Background(), chAPI, invalidPayee, amountToSend)
		require.True(t, errors.Is(gotErr, perun.ErrInvalidPayee))
	})

	t.Run("error_SendChUpdate", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("GetChInfo").Return(openedChInfo)
		chAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(assert.AnError)

		gotErr := payment.SendPayChUpdate(context.Background(), chAPI, peerAlias, amountToSend)
		require.Error(t, gotErr)
		t.Log(gotErr)
	})
}

func Test_GetBalInfo(t *testing.T) {
	t.Run("happy1", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("GetChInfo").Return(openedChInfo)

		gotBalInfo := payment.GetBalInfo(chAPI)
		assert.Equal(t, openingBalInfo, gotBalInfo)
	})
	t.Run("happy2", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("GetChInfo").Return(updatedChInfo)

		gotBalInfo := payment.GetBalInfo(chAPI)
		assert.Equal(t, updatedBalInfo, gotBalInfo)
	})
}

func Test_SubPayChUpdates(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		var notifier perun.ChUpdateNotifier
		var notif payment.PayChUpdateNotif
		dummyNotifier := func(gotNotif payment.PayChUpdateNotif) {
			notif = gotNotif
		}
		chAPI := &mocks.ChAPI{}
		chAPI.On("SubChUpdates", mock.MatchedBy(func(gotNotifier perun.ChUpdateNotifier) bool {
			notifier = gotNotifier
			return true
		})).Return(nil)

		gotErr := payment.SubPayChUpdates(chAPI, dummyNotifier)
		assert.NoError(t, gotErr)
		require.NotNil(t, notifier)

		// Test the notifier function, that interprets the notification for payment app.
		require.NotNil(t, notifier)
		notifier(chUpdateNotif)
		require.Equal(t, wantPayChUpdateNotif, notif)
	})
	t.Run("error", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("SubChUpdates", mock.Anything).Return(assert.AnError)

		dummyNotifier := func(notif payment.PayChUpdateNotif) {}
		gotErr := payment.SubPayChUpdates(chAPI, dummyNotifier)
		assert.Error(t, gotErr)
		t.Log(gotErr)
	})
}

func Test_UnsubPayChUpdates(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("UnsubChUpdates").Return(nil)

		gotErr := payment.UnsubPayChUpdates(chAPI)
		assert.NoError(t, gotErr)
	})
	t.Run("error", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("UnsubChUpdates").Return(assert.AnError)

		gotErr := payment.UnsubPayChUpdates(chAPI)
		assert.Error(t, gotErr)
	})
}

// nolint: dupl	// not duplicate of Test_RespondPayChProposal.
func Test_RespondPayChUpdate(t *testing.T) {
	updateID := "update-id-1"
	t.Run("happy_accept", func(t *testing.T) {
		accept := true
		chAPI := &mocks.ChAPI{}
		chAPI.On("RespondChUpdate", context.Background(), updateID, accept).Return(nil)

		gotErr := payment.RespondPayChUpdate(context.Background(), chAPI, updateID, accept)
		assert.NoError(t, gotErr)
	})
	t.Run("happy_reject", func(t *testing.T) {
		accept := false
		chAPI := &mocks.ChAPI{}
		chAPI.On("RespondChUpdate", context.Background(), updateID, accept).Return(nil)

		gotErr := payment.RespondPayChUpdate(context.Background(), chAPI, updateID, accept)
		assert.NoError(t, gotErr)
	})
	t.Run("error", func(t *testing.T) {
		accept := true
		chAPI := &mocks.ChAPI{}
		chAPI.On("RespondChUpdate", context.Background(), updateID, accept).Return(assert.AnError)

		gotErr := payment.RespondPayChUpdate(context.Background(), chAPI, updateID, accept)
		assert.Error(t, gotErr)
		t.Log(gotErr)
	})
}

func Test_ClosePayCh(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("Close", context.Background()).Return(updatedChInfo, nil)

		gotPayChInfo, err := payment.ClosePayCh(context.Background(), chAPI)
		require.NoError(t, err)
		assert.Equal(t, wantUpdatedPayChInfo, gotPayChInfo)
	})
	t.Run("error", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("Close", context.Background()).Return(updatedChInfo, assert.AnError)

		_, gotErr := payment.ClosePayCh(context.Background(), chAPI)
		require.Error(t, gotErr)
		t.Log(gotErr)
	})
}
