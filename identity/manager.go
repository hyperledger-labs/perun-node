// Copyright (c) 2019 - for information on the respective copyright owner
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

package identity

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/direct-state-transfer/dst-go/ethereum/keystore"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/log"
)

var packageName = "identity"

// OffChainID represents the offchain identity of a user.
type OffChainID struct {
	OnChainID        types.Address `json:"on_chain_id"`       // On chain address
	ListenerIPAddr   string        `json:"listener_ip_addr"`  // URL with port number for off chain transactions
	ListenerEndpoint string        `json:"listener_endpoint"` // URL Endpoint for offchain transactions

	KeyStore *keystore.KeyStore `json:"-"` // Optional, set when using identity for signing
	Password string             `json:"-"` // Optional, set when using identity for signing
}

// String implements fmt.Stringer interface.
func (id OffChainID) String() string {
	return fmt.Sprintf("OnChainID :0x%x, ListenerIPAddr %s", id.OnChainID, id.ListenerIPAddr+id.ListenerEndpoint)
}

// ListenerLocalAddr returns the URL on which channel listener should be started.
func (id *OffChainID) ListenerLocalAddr() (listeningAddr string, err error) {

	//Retreive the port number from complete URL
	stringsList := strings.Split(id.ListenerIPAddr, ":")
	if len(stringsList) != 2 {
		return listeningAddr, fmt.Errorf("Invalid offchain address format, missing \":port number\" - %s", id.ListenerIPAddr)
	}

	//Replace ip address with "localhost" in the URL
	listeningAddr = "localhost:" + stringsList[1]
	return listeningAddr, nil
}

// SetCredentials sets the password and keystore on the offchain id.
func (id *OffChainID) SetCredentials(ks *keystore.KeyStore, password string) (success bool) {

	if ks == nil {
		return false
	}
	id.KeyStore = ks
	id.Password = password
	return true
}

// GetCredentials fetches the password and keystore on the offchain id.
func (id *OffChainID) GetCredentials() (ks *keystore.KeyStore, password string, available bool) {
	if id.KeyStore == nil {
		return nil, "", false
	}
	return id.KeyStore, id.Password, true
}

// ClearCredentials clears the password and keystore on the offchain id.
func (id *OffChainID) ClearCredentials() {
	id.KeyStore = nil
	id.Password = ""
}

// OffChainIDStore manages an identity storage file on disk.
type OffChainIDStore struct {
	Filename string
	IDList   []OffChainID
	IDMap    map[types.Address]OffChainID
}

// InitModule initializes the identity module.
func InitModule(cfg *Config) (err error) {

	logger, err = log.NewLogger(cfg.Logger.Level, cfg.Logger.Backend, packageName)
	return err
}

// NewSession initializes and returns instances to manage keystore and offchain identity store.
func NewSession(keysDir, idFile string) (keyStore *keystore.KeyStore, idStore *OffChainIDStore, err error) {

	//ScryptN and ScryptP parameters for algorithm to encrypt keys stored on disk.
	//The are used only when creating new keys and do not have any effect with reading already stored keys.
	keyStore = keystore.NewKeyStore(keysDir, keystore.StandardScryptN, keystore.StandardScryptP)
	if len(keyStore.Accounts()) == 0 {
		err = fmt.Errorf("No keys found in keys directory -%s", keysDir)
		return keyStore, idStore, err
	}

	idStore, err = NewOffChainIDStore(idFile)
	if err != nil {
		return keyStore, idStore, err
	}

	logger.Info("Initialized identity module with a keystore")
	logger.Info("Found", len(keyStore.Accounts()), "keys in keystore")
	logger.Info("Found", len(idStore.IDList), "ids in known ids file")

	return keyStore, idStore, err
}

// NewOffChainIDStore creates a new instance of offchain identity store from file on disk.
func NewOffChainIDStore(filepath string) (idstore *OffChainIDStore, err error) {

	idstore = &OffChainIDStore{
		Filename: filepath,
		IDMap:    make(map[types.Address]OffChainID),
		IDList:   []OffChainID{},
	}

	err = idstore.Update()
	if err != nil {
		idstore = &OffChainIDStore{}
	}
	return idstore, err
}

// Update refreshes the offchain id store updating any new changes in the file on disk.
func (idStore *OffChainIDStore) Update() (err error) {
	idStoreByteArr, err := ioutil.ReadFile(idStore.Filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(idStoreByteArr, &idStore.IDList)
	if err != nil {
		return err
	}

	for _, id := range idStore.IDList {
		idStore.IDMap[id.OnChainID] = id
	}
	return nil
}

// OffChainID fetches the offchain id corresponding to the onchain id.
func (idStore *OffChainIDStore) OffChainID(onChainID types.Address) (id OffChainID, present bool) {
	id, present = idStore.IDMap[onChainID]
	return id, present
}

// Equal returns true if the two offchain identity values are equal.
// Equality is checked in terms of onchain id and listener address for offchain transactions.
func Equal(a, b OffChainID) bool {
	if a.OnChainID == b.OnChainID &&
		a.ListenerIPAddr == b.ListenerIPAddr {
		return true
	}
	return false
}

// GetKey fetches the key corresponding to the onchain id from the keystore if the password is correct.
func GetKey(ks *keystore.KeyStore, onChainID types.Address, password string) (
	key *keystore.Key, err error) {
	return ks.GetKey(onChainID, password)
}

// NewKeystore initializes keystore using standard scrypt params (for encrypting keys when storing on disk).
func NewKeystore(ksDir string) *keystore.KeyStore {

	return keystore.NewKeyStore(ksDir, keystore.StandardScryptN, keystore.StandardScryptP)
}

// IsKeysPresent checks if the keys corresponding to provided onchain ids are present in the keystore.
func IsKeysPresent(ks *keystore.KeyStore, requiredAccounts []types.Address) (missingAccounts []types.Address) {

	missingAccounts = []types.Address{}
	for _, account := range requiredAccounts {
		if !ks.HasAddress(account.Address) {
			missingAccounts = append(missingAccounts, account)
		}
	}
	return missingAccounts
}
