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

package session_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/session"
)

var (
	testdataDir       = "../testdata/session"
	validConfigFile   = "valid.yaml"
	invalidConfigFile = "invalid.yaml"
	// test cofiguration as in the testdata file at
	// ${PROJECT_ROOT}/testdata/session/valid.yaml.
	testCfg = session.Config{
		User: session.UserConfig{
			Alias:       perun.OwnAlias,
			OnChainAddr: "0x9282681723920798983380581376586951466585",
			OnChainWallet: session.WalletConfig{
				KeystorePath: "./test-keystore-on-chain",
				Password:     "test-password-on-chain",
			},
			OffChainAddr: "0x3369783337071807248093730889602727505701",
			OffChainWallet: session.WalletConfig{
				KeystorePath: "./test-keystore-off-chain",
				Password:     "test-password-off-chain",
			},
			CommAddr: "127.0.0.1:5751",
			CommType: "tcp",
		},
		IDProviderType: "yaml",
		IDProviderURL:  "./test-idprovider.yaml",
		ChainURL:       "ws://127.0.0.1:8545",
		Asset:          "0x5992089d61cE79B6CF90506F70DD42B8E42FB21d",
		Adjudicator:    "0x9daEdAcb21dce86Af8604Ba1A1D7F9BFE55ddd63",
		DatabaseDir:    "./test-db",
	}
)

func Test_ParseConfig(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		gotCfg, err := session.ParseConfig(filepath.Join(testdataDir, validConfigFile))
		require.NoError(t, err)
		assert.DeepEqual(t, testCfg, gotCfg)
	})
	t.Run("err_invalid_file", func(t *testing.T) {
		_, err := session.ParseConfig(filepath.Join(testdataDir, invalidConfigFile))
		require.Error(t, err)
		t.Log(err)
	})
	t.Run("err_missingFile", func(t *testing.T) {
		_, err := session.ParseConfig("missing_file")
		require.Error(t, err)
		t.Log(err)
	})
}
