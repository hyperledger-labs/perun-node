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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/direct-state-transfer/dst-go/blockchain"
	"github.com/direct-state-transfer/dst-go/channel"
	"github.com/direct-state-transfer/dst-go/config"
	"github.com/direct-state-transfer/dst-go/identity"
	"github.com/direct-state-transfer/dst-go/log"
)

var logger log.LoggerInterface

// Config represents the configuration for this module.
type Config struct {
	Identity   identity.Config
	Channel    channel.Config
	Blockchain blockchain.Config

	Logger log.Config
}

// ConfigDefault represents the default configuration for this module.
var ConfigDefault = Config{
	Identity:   identity.ConfigDefault,
	Channel:    channel.ConfigDefault,
	Blockchain: blockchain.ConfigDefault,

	Logger: log.Config{
		Level:   log.InfoLevel,
		Backend: log.StdoutBackend,
	},
}

// GetFlagSet initializes and returns a flagset with flags for configuring this module.
func GetFlagSet() *pflag.FlagSet {
	var nodeMgrFlags pflag.FlagSet

	nodeMgrFlags.String(
		"programLogLevel", "", "Default log level for all modules")
	nodeMgrFlags.String(
		"programLogBackend", "", "Default log backend for all modules")

	return &nodeMgrFlags
}

func addFlagSet(cmd *cobra.Command) {
	cmd.PersistentFlags().AddFlagSet(GetFlagSet())
	cmd.PersistentFlags().AddFlagSet(identity.GetFlagSet())
	cmd.PersistentFlags().AddFlagSet(channel.GetFlagSet())
	cmd.PersistentFlags().AddFlagSet(blockchain.GetFlagSet())

}

func parseFlagSet(flagSet *pflag.FlagSet, nodeConfig *Config) (
	err error) {

	err = ParseFlags(flagSet, nodeConfig)
	if err != nil {
		return err
	}

	err = identity.ParseFlags(flagSet, &nodeConfig.Identity)
	if err != nil {
		return err
	}
	resolveLoggerConfig(&nodeConfig.Logger, &nodeConfig.Identity.Logger)

	err = channel.ParseFlags(flagSet, &nodeConfig.Channel)
	if err != nil {
		return err
	}
	resolveLoggerConfig(&nodeConfig.Logger, &nodeConfig.Channel.Logger)

	err = blockchain.ParseFlags(flagSet, &nodeConfig.Blockchain)
	if err != nil {
		return err
	}
	resolveLoggerConfig(&nodeConfig.Logger, &nodeConfig.Blockchain.Logger)

	return nil
}

// ParseFlags parses the flags defined in this module.
func ParseFlags(flagSet *pflag.FlagSet, nodeConfig *Config) (
	err error) {
	var flagsToParse = []config.FlagInfo{
		{Name: "programLogLevel", Ptr: &nodeConfig.Logger.Level},
		{Name: "programLogBackend", Ptr: &nodeConfig.Logger.Backend},
	}

	return config.LookUpMultiple(flagSet, flagsToParse)

}

//TODO : Find a way to resolve configuration
func resolveLoggerConfig(defaultCfg log.Configurer, moduleCfg log.Configurer) {

	// if moduleCfg.GetLogLevel() == log.Level() {
	// 	moduleCfg.SetLogLevel(defaultCfg.GetLogLevel())
	// }

	// if moduleCfg.GetLogBackend() == "" {
	// 	moduleCfg.SetLogBackend(defaultCfg.GetLogBackend())
	// }
}
