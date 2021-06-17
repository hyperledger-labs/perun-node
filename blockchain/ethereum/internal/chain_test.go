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
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/internal"
)

func Test_ChainBackend_Interface(t *testing.T) {
	assert.Implements(t, (*perun.ChainBackend)(nil), new(internal.ChainBackend))
}

func Test_ChainBackend_Deploy(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	setup := ethereumtest.NewChainBackendSetup(t, rng, 1)

	onChainAddr := setup.Accs[0].Address()
	adjAddr, err := setup.ChainBackend.DeployAdjudicator(onChainAddr)
	require.NoError(t, err)
	assetAddr, err := setup.ChainBackend.DeployAsset(adjAddr, onChainAddr)
	require.NoError(t, err)
	assert.NoError(t, setup.ChainBackend.ValidateAdjudicator(adjAddr))
	assert.NoError(t, setup.ChainBackend.ValidateAssetHolderETH(adjAddr, assetAddr))
}

func Test_ChainBackend_ValidateAdjudicator(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	setup := ethereumtest.NewChainBackendSetup(t, rng, 1)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, setup.ChainBackend.ValidateAdjudicator(setup.AdjAddr))
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

func Test_ChainBackend_ValidateAssetHolderETH(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	setup := ethereumtest.NewChainBackendSetup(t, rng, 1)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, setup.ChainBackend.ValidateAdjudicator(setup.AdjAddr))
		assert.NoError(t, setup.ChainBackend.ValidateAssetHolderETH(setup.AdjAddr, setup.AssetAddr))
	})
	t.Run("invalid_adjudicator", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		err := setup.ChainBackend.ValidateAssetHolderETH(randomAddr1, setup.AssetAddr)

		require.Error(t, err)
		invalidContractError := blockchain.InvalidContractError{}
		ok := errors.As(err, &invalidContractError)
		require.True(t, ok)
		assert.Equal(t, blockchain.Adjudicator, invalidContractError.Name)
		assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
	})
	t.Run("invalid_assetHolderETH", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		err := setup.ChainBackend.ValidateAssetHolderETH(setup.AdjAddr, randomAddr1)

		require.Error(t, err)
		invalidContractError := blockchain.InvalidContractError{}
		ok := errors.As(err, &invalidContractError)
		require.True(t, ok)
		assert.Equal(t, blockchain.AssetHolderETH, invalidContractError.Name)
		assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
	})
}

func Test_ChainBackend_NewFunder(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	setup := ethereumtest.NewChainBackendSetup(t, rng, 1)
	randomAddr1 := ethereumtest.NewRandomAddress(rng)
	randomAddr2 := ethereumtest.NewRandomAddress(rng)

	assert.NotNil(t, setup.ChainBackend.NewFunder(randomAddr1, randomAddr2))
}

func Test_ChainBackend_NewAdjudicator(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	setup := ethereumtest.NewChainBackendSetup(t, rng, 1)
	randomAddr1 := ethereumtest.NewRandomAddress(rng)
	randomAddr2 := ethereumtest.NewRandomAddress(rng)

	assert.NotNil(t, setup.ChainBackend.NewAdjudicator(randomAddr1, randomAddr2))
}
