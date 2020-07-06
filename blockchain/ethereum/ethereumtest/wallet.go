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
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/wallet"

	"github.com/direct-state-transfer/dst-go"
	"github.com/direct-state-transfer/dst-go/blockchain/ethereum/internal"
)

// NewTestWalletBackend initializes an ethereum specific wallet backend with weak encryption parameters.
func NewTestWalletBackend() dst.WalletBackend {
	return &internal.WalletBackend{EncParams: internal.ScryptParams{
		N: internal.WeakScryptN,
		P: internal.WeakScryptP,
	}}
}

// WalletSetup can generate any number of keys for testing. To enable faster unlocking of the keys, it uses
// weak encryption parameters for the storage encryption of the keys.
type WalletSetup struct {
	WalletBackend dst.WalletBackend
	KeystorePath  string
	Keystore      *keystore.KeyStore
	Wallet        wallet.Wallet
	Accs          []wallet.Account
}

// NewWalletSetup initializes a wallet with n accounts. Empty password string and weak encrytion parameters are used.
func NewWalletSetup(t *testing.T, rng *rand.Rand, n uint) *WalletSetup {
	wb := NewTestWalletBackend()

	ksPath, err := ioutil.TempDir("", "dst-go-test-keystore-*")
	require.NoErrorf(t, err, "Error creating temp directory for keystore: %v", err)
	ks := keystore.NewKeyStore(ksPath, internal.WeakScryptN, internal.WeakScryptP)
	w, err := ethwallet.NewWallet(ks, "")
	require.NoErrorf(t, err, "Error creating wallet: %v", err)

	accs := make([]wallet.Account, n)
	for idx := uint(0); idx < n; idx++ {
		accs[idx] = w.NewRandomAccount(rng)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(ksPath); err != nil {
			t.Log("error in cleanup - ", err)
		}
	})
	return &WalletSetup{
		WalletBackend: wb,
		KeystorePath:  ksPath,
		Keystore:      ks,
		Wallet:        w,
		Accs:          accs,
	}
}

// NewRandomAddress generates a random wallet address. It generates the address only as a byte array.
// Hence it does not generate any public or private keys corresponding to the address.
// If you need an address with keys, use Wallet.NewAccount method.
func NewRandomAddress(rnd *rand.Rand) wallet.Address {
	var a common.Address
	rnd.Read(a[:])
	return ethwallet.AsWalletAddr(a)
}
