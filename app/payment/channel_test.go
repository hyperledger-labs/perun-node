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

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/app/payment"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/internal/mocks"
	"github.com/hyperledger-labs/perun-node/peruntest"
)

func Test_SendPayChUpdate(t *testing.T) {
	currencies := currency.NewRegistry()
	//nolint:errcheck	// Safe to ignore the error, as it is first register after init.
	ethCurrency, _ := currencies.Register(currency.ETHSymbol, currency.ETHMaxDecimals)

	// Returns a mock with API calls set up for currency and parts.
	newChAPIMock := func() *mocks.ChAPI {
		chAPI := &mocks.ChAPI{}
		chAPI.On("Currencies").Return([]perun.Currency{ethCurrency})
		chAPI.On("Currency", currency.ETHSymbol).Return(0, ethCurrency, true)
		chAPI.On("Parts").Return(parts)
		return chAPI
	}

	t.Run("happy_sendPayment", func(t *testing.T) {
		var updater perun.StateUpdater
		chAPI := newChAPIMock()
		chAPI.On("SendChUpdate", context.Background(), mock.MatchedBy(func(gotUpdater perun.StateUpdater) bool {
			updater = gotUpdater
			return true
		})).Return(updatedChInfo, nil)

		payments := []payment.Payment{
			makePayment(currency.ETHSymbol, peerAlias, amountToSend),
		}
		gotPayChInfo, gotErr := payment.SendPayChUpdate(context.Background(), chAPI, payments)
		require.NoError(t, gotErr)
		assert.Equal(t, wantUpdatedPayChInfo, gotPayChInfo)
		require.NotNil(t, updater)

		// TODO: Now that State is not available, how to test the updater function ?
		// stateCopy := chInfo.State.Clone()
		// updater(stateCopy)
		// assert.Equal(t, chUpdateNotif.ProposedChInfo.State.Balances, stateCopy.Allocation.Balances)
	})

	t.Run("happy_requestPayment", func(t *testing.T) {
		chAPI := newChAPIMock()
		chAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(updatedChInfo, nil)
		payments := []payment.Payment{
			makePayment(currency.ETHSymbol, perun.OwnAlias, amountToSend),
		}

		gotPayChInfo, gotErr := payment.SendPayChUpdate(context.Background(), chAPI, payments)
		require.NoError(t, gotErr)
		require.Equal(t, wantUpdatedPayChInfo, gotPayChInfo)
	})

	t.Run("error_InvalidAmount", func(t *testing.T) {
		chAPI := newChAPIMock()
		chAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(perun.ChInfo{}, nil)

		invalidAmount := "abc"
		payments := []payment.Payment{
			makePayment(currency.ETHSymbol, perun.OwnAlias, invalidAmount),
		}
		_, gotErr := payment.SendPayChUpdate(context.Background(), chAPI, payments)
		peruntest.AssertAPIError(t, gotErr, perun.ClientError, perun.ErrInvalidArgument, payment.ErrInvalidAmount.Error())
		peruntest.AssertErrInfoInvalidArgument(t, gotErr.AddInfo(), perun.ArgNameAmount, invalidAmount)
	})

	t.Run("error_InvalidPayee", func(t *testing.T) {
		chAPI := newChAPIMock()
		chAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(perun.ChInfo{}, nil)

		invalidPayee := "invalid-payee"
		payments := []payment.Payment{
			makePayment(currency.ETHSymbol, invalidPayee, amountToSend),
		}
		_, gotErr := payment.SendPayChUpdate(context.Background(), chAPI, payments)
		peruntest.AssertAPIError(t, gotErr, perun.ClientError, perun.ErrInvalidArgument, payment.ErrInvalidPayee.Error())
		peruntest.AssertErrInfoInvalidArgument(t, gotErr.AddInfo(), "payee", invalidPayee)
	})

	t.Run("error_SendChUpdate", func(t *testing.T) {
		chAPI := newChAPIMock()
		chAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(
			perun.ChInfo{}, perun.NewAPIErrUnknownInternal(assert.AnError))
		payments := []payment.Payment{
			makePayment(currency.ETHSymbol, perun.OwnAlias, amountToSend),
		}

		_, gotErr := payment.SendPayChUpdate(context.Background(), chAPI, payments)
		require.Error(t, gotErr)
		t.Log(gotErr)
	})
}

func Test_GetPayChInfo(t *testing.T) {
	t.Run("happy1", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("GetChInfo").Return(openedChInfo)

		gotPayChInfo := payment.GetPayChInfo(chAPI)
		assert.Equal(t, wantOpenedPayChInfo, gotPayChInfo)
	})
	t.Run("happy2", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("GetChInfo").Return(updatedChInfo)

		gotPayChInfo := payment.GetPayChInfo(chAPI)
		assert.Equal(t, wantUpdatedPayChInfo, gotPayChInfo)
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
		t.Run("notifier_typeOpen", func(t *testing.T) {
			notifier(chUpdateNotif)
			require.Equal(t, wantPayChUpdateNotif, notif)
		})
		t.Run("notifier_typeFinal", func(t *testing.T) {
			chUpdateNotifFinal := chUpdateNotif
			chUpdateNotifFinal.Type = perun.ChUpdateTypeFinal
			wantPayChUpdateNotifFinal := wantPayChUpdateNotif
			wantPayChUpdateNotifFinal.Type = perun.ChUpdateTypeFinal

			notifier(chUpdateNotifFinal)
			require.Equal(t, wantPayChUpdateNotifFinal, notif)
		})
		t.Run("notifier_typeClosed", func(t *testing.T) {
			chUpdateNotifClosed := chUpdateNotif
			chUpdateNotifClosed.Type = perun.ChUpdateTypeClosed
			chUpdateNotifClosed.CurrChInfo = chUpdateNotif.ProposedChInfo
			wantPayChUpdateNotifClosed := wantPayChUpdateNotif
			wantPayChUpdateNotifClosed.Type = perun.ChUpdateTypeClosed

			notifier(chUpdateNotifClosed)
			require.Equal(t, wantPayChUpdateNotifClosed, notif)
		})
		t.Run("notifier_typeClosedWithError", func(t *testing.T) {
			chUpdateNotifClosed := chUpdateNotif
			chUpdateNotifClosed.Type = perun.ChUpdateTypeClosed
			chUpdateNotifClosed.CurrChInfo = chUpdateNotif.ProposedChInfo
			chUpdateNotifClosed.Error = perun.NewAPIErrUnknownInternal(assert.AnError)
			wantPayChUpdateNotifClosed := wantPayChUpdateNotif
			wantPayChUpdateNotifClosed.Type = perun.ChUpdateTypeClosed
			wantPayChUpdateNotifClosed.Error = perun.NewAPIErrUnknownInternal(assert.AnError)

			notifier(chUpdateNotifClosed)
			require.Equal(t, wantPayChUpdateNotifClosed, notif)
		})
	})
	t.Run("error", func(t *testing.T) {
		chAPI := &mocks.ChAPI{}
		chAPI.On("SubChUpdates", mock.Anything).Return(perun.NewAPIErrUnknownInternal(assert.AnError))

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
		chAPI.On("UnsubChUpdates").Return(perun.NewAPIErrUnknownInternal(assert.AnError))

		gotErr := payment.UnsubPayChUpdates(chAPI)
		assert.Error(t, gotErr)
	})
}

func Test_RespondPayChUpdate(t *testing.T) {
	updateID := "update-id-1"
	t.Run("happy_accept", func(t *testing.T) {
		accept := true
		chAPI := &mocks.ChAPI{}
		chAPI.On("RespondChUpdate", context.Background(), updateID, accept).Return(updatedChInfo, nil)

		gotUpdatedPayChInfo, gotErr := payment.RespondPayChUpdate(context.Background(), chAPI, updateID, accept)
		assert.NoError(t, gotErr)
		assert.Equal(t, wantUpdatedPayChInfo, gotUpdatedPayChInfo)
	})
	t.Run("happy_reject", func(t *testing.T) {
		accept := false
		chAPI := &mocks.ChAPI{}
		chAPI.On("RespondChUpdate", context.Background(), updateID, accept).Return(perun.ChInfo{}, nil)

		_, gotErr := payment.RespondPayChUpdate(context.Background(), chAPI, updateID, accept)
		assert.NoError(t, gotErr)
	})
	t.Run("error", func(t *testing.T) {
		accept := true
		chAPI := &mocks.ChAPI{}
		err := perun.NewAPIErrUnknownInternal(assert.AnError)
		chAPI.On("RespondChUpdate", context.Background(), updateID, accept).Return(perun.ChInfo{}, err)

		_, gotErr := payment.RespondPayChUpdate(context.Background(), chAPI, updateID, accept)
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
		chAPI.On("Close", context.Background()).Return(updatedChInfo, perun.NewAPIErrUnknownInternal(assert.AnError))

		_, gotErr := payment.ClosePayCh(context.Background(), chAPI)
		require.Error(t, gotErr)
		t.Log(gotErr)
	})
}

//nolint:unparam
func makePayment(currency, payee, amount string) payment.Payment {
	return payment.Payment{
		Currency: currency,
		Payee:    payee,
		Amount:   amount,
	}
}
