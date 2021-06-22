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
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	copyutil "github.com/otiai10/copy"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/comm/tcp"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/idprovider/idprovidertest"
	"github.com/hyperledger-labs/perun-node/peruntest"
	"github.com/hyperledger-labs/perun-node/session"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

func Test_Integ_New(t *testing.T) {
	peerIDs := newPeerIDs(t, uint(2))
	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	cfg := sessiontest.NewConfigT(t, prng, peerIDs...)

	contracts := ethereumtest.SetupContractsT(t,
		cfg.ChainURL, cfg.ChainID, cfg.OnChainTxTimeout, false)
	currencies := currency.NewRegistry()
	_, err := currencies.Register(currency.ETHSymbol, currency.ETHMaxDecimals)
	require.NoError(t, err)
	// TODO: (mano) Test if handle and listener are running as expected.

	t.Run("happy", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)

		// Listener will start listening on this port.
		// Use a different port number to not affect other tests.
		port, err := freeport.GetFreePort()
		require.NoError(t, err)
		cfgCopy.User.CommAddr = fmt.Sprintf("127.0.0.1:%d", port)

		sess, err := session.New(cfgCopy, currencies, contracts)
		require.NoError(t, err)
		assert.NotNil(t, sess)
	})
	t.Run("invalidConfig_databaseDir_alreadyInUse", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)

		// Listener will start listening on this port.
		// Use a different port number to not affect other tests.
		port, err := freeport.GetFreePort()
		require.NoError(t, err)
		cfgCopy.User.CommAddr = fmt.Sprintf("127.0.0.1:%d", port)

		_, apiErr := session.New(cfgCopy, currencies, contracts)
		require.NoError(t, apiErr)
		// Start a session so that persistence directory is already in use,
		// Keep the database directory as same,
		// Change the port number (so that new listener can be stared without error.
		port, err = freeport.GetFreePort()
		require.NoError(t, err)
		cfgCopy.User.CommAddr = fmt.Sprintf("127.0.0.1:%d", port)

		_, apiErr = session.New(cfgCopy, currencies, contracts)
		require.Error(t, apiErr)
		peruntest.AssertAPIError(t, apiErr, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, apiErr.AddInfo(), "databaseDir", cfgCopy.DatabaseDir)
	})
	t.Run("invalidConfig_chainURL", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)
		cfgCopy.ChainURL = "invalid-url"
		_, err := session.New(cfgCopy, currencies, contracts)
		require.Error(t, err)
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, err.AddInfo(), "chainURL", cfgCopy.ChainURL)
	})
	t.Run("invalidConfig_chainURL_chainConnTimeout", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)
		cfgCopy.ChainConnTimeout = 0
		_, err := session.New(cfgCopy, currencies, contracts)
		require.Error(t, err)
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, err.AddInfo(), "chainURL", cfgCopy.ChainURL)
	})

	t.Run("invalidConfig_onChainAddr", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)
		cfgCopy.User.OnChainAddr = "invalid-addr" //nolint: goconst	// it's okay to repeat this phrase.
		_, err := session.New(cfgCopy, currencies, contracts)
		require.Error(t, err)
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, err.AddInfo(), "onChainAddr", cfgCopy.User.OnChainAddr)
	})
	t.Run("invalidConfig_offChainAddr", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)
		cfgCopy.User.OffChainAddr = "invalid-addr"
		_, err := session.New(cfgCopy, currencies, contracts)
		require.Error(t, err)
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, err.AddInfo(), "offChainAddr", cfgCopy.User.OffChainAddr)
	})
	t.Run("invalidConfig_onChainWallet_password", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)
		cfgCopy.User.OnChainWallet.Password = "invalid-password"
		wantValue := fmt.Sprintf("%s, %s", cfgCopy.User.OnChainWallet.KeystorePath, cfgCopy.User.OnChainWallet.Password)
		_, err := session.New(cfgCopy, currencies, contracts)
		require.Error(t, err)
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, err.AddInfo(), "onChainWallet", wantValue)
	})
	t.Run("invalidConfig_offChainWallet_password", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)
		cfgCopy.User.OffChainWallet.Password = "invalid-password"
		wantValue := fmt.Sprintf("%s, %s", cfgCopy.User.OffChainWallet.KeystorePath, cfgCopy.User.OffChainWallet.Password)
		_, err := session.New(cfgCopy, currencies, contracts)
		require.Error(t, err)
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, err.AddInfo(), "offChainWallet", wantValue)
	})

	t.Run("invalidConfig_commType", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)
		cfgCopy.User.CommType = "unsupported"
		_, err := session.New(cfgCopy, currencies, contracts)
		require.Error(t, err)
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, err.AddInfo(), "commType", cfgCopy.User.CommType)
	})
	t.Run("invalidConfig_commAddr", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)
		cfgCopy.User.CommAddr = "invalid-addr"
		_, err := session.New(cfgCopy, currencies, contracts)
		require.Error(t, err)
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, err.AddInfo(), "commAddr", cfgCopy.User.CommAddr)
	})
	t.Run("invalidConfig_commAddr_portInUse", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)

		// Start listening on the port already so that session.Init fails.
		// Use a different port number to not affect other tests.
		port, err := freeport.GetFreePort()
		require.NoError(t, err)
		cfgCopy.User.CommAddr = fmt.Sprintf("127.0.0.1:%d", port)
		listener, err := tcp.NewTCPBackend(1 * time.Second).NewListener(cfgCopy.User.CommAddr)
		require.NoError(t, err)
		go func() {
			_, _ = listener.Accept() //nolint: errcheck		// no need to check error.
		}()
		defer listener.Close() //nolint: errcheck		// no need to check error.

		_, apiErr := session.New(cfgCopy, currencies, contracts)
		require.Error(t, apiErr)
		peruntest.AssertAPIError(t, apiErr, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, apiErr.AddInfo(), "commAddr", cfgCopy.User.CommAddr)
	})

	t.Run("invalidConfig_idProviderType", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)
		cfgCopy.IDProviderType = "unsupported"
		_, err := session.New(cfgCopy, currencies, contracts)
		require.Error(t, err)
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, err.AddInfo(), "idProviderType", cfgCopy.IDProviderType)
	})
	t.Run("invalidConfig_idProviderURL", func(t *testing.T) {
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)
		cfgCopy.IDProviderURL = newCorruptedYAMLFile(t)
		_, err := session.New(cfgCopy, currencies, contracts)
		require.Error(t, err)
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, err.AddInfo(), "idProviderURL", cfgCopy.IDProviderURL)
	})
	t.Run("invalidConfig_idProviderURL_hasEntryForSelf", func(t *testing.T) {
		ownPeer := perun.PeerID{
			Alias: perun.OwnAlias,
		}
		cfgCopy := cfg
		cfgCopy.DatabaseDir = newDatabaseDir(t)
		cfgCopy.IDProviderURL = idprovidertest.NewIDProviderT(t, ownPeer)
		_, err := session.New(cfgCopy, currencies, contracts)
		require.Error(t, err)
		peruntest.AssertAPIError(t, err, perun.ClientError, perun.ErrInvalidConfig, "")
		peruntest.AssertErrInfoInvalidConfig(t, err.AddInfo(), "idProviderURL", cfgCopy.IDProviderURL)
		t.Logf("%+v", err)
	})
}

func Test_Integ_Persistence(t *testing.T) {
	// Generate session config for each test case to have unique value of listener address.
	currencies := currency.NewRegistry()
	_, err := currencies.Register(currency.ETHSymbol, currency.ETHMaxDecimals)
	require.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
		cfg := sessiontest.NewConfigT(t, prng)
		// Use idprovider and databaseDir from a session that was persisted already.
		// Copy database directory to tmp before using as it will be modified when reading as well.
		// ID provider file can be used as such.
		cfg.DatabaseDir = copyDirToTmp(t, "../testdata/session/persistence/alice-database")
		cfg.IDProviderURL = "../testdata/session/persistence/alice-idprovider.yaml"
		contracts := ethereumtest.SetupContractsT(t, cfg.ChainURL, cfg.ChainID, cfg.OnChainTxTimeout, false)

		alice, err := session.New(cfg, currencies, contracts)
		require.NoErrorf(t, err, "initializing alice session")
		t.Logf("alice session id: %s\n", alice.ID())
		t.Logf("alice database dir is: %s\n", cfg.DatabaseDir)

		require.Equal(t, 2, len(alice.GetChsInfo()))
	})

	t.Run("happy_drop_unknownPeers", func(t *testing.T) {
		prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
		cfg := sessiontest.NewConfigT(t, prng) // Get a session config with no peerIDs in the ID provider.
		cfg.DatabaseDir = copyDirToTmp(t, "../testdata/session/persistence/alice-database")
		contracts := ethereumtest.SetupContractsT(t, cfg.ChainURL, cfg.ChainID, cfg.OnChainTxTimeout, false)

		_, err := session.New(cfg, currencies, contracts)
		require.NoErrorf(t, err, "initializing alice session")
	})

	t.Run("err_database_init", func(t *testing.T) {
		prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
		cfg := sessiontest.NewConfigT(t, prng) // Get a session config with no peerIDs in the ID provider.
		tempFile, err := ioutil.TempFile("", "")
		require.NoError(t, err)
		tempFile.Close() // nolint:errcheck
		cfg.DatabaseDir = tempFile.Name()
		contracts := ethereumtest.SetupContractsT(t, cfg.ChainURL, cfg.ChainID, cfg.OnChainTxTimeout, false)

		_, err = session.New(cfg, currencies, contracts)
		require.Errorf(t, err, "initializing alice session")
		t.Log(err)
	})
}

func newCorruptedYAMLFile(t *testing.T) string {
	// First line has yaml syntax error (two colons).
	corruptedYaml := `
Alice: alias: Alice
    offchain_address: 0x9282681723920798983380581376586951466585
    comm_address: 127.0.0.1:5751
    comm_type: tcpip
Bob:
    alias: Bob
    offchain_address: 0x3369783337071807248093730889602727505701
    comm_address: 127.0.0.1:5750
    comm_type: tcpip`

	tempFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err = os.Remove(tempFile.Name()); err != nil {
			t.Log("Error in test cleanup: removing file - " + tempFile.Name())
		}
	})
	_, err = tempFile.WriteString(corruptedYaml)
	require.NoErrorf(t, tempFile.Close(), "closing temporary file")
	require.NoError(t, err)
	return tempFile.Name()
}

func copyDirToTmp(t *testing.T, src string) (tempDirName string) {
	var err error
	tempDirName, err = ioutil.TempDir("", "")
	require.NoError(t, err)
	require.NoError(t, copyutil.Copy(src, tempDirName))
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDirName); err != nil {
			t.Logf("Error in removing the file in test cleanup - %v", err)
		}
	})
	return tempDirName
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
