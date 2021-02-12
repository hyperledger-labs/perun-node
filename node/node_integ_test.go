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

package node_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/node"
	"github.com/hyperledger-labs/perun-node/node/nodetest"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

var validConfig perun.NodeConfig

func init() {
	validConfig = nodetest.NewConfig()
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
		_, err := node.New(cfg)
		require.Error(t, err)
	})

	t.Run("err_invalid_adjudicator", func(t *testing.T) {
		cfg := validConfig
		cfg.Adjudicator = "invalid-addr"
		_, err := node.New(cfg)
		require.Error(t, err)
	})

	t.Run("err_invalid_asset", func(t *testing.T) {
		cfg := validConfig
		cfg.Asset = "invalid-addr"
		_, err := node.New(cfg)
		require.Error(t, err)
	})

	t.Run("happy", func(t *testing.T) {
		var err error
		n, err = node.New(validConfig)
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
	var sessionID string
	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	t.Run("happy_OpenSession", func(t *testing.T) {
		var err error
		sessionCfg := sessiontest.NewConfigT(t, prng)
		sessionCfgFile := sessiontest.NewConfigFileT(t, sessionCfg)
		sessionID, _, err = n.OpenSession(sessionCfgFile)
		require.NoError(t, err)
		assert.NotZero(t, sessionID)
	})
	t.Run("happy_GetSession", func(t *testing.T) {
		sess, err := n.GetSession(sessionID)
		require.NoError(t, err)
		assert.NotNil(t, sess)
	})
	t.Run("err_GetSession_not_found", func(t *testing.T) {
		unknownSessID := "unknown session id"
		_, err := n.GetSessionV2(unknownSessID)
		require.Error(t, err)
		t.Log(err)

		wantMessage := node.ErrUnknownSessionID.Error()
		assert.Equal(t, perun.ClientError, err.Category())
		assert.Equal(t, perun.ErrV2ResourceNotFound, err.Code())
		assert.Equal(t, wantMessage, err.Message())
		addInfo, ok := err.AddInfo().(perun.ErrV2InfoResourceNotFound)
		require.True(t, ok)
		assert.Equal(t, addInfo.Type, "session id")
		assert.Equal(t, addInfo.ID, unknownSessID)
	})
	t.Run("err_OpenSession_config_file_error", func(t *testing.T) {
		_, _, err := n.OpenSession("random-config-file")
		require.Error(t, err)
		t.Log(err)
	})
	// Simulate one error to fail session.New
	// Complete test of session.New is done in the session package.
	t.Run("err_OpenSession_init_error", func(t *testing.T) {
		sessionCfg := sessiontest.NewConfigT(t, prng)
		sessionCfg.ChainURL = "invalid-url"
		sessionCfgFile := sessiontest.NewConfigFileT(t, sessionCfg)
		_, _, err := n.OpenSession(sessionCfgFile)
		require.Error(t, err)
		t.Log(err)
	})
}
