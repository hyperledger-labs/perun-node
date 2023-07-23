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

//go:build integration
// +build integration

package grpc_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/app/payment"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/node"
	"github.com/hyperledger-labs/perun-node/node/nodetest"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

// This file contains integration tests for all the APIs. Start the ganache cli node using the
// below command before running the tests:
//
// ganache-cli -b 1 \
// --account="0x1fedd636dbc7e8d41a0622a2040b86fea8842cef9d4aa4c582aad00465b7acff,100000000000000000000" \
// --account="0xb0309c60b4622d3071fad3e16c2ce4d0b1e7758316c187754f4dd0cfb44ceb33,100000000000000000000"

var (
	grpcPort = "127.0.0.1:50001"

	// singleton instance of client and context that will be used for all tests.
	client pb.Payment_APIClient
	ctx    context.Context
)

func StartServer(t *testing.T, nodeCfg perun.NodeConfig, grpcPort string) {
	nodeAPI, err := node.New(nodeCfg)
	require.NoErrorf(t, err, "initializing nodeAPI")

	t.Log("Started ListenAndServePayChAPI")
	go func() {
		if err := grpc.ListenAndServePayChAPI(nodeAPI, grpcPort); err != nil {
			t.Logf("server returned with error: %v", err)
		}
	}()
	time.Sleep(1 * time.Second) // Wait for the server to start.
}

func Test_Integ_Role(t *testing.T) {
	// Deploy contracts on blockchain.
	ethereumtest.SetupContractsT(t, ethereumtest.ChainURL, ethereumtest.ChainID, ethereumtest.OnChainTxTimeout, false)

	// Run server in a go routine.
	// Do not include asset contracts, so that RegisterCurrency API can be tested.
	StartServer(t, nodetest.NewConfig(false), grpcPort)

	// Initialize client.
	conn, err := grpclib.Dial(grpcPort, grpclib.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "dialing to grpc server")
	t.Log("connected to server")
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
	var alicePeerID, bobPeerID *pb.PeerID
	var chETHnPRN, chETH, chPRN string

	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	aliceCfg := sessiontest.NewConfigT(t, prng)
	bobCfg := sessiontest.NewConfigT(t, prng)
	aliceCfgFile := sessiontest.NewConfigFileT(t, aliceCfg)
	bobCfgFile := sessiontest.NewConfigFileT(t, bobCfg)
	wg := &sync.WaitGroup{}

	// Run OpenSession for Alice, Bob in top level test, because cleanup functions
	// for removing the keystore directory, ID Provider files are registered to this
	// testing.T.

	// Alice Open Session.
	aliceSessionID = OpenSession(t, aliceCfgFile)
	t.Logf("%s session id is %s", aliceAlias, aliceSessionID)

	// Bob Open Session.
	bobSessionID = OpenSession(t, bobCfgFile)
	t.Logf("%s session id is %s", bobAlias, bobSessionID)

	t.Run("Node.RegisterCurrency", func(t *testing.T) {
		_, _, assetERC20s := ethereumtest.ContractAddrs()
		for tokenAddr := range assetERC20s {
			assetAddr := DeployAssetERC20(t, aliceSessionID, tokenAddr.String())
			require.NotZero(t, assetAddr)

			symbol := RegisterCurrency(t, tokenAddr.String(), assetAddr)
			require.Equal(t, "PRN", symbol)
		}
	})

	t.Run("GetPeerID", func(t *testing.T) {
		// Get own peer ID of alice and bob.
		alicePeerID = GetPeerID(t, aliceSessionID, perun.OwnAlias)
		alicePeerID.Alias = aliceAlias
		bobPeerID = GetPeerID(t, bobSessionID, perun.OwnAlias)
		bobPeerID.Alias = bobAlias
	})

	t.Run("AddPeerID", func(t *testing.T) {
		// Add bob's peer ID to alice and vice versa.
		AddPeerID(t, aliceSessionID, bobPeerID)
		AddPeerID(t, bobSessionID, alicePeerID)
	})

	t.Run("OpenCh_Sub_Unsub_Respond_Accept", func(t *testing.T) {
		openCh := func(t *testing.T, currencies []string, balances [][]string) (chID string) {
			// Alice proposes a channel and bob accepts.
			wg.Add(1)
			go func() {
				chID = OpenPayCh(t, aliceSessionID,
					currencies,
					[]string{perun.OwnAlias, bobAlias},
					balances, false)

				wg.Done()
			}()
			sub := SubPayChProposal(t, bobSessionID)
			notif := ReadPayChProposalNotif(t, sub, false)
			RespondPayChProposal(t, bobSessionID, notif.Notify.ProposalID, true, false)
			UnsubPayChProposal(t, bobSessionID, false)

			wg.Wait()
			return chID
		}
		tests := []struct {
			name       string
			chID       *string // Store chID to this variable.
			currencies []string
			balances   [][]string
		}{
			{"ETHnPRNCh", &chETHnPRN, []string{"ETH", "PRN"}, [][]string{{"1", "2"}, {"1", "2"}}},
			{"ETHCh", &chETH, []string{"ETH"}, [][]string{{"1", "2"}}},
			{"PRNCh", &chPRN, []string{"PRN"}, [][]string{{"1", "2"}}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				*tt.chID = openCh(t, tt.currencies, tt.balances)
			})
		}
	})

	t.Run("CloseSession_NoForce_ErrOpenPayChs", func(t *testing.T) {
		openPayChsInfo := CloseSession(t, aliceSessionID, false, true)
		require.Len(t, openPayChsInfo, 3)
		assert.Equal(t, chETHnPRN, openPayChsInfo[0].ChID)
		assert.Equal(t, chETH, openPayChsInfo[1].ChID)
		assert.Equal(t, chPRN, openPayChsInfo[2].ChID)

		openPayChsInfo = CloseSession(t, bobSessionID, false, true)
		require.Len(t, openPayChsInfo, 3)
		assert.Equal(t, chETHnPRN, openPayChsInfo[0].ChID)
		assert.Equal(t, chETH, openPayChsInfo[1].ChID)
		assert.Equal(t, chPRN, openPayChsInfo[2].ChID)
	})

	t.Run("OpenCh_Sub_Unsub_Respond_Reject", func(t *testing.T) {
		// Bob proposes a channel and alice accepts.
		wg.Add(1)
		go func() {
			OpenPayCh(t, bobSessionID,
				[]string{"ETH"}, []string{perun.OwnAlias, aliceAlias}, [][]string{{"1", "2"}}, true)

			wg.Done()
		}()
		sub := SubPayChProposal(t, aliceSessionID)
		notif := ReadPayChProposalNotif(t, sub, false)
		RespondPayChProposal(t, aliceSessionID, notif.Notify.ProposalID, false, false)
		UnsubPayChProposal(t, aliceSessionID, false)

		wg.Wait()
	})

	t.Run("Send_Request_Payments", func(t *testing.T) {
		sendPayment := func(t *testing.T, chID string, payments []payment.Payment, accept bool) {
			// Bob sends a payment and alice accepts.
			wg.Add(1)
			go func() {
				wantErr := !accept // if accept = true, then should not error; if false, then should error.
				SendPayChUpdate(t, bobSessionID, chID, payments, wantErr)

				wg.Done()
			}()

			sub := SubPayChUpdate(t, aliceSessionID, chID)
			notif := ReadPayChUpdateNotif(t, sub)
			assert.EqualValues(t, perun.ChUpdateTypeOpen, notif.Notify.Type)
			RespondPayChUpdate(t, aliceSessionID, chID, notif.Notify.UpdateID, accept)
			UnsubPayChUpdate(t, aliceSessionID, chID)

			wg.Wait()
		}

		//nolint:govet	// it is okay to use unkeyed fields in Payment struct.
		tests := []struct {
			name     string
			chID     string
			payments []payment.Payment
		}{
			{"send_ETH_on_ETHnPRNCh", chETHnPRN, []payment.Payment{{"ETH", aliceAlias, "0.1"}}},
			{"req_ETH_on_ETHnPRNCh", chETHnPRN, []payment.Payment{{"ETH", perun.OwnAlias, "0.1"}}},
			{"send_PRN_on_ETHnPRNCh", chETHnPRN, []payment.Payment{{"PRN", aliceAlias, "0.1"}}},
			{"req_PRN_on_ETHnPRNCh", chETHnPRN, []payment.Payment{{"PRN", perun.OwnAlias, "0.1"}}},

			{"send_ETH_req_PRN_on_ETHnPRNCh", chETHnPRN, []payment.Payment{
				{"ETH", aliceAlias, "0.1"}, {"PRN", perun.OwnAlias, "0.1"},
			}},

			{"req_ETH_send_PRN_on_ETHnPRNCh", chETHnPRN, []payment.Payment{
				{"ETH", perun.OwnAlias, "0.1"}, {"PRN", aliceAlias, "0.1"},
			}},

			{"send_ETH_send_PRN_on_ETHnPRNCh", chETHnPRN, []payment.Payment{
				{"ETH", aliceAlias, "0.1"}, {"PRN", aliceAlias, "0.1"},
			}},

			{"req_ETH_req_PRN_on_ETHnPRNCh", chETHnPRN, []payment.Payment{
				{"ETH", perun.OwnAlias, "0.1"}, {"PRN", perun.OwnAlias, "0.1"},
			}},

			{"send_ETH_on_ETHCh", chETH, []payment.Payment{{"ETH", aliceAlias, "0.1"}}},
			{"req_ETH_on_ETHCh", chETH, []payment.Payment{{"ETH", perun.OwnAlias, "0.1"}}},
			{"send_PRN_on_PRNCh", chPRN, []payment.Payment{{"PRN", aliceAlias, "0.1"}}},
			{"req_PRN_on_PRNCh", chPRN, []payment.Payment{{"PRN", perun.OwnAlias, "0.1"}}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Test both cases that a payment is accepted / rejected.
				sendPayment(t, tt.chID, tt.payments, true)
				sendPayment(t, tt.chID, tt.payments, false)
			})
		}
	})

	isClosePayChSuccessful := make(chan bool, 1)
	t.Run("Close_Sub_Unsub", func(t *testing.T) {
		closeCh := func(t *testing.T, chID string) {
			// Bob initiates channel close, alice receives an update request for "finalized state" that he acknowledges.
			// The channel is then closed on-chain by bob and the funds are withdrawn immediately.
			wg.Add(2)
			go func() {
				sub := SubPayChUpdate(t, aliceSessionID, chID)
				notif := ReadPayChUpdateNotif(t, sub)
				assert.EqualValues(t, perun.ChUpdateTypeFinal, notif.Notify.Type)
				RespondPayChUpdate(t, aliceSessionID, chID, notif.Notify.UpdateID, true)
				notif = ReadPayChUpdateNotif(t, sub)
				assert.EqualValues(t, perun.ChUpdateTypeClosed, notif.Notify.Type)

				wg.Done()
			}()
			go func() {
				sub := SubPayChUpdate(t, bobSessionID, chID)
				notif := ReadPayChUpdateNotif(t, sub)
				assert.EqualValues(t, perun.ChUpdateTypeClosed, notif.Notify.Type)
				RespondPayChUpdateExpectError(t, bobSessionID, chID, notif.Notify.UpdateID, true)

				wg.Done()
			}()

			time.Sleep(1 * time.Second) // Wait for the subscriptions to be made.
			isClosePayChSuccessful <- ClosePayCh(t, bobSessionID, chID)

			wg.Wait()
		}
		tests := []struct {
			chID string
		}{
			{chETHnPRN}, {chETH}, {chPRN},
		}

		for _, tt := range tests {
			t.Run(tt.chID, func(t *testing.T) {
				closeCh(t, tt.chID)
				require.True(t, <-isClosePayChSuccessful)
			})
		}
	})

	t.Run("CloseSession_NoForce_", func(t *testing.T) {
		openPayChsInfo := CloseSession(t, aliceSessionID, false, false)
		require.Len(t, openPayChsInfo, 0)
		openPayChsInfo = CloseSession(t, bobSessionID, false, false)
		require.Len(t, openPayChsInfo, 0)
	})
	t.Run("APIs error when session is closed", func(t *testing.T) {
		OpenPayCh(t, bobSessionID,
			[]string{"ETH"}, []string{perun.OwnAlias, aliceAlias}, [][]string{{"1", "2"}}, true)
		sub := SubPayChProposal(t, aliceSessionID)
		ReadPayChProposalNotif(t, sub, true)
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

func DeployAssetERC20(t *testing.T, sessionID, tokenAddr string) string {
	req := pb.DeployAssetERC20Req{
		SessionID: sessionID,
		TokenAddr: tokenAddr,
	}
	resp, err := client.DeployAssetERC20(ctx, &req)
	require.NoErrorf(t, err, "DeployAssetERC20")
	msg, ok := resp.Response.(*pb.DeployAssetERC20Resp_MsgSuccess_)
	require.True(t, ok, "DeployAssetERC20 returned error response")
	return msg.MsgSuccess.AssetAddr
}

func RegisterCurrency(t *testing.T, tokenAddr, assetAddr string) string {
	req := pb.RegisterCurrencyReq{
		TokenAddr: tokenAddr,
		AssetAddr: assetAddr,
	}
	resp, err := client.RegisterCurrency(ctx, &req)
	require.NoErrorf(t, err, "RegisterCurrency")
	msg, ok := resp.Response.(*pb.RegisterCurrencyResp_MsgSuccess_)
	require.True(t, ok, "RegisterCurrency returned error response")
	return msg.MsgSuccess.Symbol
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
		t.Log(msg.Error)
		addInfo, ok := msg.Error.AddInfo.(*pb.MsgError_ErrInfoFailedPreCondUnclosedChs)
		if !ok {
			return nil
		}
		return addInfo.ErrInfoFailedPreCondUnclosedChs.Chs
	}

	msg, ok := resp.Response.(*pb.CloseSessionResp_MsgSuccess_)
	require.True(t, ok, "CloseSession returned error response")
	return msg.MsgSuccess.OpenPayChsInfo
}

func GetPeerID(t *testing.T, sessionID string, alias string) *pb.PeerID {
	req := pb.GetPeerIDReq{
		SessionID: sessionID,
		Alias:     alias,
	}
	resp, err := client.GetPeerID(ctx, &req)
	require.NoErrorf(t, err, "GetPeerID")
	msg, ok := resp.Response.(*pb.GetPeerIDResp_MsgSuccess_)
	require.True(t, ok, "GetPeerID returned error response")
	return msg.MsgSuccess.PeerID
}

func AddPeerID(t *testing.T, sessionID string, peerID *pb.PeerID) {
	req := pb.AddPeerIDReq{
		SessionID: sessionID,
		PeerID:    peerID,
	}
	resp, err := client.AddPeerID(ctx, &req)
	require.NoErrorf(t, err, "AddPeerID")
	_, ok := resp.Response.(*pb.AddPeerIDResp_MsgSuccess_)
	require.True(t, ok, "AddPeerID returned error response")
}

func OpenPayCh(t *testing.T, sessionID string, currencies, parts []string, bals [][]string, wantErr bool) string {
	req := pb.OpenPayChReq{
		SessionID: sessionID,
		OpeningBalInfo: pb.FromBalInfo(perun.BalInfo{
			Currencies: currencies,
			Parts:      parts,
			Bals:       bals,
		}),
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

func ReadPayChProposalNotif(t *testing.T, sub pb.Payment_API_SubPayChProposalsClient,
	wantErr bool,
) *pb.SubPayChProposalsResp_Notify_ {
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

func SendPayChUpdate(t *testing.T, sessionID, chID string, payments []payment.Payment, wantErr bool) {
	req := pb.SendPayChUpdateReq{
		SessionID: sessionID,
		ChID:      chID,
		Payments:  pb.FromPayments(payments),
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

func ReadPayChUpdateNotif(t *testing.T,
	sub pb.Payment_API_SubPayChUpdatesClient,
) *pb.SubPayChUpdatesResp_Notify_ {
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
	require.NotZero(t, respErrMsg.Error, "RespondPayChUpdate for closed channel notif")
	t.Log("Error responding to channel close notif", respErrMsg.Error)
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
}
