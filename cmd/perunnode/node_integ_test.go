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

package perunnode_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/cmd/perunnode"
)

var validConfig = perun.NodeConfig{
	LogFile:      "",
	LogLevel:     "debug",
	ChainURL:     "ws://127.0.0.1:8545",
	Adjudicator:  "0x9daEdAcb21dce86Af8604Ba1A1D7F9BFE55ddd63",
	Asset:        "0x5992089d61cE79B6CF90506F70DD42B8E42FB21d",
	CommTypes:    []string{"tcp"},
	ContactTypes: []string{"yaml"},
	Currencies:   []string{"ETH"},

	ChainConnTimeout: 30 * time.Second,
	OnChainTxTimeout: 10 * time.Second,
	ResponseTimeout:  10 * time.Second,
}

// This NodeAPI instance will be set upon first happy test.
// This will be used in the subsequent tests.
var n perun.NodeAPI

// Node can be initialized only once and hence run all the error cases before a happy test.
// This test expects a working ganache-cli node to be started using specific options.
// See ethereumtest/setupcontracts.go for the details on the command.
func Test_Integ_New(t *testing.T) {
	t.Run("err_invalid_log_level", func(t *testing.T) {
		cfg := validConfig
		cfg.LogLevel = ""
		_, err := perunnode.New(cfg)
		require.Error(t, err)
	})

	t.Run("err_invalid_adjudicator", func(t *testing.T) {
		cfg := validConfig
		cfg.Adjudicator = "invalid-addr"
		_, err := perunnode.New(cfg)
		require.Error(t, err)
	})

	t.Run("err_invalid_asset", func(t *testing.T) {
		cfg := validConfig
		cfg.Asset = "invalid-addr"
		_, err := perunnode.New(cfg)
		require.Error(t, err)
	})

	t.Run("err_invalid_asset", func(t *testing.T) {
		var err error
		n, err = perunnode.New(validConfig)
		require.NoError(t, err)
		require.NotNil(t, n)
	})

	t.Run("happy_Time", func(t *testing.T) {
		assert.GreaterOrEqual(t, time.Now().UTC().Unix()+5, n.Time())
	})

	t.Run("happy_GetConfig", func(t *testing.T) {
		cfg := n.GetConfig()
		assert.Equal(t, validConfig, cfg)
	})

	t.Run("happy_Help", func(t *testing.T) {
		apis := n.Help()
		assert.Equal(t, []string{"payment"}, apis)
	})
}
