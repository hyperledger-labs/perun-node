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
	"encoding/hex"
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

var (
	logLevel   = log.DebugLevel
	logBackend = log.StdoutBackend

	aliceID, bobID             identity.OffChainID
	alicePassword, bobPassword string
	ethereumNodeURL            string

	//balanceList for simulated backend genesis block
	balanceList map[types.Address]*big.Int

	testKeyStorePath string
	testKeyStore     *keystore.KeyStore

	libSignRuntimeBinFile, mscontractRuntimeBinFile, vpcRuntimeBinFile            string
	libSignaturesRuntimeBin, msContractRuntimeBin, vpcRuntimeBin, dummyRuntimeBin []byte

	defaultConfigFile = "../testdata/test_addresses.json"
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
		fmt.Println("cannot open test config file -", err)
		os.Exit(1)
	}

	jsonData := TestData{}
	err = json.Unmarshal(jsonFile, &jsonData)
	if err != nil {
		fmt.Println("cannot parse test config data -", err)
		os.Exit(-1)
	}

	aliceID = jsonData.AliceID
	bobID = jsonData.BobID
	alicePassword = jsonData.AlicePassword
	bobPassword = jsonData.BobPassword
	ethereumNodeURL = jsonData.EthereumNodeURL

	balanceList = make(map[types.Address]*big.Int)
	balanceList[aliceID.OnChainID] = types.EtherToWei(big.NewInt(1000))
	balanceList[bobID.OnChainID] = types.EtherToWei(big.NewInt(1000))

	testKeyStorePath = filepath.Join(filepath.Dir(configFile), jsonData.KeystoreDir)
	testKeyStorePath, err := filepath.Abs(testKeyStorePath)
	if err != nil {
		fmt.Println("test keystore file path error -", err)
		os.Exit(1)
	}
	testKeyStore = identity.NewKeystore(testKeyStorePath)

	//Read contract bin runtime data
	testContractBinariesDir := filepath.Join(filepath.Dir(configFile), "contracts")
	libSignRuntimeBinFile = filepath.Join(testContractBinariesDir, "LibSignatures.bin-runtime")
	mscontractRuntimeBinFile = filepath.Join(testContractBinariesDir, "MSContract.bin-runtime")
	vpcRuntimeBinFile = filepath.Join(testContractBinariesDir, "VPC.bin-runtime")

	libSignaturesRuntimeBin = getRuntimeBinFromFile(libSignRuntimeBinFile)
	msContractRuntimeBin = getRuntimeBinFromFile(mscontractRuntimeBinFile)
	vpcRuntimeBin = getRuntimeBinFromFile(vpcRuntimeBinFile)
	dummyRuntimeBin = []byte("83794734974343483483847834783748923784274837284304328043")

	setupLogger()

	os.Exit(m.Run())
}

func getRuntimeBinFromFile(pathToFile string) (data []byte) {

	fullPathToFile, err := filepath.Abs(pathToFile)
	if err != nil {
		fmt.Printf("\nerror resolving file path %s", pathToFile)
		os.Exit(-1)
	}

	dataBytes, err := ioutil.ReadFile(fullPathToFile)
	if err != nil {
		fmt.Printf("\nerror reading file %s", fullPathToFile)
		os.Exit(-1)
	}

	dataHex := string(dataBytes)
	data, err = hex.DecodeString(dataHex)
	if err != nil {
		fmt.Printf("\nerror decoding hex string from %s - %s", fullPathToFile, err.Error())
		os.Exit(-1)
	}

	return data
}

func setupLogger() {
	var err error
	logger, err = log.NewLogger(logLevel, logBackend, "blockchain-test")
	if err != nil {
		fmt.Printf("error setting up logger - %s\n", err)
		os.Exit(1)
	}
}
