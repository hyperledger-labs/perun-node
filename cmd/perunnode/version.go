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

	"github.com/spf13/cobra"
)

var (
	// version holds the version of perun node binary. Default value is empty string.
	// If the package is built from a tagged version of the source code, then the variable will be set to
	// tag name (usually a semantic version string) during build using linker flags.
	version string

	// gitCommitID holds the git commit ID of the source code used for building the package. This variable
	// be set during build using linker flags.
	gitCommitID string

	// goperunVersion holds the git commit ID of the go-perun dependecy used by perun-node. This is retrieved
	// from the go.mod file and set during build using linker flags.
	goperunVersion string
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information for perunnode",
	Long:  `Print the version information for perunnode`,
	Run:   versionFn,
}

func versionFn(_ *cobra.Command, _ []string) {
	fmt.Printf("%s Git revision: %s (go-perun version: %s)\n", version, gitCommitID, goperunVersion)
}
