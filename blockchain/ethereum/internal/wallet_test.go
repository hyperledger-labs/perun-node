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

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/direct-state-transfer/perun-node"
	"github.com/direct-state-transfer/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/direct-state-transfer/perun-node/blockchain/ethereum/internal"
)

func Test_WalletBackend_Interface(t *testing.T) {
	assert.Implements(t, (*perun.WalletBackend)(nil), new(internal.WalletBackend))
}

func Test_WalletBackend_NewWallet(t *testing.T) {
	rng := rand.New(rand.NewSource(1729))
	wb := ethereumtest.NewTestWalletBackend()
	setup := ethereumtest.NewWalletSetup(t, rng, 1)

	t.Run("happy", func(t *testing.T) {
		w, err := wb.NewWallet(setup.KeystorePath, "")
		assert.NoError(t, err)
		assert.NotNil(t, w)
	})
	t.Run("invalid_pwd", func(t *testing.T) {
		w, err := wb.NewWallet(setup.KeystorePath, "invalid-pwd")
		assert.Error(t, err)
		assert.Nil(t, w)
	})
	t.Run("invalid_keystore_path", func(t *testing.T) {
		w, err := wb.NewWallet("invalid-ks-path", "")
		assert.Error(t, err)
		assert.Nil(t, w)
	})
}

func Test_WalletBackend_NewAccount(t *testing.T) {
	rng := rand.New(rand.NewSource(1729))
	wb := ethereumtest.NewTestWalletBackend()
	setup := ethereumtest.NewWalletSetup(t, rng, 1)

	t.Run("happy", func(t *testing.T) {
		w, err := wb.UnlockAccount(setup.Wallet, setup.Accs[0].Address())
		assert.NoError(t, err)
		assert.NotNil(t, w)
	})
	t.Run("multiple calls", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			w, err := wb.UnlockAccount(setup.Wallet, setup.Accs[0].Address())
			assert.NoError(t, err)
			assert.NotNil(t, w)
		}
	})
	t.Run("account_not_present", func(t *testing.T) {
		randomAddr := ethereumtest.NewRandomAddress(rng)
		w, err := wb.UnlockAccount(setup.Wallet, randomAddr)
		assert.Error(t, err)
		assert.Nil(t, w)
	})
}

func Test_WalletBackend_ParseAddr(t *testing.T) {
	rng := rand.New(rand.NewSource(1729))
	wb := ethereumtest.NewTestWalletBackend()

	t.Run("happy_non_zero_value", func(t *testing.T) {
		validAddr := ethereumtest.NewRandomAddress(rng)
		gotAddr, err := wb.ParseAddr(validAddr.String())
		assert.NoError(t, err)
		require.NotNil(t, gotAddr)
		assert.True(t, validAddr.Equals(gotAddr))
	})
	t.Run("happy_zero_value", func(t *testing.T) {
		validAddr := ethwallet.Address{}
		gotAddr, err := wb.ParseAddr(validAddr.String())
		assert.NoError(t, err)
		require.NotNil(t, gotAddr)
		assert.True(t, validAddr.Equals(gotAddr))
	})
	t.Run("invalid_addr", func(t *testing.T) {
		gotAddr, err := wb.ParseAddr("invalid-addr")
		assert.Error(t, err)
		require.Nil(t, gotAddr)
	})

	t.Run("fixed_data", func(t *testing.T) {
		tests := []struct {
			name   string
			input  string
			output string
		}{
			{"lower_case", "0x931d387731bbbc988b312206c74f77d004d6b84b", "0x931D387731bBbC988B312206c74F77D004D6B84b"},
			{"upper_case", "0X931D387731BBBC988B312206C74F77D004D6B84B", "0x931D387731bBbC988B312206c74F77D004D6B84b"},
			{"mixed_case", "0X931D387731bbbc988b312206c74f77d004d6b84b", "0x931D387731bBbC988B312206c74F77D004D6B84b"},
			{"no_prefix", "931d387731bbbc988b312206c74f77d004d6b84b", "0x931D387731bBbC988B312206c74F77D004D6B84b"},
			{"zero_addr_1", "", "0x0000000000000000000000000000000000000000"},
			{"zero_addr_2", "0x", "0x0000000000000000000000000000000000000000"},
			{"zero_addr_3", "0x00000000", "0x0000000000000000000000000000000000000000"},
			{"zero_addr_4", "00000000", "0x0000000000000000000000000000000000000000"},
			{"zero_addr_5", "0x0000000000000000000000000000000000000000", "0x0000000000000000000000000000000000000000"},
			{"odd_no_of_chars", "0xbd5465321", "0x0000000000000000000000000000000Bd5465321"},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				gotAddr, err := wb.ParseAddr(test.input)
				assert.NoError(t, err)
				require.NotNil(t, gotAddr)
				assert.Equal(t, test.output, gotAddr.String())
			})
		}
	})
}
