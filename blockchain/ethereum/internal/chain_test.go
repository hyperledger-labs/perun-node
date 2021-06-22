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

package internal_test

import (
	"errors"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/internal"
)

func Test_ChainBackend_Interface(t *testing.T) {
	assert.Implements(t, (*perun.ChainBackend)(nil), new(internal.ChainBackend))
	assert.Implements(t, (*perun.ROChainBackend)(nil), new(internal.ChainBackend))
}

func Test_ChainBackend_Deploy(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	setup := ethereumtest.NewSimChainBackendSetup(t, rng, 1)

	onChainAcc := setup.Accs[0].Address()
	adjudicator, err := setup.ChainBackend.DeployAdjudicator(onChainAcc)
	require.NoError(t, err)
	require.NotNil(t, adjudicator)

	assetETH, err := setup.ChainBackend.DeployAssetETH(adjudicator, onChainAcc)
	require.NoError(t, err)
	require.NotNil(t, assetETH)

	initAccs := []pwallet.Address{setup.Accs[0].Address()}
	initBal := big.NewInt(10)
	tokenERC20PRN, err := setup.ChainBackend.DeployPerunToken(initAccs, initBal, onChainAcc)
	require.NoError(t, err)
	require.NotNil(t, tokenERC20PRN)

	assetERC20PRN, err := setup.ChainBackend.DeployAssetERC20(adjudicator, tokenERC20PRN, onChainAcc)
	require.NoError(t, err)
	require.NotNil(t, assetERC20PRN)
}

func Test_ChainBackend_ValidateAdjudicator(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	setup := ethereumtest.NewSimChainBackendSetup(t, rng, 1)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, setup.ChainBackend.ValidateAdjudicator(setup.Adjudicator))
	})
	t.Run("invalid_adjudicator", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		err := setup.ChainBackend.ValidateAdjudicator(randomAddr1)

		require.Error(t, err)
		invalidContractError := blockchain.InvalidContractError{}
		ok := errors.As(err, &invalidContractError)
		require.True(t, ok)
		assert.Equal(t, blockchain.Adjudicator, invalidContractError.Name)
		assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
	})
}

func Test_ChainBackend_ValidateAssetETH(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	setup := ethereumtest.NewSimChainBackendSetup(t, rng, 1)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, setup.ChainBackend.ValidateAdjudicator(setup.Adjudicator))
		assert.NoError(t, setup.ChainBackend.ValidateAssetETH(setup.Adjudicator, setup.AssetETH))
	})
	t.Run("invalid_adjudicator", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		err := setup.ChainBackend.ValidateAssetETH(randomAddr1, setup.AssetETH)

		require.Error(t, err)
		invalidContractError := blockchain.InvalidContractError{}
		ok := errors.As(err, &invalidContractError)
		require.True(t, ok)
		assert.Equal(t, blockchain.Adjudicator, invalidContractError.Name)
		assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
	})
	t.Run("invalid_assetETH", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		err := setup.ChainBackend.ValidateAssetETH(setup.Adjudicator, randomAddr1)

		require.Error(t, err)
		invalidContractError := blockchain.InvalidContractError{}
		ok := errors.As(err, &invalidContractError)
		require.True(t, ok)
		assert.Equal(t, blockchain.AssetETH, invalidContractError.Name)
		assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
	})
}

func Test_ChainBackend_ValidateAssetERC20(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	setup := ethereumtest.NewSimChainBackendSetup(t, rng, 1)

	t.Run("happy", func(t *testing.T) {
		require.Len(t, setup.AssetERC20s, 1)
		for tokenERC20, assetERC20 := range setup.AssetERC20s {
			symbol, decimals, err := setup.ChainBackend.ValidateAssetERC20(setup.Adjudicator, tokenERC20, assetERC20)

			require.NoError(t, err)
			require.Equal(t, "PRN", symbol)       // Symbol for the perun ERC20 token deployed for test.
			require.Equal(t, uint8(18), decimals) // MaxDecimals for the perun ERC20 token deployed for tests.
		}
	})
	t.Run("invalid_adjudicator", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		for tokenERC20, assetERC20 := range setup.AssetERC20s {
			_, _, err := setup.ChainBackend.ValidateAssetERC20(randomAddr1, tokenERC20, assetERC20)

			require.Error(t, err)
			invalidContractError := blockchain.InvalidContractError{}
			ok := errors.As(err, &invalidContractError)
			require.True(t, ok)
			assert.Equal(t, blockchain.Adjudicator, invalidContractError.Name)
			assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
		}
	})
	t.Run("invalid_assetERC20", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		for tokenERC20 := range setup.AssetERC20s {
			_, _, err := setup.ChainBackend.ValidateAssetERC20(setup.Adjudicator, tokenERC20, randomAddr1)

			require.Error(t, err)
			t.Logf("%+v", err)
			invalidContractError := blockchain.InvalidContractError{}
			ok := errors.As(err, &invalidContractError)
			require.True(t, ok)
			t.Logf("%+v", invalidContractError)
			assert.Equal(t, "PRN", invalidContractError.Name)
			assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
		}
	})
	t.Run("invalid_tokenAddr", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		for _, assetERC20 := range setup.AssetERC20s {
			_, _, err := setup.ChainBackend.ValidateAssetERC20(setup.Adjudicator, randomAddr1, assetERC20)

			require.Error(t, err)
			invalidContractError := blockchain.InvalidContractError{}
			ok := errors.As(err, &invalidContractError)
			require.False(t, ok)
		}
	})
}

func Test_ChainBackend_NewFunder(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	setup := ethereumtest.NewSimChainBackendSetup(t, rng, 1)
	randomAddr1 := ethereumtest.NewRandomAddress(rng)
	randomAddr2 := ethereumtest.NewRandomAddress(rng)

	assert.NotNil(t, setup.ChainBackend.NewFunder(randomAddr1, randomAddr2))
}

func Test_ChainBackend_NewAdjudicator(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	setup := ethereumtest.NewSimChainBackendSetup(t, rng, 1)
	randomAddr1 := ethereumtest.NewRandomAddress(rng)
	randomAddr2 := ethereumtest.NewRandomAddress(rng)

	assert.NotNil(t, setup.ChainBackend.NewAdjudicator(randomAddr1, randomAddr2))
}
