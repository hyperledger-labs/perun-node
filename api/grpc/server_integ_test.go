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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	grpclib "google.golang.org/grpc"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/node"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

// This file contains integration tests for all the APIs. Start the ganache cli node using the
// below command before running the tests:
//
// ganache-cli -b 1 \
// --account="0x1fedd636dbc7e8d41a0622a2040b86fea8842cef9d4aa4c582aad00465b7acff,100000000000000000000" \
// --account="0xb0309c60b4622d3071fad3e16c2ce4d0b1e7758316c187754f4dd0cfb44ceb33,100000000000000000000"

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
	nodeAPI, err := node.New(nodeCfg)
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
	var chID string
	prng := rand.New(rand.NewSource(1729))
	aliceCfg := sessiontest.NewConfig(t, prng)
	bobCfg := sessiontest.NewConfig(t, prng)
	aliceCfgFile := sessiontest.NewConfigFile(t, aliceCfg)
	bobCfgFile := sessiontest.NewConfigFile(t, bobCfg)
	wg := &sync.WaitGroup{}

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

	t.Run("OpenCh_Sub_Unsub_Respond_Accept", func(t *testing.T) {
		// Alice proposes a channel and bob accepts.
		wg.Add(1)
		go func() {
			chID = OpenPayCh(t, aliceSessionID, []string{perun.OwnAlias, bobAlias}, []string{"1", "2"}, false)

			wg.Done()
		}()
		sub := SubPayChProposal(t, bobSessionID)
		notif := ReadPayChProposalNotif(t, bobSessionID, sub, false)
		RespondPayChProposal(t, bobSessionID, notif.Notify.ProposalID, true, false)
		UnsubPayChProposal(t, bobSessionID, false)

		wg.Wait()
	})

	t.Run("CloseSession_NoForce_ErrOpenPayChs", func(t *testing.T) {
		openPayChsInfo := CloseSession(t, aliceSessionID, false, true)
		require.Len(t, openPayChsInfo, 1)
		assert.Equal(t, chID, openPayChsInfo[0].ChID)
		openPayChsInfo = CloseSession(t, bobSessionID, false, true)
		require.Len(t, openPayChsInfo, 1)
		assert.Equal(t, chID, openPayChsInfo[0].ChID)
	})

	t.Run("OpenCh_Sub_Unsub_Respond_Reject", func(t *testing.T) {
		// Bob proposes a channel and alice accepts.
		wg.Add(1)
		go func() {
			OpenPayCh(t, bobSessionID, []string{perun.OwnAlias, aliceAlias}, []string{"1", "2"}, true)

			wg.Done()
		}()
		sub := SubPayChProposal(t, aliceSessionID)
		notif := ReadPayChProposalNotif(t, aliceSessionID, sub, false)
		RespondPayChProposal(t, aliceSessionID, notif.Notify.ProposalID, false, false)
		UnsubPayChProposal(t, aliceSessionID, false)

		wg.Wait()
	})

	t.Run("SendPayChUpdate_Sub_Unsub_Respond_Accept", func(t *testing.T) {
		// Bob sends a payment and alice accepts.
		wg.Add(1)
		go func() {
			SendPayChUpdate(t, bobSessionID, chID, aliceAlias, "0.5", false)

			wg.Done()
		}()

		sub := SubPayChUpdate(t, aliceSessionID, chID)
		notif := ReadPayChUpdateNotif(t, aliceSessionID, chID, sub)
		assert.EqualValues(t, perun.ChUpdateTypeOpen, notif.Notify.Type)
		RespondPayChUpdate(t, aliceSessionID, chID, notif.Notify.UpdateID, true)
		UnsubPayChUpdate(t, aliceSessionID, chID)

		wg.Wait()
	})

	t.Run("SendPayChUpdate_Sub_Unsub_Respond_Reject", func(t *testing.T) {
		// Alice sends a payment and bob accepts.
		wg.Add(1)
		go func() {
			SendPayChUpdate(t, aliceSessionID, chID, bobAlias, "0.5", true)

			wg.Done()
		}()

		sub := SubPayChUpdate(t, bobSessionID, chID)
		notif := ReadPayChUpdateNotif(t, bobSessionID, chID, sub)
		assert.EqualValues(t, perun.ChUpdateTypeOpen, notif.Notify.Type)
		RespondPayChUpdate(t, bobSessionID, chID, notif.Notify.UpdateID, false)
		UnsubPayChUpdate(t, bobSessionID, chID)

		wg.Wait()
	})

	isClosePayChSuccessful := make(chan bool, 1)
	t.Run("Close_Sub_Unsub", func(t *testing.T) {
		// Bob closes payment channel, Alice receives notification for rejects and final update,
		// both receive channel closed notifications.
		wg.Add(2)
		go func() {
			sub := SubPayChUpdate(t, aliceSessionID, chID)
			notif := ReadPayChUpdateNotif(t, aliceSessionID, chID, sub)
			assert.EqualValues(t, perun.ChUpdateTypeFinal, notif.Notify.Type)
			RespondPayChUpdate(t, aliceSessionID, chID, notif.Notify.UpdateID, false)
			notif = ReadPayChUpdateNotif(t, aliceSessionID, chID, sub)
			assert.EqualValues(t, perun.ChUpdateTypeClosed, notif.Notify.Type)
			UnsubPayChUpdate(t, aliceSessionID, chID)

			wg.Done()
		}()
		go func() {
			sub := SubPayChUpdate(t, bobSessionID, chID)
			notif := ReadPayChUpdateNotif(t, bobSessionID, chID, sub)
			assert.EqualValues(t, perun.ChUpdateTypeClosed, notif.Notify.Type)
			RespondPayChUpdateExpectError(t, bobSessionID, chID, notif.Notify.UpdateID, true)
			UnsubPayChUpdate(t, bobSessionID, chID)

			wg.Done()
		}()

		time.Sleep(2 * time.Second) // Wait for the subscriptions to be made.
		isClosePayChSuccessful <- ClosePayCh(t, bobSessionID, chID)

		wg.Wait()
	})

	require.True(t, <-isClosePayChSuccessful)
	t.Run("CloseSession_NoForce_", func(t *testing.T) {
		openPayChsInfo := CloseSession(t, aliceSessionID, false, false)
		require.Len(t, openPayChsInfo, 0)
		openPayChsInfo = CloseSession(t, bobSessionID, false, false)
		require.Len(t, openPayChsInfo, 0)
	})
	t.Run("APIs error when session is closed", func(t *testing.T) {
		OpenPayCh(t, bobSessionID, []string{perun.OwnAlias, aliceAlias}, []string{"1", "2"}, true)
		sub := SubPayChProposal(t, aliceSessionID)
		ReadPayChProposalNotif(t, aliceSessionID, sub, true)
		RespondPayChProposal(t, aliceSessionID, "", false, true)
		UnsubPayChProposal(t, aliceSessionID, true)
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

func CloseSession(t *testing.T, sessionID string, force bool, wantErr bool) []*pb.PayChInfo {
	req := pb.CloseSessionReq{
		SessionID: sessionID,
		Force:     force,
	}
	resp, err := client.CloseSession(ctx, &req)
	require.NoErrorf(t, err, "CloseSession")
	if wantErr {
		msg, ok := resp.Response.(*pb.CloseSessionResp_Error)
		require.True(t, ok, "CloseSession returned success response")
		t.Log(msg.Error.Error)
		return msg.Error.OpenPayChsInfo
	}

	msg, ok := resp.Response.(*pb.CloseSessionResp_MsgSuccess_)
	require.True(t, ok, "CloseSession returned error response")
	return msg.MsgSuccess.OpenPayChsInfo
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
	req := pb.AddContactReq{
		SessionID: sessionID,
		Peer:      peer,
	}
	resp, err := client.AddContact(ctx, &req)
	require.NoErrorf(t, err, "AddContact")
	_, ok := resp.Response.(*pb.AddContactResp_MsgSuccess_)
	require.True(t, ok, "AddContact returned error response")
}

func OpenPayCh(t *testing.T, sessionID string, parts, bal []string, wantErr bool) string {
	req := pb.OpenPayChReq{
		SessionID: sessionID,
		OpeningBalInfo: &pb.BalInfo{
			Currency: currency.ETH,
			Parts:    parts,
			Bal:      bal,
		},
		ChallengeDurSecs: 10,
	}
	resp, err := client.OpenPayCh(ctx, &req)
	require.NoErrorf(t, err, "OpenPayCh")

	if wantErr {
		errMsg, ok := resp.Response.(*pb.OpenPayChResp_Error)
		require.True(t, ok, "OpenPayCh returned success response")
		t.Log(errMsg)
		return ""
	}
	msg, ok := resp.Response.(*pb.OpenPayChResp_MsgSuccess_)
	assert.True(t, ok, "OpenPayCh returned error response")
	return msg.MsgSuccess.OpenedPayChInfo.ChID
}

func SubPayChProposal(t *testing.T, sessionID string) pb.Payment_API_SubPayChProposalsClient {
	subReq := pb.SubPayChProposalsReq{
		SessionID: sessionID,
	}
	subClient, err := client.SubPayChProposals(ctx, &subReq)
	require.NoErrorf(t, err, "SubPayChProposals")
	return subClient
}

func ReadPayChProposalNotif(t *testing.T, sessionID string, sub pb.Payment_API_SubPayChProposalsClient,
	wantErr bool) *pb.SubPayChProposalsResp_Notify_ {
	notifMsg, err := sub.Recv()

	if wantErr {
		require.Error(t, err)
		t.Log(err)
		return nil
	}
	require.NoErrorf(t, err, "subClient.Recv")
	notif, ok := notifMsg.Response.(*pb.SubPayChProposalsResp_Notify_)
	require.True(t, ok, "subClient.Recv returned error response")
	return notif
}

func RespondPayChProposal(t *testing.T, sessionID, proposalID string, accept bool, wantErr bool) {
	respondReq := pb.RespondPayChProposalReq{
		SessionID:  sessionID,
		ProposalID: proposalID,
		Accept:     accept,
	}
	resp, err := client.RespondPayChProposal(ctx, &respondReq)
	require.NoErrorf(t, err, "RespondPayChProposal")

	if wantErr {
		errMsg, ok := resp.Response.(*pb.RespondPayChProposalResp_Error)
		require.True(t, ok, "RespondPayChProposal returned success response")
		t.Log(errMsg)
		return
	}
	msg, ok := resp.Response.(*pb.RespondPayChProposalResp_MsgSuccess_)
	require.True(t, ok, "RespondPayChProposal returned error response")
	t.Log(msg)
}

func UnsubPayChProposal(t *testing.T, sessionID string, wantErr bool) {
	unsubReq := pb.UnsubPayChProposalsReq{
		SessionID: sessionID,
	}
	resp, err := client.UnsubPayChProposals(ctx, &unsubReq)
	require.NoErrorf(t, err, "UnsubPayChProposals")
	if wantErr {
		errMsg, ok := resp.Response.(*pb.UnsubPayChProposalsResp_Error)
		require.True(t, ok, "UnsubPayChProposals returned success response")
		t.Log(errMsg)
		return
	}
	_, ok := resp.Response.(*pb.UnsubPayChProposalsResp_MsgSuccess_)
	require.True(t, ok, "UnsubPayChProposals returned error response")
}

func SendPayChUpdate(t *testing.T, sessionID, chID, peerAlias, amount string, wantErr bool) {
	req := pb.SendPayChUpdateReq{
		SessionID: sessionID,
		ChID:      chID,
		Payee:     peerAlias,
		Amount:    amount,
	}
	resp, err := client.SendPayChUpdate(ctx, &req)
	require.NoErrorf(t, err, "SendPayChUpdate")
	if wantErr {
		errMsg, ok := resp.Response.(*pb.SendPayChUpdateResp_Error)
		require.True(t, ok, "SendPayChUpdate returned success response")
		t.Log(errMsg)
		return
	}
	_, ok := resp.Response.(*pb.SendPayChUpdateResp_MsgSuccess_)
	require.True(t, ok, "SendPayChUpdate returned error response")
}

func SubPayChUpdate(t *testing.T, sessionID, chID string) pb.Payment_API_SubPayChUpdatesClient {
	subReq := pb.SubpayChUpdatesReq{
		SessionID: sessionID,
		ChID:      chID,
	}
	subClient, err := client.SubPayChUpdates(ctx, &subReq)
	require.NoErrorf(t, err, "SubPayChUpdates")
	return subClient
}

func ReadPayChUpdateNotif(t *testing.T, sessionID, chID string,
	sub pb.Payment_API_SubPayChUpdatesClient) *pb.SubPayChUpdatesResp_Notify_ {
	notifMsg, err := sub.Recv()
	require.NoErrorf(t, err, "subClient.Recv")
	notif, ok := notifMsg.Response.(*pb.SubPayChUpdatesResp_Notify_)
	require.True(t, ok, "SendPayChUpdate returned error response")
	return notif
}

func RespondPayChUpdate(t *testing.T, sessionID, chID, updateID string, accept bool) {
	respondReq := pb.RespondPayChUpdateReq{
		SessionID: sessionID,
		UpdateID:  updateID,
		ChID:      chID,
		Accept:    accept,
	}
	_, err := client.RespondPayChUpdate(ctx, &respondReq)
	require.NoErrorf(t, err, "RespondPayChUpdate")
}

func RespondPayChUpdateExpectError(t *testing.T, sessionID, chID, updateID string, accept bool) {
	respondReq := pb.RespondPayChUpdateReq{
		SessionID: sessionID,
		UpdateID:  updateID,
		ChID:      chID,
		Accept:    accept,
	}
	resp, err := client.RespondPayChUpdate(ctx, &respondReq)
	require.NoError(t, err, "client.RespondPayChUpdate")
	respErrMsg, ok := resp.Response.(*pb.RespondPayChUpdateResp_Error)
	require.True(t, ok)
	require.NotZero(t, respErrMsg.Error.Error, "RespondPayChUpdate for closed channel notif")
	t.Log("Error responding to channel close notif", respErrMsg.Error.Error)
}

func UnsubPayChUpdate(t *testing.T, sessionID, chID string) {
	unsubReq := pb.UnsubPayChUpdatesReq{
		SessionID: sessionID,
		ChID:      chID,
	}
	_, err := client.UnsubPayChUpdates(ctx, &unsubReq)
	require.NoErrorf(t, err, "UnsubPayChUpdates")
}

func ClosePayCh(t *testing.T, sessionID, chID string) (isClosePayChSuccessful bool) {
	req := pb.ClosePayChReq{
		SessionID: sessionID,
		ChID:      chID,
	}
	resp, err := client.ClosePayCh(ctx, &req)
	require.NoErrorf(t, err, "ClosePayCh")
	_, isClosePayChSuccessful = resp.Response.(*pb.ClosePayChResp_MsgSuccess_)
	return isClosePayChSuccessful
} // nolint:gofumpt // unknown error, maybe a false positive
