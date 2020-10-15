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
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}

var rootCmd = &cobra.Command{
	Use:   "perunnode",
	Short: "A node for connecting to perun state channel network.",
	Long: `
A node for connecting to perun state channel network. It is a multi-user node,
which can be simultaneously used by many users, with each user opening a
separate, isolated session in the node. The user can first open a session and
within the session use state channels APIs.Currently it implements and supports
only ethereum payment channels.`,
}
