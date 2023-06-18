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
	"math/rand"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	pethwallet "perun.network/go-perun/backend/ethereum/wallet"
	pkswallet "perun.network/go-perun/backend/ethereum/wallet/keystore"
	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/internal"
)

// NewTestWalletBackend initializes an ethereum specific wallet backend with weak encryption parameters.
func NewTestWalletBackend() perun.WalletBackend {
	return &internal.WalletBackend{EncParams: internal.ScryptParams{
		N: internal.WeakScryptN,
		P: internal.WeakScryptP,
	}}
}

// WalletSetup can generate any number of keys for testing. To enable faster unlocking of the keys, it uses
// weak encryption parameters for the storage encryption of the keys.
type WalletSetup struct {
	WalletBackend perun.WalletBackend
	KeystorePath  string
	Keystore      *keystore.KeyStore
	Wallet        pwallet.Wallet
	Accs          []pwallet.Account
}

// NewWalletSetupT is the test friendly version of NewWalletSetup.
// It uses the passed testing.T to handle the errors and registers the cleanup functions on it.
func NewWalletSetupT(t *testing.T, rng *rand.Rand, n uint) *WalletSetup {
	ws, err := NewWalletSetup(rng, n)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := os.RemoveAll(ws.KeystorePath); err != nil {
			t.Log("error in cleanup - ", err)
		}
	})
	return ws
}

// NewWalletSetup initializes a wallet with n accounts. Empty password string and weak encrytion parameters are used.
func NewWalletSetup(rng *rand.Rand, n uint) (*WalletSetup, error) {
	wb := NewTestWalletBackend()

	ksPath, err := os.MkdirTemp("", "perun-node-test-keystore-*")
	if err != nil {
		return nil, errors.Wrap(err, "creating temp directory for keystore")
	}
	ks := keystore.NewKeyStore(ksPath, internal.WeakScryptN, internal.WeakScryptP)
	w, err := pkswallet.NewWallet(ks, "")
	if err != nil {
		os.RemoveAll(ksPath) //nolint:errcheck
		return nil, errors.Wrap(err, "creating creating wallet")
	}

	accs := make([]pwallet.Account, n)
	for idx := uint(0); idx < n; idx++ {
		accs[idx] = w.NewRandomAccount(rng)
	}
	return &WalletSetup{
		WalletBackend: wb,
		KeystorePath:  ksPath,
		Keystore:      ks,
		Wallet:        w,
		Accs:          accs,
	}, nil
}

// NewRandomAddress generates a random wallet address. It generates the address only as a byte array.
// Hence it does not generate any public or private keys corresponding to the address.
// If you need an address with keys, use Wallet.NewAccount method.
func NewRandomAddress(rnd *rand.Rand) pwallet.Address {
	var a common.Address
	rnd.Read(a[:])
	return pethwallet.AsWalletAddr(a)
}
