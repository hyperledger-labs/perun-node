// Copyright (c) 2021 - for information on the respective copyright owner
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

package internal_test

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/internal"
)

func Test_ROChainBackend_Interface(t *testing.T) {
	assert.Implements(t, (*perun.ChainBackend)(nil), new(internal.ChainBackend))
}

func Test_ROChainBackend_ValidateAdjudicator(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	// use a real chain backend instead of simulated because, multiple connections are to be made:
	// 1. With a ChainBackend for deploying contracts.
	// 2. Use a ROChainbackend for validating contracts.
	adjudicator, _ := ethereumtest.SetupContractsT(t,
		ethereumtest.ChainURL, ethereumtest.ChainID, ethereumtest.OnChainTxTimeout)
	roChainBackend, err := ethereum.NewROChainBackend(
		ethereumtest.ChainURL, ethereumtest.ChainID, ethereumtest.ChainConnTimeout)
	require.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, roChainBackend.ValidateAdjudicator(adjudicator))
	})

	t.Run("invalid_adjudicator", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		err := roChainBackend.ValidateAdjudicator(randomAddr1)

		require.Error(t, err)
		invalidContractError := blockchain.InvalidContractError{}
		ok := errors.As(err, &invalidContractError)
		require.True(t, ok)
		assert.Equal(t, blockchain.Adjudicator, invalidContractError.Name)
		assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
	})
}

func Test_Integ_ROChainBackend_ValidateAssetHolderETH(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	// use a real chain backend instead of simulated because, multiple connections are to be made:
	// 1. With a ChainBackend for deploying contracts.
	// 2. Use a ROChainbackend for validating contracts.
	adjudicator, asset := ethereumtest.SetupContractsT(t,
		ethereumtest.ChainURL, ethereumtest.ChainID, ethereumtest.OnChainTxTimeout)
	roChainBackend, err := ethereum.NewROChainBackend(
		ethereumtest.ChainURL, ethereumtest.ChainID, ethereumtest.ChainConnTimeout)
	require.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, roChainBackend.ValidateAdjudicator(adjudicator))
		assert.NoError(t, roChainBackend.ValidateAssetHolderETH(adjudicator, asset))
	})
	t.Run("invalid_adjudicator", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		err := roChainBackend.ValidateAssetHolderETH(randomAddr1, asset)

		require.Error(t, err)
		invalidContractError := blockchain.InvalidContractError{}
		ok := errors.As(err, &invalidContractError)
		require.True(t, ok)
		assert.Equal(t, blockchain.Adjudicator, invalidContractError.Name)
		assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
	})
	t.Run("invalid_assetHolderETH", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		err := roChainBackend.ValidateAssetHolderETH(adjudicator, randomAddr1)

		require.Error(t, err)
		invalidContractError := blockchain.InvalidContractError{}
		ok := errors.As(err, &invalidContractError)
		require.True(t, ok)
		assert.Equal(t, blockchain.AssetHolderETH, invalidContractError.Name)
		assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
	})
}
