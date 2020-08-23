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

package sessiontest

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/contacts/contactstest"
	"github.com/hyperledger-labs/perun-node/session"
)

// NewConfigFile creates a temporary file containing the given session configuration and
// returns the path to it. It also registers a cleanup function on the passed test handler.
func NewConfigFile(t *testing.T, config session.Config) string {
	tempFile, err := ioutil.TempFile("", "*.yaml")
	defer func() {
		require.NoErrorf(t, tempFile.Close(), "closing temporary file")
	}()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err = os.Remove(tempFile.Name()); err != nil {
			t.Log("Error in test cleanup: removing file - " + tempFile.Name())
		}
	})

	encoder := yaml.NewEncoder(tempFile)
	require.NoErrorf(t, encoder.Encode(config), "encoding config")
	require.NoErrorf(t, encoder.Close(), "closing encoder")
	return tempFile.Name()
}

// NewConfig generates random configuration data for the session using the given prng and contacts.
// A contacts file is created with the given set of peers and path to it is added in the config.
// This function also registers cleanup functions for removing all the temp files and dirs after the test.
//
// This function uses its own prng seed instead of receiving it from the calling test function because the
// specific seed is required to match the accounts funded in on the blockchain.
func NewConfig(t *testing.T, contacts ...perun.Peer) session.Config {
	rng := rand.New(rand.NewSource(1729))
	walletSetup, userCfg := newUserConfig(t, rng, 0)
	cred := perun.Credential{
		Addr:     walletSetup.Accs[0].Address(),
		Wallet:   walletSetup.Wallet,
		Keystore: walletSetup.KeystorePath,
		Password: "",
	}
	chainURL := ethereumtest.ChainURL
	onChainTxTimeout := ethereumtest.OnChainTxTimeout
	adjudicator, asset := ethereumtest.SetupContracts(t, cred, chainURL, onChainTxTimeout)

	return session.Config{
		User:             userCfg,
		ChainURL:         chainURL,
		Adjudicator:      adjudicator.String(),
		Asset:            asset.String(),
		ChainConnTimeout: 30 * time.Second,
		ResponseTimeout:  10 * time.Second,
		OnChainTxTimeout: 5 * time.Second,
		DatabaseDir:      newDatabaseDir(t),

		ContactsType: "yaml",
		ContactsURL:  contactstest.NewYAMLFile(t, contacts...),
	}
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

// NewUserConfig returns a test user configuration with random data generated using the given rng.
// It creates "n" participant accounts to the user.
func NewUserConfig(t *testing.T, rng *rand.Rand, n uint) (perun.WalletBackend, session.UserConfig) {
	ws, userCfg := newUserConfig(t, rng, 2+n)
	return ws.WalletBackend, userCfg
}

func newUserConfig(t *testing.T, rng *rand.Rand, n uint) (*ethereumtest.WalletSetup, session.UserConfig) {
	ws := ethereumtest.NewWalletSetup(t, rng, 2+n)
	port, err := freeport.GetFreePort()
	require.NoError(t, err)

	cfg := session.UserConfig{}
	cfg.Alias = "test-user"
	cfg.OnChainAddr = ws.Accs[0].Address().String()
	cfg.OnChainWallet = session.WalletConfig{
		KeystorePath: ws.KeystorePath,
		Password:     "",
	}
	cfg.OffChainAddr = ws.Accs[1].Address().String()
	cfg.OffChainWallet = session.WalletConfig{
		KeystorePath: ws.KeystorePath,
		Password:     "",
	}
	cfg.PartAddrs = make([]string, n)
	for i := range ws.Accs[2:] {
		cfg.PartAddrs[i] = ws.Accs[i].Address().String()
	}
	cfg.CommType = "tcp"
	cfg.CommAddr = fmt.Sprintf("127.0.0.1:%d", port)
	return ws, cfg
}
