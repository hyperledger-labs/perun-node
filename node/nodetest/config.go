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

package nodetest

import (
	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

// NewConfig generates random configuration data for the node.
//
// Contract addressses and chain parameters are fetched from the ethereumtest package.
func NewConfig() perun.NodeConfig {
	adjudicator, assetETH, assetERC20s := ethereumtest.ContractAddrs()

	assetERC20sString := make(map[string]string)
	for tokenERC20, assetERC20 := range assetERC20s {
		assetERC20sString[tokenERC20.String()] = assetERC20.String()
	}

	return perun.NodeConfig{
		LogFile:              "",
		LogLevel:             "debug",
		ChainURL:             ethereumtest.ChainURL,
		ChainID:              ethereumtest.ChainID,
		Adjudicator:          adjudicator.String(),
		AssetETH:             assetETH.String(),
		AssetERC20s:          assetERC20sString,
		CommTypes:            []string{"tcp"},
		IDProviderTypes:      []string{"local"},
		CurrencyInterpreters: []string{"ETH"},

		ChainConnTimeout: ethereumtest.ChainConnTimeout,
		OnChainTxTimeout: ethereumtest.OnChainTxTimeout,
		ResponseTimeout:  sessiontest.ResponseTimeout,
	}
}
