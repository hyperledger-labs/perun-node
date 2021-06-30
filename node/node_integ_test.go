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
	"github.com/hyperledger-labs/perun-node/peruntest"
	"github.com/hyperledger-labs/perun-node/session"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

func Test_Integ_New(t *testing.T) {
	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))

	// Deploy contracts.
	ethereumtest.SetupContractsT(t, ethereumtest.ChainURL, ethereumtest.ChainID, ethereumtest.OnChainTxTimeout, false)

	t.Run("happy", func(t *testing.T) {
		n, err := node.New(nodetest.NewConfig())
		require.NoError(t, err)
		require.NotNil(t, n)
	})

	t.Run("err_invalid_log_level", func(t *testing.T) {
		cfg := nodetest.NewConfig()
		cfg.LogLevel = ""
		_, err := node.New(cfg)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("err_invalid_adjudicator_address", func(t *testing.T) {
		cfg := nodetest.NewConfig()
		cfg.Adjudicator = "invalid-addr"
		_, err := node.New(cfg)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("err_invalid_asset_address", func(t *testing.T) {
		cfg := nodetest.NewConfig()
		cfg.AssetETH = "invalid-addr"
		_, err := node.New(cfg)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("err_invalid_adjudicator_contract", func(t *testing.T) {
		cfg := nodetest.NewConfig()
		cfg.Adjudicator = ethereumtest.NewRandomAddress(prng).String()
		_, err := node.New(cfg)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("err_invalid_assetETH_contract", func(t *testing.T) {
		cfg := nodetest.NewConfig()
		cfg.AssetETH = ethereumtest.NewRandomAddress(prng).String()
		_, err := node.New(cfg)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("err_invalid_adjudicator_assetETH_contract", func(t *testing.T) {
		cfg := nodetest.NewConfig()
		cfg.Adjudicator = ethereumtest.NewRandomAddress(prng).String()
		cfg.AssetETH = ethereumtest.NewRandomAddress(prng).String()
		_, err := node.New(cfg)
		require.Error(t, err)
		t.Log(err)
	})

}

func Test_Integ_Node_Time(t *testing.T) {
	n, err := node.New(nodetest.NewConfig())
	require.NoError(t, err)
	require.NotNil(t, n)

	assert.GreaterOrEqual(t, time.Now().UTC().Unix()+5, n.Time())
}

func Test_Integ_Node_GetConfig(t *testing.T) {
	n, err := node.New(nodetest.NewConfig())
	require.NoError(t, err)
	require.NotNil(t, n)

	cfg := n.GetConfig()
	assert.Equal(t, nodetest.NewConfig(), cfg)
}

func Test_Integ_Node_Help(t *testing.T) {
	n, err := node.New(nodetest.NewConfig())
	require.NoError(t, err)
	require.NotNil(t, n)

	apis := n.Help()
	assert.Equal(t, []string{"payment"}, apis)
}

func Test_Integ_Node_OpenSession_GetSession(t *testing.T) {

	t.Run("happy", func(t *testing.T) {
		n, err := node.New(nodetest.NewConfig())
		require.NoError(t, err)
		require.NotNil(t, n)
		prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
		sessionCfg := sessiontest.NewConfigT(t, prng)
		sessionCfgFile := sessiontest.NewConfigFileT(t, sessionCfg)

		sessionID, _, err := n.OpenSession(sessionCfgFile)
		require.NoError(t, err)
		assert.NotZero(t, sessionID)

		sess, err := n.GetSession(sessionID)
		require.NoError(t, err)
		assert.NotNil(t, sess)
	})

	// Simulate one error to fail session.New and check if err is not nil.
	// Complete test of session.New is done in the session package.
	t.Run("OpenSession_init_error", func(t *testing.T) {
		n, err := node.New(nodetest.NewConfig())
		require.NoError(t, err)
		require.NotNil(t, n)
		invalidChainURL := "invalid-url"
		prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
		sessionCfg := sessiontest.NewConfigT(t, prng)
		sessionCfg.ChainURL = invalidChainURL
		sessionCfgFile := sessiontest.NewConfigFileT(t, sessionCfg)

		_, _, err = n.OpenSession(sessionCfgFile)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("GetSession_not_found", func(t *testing.T) {
		n, err := node.New(nodetest.NewConfig())
		require.NoError(t, err)
		require.NotNil(t, n)
		unknownSessID := "unknown session id"

		_, apiErr := n.GetSession(unknownSessID)
		peruntest.AssertAPIError(t, apiErr, perun.ClientError, perun.ErrResourceNotFound)
		peruntest.AssertErrInfoResourceNotFound(t, apiErr.AddInfo(), session.ResTypeSession, unknownSessID)
	})
}

func Test_Integ_Node_RegisterCurrency(t *testing.T) {
	t.Run("RegisterCurrency_Exists", func(t *testing.T) {
		n, err := node.New(nodetest.NewConfig())
		require.NoError(t, err)
		require.NotNil(t, n)

		// PRN token is deployed by SetupContracts function and included in node config.
		// So registering it once again will return an error.
		//
		// TODO: Make deploy perun token function to take symbol and maxDecimals as input arguments. Do happy test.
		_, _, assetERC20s := ethereumtest.ContractAddrs()
		require.Len(t, assetERC20s, 1)
		for tokenERC20Addr, assetERC20Addr := range assetERC20s {
			_, apiErr := n.RegisterCurrency(tokenERC20Addr.String(), assetERC20Addr.String())
			require.Error(t, apiErr)

			peruntest.AssertAPIError(t, apiErr, perun.ClientError, perun.ErrResourceExists)
			peruntest.AssertErrInfoResourceExists(t, apiErr.AddInfo(), session.ResTypeCurrency, "PRN")
		}
	})
}
