// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/dst-go
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

package primitives

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
)

var testTime time.Time

func init() {
	testTime, _ = time.Parse(
		time.RFC3339, "2012-11-01T22:08:41+00:00")
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
		wantMsgPkt ChMsgPkt
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
			wantMsgPkt: ChMsgPkt{
				Version:   "1.0",
				MessageID: MsgIdentityRequest,
				Message: JSONMsgIdentity{
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
			wantMsgPkt: ChMsgPkt{
				Version:   "1.0",
				MessageID: MsgIdentityResponse,
				Message: JSONMsgIdentity{
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
			wantMsgPkt: ChMsgPkt{
				Version:   "1.0",
				MessageID: MsgNewChannelRequest,
				Message: JSONMsgNewChannel{
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
			wantMsgPkt: ChMsgPkt{
				Version:   "1.0",
				MessageID: MsgNewChannelResponse,
				Message: JSONMsgNewChannel{
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
			wantMsgPkt: ChMsgPkt{
				Version:   "1.0",
				MessageID: MsgSessionIDRequest,
				Message: JSONMsgSessionID{
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
			wantMsgPkt: ChMsgPkt{
				Version:   "1.0",
				MessageID: MsgSessionIDResponse,
				Message: JSONMsgSessionID{
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
			wantMsgPkt: ChMsgPkt{
				Version:   "1.0",
				MessageID: MsgContractAddrRequest,
				Message: JSONMsgContractAddr{
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
			wantMsgPkt: ChMsgPkt{
				Version:   "1.0",
				MessageID: MsgContractAddrResponse,
				Message: JSONMsgContractAddr{
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
			wantMsgPkt: ChMsgPkt{
				Version:   "1.0",
				MessageID: MsgMSCBaseStateRequest,
				Message: JSONMsgMSCBaseState{
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
			wantMsgPkt: ChMsgPkt{
				Version:   "1.0",
				MessageID: MsgVPCStateRequest,
				Message: JSONMsgVPCState{
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
			wantMsgPkt: ChMsgPkt{
				Version:   "1.0",
				MessageID: MsgVPCStateResponse,
				Message: JSONMsgVPCState{
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

			msgPkt := ChMsgPkt{}
			if err := msgPkt.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Fatalf("chMsgPkt.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(msgPkt, tt.wantMsgPkt) {
				t.Errorf("jsonMsg unmarshall() got %v, want %v", msgPkt, tt.wantMsgPkt)
			}

		})
	}
}

func Test_ContainsStatus(t *testing.T) {
	type args struct {
		list          []MessageStatus
		requiredValue MessageStatus
	}
	tests := []struct {
		name       string
		args       args
		wantStatus bool
	}{
		{
			name: "status_present",
			args: args{
				list:          []MessageStatus{MessageStatusAccept, MessageStatusDecline},
				requiredValue: MessageStatusAccept,
			},
			wantStatus: true,
		},
		{
			name: "status_not_present",
			args: args{
				list:          []MessageStatus{MessageStatusAccept, MessageStatusDecline},
				requiredValue: MessageStatusRequire,
			},
			wantStatus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotStatus := ContainsStatus(tt.args.list, tt.args.requiredValue)
			if gotStatus != tt.wantStatus {
				t.Errorf("ContainsStatus() got %t, want %t", gotStatus, tt.wantStatus)
			}
		})
	}

}
