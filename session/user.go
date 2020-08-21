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
	"github.com/pkg/errors"
	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node"
)

// NewUnlockedUser initializes a user and unlocks all the accounts,
// those corresponding to on-chain address, off-chain address and all participant addresses.
func NewUnlockedUser(wb perun.WalletBackend, cfg UserConfig) (perun.User, error) {
	var err error
	u := perun.User{}

	if u.OnChain, err = newCred(wb, cfg.OnChainWallet, cfg.OnChainAddr); err != nil {
		return perun.User{}, errors.WithMessage(err, "on-chain wallet")
	}
	if u.OffChain, err = newCred(wb, cfg.OffChainWallet, cfg.OffChainAddr); err != nil {
		return perun.User{}, errors.WithMessage(err, "off-chain wallet")
	}
	if u.PartAddrs, err = parseUnlock(wb, u.OffChain.Wallet, cfg.PartAddrs...); err != nil {
		return perun.User{}, errors.WithMessage(err, "participant addresses")
	}

	u.Peer.Alias = perun.OwnAlias
	u.Peer.CommAddr = cfg.CommAddr
	u.Peer.CommType = cfg.CommType
	u.Peer.OffChainAddr = u.OffChain.Addr
	u.Peer.OffChainAddrString = u.OffChain.Addr.String()

	return u, nil
}

// newCred initializes the wallet using the wallet backend, unlocks the accounts and
// returns the credential to access the account specified in the config.
func newCred(wb perun.WalletBackend, cfg WalletConfig, addr string) (perun.Credential, error) {
	w, err := wb.NewWallet(cfg.KeystorePath, cfg.Password)
	if err != nil {
		return perun.Credential{}, err
	}
	addrs, err := parseUnlock(wb, w, addr)
	if err != nil {
		return perun.Credential{}, err
	}
	return perun.Credential{
		Addr:     addrs[0],
		Wallet:   w,
		Keystore: cfg.KeystorePath,
		Password: cfg.Password,
	}, nil
}

// parseUnlock parses the given addresses string using the wallet backend and unlocks accounts
// corresponding to each of the given addresses.
func parseUnlock(wb perun.WalletBackend, w pwallet.Wallet, addrs ...string) ([]pwallet.Address, error) {
	var err error
	parsedAddrs := make([]pwallet.Address, len(addrs))
	for i, addr := range addrs {
		if parsedAddrs[i], err = wb.ParseAddr(addr); err != nil {
			return nil, err
		}
		if _, err = w.Unlock(parsedAddrs[i]); err != nil {
			return nil, err
		}
	}
	return parsedAddrs, nil
}
