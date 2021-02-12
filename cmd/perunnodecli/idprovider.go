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
	peerIDCmdUsage = "Usage: peer-id [sub-command]"
	peerIDCmd      = &ishell.Cmd{
		Name: "peer-id",
		Help: "Use this command to add/get peer ID." + peerIDCmdUsage,
		Func: peerIDFn,
	}

	peerIDAddCmdUsage = "Usage: peer-id add [peer alias] [off-chain address] [comm address] [comm type]"
	peerIDAddCmd      = &ishell.Cmd{
		Name: "add",
		Help: "Add a peer ID to the ID provider." + peerIDAddCmdUsage,
		Func: peerIDAddFn,
	}

	peerIDGetCmdUsage = "Usage: peer-id get [peer alias]"
	peerIDGetCmd      = &ishell.Cmd{
		Name: "get",
		Help: "Get peer ID corresponding to the given Alias from the ID provider." + peerIDGetCmdUsage,
		Func: peerIDGetFn,
	}

	// List of known aliases that will be used for autocompletion. Entries will be
	// added when "peer-id get" or "peer-id add" commands return without error.
	knownAliasesList = []string{}
)

func init() {
	peerIDCmd.AddCmd(peerIDAddCmd)
	peerIDCmd.AddCmd(peerIDGetCmd)
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

func peerIDFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	c.Println(c.Cmd.HelpText())
}

func peerIDAddFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}

	// Usage: peer-id add [peer alias] [off-chain address] [comm address] [comm type]",
	countReqArgs := 4
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	req := pb.AddPeerIDReq{
		SessionID: sessionID,
		PeerID: &pb.PeerID{

			Alias:           c.Args[0],
			OffChainAddress: c.Args[1],
			CommAddress:     c.Args[2],
			CommType:        c.Args[3],
		},
	}
	resp, err := client.AddPeerID(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.AddPeerIDResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error adding peer ID : %v", apiErrorString(msgErr.Error)))
		return
	}
	addPeerAlias(c.Args[0])
	c.Printf("%s\n\n", greenf("Peer ID added successfully."))
}

func peerIDGetFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}

	// Usage: peer-id get [peer alias]
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	req := pb.GetPeerIDReq{
		SessionID: sessionID,
		Alias:     c.Args[0],
	}
	resp, err := client.GetPeerID(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.GetPeerIDResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error adding peer ID : %v", msgErr.Error.Error))
		return
	}
	msg := resp.Response.(*pb.GetPeerIDResp_MsgSuccess_)
	addPeerAlias(c.Args[0])
	c.Printf("%s\n\n", greenf("%s", prettifyPeer(msg.MsgSuccess.PeerID)))
}

func prettifyPeer(p *pb.PeerID) string {
	return fmt.Sprintf("Alias: %s, Off-chain Addr: %s, Comm Addr: %s, Comm Type: %s",
		p.Alias, p.OffChainAddress, p.CommAddress, p.CommType)
}
