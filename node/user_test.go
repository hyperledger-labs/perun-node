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

package node_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/direct-state-transfer/dst-go"
	"github.com/direct-state-transfer/dst-go/blockchain/ethereum/ethereumtest"
	"github.com/direct-state-transfer/dst-go/node"
	"github.com/direct-state-transfer/dst-go/node/nodetest"
)

func Test_New_Happy(t *testing.T) {
	rng := rand.New(rand.NewSource(1729))
	cntParts := uint(4)
	wb, testUser := nodetest.NewTestUser(t, rng, cntParts)
	userCfg := node.UserConfig{
		Alias:       testUser.Alias,
		OnChainAddr: testUser.OnChain.Addr.String(),
		OnChainWallet: node.WalletConfig{
			KeystorePath: testUser.OnChain.Keystore,
			Password:     "",
		},
		OffChainAddr: testUser.OffChain.Addr.String(),
		OffChainWallet: node.WalletConfig{
			KeystorePath: testUser.OffChain.Keystore,
			Password:     "",
		},
	}

	userCfg.PartAddrs = make([]string, len(testUser.PartAddrs))
	for i, addr := range testUser.PartAddrs {
		userCfg.PartAddrs[i] = addr.String()
	}

	gotUser, err := node.NewUnlockedUser(wb, userCfg)
	require.NoError(t, err)
	require.NotZero(t, gotUser)
	require.Len(t, gotUser.PartAddrs, int(cntParts))
}

func Test_New_Invalid_Parts(t *testing.T) {
	rng := rand.New(rand.NewSource(1729))
	cntParts := uint(1)
	wb, testUser := nodetest.NewTestUser(t, rng, cntParts)
	userCfg := node.UserConfig{
		Alias:       testUser.Alias,
		OnChainAddr: testUser.OnChain.Addr.String(),
		OnChainWallet: node.WalletConfig{
			KeystorePath: testUser.OnChain.Keystore,
			Password:     "",
		},
		OffChainAddr: testUser.OffChain.Addr.String(),
		OffChainWallet: node.WalletConfig{
			KeystorePath: testUser.OffChain.Keystore,
			Password:     "",
		},
	}

	t.Run("no_parts", func(t *testing.T) {
		gotUser, err := node.NewUnlockedUser(wb, node.UserConfig{})
		require.Error(t, err)
		require.Zero(t, gotUser)
	})
	t.Run("invalid_parts_address", func(t *testing.T) {
		userCfg.PartAddrs = make([]string, cntParts)
		for i := range testUser.PartAddrs {
			userCfg.PartAddrs[i] = "invalid-addr"
		}

		gotUser, err := node.NewUnlockedUser(wb, userCfg)
		require.Error(t, err)
		require.Zero(t, gotUser)
	})
	t.Run("missing_parts_address", func(t *testing.T) {
		userCfg.PartAddrs = make([]string, cntParts)
		for i := range testUser.PartAddrs {
			userCfg.PartAddrs[i] = ethereumtest.NewRandomAddress(rng).String()
		}
		gotUser, err := node.NewUnlockedUser(wb, userCfg)
		require.Error(t, err)
		require.Zero(t, gotUser)
	})
}

func Test_New_Invalid_Wallets(t *testing.T) {
	rng := rand.New(rand.NewSource(1729))
	wb, testUser := nodetest.NewTestUser(t, rng, 0)

	type args struct {
		wb  dst.WalletBackend
		cfg node.UserConfig
	}
	tests := []struct {
		name string
		args args
		want dst.User
	}{
		{
			name: "invalid_on-chain_address",
			args: args{
				wb: wb,
				cfg: node.UserConfig{
					Alias:       testUser.Alias,
					OnChainAddr: "invalid-addr",
					OnChainWallet: node.WalletConfig{
						KeystorePath: testUser.OnChain.Keystore,
						Password:     "",
					},
					OffChainAddr: testUser.OffChain.Addr.String(),
					OffChainWallet: node.WalletConfig{
						KeystorePath: testUser.OffChain.Keystore,
						Password:     "",
					},
				},
			},
		},
		{
			name: "invalid_off-chain_address",
			args: args{
				wb: wb,
				cfg: node.UserConfig{
					Alias:       testUser.Alias,
					OnChainAddr: testUser.OnChain.Addr.String(),
					OnChainWallet: node.WalletConfig{
						KeystorePath: testUser.OnChain.Keystore,
						Password:     "",
					},
					OffChainAddr: "invalid-addr",
					OffChainWallet: node.WalletConfig{
						KeystorePath: testUser.OffChain.Keystore,
						Password:     "",
					},
				},
			},
		},
		{
			name: "missing_on-chain_account",
			args: args{
				wb: wb,
				cfg: node.UserConfig{
					Alias:       testUser.Alias,
					OnChainAddr: ethereumtest.NewRandomAddress(rng).String(),
					OnChainWallet: node.WalletConfig{
						KeystorePath: testUser.OnChain.Keystore,
						Password:     "",
					},
					OffChainAddr: testUser.OffChain.Addr.String(),
					OffChainWallet: node.WalletConfig{
						KeystorePath: testUser.OffChain.Keystore,
						Password:     "",
					},
				},
			},
		},
		{
			name: "missing_off-chain_account",
			args: args{
				wb: wb,
				cfg: node.UserConfig{
					Alias:       testUser.Alias,
					OnChainAddr: testUser.OnChain.Addr.String(),
					OnChainWallet: node.WalletConfig{
						KeystorePath: testUser.OnChain.Keystore,
						Password:     "",
					},
					OffChainAddr: ethereumtest.NewRandomAddress(rng).String(),
					OffChainWallet: node.WalletConfig{
						KeystorePath: testUser.OffChain.Keystore,
						Password:     "",
					},
				},
			},
		},
		{
			name: "invalid_on-chain_password",
			args: args{
				wb: wb,
				cfg: node.UserConfig{
					Alias:       testUser.Alias,
					OnChainAddr: testUser.OnChain.Addr.String(),
					OnChainWallet: node.WalletConfig{
						KeystorePath: testUser.OnChain.Keystore,
						Password:     "invalid-password",
					},
					OffChainAddr: testUser.OffChain.Addr.String(),
					OffChainWallet: node.WalletConfig{
						KeystorePath: testUser.OffChain.Keystore,
						Password:     "",
					},
				},
			},
		},
		{
			name: "valid_on-chain_invalid_off-chain_password",
			args: args{
				wb: wb,
				cfg: node.UserConfig{
					Alias:       testUser.Alias,
					OnChainAddr: testUser.OnChain.Addr.String(),
					OnChainWallet: node.WalletConfig{
						KeystorePath: testUser.OnChain.Keystore,
						Password:     "",
					},
					OffChainAddr: testUser.OffChain.Addr.String(),
					OffChainWallet: node.WalletConfig{
						KeystorePath: testUser.OffChain.Keystore,
						Password:     "invalid-pwd",
					},
				},
			},
		},
		{
			name: "invalid_keystore_path",
			args: args{
				wb: wb,
				cfg: node.UserConfig{
					Alias:       testUser.Alias,
					OnChainAddr: testUser.OnChain.Addr.String(),
					OnChainWallet: node.WalletConfig{
						KeystorePath: "invalid-keystore-path",
						Password:     "",
					},
					OffChainAddr: testUser.OffChain.Addr.String(),
					OffChainWallet: node.WalletConfig{
						KeystorePath: testUser.OffChain.Keystore,
						Password:     "",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := node.NewUnlockedUser(tt.args.wb, tt.args.cfg)
			require.Error(t, err)
			assert.Zero(t, got)
		})
	}
}
