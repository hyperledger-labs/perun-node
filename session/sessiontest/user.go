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

package sessiontest

import (
	"math/rand"
	"testing"

	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
)

// NewTestUser returns a test user with with on-chain, off-chain and n participant accounts.
func NewTestUser(t *testing.T, rng *rand.Rand, n uint) (perun.WalletBackend, perun.User) {
	ws := ethereumtest.NewWalletSetup(t, rng, 2+n)
	u := perun.User{}
	u.Alias = "test-user"

	u.OnChain.Addr = ws.Accs[0].Address()
	u.OnChain.Wallet = ws.Wallet
	u.OnChain.Keystore = ws.KeystorePath
	u.OnChain.Password = ""

	u.OffChain.Addr = ws.Accs[1].Address()
	u.OffChain.Wallet = ws.Wallet
	u.OffChain.Keystore = ws.KeystorePath
	u.OffChain.Password = ""
	u.OffChainAddr = ws.Accs[1].Address()

	u.PartAddrs = make([]pwallet.Address, n)
	for i := range ws.Accs[2:] {
		u.PartAddrs[i] = ws.Accs[i].Address()
	}
	return ws.WalletBackend, u
}
