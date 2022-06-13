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
	"time"

	"github.com/abiosoft/ishell"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

var (
	nodeCmdUsage = "Usage: node [sub-command]"
	nodeCmd      = &ishell.Cmd{
		Name: "node",
		Help: "Use the command to access the node related functionalities." + nodeCmdUsage,
		Func: nodeFn,
	}

	nodeConnectCmdUsage = "Usage: node connect [url]"
	nodeConnectCmd      = &ishell.Cmd{
		Name: "connect",
		Help: "Connect to a running perun node instance. Use tab completion to cycle through default values." +
			nodeConnectCmdUsage,
		Completer: func([]string) []string {
			return []string{":50001"} // Provide default values as autocompletion.
		},
		Func: nodeConnectFn,
	}

	nodeTimeCmdUsage = "Usage: node time"
	nodeTimeCmd      = &ishell.Cmd{
		Name: "time",
		Help: "Print node time." + nodeTimeCmdUsage,
		Func: nodeTimeFn,
	}

	nodeConfigCmdUsage = "Usage: node config"
	nodeConfigCmd      = &ishell.Cmd{
		Name: "config",
		Help: "Print node config." + nodeConfigCmdUsage,
		Func: nodeConfigFn,
	}
)

func init() {
	nodeCmd.AddCmd(nodeConnectCmd)
	nodeCmd.AddCmd(nodeTimeCmd)
	nodeCmd.AddCmd(nodeConfigCmd)
}

func nodeFn(c *ishell.Context) {
	c.Println(c.Cmd.HelpText())
}

func nodeConnectFn(c *ishell.Context) {
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	nodeAddr := c.Args[0]
	conn, err := grpclib.Dial(nodeAddr, grpclib.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		sh.Printf("Error connecting to perun node at %s: %v", nodeAddr, err)
	}
	client = pb.NewPayment_APIClient(conn)
	t, err := getNodeTime()
	if err != nil {
		c.Printf("%s\n\n", redf("Error connecting to perun node: %v", err))
		return
	}
	c.Printf("%s\n\n", greenf("Connected to perun node at %s. Node time is %v", nodeAddr, time.Unix(t, 0)))
}

func nodeTimeFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	countReqArgs := 0
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	t, err := getNodeTime()
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	c.Printf("%s\n\n", greenf("Perun node time: %s", time.Unix(t, 0)))
}

func getNodeTime() (int64, error) {
	timeReq := pb.TimeReq{}
	timeResp, err := client.Time(context.Background(), &timeReq)
	if err != nil {
		return 0, err
	}
	return timeResp.Time, err
}

func nodeConfigFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	countReqArgs := 0
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	getConfigReq := pb.GetConfigReq{}
	getConfigResp, err := client.GetConfig(context.Background(), &getConfigReq)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	c.Printf("%s\n\n", greenf("Perun node config:\n%v", prettify(getConfigResp)))
}
