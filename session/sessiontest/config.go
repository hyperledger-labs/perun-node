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
	"fmt"
	"math/rand"
	"testing"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/session"
)

// NewUserConfig returns a test user configuration with random data generated using the given rng.
// It creates "n" participant accounts to the user.
func NewUserConfig(t *testing.T, rng *rand.Rand, n uint) (perun.WalletBackend, session.UserConfig) {
	ws, userCfg := newUserConfig(t, rng, 2+n)
	return ws.WalletBackend, userCfg
}

func newUserConfig(t *testing.T, rng *rand.Rand, n uint) (*ethereumtest.WalletSetup, session.UserConfig) {
	ws := ethereumtest.NewWalletSetup(t, rng, 2+n)
	port, err := freeport.GetFreePort()
	require.NoError(t, err)

	cfg := session.UserConfig{}
	cfg.Alias = "test-user"
	cfg.OnChainAddr = ws.Accs[0].Address().String()
	cfg.OnChainWallet = session.WalletConfig{
		KeystorePath: ws.KeystorePath,
		Password:     "",
	}
	cfg.OffChainAddr = ws.Accs[1].Address().String()
	cfg.OffChainWallet = session.WalletConfig{
		KeystorePath: ws.KeystorePath,
		Password:     "",
	}
	cfg.PartAddrs = make([]string, n)
	for i := range ws.Accs[2:] {
		cfg.PartAddrs[i] = ws.Accs[i].Address().String()
	}
	cfg.CommType = "tcp"
	cfg.CommAddr = fmt.Sprintf("127.0.0.1:%d", port)
	return ws, cfg
}
