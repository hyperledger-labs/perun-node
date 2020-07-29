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

package session

import (
	"github.com/pkg/errors"
	"perun.network/go-perun/wallet"

	"github.com/direct-state-transfer/perun-node"
)

// NewUnlockedUser initializes a user and unlocks all the accounts,
// those corresponding to on-chain address, off-chain address and all participant addresses.
func NewUnlockedUser(wb perun.WalletBackend, cfg UserConfig) (perun.User, error) {
	var err error
	u := perun.User{}

	if u.OnChain.Wallet, err = newWallet(wb, cfg.OnChainWallet, cfg.OnChainAddr); err != nil {
		return perun.User{}, errors.WithMessage(err, "on-chain wallet")
	}
	if u.OffChain.Wallet, err = newWallet(wb, cfg.OffChainWallet, cfg.OffChainAddr); err != nil {
		return perun.User{}, errors.WithMessage(err, "off-chain wallet")
	}
	if u.PartAddrs, err = parseUnlock(wb, u.OffChain.Wallet, cfg.PartAddrs...); err != nil {
		return perun.User{}, errors.WithMessage(err, "participant addresses")
	}
	u.OffChainAddr = u.OffChain.Addr
	u.CommAddr = cfg.CommAddr
	u.CommType = cfg.CommType

	return u, nil
}

// newWallet initializes the wallet using the wallet backend and unlocks accounts corresponding
// to the given address.
func newWallet(wb perun.WalletBackend, cfg WalletConfig, addr string) (wallet.Wallet, error) {
	w, err := wb.NewWallet(cfg.KeystorePath, cfg.Password)
	if err != nil {
		return nil, err
	}
	_, err = parseUnlock(wb, w, addr)
	return w, err
}

// parseUnlock parses the given addresses string using the wallet backend and unlocks accounts
// corresponding to each of the given addresses.
func parseUnlock(wb perun.WalletBackend, w wallet.Wallet, addrs ...string) ([]wallet.Address, error) {
	var err error
	parsedAddrs := make([]wallet.Address, len(addrs))
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
