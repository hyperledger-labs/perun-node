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
	"context"
	"time"

	"github.com/pkg/errors"
	pethchannel "perun.network/go-perun/backend/ethereum/channel"
	pethwallet "perun.network/go-perun/backend/ethereum/wallet"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
)

// ChainBackend provides ethereum specific contract backend functionality.
type ChainBackend struct {
	// Cb is the instance of contract backend that will be used for all on-chain communications.
	Cb *pethchannel.ContractBackend
	// TxTimeout is the max time to wait for confirmation of transactions on blockchain.
	// If this expires, a transactions is considered failed.
	// Use sufficiently large values when connecting to mainnet.
	TxTimeout time.Duration
}

// NewFunder initializes and returns an instance of ethereum funder.
func (cb *ChainBackend) NewFunder(assetAddr pwallet.Address) pchannel.Funder {
	return pethchannel.NewETHFunder(*cb.Cb, pethwallet.AsEthAddr(assetAddr))
}

// NewAdjudicator initializes and returns an instance of ethereum adjudicator.
func (cb *ChainBackend) NewAdjudicator(adjAddr, receiverAddr pwallet.Address) pchannel.Adjudicator {
	return pethchannel.NewAdjudicator(*cb.Cb, pethwallet.AsEthAddr(adjAddr), pethwallet.AsEthAddr(receiverAddr))
}

// ValidateContracts validates the integrity of given adjudicator and asset holder contracts.
func (cb *ChainBackend) ValidateContracts(adjAddr, assetAddr pwallet.Address) error {
	ctx, cancel := context.WithTimeout(context.Background(), cb.TxTimeout)
	defer cancel()

	// Integrity of Adjudicator is implicitly done during validation of asset holder contract.
	err := pethchannel.ValidateAssetHolderETH(ctx, *cb.Cb, pethwallet.AsEthAddr(assetAddr), pethwallet.AsEthAddr(adjAddr))
	if pethchannel.IsContractBytecodeError(err) {
		return errors.Wrap(err, "invalid contracts at given addresses")
	}
	return errors.Wrap(err, "validating contracts")
}

// DeployAdjudicator deploys the adjudicator contract.
func (cb *ChainBackend) DeployAdjudicator() (pwallet.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cb.TxTimeout)
	defer cancel()
	addr, err := pethchannel.DeployAdjudicator(ctx, *cb.Cb)
	return pethwallet.AsWalletAddr(addr), errors.Wrap(err, "deploying adjudicator contract")
}

// DeployAsset deploys the asset holder contract, setting the adjudicator address to given value.
func (cb *ChainBackend) DeployAsset(adjAddr pwallet.Address) (pwallet.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cb.TxTimeout)
	defer cancel()

	addr, err := pethchannel.DeployETHAssetholder(ctx, *cb.Cb, pethwallet.AsEthAddr(adjAddr))
	return pethwallet.AsWalletAddr(addr), errors.Wrap(err, "deploying asset contract")
}
