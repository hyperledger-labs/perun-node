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

package client_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger-labs/perun-node/client"
	"github.com/hyperledger-labs/perun-node/internal/mocks"
)

func Test_Client_Close(t *testing.T) {
	// happy path test is covered in integration test, as internal components of
	// the client should be initialized.
	t.Run("err_channelClient_Err", func(t *testing.T) {
		chClient := &mocks.ChClient{}
		msgBus := &mocks.WireBus{}
		dbConnCloser := &mocks.Closer{}
		Client := client.NewClientForTest(chClient, msgBus, nil, dbConnCloser)

		chClient.On("Close").Return(errors.New("error for test"))
		msgBus.On("Close").Return(nil)
		assert.Error(t, Client.Close())
	})

	t.Run("err_wireBus_Err", func(t *testing.T) {
		chClient := &mocks.ChClient{}
		msgBus := &mocks.WireBus{}
		dbConnCloser := &mocks.Closer{}
		Client := client.NewClientForTest(chClient, msgBus, nil, dbConnCloser)

		chClient.On("Close").Return(nil)
		msgBus.On("Close").Return(errors.New("error for test"))
		assert.Error(t, Client.Close())
	})
}

func Test_Client_Register(t *testing.T) {
	// happy path test is covered in integration test, as internal components of
	// the client should be initialized.
	t.Run("happy", func(t *testing.T) {
		registerer := &mocks.Registerer{}
		dbConnCloser := &mocks.Closer{}
		Client := client.NewClientForTest(nil, nil, registerer, dbConnCloser)
		registerer.On("Register", nil, "").Return()
		Client.Register(nil, "")
	})
}
