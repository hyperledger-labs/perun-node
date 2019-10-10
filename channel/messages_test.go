// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
//     https://github.com/direct-state-transfer/dst-go/NOTICE
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

package channel

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
)

var testTime time.Time
var contractStoreVersionForTest = []byte("8b9df51d7a7343a92133f24aaa1ba544d92eac1ccd66c731aa243e3273cb0f69")

func init() {
	testTime, _ = time.Parse(
		time.RFC3339, "2012-11-01T22:08:41+00:00")
}

type jsonMsgTypeForTest struct{}

var MsgIDForTest MessageID = "junk-message-type-for-test"

//ChReadMock sends a mock packet to read
//It takes an initialized channel and should be invoked as go-routine
func ChReadMock(ch *genericChannelAdapter, responseMsg jsonMsgPacket, wg *sync.WaitGroup) {

	defer wg.Done()

	//channel is guarded by mutex in Read/Write and other calls,
	//so use pipes directly

	ch.readHandlerPipe.msgPacket <- responseMsg
}

//ChWriteMock waits for a message write to occur, responds to it
//It takes an initialized channel and should be invoked as go-routine
func ChWriteMock(t *testing.T, ch *genericChannelAdapter, expectRequest chMsgPkt, expectMatchInMock bool, respondWithErr error, wg *sync.WaitGroup) {

	defer wg.Done()

	//channel is guarded by mutex in Read/Write and other calls,
	//so use pipes directly

	inMsg1 := <-ch.writeHandlerPipe.msgPacket

	//Write expects a response from handler with error status
	inMsg1.err = respondWithErr
	ch.writeHandlerPipe.msgPacket <- inMsg1

	if respondWithErr != nil {
		return
	}

	if expectMatchInMock != compareMsg(inMsg1.message, expectRequest) {
		t.Fatalf("inMsg1 = %v, want %v", inMsg1.message, expectRequest)
	}
}

//ChWriteReadMock waits for a message write to occur, responds to it
//Then also it sends a mock packet to read
//It takes an initialized channel and should be invoked as go-routine
func ChWriteReadMock(t *testing.T, ch *genericChannelAdapter, expectRequestMsg chMsgPkt, responseMsg jsonMsgPacket, respondWithErr error, expectRequestMsgMatch bool, wg *sync.WaitGroup) {

	defer wg.Done()

	//channel is guarded by mutex in Read/Write and other calls,
	//so use pipes directly

	inMsg1 := <-ch.writeHandlerPipe.msgPacket

	//Write expects a response from handler with error status
	inMsg1.err = respondWithErr
	ch.writeHandlerPipe.msgPacket <- inMsg1

	if respondWithErr != nil {
		return
	}

	if expectRequestMsgMatch != compareMsg(inMsg1.message, expectRequestMsg) {
		t.Errorf("inMsg1 = %+v, want %+v", inMsg1.message, expectRequestMsg)
	}

	ch.readHandlerPipe.msgPacket <- responseMsg
}

func compareMsg(msg1, msg2 chMsgPkt) bool {

	if (reflect.TypeOf(msg1) == reflect.TypeOf(msg2)) &&
		(msg1.MessageID == msg2.MessageID) &&
		reflect.DeepEqual(msg1.Message, msg2.Message) {
		return true
	}
	return false
}
func Test_chMsgPkt_UnmarshalJSON(t *testing.T) {

	//TestContract object with gas units and hash go file values stripped off.
	testContractLibSignature := contract.Store.LibSignatures()
	testContractLibSignature.GasUnits = 0
	testContractLibSignature.HashGoFile = ""

	type args struct {
		data []byte
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantMsgPkt chMsgPkt
	}{
		{
			name: "invalid_chRawMsgPkt",
			args: args{
				data: []byte(`{sasa`), //invalid raw json
			},
			wantErr: true,
		},
		{
			name: "valid_MsgIdentityRequest",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgIdentityRequest",
				"message":{
					"id":{
						"on_chain_id":"0x932a74da117eb9288ea759487360cd700e7777e1",
						"listener_ip_addr":"http://localhost:1250",
						"listener_endpoint":"/state-channel",
						"KeyStore":{},
						"Password":"some-non-empty-password"}},
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: false,
			wantMsgPkt: chMsgPkt{
				Version:   "1.0",
				MessageID: MsgIdentityRequest,
				Message: jsonMsgIdentity{
					ID: identity.OffChainID{
						OnChainID:        types.HexToAddress("932a74da117eb9288ea759487360cd700e7777e1"),
						ListenerIPAddr:   "http://localhost:1250",
						ListenerEndpoint: "/state-channel",
						KeyStore:         nil,
						Password:         ""},
				}},
		},
		{
			name: "invalid_MsgIdentityRequest",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgIdentityRequest",
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "valid_MsgIdentityResponse",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgIdentityResponse",
				"message":{
					"id":{
						"on_chain_id":"0x932a74da117eb9288ea759487360cd700e7777e1",
						"listener_ip_addr":"http://localhost:1250",
						"listener_endpoint":"/state-channel",
						"KeyStore":{},
						"Password":"some-non-empty-password"}},
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: false,
			wantMsgPkt: chMsgPkt{
				Version:   "1.0",
				MessageID: MsgIdentityResponse,
				Message: jsonMsgIdentity{
					ID: identity.OffChainID{
						OnChainID:        types.HexToAddress("932a74da117eb9288ea759487360cd700e7777e1"),
						ListenerIPAddr:   "http://localhost:1250",
						ListenerEndpoint: "/state-channel",
						KeyStore:         nil,
						Password:         ""},
				}},
		},
		{
			name: "invalid_MsgIdentityResponse",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgIdentityResponse",
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "valid_MsgNewChannelRequest",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgNewChannelRequest",
				"message":{
						"status":"require",
						"reason":"ok"},
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: false,
			wantMsgPkt: chMsgPkt{
				Version:   "1.0",
				MessageID: MsgNewChannelRequest,
				Message: jsonMsgNewChannel{
					Status: MessageStatusRequire,
					Reason: "ok",
				}},
		},
		{
			name: "invalid_MsgNewChannelRequest",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgNewChannelRequest",
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "valid_MsgNewChannelResponse",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgNewChannelResponse",
				"message":{
						"status":"accept",
						"reason":"ok"},
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: false,
			wantMsgPkt: chMsgPkt{
				Version:   "1.0",
				MessageID: MsgNewChannelResponse,
				Message: jsonMsgNewChannel{
					Status: MessageStatusAccept,
					Reason: "ok",
				}},
		},
		{
			name: "invalid_MsgNewChannelResponse",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgNewChannelResponse",
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "valid_MsgSessionIdRequest",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgSessionIdRequest",
				"message":{
					"sid":{
						"addr_sender":"0x932a74da117eb9288ea759487360cd700e7777e1",
						"nonce_sender":"qxLNNA==",
						"sid_sender_part":"qxLNNA==",
						"addr_receiver":"0x0000000000000000000000000000000000000000",
						"nonce_receiver":null,
						"sid_receiver_part":null,
						"sid_complete":null,
						"locked":false
					},
					"status":"require"
				},
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: false,
			wantMsgPkt: chMsgPkt{
				Version:   "1.0",
				MessageID: MsgSessionIDRequest,
				Message: jsonMsgSessionID{
					Sid: SessionID{
						AddrSender:    aliceID.OnChainID,
						NonceSender:   types.Hex2Bytes("ab12cd34"),
						SidSenderPart: types.Hex2Bytes("ab12cd34"),
						Locked:        false,
					},
					Status: "require",
				}},
		},
		{
			name: "invalid_MsgSessionIdRequest",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgSessionIdRequest",
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "valid_MsgSessionIdResponse",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgSessionIdResponse",
				"message":{
					"sid":{
						"addr_sender":"0x932a74da117eb9288ea759487360cd700e7777e1",
						"nonce_sender":"qxLNNA==",
						"sid_sender_part":"qxLNNA==",
						"addr_receiver":"0x815430d6ea7275317d09199a5a5675f017e011ef",
						"nonce_receiver":"qxLNNA==",
						"sid_receiver_part":"qxLNNA==",
						"sid_complete":121,
						"locked":true
					},
					"status":"require"
				},
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: false,
			wantMsgPkt: chMsgPkt{
				Version:   "1.0",
				MessageID: MsgSessionIDResponse,
				Message: jsonMsgSessionID{
					Sid: SessionID{
						AddrSender:      aliceID.OnChainID,
						NonceSender:     types.Hex2Bytes("ab12cd34"),
						SidSenderPart:   types.Hex2Bytes("ab12cd34"),
						SidReceiverPart: types.Hex2Bytes("ab12cd34"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   types.Hex2Bytes("ab12cd34"),
						SidComplete:     big.NewInt(121),
						Locked:          true,
					},
					Status: "require",
				}},
		},
		{
			name: "invalid_MsgSessionIdResponse",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgSessionIdResponse",
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "valid_MsgContractAddrRequest",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgContractAddrRequest",
				"message":{
					"addr":"0x847b3655b5beb829cb3cd41c00a27648de737c39",
					"contract_type":{
						"Name":"LibSignatures",
						"Version":"0.0.1",
						"GasUnits":4000000,
						"HashSolFile":"359e2e9f7bacdcefc6962c46182aba7f16b8b0a8314468ca8dd88edd25299209",
						"HashGoFile":"1b6a7102bd6726168ea973b0bb2107655b638bc09b702cffe05a9d90e67373d7",
						"HashBinRuntimeFile":"3c0f29dfe76fd55ab0b023b26c97d2a306805a03788277cd0c7d2817cb7a9bf9"
						},
					"status":"require"
				},
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: false,
			wantMsgPkt: chMsgPkt{
				Version:   "1.0",
				MessageID: MsgContractAddrRequest,
				Message: jsonMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: testContractLibSignature,
					Status:       "require",
				}},
		},
		{
			name: "invalid_MsgContractAddrRequest",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgContractAddrRequest",
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "valid_MsgContractAddrResponse",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgContractAddrResponse",
				"message":{
					"addr":"0x847b3655b5beb829cb3cd41c00a27648de737c39",
					"contract_type":{
						"Name":"LibSignatures",
						"Version":"0.0.1",
						"GasUnits":4000000,
						"HashSolFile":"359e2e9f7bacdcefc6962c46182aba7f16b8b0a8314468ca8dd88edd25299209",
						"HashGoFile":"1b6a7102bd6726168ea973b0bb2107655b638bc09b702cffe05a9d90e67373d7",
						"HashBinRuntimeFile":"3c0f29dfe76fd55ab0b023b26c97d2a306805a03788277cd0c7d2817cb7a9bf9"
						},
					"status":"require"
				},
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: false,
			wantMsgPkt: chMsgPkt{
				Version:   "1.0",
				MessageID: MsgContractAddrResponse,
				Message: jsonMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: testContractLibSignature,
					Status:       "require",
				}},
		},
		{
			name: "invalid_MsgContractAddrResponse",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgContractAddrResponse",
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "valid_MsgMSCBaseStateRequest",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgMSCBaseStateRequest",
				"message":{
					"signed_state_val":{
						"ms_contract_state":{
							"vpc_address":"0x932a74da117eb9288ea759487360cd700e7777e1",
							"sid":123456,
							"blocked_sender":10,
							"blocked_receiver":20,
							"version":1
							},
						"sign_sender":"c2lnbi1zZW5kZXI="
					},
					"status":"require"
				},
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: false,
			wantMsgPkt: chMsgPkt{
				Version:   "1.0",
				MessageID: MsgMSCBaseStateRequest,
				Message: jsonMsgMSCBaseState{
					SignedStateVal: MSCBaseStateSigned{
						MSContractBaseState: MSCBaseState{
							VpcAddress:      aliceID.OnChainID,
							Sid:             big.NewInt(123456),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(20),
							Version:         big.NewInt(1),
						},
						SignSender: []byte("sign-sender"),
					},
					Status: "require",
				}},
		},
		{
			name: "invalid_MsgMSCBaseStateRequest",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgMSCBaseStateRequest",
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "valid_MsgMSCBaseStateResponse",
			args: args{
				data: []byte(`{
				"version":"1.0",
				"message_id":"MsgMSCBaseStateResponse",
				"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "valid_MsgMSCBaseStateRequest",
			args: args{
				data: []byte(`{
					"version":"1.0",
					"message_id":"MsgVPCStateRequest",
					"message":{
						"signed_state_val":{
							"vpc_state":{
								"id":"c2FtcGxlLWlk",
								"version":1,
								"blocked_alice":10,
								"blocked_bob":20
							},
							"sign_sender":"c2lnbi1zZW5kZXI=",
							"sign_receiver":null
						},
						"status":"require"
					},
					"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: false,
			wantMsgPkt: chMsgPkt{
				Version:   "1.0",
				MessageID: MsgVPCStateRequest,
				Message: jsonMsgVPCState{
					SignedStateVal: VPCStateSigned{
						VPCState: VPCState{
							ID:              []byte("sample-id"),
							Version:         big.NewInt(1),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(20),
						},
						SignSender: []byte("sign-sender"),
					},
					Status: "require",
				}},
		},
		{
			name: "invalid_MsgMSCBaseStateRequest",
			args: args{
				data: []byte(`{
					"version":"1.0",
					"message_id":"MsgVPCStateRequest",
					"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "valid_MsgMSCBaseStateResponse",
			args: args{
				data: []byte(`{
					"version":"1.0",
					"message_id":"MsgVPCStateResponse",
					"message":{
						"signed_state_val":{
							"vpc_state":{
								"id":"c2FtcGxlLWlk",
								"version":1,
								"blocked_alice":10,
								"blocked_bob":20
							},
						"sign_sender":"c2lnbi1zZW5kZXI=",
						"sign_receiver":"c2lnbi1yZWNlaXZlcg=="
					},
					"status":"require"
					},
					"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: false,
			wantMsgPkt: chMsgPkt{
				Version:   "1.0",
				MessageID: MsgVPCStateResponse,
				Message: jsonMsgVPCState{
					SignedStateVal: VPCStateSigned{
						VPCState: VPCState{
							ID:              []byte("sample-id"),
							Version:         big.NewInt(1),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(20),
						},
						SignSender:   []byte("sign-sender"),
						SignReceiver: []byte("sign-receiver"),
					},
					Status: "require",
				}},
		},
		{
			name: "invalid_MsgMSCBaseStateResponse",
			args: args{
				data: []byte(`{
					"version":"1.0",
					"message_id":"MsgVPCStateResponse",
					"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
		{
			name: "invalid_MessageID",
			args: args{
				data: []byte(`{
					"version":"1.0",
					"message_id":"unknown-msg-id",
					"timestamp":"0001-01-01T00:00:00Z"}`),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// //use for generating raw json in test cases
			// var err error
			// tt.args.data, err = json.Marshal(tt.wantMsgPkt)
			// if err != nil {
			// 	t.Fatalf("json.Marshall error - %v", err.Error())
			// }

			// fmt.Printf("\n\n%s\n\n", tt.args.data)

			msgPkt := chMsgPkt{}
			if err := msgPkt.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Fatalf("chMsgPkt.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(msgPkt, tt.wantMsgPkt) {
				t.Errorf("jsonMsg unmarshall() got %v, want %v", msgPkt, tt.wantMsgPkt)
			}

		})
	}
}
func Test_channel_IdentityRequest(t *testing.T) {
	type args struct {
		selfID identity.OffChainID
	}
	tests := []struct {
		name              string
		args              args
		expectRequest     chMsgPkt
		expectMatchInMock bool
		mockResponse      jsonMsgPacket
		responseError     error
		wantErr           bool
		wantPeerID        identity.OffChainID
	}{
		{
			name: "valid-1",
			args: args{
				selfID: aliceID,
			},
			expectRequest: chMsgPkt{
				MessageID: MsgIdentityRequest,
				Message:   jsonMsgIdentity{aliceID},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgIdentityResponse,
					Message: jsonMsgIdentity{
						ID: aliceID,
					},
				}},
			wantErr:    false,
			wantPeerID: aliceID,
		},
		{
			name: "invalid-message-id",
			args: args{
				selfID: aliceID,
			},
			expectRequest: chMsgPkt{
				MessageID: MsgIdentityRequest,
				Message: jsonMsgIdentity{
					ID: aliceID,
				}},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			args: args{
				selfID: aliceID,
			},
			expectRequest: chMsgPkt{
				MessageID: MsgIdentityRequest,
				Message: jsonMsgIdentity{
					ID: aliceID,
				}},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgIdentityResponse,
					Message:   nil,
				}},
			wantErr: true,
		},
		{
			name:          "write-error",
			args:          args{},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
		{
			name: "read-error",
			args: args{},
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteReadMock(t, adapter, tt.expectRequest, tt.mockResponse, tt.responseError, tt.expectMatchInMock, wg)

			gotPeerID, err := ch.IdentityRequest(tt.args.selfID)
			if err != nil {
				t.Logf("channel.IdentityRequest() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if !reflect.DeepEqual(gotPeerID, tt.wantPeerID) {
				t.Errorf("channel.IdentityRequest() = %v, want %v", gotPeerID, tt.wantPeerID)
			}
			wg.Wait() //Wait for mock object to exit
		})
	}
}
func Test_channel_IdentityRead(t *testing.T) {
	tests := []struct {
		name            string
		mockResponse    jsonMsgPacket
		wantErr         bool
		wantPeerID      identity.OffChainID
		wantPeerIDMatch bool
	}{
		{
			name: "valid-1",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgIdentityRequest,
					Message: jsonMsgIdentity{
						ID: aliceID,
					},
				}},
			wantErr:         false,
			wantPeerID:      aliceID,
			wantPeerIDMatch: true,
		},
		{
			name: "read-error",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
		{
			name: "id-mismatch",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgIdentityRequest,
					Message: jsonMsgIdentity{
						ID: bobID,
					},
				}},
			wantErr:         false,
			wantPeerID:      aliceID,
			wantPeerIDMatch: false,
		},
		{
			name: "invalid-message-id",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-id"),
				}},
			wantErr:         true,
			wantPeerID:      aliceID,
			wantPeerIDMatch: false,
		},
		{
			name: "invalid-message",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgIdentityRequest,
					Message:   nil,
				}},
			wantErr:         true,
			wantPeerID:      aliceID,
			wantPeerIDMatch: false,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChReadMock(adapter, tt.mockResponse, wg)

			gotPeerID, err := ch.IdentityRead()
			if err != nil {
				t.Logf("channel.IdentityRead() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if !reflect.DeepEqual(gotPeerID, tt.wantPeerID) && tt.wantPeerIDMatch {
				t.Errorf("channel.IdentityRead() = %v, want %v", gotPeerID, tt.wantPeerID)
			}

			wg.Wait()
		})
	}
}

func Test_channel_IdentityRespond(t *testing.T) {
	type args struct {
		selfID identity.OffChainID
	}
	tests := []struct {
		name              string
		args              args
		expectResponse    chMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "valid-1",
			args: args{
				selfID: aliceID,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgIdentityResponse,
				Message: jsonMsgIdentity{
					ID: aliceID,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           true,
		},
		{
			name: "id-mismatch",
			args: args{
				selfID: bobID,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgIdentityResponse,
				Message: jsonMsgIdentity{
					ID: aliceID,
				},
			},
			expectMatchInMock: false,
			responseError:     nil,
			wantErr:           true,
		},
		{
			name: "wrong-msg-id",
			args: args{
				selfID: aliceID,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgIDForTest,
			},
			expectMatchInMock: false,
			responseError:     nil,
			wantErr:           true,
		},
		{
			name: "wrong-message-packet-type",
			args: args{
				selfID: aliceID,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgIdentityResponse,
				Message:   jsonMsgTypeForTest{},
			},
			expectMatchInMock: false,
			responseError:     nil,
			wantErr:           true,
		},
		{
			name: "write-error",
			args: args{
				selfID: aliceID,
			},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteMock(t, adapter, tt.expectResponse, tt.expectMatchInMock, tt.responseError, wg)

			err := ch.IdentityRespond(tt.args.selfID)
			if err != nil {
				t.Logf("channel.RespondToContractAddr() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			wg.Wait()
		})
	}
}

func Test_channel_NewChannelRequest(t *testing.T) {
	type args struct {
		msgProtocolVersion   string
		contractStoreVersion []byte
	}
	tests := []struct {
		name              string
		args              args
		expectRequest     chMsgPkt
		expectMatchInMock bool
		mockResponse      jsonMsgPacket
		responseError     error
		wantErr           bool
		wantAccept        MessageStatus
		wantReason        string
	}{
		{
			name: "accept",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0"),
			},
			expectRequest: chMsgPkt{
				MessageID: MsgNewChannelRequest,
				Message: jsonMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgNewChannelResponse,
					Message: jsonMsgNewChannel{
						MsgProtocolVersion:   "1.0",
						ContractStoreVersion: []byte("v1.0"),
						Status:               MessageStatusAccept,
						Reason:               "",
					},
				}},
			wantErr:    false,
			wantAccept: MessageStatusAccept,
			wantReason: "",
		},
		{
			name: "decline",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0"),
			},
			expectRequest: chMsgPkt{
				MessageID: MsgNewChannelRequest,
				Message: jsonMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgNewChannelResponse,
					Message: jsonMsgNewChannel{
						MsgProtocolVersion:   "1.0",
						ContractStoreVersion: []byte("v1.0"),
						Status:               MessageStatusDecline,
						Reason:               "wrong-version",
					},
				}},
			wantErr:    false,
			wantAccept: MessageStatusDecline,
			wantReason: "wrong-version",
		},
		{
			name: "msgProtocolVersion-modified-by-peer",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0"),
			},
			expectRequest: chMsgPkt{
				MessageID: MsgNewChannelRequest,
				Message: jsonMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgNewChannelResponse,
					Message: jsonMsgNewChannel{
						MsgProtocolVersion:   "2.0",
						ContractStoreVersion: []byte("v1.0"),
						Status:               MessageStatusAccept,
						Reason:               "",
					},
				}},
			wantErr: true,
		},
		{
			name: "contractStoreVersion-modified-by-peer",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0"),
			},
			expectRequest: chMsgPkt{
				MessageID: MsgNewChannelRequest,
				Message: jsonMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgNewChannelResponse,
					Message: jsonMsgNewChannel{
						MsgProtocolVersion:   "1.0",
						ContractStoreVersion: []byte("v2.0"),
						Status:               MessageStatusAccept,
						Reason:               "",
					},
				}},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0"),
			},
			expectRequest: chMsgPkt{
				MessageID: MsgNewChannelRequest,
				Message: jsonMsgNewChannel{
					ContractStoreVersion: []byte("v1.0"),
					MsgProtocolVersion:   "1.0",
					Status:               "require",
					Reason:               "",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0"),
			},
			expectRequest: chMsgPkt{
				MessageID: MsgNewChannelRequest,
				Message: jsonMsgNewChannel{
					ContractStoreVersion: []byte("v1.0"),
					MsgProtocolVersion:   "1.0",
					Status:               "require",
					Reason:               "",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgNewChannelResponse,
					Message:   nil,
				}},
			wantErr: true,
		},
		{
			name:          "write-error",
			args:          args{},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
		{
			name: "read-error",
			args: args{},
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteReadMock(t, adapter, tt.expectRequest, tt.mockResponse, tt.responseError, tt.expectMatchInMock, wg)

			gotAccept, gotReason, err := ch.NewChannelRequest(tt.args.msgProtocolVersion, tt.args.contractStoreVersion)
			if err != nil {
				t.Logf("channel.NewChannelRequest() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if gotAccept != tt.wantAccept {
				t.Errorf("channel.NewChannelRequest() gotAccept = %v, want %v", gotAccept, tt.wantAccept)
			}
			if gotReason != tt.wantReason {
				t.Errorf("channel.NewChannelRequest() gotReason = %v, want %v", gotReason, tt.wantReason)
			}
			wg.Wait()
		})
	}
}

func Test_channel_NewChannelRead(t *testing.T) {
	tests := []struct {
		name                     string
		mockResponse             jsonMsgPacket
		wantErr                  bool
		wantMsgProtocolVersion   string
		wantContractStoreVersion []byte
	}{
		{
			name: "valid-1",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgNewChannelRequest,
					Message: jsonMsgNewChannel{
						MsgProtocolVersion:   "1.0",
						ContractStoreVersion: contractStoreVersionForTest,
						Status:               MessageStatusRequire,
					},
				}},
			wantErr:                  false,
			wantMsgProtocolVersion:   "1.0",
			wantContractStoreVersion: contractStoreVersionForTest,
		},
		{
			name: "read-error",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error"),
			},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-id"),
				}},
			wantErr:                  true,
			wantContractStoreVersion: contractStoreVersionForTest,
		},
		{
			name: "invalid-message",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgNewChannelRequest,
					Message:   nil,
				}},
			wantErr:                  true,
			wantContractStoreVersion: contractStoreVersionForTest,
		},
		{
			name: "invalid-status",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgNewChannelRequest,
					Message: jsonMsgNewChannel{
						MsgProtocolVersion:   "1.0",
						ContractStoreVersion: contractStoreVersionForTest,
						Status:               MessageStatus("invalid-message-status"),
					},
				}},
			wantErr: true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChReadMock(adapter, tt.mockResponse, wg)

			msgProtocolVersion, contractStoreVersion, err := ch.NewChannelRead()
			if err != nil {
				t.Logf("channel.NewChannelRead() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}

			if !reflect.DeepEqual(contractStoreVersion, tt.wantContractStoreVersion) {
				t.Errorf("channel.NewChannelRead() contractStoreVersion = %v, wantContractStoreVersion %v", contractStoreVersion, tt.wantContractStoreVersion)
			}

			if !reflect.DeepEqual(msgProtocolVersion, tt.wantMsgProtocolVersion) {
				t.Errorf("channel.NewChannelRead() msgProtocolVersion = %v, wantMsgProtocolVersion %v", msgProtocolVersion, tt.wantMsgProtocolVersion)
			}

			wg.Wait()
		})
	}
}

func Test_channel_NewChannelRespond(t *testing.T) {
	type args struct {
		msgProtocolVersion   string
		contractStoreVersion []byte
		accept               MessageStatus
		reason               string
	}
	tests := []struct {
		name              string
		args              args
		expectResponse    chMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "accept",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: contractStoreVersionForTest,
				accept:               MessageStatusAccept,
				reason:               "",
			},
			expectResponse: chMsgPkt{
				MessageID: MsgNewChannelResponse,
				Message: jsonMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               MessageStatusAccept,
					Reason:               "",
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "decline",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: contractStoreVersionForTest,
				accept:               MessageStatusDecline,
				reason:               "",
			},
			expectResponse: chMsgPkt{
				MessageID: MsgNewChannelResponse,
				Message: jsonMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               MessageStatusDecline,
					Reason:               "",
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "decline-expect-wrong-reason",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: contractStoreVersionForTest,
				accept:               MessageStatusDecline,
				reason:               "",
			},
			expectResponse: chMsgPkt{
				MessageID: MsgNewChannelResponse,
				Message: jsonMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               MessageStatusDecline,
					Reason:               "some-random-reason",
				},
			},
			expectMatchInMock: false,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "write-error",
			args: args{
				msgProtocolVersion:   "",
				contractStoreVersion: []byte{},
				accept:               MessageStatusAccept,
				reason:               "",
			},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteMock(t, adapter, tt.expectResponse, tt.expectMatchInMock, tt.responseError, wg)

			err := ch.NewChannelRespond(tt.args.msgProtocolVersion, tt.args.contractStoreVersion, tt.args.accept, tt.args.reason)
			if (err != nil) != tt.wantErr {
				t.Errorf("channel.NewChannelRespond() error = %v, wantErr %v", err, tt.wantErr)
			}
			wg.Wait()
		})
	}
}

func Test_channel_ContractAddrRequest(t *testing.T) {
	type args struct {
		addr types.Address
		id   contract.Handler
	}
	tests := []struct {
		name              string
		args              args
		expectRequest     chMsgPkt
		expectMatchInMock bool
		mockResponse      jsonMsgPacket
		responseError     error
		wantErr           bool
		wantStatus        MessageStatus
	}{
		{
			name: "accept",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},
			expectRequest: chMsgPkt{
				MessageID: MsgContractAddrRequest,
				Message: jsonMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgContractAddrResponse,
					Message: jsonMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.LibSignatures(),
						Status:       MessageStatusAccept,
					},
				}},
			wantErr:    false,
			wantStatus: MessageStatusAccept,
		},
		{
			name: "decline",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},
			expectRequest: chMsgPkt{
				MessageID: MsgContractAddrRequest,
				Message: jsonMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgContractAddrResponse,
					Message: jsonMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.LibSignatures(),
						Status:       MessageStatusDecline,
					},
				}},
			wantErr:    false,
			wantStatus: MessageStatusDecline,
		},
		{
			name: "contract-type-modified-by-peer",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},
			expectRequest: chMsgPkt{
				MessageID: MsgContractAddrRequest,
				Message: jsonMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgContractAddrResponse,
					Message: jsonMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.MSContract(),
						Status:       MessageStatusDecline,
					},
				}},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			args: args{
				addr: types.Address{},
				id:   contract.Store.LibSignatures(),
			},
			expectRequest: chMsgPkt{
				MessageID: MsgContractAddrRequest,
				Message: jsonMsgContractAddr{
					Addr:         types.Address{},
					ContractType: contract.Store.LibSignatures(),
					Status:       "require",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-od"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			args: args{
				addr: types.Address{},
				id:   contract.Store.LibSignatures(),
			},
			expectRequest: chMsgPkt{
				MessageID: MsgContractAddrRequest,
				Message: jsonMsgContractAddr{
					Addr:         types.Address{},
					ContractType: contract.Store.LibSignatures(),
					Status:       "require",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgContractAddrResponse,
					Message:   nil,
				}},
			wantErr: true,
		},
		{
			name:          "write-error",
			args:          args{},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
		{
			name: "read-error",
			args: args{},
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteReadMock(t, adapter, tt.expectRequest, tt.mockResponse, tt.responseError, tt.expectMatchInMock, wg)

			gotStatus, err := ch.ContractAddrRequest(tt.args.addr, tt.args.id)
			if err != nil {
				t.Logf("channel.SendContractAddr() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("channel.SendContractAddr() = %v, want %v", gotStatus, tt.wantStatus)
			}
			wg.Wait()
		})
	}
}

func Test_channel_ContractAddrRead(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse jsonMsgPacket
		wantErr      bool
		wantAddr     types.Address
		wantID       contract.Handler
	}{
		{
			name: "valid",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgContractAddrRequest,
					Message: jsonMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.LibSignatures(),
						Status:       MessageStatusRequire,
					},
				}},
			wantErr:  false,
			wantAddr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
			wantID:   contract.Store.LibSignatures(),
		},
		{
			name: "invalid-status-accept",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgContractAddrRequest,
					Message: jsonMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.LibSignatures(),
						Status:       MessageStatusAccept,
					}},
			},
			wantErr:  true,
			wantAddr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
			wantID:   contract.Store.LibSignatures(),
		},
		{
			name: "invalid-status-decline",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgContractAddrRequest,
					Message: jsonMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.LibSignatures(),
						Status:       MessageStatusDecline,
					}},
			},
			wantErr:  true,
			wantAddr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
			wantID:   contract.Store.LibSignatures(),
		},
		{
			name: "read-error",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgContractAddrRequest,
					Message:   nil,
				}},
			wantErr: true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChReadMock(adapter, tt.mockResponse, wg)

			gotAddr, gotID, err := ch.ContractAddrRead()
			if err != nil {
				t.Logf("channel.ReadContractAddr() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if !reflect.DeepEqual(gotAddr, tt.wantAddr) {
				t.Errorf("channel.ReadContractAddr() gotAddr = %v, want %v", gotAddr, tt.wantAddr)
			}
			if !reflect.DeepEqual(gotID, tt.wantID) {
				t.Errorf("channel.ReadContractAddr() gotId = %v, want %v", gotID, tt.wantID)
			}

			wg.Wait()
		})
	}
}
func Test_channel_ContractAddrRespond(t *testing.T) {
	type args struct {
		addr   types.Address
		id     contract.Handler
		status MessageStatus
	}
	tests := []struct {
		name              string
		args              args
		expectResponse    chMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "valid-accept",
			args: args{
				addr:   types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:     contract.Store.LibSignatures(),
				status: MessageStatusAccept,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgContractAddrResponse,
				Message: jsonMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       MessageStatusAccept,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "valid-decline",
			args: args{
				addr:   types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:     contract.Store.LibSignatures(),
				status: MessageStatusDecline,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgContractAddrResponse,
				Message: jsonMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       MessageStatusDecline,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "invalid-status",
			args: args{
				addr:   types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:     contract.Store.LibSignatures(),
				status: MessageStatus("some-invalid-status"),
			},
			responseError: nil,
			wantErr:       true,
		},
		{
			name: "write-error",
			args: args{
				addr:   types.Address{},
				id:     contract.Store.LibSignatures(),
				status: MessageStatusAccept,
			},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteMock(t, adapter, tt.expectResponse, tt.expectMatchInMock, tt.responseError, wg)

			err := ch.ContractAddrRespond(tt.args.addr, tt.args.id, tt.args.status)
			if err != nil {
				t.Logf("channel.RespondToContractAddr() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			wg.Wait()
		})
	}
}

func Test_channel_NewMSCBaseStateRequest(t *testing.T) {
	type args struct {
		newState MSCBaseStateSigned
	}
	tests := []struct {
		name              string
		args              args
		expectRequest     chMsgPkt
		expectMatchInMock bool
		mockResponse      jsonMsgPacket
		responseError     error
		wantErr           bool
		wantResponseState MSCBaseStateSigned
		wantStatus        MessageStatus
	}{
		{
			name: "valid-accept",
			args: args{
				newState: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte(""),
				},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgMSCBaseStateRequest,
				Message: jsonMsgMSCBaseState{
					SignedStateVal: MSCBaseStateSigned{
						MSContractBaseState: MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgMSCBaseStateResponse,
					Message: jsonMsgMSCBaseState{
						SignedStateVal: MSCBaseStateSigned{
							MSContractBaseState: MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(10),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte("receiver-signature"),
						},
						Status: MessageStatusAccept,
					},
				}},
			wantErr: false,
			wantResponseState: MSCBaseStateSigned{
				MSContractBaseState: MSCBaseState{
					VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					Sid:             big.NewInt(10),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(10),
					Version:         big.NewInt(10),
				},
				SignSender:   []byte("sender-signature"),
				SignReceiver: []byte("receiver-signature"),
			},
			wantStatus: MessageStatusAccept,
		},
		{
			name: "valid-decline",
			args: args{
				newState: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte(""),
				},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgMSCBaseStateRequest,
				Message: jsonMsgMSCBaseState{
					SignedStateVal: MSCBaseStateSigned{
						MSContractBaseState: MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgMSCBaseStateResponse,
					Message: jsonMsgMSCBaseState{
						SignedStateVal: MSCBaseStateSigned{
							MSContractBaseState: MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(10),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: MessageStatusDecline,
					},
				}},
			wantErr: false,
			wantResponseState: MSCBaseStateSigned{
				MSContractBaseState: MSCBaseState{
					VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					Sid:             big.NewInt(10),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(10),
					Version:         big.NewInt(10),
				},
				SignSender:   []byte("sender-signature"),
				SignReceiver: []byte(""),
			},
			wantStatus: MessageStatusDecline,
		},
		{
			name: "msc-base-state-modified-by-peer",
			args: args{
				newState: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte(""),
				},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgMSCBaseStateRequest,
				Message: jsonMsgMSCBaseState{
					SignedStateVal: MSCBaseStateSigned{
						MSContractBaseState: MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgMSCBaseStateResponse,
					Message: jsonMsgMSCBaseState{
						SignedStateVal: MSCBaseStateSigned{
							MSContractBaseState: MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(1000),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte("receiver-signature"),
						},
						Status: MessageStatusAccept,
					},
				}},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			args: args{
				newState: MSCBaseStateSigned{},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgMSCBaseStateRequest,
				Message: jsonMsgMSCBaseState{
					SignedStateVal: MSCBaseStateSigned{},
					Status:         MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			args: args{
				newState: MSCBaseStateSigned{},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgMSCBaseStateRequest,
				Message: jsonMsgMSCBaseState{
					SignedStateVal: MSCBaseStateSigned{},
					Status:         MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgMSCBaseStateResponse,
					Message:   nil,
				}},
			wantErr: true,
		},
		{
			name:          "write-error",
			args:          args{},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
		{
			name: "read-error",
			args: args{},
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteReadMock(t, adapter, tt.expectRequest, tt.mockResponse, tt.responseError, tt.expectMatchInMock, wg)

			gotResponseState, gotStatus, err := ch.NewMSCBaseStateRequest(tt.args.newState)
			if err != nil {
				t.Logf("channel.NewMSCBaseStateRequest() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if !(gotResponseState.Equal(tt.wantResponseState)) {
				t.Errorf("channel.NewMSCBaseStateRequest() gotResponseState = %v, want %v", gotResponseState, tt.wantResponseState)
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("channel.NewMSCBaseStateRequest() gotStatus = %v, want %v", gotStatus, tt.wantStatus)
			}
			wg.Wait()
		})
	}
}

func Test_channel_NewMSCBaseStateRead(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse jsonMsgPacket
		wantErr      bool
		wantState    MSCBaseStateSigned
	}{
		{
			name: "valid",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgMSCBaseStateRequest,
					Message: jsonMsgMSCBaseState{
						SignedStateVal: MSCBaseStateSigned{
							MSContractBaseState: MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(10),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: MessageStatusRequire,
					},
				}},
			wantErr: false,
			wantState: MSCBaseStateSigned{
				MSContractBaseState: MSCBaseState{
					VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					Sid:             big.NewInt(10),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(10),
					Version:         big.NewInt(10),
				},
				SignSender:   []byte("sender-signature"),
				SignReceiver: []byte(""),
			},
		},
		{
			name: "read-error",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgMSCBaseStateRequest,
					Message:   nil,
				}},
			wantErr: true,
		},
		{
			name: "invalid-status-accept",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgMSCBaseStateRequest,
					Message: jsonMsgMSCBaseState{
						SignedStateVal: MSCBaseStateSigned{
							MSContractBaseState: MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(10),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: MessageStatusAccept,
					},
				}},
			wantErr: true,
		},
		{
			name: "invalid-status-decline",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgMSCBaseStateRequest,
					Message: jsonMsgMSCBaseState{
						SignedStateVal: MSCBaseStateSigned{
							MSContractBaseState: MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(10),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: MessageStatusDecline,
					},
				}},
			wantErr: true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChReadMock(adapter, tt.mockResponse, wg)

			gotState, err := ch.NewMSCBaseStateRead()
			if err != nil {
				t.Logf("channel.NewMSCBaseStateRead() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if !gotState.Equal(tt.wantState) {
				t.Errorf("channel.NewMSCBaseStateRead() = %v, want %v", gotState, tt.wantState)
			}
			wg.Wait()
		})
	}
}

func Test_channel_NewMSCBaseStateRespond(t *testing.T) {
	type args struct {
		state  MSCBaseStateSigned
		status MessageStatus
	}
	tests := []struct {
		name              string
		args              args
		expectResponse    chMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "valid-accept",
			args: args{
				state: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: MessageStatusAccept,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgMSCBaseStateResponse,
				Message: jsonMsgMSCBaseState{
					SignedStateVal: MSCBaseStateSigned{
						MSContractBaseState: MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: MessageStatusAccept,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "valid-decline",
			args: args{
				state: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: MessageStatusDecline,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgMSCBaseStateResponse,
				Message: jsonMsgMSCBaseState{
					SignedStateVal: MSCBaseStateSigned{
						MSContractBaseState: MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: MessageStatusDecline,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "invalid-status-require",
			args: args{
				state: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: MessageStatusRequire,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgMSCBaseStateResponse,
				Message: jsonMsgMSCBaseState{
					SignedStateVal: MSCBaseStateSigned{
						MSContractBaseState: MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           true,
		},
		{
			name: "write-error",
			args: args{
				state:  MSCBaseStateSigned{},
				status: MessageStatusRequire,
			},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteMock(t, adapter, tt.expectResponse, tt.expectMatchInMock, tt.responseError, wg)

			err := ch.NewMSCBaseStateRespond(tt.args.state, tt.args.status)
			if err != nil {
				t.Logf("channel.NewMSCBaseStateRespond() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			wg.Wait()
		})
	}
}

func Test_channel_NewVPCStateRequest(t *testing.T) {
	type args struct {
		newStateSigned VPCStateSigned
	}
	tests := []struct {
		name              string
		args              args
		expectRequest     chMsgPkt
		expectMatchInMock bool
		mockResponse      jsonMsgPacket
		responseError     error
		wantErr           bool
		wantResponseState VPCStateSigned
		wantStatus        MessageStatus
	}{
		{
			name: "valid-accept",
			args: args{
				newStateSigned: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte(""),
				},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgVPCStateRequest,
				Message: jsonMsgVPCState{
					SignedStateVal: VPCStateSigned{
						VPCState: VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgVPCStateResponse,
					Message: jsonMsgVPCState{
						SignedStateVal: VPCStateSigned{
							VPCState: VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte("receiver-signature"),
						},
						Status: MessageStatusAccept,
					},
				}},
			wantErr: false,
			wantResponseState: VPCStateSigned{
				VPCState: VPCState{
					ID:              []byte("some-valid-id"),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(10),
					Version:         big.NewInt(10),
				},
				SignSender:   []byte("sender-signature"),
				SignReceiver: []byte("receiver-signature"),
			},
			wantStatus: MessageStatusAccept,
		},
		{
			name: "valid-decline",
			args: args{
				newStateSigned: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte(""),
				},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgVPCStateRequest,
				Message: jsonMsgVPCState{
					SignedStateVal: VPCStateSigned{
						VPCState: VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgVPCStateResponse,
					Message: jsonMsgVPCState{
						SignedStateVal: VPCStateSigned{
							VPCState: VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: MessageStatusDecline,
					},
				}},
			wantErr: false,
			wantResponseState: VPCStateSigned{
				VPCState: VPCState{
					ID:              []byte("some-valid-id"),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(10),
					Version:         big.NewInt(10),
				},
				SignSender:   []byte("sender-signature"),
				SignReceiver: []byte(""),
			},
			wantStatus: MessageStatusDecline,
		},
		{
			name: "invalid-message-id",
			args: args{
				newStateSigned: VPCStateSigned{
					VPCState: VPCState{},
				},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgVPCStateRequest,
				Message: jsonMsgVPCState{
					SignedStateVal: VPCStateSigned{
						VPCState: VPCState{},
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			args: args{
				newStateSigned: VPCStateSigned{
					VPCState: VPCState{},
				},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgVPCStateRequest,
				Message: jsonMsgVPCState{
					SignedStateVal: VPCStateSigned{
						VPCState: VPCState{},
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgVPCStateResponse,
					Message:   nil,
				},
			},
			wantErr: true,
		},
		{
			name: "vpc-state-modified-by-peer",
			args: args{
				newStateSigned: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte(""),
				},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgVPCStateRequest,
				Message: jsonMsgVPCState{
					SignedStateVal: VPCStateSigned{
						VPCState: VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgVPCStateResponse,
					Message: jsonMsgVPCState{
						SignedStateVal: VPCStateSigned{
							VPCState: VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(1000),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte("receiver-signature"),
						},
						Status: MessageStatusAccept,
					},
				}},
			wantErr: true,
		},
		{
			name:          "write-error",
			args:          args{},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
		{
			name: "read-error",
			args: args{},
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteReadMock(t, adapter, tt.expectRequest, tt.mockResponse, tt.responseError, tt.expectMatchInMock, wg)

			gotResponseState, gotStatus, err := ch.NewVPCStateRequest(tt.args.newStateSigned)
			if err != nil {
				t.Logf("channel.NewVPCStateRequest() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if !(gotResponseState.Equal(tt.wantResponseState)) {
				t.Errorf("channel.NewVPCStateRequest() gotResponseState = %v, want %v", gotResponseState, tt.wantResponseState)
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("channel.NewVPCStateRequest() gotStatus = %v, want %v", gotStatus, tt.wantStatus)
			}
			wg.Wait()
		})
	}
}

func Test_channel_NewVPCStateRead(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse jsonMsgPacket
		wantErr      bool
		wantState    VPCStateSigned
	}{
		{
			name: "valid",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgVPCStateRequest,
					Message: jsonMsgVPCState{
						SignedStateVal: VPCStateSigned{
							VPCState: VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: MessageStatusRequire,
					},
				}},
			wantErr: false,
			wantState: VPCStateSigned{
				VPCState: VPCState{
					ID:              []byte("some-valid-id"),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(10),
					Version:         big.NewInt(10),
				},
				SignSender:   []byte("sender-signature"),
				SignReceiver: []byte(""),
			},
		},
		{
			name: "read-error",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgVPCStateRequest,
					Message:   nil,
				}},
			wantErr: true,
		},
		{
			name: "invalid-status-accept",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgVPCStateRequest,
					Message: jsonMsgVPCState{
						SignedStateVal: VPCStateSigned{
							VPCState: VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: MessageStatusAccept,
					},
				}},
			wantErr: true,
		},
		{
			name: "invalid-status-decline",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgVPCStateRequest,
					Message: jsonMsgVPCState{
						SignedStateVal: VPCStateSigned{
							VPCState: VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: MessageStatusDecline,
					},
				}},
			wantErr: true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChReadMock(adapter, tt.mockResponse, wg)

			gotState, err := ch.NewVPCStateRead()
			if err != nil {
				t.Logf("channel.NewVPCStateRead() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if !gotState.Equal(tt.wantState) {
				t.Errorf("channel.NewVPCStateRead() = %v, want %v", gotState, tt.wantState)
			}
			wg.Wait()
		})
	}
}

func Test_channel_NewVPCStateRespond(t *testing.T) {
	type args struct {
		state  VPCStateSigned
		status MessageStatus
	}
	tests := []struct {
		name              string
		args              args
		expectResponse    chMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "valid-accept",
			args: args{
				state: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: MessageStatusAccept,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgVPCStateResponse,
				Message: jsonMsgVPCState{
					SignedStateVal: VPCStateSigned{
						VPCState: VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: MessageStatusAccept,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "valid-decline",
			args: args{
				state: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: MessageStatusDecline,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgVPCStateResponse,
				Message: jsonMsgVPCState{
					SignedStateVal: VPCStateSigned{
						VPCState: VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: MessageStatusDecline,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "invalid-status-require",
			args: args{
				state: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: MessageStatusRequire,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgVPCStateResponse,
				Message: jsonMsgVPCState{
					SignedStateVal: VPCStateSigned{
						VPCState: VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           true,
		},
		{
			name: "write-error",
			args: args{
				state:  VPCStateSigned{},
				status: MessageStatusRequire,
			},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteMock(t, adapter, tt.expectResponse, tt.expectMatchInMock, tt.responseError, wg)

			err := ch.NewVPCStateRespond(tt.args.state, tt.args.status)
			if err != nil {
				t.Logf("channel.NewVPCStateRespond() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			wg.Wait()
		})
	}
}

func Test_channel_SessionIdRequest(t *testing.T) {
	type args struct {
		sid SessionID
	}
	tests := []struct {
		name              string
		args              args
		expectRequest     chMsgPkt
		expectMatchInMock bool
		mockResponse      jsonMsgPacket
		responseError     error
		wantErr           bool
		wantStatus        MessageStatus
		wantSessionID     SessionID
	}{
		{
			name: "valid-accept",
			args: args{
				sid: SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgSessionIDRequest,
				Message: jsonMsgSessionID{
					Sid: SessionID{
						SidSenderPart: []byte("sid-sender-part"),
						AddrSender:    aliceID.OnChainID,
						NonceSender:   big.NewInt(10).Bytes(),
						Locked:        false,
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgSessionIDResponse,
					Message: jsonMsgSessionID{
						Sid: SessionID{
							SidSenderPart:   []byte("sid-sender-part"),
							AddrSender:      aliceID.OnChainID,
							NonceSender:     big.NewInt(10).Bytes(),
							SidReceiverPart: []byte("sid-receiver-part"),
							AddrReceiver:    bobID.OnChainID,
							NonceReceiver:   big.NewInt(20).Bytes(),
							Locked:          false,
						},
						Status: MessageStatusAccept,
					},
				}},
			wantErr:    false,
			wantStatus: MessageStatusAccept,
			wantSessionID: SessionID{
				SidSenderPart:   []byte("sid-sender-part"),
				AddrSender:      aliceID.OnChainID,
				NonceSender:     big.NewInt(10).Bytes(),
				SidReceiverPart: []byte("sid-receiver-part"),
				AddrReceiver:    bobID.OnChainID,
				NonceReceiver:   big.NewInt(20).Bytes(),
				Locked:          false,
			},
		},
		{
			name: "valid-decline",
			args: args{
				sid: SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgSessionIDRequest,
				Message: jsonMsgSessionID{
					Sid: SessionID{
						SidSenderPart: []byte("sid-sender-part"),
						AddrSender:    aliceID.OnChainID,
						NonceSender:   big.NewInt(10).Bytes(),
						Locked:        false,
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgSessionIDResponse,
					Message: jsonMsgSessionID{
						Sid: SessionID{
							SidSenderPart: []byte("sid-sender-part"),
							AddrSender:    aliceID.OnChainID,
							NonceSender:   big.NewInt(10).Bytes(),
							Locked:        false,
						},
						Status: MessageStatusDecline,
					},
				}},
			wantErr:    false,
			wantStatus: MessageStatusDecline,
			wantSessionID: SessionID{
				SidSenderPart: []byte("sid-sender-part"),
				AddrSender:    aliceID.OnChainID,
				NonceSender:   big.NewInt(10).Bytes(),
				Locked:        false,
			},
		},
		{
			name: "write-error",
			args: args{
				sid: SessionID{},
			},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
		{
			name: "read-error",
			args: args{
				sid: SessionID{},
			},
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			args: args{
				sid: SessionID{},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgSessionIDRequest,
				Message: jsonMsgSessionID{
					Sid:    SessionID{},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			args: args{
				sid: SessionID{},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgSessionIDRequest,
				Message: jsonMsgSessionID{
					Sid:    SessionID{},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgSessionIDResponse,
					Message:   nil,
				}},
			wantErr: true,
		},
		{
			name: "peer-modifies-sender-component",
			args: args{
				sid: SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},
			expectRequest: chMsgPkt{
				MessageID: MsgSessionIDRequest,
				Message: jsonMsgSessionID{
					Sid: SessionID{
						SidSenderPart: []byte("sid-sender-part"),
						AddrSender:    aliceID.OnChainID,
						NonceSender:   big.NewInt(10).Bytes(),
						Locked:        false,
					},
					Status: MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgSessionIDResponse,
					Message: jsonMsgSessionID{
						Sid: SessionID{
							SidSenderPart:   []byte("sid-modified-sender-part"),
							AddrSender:      aliceID.OnChainID,
							NonceSender:     big.NewInt(10).Bytes(),
							SidReceiverPart: []byte("sid-receiver-part"),
							AddrReceiver:    bobID.OnChainID,
							NonceReceiver:   big.NewInt(20).Bytes(),
							Locked:          false,
						},
						Status: MessageStatusAccept,
					},
				}},
			wantErr: true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteReadMock(t, adapter, tt.expectRequest, tt.mockResponse, tt.responseError, tt.expectMatchInMock, wg)

			gotSessionID, gotStatus, err := ch.SessionIDRequest(tt.args.sid)
			if err != nil {
				t.Logf("channel.SessionIdRequest() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("channel.SessionIdRequest() = %v, want %v", gotStatus, tt.wantStatus)
			}
			if !gotSessionID.Equal(tt.wantSessionID) {
				t.Errorf("channel.SessionIdRequest() = %v, want %v", gotSessionID, tt.wantSessionID)
			}
			wg.Wait()
		})
	}
}

func Test_channel_SessionIdRead(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse jsonMsgPacket
		wantErr      bool
		wantSid      SessionID
	}{
		{
			name: "valid-1",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgSessionIDRequest,
					Message: jsonMsgSessionID{
						Sid: SessionID{
							SidSenderPart: []byte("sid-sender-part"),
							AddrSender:    aliceID.OnChainID,
							NonceSender:   big.NewInt(10).Bytes(),
							Locked:        false,
						},
						Status: MessageStatusRequire,
					},
				}},
			wantErr: false,
			wantSid: SessionID{
				SidSenderPart: []byte("sid-sender-part"),
				AddrSender:    aliceID.OnChainID,
				NonceSender:   big.NewInt(10).Bytes(),
				Locked:        false,
			},
		},
		{
			name: "read-error",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{},
				err:     fmt.Errorf("read-error"),
			},
			wantErr: true,
		},
		{
			name: "messageId-error",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MessageID("invalid-message-id"),
				},
			},
			wantErr: true,
		},
		{
			name: "message-error",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgSessionIDRequest,
					Message:   nil,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-status-accept",
			mockResponse: jsonMsgPacket{
				message: chMsgPkt{
					MessageID: MsgSessionIDRequest,
					Message: jsonMsgSessionID{
						Sid: SessionID{
							SidSenderPart: []byte("sid-sender-part"),
							AddrSender:    aliceID.OnChainID,
							NonceSender:   big.NewInt(10).Bytes(),
							Locked:        false,
						},
						Status: MessageStatusAccept,
					},
				}},
			wantErr: true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChReadMock(adapter, tt.mockResponse, wg)

			gotSid, err := ch.SessionIDRead()
			if err != nil {
				t.Logf("channel.SessionIdRead() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if !reflect.DeepEqual(gotSid, tt.wantSid) {
				t.Errorf("channel.SessionIdRead() gotSid = %v, want %v", gotSid, tt.wantSid)
			}

			wg.Wait()
		})
	}
}

func Test_channel_SessionIdRespond(t *testing.T) {
	type args struct {
		sid    SessionID
		status MessageStatus
	}
	tests := []struct {
		name              string
		args              args
		expectResponse    chMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "valid-accept",
			args: args{
				sid: SessionID{
					SidSenderPart:   []byte("sid-sender-part"),
					AddrSender:      aliceID.OnChainID,
					NonceSender:     big.NewInt(10).Bytes(),
					SidReceiverPart: []byte("sid-receiver-part"),
					AddrReceiver:    bobID.OnChainID,
					NonceReceiver:   big.NewInt(20).Bytes(),
					Locked:          false,
				},
				status: MessageStatusAccept,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgSessionIDResponse,
				Message: jsonMsgSessionID{
					Sid: SessionID{
						SidSenderPart:   []byte("sid-sender-part"),
						AddrSender:      aliceID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false,
					},
					Status: MessageStatusAccept,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "valid-decline",
			args: args{
				sid: SessionID{
					SidSenderPart:   []byte("sid-sender-part"),
					AddrSender:      aliceID.OnChainID,
					NonceSender:     big.NewInt(10).Bytes(),
					SidReceiverPart: []byte("sid-receiver-part"),
					AddrReceiver:    bobID.OnChainID,
					NonceReceiver:   big.NewInt(20).Bytes(),
					Locked:          false,
				},
				status: MessageStatusDecline,
			},
			expectResponse: chMsgPkt{
				MessageID: MsgSessionIDResponse,
				Message: jsonMsgSessionID{
					Sid: SessionID{
						SidSenderPart:   []byte("sid-sender-part"),
						AddrSender:      aliceID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false,
					},
					Status: MessageStatusDecline,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "invalid-status",
			args: args{
				sid:    SessionID{},
				status: MessageStatus("invalid-status"),
			},
			expectResponse: chMsgPkt{
				MessageID: MsgSessionIDResponse,
				Message: jsonMsgSessionID{
					Sid:    SessionID{},
					Status: MessageStatus("invalid-status"),
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           true,
		},
		{
			name: "write-error",
			args: args{
				sid:    SessionID{},
				status: MessageStatusDecline,
			},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ch, adapter := setupMockChannel()
			adapter.connected = true

			wg.Add(1)
			go ChWriteMock(t, adapter, tt.expectResponse, tt.expectMatchInMock, tt.responseError, wg)

			err := ch.SessionIDRespond(tt.args.sid, tt.args.status)
			if err != nil {
				t.Logf("channel.SessionIdResponse() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			wg.Wait()
		})
	}
}
