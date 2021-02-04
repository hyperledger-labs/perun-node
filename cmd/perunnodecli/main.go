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

	"github.com/abiosoft/ishell"
	"github.com/fatih/color"

	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

var (
	// File that stores history of commands used in the interactive shell.
	// This will be preserved across the multiple runs of perunnode cli.
	// It will be located in the home directory.
	historyFile = ".perunnodecli_history"

	// Singleton instance of ishell that is used throughout this program.
	// this will be initialized in main() and be accessed by subscription
	// handler to print the received notification messages.
	sh *ishell.Shell

	// Singleton instance of grpc payment channel client that will be
	// used by all functions in this program. This is safe for concurrent
	// access without a mutex.
	client pb.Payment_APIClient

	// Session ID for the currently active session. The cli application
	// allows only one session to be open at a time and all channel requests,
	// payments and payment requests are sent and received in this context of
	// this session.
	// It is set when a session is opened and closed when a session is closed.
	sessionID string

	// standard value of challenge duration for all outgoing channel open requests.
	challengeDurSecs uint64 = 10

	// SPrintf style functions that produce colored text.
	redf   = color.New(color.FgRed).SprintfFunc()
	greenf = color.New(color.FgGreen).SprintfFunc()
)

func main() {
	// New shell includes help, clear, exit commands by default.
	sh = ishell.New()

	// Read and write history to $HOME/historyFile
	sh.SetHomeHistoryPath(historyFile)

	sh.AddCmd(chainCmd)
	sh.AddCmd(nodeCmd)
	sh.AddCmd(sessionCmd)
	sh.AddCmd(peerIDCmd)
	sh.AddCmd(channelCmd)
	sh.AddCmd(paymentCmd)

	sh.Printf("Perun node cli application.\n\n")
	sh.Printf("%s\n\n", greenf("Connect to a perun node instance using 'node connect' for making any transactions."))

	sh.Run()
}

// printNodeNotConnectedError is a helper function to print error message that is used across mutiple commands.
func printNodeNotConnectedError(c ishell.Actions) {
	c.Printf("%s\n\n", redf("Not connected to perun node, connect using 'node connect' command."))
}

// printArgCountError is a helper function to print error message that is used across mutiple commands.
func printArgCountError(c *ishell.Context, reqArgCount int) {
	c.Printf("%s\n\n", redf("Got %d arg(s). Want %d.", len(c.Args), reqArgCount))
	c.Printf("Command help:\t%s\n\n", c.Cmd.Help)
}

// printCommandSendingError is a helper function to print error message that is used across mutiple commands.
func printCommandSendingError(c ishell.Actions, err error) {
	c.Printf("%s\n\n", redf("Error sending command to perun node: %v.", err))
}

// printUnknownChAliasError is a helper function to print error message that is used across mutiple commands.
func printUnknownChannelAliasError(c ishell.Actions, chAlias string) {
	c.Printf("%s\n\n", redf("Unknown channel alias %s", chAlias))
	c.Printf("%s\n\n", redf("Known channel aliases:\n%v", openChannelsList))
}

// printUnknownChNotifError is a helper function to print error message that is used across mutiple commands.
func printUnknownChNotifError(c ishell.Actions, channelNotifAlias string) {
	c.Printf("%s\n\n", redf("Unknown channel opening request notification alias.%s", channelNotifAlias))
	c.Printf("%s\n\n", redf("Known channel opening request notification aliases:\n%v.", channelNotifList))
}

// printNoPaymentNotifError is a helper function to print error message that is used across mutiple commands.
func printNoPaymentNotifError(c ishell.Actions, chAlias string) {
	c.Printf("%s\n\n", redf("No payment notifications pending a response on this channel.%s", chAlias))
}

// apiErrorString formats the error message returned by the API into pretty strings.
func apiErrorString(e *pb.MsgErrorV2) string {
	return fmt.Sprintf("category: %s, code: %d, message: %s, additional info: %+v",
		e.Category, e.Code, e.Message, e.AddInfo)
}
