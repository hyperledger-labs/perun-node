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

package internal

import (
	"context"
	"time"

	"github.com/pkg/errors"
	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// ChainBackend provides ethereum specific contract backend functionality.
type ChainBackend struct {
	// Cb is the instance of contract backend that will be used for all on-chain communications.
	Cb *ethchannel.ContractBackend
	// TxTimeout is the max time to wait for confirmation of transactions on blockchain.
	// If this expires, a transactions is considered failed.
	// Use sufficiently large values when connecting to mainnet.
	TxTimeout time.Duration
}

// NewFunder initializes and returns an instance of ethereum funder.
func (cb *ChainBackend) NewFunder(assetAddr wallet.Address) channel.Funder {
	return ethchannel.NewETHFunder(*cb.Cb, ethwallet.AsEthAddr(assetAddr))
}

// NewAdjudicator initializes and returns an instance of ethereum adjudicator.
func (cb *ChainBackend) NewAdjudicator(adjAddr, receiverAddr wallet.Address) channel.Adjudicator {
	return ethchannel.NewAdjudicator(*cb.Cb, ethwallet.AsEthAddr(adjAddr), ethwallet.AsEthAddr(receiverAddr))
}

// ValidateContracts validates the integrity of given adjudicator and asset holder contracts.
func (cb *ChainBackend) ValidateContracts(adjAddr, assetAddr wallet.Address) error {
	ctx, cancel := context.WithTimeout(context.Background(), cb.TxTimeout)
	defer cancel()

	// Integrity of Adjudicator is implicitly done during validation of asset holder contract.
	err := ethchannel.ValidateAssetHolderETH(ctx, *cb.Cb, ethwallet.AsEthAddr(assetAddr), ethwallet.AsEthAddr(adjAddr))
	if ethchannel.IsContractBytecodeError(err) {
		return errors.Wrap(err, "invalid contracts at given addresses")
	}
	return errors.Wrap(err, "validating contracts")
}

// DeployAdjudicator deploys the adjudicator contract.
func (cb *ChainBackend) DeployAdjudicator() (wallet.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cb.TxTimeout)
	defer cancel()
	addr, err := ethchannel.DeployAdjudicator(ctx, *cb.Cb)
	return ethwallet.AsWalletAddr(addr), errors.Wrap(err, "deploying adjudicator contract")
}

// DeployAsset deploys the asset holder contract, setting the adjudicator address to given value.
func (cb *ChainBackend) DeployAsset(adjAddr wallet.Address) (wallet.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cb.TxTimeout)
	defer cancel()

	addr, err := ethchannel.DeployETHAssetholder(ctx, *cb.Cb, ethwallet.AsEthAddr(adjAddr))
	return ethwallet.AsWalletAddr(addr), errors.Wrap(err, "deploying asset contract")
}
