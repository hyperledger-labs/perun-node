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

package identity

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/direct-state-transfer/dst-go/ethereum/keystore"
	"github.com/direct-state-transfer/dst-go/log"
)

type TestData struct {
	KeystoreDir   string     `json:"keystore_dir"`
	KnownIdsFile  string     `json:"known_ids_file"`
	AlicePassword string     `json:"alice_password"`
	BobPassword   string     `json:"bob_password"`
	AliceID       OffChainID `json:"alice_id"`
	BobID         OffChainID `json:"bob_id"`
}

var (
	aliceID, bobID             OffChainID
	alicePassword, bobPassword string

	dummyKeyStore = NewKeystore("random-keys-dir")

	testKeyStorePath string
	testKeyStore     *keystore.KeyStore

	knownIdsFile string

	configFile             = "./testdata/test_addresses.json"
	dummyConfigFileFlagVar string
)

func TestMain(m *testing.M) {

	//Define and parse flag for compatibility with other models
	flag.StringVar(&dummyConfigFileFlagVar, "configFile", "", "Flag defined for compatibility, do not use. See config_for_test.go instead")
	flag.Parse()

	//Flag value is ignored, use hardcoded value
	//Because signatures used in tests are for these keys
	var err error
	configFile, err = filepath.Abs(configFile)
	if err != nil {
		fmt.Println("test config file path error -", err)
		os.Exit(1)
	}
	fmt.Println("using test config file -", configFile)

	jsonFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Cannot open test config file -", err)
		os.Exit(1)
	}

	jsonData := TestData{}
	err = json.Unmarshal(jsonFile, &jsonData)
	if err != nil {
		fmt.Println("Cannot parse test config data -", err)
		os.Exit(-1)
	}

	aliceID = jsonData.AliceID
	bobID = jsonData.BobID
	alicePassword = jsonData.AlicePassword
	bobPassword = jsonData.BobPassword

	testKeyStorePath = filepath.Join(filepath.Dir(configFile), jsonData.KeystoreDir)
	testKeyStorePath, err = filepath.Abs(testKeyStorePath)
	if err != nil {
		fmt.Println("test keystore file path error -", err)
		os.Exit(1)
	}
	testKeyStore = NewKeystore(testKeyStorePath)

	knownIdsFile = filepath.Join(filepath.Dir(configFile), jsonData.KnownIdsFile)
	knownIdsFile, err = filepath.Abs(knownIdsFile)
	if err != nil {
		fmt.Println("known ids file path error -", err)
		os.Exit(1)
	}

	setupLogger()

	os.Exit(m.Run())
}

func setupLogger() {

	var err error
	logger, err = log.NewLogger(log.DebugLevel, log.StdoutBackend, "identity-test")
	if err != nil {
		fmt.Println("Error initializing logger for tests")
	}

	keystoreLogger, err := log.NewLogger(log.DebugLevel, log.StdoutBackend, "keystore-test")
	if err != nil {
		fmt.Println("Error initializing keystore logger for tests")
	}
	keystore.SetLogger(keystoreLogger)
}
