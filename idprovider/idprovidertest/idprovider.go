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

package idprovidertest

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/hyperledger-labs/perun-node"
)

// NewIDProviderT is the test friendly version of NewIDProvider.
// It uses the passed testing.T to handle the errors and registers the cleanup functions on it.
func NewIDProviderT(t *testing.T, peers ...perun.Peer) string {
	idProviderFile, err := NewIDProvider(peers...)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err = os.Remove(idProviderFile); err != nil {
			t.Log("Error in test cleanup: removing file - " + idProviderFile)
		}
	})
	return idProviderFile
}

// NewIDProvider creates a temporary file containing the details of given peers and
// returns the path to it. It also registers a cleanup function on the passed test handler.
func NewIDProvider(peers ...perun.Peer) (string, error) {
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", errors.Wrap(err, "creating temp file for local idProvider")
	}
	// if err = os.Remove(tempFile.Name()); err != nil {
	idProvider := make(map[string]perun.Peer, len(peers))
	for _, peer := range peers {
		idProvider[peer.Alias] = peer
	}

	encoder := yaml.NewEncoder(tempFile)
	if err := encoder.Encode(idProvider); err != nil {
		tempFile.Close()           // nolint: errcheck
		os.Remove(tempFile.Name()) // nolint: errcheck
		return "", errors.Wrap(err, "encoding idProvider")
	}
	if err := encoder.Close(); err != nil {
		tempFile.Close()           // nolint: errcheck
		os.Remove(tempFile.Name()) // nolint: errcheck
		return "", errors.Wrap(err, "closing encoder")
	}
	return tempFile.Name(), tempFile.Close()
}
