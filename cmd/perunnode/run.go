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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc"
	"github.com/hyperledger-labs/perun-node/node"
)

const (
	// flag names for run command.
	loglevelF         = "loglevel"
	logfileF          = "logfile"
	chainurlF         = "chainurl"
	adjudicatorF      = "adjudicator"
	assetF            = "asset"
	chainconntimeoutF = "chainconntimeout"
	onchaintxtimeoutF = "onchaintxtimeout"
	responsetimeoutF  = "responsetimeout"
	configfileF       = "configfile" // can only be specified in flag, not via config file.
	grpcPortF         = "grpcport"   // can only be specified in flag, not via config file.

	// default values for flags in run command.
	defaultConfigFile = "node.yaml"
	defaultGrpcPort   = 50001
)

var (
	// Viper instance for parsing node configuration file. Each flag in the nodeCfgFlags list (that are defined
	// on the run command) will also be attached to the viper instance, so that the values from flags (when
	// specified), override the values defined in the configuration files.
	nodeCfgViper *viper.Viper

	// Flags corresponding to node configuration parameters. Each of this flag can individually override the
	// default values in config file. Also, the node configuration can be fully specified by using all of these
	// flags, in which case no config file
	// is needed and configFile flag can be unspecified.
	nodeCfgFlags = []string{
		logfileF,
		loglevelF,
		chainurlF,
		adjudicatorF,
		assetF,
		chainconntimeoutF,
		onchaintxtimeoutF,
		responsetimeoutF,
	}

	// List of supported adapters by the node for the respective components.
	// Currently this is fixed and hence hard coded here.
	// It can be moved to config file or flags at the point when the user will
	// be able to choose (when starting the node) which ones to load or support.
	supportedCommTypes             = []string{"tcp"}
	supportedContactTypes          = []string{"yaml"}
	supportedCurrencyInterpretters = []string{"ETH"}
)

func init() {
	rootCmd.AddCommand(runCmd)
	defineFlags()

	nodeCfgViper = viper.New()

	// Bind the configuration flags to viper instance,
	// values in flags (when specified), takes precedence over those in config file.
	var err error
	for i := range nodeCfgFlags {
		if err = nodeCfgViper.BindPFlag(nodeCfgFlags[i], runCmd.Flags().Lookup(nodeCfgFlags[i])); err != nil {
			panic(err)
		}
	}
}

func defineFlags() {
	runCmd.Flags().String(configfileF, defaultConfigFile, "node config file")
	runCmd.Flags().Uint64(grpcPortF, defaultGrpcPort, "port for grpc payment channel API server to listen")

	// All these flags should have zero values for defaults, as their only purpose is allow	the user to
	// explicitly specify the configuration.
	runCmd.Flags().String(loglevelF, "", "Log level. Supported levels: debug, info, error")
	runCmd.Flags().String(logfileF, "", "Log file path. Use empty string for stdout")
	runCmd.Flags().String(chainurlF, "", "URL of the blockchain node")
	runCmd.Flags().String(adjudicatorF, "", "Address as of the adjudicator contract as hex string with 0x prefix")
	runCmd.Flags().String(assetF, "", "Address as of the asset contract as hex string with 0x prefix")
	runCmd.Flags().Duration(chainconntimeoutF, time.Duration(0),
		"Connection timeout for connecting to the blockchain node")
	runCmd.Flags().Duration(onchaintxtimeoutF, time.Duration(0),
		"Max duration to wait for an on-chain transaction to be mined.")
	runCmd.Flags().Duration(responsetimeoutF, time.Duration(0),
		"Max duration to wait for a response in off-chain communication.")
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the perunnode",
	Long: `Start the perun node. Currently, the node serves the payment API via grpc.

Configuration can be specified in the config file or via flags. Values in the
flags override that in the config file.

If no flags are specified, default path for config file is used. However, if
all the config flags are specified, config file is ignored.`,
	Run: run,
}

func run(cmd *cobra.Command, args []string) {
	nodeCfg := parseNodeConfig(cmd.LocalNonPersistentFlags(), nodeCfgViper)
	grpcPort, err := cmd.Flags().GetUint64(grpcPortF)
	if err != nil {
		panic("unknown flag port\n")
	}
	grpcAddr := fmt.Sprintf(":%d", grpcPort)

	nodeAPI, err := node.New(nodeCfg)
	if err != nil {
		fmt.Printf("Error initializing nodeAPI: %v\n", err)
		return
	}

	fmt.Printf("Running perun node with the below config:\n%s.\n\nServing payment channel API via grpc at port %s\n\n",
		prettify(nodeCfg), grpcAddr)
	if err := grpc.ListenAndServePayChAPI(nodeAPI, grpcAddr); err != nil {
		fmt.Printf("Server returned with error: %v\n", err)
	}
}

func parseNodeConfig(fs *pflag.FlagSet, v *viper.Viper) perun.NodeConfig {
	// Ignore config file, if all config flags are specified.
	if !areAllFlagsSpecified(fs, nodeCfgFlags...) {
		nodeCfgFile, err := fs.GetString(configfileF)
		if err != nil {
			panic("unknown flag configfile\n")
		}

		// Read config from file.
		v.SetConfigFile(filepath.Clean(nodeCfgFile))
		v.SetConfigType("yaml")
		err = v.ReadInConfig()
		if err != nil {
			fmt.Printf("Error reading node config file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Using node config file - %s\n", nodeCfgFile)
	}

	// Copy the configuration from viper to struct.
	var nodeCfg perun.NodeConfig
	err := v.Unmarshal(&nodeCfg)
	if err != nil {
		fmt.Printf("Error marshaling node config from viper instance: %v\n", err)
		os.Exit(1)
	}

	nodeCfg.CommTypes = supportedCommTypes
	nodeCfg.ContactTypes = supportedContactTypes
	nodeCfg.CurrencyInterpreters = supportedCurrencyInterpretters
	return nodeCfg
}
