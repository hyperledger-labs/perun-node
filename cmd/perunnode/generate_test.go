// Copyright (c) 2023 - for information on the respective copyright owner
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

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSessionConfig(t *testing.T) {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "session_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir) //nolint:errcheck

	// Change to the temporary directory
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Now you can test the function with a clean temporary directory
	err = generateSessionConfig()
	require.NoError(t, err)

	// List of aliases for testing
	aliasesList := []string{"alice", "bob", "api"}

	// Verify the directories and files were created as expected
	for _, alias := range aliasesList {
		dirPath := filepath.Join(tempDir, alias)
		require.DirExists(t, dirPath)

		require.DirExists(t, filepath.Join(dirPath, "database"))
		require.FileExists(t, filepath.Join(dirPath, "idprovider.yaml"))
		require.DirExists(t, filepath.Join(dirPath, "keystore"))

		keystoreFiles, err := ioutil.ReadDir(filepath.Join(dirPath, "keystore"))
		require.NoError(t, err)
		require.Len(t, keystoreFiles, 2)
		for _, file := range keystoreFiles {
			require.True(t, strings.HasPrefix(file.Name(), "UTC"))
		}

		require.FileExists(t, filepath.Join(dirPath, "session.yaml"))
	}
}
