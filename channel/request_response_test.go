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
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/direct-state-transfer/dst-go/channel/primitives"
	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
)

const testChannelMessageProtocolVersion = "0.1"

var testTime time.Time
var contractStoreVersionForTest = []byte("8b9df51d7a7343a92133f24aaa1ba544d92eac1ccd66c731aa243e3273cb0f69")

func init() {
	testTime, _ = time.Parse(
		time.RFC3339, "2012-11-01T22:08:41+00:00")
}

func marshalChMsgPkt(pkt primitives.ChMsgPkt) []byte {
	byteArray, err := json.Marshal(pkt)
	if err != nil {
		byteArray = []byte{}
	}
	return byteArray
}

func Test_Instance_IdentityRequest(t *testing.T) {
	type args struct {
		selfID identity.OffChainID
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error
		mockReadReturnMsg  primitives.ChMsgPkt
		mockReadReturnErr  error

		wantErr    bool
		wantPeerID identity.OffChainID
	}{
		{
			name: "valid-1",
			args: args{
				selfID: aliceID,
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message:   primitives.JSONMsgIdentity{ID: aliceID}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityResponse,
				Message:   primitives.JSONMsgIdentity{ID: bobID}},
			mockReadReturnErr: nil,

			wantErr:    false,
			wantPeerID: bobID,
		},
		{
			name: "write-error",
			args: args{
				selfID: aliceID,
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message:   primitives.JSONMsgIdentity{ID: aliceID},
			},
			mockWriteReturnErr: fmt.Errorf("mock-error"),

			wantErr: true,
		},
		{
			name: "read-error",
			args: args{
				selfID: aliceID,
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message:   primitives.JSONMsgIdentity{ID: aliceID},
			},
			mockWriteReturnErr: nil,
			mockReadReturnMsg:  primitives.ChMsgPkt{},
			mockReadReturnErr:  fmt.Errorf("mock-error"),

			wantErr: true,
		},
		{
			name: "message-id-mismatch",
			args: args{
				selfID: aliceID,
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message: primitives.JSONMsgIdentity{
					ID: aliceID,
				}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDRequest,
				Message:   primitives.JSONMsgSessionID{},
			},
			mockReadReturnErr: nil,

			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

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
		})
	}
}

func Test_Instance_IdentityRead(t *testing.T) {
	tests := []struct {
		name string

		mockReadReturnMsg primitives.ChMsgPkt
		mockReadReturnErr error

		wantErr    bool
		wantPeerID identity.OffChainID
	}{
		{
			name: "valid",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message: primitives.JSONMsgIdentity{
					ID: aliceID}},
			mockReadReturnErr: nil,
			wantErr:           false,
			wantPeerID:        aliceID,
		},
		{
			name: "read-error",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message: primitives.JSONMsgIdentity{
					ID: aliceID}},
			mockReadReturnErr: fmt.Errorf("mock-error"),
			wantErr:           true,
		},
		{
			name: "message-id-mismatch",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message:   primitives.JSONMsgContractAddr{}},
			mockReadReturnErr: nil,
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()
			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			mockAdapter := MockReadWriteCloser{}
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

			gotPeerID, err := ch.IdentityRead()
			if err != nil {
				t.Logf("channel.IdentityRead() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
			if !reflect.DeepEqual(gotPeerID, tt.wantPeerID) {
				t.Errorf("channel.IdentityRead() = %v, want %v", gotPeerID, tt.wantPeerID)
			}

		})
	}
}

func Test_Instance_IdentityRespond(t *testing.T) {
	type args struct {
		selfID identity.OffChainID
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error

		wantErr bool
	}{
		{
			name: "valid-1",
			args: args{
				selfID: aliceID},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityResponse,
				Message: primitives.JSONMsgIdentity{
					ID: aliceID,
				},
			},
			mockWriteReturnErr: nil,

			wantErr: false,
		},
		{
			name: "write-error",
			args: args{
				selfID: aliceID},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityResponse,
				Message: primitives.JSONMsgIdentity{
					ID: aliceID,
				},
			},
			mockWriteReturnErr: fmt.Errorf("mock-error"),

			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

			err := ch.IdentityRespond(tt.args.selfID)
			if err != nil {
				t.Logf("channel.RespondToContractAddr() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
		})
	}
}
func Test_Instance_NewChannelRequest(t *testing.T) {
	type args struct {
		msgProtocolVersion   string
		contractStoreVersion []byte
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error
		mockReadReturnMsg  primitives.ChMsgPkt
		mockReadReturnErr  error

		wantErr    bool
		wantAccept primitives.MessageStatus
		wantReason string
	}{
		{
			name: "accept",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0")},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelResponse,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusAccept,
					Reason:               ""}},
			mockReadReturnErr: nil,

			wantErr:    false,
			wantAccept: primitives.MessageStatusAccept,
			wantReason: "",
		},
		{
			name: "decline",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0")},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelResponse,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusDecline,
					Reason:               "decline-for-test"}},
			mockReadReturnErr: nil,

			wantErr:    false,
			wantAccept: primitives.MessageStatusDecline,
			wantReason: "decline-for-test",
		},
		{
			name: "write-error",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0")},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusRequire}},
			mockWriteReturnErr: fmt.Errorf("mock-error"),

			wantErr: true,
		},
		{
			name: "read-error",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0")},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg:  primitives.ChMsgPkt{},
			mockReadReturnErr:  fmt.Errorf("mock-error"),

			wantErr: true,
		},
		{
			name: "message-id-mismatch",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v1.0")},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDRequest,
				Message:   primitives.JSONMsgSessionID{},
			},
			mockReadReturnErr: nil,

			wantErr: true,
		},
		{
			name: "contract-version-mismatch",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: []byte("v2.0")},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v2.0"),
					Status:               primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelResponse,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusAccept,
					Reason:               ""}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
		{
			name: "protocol-version-mismatch",
			args: args{
				msgProtocolVersion:   "2.0",
				contractStoreVersion: []byte("v1.0")},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "2.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelResponse,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: []byte("v1.0"),
					Status:               primitives.MessageStatusAccept,
					Reason:               ""}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

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
		})
	}
}

func Test_Instance_NewChannelRead(t *testing.T) {
	tests := []struct {
		name string

		mockReadReturnMsg primitives.ChMsgPkt
		mockReadReturnErr error

		wantErr                  bool
		wantMsgProtocolVersion   string
		wantContractStoreVersion []byte
	}{
		{
			name: "valid-1",

			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               primitives.MessageStatusRequire}},
			mockReadReturnErr: nil,

			wantErr:                  false,
			wantMsgProtocolVersion:   "1.0",
			wantContractStoreVersion: contractStoreVersionForTest,
		},
		{
			name: "read-error",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               primitives.MessageStatusRequire}},
			mockReadReturnErr: fmt.Errorf("mock-error"),
			wantErr:           true,
		},
		{
			name: "message-id-mismatch",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message:   primitives.JSONMsgMSCBaseState{}},
			mockReadReturnErr: nil,
			wantErr:           true,
		},
		{
			name: "status-is-not-require",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelRequest,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()
			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			mockAdapter := MockReadWriteCloser{}
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

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

		})
	}
}

func Test_Instance_NewChannelRespond(t *testing.T) {
	type args struct {
		msgProtocolVersion   string
		contractStoreVersion []byte
		accept               primitives.MessageStatus
		reason               string
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error

		wantErr bool
	}{
		{
			name: "valid-accept",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: contractStoreVersionForTest,
				accept:               primitives.MessageStatusAccept,
				reason:               ""},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelResponse,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               primitives.MessageStatusAccept,
					Reason:               ""}},
			mockWriteReturnErr: nil,
			wantErr:            false,
		},
		{
			name: "valid-decline",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: contractStoreVersionForTest,
				accept:               primitives.MessageStatusDecline,
				reason:               ""},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelResponse,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               primitives.MessageStatusDecline,
					Reason:               ""}},
			mockWriteReturnErr: nil,
			wantErr:            false,
		},
		{
			name: "invalid-status-require",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: contractStoreVersionForTest,
				accept:               primitives.MessageStatusRequire,
				reason:               ""},
			wantErr: true,
		},
		{
			name: "write-error",
			args: args{
				msgProtocolVersion:   "1.0",
				contractStoreVersion: contractStoreVersionForTest,
				accept:               primitives.MessageStatusAccept,
				reason:               ""},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgNewChannelResponse,
				Message: primitives.JSONMsgNewChannel{
					MsgProtocolVersion:   "1.0",
					ContractStoreVersion: contractStoreVersionForTest,
					Status:               primitives.MessageStatusAccept,
					Reason:               ""}},
			mockWriteReturnErr: fmt.Errorf("mock-error"),
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

			err := ch.NewChannelRespond(tt.args.msgProtocolVersion, tt.args.contractStoreVersion, tt.args.accept, tt.args.reason)
			if (err != nil) != tt.wantErr {
				t.Errorf("channel.NewChannelRespond() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func Test_Instance_ContractAddrRequest(t *testing.T) {
	type args struct {
		addr types.Address
		id   contract.Handler
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error
		mockReadReturnMsg  primitives.ChMsgPkt
		mockReadReturnErr  error

		wantErr    bool
		wantStatus primitives.MessageStatus
	}{
		{
			name: "accept",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require"}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrResponse,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,

			wantErr:    false,
			wantStatus: primitives.MessageStatusAccept,
		},
		{
			name: "decline",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require"}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrResponse,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusDecline}},
			mockReadReturnErr: nil,

			wantErr:    false,
			wantStatus: primitives.MessageStatusDecline,
		},
		{
			name: "write-error",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require"}},
			mockWriteReturnErr: fmt.Errorf("mock-error"),

			wantErr: true,
		},
		{
			name: "read-error",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require"}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrResponse,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusAccept}},
			mockReadReturnErr: fmt.Errorf("mock-error"),

			wantErr: true,
		},
		{
			name: "message-id-mismatch",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require"}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDRequest,
				Message:   primitives.JSONMsgSessionID{}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
		{
			name: "contract-type-mismatch",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require"}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrResponse,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.MSContract(),
					Status:       primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
		{
			name: "contract-addr-mismatch",
			args: args{
				addr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:   contract.Store.LibSignatures(),
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       "require"}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrResponse,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("9cB3cD41C00847b3655B5bEB82A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

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
		})
	}
}

func Test_Instance_ContractAddrRead(t *testing.T) {
	tests := []struct {
		name string

		mockReadReturnMsg primitives.ChMsgPkt
		mockReadReturnErr error

		wantErr  bool
		wantAddr types.Address
		wantID   contract.Handler
	}{
		{
			name: "valid",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusRequire}},
			wantErr:  false,
			wantAddr: types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
			wantID:   contract.Store.LibSignatures(),
		},
		{
			name: "read-error",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusRequire}},
			mockReadReturnErr: fmt.Errorf("mock-error"),
			wantErr:           true,
		},
		{
			name: "message-id-mismatch",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateRequest,
				Message:   primitives.JSONMsgMSCBaseState{}},
			mockReadReturnErr: nil,
			wantErr:           true,
		},
		{
			name: "status-is-not-require",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,
			wantErr:           true,
		},
	}
	wg := &sync.WaitGroup{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()
			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			mockAdapter := MockReadWriteCloser{}
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

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
			if !gotID.Equal(tt.wantID) {
				t.Errorf("channel.ReadContractAddr() gotId = %v, want %v", gotID, tt.wantID)
			}

			wg.Wait()
		})
	}
}

func Test_Instance_ContractAddrRespond(t *testing.T) {
	type args struct {
		addr   types.Address
		id     contract.Handler
		status primitives.MessageStatus
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error

		wantErr bool
	}{
		{
			name: "valid-accept",
			args: args{
				addr:   types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:     contract.Store.LibSignatures(),
				status: primitives.MessageStatusAccept,
			},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrResponse,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusAccept,
				},
			},
			mockWriteReturnErr: nil,
			wantErr:            false,
		},
		{
			name: "valid-decline",
			args: args{
				addr:   types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:     contract.Store.LibSignatures(),
				status: primitives.MessageStatusDecline,
			},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrResponse,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusDecline,
				},
			},
			mockWriteReturnErr: nil,
			wantErr:            false,
		},
		{
			name: "invalid-status-require",
			args: args{
				addr:   types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:     contract.Store.LibSignatures(),
				status: primitives.MessageStatusRequire,
			},
			wantErr: true,
		},
		{
			name: "write-error",
			args: args{
				addr:   types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
				id:     contract.Store.LibSignatures(),
				status: primitives.MessageStatusAccept,
			},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrResponse,
				Message: primitives.JSONMsgContractAddr{
					Addr:         types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
					ContractType: contract.Store.LibSignatures(),
					Status:       primitives.MessageStatusAccept,
				},
			},
			mockWriteReturnErr: fmt.Errorf("mock-error"),
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

			err := ch.ContractAddrRespond(tt.args.addr, tt.args.id, tt.args.status)
			if err != nil {
				t.Logf("channel.RespondToContractAddr() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
		})
	}
}

func Test_Instance_NewMSCBaseStateRequest(t *testing.T) {
	type args struct {
		newState primitives.MSCBaseStateSigned
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error
		mockReadReturnMsg  primitives.ChMsgPkt
		mockReadReturnErr  error

		wantErr           bool
		wantResponseState primitives.MSCBaseStateSigned
		wantStatus        primitives.MessageStatus
	}{
		{
			name: "accept",
			args: args{
				newState: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
					Status: primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
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
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,

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
			name: "decline",
			args: args{
				newState: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
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
					Status: primitives.MessageStatusDecline}},
			mockReadReturnErr: nil,

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
			wantStatus: primitives.MessageStatusDecline,
		},
		{
			name: "write-error",
			args: args{
				newState: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: fmt.Errorf("mock-error"),

			wantErr: true,
		},
		{
			name: "read-error",
			args: args{
				newState: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
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
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: fmt.Errorf("mock-error"),

			wantErr: true,
		},
		{
			name: "message-id-mismatch",
			args: args{
				newState: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message:   primitives.JSONMsgContractAddr{}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
		{
			name: "state-mismatch",
			args: args{
				newState: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateResponse,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							//
							//Address modified by user
							//
							VpcAddress:      types.HexToAddress("C00A2764847b3655B5bEB829cB3cD418de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

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
		})
	}
}

func Test_Instance_NewMSCBaseStateRead(t *testing.T) {
	tests := []struct {
		name string

		mockReadReturnMsg primitives.ChMsgPkt
		mockReadReturnErr error

		wantErr   bool
		wantState primitives.MSCBaseStateSigned
	}{
		{
			name: "valid",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateRequest,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusRequire}},
			mockReadReturnErr: nil,
			wantErr:           false,
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
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateRequest,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusRequire}},
			mockReadReturnErr: fmt.Errorf("mock-error"),
			wantErr:           true,
		},
		{
			name: "message-id-mismatch",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message:   primitives.JSONMsgIdentity{}},
			mockReadReturnErr: nil,
			wantErr:           true,
		},
		{
			name: "status-is-not-require",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateRequest,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()
			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			mockAdapter := MockReadWriteCloser{}
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

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
		})
	}
}

func Test_Instance_NewMSCBaseStateRespond(t *testing.T) {
	type args struct {
		state  primitives.MSCBaseStateSigned
		status primitives.MessageStatus
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error

		wantErr bool
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
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature")},
				status: primitives.MessageStatusAccept},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateResponse,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature")},
					Status: primitives.MessageStatusAccept}},
			mockWriteReturnErr: nil,
			wantErr:            false,
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
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature")},
				status: primitives.MessageStatusDecline},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateResponse,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature")},
					Status: primitives.MessageStatusDecline}},
			mockWriteReturnErr: nil,
			wantErr:            true,
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
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature")},
				status: primitives.MessageStatusRequire},
			wantErr: true,
		},
		{
			name: "write-error",
			args: args{
				state: primitives.MSCBaseStateSigned{
					MSContractBaseState: primitives.MSCBaseState{
						VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
						Sid:             big.NewInt(10),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature")},
				status: primitives.MessageStatusAccept},

			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgMSCBaseStateResponse,
				Message: primitives.JSONMsgMSCBaseState{
					SignedStateVal: primitives.MSCBaseStateSigned{
						MSContractBaseState: primitives.MSCBaseState{
							VpcAddress:      types.HexToAddress("847b3655B5bEB829cB3cD41C00A27648de737C39"),
							Sid:             big.NewInt(10),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature")},
					Status: primitives.MessageStatusAccept}},
			mockWriteReturnErr: fmt.Errorf("mock-error"),
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

			err := ch.NewMSCBaseStateRespond(tt.args.state, tt.args.status)
			if err != nil {
				t.Logf("channel.NewMSCBaseStateRespond() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
		})
	}
}

func Test_Instance_NewVPCStateRequest(t *testing.T) {
	type args struct {
		newStateSigned primitives.VPCStateSigned
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error
		mockReadReturnMsg  primitives.ChMsgPkt
		mockReadReturnErr  error

		wantErr           bool
		wantResponseState primitives.VPCStateSigned
		wantStatus        primitives.MessageStatus
	}{
		{
			name: "accept",
			args: args{
				newStateSigned: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")}},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
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
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,

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
			name: "decline",
			args: args{
				newStateSigned: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")}},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
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
					Status: primitives.MessageStatusDecline}},
			mockReadReturnErr: nil,

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
			wantStatus: primitives.MessageStatusDecline,
		},
		{
			name: "write-error",
			args: args{
				newStateSigned: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")}},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusRequire}},
			mockWriteReturnErr: fmt.Errorf("mock-error"),

			wantErr: true,
		},
		{
			name: "read-error",
			args: args{
				newStateSigned: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")}},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
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
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: fmt.Errorf("mock-error"),

			wantErr: true,
		},
		{
			name: "message-id-mismatch",
			args: args{
				newStateSigned: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")}},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message:   primitives.JSONMsgIdentity{}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
		{
			name: "state-mismatch",
			args: args{
				newStateSigned: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10),
					},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("")}},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusRequire}},
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateResponse,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id-modified-by-peer"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10),
						},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature"),
					},
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

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
		})
	}
}

func Test_Instance_NewVPCStateRead(t *testing.T) {
	tests := []struct {
		name string

		mockReadReturnMsg primitives.ChMsgPkt
		mockReadReturnErr error

		wantErr   bool
		wantState primitives.VPCStateSigned
	}{
		{
			name: "valid",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateRequest,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusRequire}},
			mockReadReturnErr: nil,
			wantErr:           false,
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
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateRequest,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusRequire}},
			mockReadReturnErr: fmt.Errorf("mock-error"),
			wantErr:           true,
		},
		{
			name: "message-id-mismatch",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgContractAddrRequest,
				Message:   primitives.JSONMsgContractAddr{}},
			mockReadReturnErr: nil,
			wantErr:           true,
		},
		{
			name: "status-is-not-require",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateRequest,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("")},
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()
			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			mockAdapter := MockReadWriteCloser{}
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

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
		})
	}
}

func Test_Instance_NewVPCStateRespond(t *testing.T) {
	type args struct {
		state  primitives.VPCStateSigned
		status primitives.MessageStatus
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error

		wantErr bool
	}{
		{
			name: "valid-accept",
			args: args{
				state: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature")},
				status: primitives.MessageStatusAccept},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateResponse,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature")},
					Status: primitives.MessageStatusAccept}},
			mockWriteReturnErr: nil,
			wantErr:            false,
		},
		{
			name: "valid-decline",
			args: args{
				state: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature")},
				status: primitives.MessageStatusDecline},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateResponse,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature")},
					Status: primitives.MessageStatusDecline}},
			mockWriteReturnErr: nil,
			wantErr:            false,
		},
		{
			name: "invalid-require",
			args: args{
				state: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature")},
				status: primitives.MessageStatusRequire},
			wantErr: true,
		},
		{
			name: "write-error",
			args: args{
				state: primitives.VPCStateSigned{
					VPCState: primitives.VPCState{
						ID:              []byte("some-valid-id"),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(10),
						Version:         big.NewInt(10)},
					SignSender:   []byte("sender-signature"),
					SignReceiver: []byte("receiver-signature")},
				status: primitives.MessageStatusAccept},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgVPCStateResponse,
				Message: primitives.JSONMsgVPCState{
					SignedStateVal: primitives.VPCStateSigned{
						VPCState: primitives.VPCState{
							ID:              []byte("some-valid-id"),
							BlockedSender:   big.NewInt(10),
							BlockedReceiver: big.NewInt(10),
							Version:         big.NewInt(10)},
						SignSender:   []byte("sender-signature"),
						SignReceiver: []byte("receiver-signature")},
					Status: primitives.MessageStatusAccept}},
			mockWriteReturnErr: fmt.Errorf("mock-error"),
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

			err := ch.NewVPCStateRespond(tt.args.state, tt.args.status)
			if err != nil {
				t.Logf("channel.NewVPCStateRespond() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
		})
	}
}
func Test_Instance_SessionIdRequest(t *testing.T) {
	type args struct {
		sid primitives.SessionID
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error
		mockReadReturnMsg  primitives.ChMsgPkt
		mockReadReturnErr  error

		wantErr       bool
		wantStatus    primitives.MessageStatus
		wantSessionID primitives.SessionID
	}{
		{
			name: "accept",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart:   []byte("sid-sender-part"),
						AddrSender:      aliceID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false},
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,

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
			name: "decline",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart:   []byte("sid-sender-part"),
						AddrSender:      aliceID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false},
					Status: primitives.MessageStatusDecline}},
			mockReadReturnErr: nil,

			wantErr:    false,
			wantStatus: primitives.MessageStatusDecline,
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
			name: "write-error",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: fmt.Errorf("mock-error"),
			wantErr:            true,
		},
		{
			name: "read-error",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart:   []byte("sid-sender-part"),
						AddrSender:      aliceID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false},
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: fmt.Errorf("mock-error"),

			wantErr: true,
		},
		{
			name: "message-id-mismatch",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message:   primitives.JSONMsgIdentity{}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
		{
			name: "sender-part-modified",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						//
						// Sender Part Modified
						SidSenderPart:   []byte("sid-sender-part-modified"),
						AddrSender:      aliceID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false},
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
		{
			name: "sender-addr-modified",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart: []byte("sid-sender-part"),
						//
						// Sender Addr Modified
						AddrSender:      bobID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false},
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
		{
			name: "sender-nonce-modified",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart: []byte("sid-sender-part"),
					AddrSender:    aliceID.OnChainID,
					NonceSender:   big.NewInt(10).Bytes(),
					Locked:        false,
				},
			},

			mockWriteExpectMsg: primitives.ChMsgPkt{
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
			mockWriteReturnErr: nil,
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart: []byte("sid-sender-part"),
						AddrSender:    aliceID.OnChainID,
						//
						// Sender Nonce Modified
						NonceSender:     big.NewInt(100).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false},
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,

			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

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
		})
	}
}

func Test_Instance_SessionIdRead(t *testing.T) {
	tests := []struct {
		name string

		mockReadReturnMsg primitives.ChMsgPkt
		mockReadReturnErr error

		wantErr bool
		wantSid primitives.SessionID
	}{
		{
			name: "valid",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDRequest,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart: []byte("sid-sender-part"),
						AddrSender:    aliceID.OnChainID,
						NonceSender:   big.NewInt(10).Bytes(),
						Locked:        false,
					},
					Status: primitives.MessageStatusRequire}},
			mockReadReturnErr: nil,
			wantErr:           false,
			wantSid: primitives.SessionID{
				SidSenderPart: []byte("sid-sender-part"),
				AddrSender:    aliceID.OnChainID,
				NonceSender:   big.NewInt(10).Bytes(),
				Locked:        false,
			},
		},
		{
			name: "read-error",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDRequest,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart: []byte("sid-sender-part"),
						AddrSender:    aliceID.OnChainID,
						NonceSender:   big.NewInt(10).Bytes(),
						Locked:        false,
					},
					Status: primitives.MessageStatusRequire}},
			mockReadReturnErr: fmt.Errorf("mock-error"),
			wantErr:           true,
		},
		{
			name: "message-id-mismatch",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgIdentityRequest,
				Message:   primitives.JSONMsgIdentity{}},
			mockReadReturnErr: nil,
			wantErr:           true,
		},
		{
			name: "status-is-not-require",
			mockReadReturnMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDRequest,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart: []byte("sid-sender-part"),
						AddrSender:    aliceID.OnChainID,
						NonceSender:   big.NewInt(10).Bytes(),
						Locked:        false,
					},
					Status: primitives.MessageStatusAccept}},
			mockReadReturnErr: nil,
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()
			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			mockAdapter := MockReadWriteCloser{}
			mockAdapter.On("Read").Return(marshalChMsgPkt(tt.mockReadReturnMsg), tt.mockReadReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

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

		})
	}
}

func Test_Instance_SessionIdRespond(t *testing.T) {
	type args struct {
		sid    primitives.SessionID
		status primitives.MessageStatus
	}
	tests := []struct {
		name string
		args args

		mockWriteExpectMsg primitives.ChMsgPkt
		mockWriteReturnErr error

		wantErr bool
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
					Locked:          false},
				status: primitives.MessageStatusAccept},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart:   []byte("sid-sender-part"),
						AddrSender:      aliceID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false},
					Status: primitives.MessageStatusAccept}},
			mockWriteReturnErr: nil,
			wantErr:            false,
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
					Locked:          false},
				status: primitives.MessageStatusDecline},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart:   []byte("sid-sender-part"),
						AddrSender:      aliceID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false},
					Status: primitives.MessageStatusDecline}},
			mockWriteReturnErr: nil,
			wantErr:            true,
		},
		{
			name: "invalid-status-require",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart:   []byte("sid-sender-part"),
					AddrSender:      aliceID.OnChainID,
					NonceSender:     big.NewInt(10).Bytes(),
					SidReceiverPart: []byte("sid-receiver-part"),
					AddrReceiver:    bobID.OnChainID,
					NonceReceiver:   big.NewInt(20).Bytes(),
					Locked:          false},
				status: primitives.MessageStatusRequire},
			wantErr: true,
		},
		{
			name: "write-error",
			args: args{
				sid: primitives.SessionID{
					SidSenderPart:   []byte("sid-sender-part"),
					AddrSender:      aliceID.OnChainID,
					NonceSender:     big.NewInt(10).Bytes(),
					SidReceiverPart: []byte("sid-receiver-part"),
					AddrReceiver:    bobID.OnChainID,
					NonceReceiver:   big.NewInt(20).Bytes(),
					Locked:          false},
				status: primitives.MessageStatusAccept},
			mockWriteExpectMsg: primitives.ChMsgPkt{
				MessageID: primitives.MsgSessionIDResponse,
				Message: primitives.JSONMsgSessionID{
					Sid: primitives.SessionID{
						SidSenderPart:   []byte("sid-sender-part"),
						AddrSender:      aliceID.OnChainID,
						NonceSender:     big.NewInt(10).Bytes(),
						SidReceiverPart: []byte("sid-receiver-part"),
						AddrReceiver:    bobID.OnChainID,
						NonceReceiver:   big.NewInt(20).Bytes(),
						Locked:          false},
					Status: primitives.MessageStatusAccept}},
			mockWriteReturnErr: fmt.Errorf("mock-error"),
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			timestamp := time.Now()

			mockAdapter := MockReadWriteCloser{}

			timestampProvider := mockClock{}
			timestampProvider.On("Now").Return(timestamp)
			timestampProvider.On("SetLocation", mock.Anything).Return(nil)

			_ = timestampProvider.SetLocation("Local")

			tt.mockWriteExpectMsg.Version = testChannelMessageProtocolVersion
			tt.mockWriteExpectMsg.Timestamp = timestamp

			mockAdapter.On("Write", marshalChMsgPkt((tt.mockWriteExpectMsg))).Return(tt.mockWriteReturnErr)

			ch := &Instance{
				timestampProvider: &timestampProvider,
				adapter:           &mockAdapter,
			}

			err := ch.SessionIDRespond(tt.args.sid, tt.args.status)
			if err != nil {
				t.Logf("channel.SessionIdResponse() error = %v, wantErr %v", err, tt.wantErr)
				if !tt.wantErr {
					t.Fail()
				}
				return
			}
		})
	}
}
