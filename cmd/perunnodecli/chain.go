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

	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
)

var (
	chainCmdUsage = "Usage: chain [sub-command]"
	chainCmd      = &ishell.Cmd{
		Name: "chain",
		Help: "Use this command to functionalities related to direct interaction with blockchain" + chainCmdUsage,
		Func: chainFn,
	}

	deployPerunContractsCmdUsage = "Usage: deploy-perun-contracts."
	deployPerunContractsCmd      = &ishell.Cmd{
		Name: "deploy-perun-contracts",
		Help: "Deploy perun smart contracts." + deployPerunContractsCmdUsage,
		Completer: func([]string) []string {
			return []string{"ws://127.0.0.1:8545"} // Provide default values as autocompletion.
		},
		Func: deployPerunContractsFn,
	}
)

func init() {
	chainCmd.AddCmd(deployPerunContractsCmd)
}

func chainFn(c *ishell.Context) {
	c.Println(c.Cmd.HelpText())
}

func deployPerunContractsFn(c *ishell.Context) {
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	adjudicator, asset, err := ethereumtest.SetupContracts(c.Args[0], ethereumtest.OnChainTxTimeout)
	if err != nil {
		c.Printf("%s\n\n", redf("Error deploying contracts: %v", err))
	}
	c.Printf("%s\n\n", greenf("Contracts deployed successfully.\nAdjudicator address: %v\nAsset address: %v\n",
		adjudicator, asset))
}
