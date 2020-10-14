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
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/contacts/contactstest"
	"github.com/hyperledger-labs/perun-node/session"
)

// NewConfigFileT is the test friendly version of NewConfigFile.
// It uses the passed testing.T to handle the errors and registers the cleanup functions on it.
func NewConfigFileT(t *testing.T, config session.Config) string {
	configFile, err := NewConfigFile(config)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err = os.Remove(configFile); err != nil {
			t.Log("Error in test cleanup: removing file - " + configFile)
		}
	})
	return configFile
}

// NewConfigFile creates a temporary file containing the given session configuration and
// returns the path to it. It also registers a cleanup function on the passed test handler.
func NewConfigFile(config interface{}) (string, error) {
	tempFile, err := ioutil.TempFile("", "*.yaml")
	if err != nil {
		return "", errors.Wrap(err, "creating temp file for config")
	}
	encoder := yaml.NewEncoder(tempFile)
	if err := encoder.Encode(config); err != nil {
		tempFile.Close()           // nolint: errcheck
		os.Remove(tempFile.Name()) // nolint: errcheck
		return "", errors.Wrap(err, "encoding config")
	}
	if err := encoder.Close(); err != nil {
		tempFile.Close()           // nolint: errcheck
		os.Remove(tempFile.Name()) // nolint: errcheck
		return "", errors.Wrap(err, "closing encoder")
	}
	return tempFile.Name(), tempFile.Close()
}

// NewConfigT is the test friendly version of NewConfig.
// It uses the passed testing.T to handle the errors and registers the cleanup functions on it.
func NewConfigT(t *testing.T, rng *rand.Rand, contacts ...perun.Peer) session.Config {
	sessionCfg, err := NewConfig(rng, contacts...)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err = os.RemoveAll(sessionCfg.DatabaseDir); err != nil {
			t.Log("Error in test cleanup: removing directory - " + sessionCfg.DatabaseDir)
		}
		// Currently, onChainWallet & offChainWallet use same keystore, so delete once.
		if err = os.RemoveAll(sessionCfg.User.OnChainWallet.KeystorePath); err != nil {
			t.Log("Error in test cleanup: removing directory - " + sessionCfg.User.OnChainWallet.KeystorePath)
		}
		if err = os.Remove(sessionCfg.ContactsURL); err != nil {
			t.Log("Error in test cleanup: removing file - " + sessionCfg.ContactsURL)
		}
	})
	return sessionCfg
}

// NewConfig generates random configuration data for the session using the given prng and contacts.
// A contacts file is created with the given set of peers and path to it is added in the config.
// This function also registers cleanup functions for removing all the temp files and dirs after the test.
//
// This function returns a session config with user on-chain addresses that are funded on blockchain when
// using a particular seed for prng. The first two consecutive calls to this function will return
// funded accounts when using prng := rand.New(rand.NewSource(1729)).
func NewConfig(rng *rand.Rand, contacts ...perun.Peer) (session.Config, error) {
	_, userCfg, err := newUserConfig(rng, 0)
	if err != nil {
		return session.Config{}, errors.WithMessage(err, "new user config")
	}
	adjudicator, asset := ethereumtest.ContractAddrs()
	databaseDir, err := newDatabaseDir()
	if err != nil {
		return session.Config{}, err
	}
	contactsYAMLFile, err := contactstest.NewYAMLFile(contacts...)
	if err != nil {
		return session.Config{}, err
	}

	return session.Config{
		User:             userCfg,
		ChainURL:         ethereumtest.ChainURL,
		Adjudicator:      adjudicator.String(),
		Asset:            asset.String(),
		ChainConnTimeout: 30 * time.Second,
		ResponseTimeout:  10 * time.Second,
		OnChainTxTimeout: 5 * time.Second,
		DatabaseDir:      databaseDir,

		ContactsType: "yaml",
		ContactsURL:  contactsYAMLFile,
	}, nil
}

func newDatabaseDir() (string, error) {
	databaseDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", errors.Wrap(err, "creating temp directory for database")
	}
	return databaseDir, nil
}

// NewUserConfigT is the test friendly version of NewUserConfig.
// It uses the passed testing.T to handle the errors and registers the cleanup functions on it.
func NewUserConfigT(t *testing.T, rng *rand.Rand, n uint) (perun.WalletBackend, session.UserConfig) {
	wb, userCfg, err := NewUserConfig(rng, n)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := os.RemoveAll(userCfg.OnChainWallet.KeystorePath); err != nil {
			t.Log("error in cleanup - ", err)
		}
	})
	return wb, userCfg
}

// NewUserConfig returns a test user configuration with random data generated using the given rng.
// It creates "n" participant accounts to the user.
func NewUserConfig(rng *rand.Rand, n uint) (perun.WalletBackend, session.UserConfig, error) {
	ws, userCfg, err := newUserConfig(rng, 2+n)
	return ws.WalletBackend, userCfg, err
}

func newUserConfig(rng *rand.Rand, n uint) (*ethereumtest.WalletSetup, session.UserConfig, error) {
	ws, err := ethereumtest.NewWalletSetup(rng, 2+n)
	if err != nil {
		return nil, session.UserConfig{}, errors.WithMessage(err, "new wallet setup")
	}
	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, session.UserConfig{}, errors.Wrap(err, "acquiring free port")
	}

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
	return ws, cfg, nil
}
