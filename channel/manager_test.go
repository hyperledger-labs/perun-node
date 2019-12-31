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
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
	"github.com/direct-state-transfer/dst-go/log"
)

func Test_InitModule(t *testing.T) {
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "1",
			args: args{&Config{
				Logger: log.Config{
					Level:   log.DebugLevel,
					Backend: log.StdoutBackend,
				},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InitModule(tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("InitModule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_Instance_Connected(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     bool
	}{
		{
			name: "valid-connected",
			instance: &Instance{
				adapter: &genericChannelAdapter{
					connected: true,
				},
			},
			want: true,
		},
		{
			name: "valid-not-connected",
			instance: &Instance{
				adapter: &genericChannelAdapter{
					connected: false,
				},
			},
			want: false,
		},
		{
			name:     "invalid-nil-adapter",
			instance: &Instance{},
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.Connected()
			if got != tt.want {
				t.Errorf("Instance.Connected() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Instance_Close(t *testing.T) {

	t.Run("valid-close-no-error", func(t *testing.T) {

		adapter := &MockReadWriteCloser{}
		adapter.On("Close").Return(nil)

		inst := Instance{
			adapter: adapter,
		}

		err := inst.Close()
		if err != nil {
			t.Errorf("Instance.Close() error = nil, wantErr %v", err)
		}

		if !adapter.AssertNumberOfCalls(t, "Close", 1) {
			t.Errorf("Instance.Close() adapter.Close was not called")
		}
	})

	t.Run("valid-close-with-error", func(t *testing.T) {

		adapter := &MockReadWriteCloser{}
		adapter.On("Close").Return(fmt.Errorf("close-error"))

		inst := Instance{
			adapter: adapter,
		}

		err := inst.Close()
		if err == nil {
			t.Errorf("Instance.Close() error = nil, wantErr %v", err)
		}

		if !adapter.AssertNumberOfCalls(t, "Close", 1) {
			t.Errorf("Instance.Close() adapter.Close was not called")
		}
	})

	t.Run("invalid-nil-adapter", func(t *testing.T) {
		inst := Instance{}

		err := inst.Close()
		if err == nil {
			t.Errorf("Instance.Close() error = nil, wantErr %v", err)
		}
	})
}

func Test_Instance_SetClosingMode(t *testing.T) {
	type args struct {
		closingMode ClosingMode
	}
	tests := []struct {
		name     string
		instance *Instance
		args     args
		wantSet  bool
	}{
		{
			name:     "valid-Manual",
			instance: &Instance{},
			args: args{
				closingMode: ClosingModeManual,
			},
			wantSet: true,
		},
		{
			name:     "valid-AutoNormal",
			instance: &Instance{},
			args: args{
				closingMode: ClosingModeAutoNormal,
			},
			wantSet: true,
		},
		{
			name:     "valid-AutoImmediate",
			instance: &Instance{},
			args: args{
				closingMode: ClosingModeAutoImmediate,
			},
			wantSet: true,
		},
		{
			name:     "invalid-closing-mode",
			instance: &Instance{},
			args: args{
				closingMode: ClosingMode("invalid-closing-mode"),
			},
			wantSet: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.instance.SetClosingMode(tt.args.closingMode)
			if tt.wantSet && tt.instance.closingMode != tt.args.closingMode {
				t.Errorf("Instance.SetClosingMode() not set")
			}

		})
	}
}
func Test_Instance_ClosingMode(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     ClosingMode
	}{
		{
			name: "valid-Manual",
			instance: &Instance{
				closingMode: ClosingModeManual,
			},
			want: ClosingModeManual,
		},
		{
			name: "valid-AutoNormal",
			instance: &Instance{
				closingMode: ClosingModeAutoNormal,
			},
			want: ClosingModeAutoNormal,
		},
		{
			name: "valid-AutoImmediate",
			instance: &Instance{
				closingMode: ClosingModeAutoImmediate,
			},
			want: ClosingModeAutoImmediate,
		},
		{
			name: "invalid-closing-mode",
			instance: &Instance{
				closingMode: ClosingMode("invalid-closing-mode"),
			},
			want: ClosingMode("invalid-closing-mode"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.ClosingMode()
			if got != tt.want {
				t.Errorf("Instance.ClosingMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_Instance_setSelfID(t *testing.T) {
	type args struct {
		selfID identity.OffChainID
	}
	tests := []struct {
		name     string
		instance *Instance
		args     args
		wantSet  bool
	}{
		{
			name:     "valid-nil",
			instance: &Instance{},
			args: args{
				selfID: identity.OffChainID{},
			},
			wantSet: true,
		},
		{
			name:     "valid-not-nil",
			instance: &Instance{},
			args: args{
				selfID: aliceID,
			},
			wantSet: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.instance.setSelfID(tt.args.selfID)
			if tt.wantSet && tt.instance.selfID != tt.args.selfID {
				t.Errorf("Instance.setSelfID() not set")
			}

		})
	}
}

func Test_Instance_SelfID(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     identity.OffChainID
	}{
		{
			name: "valid-nil",
			instance: &Instance{
				selfID: identity.OffChainID{},
			},
			want: identity.OffChainID{},
		},
		{
			name: "valid-not-nil",
			instance: &Instance{
				selfID: aliceID,
			},
			want: aliceID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.SelfID()
			if got != tt.want {
				t.Errorf("Instance.SelfID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Instance_setPeerID(t *testing.T) {
	type args struct {
		peerID identity.OffChainID
	}
	tests := []struct {
		name     string
		instance *Instance
		args     args
		wantSet  bool
	}{
		{
			name:     "valid-nil",
			instance: &Instance{},
			args: args{
				peerID: identity.OffChainID{},
			},
			wantSet: true,
		},
		{
			name:     "valid-not-nil",
			instance: &Instance{},
			args: args{
				peerID: aliceID,
			},
			wantSet: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.instance.setPeerID(tt.args.peerID)
			if tt.wantSet && tt.instance.peerID != tt.args.peerID {
				t.Errorf("Instance.setPeerID() not set")
			}

		})
	}
}

func Test_Instance_PeerID(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     identity.OffChainID
	}{
		{
			name: "valid-nil",
			instance: &Instance{
				peerID: identity.OffChainID{},
			},
			want: identity.OffChainID{},
		},
		{
			name: "valid-not-nil",
			instance: &Instance{
				peerID: aliceID,
			},
			want: aliceID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.PeerID()
			if got != tt.want {
				t.Errorf("Instance.PeerID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Instance_SenderID(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     identity.OffChainID
	}{
		{
			name: "valid-nil",
			instance: &Instance{
				selfID:      identity.OffChainID{},
				roleChannel: Sender,
			},
			want: identity.OffChainID{},
		},
		{
			name: "valid-not-nil-self",
			instance: &Instance{
				selfID:      aliceID,
				roleChannel: Sender,
			},
			want: aliceID,
		},
		{
			name: "valid-not-nil-peer",
			instance: &Instance{
				peerID:      aliceID,
				roleChannel: Receiver,
			},
			want: aliceID,
		},
		{
			name: "invalid-role",
			instance: &Instance{
				roleChannel: Role("invalid-role"),
			},
			want: identity.OffChainID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.SenderID()
			if got != tt.want {
				t.Errorf("Instance.SenderID() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_Instance_ReceiverID(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     identity.OffChainID
	}{
		{
			name: "valid-nil",
			instance: &Instance{
				selfID:      identity.OffChainID{},
				roleChannel: Receiver,
			},
			want: identity.OffChainID{},
		},
		{
			name: "valid-not-nil-self",
			instance: &Instance{
				selfID:      aliceID,
				roleChannel: Receiver,
			},
			want: aliceID,
		},
		{
			name: "valid-not-nil-peer",
			instance: &Instance{
				peerID:      aliceID,
				roleChannel: Sender,
			},
			want: aliceID,
		},
		{
			name: "invalid-role",
			instance: &Instance{
				roleChannel: Role("invalid-role"),
			},
			want: identity.OffChainID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.ReceiverID()
			if got != tt.want {
				t.Errorf("Instance.ReceiverID() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_Instance_SetRoleChannel(t *testing.T) {
	type args struct {
		roleChannel Role
	}
	tests := []struct {
		name     string
		instance *Instance
		args     args
		wantSet  bool
	}{
		{
			name:     "valid-Sender",
			instance: &Instance{},
			args: args{
				roleChannel: Sender,
			},
			wantSet: true,
		},
		{
			name:     "valid-Receiver",
			instance: &Instance{},
			args: args{
				roleChannel: Receiver,
			},
			wantSet: true,
		},
		{
			name:     "invalid-role",
			instance: &Instance{},
			args: args{
				roleChannel: Role("invalid-role"),
			},
			wantSet: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.instance.SetRoleChannel(tt.args.roleChannel)
			if tt.wantSet && tt.instance.roleChannel != tt.args.roleChannel {
				t.Errorf("Instance.SetRoleChannel() not set")
			}

		})
	}
}
func Test_Instance_RoleChannel(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     Role
	}{
		{
			name: "valid-Sender",
			instance: &Instance{
				roleChannel: Sender,
			},
			want: Sender,
		},
		{
			name: "valid-Receiver",
			instance: &Instance{
				roleChannel: Receiver,
			},
			want: Receiver,
		},
		{
			name: "invalid-Role",
			instance: &Instance{
				roleChannel: Role("invalid-role"),
			},
			want: Role("invalid-role"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.RoleChannel()
			if got != tt.want {
				t.Errorf("Instance.RoleChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Instance_SetRoleClosing(t *testing.T) {
	type args struct {
		roleClosing Role
	}
	tests := []struct {
		name     string
		instance *Instance
		args     args
		wantSet  bool
	}{
		{
			name:     "valid-Sender",
			instance: &Instance{},
			args: args{
				roleClosing: Sender,
			},
			wantSet: true,
		},
		{
			name:     "valid-Receiver",
			instance: &Instance{},
			args: args{
				roleClosing: Receiver,
			},
			wantSet: true,
		},
		{
			name:     "invalid-role",
			instance: &Instance{},
			args: args{
				roleClosing: Role("invalid-role"),
			},
			wantSet: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.instance.SetRoleClosing(tt.args.roleClosing)
			if tt.wantSet && tt.instance.roleClosing != tt.args.roleClosing {
				t.Errorf("Instance.SetRoleClosing() not set")
			}

		})
	}
}
func Test_Instance_RoleClosing(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     Role
	}{
		{
			name: "valid-Sender",
			instance: &Instance{
				roleClosing: Sender,
			},
			want: Sender,
		},
		{
			name: "valid-Receiver",
			instance: &Instance{
				roleClosing: Receiver,
			},
			want: Receiver,
		},
		{
			name: "invalid-Role",
			instance: &Instance{
				roleClosing: Role("invalid-role"),
			},
			want: Role("invalid-role"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.RoleClosing()
			if got != tt.want {
				t.Errorf("Instance.RoleClosing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Instance_SetStatus(t *testing.T) {
	type args struct {
		status Status
	}
	tests := []struct {
		name     string
		instance *Instance
		args     args
		wantSet  bool
	}{
		{
			name: "valid-presetup-to-setup",
			instance: &Instance{
				status: PreSetup,
			},
			args: args{
				status: Setup,
			},
			wantSet: true,
		},
		{
			name: "invalid-init-to-setup",
			instance: &Instance{
				status: Init,
			},
			args: args{
				status: Setup,
			},
			wantSet: false,
		},
		{
			name: "valid-init-to-open",
			instance: &Instance{
				status: Init,
			},
			args: args{
				status: Open,
			},
			wantSet: true,
		},
		{
			name: "invalid-presetup-to-open",
			instance: &Instance{
				status: PreSetup,
			},
			args: args{
				status: Open,
			},
			wantSet: false,
		},
		{
			name: "valid-open-to-inconflict",
			instance: &Instance{
				status: Open,
			},
			args: args{
				status: InConflict,
			},
			wantSet: true,
		},
		{
			name: "valid-waitingtoclose-to-inconflict",
			instance: &Instance{
				status: WaitingToClose,
			},
			args: args{
				status: InConflict,
			},
			wantSet: true,
		},
		{
			name: "invalid-init-to-inconflict",
			instance: &Instance{
				status: Init,
			},
			args: args{
				status: InConflict,
			},
			wantSet: false,
		},
		{
			name: "valid-inconflict-to-settled",
			instance: &Instance{
				status: InConflict,
			},
			args: args{
				status: Settled,
			},
			wantSet: true,
		},
		{
			name: "invalid-init-to-settled",
			instance: &Instance{
				status: Init,
			},
			args: args{
				status: Settled,
			},
			wantSet: false,
		},
		{
			name: "valid-open-to-waitingtoclose",
			instance: &Instance{
				status: Open,
			},
			args: args{
				status: WaitingToClose,
			},
			wantSet: true,
		},
		{
			name: "invalid-init-to-waitingtoclose",
			instance: &Instance{
				status: Init,
			},
			args: args{
				status: WaitingToClose,
			},
			wantSet: false,
		},
		{
			name: "valid-settled-to-vpcclosing",
			instance: &Instance{
				status: Settled,
			},
			args: args{
				status: VPCClosing,
			},
			wantSet: true,
		},
		{
			name: "invalid-init-to-vpcclosing",
			instance: &Instance{
				status: Init,
			},
			args: args{
				status: VPCClosing,
			},
			wantSet: false,
		},
		{
			name: "valid-vpcclosing-to-vpcclosed",
			instance: &Instance{
				status: VPCClosing,
			},
			args: args{
				status: VPCClosed,
			},
			wantSet: true,
		},
		{
			name: "invalid-init-to-vpcclosed",
			instance: &Instance{
				status: Init,
			},
			args: args{
				status: VPCClosed,
			},
			wantSet: false,
		},
		{
			name: "valid-init-to-closed",
			instance: &Instance{
				status: Init,
			},
			args: args{
				status: Closed,
			},
			wantSet: true,
		},
		{
			name: "valid-vpcclosing-to-closed",
			instance: &Instance{
				status: VPCClosing,
			},
			args: args{
				status: Closed,
			},
			wantSet: true,
		},
		{
			name: "valid-vpcclosed-to-closed",
			instance: &Instance{
				status: VPCClosed,
			},
			args: args{
				status: Closed,
			},
			wantSet: true,
		},
		{
			name: "valid-waitingtoclose-to-closed",
			instance: &Instance{
				status: WaitingToClose,
			},
			args: args{
				status: Closed,
			},
			wantSet: true,
		},
		{
			name: "invalid-settled-to-closed",
			instance: &Instance{
				status: Settled,
			},
			args: args{
				status: Closed,
			},
			wantSet: false,
		},
		{
			name:     "invalid-status",
			instance: &Instance{},
			args: args{
				status: Status("invalid-status"),
			},
			wantSet: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.instance.SetStatus(tt.args.status)
			if tt.wantSet && tt.instance.status != tt.args.status {
				t.Errorf("Instance.SetStatus() not set")
			}

		})
	}
}
func Test_Instance_Status(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     Status
	}{
		{
			name: "valid-Init",
			instance: &Instance{
				status: Init,
			},
			want: Init,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.Status()
			if got != tt.want {
				t.Errorf("Instance.Status() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Instance_SetSessionID(t *testing.T) {
	type args struct {
		sessionID SessionID
	}
	tests := []struct {
		name     string
		instance *Instance
		args     args
		wantErr  bool
		wantSet  bool
	}{
		{
			name:     "invalid-nil,session-id",
			instance: &Instance{},
			args: args{
				sessionID: SessionID{},
			},
			wantErr: true,
			wantSet: false,
		},
		{
			name:     "invalid-nil-id",
			instance: &Instance{},
			args: args{
				sessionID: SessionID{
					SidComplete:     big.NewInt(0).SetBytes(types.Hex2Bytes("e7b9f1350657e7c272508a7cc9451766473bc4c00ece53ea71ac1485e7a7769c")),
					SidSenderPart:   types.Hex2Bytes("fa5a62c4471e12cb2390d3cf31b40a3dc3ba89a5da69203d0064fca83e14fafa"),
					SidReceiverPart: types.Hex2Bytes("4e1f45d6f62c11356dd567d018b40e2e64056c5d0e0b1bdcfa0aa8615e853530"),
					AddrSender:      aliceID.OnChainID,
					AddrReceiver:    bobID.OnChainID,
					NonceSender:     types.Hex2Bytes("2041bb9fe4740252f442c99d8b1f100e8294978e10f10ccfa1409fd270ebb326"),
					NonceReceiver:   types.Hex2Bytes("249613dbc237c4b3a8f4ab3ec52047dfa53d33a4b068b6c98e88abfa4170b18f"),
					Locked:          true,
				},
			},
			wantErr: false,
			wantSet: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.instance.SetSessionID(tt.args.sessionID)
			if tt.wantErr != (gotErr != nil) {
				t.Errorf("Instance.SetSessionID() = %v, wantErr = %v", gotErr, tt.wantErr)
			}

			if tt.wantSet && !tt.instance.sessionID.Equal(tt.args.sessionID) {
				t.Errorf("Instance.SetSessionID() not set")
			}

		})
	}
}
func Test_Instance_SessionID(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     SessionID
	}{
		{
			name: "valid-nil",
			instance: &Instance{
				sessionID: SessionID{},
			},
			want: SessionID{},
		},
		{
			name: "valid-non-nil",
			instance: &Instance{
				sessionID: SessionID{
					SidComplete:     big.NewInt(0).SetBytes(types.Hex2Bytes("e7b9f1350657e7c272508a7cc9451766473bc4c00ece53ea71ac1485e7a7769c")),
					SidSenderPart:   types.Hex2Bytes("fa5a62c4471e12cb2390d3cf31b40a3dc3ba89a5da69203d0064fca83e14fafa"),
					SidReceiverPart: types.Hex2Bytes("4e1f45d6f62c11356dd567d018b40e2e64056c5d0e0b1bdcfa0aa8615e853530"),
					AddrSender:      aliceID.OnChainID,
					AddrReceiver:    bobID.OnChainID,
					NonceSender:     types.Hex2Bytes("2041bb9fe4740252f442c99d8b1f100e8294978e10f10ccfa1409fd270ebb326"),
					NonceReceiver:   types.Hex2Bytes("249613dbc237c4b3a8f4ab3ec52047dfa53d33a4b068b6c98e88abfa4170b18f"),
					Locked:          true,
				},
			},
			want: SessionID{
				SidComplete:     big.NewInt(0).SetBytes(types.Hex2Bytes("e7b9f1350657e7c272508a7cc9451766473bc4c00ece53ea71ac1485e7a7769c")),
				SidSenderPart:   types.Hex2Bytes("fa5a62c4471e12cb2390d3cf31b40a3dc3ba89a5da69203d0064fca83e14fafa"),
				SidReceiverPart: types.Hex2Bytes("4e1f45d6f62c11356dd567d018b40e2e64056c5d0e0b1bdcfa0aa8615e853530"),
				AddrSender:      aliceID.OnChainID,
				AddrReceiver:    bobID.OnChainID,
				NonceSender:     types.Hex2Bytes("2041bb9fe4740252f442c99d8b1f100e8294978e10f10ccfa1409fd270ebb326"),
				NonceReceiver:   types.Hex2Bytes("249613dbc237c4b3a8f4ab3ec52047dfa53d33a4b068b6c98e88abfa4170b18f"),
				Locked:          true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.SessionID()
			if !got.Equal(tt.want) {
				t.Errorf("Instance.SessionID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Instance_SetContractStore(t *testing.T) {
	type args struct {
		contractStore contract.StoreType
	}
	tests := []struct {
		name     string
		instance *Instance
		args     args
		wantSet  bool
	}{
		{
			name:     "valid",
			instance: &Instance{},
			args: args{
				contractStore: contract.StoreType{},
			},
			wantSet: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.instance.SetContractStore(tt.args.contractStore)
			if tt.wantSet && tt.instance.contractStore != tt.args.contractStore {
				t.Errorf("Instance.SetContractStore() not set")
			}

		})
	}
}
func Test_Instance_ContractStore(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     contract.StoreType
	}{
		{
			name: "valid-Sender",
			instance: &Instance{
				contractStore: contract.StoreType{},
			},
			want: contract.StoreType{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.ContractStore()
			if got != tt.want {
				t.Errorf("Instance.ContractStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Instance_SetMSCBaseState(t *testing.T) {
	type args struct {
		mscBaseState MSCBaseStateSigned
	}
	tests := []struct {
		name     string
		instance *Instance
		args     args
		wantErr  bool
		wantSet  bool
	}{
		{
			name: "valid",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				mscBaseState: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      aliceID.OnChainID,
						Sid:             big.NewInt(123456),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
						Version:         big.NewInt(1),
					},
					SignSender:   types.Hex2Bytes("2be8ccd7928d47e38a2bc04377817d2f7364391ef0d39b516f64f4d036b499a00a4f50da3b87d12fed65532e51b401c507713507cea97b9d3e68e07defd531241c"),
					SignReceiver: types.Hex2Bytes("f53a6a46f3d6a94b035b9bc5fb17d0fc762cbf5119e91d3c86515e9db2499d9d5676e333069ebc7b7c9edc2c216cc3dfc15f3b480ba6aa1a5157dd7c28d7a9c61c"),
				},
			},
			wantErr: false,
			wantSet: true,
		},
		{
			name:     "invalid-sender-signature",
			instance: &Instance{},
			args: args{
				mscBaseState: MSCBaseStateSigned{

					MSContractBaseState: MSCBaseState{
						VpcAddress:      aliceID.OnChainID,
						Sid:             big.NewInt(123456),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
						Version:         big.NewInt(1),
					},
					SignSender:   []byte("less-than-65-bytes"),
					SignReceiver: types.Hex2Bytes("f53a6a46f3d6a94b035b9bc5fb17d0fc762cbf5119e91d3c86515e9db2499d9d5676e333069ebc7b7c9edc2c216cc3dfc15f3b480ba6aa1a5157dd7c28d7a9c61c"),
				},
			},
			wantErr: true,
			wantSet: false,
		},
		{
			name: "invalid-sender",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				mscBaseState: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      aliceID.OnChainID,
						Sid:             big.NewInt(123456),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
						Version:         big.NewInt(1),
					},
					//This SignSender is by bob (invalid)
					SignSender:   types.Hex2Bytes("f53a6a46f3d6a94b035b9bc5fb17d0fc762cbf5119e91d3c86515e9db2499d9d5676e333069ebc7b7c9edc2c216cc3dfc15f3b480ba6aa1a5157dd7c28d7a9c61c"),
					SignReceiver: types.Hex2Bytes("f53a6a46f3d6a94b035b9bc5fb17d0fc762cbf5119e91d3c86515e9db2499d9d5676e333069ebc7b7c9edc2c216cc3dfc15f3b480ba6aa1a5157dd7c28d7a9c61c"),
				},
			},
			wantErr: true,
			wantSet: false,
		},
		{
			name: "invalid-receiver-signature",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				mscBaseState: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      aliceID.OnChainID,
						Sid:             big.NewInt(123456),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
						Version:         big.NewInt(1),
					},
					SignSender:   types.Hex2Bytes("2be8ccd7928d47e38a2bc04377817d2f7364391ef0d39b516f64f4d036b499a00a4f50da3b87d12fed65532e51b401c507713507cea97b9d3e68e07defd531241c"),
					SignReceiver: types.Hex2Bytes("less-than-65-bytes"),
				},
			},
			wantErr: true,
			wantSet: false,
		},
		{
			name: "invalid-receiver",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				mscBaseState: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      aliceID.OnChainID,
						Sid:             big.NewInt(123456),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
						Version:         big.NewInt(1),
					},
					SignSender: types.Hex2Bytes("2be8ccd7928d47e38a2bc04377817d2f7364391ef0d39b516f64f4d036b499a00a4f50da3b87d12fed65532e51b401c507713507cea97b9d3e68e07defd531241c"),
					//This SignReceiver is by alice (invalid)
					SignReceiver: types.Hex2Bytes("2be8ccd7928d47e38a2bc04377817d2f7364391ef0d39b516f64f4d036b499a00a4f50da3b87d12fed65532e51b401c507713507cea97b9d3e68e07defd531241c"),
				},
			},
			wantErr: true,
			wantSet: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.instance.SetMSCBaseState(tt.args.mscBaseState)
			if tt.wantErr != (gotErr != nil) {
				t.Errorf("Instance.SetMSCBaseState() = %v, wantErr = %v", gotErr, tt.wantErr)
			}

			if tt.wantSet && !tt.instance.mscBaseState.Equal(tt.args.mscBaseState) {
				t.Errorf("Instance.SetMSCBaseState() not set")
			}

		})
	}
}
func Test_Instance_MSCBaseState(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     MSCBaseStateSigned
	}{
		{
			name: "valid",
			instance: &Instance{
				mscBaseState: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      aliceID.OnChainID,
						Sid:             big.NewInt(123456),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
						Version:         big.NewInt(1),
					},
					SignSender:   []byte("sign-sender"),
					SignReceiver: []byte("sign-receiver"),
				},
			},
			want: MSCBaseStateSigned{
				MSContractBaseState: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
				SignSender:   []byte("sign-sender"),
				SignReceiver: []byte("sign-receiver"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.MscBaseState()
			if !got.Equal(tt.want) {
				t.Errorf("Instance.MscBaseState() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_Instance_SetCurrentVPCState(t *testing.T) {
	type args struct {
		vpcState VPCStateSigned
	}
	tests := []struct {
		name     string
		instance *Instance
		args     args
	}{
		{
			name: "valid-state",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				},
			},
		},
		{
			name: "invalid-state",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.instance.SetCurrentVPCState(tt.args.vpcState)

			currentVPCStateIndex := len(tt.instance.vpcStatesList) - 1
			if currentVPCStateIndex >= 0 {
				if !tt.instance.vpcStatesList[currentVPCStateIndex].Equal(tt.args.vpcState) {
					t.Errorf("Instance.SetCurrentVPCState() not set")
				}
			}

		})
	}
}
func Test_Instance_ValidateIncomingState(t *testing.T) {
	type args struct {
		vpcState VPCStateSigned
	}
	tests := []struct {
		name      string
		instance  *Instance
		args      args
		wantValid bool
	}{
		{
			name: "valid-1",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				},
			},
			wantValid: true,
		},
		{
			name: "valid-2",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Receiver,
				vpcStatesList: []VPCStateSigned{{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(0),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				}},
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				},
			},
			wantValid: true,
		},
		{
			name: "invalid-version",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
				vpcStatesList: []VPCStateSigned{{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(2),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				}},
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				},
			},
			wantValid: false,
		},
		{
			name: "invalid-peer-signature",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Receiver,
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender: []byte("less-than-65-bytes"),
				},
			},
			wantValid: false,
		},
		{
			name: "invalid-peer-signature-2",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					//This SignReceiver is by alice (invalid)
					SignReceiver: types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
				},
			},
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			vpcStateInput := tt.args.vpcState

			isValid, reason := tt.instance.ValidateIncomingState(tt.args.vpcState)
			if tt.wantValid != isValid {
				t.Errorf("Instance.ValidateIncomingState() = %v, reason = %v wantValid = %v", isValid, reason, tt.wantValid)
			}

			//Check if input was modified
			if !vpcStateInput.Equal(tt.args.vpcState) {
				t.Errorf("Instance.ValidateIncomingState() modified input value. got = %v,  want = %v", tt.args.vpcState, vpcStateInput)
			}

		})
	}
}

func Test_Instance_ValidateFullState(t *testing.T) {
	type args struct {
		vpcState VPCStateSigned
	}
	tests := []struct {
		name      string
		instance  *Instance
		args      args
		wantValid bool
	}{
		{
			name: "valid-1",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				},
			},
			wantValid: true,
		},
		{
			name: "valid-2",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
				vpcStatesList: []VPCStateSigned{{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(0),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				}},
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				},
			},
			wantValid: true,
		},
		{
			name: "invalid-version",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
				vpcStatesList: []VPCStateSigned{{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(2),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				}},
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				},
			},
			wantValid: false,
		},
		{
			name:     "invalid-sender-signature",
			instance: &Instance{},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   []byte("less-than-65-bytes"),
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				},
			},
			wantValid: false,
		},
		{
			name: "invalid-sender",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					//This SignSender is by bob (invalid)
					SignSender:   types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
					SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
				},
			},
			wantValid: false,
		},
		{
			name: "invalid-receiver-signature",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
					SignReceiver: types.Hex2Bytes("less-than-65-bytes"),
				},
			},
			wantValid: false,
		},
		{
			name: "invalid-receiver",
			instance: &Instance{
				selfID:      aliceID,
				peerID:      bobID,
				roleChannel: Sender,
			},
			args: args{
				vpcState: VPCStateSigned{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender: types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
					//This SignReceiver is by alice (invalid)
					SignReceiver: types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
				},
			},
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			vpcStateInput := tt.args.vpcState

			isValid, reason := tt.instance.ValidateFullState(tt.args.vpcState)
			if tt.wantValid != isValid {
				t.Errorf("Instance.ValidateFullState() = %v, reason = %v wantValid = %v", isValid, reason, tt.wantValid)
			}

			//Check if input was modified
			if !vpcStateInput.Equal(tt.args.vpcState) {
				t.Errorf("Instance.ValidateFullState() modified input value. got = %v,  want = %v", tt.args.vpcState, vpcStateInput)
			}

		})
	}
}

func Test_Instance_CurrentVpcState(t *testing.T) {
	tests := []struct {
		name     string
		instance *Instance
		want     VPCStateSigned
	}{
		{
			name: "valid",
			instance: &Instance{
				vpcStatesList: []VPCStateSigned{{
					VPCState: VPCState{
						ID:              []byte("sample-id"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   []byte("sign-sender"),
					SignReceiver: []byte("sign-receiver"),
				}}},
			want: VPCStateSigned{
				VPCState: VPCState{
					ID:              []byte("sample-id"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				SignSender:   []byte("sign-sender"),
				SignReceiver: []byte("sign-receiver"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instance.CurrentVpcState()
			if !got.Equal(tt.want) {
				t.Errorf("Instance.MscBaseState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NewSession(t *testing.T) {
	type args struct {
		selfID      identity.OffChainID
		adapterType AdapterType
		maxConn     uint32
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		shutdown bool
	}{
		{
			name: "valid",
			args: args{
				selfID: identity.OffChainID{
					OnChainID:        bobID.OnChainID,
					ListenerIPAddr:   bobID.ListenerIPAddr,
					ListenerEndpoint: "/listen-new-session-test-1",
				},
				adapterType: WebSocket,
				maxConn:     100,
			},
			wantErr:  false,
			shutdown: true,
		},
		{
			name: "invalid-ip-local",
			args: args{
				selfID: identity.OffChainID{
					OnChainID:        bobID.OnChainID,
					ListenerIPAddr:   invalidOffchainAddr,
					ListenerEndpoint: "/listen-new-session-test-2",
				},
				adapterType: WebSocket,
				maxConn:     100,
			},
			wantErr:  true,
			shutdown: false,
		},
		{
			name: "invalid-adapter-type",
			args: args{
				selfID: identity.OffChainID{
					OnChainID:        bobID.OnChainID,
					ListenerIPAddr:   bobID.ListenerIPAddr,
					ListenerEndpoint: "/listen-new-session-test-1",
				},
				adapterType: AdapterType("invalid-adapter-type"),
				maxConn:     100,
			},
			wantErr:  true,
			shutdown: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, listener, err := NewSession(tt.args.selfID, tt.args.adapterType, tt.args.maxConn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewSession() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.shutdown {
				if listener != nil {
					_ = listener.Shutdown(context.Background())
				} else {
					t.Errorf("NewSession() want listener, got nil")
				}
			}

		})
	}
}
