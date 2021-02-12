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

package internal

import (
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	pethwallet "perun.network/go-perun/backend/ethereum/wallet"
	pkeystore "perun.network/go-perun/backend/ethereum/wallet/keystore"
	pwallet "perun.network/go-perun/wallet"
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

// Length of a hex representation (excluding '0x' prefix) of a valid address in canonical form.
const addressLengthHex = 40

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
func (wb *WalletBackend) NewWallet(keystorePath, password string) (pwallet.Wallet, error) {
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		return nil, errors.Wrap(err, "initializing new wallet, cannot find keystore directory")
	}
	ks := keystore.NewKeyStore(keystorePath, wb.EncParams.N, wb.EncParams.P)
	w, err := pkeystore.NewWallet(ks, password)
	return w, errors.Wrap(err, "initializing new wallet")
}

// UnlockAccount retrieves the account corresponding to the given address, unlocks and returns it.
func (wb *WalletBackend) UnlockAccount(w pwallet.Wallet, addr pwallet.Address) (pwallet.Account, error) {
	acc, err := w.Unlock(addr)
	return acc, errors.Wrap(err, "unlocking account")
}

// ParseAddr parses the ethereum address from the given string. It should be in hexadecimal
// representation of the address, optionally prefixed by "0x" or "0X".
// It pads zeros in the beginning if the address string is less than required length and
// returns an error if it is greater than required length.
func (wb *WalletBackend) ParseAddr(str string) (pwallet.Address, error) {
	// If address string is longer than address length, return an error.
	hasPrefix := strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X")
	if !hasPrefix && len(str) > addressLengthHex || len(str) > addressLengthHex+2 {
		return nil, errors.New("hex string too long, should be <= 40 chars")
	}

	addr := pethwallet.AsWalletAddr(common.HexToAddress(str))

	// common.HexToAddress parses even invalid strings to zero value of the address type.
	// So return an error when addr has zero value and the input string is not a valid
	// zero value representation of the address type. Valid zero value representations are
	// "", "0x", "0x00000" (any number of zeros) or the canonical form of forty zeros.
	zeroValue := pethwallet.Address{}
	if addr.Equals(&zeroValue) && !strings.Contains(zeroValue.String(), str) {
		return nil, errors.New("invalid string")
	}
	return addr, nil
}
