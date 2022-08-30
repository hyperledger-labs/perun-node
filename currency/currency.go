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

package currency

import (
	"math/big"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// currency is the currency interpreter with the all the information required
// for parsing amount strings and printing amount.
type currency struct {
	symbol   string
	decimals decimal.Decimal
}

// Parse parses the amount string, converts it into base unit of the currency
// and returns it.
//
// Calculations are done using decimal library to ensure no loss of accuracy.
//
// It returns an error if the input string is not valid or if the absolute
// value of amount after converting to base unit is less than 1.
func (c currency) Parse(input string) (*big.Int, error) {
	amount, err := decimal.NewFromString(input)
	if err != nil {
		return nil, errors.Wrap(err, "invalid decimal string")
	}

	amountBaseUnit := amount.Mul(c.decimals)
	if !amountBaseUnit.IsZero() && amountBaseUnit.LessThan(decimal.NewFromInt(1)) {
		return nil, errors.New("amount is too small, should be larger than 1e-18")
	}
	return amountBaseUnit.BigInt(), nil
}

// Print returns the string representation after converting the amount given in
// base unit of the currency to its standard representation. Conversion is done
// by multiplying the amount in base unit with the number of decimal points for
// this currency.
//
// Calculations are done using decimal library to ensure no loss of accuracy.
func (c currency) Print(input *big.Int) string {
	amountBaseUnit := decimal.NewFromBigInt(input, 0)
	return amountBaseUnit.Div(c.decimals).String()
}

// Symbol returns the symbol for this currency.
func (c currency) Symbol() string { return c.symbol }
