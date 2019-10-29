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

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/direct-state-transfer/dst-go/blockchain"
	"github.com/direct-state-transfer/dst-go/channel"
	"github.com/direct-state-transfer/dst-go/ethereum/adapter"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
	"github.com/direct-state-transfer/dst-go/log"
)

var packageName = "dst-go"

// BlockchainConn is the blockchain connection used by users in this node.
var BlockchainConn adapter.ContractBackend

// LibSignAddr is the address of libSignatures contract that can be used by users in this node.
var LibSignAddr types.Address

func main() {

	app := &cobra.Command{
		Use: "dst-go",
		Run: nodeInit,
	}
	addFlagSet(app)
	err := app.Execute()

	if err != nil {
		logger.Error("Error running dst-go -", err)
	}

}
func nodeInit(cmd *cobra.Command, args []string) {

	err := parseFlagSet(cmd.Flags(), &ConfigDefault)
	if err != nil {
		fmt.Println("error parsing configuration\n,", err)
		return
	}

	logger, err = log.NewLogger(ConfigDefault.Logger.Level, ConfigDefault.Logger.Backend, packageName)
	if err != nil {
		fmt.Println("error initializing logger\n", err)
		return
	}

	err = identity.InitModule(&ConfigDefault.Identity)
	if err != nil {
		logger.Error("error initializing identity module -", err)
		return
	}

	err = channel.InitModule(&ConfigDefault.Channel)
	if err != nil {
		logger.Error("error initializing channel module -", err)
		return
	}

	BlockchainConn, LibSignAddr, err = blockchain.InitModule(&ConfigDefault.Blockchain)
	if err != nil {
		logger.Error("error initializing blockchain module -", err)
		return
	}
	if err == nil {
		logger.Info("Node successfully initialized")
	}

	//TODO : Exit signal from keyboard interrupt and other modules
	for {
		logger.Error("The event handler is yet to be implemented")
	}
}
