// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
//     https://github.com/direct-state-transfer/dst-go/NOTICE
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

package keystore

import (
	"io/ioutil"

	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/ethereum/go-ethereum/accounts"
	ethereumKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
)

// Standard parameter values for N,P in Scrypt algorithm used for encrypting keys stored on disk.
var (
	StandardScryptN = ethereumKeystore.StandardScryptN
	StandardScryptP = ethereumKeystore.StandardScryptP
)

// Account wraps the Account type defined in go-ethereum/accounts.
// It consists of onchain address and url to locate the key file in the backend.
type Account struct {
	accounts.Account
}

// MakeAccount creates and returns and ethereum account (with empty url) from ethereum onchain id.
func MakeAccount(ethAddr types.Address) Account {
	return Account{Account: accounts.Account{Address: ethAddr.Address}}
}

// KeyStore wraps the ethereum keystore that is used to manage a keys storage directory on disk.
type KeyStore struct {
	*ethereumKeystore.KeyStore
}

// Key wraps the Key type from ethereumKeystore. It stores the private key in plain text and ethereum address corresponding to it.
type Key struct {
	*ethereumKeystore.Key
}

// NewKeyStore initializes and returns an new instance of ethereum keystore.
func NewKeyStore(keysDir string, scryptN, scryptP int) *KeyStore {
	return &KeyStore{
		ethereumKeystore.NewKeyStore(keysDir, scryptN, scryptP),
	}
}

// GetKey returns the decrypted private key corresponding to the ethereum address if the password is correct.
// In case the password is wrong or key not found in keystore, an error is returned.
func (ks *KeyStore) GetKey(ethAddr types.Address, password string) (
	*Key, error) {

	ethAccount, err := ks.Find(MakeAccount(ethAddr))
	if err != nil {
		logger.Error("Ethereum Account not found", err)
		return nil, err
	}

	keyjson, err := ioutil.ReadFile(ethAccount.URL.Path)
	if err != nil {
		logger.Error("Account key file not found", err)
		return nil, err
	}
	keyEth, err := ethereumKeystore.DecryptKey(keyjson, password)
	if err != nil {
		logger.Error("Error in decrypting key", err)
		return nil, err
	}

	return &Key{keyEth}, nil
}

// Find retrieves and returns an unique account corresponding to the ethereum address from the keystore.
// If no account or multiple matching accounts are found, an error is returned.
func (ks *KeyStore) Find(a Account) (Account, error) {
	accEth, err := ks.KeyStore.Find(a.Account)
	return Account{accEth}, err
}

// SignHashWithPassphrase generates a signatures over the hash, provided correct passphrase is provided.
func (ks *KeyStore) SignHashWithPassphrase(a Account, passphrase string, hash []byte) (signature []byte, err error) {
	return ks.KeyStore.SignHashWithPassphrase(a.Account, passphrase, hash)
}
