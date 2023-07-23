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
	"math/rand"
	"os"
	"path/filepath"

	"github.com/idoall/gocryptotrader/common/file"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/idprovider/idprovidertest"
	"github.com/hyperledger-labs/perun-node/node/nodetest"
	"github.com/hyperledger-labs/perun-node/session"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

const (
	aliceAlias, bobAlias = "alice", "bob"
	apiAlias             = "api"
	nodeConfigFile       = "node.yaml"
	sessionConfigFile    = "session.yaml"
	keystoreDir          = "keystore"
	idProviderFile       = "idprovider.yaml"
	databaseDir          = "database"

	onlyNodeF    = "only-node"
	onlySessionF = "only-session"

	dirFileMode = os.FileMode(0o750) // file mode for creating the directories for alice and bob.
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate demo artifacts",
	Long: `
Generate demo artifacts for node and session configuration.

- Node: node.yaml file.
- Session: Two directories (alice and bob) each containing session.yaml file,
  idprovider.yaml file and keystore directory with keys corresponding to the
  on-chain and off-chain accounts.

Note:
=====
Use the below command to start a ganache cli node with accounts corresponding
to the generated keys pre-funded with 100 ETH each.

ganache-cli -b 1 \
 --account="0x7d51a817ee07c3f28581c47a5072142193337fdca4d7911e58c5af2d03895d1a,\
100000000000000000000" \
 --account="0x6aeeb7f09e757baa9d3935a042c3d0d46a2eda19e9b676283dce4eaf32e29dc9,\
100000000000000000000"
`,

	Run: generate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	defineGenerateCmdFlags()
}

func defineGenerateCmdFlags() {
	generateCmd.Flags().Bool(onlyNodeF, false, "generate only node configuration artifacts.")
	generateCmd.Flags().Bool(onlySessionF, false, "generate only session configuration artifacts for both alice & bob.")
}

func generate(cmd *cobra.Command, _ []string) {
	genNodeConfig, err := cmd.Flags().GetBool(onlyNodeF)
	if err != nil {
		panic("unknown flag " + onlyNodeF + "\n")
	}

	genSessionConfig, err := cmd.Flags().GetBool(onlySessionF)
	if err != nil {
		panic("unknown flag " + onlySessionF + "\n")
	}

	// Enable generation of both config when no flags are specified.
	if areAllFlagsUnspecified(cmd.Flags(), onlyNodeF, onlySessionF) {
		genNodeConfig = true
		genSessionConfig = true
	}

	// Generate configuration artifacts.
	if genNodeConfig {
		if err = generateNodeConfig(); err != nil {
			fmt.Printf("Error generating node configuration artifacts: %v\n", err)
		} else {
			fmt.Printf("Generated node configuration file: %s\n", nodeConfigFile)
		}
	}
	if genSessionConfig {
		if err = generateSessionConfig(); err != nil {
			fmt.Printf("Error generating session configuration artifacts: %v\n", err)
		} else {
			fmt.Printf("Generated session configuration artifacts: %s, %s, %s directories\n",
				aliceAlias, bobAlias, apiAlias)
		}
	}
	if err != nil {
		os.Exit(1)
	}
}

// generateNodeConfig generates node configuration artifact (node.yaml) in the current directory.
func generateNodeConfig() error {
	// Check if file exists.
	if _, err := os.Stat(nodeConfigFile); !os.IsNotExist(err) {
		return errors.New("file exists - " + nodeConfigFile)
	}
	nodeCfg := nodetest.NewConfig(true)
	adjudicator, assetETH, _ := ethereumtest.ContractAddrs()
	nodeCfg.Adjudicator = adjudicator.String()
	nodeCfg.AssetETH = assetETH.String()
	// Create file in temp dir.
	tempNodeConfigFile, err := sessiontest.NewConfigFile(nodeCfg)
	if err != nil {
		return err
	}
	// Move the file to current directory.
	filesToMove := map[string]string{tempNodeConfigFile: filepath.Join(nodeConfigFile)}
	return moveFiles(filesToMove)
}

// generateSessionConfig generates two sets of session configuration artifacts in two directories named alice and bob.
// Each directory would have: session.yaml, idprovider.yaml and keystore (containing 2 key files - on-chain
// & off-chain). To use this configuration, start the node from same directory containing the session config artifacts
// directory and pass the path "alice/session.yaml" and "bob/session.yaml" for alice and bob respectively.
func generateSessionConfig() error {
	if isPresent, dirName := isAnyDirPresent(aliceAlias, bobAlias, apiAlias); isPresent {
		return errors.New("dir exists - " + dirName)
	}
	err := makeDirs(aliceAlias, bobAlias, apiAlias)
	if err != nil {
		return err
	}

	const count = 3
	aliases := [count]string{aliceAlias, bobAlias, apiAlias}
	cfgs := [count]session.Config{}
	providersFile := [count]string{}
	cfgFile := [count]string{}

	// Generate session config, the seed ethereumtest.RandSeedForTestAccs generates two accounts which were funded
	// when starting the ganache cli node with the command documented in help message.
	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs)) //nolint:gosec		// okay to use weak rand in tests.
	for i := 0; i < count; i++ {
		cfgs[i], err = sessiontest.NewConfig(prng)
		if err != nil {
			return err
		}
		cfgs[i].User.Alias = aliases[i]
	}

	for i := 0; i < count; i++ {
		providersFile[i], err = idprovidertest.NewIDProvider(peerID(cfgs[i].User))
		if err != nil {
			return err
		}
	}

	for i := 0; i < count; i++ {
		cfgFile[i], err = sessiontest.NewConfigFile(updatedConfigCopy(cfgs[i]))
		if err != nil {
			return err
		}
	}

	// Move the artifacts to currenct directory.
	filesToMove := make(map[string]string)
	for i := 0; i < count; i++ {
		cfgFile[i] = filepath.Join(aliases[i], sessionConfigFile)
		providersFile[i] = filepath.Join(aliases[i], idProviderFile)
		cfgs[i].DatabaseDir = filepath.Join(aliases[i], databaseDir)
		cfgs[i].User.OnChainWallet.KeystorePath = filepath.Join(aliases[i], keystoreDir)
	}
	return moveFiles(filesToMove)
}

func isAnyDirPresent(dirNames ...string) (bool, string) {
	for i := range dirNames {
		if _, err := os.Stat(dirNames[i]); !os.IsNotExist(err) {
			return true, dirNames[i]
		}
	}
	return false, ""
}

func makeDirs(dirNames ...string) error {
	var err error
	for i := range dirNames {
		if err = os.Mkdir(dirNames[i], dirFileMode); err != nil {
			return errors.New("creating dir - " + dirNames[i])
		}
	}
	return nil
}

func peerID(userCfg session.UserConfig) perun.PeerID {
	return perun.PeerID{
		Alias:              userCfg.Alias,
		OffChainAddrString: userCfg.OffChainAddr,
		CommAddr:           userCfg.CommAddr,
		CommType:           userCfg.CommType,
	}
}

func updatedConfigCopy(cfg session.Config) session.Config {
	cfgCopy := cfg
	cfgCopy.IDProviderURL = filepath.Join(cfg.User.Alias, idProviderFile)
	cfgCopy.DatabaseDir = filepath.Join(cfg.User.Alias, databaseDir)
	cfgCopy.User.OnChainWallet.KeystorePath = filepath.Join(cfg.User.Alias, keystoreDir)
	cfgCopy.User.OffChainWallet.KeystorePath = filepath.Join(cfg.User.Alias, keystoreDir)
	return cfgCopy
}

func moveFiles(srcDest map[string]string) error {
	errs := []string{}
	for src, dest := range srcDest {
		//  avoiding invalid cross-device link, therefore dir has to be checked and os.Rename cannot be used
		fileInfo, err := os.Stat(src)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s to %s: %v", src, dest, err))
		}
		if fileInfo.IsDir() {
			if err := handleDir(src, dest); err != nil {
				errs = append(errs, fmt.Sprintf("%s to %s: %v", src, dest, err))
			}
		} else {
			if err := file.Move(src, dest); err != nil {
				errs = append(errs, fmt.Sprintf("%s to %s: %v", src, dest, err))
			}
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf("moving files: %v", errs)
	}
	return nil
}

func handleDir(srcDir string, destDir string) error {
	errs := []string{}
	if err := os.Mkdir(destDir, dirFileMode); err != nil {
		errs = append(errs, fmt.Sprintf("creating dir - %s: %v", destDir, err))
	}
	files, err := os.ReadDir(srcDir)
	if err != nil {
		errs = append(errs, fmt.Sprintf("reading dir - %s: %v", srcDir, err))
	}
	for _, fileToMove := range files {
		if err := file.Move(srcDir+"/"+fileToMove.Name(), destDir+"/"+fileToMove.Name()); err != nil {
			errs = append(errs, fmt.Sprintf("%s to %s: %v", srcDir+"/"+fileToMove.Name(), destDir+"/"+fileToMove.Name(), err))
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf("handleDir files: %v", errs)
	}
	return nil
}
