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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	ppayment "perun.network/go-perun/apps/payment"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/session"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

// init() initializes the payment app in go-perun.
func init() {
	wb := ethereum.NewWalletBackend()
	emptyAddr, err := wb.ParseAddr("0x0")
	if err != nil {
		panic("Error parsing zero address for app payment def: " + err.Error())
	}
	ppayment.SetAppDef(emptyAddr) // dummy app def.
}

// This test includes all methods on SessionAPI and ChannelAPI.
func Test_Integ_Role(t *testing.T) {
	prng := rand.New(rand.NewSource(1729))

	aliceAlias, bobAlias := "alice", "bob"

	// Start with empty contacts.
	aliceCfg := sessiontest.NewConfig(t, prng)
	bobCfg := sessiontest.NewConfig(t, prng)

	alice, err := session.New(aliceCfg)
	require.NoErrorf(t, err, "initializing alice session")
	t.Logf("alice session id: %s\n", alice.ID())

	bob, err := session.New(bobCfg)
	require.NoErrorf(t, err, "initializing bob session")
	t.Logf("bob session id: %s\n", bob.ID())

	var aliceContact, bobContact perun.Peer
	t.Run("GetContact", func(t *testing.T) {
		t.Run("happy", func(t *testing.T) {
			aliceContact, err = alice.GetContact(perun.OwnAlias)
			require.NoErrorf(t, err, "Alice: GetContact")
			aliceContact.Alias = aliceAlias

			bobContact, err = bob.GetContact(perun.OwnAlias)
			require.NoErrorf(t, err, "Bob: GetContact")
			bobContact.Alias = bobAlias
		})
		t.Run("missing", func(t *testing.T) {
			_, err = alice.GetContact("random alias")
			assert.Errorf(t, err, "Alice: GetContact")
			t.Log(err)
		})
	})

	t.Run("AddContact", func(t *testing.T) {
		t.Run("happy", func(t *testing.T) {
			err = alice.AddContact(bobContact)
			require.NoErrorf(t, err, "Alice: AddContact")

			err = bob.AddContact(aliceContact)
			require.NoErrorf(t, err, "Bob: GetContact")
		})
		t.Run("already_exists", func(t *testing.T) {
			// Try to add bob contact again
			err = alice.AddContact(bobContact)
			assert.Errorf(t, err, "Alice: AddContact")
			t.Log(err)
		})
	})

	const challengeDurSecs uint64 = 10
	// wg := &sync.WaitGroup{}
	ctx := context.Background()

	t.Run("OpenCh", func(t *testing.T) {
		// Propose Channel by alice.
		bals := make(map[string]string)
		bals[perun.OwnAlias] = "1"
		bals[bobAlias] = "2"
		balInfo := perun.BalInfo{
			Currency: currency.ETH,
			Bals:     bals,
		}
		app := perun.App{
			Def:  ppayment.AppDef(),
			Data: &ppayment.NoData{},
		}
		aliceCh1Info, err := alice.OpenCh(ctx, bobAlias, balInfo, app, challengeDurSecs)
		require.NoErrorf(t, err, "alice opening channel with bob")
		t.Log(aliceCh1Info)
	})
}
