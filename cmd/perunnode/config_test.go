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

package perunnode_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/cmd/perunnode"
)

var (
	testdataDir       = "../../testdata/perunnode"
	validConfigFile   = "valid.yaml"
	invalidConfigFile = "invalid.yaml"
	// test config as expected from the testdata file at
	// ${PROJECT_ROOT}/testdata/perunnode/valid.yaml.
	testCfg = perun.NodeConfig{
		LogLevel:         "debug",
		LogFile:          "Node.log",
		ChainURL:         "ws://127.0.0.1:8545",
		Adjudicator:      "0x9daEdAcb21dce86Af8604Ba1A1D7F9BFE55ddd63",
		Asset:            "0x5992089d61cE79B6CF90506F70DD42B8E42FB21d",
		ChainConnTimeout: 10 * time.Second,
		OnChainTxTimeout: 10 * time.Second,
		ResponseTimeout:  30 * time.Second,
	}
)

func Test_ParseConfig(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		gotCfg, err := perunnode.ParseConfig(filepath.Join(testdataDir, validConfigFile))
		require.NoError(t, err)
		assert.DeepEqual(t, testCfg, gotCfg)
	})
	t.Run("err_invalid_file", func(t *testing.T) {
		_, err := perunnode.ParseConfig(filepath.Join(testdataDir, invalidConfigFile))
		require.Error(t, err)
		t.Log(err)
	})
	t.Run("err_missingFile", func(t *testing.T) {
		_, err := perunnode.ParseConfig("missing_file")
		require.Error(t, err)
		t.Log(err)
	})
}
