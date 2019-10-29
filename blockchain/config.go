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

package blockchain

import (
	"math/big"

	"github.com/spf13/pflag"

	"github.com/direct-state-transfer/dst-go/config"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/log"
)

var logger log.LoggerInterface

// SetLogger sets the logger instance for this module.
func SetLogger(moduleLogger log.LoggerInterface) {
	logger = moduleLogger
}

// Config represents the configuration for this module.
type Config struct {
	//Logger config - Do not specify default value
	//will be inherited from node manager
	Logger log.Config

	libSignaturesAddr types.Address

	gethURL   string
	networkID *big.Int
}

// ConfigDefault represents the default configuration for this module.
var ConfigDefault = Config{
	gethURL: "ws://localhost:8546",
}

// GetFlagSet initializes and returns a flagset with flags for configuring this module.
func GetFlagSet() *pflag.FlagSet {

	var bcFlags pflag.FlagSet

	bcFlags.String("blockchainLogLevel", "", "Log level for blockchain module")
	bcFlags.String("blockchainLogBackend", "", "Log Backend for blockchain module")
	bcFlags.String("libSignAddr", "", "Lib signatures address for node")
	bcFlags.String("gethURL", "", "Geth node URL for connection")

	return &bcFlags
}

// ParseFlags parses the flags defined in this module.
func ParseFlags(flagSet *pflag.FlagSet, cfg *Config) error {

	var flagsToParse = []config.FlagInfo{
		{Name: "blockchainLogLevel", Ptr: &cfg.Logger.Level},
		{Name: "blockchainLogBackend", Ptr: &cfg.Logger.Backend},
		{Name: "libSignAddr", Ptr: &cfg.libSignaturesAddr},
		{Name: "gethURL", Ptr: &cfg.gethURL},
	}
	return config.LookUpMultiple(flagSet, flagsToParse)

}
