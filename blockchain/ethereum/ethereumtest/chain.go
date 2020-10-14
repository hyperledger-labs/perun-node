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

package ethereumtest

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/stretchr/testify/require"
	pethchannel "perun.network/go-perun/backend/ethereum/channel"
	pethchanneltest "perun.network/go-perun/backend/ethereum/channel/test"
	pethwallet "perun.network/go-perun/backend/ethereum/wallet"
	pkeystore "perun.network/go-perun/backend/ethereum/wallet/keystore"
	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/internal"
)

// ChainTxTimeout is the timeout for on-chain transactions.
// Since test setups are expected to run with a simulated backend or ganache node,
// a small timeout value is used.
const ChainTxTimeout = 1 * time.Minute

// ChainBackendSetup is a test setup that uses a simulated blockchain backend (for details on this backend,
// see go-ethereum) with required contracts deployed on it and a UserSetup.
type ChainBackendSetup struct {
	*WalletSetup
	ChainBackend       perun.ChainBackend
	AdjAddr, AssetAddr pwallet.Address
}

// NewChainBackendSetup returns a simulated contract backend with assetHolder and adjudicator contracts deployed.
// It also generates the given number of accounts and funds them each with 10 ether.
// and returns a test ChainBackend using the given randomness.
func NewChainBackendSetup(t *testing.T, rng *rand.Rand, numAccs uint) *ChainBackendSetup {
	walletSetup := NewWalletSetupT(t, rng, numAccs)

	cbEth := newSimContractBackend(t, walletSetup.Accs, walletSetup.Keystore)
	cb := &internal.ChainBackend{Cb: &cbEth, TxTimeout: ChainTxTimeout}

	onChainAddr := walletSetup.Accs[0].Address()
	adjudicator, err := cb.DeployAdjudicator(onChainAddr)
	require.NoError(t, err)
	asset, err := cb.DeployAsset(adjudicator, onChainAddr)
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
func newSimContractBackend(t *testing.T, accs []pwallet.Account, ks *keystore.KeyStore) pethchannel.ContractBackend {
	simBackend := pethchanneltest.NewSimulatedBackend()
	ctx, cancel := context.WithTimeout(context.Background(), ChainTxTimeout)
	defer cancel()
	for _, acc := range accs {
		simBackend.FundAddress(ctx, pethwallet.AsEthAddr(acc.Address()))
	}

	ksWallet, err := pkeystore.NewWallet(ks, "") // Password for test accounts is always empty string.
	require.NoError(t, err)

	tr := pkeystore.NewTransactor(*ksWallet)
	return pethchannel.NewContractBackend(simBackend, tr)
}
