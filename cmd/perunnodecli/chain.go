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

package main

import (
	"github.com/abiosoft/ishell"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/currency/currencytest"
)

var (
	aliceOnChainAddr = "0x8450c0055cB180C7C37A25866132A740b812937B"
	bobOnChainAddr   = "0x8946Ee6a3Ba1AD6CF79faaa31F4D4aBC17A9424b"
	chainAddr        string

	chainCmdUsage = "Usage: chain [sub-command]"
	chainCmd      = &ishell.Cmd{
		Name: "chain",
		Help: "Use this command to functionalities related to direct interaction with blockchain" + chainCmdUsage,
		Func: chainFn,
	}

	setChainAddrCmdUsage = "Usage: set-blockchain-address [address]."
	setChainAddrCmd      = &ishell.Cmd{
		Name: "set-blockchain-address",
		Help: "Set address of the blockchain node for reading balance." + setChainAddrCmdUsage,
		Completer: func([]string) []string {
			return []string{"ws://127.0.0.1:8545"} // Provide default values as autocompletion.
		},
		Func: setChainAddrFn,
	}

	getChainAddrCmdUsage = "Usage: get-blockchain-address."
	getChainAddrCmd      = &ishell.Cmd{
		Name: "get-blockchain-address",
		Help: "Get address of the blockchain node for reading balance." + getChainAddrCmdUsage,
		Func: getChainAddrFn,
	}

	deployPerunContractsCmdUsage = "Usage: deploy-perun-contracts."
	deployPerunContractsCmd      = &ishell.Cmd{
		Name: "deploy-perun-contracts",
		Help: "Deploy perun smart contracts." + deployPerunContractsCmdUsage,
		Func: deployPerunContractsFn,
	}

	getOnChainBalanceCmdUsage = "Usage: deploy-perun-contracts."
	getOnChainBalanceCmd      = &ishell.Cmd{
		Name: "get-on-chain-balance",
		Help: "Get on-chain balance." + getOnChainBalanceCmdUsage,
		Completer: func([]string) []string {
			return []string{aliceOnChainAddr, bobOnChainAddr}
		},
		Func: getOnChainBalanceFn,
	}

	// Max length of the amount string representing on-chain balance.
	// Digits after this will be trucated.
	// This is assuming, a maximum of 3 digits before decimal points which
	// allows upto 6 digits after the decimal place to be shown.
	amountMaxLength = 10
	// This currency parser is used to parse on-chain ETH balances.
	ethCurrency perun.Currency
)

func init() {
	chainCmd.AddCmd(setChainAddrCmd)
	chainCmd.AddCmd(getChainAddrCmd)
	chainCmd.AddCmd(deployPerunContractsCmd)
	chainCmd.AddCmd(getOnChainBalanceCmd)

	// Registry in currencytest has all currency parsers used in tests
	// pre-registered.
	ethCurrency = currencytest.Registry().Currency(currency.ETHSymbol)
}

func chainFn(c *ishell.Context) {
	c.Println(c.Cmd.HelpText())
}

func setChainAddrFn(c *ishell.Context) {
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}
	chainAddr = c.Args[0]

	c.Printf("%s\n\n", greenf("Blockchain node address: %s\n", chainAddr))
}

func getChainAddrFn(c *ishell.Context) {
	countReqArgs := 0
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	c.Printf("%s\n\n", greenf("Blockchain node address: %s\n", chainAddr))
}

func deployPerunContractsFn(c *ishell.Context) {
	countReqArgs := 0
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}
	if chainAddr == "" {
		c.Printf("%s\n\n", redf("Chain address is not set, set using below command %s", setChainAddrCmdUsage))
		return
	}

	contracts, err := ethereumtest.SetupContracts(chainAddr, ethereumtest.ChainID, ethereumtest.OnChainTxTimeout, false)
	if err != nil {
		c.Printf("%s\n\n", redf("Error deploying contracts: %v", err))
		return
	}
	c.Printf("%s\n\n", greenf("Contracts deployed successfully.\nAdjudicator address: %v\nAsset ETH address: %v\n",
		contracts.Adjudicator(), contracts.AssetETH()))
}

func getOnChainBalanceFn(c *ishell.Context) {
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}
	if chainAddr == "" {
		c.Printf("%s\n\n", redf("Chain address is not set, set using below command %s", setChainAddrCmdUsage))
		return
	}

	addr, err := ethereum.NewWalletBackend().ParseAddr(c.Args[0])
	if err != nil {
		c.Printf("%s\n\n", redf("Error parsing address: %v", err))
		return
	}

	bal, err := ethereum.BalanceAt(chainAddr, ethereumtest.ChainConnTimeout, ethereumtest.OnChainTxTimeout, addr)
	if err != nil {
		c.Printf("%s\n\n", redf("Error connecting to blockchain node: %v", err))
		return
	}

	ethAmount := ethCurrency.Print(bal)
	if len(ethAmount) > amountMaxLength {
		ethAmount = ethAmount[:amountMaxLength]
	}
	c.Printf("%s\n\n", greenf("On-chain balance: %s", ethAmount))
}
