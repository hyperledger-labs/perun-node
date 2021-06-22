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

package currency

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/hyperledger-labs/perun-node"
)

// Define symbol and max decimals for ETH, because there is no token contract
// for ETH, from which these details can be fetched from.
const (
	// ETHSymbol is the symbol for ethereum's native currency ETHSymbol.
	ETHSymbol = "ETH"

	// ETHMaxDecimals is the maximum number of decimal places allowed in ETH representation.
	ETHMaxDecimals uint8 = 18
)

// Registry implements a currency registry with currency parsers indexed by
// symbols.
//
// It uses a slice to keep track of registered symbols because iterating over
// map to retrieve the symbols each time will result in different ordering of
// symbols in the list.
type Registry struct {
	mtx        sync.RWMutex
	symbols    []string
	currencies map[string]perun.Currency
}

// NewRegistry initializes a currency registry.
func NewRegistry() *Registry {
	r := Registry{
		currencies: make(map[string]perun.Currency),
	}
	return &r
}

// Symbols returns a list of all the currencies registered in
// this module.
func (r *Registry) Symbols() []string {
	r.mtx.RLock()
	regsiteredSymbolsCopy := make([]string, len(r.symbols))
	copy(regsiteredSymbolsCopy, r.symbols)
	r.mtx.RUnlock()
	return regsiteredSymbolsCopy
}

// IsRegistered checks if there is parser registered for the currency
// represented by the given string.
func (r *Registry) IsRegistered(symbol string) bool {
	r.mtx.RLock()
	ok := r.isRegistered(symbol)
	r.mtx.RUnlock()
	return ok
}

func (r *Registry) isRegistered(symbol string) bool {
	p, ok := r.currencies[symbol]
	return ok && p != nil
}

// Register initializes a currency parser, registers it with the registry and
// returns it.
//
// Returns an error if the parser is already registered.
func (r *Registry) Register(symbol string, maxDecimals uint8) (perun.Currency, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.isRegistered(symbol) {
		return nil, errors.New("currency parser already registered for the given symbol")
	}
	c := currency{
		symbol:   symbol,
		decimals: decimal.New(1, int32(maxDecimals)),
	}
	r.currencies[symbol] = c
	r.symbols = append(r.symbols, symbol)
	return c, nil
}

// Currency returns the currency parser registered for the given currency
// symbol. If no parser is registered, it returns nil.
//
// So, caller should do a nil check before using the currency parser.
func (r *Registry) Currency(symbol string) perun.Currency {
	r.mtx.RLock()
	parser := r.currencies[symbol]
	r.mtx.RUnlock()
	return parser
}
