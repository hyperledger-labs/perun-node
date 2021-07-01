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

package ethereum_test

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node/blockchain"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/currency"
)

func Test_ContractRegistry_Adjudicator_AssetETH(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		// Each case not organized as sub-tests because order of execution is to be
		// maintained.

		rng := rand.New(rand.NewSource(rand.Int63()))
		setup := ethereumtest.NewSimChainBackendSetup(t, rng, 2)

		// happy/NewContractRegistry.
		r, err := ethereum.NewContractRegistry(setup.ChainBackend, setup.Adjudicator, setup.AssetETH)
		require.NoError(t, err)
		require.NotNil(t, r)

		// happy/registry.adjudicator.
		gotAdjudicator := r.Adjudicator()
		assert.NotNil(t, gotAdjudicator)
		assert.Equal(t, setup.Adjudicator, gotAdjudicator)

		// happy/registry.asset.
		assert.Equal(t, setup.AssetETH, r.AssetETH())
		gotAssetETH, found := r.Asset("ETH")
		assert.True(t, found)
		assert.Equal(t, setup.AssetETH, gotAssetETH)

		// happy/registry.assets.
		assets := r.Assets()
		require.Len(t, assets, 1, "should contain only one asset: ETH after init")
		require.Equal(t, setup.AssetETH.String(), assets[currency.ETHSymbol])
	})

	t.Run("invalid_adjudicator", func(t *testing.T) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		setup := ethereumtest.NewSimChainBackendSetup(t, rng, 2)
		randomAddr1 := ethereumtest.NewRandomAddress(rng)

		_, err := ethereum.NewContractRegistry(setup.ChainBackend, randomAddr1, setup.AssetETH)
		require.Error(t, err)
	})

	t.Run("invalid_asset", func(t *testing.T) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		setup := ethereumtest.NewSimChainBackendSetup(t, rng, 2)
		randomAddr1 := ethereumtest.NewRandomAddress(rng)

		_, err := ethereum.NewContractRegistry(setup.ChainBackend, setup.Adjudicator, randomAddr1)
		require.Error(t, err)
	})
}

func Test_ContractRegistry_ERC20(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		// Each case not organized as sub-tests because order of execution is to be
		// maintained.

		rng := rand.New(rand.NewSource(rand.Int63()))
		setup := ethereumtest.NewSimChainBackendSetup(t, rng, 2)
		perunSymbol := "PRN"

		// happy/NewContractRegistry
		r, err := ethereum.NewContractRegistry(setup.ChainBackend, setup.Adjudicator, setup.AssetETH)
		require.NoError(t, err)
		require.NotNil(t, r)

		// happy/registry.Asset.
		_, found := r.Asset(perunSymbol)
		assert.False(t, found, "No ERC20 tokens should be registered during init")

		var tokenERC20, assetERC20 pwallet.Address
		var symbol string
		require.Len(t, setup.AssetERC20s, 1, "setup should contain only one erc20 asset info")
		for tokenERC20, assetERC20 = range setup.AssetERC20s {
			symbol, _, err = r.RegisterAssetERC20(tokenERC20, assetERC20)
			require.NoError(t, err)
		}

		// happy/registry.Asset.
		gotAsset, found := r.Asset(perunSymbol)
		assert.True(t, found, "No ERC20 tokens should be registered during init")
		require.True(t, assetERC20.Equals(gotAsset))

		// happy/registry.Token.
		gotToken, found := r.Token(perunSymbol)
		assert.True(t, found, "No ERC20 tokens should be registered during init")
		require.True(t, tokenERC20.Equals(gotToken))

		// happy/registry.Symbol.
		gotSymbol, found := r.Symbol(assetERC20)
		assert.True(t, found, "No ERC20 tokens should be registered during init")
		require.Equal(t, perunSymbol, gotSymbol)

		// happy/registry.Assets.
		assets := r.Assets()
		require.Len(t, assets, 2, "should contain only two assets: ETH, PRN")
		require.Equal(t, setup.AssetETH.String(), assets[currency.ETHSymbol])
		require.Equal(t, assetERC20.String(), assets[symbol])
	})

	t.Run("invalid_tokenERC20", func(t *testing.T) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		setup := ethereumtest.NewSimChainBackendSetup(t, rng, 2)
		randomAddr1 := ethereumtest.NewRandomAddress(rng)

		r, err := ethereum.NewContractRegistry(setup.ChainBackend, setup.Adjudicator, setup.AssetETH)
		require.NoError(t, err)
		require.NotNil(t, r)

		for _, assetERC20 := range setup.AssetERC20s {
			_, _, err = r.RegisterAssetERC20(randomAddr1, assetERC20)
			require.Error(t, err)

			invalidContractErr := blockchain.InvalidContractError{}
			assert.False(t, errors.As(err, &invalidContractErr))
		}
	})

	t.Run("invalid_assetERC20", func(t *testing.T) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		setup := ethereumtest.NewSimChainBackendSetup(t, rng, 2)
		perunSymbol := "PRN"
		randomAddr1 := ethereumtest.NewRandomAddress(rng)

		r, err := ethereum.NewContractRegistry(setup.ChainBackend, setup.Adjudicator, setup.AssetETH)
		require.NoError(t, err)
		require.NotNil(t, r)

		for tokenERC20 := range setup.AssetERC20s {
			_, _, err = r.RegisterAssetERC20(tokenERC20, randomAddr1)
			require.Error(t, err)

			invalidContractErr := blockchain.InvalidContractError{}
			assert.True(t, errors.As(err, &invalidContractErr))
			assert.Equal(t, perunSymbol, invalidContractErr.Name)
		}
	})

	t.Run("re-register", func(t *testing.T) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		setup := ethereumtest.NewSimChainBackendSetup(t, rng, 2)
		perunSymbol := "PRN"

		r, err := ethereum.NewContractRegistry(setup.ChainBackend, setup.Adjudicator, setup.AssetETH)
		require.NoError(t, err)
		require.NotNil(t, r)

		require.Len(t, setup.AssetERC20s, 1, "setup should contain only one erc20 asset info")
		var tokenERC20, assetERC20 pwallet.Address
		var symbol string
		for tokenERC20, assetERC20 = range setup.AssetERC20s {
			// Register the token once.
			symbol, _, err = r.RegisterAssetERC20(tokenERC20, assetERC20)
			require.NoError(t, err)
			assert.Equal(t, perunSymbol, symbol)
		}
		t.Run("sameAsset_sameToken", func(t *testing.T) {
			_, _, err = r.RegisterAssetERC20(tokenERC20, assetERC20)
			require.Error(t, err)
			assetERC20RegisteredError := blockchain.AssetERC20RegisteredError{}
			require.True(t, errors.As(err, &assetERC20RegisteredError))
			assert.Equal(t, assetERC20.String(), assetERC20RegisteredError.Asset)
			assert.Equal(t, perunSymbol, assetERC20RegisteredError.Symbol)
		})
		// TODO: This test requires a second erc20 sample contract with a different symbol.
		// Add the test after another erc20 contract is added.
		// t.Run("sameAsset_differentToken", func(t *testing.T) {
		// })
	})
}
