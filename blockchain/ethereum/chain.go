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

package ethereum

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	pethchannel "perun.network/go-perun/backend/ethereum/channel"
	pethwallet "perun.network/go-perun/backend/ethereum/wallet"
	pkeystore "perun.network/go-perun/backend/ethereum/wallet/keystore"
	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/internal"
)

// NewChainBackend initializes a connection to blockchain node and sets up a wallet with given credentials
// for funding on-chain transactions and channel balances.
//
// It uses the provided credentials to initialize a new keystore wallet.
//
// The function signature uses only types defined in the root package of this project and types from std lib.
// This enables the function to be loaded as symbol without importing this package when it is compiled as plugin.
func NewChainBackend(url string, chainConnTimeout, onChainTxTimeout time.Duration, cred perun.Credential) (
	perun.ChainBackend, error) {
	ctx, cancel := context.WithTimeout(context.Background(), chainConnTimeout)
	defer cancel()
	ethereumBackend, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to ethereum node at "+url)
	}

	ks := keystore.NewKeyStore(cred.Keystore, internal.StandardScryptN, internal.StandardScryptP)
	acc := accounts.Account{Address: pethwallet.AsEthAddr(cred.Addr)}
	if err = ks.Unlock(acc, cred.Password); err != nil {
		return nil, errors.Wrap(err, "unlocking on-chain keystore for addr - "+cred.Addr.String())
	}

	ksWallet, err := pkeystore.NewWallet(ks, cred.Password)
	if err != nil {
		return nil, err
	}
	tr := pkeystore.NewTransactor(*ksWallet)
	cb := pethchannel.NewContractBackend(ethereumBackend, tr)
	return &internal.ChainBackend{Cb: &cb, TxTimeout: onChainTxTimeout}, nil
}

// BalanceAt reads the on-chain balance of the given address.
func BalanceAt(url string, chainConnTimeout, onChainTxTimeout time.Duration, addr pwallet.Address) (
	*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), chainConnTimeout)
	defer cancel()
	ethereumBackend, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to ethereum node at "+url)
	}

	ctx, cancel = context.WithTimeout(context.Background(), onChainTxTimeout)
	defer cancel()
	bal, err := ethereumBackend.BalanceAt(ctx, pethwallet.AsEthAddr(addr), nil)
	if err != nil {
		return nil, errors.Wrap(err, "reading on-chain balance for "+addr.String())
	}
	return bal, nil
}
