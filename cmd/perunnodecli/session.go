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

	"github.com/abiosoft/ishell"

	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

var (
	sessionCmdUsage = "Usage: session [sub-command]"
	sessionCmd      = &ishell.Cmd{
		Name: "session",
		Help: "Use this command to open and close sessions." + sessionCmdUsage,
		Func: sessionFn,
	}

	sessionOpenCmdUsage = "Usage: session open [session config file]"
	sessionOpenCmd      = &ishell.Cmd{
		Name: "open",
		Help: "Open a new session. Use tab completion to cycle through default values." + sessionOpenCmdUsage,
		Completer: func([]string) []string {
			return []string{"alice/session.yaml", "bob/session.yaml"} // Provide default values as autocompletion.
		},
		Func: sessionOpenFn,
	}
	sessionCloseOpts     = []string{"force", "no-force"}
	sessionCloseCmdUsage = "Usage: session close force|no-force"
	sessionCloseCmd      = &ishell.Cmd{
		Name: "close",
		Help: "Close the current session. Force will persist open chs" + sessionCloseCmdUsage,
		Completer: func([]string) []string {
			return sessionCloseOpts
		},
		Func: sessionCloseFn,
	}
)

func init() {
	sessionCmd.AddCmd(sessionOpenCmd)
	sessionCmd.AddCmd(sessionCloseCmd)
}

func sessionFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	c.Println(c.Cmd.HelpText())
}

func sessionOpenFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	req := pb.OpenSessionReq{
		ConfigFile: c.Args[0],
	}
	resp, err := client.OpenSession(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.OpenSessionResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error opening session : %v", apiErrorString(msgErr.Error)))
		return
	}
	msg := resp.Response.(*pb.OpenSessionResp_MsgSuccess_)
	sessionID = msg.MsgSuccess.SessionID
	c.Printf("%s\n\n", greenf("Session opened."))

	for i := range msg.MsgSuccess.RestoredChs {
		chAlias := addOpenChannelID(msg.MsgSuccess.RestoredChs[i].ChID,
			findPeerAlias(msg.MsgSuccess.RestoredChs[i].BalInfo.Parts))
		c.Printf("%s\n", greenf("Channel restored. Alias: %s.\n%s.", chAlias,
			prettifyPayChInfo(msg.MsgSuccess.RestoredChs[i])))
		paymentSub(c, chAlias)
	}
	c.Printf("\n")
	// Automatically subscribe to channel opening request notifications in this session.
	channelSub(c)
}

func sessionCloseFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	channelUnsub(c) // Close the channel opening request subscriptions before closing the session.

	req := pb.CloseSessionReq{
		SessionID: sessionID,
	}
	if c.Args[0] == "force" {
		req.Force = true
	} else if c.Args[0] == "no-force" {
		req.Force = false
	} else {
		c.Printf("%s\n\n", redf("Parameter should be one of these values: %v", sessionCloseOpts))
		return
	}

	resp, err := client.CloseSession(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.CloseSessionResp_Error)
	if ok {
		channelSub(c) // If there is an error in session close, re-subscribe to channel opening request notifications.

		c.Printf("%s\n\n", redf("Error closing session : %v", apiErrorString(msgErr.Error)))
		return
	}
	msg := resp.Response.(*pb.CloseSessionResp_MsgSuccess_)
	resetLocalCache()
	c.Printf("%s\n\n", greenf("Session closed. ID: %s.", sessionID))
	if c.Args[0] == "force" {
		for i := range msg.MsgSuccess.OpenPayChsInfo {
			chAlias := openChannelsRevMap[msg.MsgSuccess.OpenPayChsInfo[i].ChID]
			c.Printf("%s\n", greenf("Channel persisted. Alias: %s.\n%s.", chAlias,
				prettifyPayChInfo(msg.MsgSuccess.OpenPayChsInfo[i])))
		}
	}
}

func resetLocalCache() {
	channelNotifCounter = 0
	channelNotifList = []string{}
	channelNotifMap = make(map[string]*pb.SubPayChProposalsResp_Notify)

	openChannelsCounter = 0
	openChannelsMap = make(map[string]*openChannelInfo)
	openChannelsRevMap = make(map[string]string)
	openChannelsList = []string{}

	knownAliasesList = []string{}
}
