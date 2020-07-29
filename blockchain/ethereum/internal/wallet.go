// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/perun-node
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

package internal

import (
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/wallet"
)

// Standard encryption parameters should be uses for real wallets. Using these parameters will
// cause the decryption to use 256MB of RAM and takes approx 1s on a modern processor.
//
// Weak encryption parameters should be used for test wallets. Using these parameters will
// cause the can be decrypted and unlocked faster.
const (
	StandardScryptN = keystore.StandardScryptN
	StandardScryptP = keystore.StandardScryptP
	WeakScryptN     = 2
	WeakScryptP     = 1
)

// WalletBackend provides ethereum specific wallet backend functionality.
type WalletBackend struct {
	EncParams ScryptParams
}

// ScryptParams defines the parameters for scrypt encryption algorithm, used or storage encryption of keys.
//
// Weak values should be used only for testing purposes (enables faster unlockcing). Use standard values otherwise.
type ScryptParams struct {
	N, P int
}

// NewWallet initializes an ethereum keystore at the given path and checks if all the keys in the keystore can
// be unlocked with the given password.
func (wb *WalletBackend) NewWallet(keystorePath, password string) (wallet.Wallet, error) {
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		return nil, errors.Wrap(err, "initializing new wallet, cannot find keystore directory")
	}
	ks := keystore.NewKeyStore(keystorePath, wb.EncParams.N, wb.EncParams.P)
	w, err := ethwallet.NewWallet(ks, password)
	return w, errors.Wrap(err, "initializing new wallet")
}

// UnlockAccount retrieves the account corresponding to the given address, unlocks and returns it.
func (wb *WalletBackend) UnlockAccount(w wallet.Wallet, addr wallet.Address) (wallet.Account, error) {
	acc, err := w.Unlock(addr)
	return acc, errors.Wrap(err, "unlocking account")
}

// ParseAddr parses the ethereum address from the given string. It be the hexadecimal
// representation of the address, optionally prefixed by "0x".
// It can be all upper or all lower or mixed case. All of them will produce identical
// result.
func (wb *WalletBackend) ParseAddr(str string) (wallet.Address, error) {
	addr := ethwallet.AsWalletAddr(common.HexToAddress(str))

	// common.HexToAddress parses even invalid strings to zero value of the address type.
	// So return an error when addr has zero value and the input string is not a valid
	// zero value representation of the address type. Valid zero value representations are
	// "", "0x", "0x00000" (any number of zeros) or the canonical form of forty zeros.
	zeroValue := ethwallet.Address{}
	if addr.Equals(&zeroValue) && !strings.Contains(zeroValue.String(), str) {
		return nil, errors.New("parsing address")
	}
	return addr, nil
}
