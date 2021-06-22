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

	"github.com/hyperledger-labs/perun-node/blockchain"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
)

func Test_ROChainBackend_ValidateAdjudicator(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	// use a real chain backend instead of simulated because, multiple connections are to be made:
	// 1. With a ChainBackend for deploying contracts.
	// 2. Use a ROChainbackend for validating contracts.
	contracts := ethereumtest.SetupContractsT(t,
		ethereumtest.ChainURL, ethereumtest.ChainID, ethereumtest.OnChainTxTimeout, false)
	roChainBackend, err := ethereum.NewROChainBackend(
		ethereumtest.ChainURL, ethereumtest.ChainID, ethereumtest.ChainConnTimeout)
	require.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, roChainBackend.ValidateAdjudicator(contracts.Adjudicator()))
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

func Test_Integ_ROChainBackend_ValidateAssetETH(t *testing.T) {
	rng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	// use a real chain backend instead of simulated because, multiple connections are to be made:
	// 1. With a ChainBackend for deploying contracts.
	// 2. Use a ROChainbackend for validating contracts.
	contracts := ethereumtest.SetupContractsT(t,
		ethereumtest.ChainURL, ethereumtest.ChainID, ethereumtest.OnChainTxTimeout, false)
	roChainBackend, err := ethereum.NewROChainBackend(
		ethereumtest.ChainURL, ethereumtest.ChainID, ethereumtest.ChainConnTimeout)
	require.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, roChainBackend.ValidateAdjudicator(contracts.Adjudicator()))
		assert.NoError(t, roChainBackend.ValidateAssetETH(contracts.Adjudicator(), contracts.AssetETH()))
	})
	t.Run("invalid_adjudicator", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		err := roChainBackend.ValidateAssetETH(randomAddr1, contracts.AssetETH())

		require.Error(t, err)
		invalidContractError := blockchain.InvalidContractError{}
		ok := errors.As(err, &invalidContractError)
		require.True(t, ok)
		assert.Equal(t, blockchain.Adjudicator, invalidContractError.Name)
		assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
	})
	t.Run("invalid_assetHolderETH", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		err := roChainBackend.ValidateAssetETH(contracts.Adjudicator(), randomAddr1)

		require.Error(t, err)
		invalidContractError := blockchain.InvalidContractError{}
		ok := errors.As(err, &invalidContractError)
		require.True(t, ok)
		assert.Equal(t, blockchain.AssetETH, invalidContractError.Name)
		assert.Equal(t, randomAddr1.String(), invalidContractError.Address)
	})
}
