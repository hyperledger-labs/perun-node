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

package adapter

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/direct-state-transfer/dst-go/ethereum/keystore"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
	"github.com/direct-state-transfer/dst-go/log"
)

type TestData struct {
	EthereumNodeURL string              `json:"ethereum_node_url"`
	KeystoreDir     string              `json:"keystore_dir"`
	AlicePassword   string              `json:"alice_password"`
	BobPassword     string              `json:"bob_password"`
	AliceID         identity.OffChainID `json:"alice_id"`
	BobID           identity.OffChainID `json:"bob_id"`
}

//User and ethereum node details parsed from testdata
var (
	aliceID, bobID             identity.OffChainID
	alicePassword, bobPassword string
	ethereumNodeURL            string

	//balanceList for simulated backend genesis block
	balanceList map[types.Address]*big.Int

	testKeyStorePath string
	testKeyStore     *keystore.KeyStore

	dummyKeyStore = identity.NewKeystore("random-keys-dir")

	defaultConfigFile = "../../testdata/test_addresses.json"
	configFile        string
)

func TestMain(m *testing.M) {

	flag.StringVar(&configFile, "configFile", defaultConfigFile, "Config file for unit tests")
	flag.Parse()
	configFile, err := filepath.Abs(configFile)
	if err != nil {
		fmt.Println("test config file path error -", err)
		os.Exit(1)
	}
	fmt.Println("using test config file -", configFile)

	jsonFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Cannot open test_addresses file -", err)
		os.Exit(1)
	}

	jsonData := TestData{}
	err = json.Unmarshal(jsonFile, &jsonData)
	if err != nil {
		fmt.Println("Cannot parse test_addresses data -", err)
		os.Exit(-1)
	}
	fmt.Printf("\n%+v\n", jsonData)

	aliceID = jsonData.AliceID
	bobID = jsonData.BobID
	alicePassword = jsonData.AlicePassword
	bobPassword = jsonData.BobPassword
	ethereumNodeURL = jsonData.EthereumNodeURL

	balanceList = make(map[types.Address]*big.Int)
	balanceList[aliceID.OnChainID] = types.EtherToWei(big.NewInt(1000))
	balanceList[bobID.OnChainID] = types.EtherToWei(big.NewInt(1000))

	configFileDir := filepath.Dir(configFile)
	testKeyStorePath = filepath.Join(configFileDir, jsonData.KeystoreDir)
	testKeyStorePath, err := filepath.Abs(testKeyStorePath)
	if err != nil {
		fmt.Println("test keystore file path error -", err)
		os.Exit(1)
	}
	fmt.Println(testKeyStorePath)
	testKeyStore = identity.NewKeystore(testKeyStorePath)

	setupLogger()

	os.Exit(m.Run())
}

func setupLogger() {

	identityLogger, err := log.NewLogger(log.DebugLevel, log.StdoutBackend, "identity-test")
	if err != nil {
		fmt.Println("Error initialising identity logger for tests")
	}
	identity.SetLogger(identityLogger)

	keystoreLogger, err := log.NewLogger(log.DebugLevel, log.StdoutBackend, "keystore-test")
	if err != nil {
		fmt.Println("Error initialising keystore logger for tests")
	}
	keystore.SetLogger(keystoreLogger)
}
