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
	"context"
	"fmt"

	"github.com/abiosoft/ishell"

	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

var (
	contactCmdUsage = "Usage: contact [sub-command]"
	contactCmd      = &ishell.Cmd{
		Name: "contact",
		Help: "Use this command to add/get idProvider." + contactCmdUsage,
		Func: contactFn,
	}

	contactAddCmdUsage = "Usage: contact add [peer alias] [off-chain address] [comm address] [comm type]"
	contactAddCmd      = &ishell.Cmd{
		Name: "add",
		Help: "Add a peer to idProvider." + contactAddCmdUsage,
		Func: contactAddFn,
	}

	contactGetCmdUsage = "Usage: contact get [peer alias]"
	contactGetCmd      = &ishell.Cmd{
		Name: "get",
		Help: "Get peer info from idProvider." + contactGetCmdUsage,
		Func: contactGetFn,
	}

	// List of known aliases that will be used for autocompletion. Entries will be
	// added when "contact get" or "contact add" commands return without error.
	knownAliasesList = []string{}
)

func init() {
	contactCmd.AddCmd(contactAddCmd)
	contactCmd.AddCmd(contactGetCmd)
}

// Add alias to known aliases list for autocompletion.
func addPeerAlias(alias string) {
	aliasIdx := -1
	for idx := range knownAliasesList {
		if knownAliasesList[idx] == alias {
			aliasIdx = idx
			break
		}
	}
	if aliasIdx == -1 {
		// Add only if the entry is not present already.
		knownAliasesList = append(knownAliasesList, alias)
	}
}

func contactFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	c.Println(c.Cmd.HelpText())
}

func contactAddFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}

	// Usage: contact add [peer alias] [off-chain address] [comm address] [comm type]",
	countReqArgs := 4
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	req := pb.AddContactReq{
		SessionID: sessionID,
		Peer: &pb.Peer{

			Alias:           c.Args[0],
			OffChainAddress: c.Args[1],
			CommAddress:     c.Args[2],
			CommType:        c.Args[3],
		},
	}
	resp, err := client.AddContact(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.AddContactResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error adding contact : %v", msgErr.Error.Error))
		return
	}
	addPeerAlias(c.Args[0])
	c.Printf("%s\n\n", greenf("Contact added successfully."))
}

func contactGetFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}

	// Usage: contact get [peer alias]
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	req := pb.GetContactReq{
		SessionID: sessionID,
		Alias:     c.Args[0],
	}
	resp, err := client.GetContact(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.GetContactResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error adding contact : %v", msgErr.Error.Error))
		return
	}
	msg := resp.Response.(*pb.GetContactResp_MsgSuccess_)
	addPeerAlias(c.Args[0])
	c.Printf("%s\n\n", greenf("%s", prettifyPeer(msg.MsgSuccess.Peer)))
}

func prettifyPeer(p *pb.Peer) string {
	return fmt.Sprintf("Alias: %s, Off-chain Addr: %s, Comm Addr: %s, Comm Type: %s",
		p.Alias, p.OffChainAddress, p.CommAddress, p.CommType)
}
