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

// +build integration

package grpc_test

import (
	"context"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	grpclib "google.golang.org/grpc"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/cmd/perunnode"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

var (
	nodeCfg = perun.NodeConfig{
		LogFile:      "",
		LogLevel:     "debug",
		ChainURL:     "ws://127.0.0.1:8545",
		Adjudicator:  "0x9daEdAcb21dce86Af8604Ba1A1D7F9BFE55ddd63",
		Asset:        "0x5992089d61cE79B6CF90506F70DD42B8E42FB21d",
		CommTypes:    []string{"tcp"},
		ContactTypes: []string{"yaml"},
		Currencies:   []string{"ETH"},

		ChainConnTimeout: 30 * time.Second,
		OnChainTxTimeout: 10 * time.Second,
		ResponseTimeout:  10 * time.Second,
	}

	grpcPort = ":50001"
)

func StartServer(t *testing.T) {
	// Initialize a listener.
	listener, err := net.Listen("tcp", grpcPort)
	require.NoErrorf(t, err, "starting listener")

	// Initialize a grpc payment API.
	nodeAPI, err := perunnode.New(nodeCfg)
	require.NoErrorf(t, err, "initializing nodeAPI")
	grpcGrpcPayChServer := grpc.NewPayChServer(nodeAPI)

	// Create grpc server.
	grpcServer := grpclib.NewServer()
	pb.RegisterPayment_APIServer(grpcServer, grpcGrpcPayChServer)

	// Run Server in a go-routine.
	t.Log("Starting server")

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Logf("failed to serve: %v", err)
		}
	}()
}

var (
	client pb.Payment_APIClient
	ctx    context.Context
)

func Test_Integ_Role(t *testing.T) {
	StartServer(t)

	conn, err := grpclib.Dial(grpcPort, grpclib.WithInsecure())
	require.NoError(t, err, "dialing to grpc server")
	t.Log("connected to server")

	// Inititalize client.
	client = pb.NewPayment_APIClient(conn)
	ctx = context.Background()

	t.Run("Node.Time", func(t *testing.T) {
		var timeResp *pb.TimeResp
		timeReq := pb.TimeReq{}
		timeResp, err = client.Time(ctx, &timeReq)
		require.NoError(t, err)
		t.Logf("Response: %+v, Error: %+v", timeResp, err)
	})

	t.Run("Node.GetConfig", func(t *testing.T) {
		var getConfigResp *pb.GetConfigResp
		getConfigReq := pb.GetConfigReq{}
		getConfigResp, err = client.GetConfig(ctx, &getConfigReq)
		require.NoError(t, err)
		t.Logf("Response: %+v, Error: %+v", getConfigResp, err)
	})

	t.Run("Node.Help", func(t *testing.T) {
		var helpResp *pb.HelpResp
		helpReq := pb.HelpReq{}
		helpResp, err = client.Help(ctx, &helpReq)
		require.NoError(t, err)
		t.Logf("Response: %+v, Error: %+v", helpResp, err)
	})

	aliceAlias, bobAlias := "alice", "bob"
	var aliceSessionID, bobSessionID string
	var alicePeer, bobPeer *pb.Peer
	prng := rand.New(rand.NewSource(1729))
	aliceCfgFile := sessiontest.NewConfigFile(t, sessiontest.NewConfig(t, prng))
	bobCfgFile := sessiontest.NewConfigFile(t, sessiontest.NewConfig(t, prng))
	// wg := &sync.WaitGroup{}

	// Run OpenSession for Alice, Bob in top level test, because cleaup functions
	// for removing the keystore directory, contacts file are registered to this
	// testing.T.

	// Alice Open Session.
	aliceSessionID = OpenSession(t, aliceCfgFile)
	t.Logf("%s session id is %s", aliceAlias, aliceSessionID)

	// Bob Open Session.
	bobSessionID = OpenSession(t, bobCfgFile)
	t.Logf("%s session id is %s", bobAlias, bobSessionID)

	t.Run("GetContact", func(t *testing.T) {
		// Get own contact of alice and bob.
		alicePeer = GetContact(t, aliceSessionID, perun.OwnAlias)
		alicePeer.Alias = aliceAlias
		bobPeer = GetContact(t, bobSessionID, perun.OwnAlias)
		bobPeer.Alias = bobAlias
	})

	t.Run("AddContact", func(t *testing.T) {
		// Add bob contact to alice and vice versa.
		AddContact(t, aliceSessionID, bobPeer)
		AddContact(t, bobSessionID, alicePeer)
	})
}

func OpenSession(t *testing.T, cfgFile string) string {
	req := pb.OpenSessionReq{
		ConfigFile: cfgFile,
	}
	resp, err := client.OpenSession(ctx, &req)
	require.NoErrorf(t, err, "OpenSession")
	msg, ok := resp.Response.(*pb.OpenSessionResp_MsgSuccess_)
	require.True(t, ok, "OpenSession returned error response")
	return msg.MsgSuccess.SessionID
}

func GetContact(t *testing.T, sessionID string, alias string) *pb.Peer {
	req := pb.GetContactReq{
		SessionID: sessionID,
		Alias:     alias,
	}
	resp, err := client.GetContact(ctx, &req)
	require.NoErrorf(t, err, "GetContact")
	msg, ok := resp.Response.(*pb.GetContactResp_MsgSuccess_)
	require.True(t, ok, "GetContact returned error response")
	return msg.MsgSuccess.Peer
}

func AddContact(t *testing.T, sessionID string, peer *pb.Peer) {
	req:= pb.AddContactReq{
		SessionID: sessionID,
		Peer:      peer,
	}
	resp, err := client.AddContact(ctx, &req)
	require.NoErrorf(t, err, "AddContact")
	_, ok := resp.Response.(*pb.AddContactResp_MsgSuccess_)
	require.True(t, ok, "AddContact returned error response")
}
