// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/dst-go
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
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/direct-state-transfer/perun-node"
	"github.com/direct-state-transfer/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/direct-state-transfer/perun-node/blockchain/ethereum/internal"
)

func Test_ChainBackend_Interface(t *testing.T) {
	assert.Implements(t, (*perun.ChainBackend)(nil), new(internal.ChainBackend))
}

func Test_ChainBackend_Deploy(t *testing.T) {
	rng := rand.New(rand.NewSource(1729))
	setup := ethereumtest.NewChainBackendSetup(t, rng, 1)

	adjAddr, err := setup.ChainBackend.DeployAdjudicator()
	require.NoError(t, err)
	assetAddr, err := setup.ChainBackend.DeployAsset(adjAddr)
	require.NoError(t, err)
	assert.NoError(t, setup.ChainBackend.ValidateContracts(adjAddr, assetAddr))
}

func Test_ChainBackend_ValidateContracts(t *testing.T) {
	rng := rand.New(rand.NewSource(1729))
	setup := ethereumtest.NewChainBackendSetup(t, rng, 1)

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, setup.ChainBackend.ValidateContracts(setup.AdjAddr, setup.AssetAddr))
	})
	t.Run("invalid_random_addrs", func(t *testing.T) {
		randomAddr1 := ethereumtest.NewRandomAddress(rng)
		randomAddr2 := ethereumtest.NewRandomAddress(rng)
		assert.Error(t, setup.ChainBackend.ValidateContracts(randomAddr1, randomAddr2))
	})
}
