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

package contactsyaml_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/contacts/contactsyaml"
)

var (
	peer1 = perun.Peer{
		Alias:              "Alice",
		OffChainAddrString: "0x9282681723920798983380581376586951466585",
		CommType:           "tcpip",
		CommAddr:           "127.0.0.1:5751",
	}
	peer2 = perun.Peer{
		Alias:              "Bob",
		OffChainAddrString: "0x3369783337071807248093730889602727505701",
		CommType:           "tcpip",
		CommAddr:           "127.0.0.1:5750",
	}
	missingPeer = perun.Peer{
		Alias:              "Tom",
		OffChainAddrString: "0x7187308896023072480933697833370727318468",
		CommType:           "tcpip",
		CommAddr:           "127.0.0.1:5753",
	}

	walletBackend = ethereum.NewWalletBackend()

	testdataDir = filepath.Join("..", "..", "testdata", "contacts")

	testDataFile            = filepath.Join(testdataDir, "test.yaml")
	updatedTestDataFile     = filepath.Join(testdataDir, "test_added_entries.yaml")
	invalidOffChainAddrFile = filepath.Join(testdataDir, "invalid_addr.yaml")
	corruptedYAML           = filepath.Join(testdataDir, "corrupted.yaml")
	missingFile             = "./con.yml"
)

func init() {
	peer1.OffChainAddr, _ = walletBackend.ParseAddr(peer1.OffChainAddrString)             // nolint:errcheck
	peer2.OffChainAddr, _ = walletBackend.ParseAddr(peer2.OffChainAddrString)             // nolint:errcheck
	missingPeer.OffChainAddr, _ = walletBackend.ParseAddr(missingPeer.OffChainAddrString) // nolint:errcheck
}

func Test_ContactsReader_Interface(t *testing.T) {
	assert.Implements(t, (*perun.ContactsReader)(nil), new(contactsyaml.Provider))
}

func Test_Contacts_Interface(t *testing.T) {
	assert.Implements(t, (*perun.Contacts)(nil), new(contactsyaml.Provider))
}

func Test_NewContactsFromYaml(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		gotContacts, err := contactsyaml.New(tempContactsFile(t), walletBackend)
		assert.NoError(t, err)

		gotPeer1, isPresent := gotContacts.ReadByAlias(peer1.Alias)
		assert.Equal(t, peer1, gotPeer1)
		assert.True(t, isPresent)

		gotPeer2, isPresent := gotContacts.ReadByAlias(peer2.Alias)
		assert.Equal(t, peer2, gotPeer2)
		assert.True(t, isPresent)

		_, isPresent = gotContacts.ReadByAlias(missingPeer.Alias)
		assert.False(t, isPresent)
	})

	t.Run("corrupted_yaml", func(t *testing.T) {
		_, err := contactsyaml.New(corruptedYAML, walletBackend)
		assert.Error(t, err)
		t.Log(err)
	})

	t.Run("invalid_offchain_addr", func(t *testing.T) {
		_, err := contactsyaml.New(invalidOffChainAddrFile, walletBackend)
		assert.Error(t, err)
		t.Log(err)
	})

	t.Run("missing_file", func(t *testing.T) {
		_, err := contactsyaml.New(missingFile, walletBackend)
		assert.Error(t, err)
		t.Log(err)
	})
}

// nolint:dupl  // False positive. ReadByAlias is diff from ReadByOffChainAddr.
func Test_YAML_ReadByAlias(t *testing.T) {
	c, err := contactsyaml.New(tempContactsFile(t), walletBackend)
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		gotPeer, isPresent := c.ReadByAlias(peer1.Alias)
		assert.True(t, isPresent)
		assert.Equal(t, gotPeer, peer1)
	})

	t.Run("missing_peer", func(t *testing.T) {
		_, isPresent := c.ReadByAlias(missingPeer.Alias)
		assert.False(t, isPresent)
	})
}

// nolint:dupl  // False positive. ReadByOffChainAddr is diff from ReadByAlias.
func Test_YAML_ReadByOffChainAddr(t *testing.T) {
	c, err := contactsyaml.New(tempContactsFile(t), walletBackend)
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		gotPeer, isPresent := c.ReadByOffChainAddr(peer1.OffChainAddr)
		assert.True(t, isPresent)
		assert.Equal(t, peer1, gotPeer)
	})

	t.Run("missing_peer", func(t *testing.T) {
		_, isPresent := c.ReadByOffChainAddr(missingPeer.OffChainAddr)
		assert.False(t, isPresent)
	})
}

func Test_YAML_Write_Read(t *testing.T) {
	c, err := contactsyaml.New(testDataFile, walletBackend)
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, c.Write(missingPeer.Alias, missingPeer))
		gotPeer, isPresent := c.ReadByAlias(missingPeer.Alias)
		assert.True(t, isPresent)
		assert.Equal(t, gotPeer, missingPeer)
	})

	t.Run("peer_already_present", func(t *testing.T) {
		err := c.Write(peer1.Alias, peer1)
		assert.Error(t, err)
		t.Log(err)
	})

	t.Run("alias_used_by_diff_peer", func(t *testing.T) {
		err := c.Write(peer1.Alias, peer2)
		assert.Error(t, err)
		t.Log(err)
	})

	t.Run("invalid_offchain_addr", func(t *testing.T) {
		c, err := contactsyaml.New(testDataFile, walletBackend)
		assert.NoError(t, err)

		missingPeerCopy := missingPeer
		missingPeerCopy.OffChainAddrString = "invalid-addr"
		err = c.Write(missingPeerCopy.Alias, missingPeerCopy)
		assert.Error(t, err)
		t.Log(err)
	})
}

func Test_YAML_Delete_Read(t *testing.T) {
	c, err := contactsyaml.New(testDataFile, walletBackend)
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, c.Delete(peer1.Alias))
		_, isPresent := c.ReadByAlias(peer1.Alias)
		assert.False(t, isPresent)
	})

	t.Run("missing_peer", func(t *testing.T) {
		err := c.Delete(missingPeer.Alias)
		assert.Error(t, err)
		t.Log(err)
	})
}

func Test_YAML_UpdateStorage(t *testing.T) {
	t.Run("happy_empty_file", func(t *testing.T) {
		// Setup: NewYAML with zero entries
		tempFile, err := ioutil.TempFile("", "")
		require.NoError(t, err)
		require.NoError(t, tempFile.Close())
		t.Cleanup(func() {
			if err = os.Remove(tempFile.Name()); err != nil {
				t.Log("Error in test cleanup: removing file - " + tempFile.Name())
			}
		})
		c, err := contactsyaml.New(tempFile.Name(), walletBackend)
		assert.NoError(t, err)

		// Setup: Add entries to cache.
		require.NoError(t, c.Write(peer1.Alias, peer1))
		require.NoError(t, c.Write(peer2.Alias, peer2))

		// Test
		assert.NoError(t, c.UpdateStorage())
		assert.True(t, compareFileContent(t, tempFile.Name(), testDataFile))
	})

	t.Run("happy_non_empty_file", func(t *testing.T) {
		// Setup: Create a copy of contacts file with test data and add entry
		tempTestDataFile := tempContactsFile(t)
		c, err := contactsyaml.New(tempTestDataFile, walletBackend)
		assert.NoError(t, err)
		assert.NoError(t, c.Write(missingPeer.Alias, missingPeer))

		// Test
		assert.NoError(t, c.UpdateStorage())
		assert.True(t, compareFileContent(t, tempTestDataFile, updatedTestDataFile))
	})

	t.Run("file_permission_error", func(t *testing.T) {
		// Setup: Create a copy of contacts file with test data and add entry
		tempTestDataFile := tempContactsFile(t)
		c, err := contactsyaml.New(tempTestDataFile, walletBackend)
		assert.NoError(t, err)
		assert.NoError(t, c.Write(missingPeer.Alias, missingPeer))

		// Change file permission
		err = os.Chmod(tempTestDataFile, 0o444)
		require.NoError(t, err)

		// Test
		err = c.UpdateStorage()
		assert.Error(t, err)
		t.Log(err)
	})
}

// tempContactsFile makes a copy of the testdata contacts file and
// returns the path to it.
func tempContactsFile(t *testing.T) string {
	tempFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	testDataFile, err := os.Open(testDataFile)
	require.NoError(t, err)

	_, err = io.Copy(tempFile, testDataFile)
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())
	require.NoError(t, testDataFile.Close())

	t.Cleanup(func() {
		if err = os.Remove(tempFile.Name()); err != nil {
			t.Log("Error in test cleanup: removing file - " + tempFile.Name())
		}
	})
	return tempFile.Name()
}

func compareFileContent(t *testing.T, file1, file2 string) bool {
	f1, err := ioutil.ReadFile(file1)
	require.NoError(t, err)
	f2, err := ioutil.ReadFile(file2)
	require.NoError(t, err)

	return bytes.Equal(f1, f2)
}
