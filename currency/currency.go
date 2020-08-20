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

	"github.com/hyperledger-labs/perun-node"
)

const (
	// ETH represents the ethereum currency.
	ETH              = "ETH"
	ether            = 1e18 // ether is the unit used for string representation of ETH.
	ethPlacesToRound = 6
)

var currencies map[string]perun.Currency

func init() {
	currencies = make(map[string]perun.Currency)

	ethMultiplier := decimal.NewFromFloat(ether)
	currencies[ETH] = ethParser{multiplier: ethMultiplier, placesToRound: ethPlacesToRound}
}

// IsSupported checks if there is parser regsitered for the currency
// represented by the given string.
func IsSupported(currency string) bool {
	p, ok := currencies[currency]
	return ok && p != nil
}

// NewParser returns the currency parser. It returns nil if unsupported currency is used.
// so check if exists before usage.
func NewParser(currency string) perun.Currency {
	return currencies[currency]
}

type ethParser struct {
	multiplier    decimal.Decimal
	placesToRound int32
}

// Parse parses the given currency string in Ether, converts it to Wei and returns a
// big.Int representation of the value.
// It can parse decimal values upto 1e-18 (equivalent of 1e-18 and the minimum value of
// the currency) and convert it to corresponding amount in Wei without loss of accuracy.
func (p ethParser) Parse(input string) (*big.Int, error) {
	amount, err := decimal.NewFromString(input)
	if err != nil {
		return nil, errors.Wrap(err, "invalid decimal string")
	}

	amountBaseUnit := amount.Mul(p.multiplier)
	if amountBaseUnit.LessThan(decimal.NewFromInt(1)) {
		return nil, errors.New("amount is too small, should be larger than 1e-18")
	}
	return amountBaseUnit.BigInt(), nil
}

// Print converts the input in Wei to Ether and returns a string representation of it.
// The returned string is rounded off to 6 decimal places for visual representation.
func (p ethParser) Print(input *big.Int) string {
	amount := decimal.NewFromBigInt(input, 0)
	return amount.Div(p.multiplier).StringFixedBank(p.placesToRound)
}
