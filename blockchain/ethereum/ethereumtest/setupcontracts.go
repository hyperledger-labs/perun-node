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

package ethereumtest

import (
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	pethwallet "perun.network/go-perun/backend/ethereum/wallet"
	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
)

// Command to start the ganache-cli node:
//
// ganache-cli --account="0x1fedd636dbc7e8d41a0622a2040b86fea8842cef9d4aa4c582aad00465b7acff,10000000000000000000" \
//  --account="0xb0309c60b4622d3071fad3e16c2ce4d0b1e7758316c187754f4dd0cfb44ceb33,10000000000000000000"
//
// Ethereum address corresponding the above accounts: 0x8450c0055cB180C7C37A25866132A740b812937B and
// 0xc4bA4815c82727554e4c12A07a139b74c6742322.
//
// The account in the command corresponds to the on-chain account of first two
// users when seeding the rand source with "RandSeedForTestAccs" and passing
// numParts as 0. If numParts is not zero, then the on-chain account is funded
// only for the first user. Hence DO NOT CHANGE THE RAND SEED for integration
// tests in this package.
//
// The contracts will be deployed only during the first run of tests and will
// be resused in subsequent runs. This saves ~0.3s of setup time in each run.
// Hence when running tests on development machine, START THE NODE ONLY ONCE.

// ContractAddrs returns the contract addresses of adjudicator, asset ETH
// asset ERC20 contracts (along with the corresponding erc20 token contracts
// used in test setups. Address generation mechanism in ethereum is used to
// pre-compute the contract address.
//
// On a fresh ganache-cli node run the setup contracts helper function to
// deploy these contracts.
func ContractAddrs() (adjudicator, assetETH pwallet.Address, assetERC20s map[pwallet.Address]pwallet.Address) {
	// If not, then all of them must be nil, generate the addresses and return the values.
	// DO NOT SET the package level variable, it will be set by SetupContracts
	// function after the contracts are deployed to the blockchain.
	prng := rand.New(rand.NewSource(RandSeedForTestAccs))
	ws, err := NewWalletSetup(prng, 2)
	if err != nil {
		panic("Cannot setup test wallet")
	}
	adjudicator = pethwallet.AsWalletAddr(crypto.CreateAddress(pethwallet.AsEthAddr(ws.Accs[0].Address()), 0))
	assetETH = pethwallet.AsWalletAddr(crypto.CreateAddress(pethwallet.AsEthAddr(ws.Accs[0].Address()), 1))
	tokenERC20PRN := pethwallet.AsWalletAddr(crypto.CreateAddress(pethwallet.AsEthAddr(ws.Accs[0].Address()), 2))
	assetERC20PRN := pethwallet.AsWalletAddr(crypto.CreateAddress(pethwallet.AsEthAddr(ws.Accs[0].Address()), 3))
	assetERC20s = map[pwallet.Address]pwallet.Address{
		tokenERC20PRN: assetERC20PRN,
	}
	return
}

// SetupContractsT is the test friendly version of SetupContracts.
// It uses the passed testing.T to handle the errors and registers the cleanup
// functions on it.
func SetupContractsT(t *testing.T,
	chainURL string, chainID int, onChainTxTimeout time.Duration, incAssetERC20s bool,
) perun.ContractRegistry {
	contracts, err := SetupContracts(chainURL, chainID, onChainTxTimeout, incAssetERC20s)
	require.NoError(t, err)
	return contracts
}

// SetupContracts on its first invocation, deploys the contracts to blockchain
// and sets the addresses to package level variables.
// It checks first time invocation by checking if all any of the package level
// variables are nil.
//
// On every consequent calls, a contract registry will be initialized using
// these addresses in package level variables and will be returned directly.
//
// Every calls returns a new instance of contract registry, so that modifying
// the contract registry in one test does not affect other tests.
func SetupContracts(chainURL string, chainID int, onChainTxTimeout time.Duration, incAssetERC20s bool) (
	perun.ContractRegistry, error,
) {
	var err error

	if !isBlockchainRunning(chainURL) {
		return nil, errors.New("cannot connect to ganache-cli node at " + chainURL)
	}

	prng := rand.New(rand.NewSource(RandSeedForTestAccs))
	ws, err := NewWalletSetup(prng, 3) // Because accounts at index 0, 2 should be funded with PRN tokens.
	if err != nil {
		return nil, err
	}
	onChainCred := perun.Credential{
		Addr:     ws.Accs[0].Address(),
		Wallet:   ws.Wallet,
		Keystore: ws.KeystorePath,
		Password: "",
	}
	// If any of the values are nil, deploy contracts, set the package level
	// variables and return the addresses.
	chain, err := ethereum.NewChainBackend(chainURL, chainID, ChainConnTimeout, onChainTxTimeout, onChainCred)
	if err != nil {
		return nil, errors.WithMessage(err, "initializaing chain backend")
	}

	// If all the values are not nil, assume they are valid contract address
	// set during a previous invocation of this function and return the
	// contract addresses.
	// If not, then deploy the contracts and set the package level address variables.

	// If adjudicator is valid, then return the addresses directly.
	// If not, assume contracts have not been deployed, then deploy all of them.
	adjudicatorAddr, _, _ := ContractAddrs()
	if chain.ValidateAdjudicator(adjudicatorAddr) == nil {
		return newContractRegistry(chain, incAssetERC20s)
	}

	// Make a list of addresses by including all the accounts in wallet setup.
	// Each of these address will be assigned initBal amount of PRN tokens.
	initAccs := make([]pwallet.Address, len(ws.Accs))
	for i := range ws.Accs {
		initAccs[i] = ws.Accs[i].Address()
	}

	initBal := new(big.Int).Mul(big.NewInt(1e18), big.NewInt(1e3))
	err = deployContracts(chain, onChainCred, initAccs, initBal)
	if err != nil {
		return nil, err
	}

	return newContractRegistry(chain, incAssetERC20s)
}

func isBlockchainRunning(url string) bool {
	_, _, err := websocket.DefaultDialer.Dial(url, nil)
	return err == nil
}

func newContractRegistry(chain perun.ROChainBackend, incAssetERC20s bool) (perun.ContractRegistry, error) {
	adjudicatorAddr, assetETHAddr, assetERC20Addrs := ContractAddrs()
	contracts, err := ethereum.NewContractRegistry(chain, adjudicatorAddr, assetETHAddr)
	if err != nil {
		return nil, errors.WithMessage(err, "initializing contract registry")
	}
	if !incAssetERC20s {
		return contracts, nil
	}

	for tokenERC20, assetERC20 := range assetERC20Addrs {
		_, _, err = contracts.RegisterAssetERC20(tokenERC20, assetERC20)
		if err != nil {
			return nil, errors.WithMessagef(err, "registering erc20 token %v with asset %v", tokenERC20, assetERC20)
		}
	}
	return contracts, nil
}

func deployContracts(chain perun.ChainBackend, onChainCred perun.Credential,
	initAccs []pwallet.Address, initBal *big.Int,
) error {
	var err error
	adjudicator, err := chain.DeployAdjudicator(onChainCred.Addr)
	if err != nil {
		return errors.WithMessage(err, "deploying adjudicator")
	}
	_, err = chain.DeployAssetETH(adjudicator, onChainCred.Addr)
	if err != nil {
		return errors.WithMessage(err, "deploying asset ETH")
	}
	tokenERC20PRN, err := chain.DeployPerunToken(initAccs, initBal, onChainCred.Addr)
	if err != nil {
		return errors.WithMessage(err, "deploying perun token")
	}
	_, err = chain.DeployAssetERC20(adjudicator, tokenERC20PRN, onChainCred.Addr)
	if err != nil {
		return errors.WithMessage(err, "deploying asset ERC20")
	}
	return nil
}
