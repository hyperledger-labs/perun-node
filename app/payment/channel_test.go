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

		channelAPI := &mocks.ChannelAPI{}
		channelAPI.On("GetInfo").Return(chInfo)
		channelAPI.On("SendChUpdate", context.Background(), mock.MatchedBy(func(gotUpdater perun.StateUpdater) bool {
			updater = gotUpdater
			return true
		})).Return(nil)

		gotErr := payment.SendPayChUpdate(context.Background(), channelAPI, peerAlias, amountToSend)
		require.NoError(t, gotErr)
		require.NotNil(t, updater)

		stateCopy := chInfo.State.Clone()
		updater(stateCopy)
		assert.Equal(t, chUpdateNotif.Update.State.Allocation.Balances, stateCopy.Allocation.Balances)
	})

	t.Run("happy_requestPayment", func(t *testing.T) {
		channelAPI := &mocks.ChannelAPI{}
		channelAPI.On("GetInfo").Return(chInfo)
		channelAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(nil)

		gotErr := payment.SendPayChUpdate(context.Background(), channelAPI, perun.OwnAlias, amountToSend)
		require.NoError(t, gotErr)
	})

	t.Run("error_InsufficientBalance", func(t *testing.T) {
		channelAPI := &mocks.ChannelAPI{}
		channelAPI.On("GetInfo").Return(chInfo)
		channelAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(nil)

		amount := "10" // This amount is greater than the channel balance of "1"
		gotErr := payment.SendPayChUpdate(context.Background(), channelAPI, peerAlias, amount)
		require.True(t, errors.Is(gotErr, perun.ErrInsufficientBal))
	})

	t.Run("error_InvalidAmount", func(t *testing.T) {
		channelAPI := &mocks.ChannelAPI{}
		channelAPI.On("GetInfo").Return(chInfo)
		channelAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(nil)

		invalidAmount := "abc"
		gotErr := payment.SendPayChUpdate(context.Background(), channelAPI, peerAlias, invalidAmount)
		require.True(t, errors.Is(gotErr, perun.ErrInvalidAmount))
	})

	t.Run("error_InvalidPayee", func(t *testing.T) {
		channelAPI := &mocks.ChannelAPI{}
		channelAPI.On("GetInfo").Return(chInfo)
		channelAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(nil)

		invalidPayee := "invalid-payee"
		gotErr := payment.SendPayChUpdate(context.Background(), channelAPI, invalidPayee, amountToSend)
		require.True(t, errors.Is(gotErr, perun.ErrInvalidPayee))
	})

	t.Run("error_SendChUpdate", func(t *testing.T) {
		channelAPI := &mocks.ChannelAPI{}
		channelAPI.On("GetInfo").Return(chInfo)
		channelAPI.On("SendChUpdate", context.Background(), mock.Anything).Return(assert.AnError)

		gotErr := payment.SendPayChUpdate(context.Background(), channelAPI, peerAlias, amountToSend)
		require.Error(t, gotErr)
	})
}

func Test_GetBalInfo(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		channelAPI := &mocks.ChannelAPI{}
		channelAPI.On("GetInfo").Return(chInfo)

		gotBalInfo := payment.GetBalInfo(channelAPI)
		assert.Equal(t, wantBalInfo, gotBalInfo)
	})
}
