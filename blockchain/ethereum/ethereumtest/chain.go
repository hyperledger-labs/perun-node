// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/dst-go
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

package ethereumtest

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/stretchr/testify/require"
	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	ethchanneltest "perun.network/go-perun/backend/ethereum/channel/test"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/wallet"

	"github.com/direct-state-transfer/dst-go"
	"github.com/direct-state-transfer/dst-go/blockchain/ethereum/internal"
)

// ChainTxTimeout is the timeout for on-chain transactions.
// Since test setups are expected to run with a simulated backend or ganache node,
// a small timeout value is used.
const ChainTxTimeout = 1 * time.Minute

// ChainBackendSetup is a test setup that uses a simulated blockchain backend (for details on this backend,
// see go-ethereum) with required contracts deployed on it and a UserSetup.
type ChainBackendSetup struct {
	*WalletSetup
	ChainBackend       dst.ChainBackend
	AdjAddr, AssetAddr wallet.Address
}

// NewChainBackendSetup returns a simulated contract backend with assetHolder and adjudicator contracts deployed.
// It also generates the given number of accounts and funds them each with 10 ether.
// and returns a test ChainBackend using the given randomness.
func NewChainBackendSetup(t *testing.T, rng *rand.Rand, numAccs uint) *ChainBackendSetup {
	walletSetup := NewWalletSetup(t, rng, numAccs)

	cbEth := newSimContractBackend(walletSetup.Accs, walletSetup.Keystore)
	cb := &internal.ChainBackend{Cb: &cbEth, TxTimeout: ChainTxTimeout}

	adjudicator, err := cb.DeployAdjudicator()
	require.NoError(t, err)
	asset, err := cb.DeployAsset(adjudicator)
	require.NoError(t, err)

	// No cleanup required.
	return &ChainBackendSetup{
		WalletSetup:  walletSetup,
		ChainBackend: cb,
		AdjAddr:      adjudicator,
		AssetAddr:    asset,
	}
}

// newSimContractBackend sets up a simulated contract backend with the first entry (index 0) in accs
// as the user account. All accounts are funded with 10 ethers.
func newSimContractBackend(accs []wallet.Account, ks *keystore.KeyStore) ethchannel.ContractBackend {
	simBackend := ethchanneltest.NewSimulatedBackend()
	ctx, cancel := context.WithTimeout(context.Background(), ChainTxTimeout)
	defer cancel()
	for _, acc := range accs {
		simBackend.FundAddress(ctx, ethwallet.AsEthAddr(acc.Address()))
	}

	onChainAcc := &accs[0].(*ethwallet.Account).Account
	return ethchannel.NewContractBackend(simBackend, ks, onChainAcc)
}
