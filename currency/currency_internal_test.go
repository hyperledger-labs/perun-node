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
	"testing"

	"github.com/stretchr/testify/assert"
)

// internal tests only check for errors is init() function in currency.go.
// see external tests file for other tests.
func Test_Exists_NewParser(t *testing.T) {
	t.Run("Err_Missing", func(t *testing.T) {
		assert.False(t, IsSupported("missing_parser_for_test"))
		assert.Nil(t, NewParser("missing_parser_for_test"))
	})

	t.Run("Err_Exists_but_nil", func(t *testing.T) {
		testCurrency := "nil_parser_for_test"
		currencies[testCurrency] = nil
		t.Cleanup(func() {
			delete(currencies, testCurrency)
		})

		assert.False(t, IsSupported(testCurrency))
		assert.Nil(t, NewParser("missing_parser_for_test"))
	})
}
