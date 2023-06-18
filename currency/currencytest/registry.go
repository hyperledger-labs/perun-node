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

package currencytest

import (
	"github.com/hyperledger-labs/perun-node/currency"
)

var r *currency.Registry

// Registry returns a currency registry for use in tests with all the symbols
// used in the tests already registered.
//
// This is intended for use in tests in session, because in actual
// implementation, this is initialized by the node and passed onto the session.
func Registry() *currency.Registry {
	if r != nil {
		return r
	}
	r = currency.NewRegistry()
	//nolint: errcheck		// Registering currencies on new registry will not fail.
	r.Register(currency.ETHSymbol, currency.ETHMaxDecimals)
	return r
}
