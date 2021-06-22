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

package currency_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node/currency"
)

func Test_Currency_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  *big.Int
		wantErr bool
	}{
		{"happy_1", "0.5", big.NewInt(5e17), false},
		{"happy_2", "0.000000000000000005", big.NewInt(5), false},
		{"happy_3_exp_form", "5e-18", big.NewInt(5), false},
		{"happy_3_exp_form_upper_case", "5E-18", big.NewInt(5), false},

		{"err_too_small_exp_form", "5e-19", nil, true},
		{"err_too_small_exp_form_upper_case", "5E-19", nil, true},
		{"err_too_small", "0.0000000000000000005", nil, true},
		{"invalid_string", "invalid-currency-string", nil, true},
	}
	r := currency.NewRegistry()
	eth, err := r.Register(currency.ETHSymbol, currency.ETHMaxDecimals)
	require.NoError(t, err)
	require.Equal(t, currency.ETHSymbol, eth.Symbol())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eth.Parse(tt.input)
			if err != nil {
				t.Log(err)
			}
			require.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.output, got)
		})
	}
}

func Test_Currency_Print(t *testing.T) {
	tests := []struct {
		name   string
		input  *big.Int
		output string
	}{
		{"happy_1_whole_number", big.NewInt(5e18), "5"},
		{"happy_1_decimal", big.NewInt(5e17), "0.5"},
		{"happy_2_decimal", big.NewInt(12345678e10), "0.12345678"},
		{"happy_3_decimal", big.NewInt(87654321e10), "0.87654321"},
		{"happy_to_zero", big.NewInt(5), "0"},
	}
	r := currency.NewRegistry()
	eth, err := r.Register(currency.ETHSymbol, currency.ETHMaxDecimals)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eth.Print(tt.input)
			assert.Equal(t, tt.output, got)
		})
	}
}
