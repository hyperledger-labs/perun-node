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

package sessiontest_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	pethwallet "perun.network/go-perun/backend/ethereum/wallet"
)

// TestPrintPrivateKeys function is used to the print the private key of
// on-chain accounts for n users used in tests.
//
// These private keys are used for initializing ganache-cli node with pre-funded accounts.
func TestPrintPrivateKeys(t *testing.T) {
	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))

	count := uint(3)
	count *= 2
	ws, err := ethereumtest.NewWalletSetup(prng, count)
	require.NoError(t, err)
	password := ""
	fmt.Printf("\nNo       Address                                       Private Key\n\n")
	for i := uint(0); i < count; i = i + 2 {
		acc := ws.Accs[i]

		addr := common.Address(*(acc.Address()).(*pethwallet.Address))
		keyJSON, err := ws.Keystore.Export(accounts.Account{Address: addr}, password, password)
		require.NoError(t, err)

		key, err := keystore.DecryptKey(keyJSON, password)
		require.NoError(t, err)
		fmt.Printf("%d:\t 0x%s: 0x%X\n", (i+2)/2, acc.Address(), key.PrivateKey.D.Bytes())
	}
}
