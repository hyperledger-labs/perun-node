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

package channel

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/direct-state-transfer/dst-go/channel/primitives"
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

var MsgIDForTest primitives.MessageID = "junk-message-type-for-test"

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
func ChWriteMock(t *testing.T, ch *genericChannelAdapter, expectRequest primitives.ChMsgPkt, expectMatchInMock bool, respondWithErr error, wg *sync.WaitGroup) {

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
func ChWriteReadMock(t *testing.T, ch *genericChannelAdapter, expectRequestMsg primitives.ChMsgPkt, responseMsg jsonMsgPacket, respondWithErr error, expectRequestMsgMatch bool, wg *sync.WaitGroup) {

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

func compareMsg(msg1, msg2 primitives.ChMsgPkt) bool {

	if (reflect.TypeOf(msg1) == reflect.TypeOf(msg2)) &&
		(msg1.MessageID == msg2.MessageID) &&
		reflect.DeepEqual(msg1.Message, msg2.Message) {
		return true
	}
	return false
}
func Test_channel_IdentityRequest(t *testing.T) {
	type args struct {
		selfID identity.OffChainID
	}
	tests := []struct {
		name              string
		args              args
		expectRequest     primitives.ChMsgPkt
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
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message:   primitives.JSONMsgIdentity{ID: aliceID},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgIdentityResponse,
					Message: primitives.JSONMsgIdentity{
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
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message: primitives.JSONMsgIdentity{
					ID: aliceID,
				}},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			args: args{
				selfID: aliceID,
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message: primitives.JSONMsgIdentity{
					ID: aliceID,
				}},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgIdentityResponse,
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
				message: primitives.ChMsgPkt{},
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
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgIdentityRequest,
					Message: primitives.JSONMsgIdentity{
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
				message: primitives.ChMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
		{
			name: "id-mismatch",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgIdentityRequest,
					Message: primitives.JSONMsgIdentity{
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
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-id"),
				}},
			wantErr:         true,
			wantPeerID:      aliceID,
			wantPeerIDMatch: false,
		},
		{
			name: "invalid-message",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgIdentityRequest,
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
		expectResponse    primitives.ChMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "valid-1",
			args: args{
				selfID: aliceID,
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityResponse,
				Message: primitives.JSONMsgIdentity{
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
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityResponse,
				Message: primitives.JSONMsgIdentity{
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
			expectResponse: primitives.ChMsgPkt{
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
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityResponse,
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
		expectRequest     primitives.ChMsgPkt
		expectMatchInMock bool
		mockResponse      jsonMsgPacket
		responseError     error
		wantErr           bool
		wantAccept        primitives.MessageStatus
		wantReason        string
	}{
		{
			name: "accept",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0"),
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgNewChannelResponse,
					Message: primitives.JSONMsgNewChannel{
						MsgProtocolVersion:   "1.0",
						ContractStoreVersion: []byte("v1.0"),
						Status:               primitives.MessageStatusAccept,
						Reason:               "",
					},
				}},
			wantErr:    false,
			wantAccept: primitives.MessageStatusAccept,
			wantReason: "",
		},
		{
			name: "decline",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0"),
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgNewChannelResponse,
					Message: primitives.JSONMsgNewChannel{
						MsgProtocolVersion:   "1.0",
						ContractStoreVersion: []byte("v1.0"),
						Status:               primitives.MessageStatusDecline,
						Reason:               "wrong-version",
					},
				}},
			wantErr:    false,
			wantAccept: primitives.MessageStatusDecline,
			wantReason: "wrong-version",
		},
		{
			name: "msgProtocolVersion-modified-by-peer",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0"),
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgNewChannelResponse,
					Message: primitives.JSONMsgNewChannel{
						MsgProtocolVersion:   "2.0",
						ContractStoreVersion: []byte("v1.0"),
						Status:               primitives.MessageStatusAccept,
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
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgNewChannelResponse,
					Message: primitives.JSONMsgNewChannel{
						MsgProtocolVersion:   "1.0",
						ContractStoreVersion: []byte("v2.0"),
						Status:               primitives.MessageStatusAccept,
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
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					ContractStoreVersion: []byte("v1.0"),
					MsgProtocolVersion:   "1.0",
					Status:               "require",
					Reason:               "",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0"),
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					ContractStoreVersion: []byte("v1.0"),
					MsgProtocolVersion:   "1.0",
					Status:               "require",
					Reason:               "",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgNewChannelResponse,
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
				message: primitives.ChMsgPkt{},
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
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgNewChannelRequest,
					Message: primitives.JSONMsgNewChannel{
						MsgProtocolVersion:   "1.0",
						ContractStoreVersion: contractStoreVersionForTest,
						Status:               primitives.MessageStatusRequire,
					},
				}},
			wantErr:                  false,
			wantMsgProtocolVersion:   "1.0",
			wantContractStoreVersion: contractStoreVersionForTest,
		},
		{
			name: "read-error",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{},
				err:     fmt.Errorf("read-error"),
			},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-id"),
				}},
			wantErr:                  true,
			wantContractStoreVersion: contractStoreVersionForTest,
		},
		{
			name: "invalid-message",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgNewChannelRequest,
					Message:   nil,
				}},
			wantErr:                  true,
			wantContractStoreVersion: contractStoreVersionForTest,
		},
		{
			name: "invalid-status",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgNewChannelRequest,
					Message: primitives.JSONMsgNewChannel{
						MsgProtocolVersion:   "1.0",
						ContractStoreVersion: contractStoreVersionForTest,
						Status:               primitives.MessageStatus("invalid-message-status"),
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
		accept               primitives.MessageStatus
		reason               string
	}
	tests := []struct {
		name              string
		args              args
		expectResponse    primitives.ChMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "accept",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: contractStoreVersionForTest,
				accept:               primitives.MessageStatusAccept,
				reason:               "",
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelResponse,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               primitives.MessageStatusAccept,
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
				accept:               primitives.MessageStatusDecline,
				reason:               "",
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelResponse,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               primitives.MessageStatusDecline,
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
				accept:               primitives.MessageStatusDecline,
				reason:               "",
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelResponse,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               primitives.MessageStatusDecline,
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
				accept:               primitives.MessageStatusAccept,
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
		expectRequest     primitives.ChMsgPkt
		expectMatchInMock bool
		mockResponse      jsonMsgPacket
		responseError     error
		wantErr           bool
		wantStatus        primitives.MessageStatus
	}{
		{
			name: "accept",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgContractAddrResponse,
					Message: primitives.JSONMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.LibSignatures(),
						Status:       primitives.MessageStatusAccept,
					},
				}},
			wantErr:    false,
			wantStatus: primitives.MessageStatusAccept,
		},
		{
			name: "decline",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgContractAddrResponse,
					Message: primitives.JSONMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.LibSignatures(),
						Status:       primitives.MessageStatusDecline,
					},
				}},
			wantErr:    false,
			wantStatus: primitives.MessageStatusDecline,
		},
		{
			name: "contract-type-modified-by-peer",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgContractAddrResponse,
					Message: primitives.JSONMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.MSContract(),
						Status:       primitives.MessageStatusDecline,
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
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.Address{},
					ContractType: contract.Store.LibSignatures(),
					Status:       "require",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-od"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			args: args{
				addr: types.Address{},
				id:   contract.Store.LibSignatures(),
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.Address{},
					ContractType: contract.Store.LibSignatures(),
					Status:       "require",
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgContractAddrResponse,
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
				message: primitives.ChMsgPkt{},
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
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgContractAddrRequest,
					Message: primitives.JSONMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.LibSignatures(),
						Status:       primitives.MessageStatusRequire,
					},
				}},
			wantErr:  false,
			wantAddr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
			wantID:   contract.Store.LibSignatures(),
		},
		{
			name: "invalid-status-accept",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgContractAddrRequest,
					Message: primitives.JSONMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.LibSignatures(),
						Status:       primitives.MessageStatusAccept,
					}},
			},
			wantErr:  true,
			wantAddr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
			wantID:   contract.Store.LibSignatures(),
		},
		{
			name: "invalid-status-decline",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgContractAddrRequest,
					Message: primitives.JSONMsgContractAddr{
						Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						ContractType: contract.Store.LibSignatures(),
						Status:       primitives.MessageStatusDecline,
					}},
			},
			wantErr:  true,
			wantAddr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
			wantID:   contract.Store.LibSignatures(),
		},
		{
			name: "read-error",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgContractAddrRequest,
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
		status primitives.MessageStatus
	}
	tests := []struct {
		name              string
		args              args
		expectResponse    primitives.ChMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "valid-accept",
			args: args{
				addr:   types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:     contract.Store.LibSignatures(),
				status: primitives.MessageStatusAccept,
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrResponse,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusAccept,
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
				status: primitives.MessageStatusDecline,
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrResponse,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusDecline,
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
				status: primitives.MessageStatus("some-invalid-status"),
			},
			responseError: nil,
			wantErr:       true,
		},
		{
			name: "write-error",
			args: args{
				addr:   types.Address{},
				id:     contract.Store.LibSignatures(),
				status: primitives.MessageStatusAccept,
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
		newState primitives.MSCBaseStateSigned
	}
	tests := []struct {
		name              string
		args              args
		expectRequest     primitives.ChMsgPkt
		expectMatchInMock bool
		mockResponse      jsonMsgPacket
		responseError     error
		wantErr           bool
		wantResponseState primitives.MSCBaseStateSigned
		wantStatus        primitives.MessageStatus
	}{
		{
			name: "valid-accept",
			args: args{
				newState: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
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
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateRequest,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgMSCBaseStateResponse,
					Message: primitives.JSONMsgMSCBaseState{
						SignedStateVal: primitives.MSCBaseStateSigned{
							MSContractBaseState: primitives.MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(10),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte("receiver-signature"),
						},
						Status: primitives.MessageStatusAccept,
					},
				}},
			wantErr: false,
			wantResponseState: primitives.MSCBaseStateSigned{
				MSContractBaseState: primitives.MSCBaseState{
					VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					Sid:             big.NewInt(10),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(10),
					Version:         big.NewInt(10),
				},
				SignSender:   []byte("sender-signature"),
				SignReceiver: []byte("receiver-signature"),
			},
			wantStatus: primitives.MessageStatusAccept,
		},
		{
			name: "valid-decline",
			args: args{
				newState: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
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
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateRequest,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgMSCBaseStateResponse,
					Message: primitives.JSONMsgMSCBaseState{
						SignedStateVal: primitives.MSCBaseStateSigned{
							MSContractBaseState: primitives.MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(10),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: primitives.MessageStatusDecline,
					},
				}},
			wantErr: false,
			wantResponseState: primitives.MSCBaseStateSigned{
				MSContractBaseState: primitives.MSCBaseState{
					VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					Sid:             big.NewInt(10),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(10),
					Version:         big.NewInt(10),
				},
				SignSender:   []byte("sender-signature"),
				SignReceiver: []byte(""),
			},
			wantStatus: primitives.MessageStatusDecline,
		},
		{
			name: "msc-base-state-modified-by-peer",
			args: args{
				newState: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
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
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateRequest,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgMSCBaseStateResponse,
					Message: primitives.JSONMsgMSCBaseState{
						SignedStateVal: primitives.MSCBaseStateSigned{
							MSContractBaseState: primitives.MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(1000),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte("receiver-signature"),
						},
						Status: primitives.MessageStatusAccept,
					},
				}},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			args: args{
				newState: primitives.MSCBaseStateSigned{},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateRequest,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{},
					Status:         primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			args: args{
				newState: primitives.MSCBaseStateSigned{},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateRequest,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{},
					Status:         primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgMSCBaseStateResponse,
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
				message: primitives.ChMsgPkt{},
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
		wantState    primitives.MSCBaseStateSigned
	}{
		{
			name: "valid",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgMSCBaseStateRequest,
					Message: primitives.JSONMsgMSCBaseState{
						SignedStateVal: primitives.MSCBaseStateSigned{
							MSContractBaseState: primitives.MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(10),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: primitives.MessageStatusRequire,
					},
				}},
			wantErr: false,
			wantState: primitives.MSCBaseStateSigned{
				MSContractBaseState: primitives.MSCBaseState{
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
				message: primitives.ChMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgMSCBaseStateRequest,
					Message:   nil,
				}},
			wantErr: true,
		},
		{
			name: "invalid-status-accept",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgMSCBaseStateRequest,
					Message: primitives.JSONMsgMSCBaseState{
						SignedStateVal: primitives.MSCBaseStateSigned{
							MSContractBaseState: primitives.MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(10),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: primitives.MessageStatusAccept,
					},
				}},
			wantErr: true,
		},
		{
			name: "invalid-status-decline",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgMSCBaseStateRequest,
					Message: primitives.JSONMsgMSCBaseState{
						SignedStateVal: primitives.MSCBaseStateSigned{
							MSContractBaseState: primitives.MSCBaseState{
								VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
								Sid:             big.NewInt(10),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: primitives.MessageStatusDecline,
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
		state  primitives.MSCBaseStateSigned
		status primitives.MessageStatus
	}
	tests := []struct {
		name              string
		args              args
		expectResponse    primitives.ChMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "valid-accept",
			args: args{
				state: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: primitives.MessageStatusAccept,
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateResponse,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: primitives.MessageStatusAccept,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "valid-decline",
			args: args{
				state: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: primitives.MessageStatusDecline,
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateResponse,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: primitives.MessageStatusDecline,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "invalid-status-require",
			args: args{
				state: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: primitives.MessageStatusRequire,
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateResponse,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           true,
		},
		{
			name: "write-error",
			args: args{
				state:  primitives.MSCBaseStateSigned{},
				status: primitives.MessageStatusRequire,
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
		newStateSigned primitives.VPCStateSigned
	}
	tests := []struct {
		name              string
		args              args
		expectRequest     primitives.ChMsgPkt
		expectMatchInMock bool
		mockResponse      jsonMsgPacket
		responseError     error
		wantErr           bool
		wantResponseState primitives.VPCStateSigned
		wantStatus        primitives.MessageStatus
	}{
		{
			name: "valid-accept",
			args: args{
				newStateSigned: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte(""),
				},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateRequest,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgVPCStateResponse,
					Message: primitives.JSONMsgVPCState{
						SignedStateVal: primitives.VPCStateSigned{
							VPCState: primitives.VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte("receiver-signature"),
						},
						Status: primitives.MessageStatusAccept,
					},
				}},
			wantErr: false,
			wantResponseState: primitives.VPCStateSigned{
				VPCState: primitives.VPCState{
					ID:              []byte("some-valid-id"),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(10),
					Version:         big.NewInt(10),
				},
				SignSender:   []byte("sender-signature"),
				SignReceiver: []byte("receiver-signature"),
			},
			wantStatus: primitives.MessageStatusAccept,
		},
		{
			name: "valid-decline",
			args: args{
				newStateSigned: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte(""),
				},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateRequest,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgVPCStateResponse,
					Message: primitives.JSONMsgVPCState{
						SignedStateVal: primitives.VPCStateSigned{
							VPCState: primitives.VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: primitives.MessageStatusDecline,
					},
				}},
			wantErr: false,
			wantResponseState: primitives.VPCStateSigned{
				VPCState: primitives.VPCState{
					ID:              []byte("some-valid-id"),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(10),
					Version:         big.NewInt(10),
				},
				SignSender:   []byte("sender-signature"),
				SignReceiver: []byte(""),
			},
			wantStatus: primitives.MessageStatusDecline,
		},
		{
			name: "invalid-message-id",
			args: args{
				newStateSigned: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{},
				},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateRequest,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{},
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			args: args{
				newStateSigned: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{},
				},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateRequest,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{},
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgVPCStateResponse,
					Message:   nil,
				},
			},
			wantErr: true,
		},
		{
			name: "vpc-state-modified-by-peer",
			args: args{
				newStateSigned: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte(""),
				},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateRequest,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte(""),
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgVPCStateResponse,
					Message: primitives.JSONMsgVPCState{
						SignedStateVal: primitives.VPCStateSigned{
							VPCState: primitives.VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(1000),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte("receiver-signature"),
						},
						Status: primitives.MessageStatusAccept,
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
				message: primitives.ChMsgPkt{},
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
		wantState    primitives.VPCStateSigned
	}{
		{
			name: "valid",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgVPCStateRequest,
					Message: primitives.JSONMsgVPCState{
						SignedStateVal: primitives.VPCStateSigned{
							VPCState: primitives.VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: primitives.MessageStatusRequire,
					},
				}},
			wantErr: false,
			wantState: primitives.VPCStateSigned{
				VPCState: primitives.VPCState{
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
				message: primitives.ChMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgVPCStateRequest,
					Message:   nil,
				}},
			wantErr: true,
		},
		{
			name: "invalid-status-accept",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgVPCStateRequest,
					Message: primitives.JSONMsgVPCState{
						SignedStateVal: primitives.VPCStateSigned{
							VPCState: primitives.VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: primitives.MessageStatusAccept,
					},
				}},
			wantErr: true,
		},
		{
			name: "invalid-status-decline",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgVPCStateRequest,
					Message: primitives.JSONMsgVPCState{
						SignedStateVal: primitives.VPCStateSigned{
							VPCState: primitives.VPCState{
								ID:              []byte("some-valid-id"),
								BlockedSender:   big.NewInt(10),
								BlockedReceiver: big.NewInt(10),
								Version:         big.NewInt(10),
							},
							SignSender:   []byte("sender-signature"),
							SignReceiver: []byte(""),
						},
						Status: primitives.MessageStatusDecline,
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
		state  primitives.VPCStateSigned
		status primitives.MessageStatus
	}
	tests := []struct {
		name              string
		args              args
		expectResponse    primitives.ChMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "valid-accept",
			args: args{
				state: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: primitives.MessageStatusAccept,
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateResponse,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: primitives.MessageStatusAccept,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "valid-decline",
			args: args{
				state: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: primitives.MessageStatusDecline,
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateResponse,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: primitives.MessageStatusDecline,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "invalid-status-require",
			args: args{
				state: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature"),
				},
				status: primitives.MessageStatusRequire,
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateResponse,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           true,
		},
		{
			name: "write-error",
			args: args{
				state:  primitives.VPCStateSigned{},
				status: primitives.MessageStatusRequire,
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
		sid primitives.SessionID
	}
	tests := []struct {
		name              string
		args              args
		expectRequest     primitives.ChMsgPkt
		expectMatchInMock bool
		mockResponse      jsonMsgPacket
		responseError     error
		wantErr           bool
		wantStatus        primitives.MessageStatus
		wantSessionID     primitives.SessionID
	}{
		{
			name: "valid-accept",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDRequest,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart: []byte("sid-sender-part"),
						AddrSender:    aliceID.OnChainID,
						NonceSender:   big.NewInt(10).Bytes(),
						Locked:        false,
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgSessionIDResponse,
					Message: primitives.JSONMsgSessionID{
						Sid: primitives.SessionID{
							SidSenderPart:   []byte("sid-sender-part"),
							AddrSender:      aliceID.OnChainID,
							NonceSender:     big.NewInt(10).Bytes(),
							SidReceiverPart: []byte("sid-receiver-part"),
							AddrReceiver:    bobID.OnChainID,
							NonceReceiver:   big.NewInt(20).Bytes(),
							Locked:          false,
						},
						Status: primitives.MessageStatusAccept,
					},
				}},
			wantErr:    false,
			wantStatus: primitives.MessageStatusAccept,
			wantSessionID: primitives.SessionID{
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
				sid: primitives.SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDRequest,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart: []byte("sid-sender-part"),
						AddrSender:    aliceID.OnChainID,
						NonceSender:   big.NewInt(10).Bytes(),
						Locked:        false,
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgSessionIDResponse,
					Message: primitives.JSONMsgSessionID{
						Sid: primitives.SessionID{
							SidSenderPart: []byte("sid-sender-part"),
							AddrSender:    aliceID.OnChainID,
							NonceSender:   big.NewInt(10).Bytes(),
							Locked:        false,
						},
						Status: primitives.MessageStatusDecline,
					},
				}},
			wantErr:    false,
			wantStatus: primitives.MessageStatusDecline,
			wantSessionID: primitives.SessionID{
				SidSenderPart: []byte("sid-sender-part"),
				AddrSender:    aliceID.OnChainID,
				NonceSender:   big.NewInt(10).Bytes(),
				Locked:        false,
			},
		},
		{
			name: "write-error",
			args: args{
				sid: primitives.SessionID{},
			},
			responseError: fmt.Errorf("write-error"),
			wantErr:       true,
		},
		{
			name: "read-error",
			args: args{
				sid: primitives.SessionID{},
			},
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{},
				err:     fmt.Errorf("read-error")},
			wantErr: true,
		},
		{
			name: "invalid-message-id",
			args: args{
				sid: primitives.SessionID{},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDRequest,
				Message: primitives.JSONMsgSessionID{
					Sid:    primitives.SessionID{},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-id"),
				}},
			wantErr: true,
		},
		{
			name: "invalid-message",
			args: args{
				sid: primitives.SessionID{},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDRequest,
				Message: primitives.JSONMsgSessionID{
					Sid:    primitives.SessionID{},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgSessionIDResponse,
					Message:   nil,
				}},
			wantErr: true,
		},
		{
			name: "peer-modifies-sender-component",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},
			expectRequest: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDRequest,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart: []byte("sid-sender-part"),
						AddrSender:    aliceID.OnChainID,
						NonceSender:   big.NewInt(10).Bytes(),
						Locked:        false,
					},
					Status: primitives.MessageStatusRequire,
				},
			},
			expectMatchInMock: true,
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgSessionIDResponse,
					Message: primitives.JSONMsgSessionID{
						Sid: primitives.SessionID{
							SidSenderPart:   []byte("sid-modified-sender-part"),
							AddrSender:      aliceID.OnChainID,
							NonceSender:     big.NewInt(10).Bytes(),
							SidReceiverPart: []byte("sid-receiver-part"),
							AddrReceiver:    bobID.OnChainID,
							NonceReceiver:   big.NewInt(20).Bytes(),
							Locked:          false,
						},
						Status: primitives.MessageStatusAccept,
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
		wantSid      primitives.SessionID
	}{
		{
			name: "valid-1",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgSessionIDRequest,
					Message: primitives.JSONMsgSessionID{
						Sid: primitives.SessionID{
							SidSenderPart: []byte("sid-sender-part"),
							AddrSender:    aliceID.OnChainID,
							NonceSender:   big.NewInt(10).Bytes(),
							Locked:        false,
						},
						Status: primitives.MessageStatusRequire,
					},
				}},
			wantErr: false,
			wantSid: primitives.SessionID{
				SidSenderPart: []byte("sid-sender-part"),
				AddrSender:    aliceID.OnChainID,
				NonceSender:   big.NewInt(10).Bytes(),
				Locked:        false,
			},
		},
		{
			name: "read-error",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{},
				err:     fmt.Errorf("read-error"),
			},
			wantErr: true,
		},
		{
			name: "messageId-error",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MessageID("invalid-message-id"),
				},
			},
			wantErr: true,
		},
		{
			name: "message-error",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgSessionIDRequest,
					Message:   nil,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-status-accept",
			mockResponse: jsonMsgPacket{
				message: primitives.ChMsgPkt{
					MessageID: primitives.MsgSessionIDRequest,
					Message: primitives.JSONMsgSessionID{
						Sid: primitives.SessionID{
							SidSenderPart: []byte("sid-sender-part"),
							AddrSender:    aliceID.OnChainID,
							NonceSender:   big.NewInt(10).Bytes(),
							Locked:        false,
						},
						Status: primitives.MessageStatusAccept,
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
		sid    primitives.SessionID
		status primitives.MessageStatus
	}
	tests := []struct {
		name              string
		args              args
		expectResponse    primitives.ChMsgPkt
		expectMatchInMock bool
		responseError     error
		wantErr           bool
	}{
		{
			name: "valid-accept",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart:   []byte("sid-sender-part"),
					AddrSender:      aliceID.OnChainID,
					NonceSender:     big.NewInt(10).Bytes(),
					SidReceiverPart: []byte("sid-receiver-part"),
					AddrReceiver:    bobID.OnChainID,
					NonceReceiver:   big.NewInt(20).Bytes(),
					Locked:          false,
				},
				status: primitives.MessageStatusAccept,
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart:   []byte("sid-sender-part"),
						AddrSender:      aliceID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false,
					},
					Status: primitives.MessageStatusAccept,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "valid-decline",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart:   []byte("sid-sender-part"),
					AddrSender:      aliceID.OnChainID,
					NonceSender:     big.NewInt(10).Bytes(),
					SidReceiverPart: []byte("sid-receiver-part"),
					AddrReceiver:    bobID.OnChainID,
					NonceReceiver:   big.NewInt(20).Bytes(),
					Locked:          false,
				},
				status: primitives.MessageStatusDecline,
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart:   []byte("sid-sender-part"),
						AddrSender:      aliceID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false,
					},
					Status: primitives.MessageStatusDecline,
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           false,
		},
		{
			name: "invalid-status",
			args: args{
				sid:    primitives.SessionID{},
				status: primitives.MessageStatus("invalid-status"),
			},
			expectResponse: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid:    primitives.SessionID{},
					Status: primitives.MessageStatus("invalid-status"),
				},
			},
			expectMatchInMock: true,
			responseError:     nil,
			wantErr:           true,
		},
		{
			name: "write-error",
			args: args{
				sid:    primitives.SessionID{},
				status: primitives.MessageStatusDecline,
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
