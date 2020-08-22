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

package contactstest

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/hyperledger-labs/perun-node"
)

// NewYAMLFile creates a temporary file containing the details of given peers and
// returns the path to it. It also registers a cleanup function on the passed test handler.
func NewYAMLFile(t *testing.T, peers ...perun.Peer) string {
	tempFile, err := ioutil.TempFile("", "")
	defer func() {
		require.NoErrorf(t, tempFile.Close(), "closing temporary file")
	}()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err = os.Remove(tempFile.Name()); err != nil {
			t.Log("Error in test cleanup: removing file - " + tempFile.Name())
		}
	})
	contacts := make(map[string]perun.Peer, len(peers))
	for _, peer := range peers {
		contacts[peer.Alias] = peer
	}

	encoder := yaml.NewEncoder(tempFile)
	require.NoErrorf(t, encoder.Encode(contacts), "encoding contacts")
	require.NoErrorf(t, encoder.Close(), "closing encoder")
	return tempFile.Name()
}
