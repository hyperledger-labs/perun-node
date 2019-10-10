// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
//     https://github.com/direct-state-transfer/dst-go/NOTICE
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	"github.com/direct-state-transfer/dst-go/channel"
	"github.com/direct-state-transfer/dst-go/ethereum/keystore"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
	"github.com/direct-state-transfer/dst-go/log"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// TestData is the collection of dummy test parameters used for walkthrough execution.
type TestData struct {
	KeystoreDir     string              `json:"keystore_dir"`
	EthereumNodeURL string              `json:"ethereum_node_url"`
	AlicePassword   string              `json:"alice_password"`
	BobPassword     string              `json:"bob_password"`
	AliceID         identity.OffChainID `json:"alice_id"`
	BobID           identity.OffChainID `json:"bob_id"`
}

var (
	//BalanceList for simulated backend genesis block
	BalanceList map[types.Address]*big.Int

	aliceID, bobID                     identity.OffChainID
	alicePassword, bobPassword         string
	aliceEthereumAddr, bobEthereumAddr types.Address
	ethereumNodeURL                    string

	testKeyStorePath string
	testKeystore     *keystore.KeyStore

	aliceColor = color.New(color.FgGreen)
	bobColor   = color.New(color.FgYellow)

	defaultConfigFile = "../testdata/test_addresses.json"
	configFile        string
)

func setupConfig(filePath string) {

	configFile, err := filepath.Abs(configFile)
	if err != nil {
		fmt.Println("walkthrough config file path error -", err)
		os.Exit(1)
	}
	fmt.Println("using walkthrough config file -", configFile)

	jsonFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Cannot open walkthrough config file -", err)
		os.Exit(1)
	}

	jsonData := TestData{}
	err = json.Unmarshal(jsonFile, &jsonData)
	if err != nil {
		fmt.Println("Cannot parse walkthrough config data -", err)
		os.Exit(-1)
	}

	ethereumNodeURL = jsonData.EthereumNodeURL
	aliceID = jsonData.AliceID
	bobID = jsonData.BobID
	alicePassword = jsonData.AlicePassword
	bobPassword = jsonData.BobPassword

	aliceEthereumAddr = aliceID.OnChainID
	bobEthereumAddr = bobID.OnChainID

	BalanceList = make(map[types.Address]*big.Int)
	BalanceList[aliceEthereumAddr] = types.EtherToWei(big.NewInt(1000))
	BalanceList[bobEthereumAddr] = types.EtherToWei(big.NewInt(1000))

	testKeyStorePath = filepath.Join(filepath.Dir(configFile), jsonData.KeystoreDir)
	testKeyStorePath, err := filepath.Abs(testKeyStorePath)
	if err != nil {
		fmt.Println("test keystore file path error -", err)
		os.Exit(1)
	}
	testKeystore = identity.NewKeystore(testKeyStorePath)
}

func main() {

	walkthroughApp := &cobra.Command{
		Use: "walkthrough [flags] (any one backend must be specified)",
		Run: walkthrough,
	}
	addFlagSet(walkthroughApp)
	err := walkthroughApp.Execute()
	if err != nil {
		fmt.Println("Error initializing walkthrough app -", err)
	}

}

func addFlagSet(app *cobra.Command) {

	app.PersistentFlags().String(
		"configFile", defaultConfigFile, "Config file for unit tests")

	app.PersistentFlags().String(
		"ethereum_address", ethereumNodeURL, "Address of ethereum node to connect. Provide complete url")
	app.PersistentFlags().Bool(
		"simulated_backend", false, "Run walkthrough with simulated backend for both alice and bob")
	app.PersistentFlags().Bool(
		"real_backend", false, "Run walkthrough with real backend for alice and bob")
	app.PersistentFlags().Bool(
		"real_backend_alice", false, "Run walkthrough with real backend for alice")
	app.PersistentFlags().Bool(
		"real_backend_bob", false, "Run walkthrough with real backend for bob")
	app.PersistentFlags().Bool(
		"ch_message_print", false, "Enable/Disable printing of channel messages")
	app.PersistentFlags().Bool(
		"dispute", false, "Run walkthrough for dispute condition during closure")

}

func walkthrough(app *cobra.Command, args []string) {

	//Parse user configurations
	chMsgPrint, _ := app.Flags().GetBool("ch_message_print")

	simulatedBackend, _ := app.Flags().GetBool("simulated_backend")
	realBackend, _ := app.Flags().GetBool("real_backend")

	realBackendAlice, _ := app.Flags().GetBool("real_backend_alice")
	realBackendBob, _ := app.Flags().GetBool("real_backend_bob")
	realBackendAlice = realBackendAlice || realBackend
	realBackendBob = realBackendBob || realBackend

	dispute, _ := app.Flags().GetBool("dispute")

	if !(simulatedBackend || realBackendAlice || realBackendBob) {
		_, _ = fmt.Fprintf(app.OutOrStderr(), "\nNo blockchain backend specified.\n\n")
		_ = app.Help()
		return
	}

	configFile, _ = app.Flags().GetString("configFile")
	setupConfig(configFile)

	configModuleLogger(log.ErrorLevel)
	channel.ReadWriteLogging = chMsgPrint

	wg := &sync.WaitGroup{}

	switch {
	case simulatedBackend:
		wg.Add(1)
		simulatedBlockchain(wg, dispute)
		wg.Wait()
	case (realBackendAlice || realBackendBob):
		wg.Add(1)
		realBlockchain(realBackendAlice, realBackendBob, wg, dispute)
		wg.Wait()
	}

	fmt.Println("Walkthrough execution complete")
}

func configModuleLogger(logLevel log.Level) {

	chConfig := channel.Config{
		Logger: log.Config{
			Level:   logLevel,
			Backend: log.StdoutBackend,
		},
	}

	err := channel.InitModule(&chConfig)
	if err != nil {
		fmt.Println("Error initializing config module")
	}
}
