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

package session_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/session"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

func Test_New_Happy(t *testing.T) {
	rng := rand.New(rand.NewSource(1729))
	cntParts := uint(4)
	wb, userCfg := sessiontest.NewUserConfig(t, rng, cntParts)

	t.Run("non_empty_parts", func(t *testing.T) {
		userCfgCopy := userCfg
		gotUser, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.NoError(t, err)
		compareUserWithCfg(t, gotUser, userCfgCopy)
	})

	t.Run("empty_parts", func(t *testing.T) {
		userCfgCopy := userCfg
		userCfgCopy.PartAddrs = nil
		gotUser, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.NoError(t, err)
		compareUserWithCfg(t, gotUser, userCfgCopy)
	})
}

func compareUserWithCfg(t *testing.T, gotUser perun.User, userCfg session.UserConfig) {
	require.NotZero(t, gotUser)

	assert.Equal(t, perun.OwnAlias, gotUser.Alias)

	require.NotNil(t, gotUser.OffChainAddr)
	assert.Equal(t, userCfg.OffChainAddr, gotUser.OffChain.Addr.String())

	require.NotNil(t, gotUser.OnChain.Addr)
	assert.Equal(t, userCfg.OnChainAddr, gotUser.OnChain.Addr.String())

	assert.Equal(t, userCfg.CommAddr, gotUser.CommAddr)
	assert.Equal(t, userCfg.CommType, gotUser.CommType)
	require.Len(t, gotUser.PartAddrs, len(userCfg.PartAddrs))
}

func Test_New_Unhappy(t *testing.T) {
	rng := rand.New(rand.NewSource(1729))
	cntParts := uint(1)
	wb, userCfg := sessiontest.NewUserConfig(t, rng, cntParts)

	t.Run("invalid_parts_address", func(t *testing.T) {
		userCfgCopy := userCfg
		userCfgCopy.PartAddrs[0] = "invalid-address"
		_, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.Error(t, err)
	})
	t.Run("missing_parts_address", func(t *testing.T) {
		userCfgCopy := userCfg
		userCfgCopy.PartAddrs[0] = ethereumtest.NewRandomAddress(rng).String()
		_, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.Error(t, err)
	})
	t.Run("invalid_on-chain_address", func(t *testing.T) {
		userCfgCopy := userCfg
		userCfgCopy.OnChainAddr = "invalid-addr"
		_, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.Error(t, err)
	})
	t.Run("invalid_off-chain_address", func(t *testing.T) {
		userCfgCopy := userCfg
		userCfgCopy.OffChainAddr = "invalid-addr"
		_, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.Error(t, err)
	})
	t.Run("missing_on-chain_address", func(t *testing.T) {
		userCfgCopy := userCfg
		userCfgCopy.OnChainAddr = ethereumtest.NewRandomAddress(rng).String()
		_, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.Error(t, err)
	})
	t.Run("missing_off-chain_address", func(t *testing.T) {
		userCfgCopy := userCfg
		userCfgCopy.OffChainAddr = ethereumtest.NewRandomAddress(rng).String()
		_, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.Error(t, err)
	})
	t.Run("invalid_on-chain_password", func(t *testing.T) {
		userCfgCopy := userCfg
		userCfgCopy.OnChainWallet.Password = "invalid-password"
		_, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.Error(t, err)
	})
	t.Run("invalid_off-chain_password", func(t *testing.T) {
		userCfgCopy := userCfg
		userCfgCopy.OffChainWallet.Password = "invalid-password"
		_, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.Error(t, err)
	})
	t.Run("invalid_on-chain_keystore", func(t *testing.T) {
		userCfgCopy := userCfg
		userCfgCopy.OnChainWallet.KeystorePath = "invalid-keystore"
		_, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.Error(t, err)
	})
	t.Run("invalid_off-chain_keystore", func(t *testing.T) {
		userCfgCopy := userCfg
		userCfgCopy.OffChainWallet.KeystorePath = "invalid-keystore"
		_, err := session.NewUnlockedUser(wb, userCfgCopy)
		require.Error(t, err)
	})
}
