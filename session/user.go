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

package session

import (
	"fmt"

	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node"
)

// User represents a participant in the off-chain network that uses a session on this node for sending transactions.
type User struct {
	perun.PeerID

	OnChain  perun.Credential // Account for funding the channel and the on-chain transactions.
	OffChain perun.Credential // Account (corresponding to off-chain address) used for signing authentication messages.

	// List of participant addresses for this user in each open channel.
	// OffChain credential is used for managing all these accounts.
	PartAddrs []pwallet.Address
}

// NewUnlockedUser initializes a user and unlocks onChaina and offChain accounts.
func NewUnlockedUser(wb perun.WalletBackend, cfg UserConfig) (User, perun.APIError) {
	var err error
	u := User{}

	onChainAddr, err := wb.ParseAddr(cfg.OnChainAddr)
	if err != nil {
		return User{}, perun.NewAPIErrInvalidConfig(err, "onChainAddr", cfg.OnChainAddr)
	}
	offChainAddr, err := wb.ParseAddr(cfg.OffChainAddr)
	if err != nil {
		return User{}, perun.NewAPIErrInvalidConfig(err, "offChainAddr", cfg.OffChainAddr)
	}
	if u.OnChain, err = newCred(wb, cfg.OnChainWallet, onChainAddr); err != nil {
		value := fmt.Sprintf("%s, %s", cfg.OnChainWallet.KeystorePath, cfg.OnChainWallet.Password)
		return User{}, perun.NewAPIErrInvalidConfig(err, "onChainWallet", value)
	}
	if u.OffChain, err = newCred(wb, cfg.OffChainWallet, offChainAddr); err != nil {
		value := fmt.Sprintf("%s, %s", cfg.OffChainWallet.KeystorePath, cfg.OffChainWallet.Password)
		return User{}, perun.NewAPIErrInvalidConfig(err, "offChainWallet", value)
	}

	u.PeerID.Alias = perun.OwnAlias
	u.PeerID.CommAddr = cfg.CommAddr
	u.PeerID.CommType = cfg.CommType
	u.PeerID.OffChainAddr = u.OffChain.Addr
	u.PeerID.OffChainAddrString = u.OffChain.Addr.String()

	return u, nil
}

// newCred initilizes the wallet and unlocks the account.
func newCred(wb perun.WalletBackend, cfg WalletConfig, addr pwallet.Address) (perun.Credential, error) {
	w, err := wb.NewWallet(cfg.KeystorePath, cfg.Password)
	if err != nil {
		return perun.Credential{}, err
	}
	_, err = wb.UnlockAccount(w, addr)
	if err != nil {
		return perun.Credential{}, err
	}
	return perun.Credential{
		Addr:     addr,
		Wallet:   w,
		Keystore: cfg.KeystorePath,
		Password: cfg.Password,
	}, nil
}
