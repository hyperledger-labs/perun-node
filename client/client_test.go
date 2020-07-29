// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/dst-go
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

package client_test

import (
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/direct-state-transfer/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/direct-state-transfer/perun-node/client"
	"github.com/direct-state-transfer/perun-node/comm/tcp"
	"github.com/direct-state-transfer/perun-node/session/sessiontest"
)

// Currently the integration setup is a quick hack requiring the user to manually start ganache-cli node. Command:
//
// ganache-cli --account="0x1fedd636dbc7e8d41a0622a2040b86fea8842cef9d4aa4c582aad00465b7acff,10000000000000000000"
//
// The account in the command corresponds to on-chain account of user when seeding the rand source with 1729.
// Hence DO NOT CHANGE THE RAND SEED for integration tests in this package.
//
// The contracts will be deployed only during the first run of tests and will be resused in subsequent runs. This
// saves ~0.3s of setup time in each run. Hence when running tests on development machine, START THE NODE ONLY ONCE.

func Test_Integ_NewEthereumPaymentClient(t *testing.T) {
	prng := rand.New(rand.NewSource(1729))
	_, user := sessiontest.NewTestUser(t, prng, 0)
	adjudicator, asset := setup(t, user.OnChain)

	cfg := client.Config{
		Chain: client.ChainConfig{
			Adjudicator: adjudicator.String(),
			Asset:       asset.String(),
			URL:         testChainURL,
			ConnTimeout: 10 * time.Second,
		},
		PeerReconnTimeout: 0,
	}
	// TODO: (mano) Test if handle and lister are running as expected.

	t.Run("happy", func(t *testing.T) {
		cfg.DatabaseDir = newDatabaseDir(t) // start with empty persistence dir each time.
		_, err := client.NewEthereumPaymentClient(cfg, user, tcp.NewTCPBackend(5*time.Second))
		assert.NoError(t, err)
	})

	t.Run("err_invalid_chain_url", func(t *testing.T) {
		invalidCfg := cfg
		invalidCfg.DatabaseDir = newDatabaseDir(t) // start with empty persistence dir each time.
		invalidCfg.Chain.URL = "invalid-url"

		_, err := client.NewEthereumPaymentClient(invalidCfg, user, tcp.NewTCPBackend(5*time.Second))
		t.Log(err)
		assert.Error(t, err)
	})

	t.Run("err_malformed_asset_addr", func(t *testing.T) {
		invalidCfg := cfg
		invalidCfg.DatabaseDir = newDatabaseDir(t) // start with empty persistence dir each time.
		invalidCfg.Chain.Asset = "invalid-addr"

		_, err := client.NewEthereumPaymentClient(invalidCfg, user, tcp.NewTCPBackend(5*time.Second))
		t.Log(err)
		assert.Error(t, err)
	})

	t.Run("err_malformed_adj_addr", func(t *testing.T) {
		invalidCfg := cfg
		invalidCfg.DatabaseDir = newDatabaseDir(t) // start with empty persistence dir each time.
		invalidCfg.Chain.Adjudicator = "invalid-addr"

		_, err := client.NewEthereumPaymentClient(invalidCfg, user, tcp.NewTCPBackend(5*time.Second))
		t.Log(err)
		assert.Error(t, err)
	})

	t.Run("err_invalid_adjudicator", func(t *testing.T) {
		invalidCfg := cfg
		invalidCfg.DatabaseDir = newDatabaseDir(t) // start with empty persistence dir each time.
		randomAddr := ethereumtest.NewRandomAddress(prng)
		invalidCfg.Chain.Adjudicator = randomAddr.String()

		_, err := client.NewEthereumPaymentClient(invalidCfg, user, tcp.NewTCPBackend(5*time.Second))
		t.Log(err)
		assert.Error(t, err)
	})

	t.Run("err_invalid_asset", func(t *testing.T) {
		invalidCfg := cfg
		invalidCfg.DatabaseDir = newDatabaseDir(t) // start with empty persistence dir each time.
		randomAddr := ethereumtest.NewRandomAddress(prng)
		invalidCfg.Chain.Asset = randomAddr.String()

		_, err := client.NewEthereumPaymentClient(invalidCfg, user, tcp.NewTCPBackend(5*time.Second))
		t.Log(err)
		assert.Error(t, err)
	})

	t.Run("err_invalid_on_chain_password", func(t *testing.T) {
		cfg.DatabaseDir = newDatabaseDir(t) // start with empty persistence dir each time.
		invalidUser := user
		ws := ethereumtest.NewWalletSetup(t, prng, 0)
		invalidUser.OnChain.Wallet = ws.Wallet
		invalidUser.OnChain.Keystore = ws.KeystorePath
		invalidUser.OnChain.Password = "invalid-password"
		_, err := client.NewEthereumPaymentClient(cfg, invalidUser, tcp.NewTCPBackend(5*time.Second))
		t.Log(err)
		assert.Error(t, err)
	})

	t.Run("err_invalid_off_chain_password", func(t *testing.T) {
		cfg.DatabaseDir = newDatabaseDir(t) // start with empty persistence dir each time.
		invalidUser := user
		ws := ethereumtest.NewWalletSetup(t, prng, 0)
		invalidUser.OffChain.Wallet = ws.Wallet
		invalidUser.OffChain.Keystore = ws.KeystorePath
		invalidUser.OffChain.Password = "invalid-password"
		_, err := client.NewEthereumPaymentClient(cfg, invalidUser, tcp.NewTCPBackend(5*time.Second))
		t.Log(err)
		assert.Error(t, err)
	})

	t.Run("err_invalid_comm_addr", func(t *testing.T) {
		cfg.DatabaseDir = newDatabaseDir(t) // start with empty persistence dir each time.
		invalidUser := user
		invalidUser.CommAddr = "invalid-addr"
		_, err := client.NewEthereumPaymentClient(cfg, invalidUser, tcp.NewTCPBackend(5*time.Second))
		assert.Error(t, err)
	})

	t.Run("err_persistence_path_is_file", func(t *testing.T) {
		emptyFile, err := ioutil.TempFile("", "")
		require.NoError(t, err)
		err = emptyFile.Close()
		require.NoError(t, err)
		t.Cleanup(func() {
			// nolint:govet // err does not shadow the prev declaration as this func will be executed later.
			if err := os.Remove(emptyFile.Name()); err != nil {
				t.Logf("Error in removing the file in test cleanup - %v", err)
			}
		})

		cfgInvalidPeristence := cfg
		cfgInvalidPeristence.DatabaseDir = emptyFile.Name()
		_, err = client.NewEthereumPaymentClient(cfgInvalidPeristence, user, tcp.NewTCPBackend(5*time.Second))
		t.Log(err)
		assert.Error(t, err)
	})

	// TODO: (mano) Faulty Persistence data and Reconnection errors.
}

func newDatabaseDir(t *testing.T) (dir string) {
	databaseDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := os.RemoveAll(databaseDir); err != nil {
			t.Logf("Error in removing the file in test cleanup - %v", err)
		}
	})
	return databaseDir
}
