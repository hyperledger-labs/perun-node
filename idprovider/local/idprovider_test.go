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

package local_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/idprovider"
	"github.com/hyperledger-labs/perun-node/idprovider/idprovidertest"
	"github.com/hyperledger-labs/perun-node/idprovider/local"
)

var (
	peer1 = perun.PeerID{
		Alias:              "Alice",
		OffChainAddrString: "0x9282681723920798983380581376586951466585",
		CommType:           "tcpip",
		CommAddr:           "127.0.0.1:5751",
	}
	peer2 = perun.PeerID{
		Alias:              "Bob",
		OffChainAddrString: "0x3369783337071807248093730889602727505701",
		CommType:           "tcpip",
		CommAddr:           "127.0.0.1:5750",
	}
	peer3 = perun.PeerID{
		Alias:              "Tom",
		OffChainAddrString: "0x7187308896023072480933697833370727318468",
		CommType:           "tcpip",
		CommAddr:           "127.0.0.1:5753",
	}

	walletBackend = ethereum.NewWalletBackend()
)

func init() {
	var err error
	peer1.OffChainAddr, err = walletBackend.ParseAddr(peer1.OffChainAddrString)
	if err != nil {
		panic(err)
	}
	peer2.OffChainAddr, err = walletBackend.ParseAddr(peer2.OffChainAddrString)
	if err != nil {
		panic(err)
	}
	peer3.OffChainAddr, err = walletBackend.ParseAddr(peer3.OffChainAddrString)
	if err != nil {
		panic(err)
	}
}

func Test_IDReader_Interface(t *testing.T) {
	assert.Implements(t, (*perun.IDReader)(nil), new(local.IDProvider))
}

func Test_IDProvider_Interface(t *testing.T) {
	assert.Implements(t, (*perun.IDProvider)(nil), new(local.IDProvider))
}

func Test_NewIDProviderFromYaml_ReadByAlias(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		idProviderFile := idprovidertest.NewIDProviderT(t, peer1, peer2)

		gotPeerIDs, err := local.NewIDprovider(idProviderFile, walletBackend)

		assert.NoError(t, err)
		gotPeer1, isPresent := gotPeerIDs.ReadByAlias(peer1.Alias)
		assert.Equal(t, peer1, gotPeer1)
		assert.True(t, isPresent)
		gotPeer2, isPresent := gotPeerIDs.ReadByAlias(peer2.Alias)
		assert.Equal(t, peer2, gotPeer2)
		assert.True(t, isPresent)
		_, isPresent = gotPeerIDs.ReadByAlias(peer3.Alias)
		assert.False(t, isPresent)
	})

	t.Run("corrupted_yaml", func(t *testing.T) {
		idProviderFile := newCorruptedYAMLFile(t)
		_, err := local.NewIDprovider(idProviderFile, walletBackend)
		assert.Error(t, err)
		t.Log(err)
	})

	t.Run("invalid_offchain_addr", func(t *testing.T) {
		peer1Copy := peer1
		peer1Copy.OffChainAddrString = "invalid address"
		idProviderFile := idprovidertest.NewIDProviderT(t, peer1Copy, peer2)

		_, err := local.NewIDprovider(idProviderFile, walletBackend)
		assert.Error(t, err)
		t.Log(err)
	})

	t.Run("missing_file", func(t *testing.T) {
		_, err := local.NewIDprovider("./random-file.yaml", walletBackend)
		assert.Error(t, err)
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

	tempFile, err := os.CreateTemp("", "")
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

// nolint:dupl  // False positive. ReadByAlias is diff from ReadByOffChainAddr.
func Test_YAML_ReadByAlias(t *testing.T) {
	idProviderFile := idprovidertest.NewIDProviderT(t, peer1, peer2)
	c, err := local.NewIDprovider(idProviderFile, walletBackend)
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		gotPeer, isPresent := c.ReadByAlias(peer1.Alias)
		assert.True(t, isPresent)
		assert.Equal(t, gotPeer, peer1)
	})

	t.Run("missing_peer", func(t *testing.T) {
		_, isPresent := c.ReadByAlias(peer3.Alias)
		assert.False(t, isPresent)
	})
}

// nolint:dupl  // False positive. ReadByOffChainAddr is diff from ReadByAlias.
func Test_YAML_ReadByOffChainAddr(t *testing.T) {
	idProviderFile := idprovidertest.NewIDProviderT(t, peer1, peer2)
	c, err := local.NewIDprovider(idProviderFile, walletBackend)
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		gotPeer, isPresent := c.ReadByOffChainAddr(peer1.OffChainAddr)
		assert.True(t, isPresent)
		assert.Equal(t, peer1, gotPeer)
	})

	t.Run("missing_peer", func(t *testing.T) {
		_, isPresent := c.ReadByOffChainAddr(peer3.OffChainAddr)
		assert.False(t, isPresent)
	})
}

func Test_YAML_Write_Read(t *testing.T) {
	idProviderFile := idprovidertest.NewIDProviderT(t, peer1, peer2)
	c, err := local.NewIDprovider(idProviderFile, walletBackend)
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, c.Write(peer3.Alias, peer3))
		gotPeer, isPresent := c.ReadByAlias(peer3.Alias)
		assert.True(t, isPresent)
		assert.Equal(t, gotPeer, peer3)
	})

	t.Run("peer_already_present", func(t *testing.T) {
		err := c.Write(peer1.Alias, peer1)
		assert.True(t, errors.Is(err, idprovider.ErrPeerIDAlreadyRegistered))
	})

	t.Run("alias_used_by_diff_peer", func(t *testing.T) {
		err := c.Write(peer1.Alias, peer2)
		assert.True(t, errors.Is(err, idprovider.ErrPeerAliasAlreadyUsed))
	})

	t.Run("invalid_offchain_addr", func(t *testing.T) {
		idProviderFile := idprovidertest.NewIDProviderT(t, peer1, peer2)
		c, err := local.NewIDprovider(idProviderFile, walletBackend)
		assert.NoError(t, err)

		peer3Copy := peer3
		peer3Copy.OffChainAddrString = "invalid-addr"
		err = c.Write(peer3Copy.Alias, peer3Copy)

		assert.True(t, errors.Is(err, idprovider.ErrParsingOffChainAddress))
	})
}

func Test_YAML_Delete_Read(t *testing.T) {
	idProviderFile := idprovidertest.NewIDProviderT(t, peer1, peer2)
	c, err := local.NewIDprovider(idProviderFile, walletBackend)
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, c.Delete(peer1.Alias))
		_, isPresent := c.ReadByAlias(peer1.Alias)
		assert.False(t, isPresent)
	})

	t.Run("missing_peer", func(t *testing.T) {
		err := c.Delete(peer3.Alias)
		assert.Error(t, err)
		t.Log(err)
	})
}

func Test_YAML_UpdateStorage(t *testing.T) {
	t.Run("happy_empty_file", func(t *testing.T) {
		// Setup: NewYAML with zero entries
		emptyFile := idprovidertest.NewIDProviderT(t)
		fileWithTwoPeerIDs := idprovidertest.NewIDProviderT(t, peer1, peer2)
		c, err := local.NewIDprovider(emptyFile, walletBackend)
		assert.NoError(t, err)

		// Setup: Add entries to cache.
		require.NoError(t, c.Write(peer1.Alias, peer1))
		require.NoError(t, c.Write(peer2.Alias, peer2))

		// Test
		assert.NoError(t, c.UpdateStorage())
		assert.True(t, compareFileContent(t, emptyFile, fileWithTwoPeerIDs))
	})

	t.Run("happy_non_empty_file", func(t *testing.T) {
		fileWithTwoPeerIDs := idprovidertest.NewIDProviderT(t, peer1, peer2)
		fileWithThreePeerIDs := idprovidertest.NewIDProviderT(t, peer1, peer2, peer3)
		c, err := local.NewIDprovider(fileWithTwoPeerIDs, walletBackend)
		assert.NoError(t, err)
		assert.NoError(t, c.Write(peer3.Alias, peer3))

		// Test
		assert.NoError(t, c.UpdateStorage())
		assert.True(t, compareFileContent(t, fileWithTwoPeerIDs, fileWithThreePeerIDs))
	})

	t.Run("file_permission_error", func(t *testing.T) {
		// Setup: Create a copy of idProvider file with test data and add entry
		idProviderFile := idprovidertest.NewIDProviderT(t, peer1, peer2)
		c, err := local.NewIDprovider(idProviderFile, walletBackend)
		assert.NoError(t, err)
		assert.NoError(t, c.Write(peer3.Alias, peer3))

		// Change file permission
		err = os.Chmod(idProviderFile, 0o444)
		require.NoError(t, err)

		// Test
		err = c.UpdateStorage()
		assert.Error(t, err)
		t.Log(err)
	})
}

func compareFileContent(t *testing.T, file1, file2 string) bool {
	f1, err := os.ReadFile(file1)
	require.NoError(t, err)
	f2, err := os.ReadFile(file2)
	require.NoError(t, err)

	return bytes.Equal(f1, f2)
}
