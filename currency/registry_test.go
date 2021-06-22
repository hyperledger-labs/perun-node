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

package currency_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/currency"
)

func Test_Implements(t *testing.T) {
	assert.Implements(t, (*perun.ROCurrencyRegistry)(nil), new(currency.Registry))
	assert.Implements(t, (*perun.CurrencyRegistry)(nil), new(currency.Registry))
}

func Test_IsRegistered_NotNil_ETH(t *testing.T) {
	r := currency.NewRegistry()
	assert.False(t, r.IsRegistered(currency.ETHSymbol),
		"no currency should be registered by default")
}

func Test_Register_Symbols(t *testing.T) {
	r := currency.NewRegistry()
	wantRegisteredSymbols := []string{}

	tests := []struct {
		name        string
		symbol      string
		maxDecimals uint8
	}{
		{"ETH", currency.ETHSymbol, currency.ETHMaxDecimals},
		{"maxDecimals-0", "ERC20Test-1", 0},
		{"maxDecimals-18", "ERC20Test-2", 18},
		{"maxDecimals-255", "ERC20Test-3", 255},
	}

	for _, tt := range tests {
		t.Run("no-error-on-first-register"+tt.name, func(t *testing.T) {
			_, err := r.Register(tt.symbol, tt.maxDecimals)
			assert.NoError(t, err)

			wantRegisteredSymbols = append(wantRegisteredSymbols, tt.symbol)
			require.Equal(t, wantRegisteredSymbols, r.Symbols(), "should match with regsitered symbols")
		})
	}

	for _, tt := range tests {
		t.Run("error-on-re-register"+tt.name, func(t *testing.T) {
			_, err := r.Register(tt.symbol, tt.maxDecimals)
			assert.Error(t, err)
		})
	}

	t.Run("symbols return deep copy", func(t *testing.T) {
		gotRegisteredSymbols1 := r.Symbols()
		gotRegisteredSymbols2 := r.Symbols()
		gotRegisteredSymbols1[0] = ""
		require.NotEqual(t, gotRegisteredSymbols1, gotRegisteredSymbols2)
		require.Equal(t, gotRegisteredSymbols2, r.Symbols())
	})
}

func Test_Parser(t *testing.T) {
	r := currency.NewRegistry()
	_, err := r.Register(currency.ETHSymbol, currency.ETHMaxDecimals)
	assert.NoError(t, err)

	eth := r.Currency(currency.ETHSymbol)
	assert.NotNil(t, eth)
}
