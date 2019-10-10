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
	"bytes"
	"math/big"
	"reflect"
	"testing"

	"github.com/direct-state-transfer/dst-go/ethereum/keystore"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
)

var testSessionID = SessionID{
	SidComplete:     big.NewInt(0).SetBytes(types.Hex2Bytes("e7b9f1350657e7c272508a7cc9451766473bc4c00ece53ea71ac1485e7a7769c")),
	SidSenderPart:   types.Hex2Bytes("fa5a62c4471e12cb2390d3cf31b40a3dc3ba89a5da69203d0064fca83e14fafa"),
	SidReceiverPart: types.Hex2Bytes("4e1f45d6f62c11356dd567d018b40e2e64056c5d0e0b1bdcfa0aa8615e853530"),
	AddrSender:      aliceID.OnChainID,
	AddrReceiver:    bobID.OnChainID,
	NonceSender:     types.Hex2Bytes("2041bb9fe4740252f442c99d8b1f100e8294978e10f10ccfa1409fd270ebb326"),
	NonceReceiver:   types.Hex2Bytes("249613dbc237c4b3a8f4ab3ec52047dfa53d33a4b068b6c98e88abfa4170b18f"),
	Locked:          true,
}

func Test_NewSessionID(t *testing.T) {
	addrSender := aliceID.OnChainID
	addrReceiver := bobID.OnChainID
	wantSessionID := SessionID{
		AddrSender:   aliceID.OnChainID,
		AddrReceiver: bobID.OnChainID,
		Locked:       false,
	}

	gotSessionID := NewSessionID(addrSender, addrReceiver)
	if !reflect.DeepEqual(gotSessionID, wantSessionID) {
		t.Errorf("NewSessionID() = %v, want %v", gotSessionID, wantSessionID)
	}
}

func Test_SessionID_Equal(t *testing.T) {
	type args struct {
		a, b SessionID
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid_nil_equal",
			args: args{
				a: SessionID{
					AddrSender:    types.Address{},
					NonceSender:   nil,
					SidSenderPart: nil,

					SidReceiverPart: nil,
					AddrReceiver:    types.Address{},
					NonceReceiver:   nil,

					SidComplete: nil,
				},
				b: SessionID{
					AddrSender:    types.Address{},
					NonceSender:   nil,
					SidSenderPart: nil,

					SidReceiverPart: nil,
					AddrReceiver:    types.Address{},
					NonceReceiver:   nil,

					SidComplete: nil,
				},
			},
			want: true,
		},
		{
			name: "valid_non_nil_equal",
			args: args{
				a: SessionID{
					AddrSender:    aliceID.OnChainID,
					NonceSender:   types.Hex2Bytes("ab12cd34"),
					SidSenderPart: types.Hex2Bytes("ab12cd34"),

					SidReceiverPart: types.Hex2Bytes("ab12cd34"),
					AddrReceiver:    bobID.OnChainID,
					NonceReceiver:   types.Hex2Bytes("ab12cd34"),

					SidComplete: big.NewInt(121),
				},
				b: SessionID{
					AddrSender:    aliceID.OnChainID,
					NonceSender:   types.Hex2Bytes("ab12cd34"),
					SidSenderPart: types.Hex2Bytes("ab12cd34"),

					SidReceiverPart: types.Hex2Bytes("ab12cd34"),
					AddrReceiver:    bobID.OnChainID,
					NonceReceiver:   types.Hex2Bytes("ab12cd34"),

					SidComplete: big.NewInt(121),
				},
			},
			want: true,
		},
		{
			name: "valid_non_nil_notequal_sidComplete",
			args: args{
				a: SessionID{
					AddrSender:    aliceID.OnChainID,
					NonceSender:   types.Hex2Bytes("ab12cd34"),
					SidSenderPart: types.Hex2Bytes("ab12cd34"),

					SidReceiverPart: types.Hex2Bytes("ab12cd34"),
					AddrReceiver:    bobID.OnChainID,
					NonceReceiver:   types.Hex2Bytes("ab12cd34"),

					SidComplete: big.NewInt(121),
				},
				b: SessionID{
					AddrSender:    aliceID.OnChainID,
					NonceSender:   types.Hex2Bytes("ab12cd34"),
					SidSenderPart: types.Hex2Bytes("ab12cd34"),

					SidReceiverPart: types.Hex2Bytes("ab12cd34"),
					AddrReceiver:    bobID.OnChainID,
					NonceReceiver:   types.Hex2Bytes("ab12cd34"),

					SidComplete: big.NewInt(321),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.a.Equal(tt.args.b); got != tt.want {
				t.Errorf("SessionID.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_SessionID_EqualSender(t *testing.T) {
	type args struct {
		a, b SessionID
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid_nil",
			args: args{
				a: SessionID{
					AddrSender:    types.Address{},
					NonceSender:   nil,
					SidSenderPart: nil,
				},
				b: SessionID{
					AddrSender:    types.Address{},
					NonceSender:   nil,
					SidSenderPart: nil,
				},
			},
			want: true,
		},
		{
			name: "valid_non_nil_notequal_addr",
			args: args{
				a: SessionID{
					AddrSender:    aliceID.OnChainID,
					NonceSender:   types.Hex2Bytes("ab12cd34"),
					SidSenderPart: types.Hex2Bytes("ab12cd34"),
				},
				b: SessionID{
					AddrSender:    bobID.OnChainID,
					NonceSender:   types.Hex2Bytes("ab12cd34"),
					SidSenderPart: types.Hex2Bytes("ab12cd34"),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.a.EqualSender(tt.args.b); got != tt.want {
				t.Errorf("SessionID.EqualSender() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_SessionID_EqualReceiver(t *testing.T) {
	type args struct {
		a, b SessionID
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid_nil",
			args: args{
				a: SessionID{
					AddrReceiver:    types.Address{},
					NonceReceiver:   nil,
					SidReceiverPart: nil,
				},
				b: SessionID{
					AddrReceiver:    types.Address{},
					NonceReceiver:   nil,
					SidReceiverPart: nil,
				},
			},
			want: true,
		},
		{
			name: "valid_non_nil_notequal_addr",
			args: args{
				a: SessionID{
					AddrReceiver:    aliceID.OnChainID,
					NonceReceiver:   types.Hex2Bytes("ab12cd34"),
					SidReceiverPart: types.Hex2Bytes("ab12cd34"),
				},
				b: SessionID{
					AddrReceiver:    bobID.OnChainID,
					NonceReceiver:   types.Hex2Bytes("ab12cd34"),
					SidReceiverPart: types.Hex2Bytes("ab12cd34"),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.a.EqualReceiver(tt.args.b); got != tt.want {
				t.Errorf("SessionID.EqualReceiver() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_SessionID_GenerateSenderPart(t *testing.T) {
	type args struct {
		AddrSender types.Address
		sid        SessionID
	}
	tests := []struct {
		name          string
		args          args
		wantSessionID SessionID
		wantErr       bool
	}{
		{
			name: "valid",
			args: args{
				sid:        SessionID{},
				AddrSender: aliceID.OnChainID,
			},
			wantErr: false,
		},
		{
			name: "locked",
			args: args{
				sid:        SessionID{Locked: true},
				AddrSender: aliceID.OnChainID,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.sid.GenerateSenderPart(tt.args.AddrSender)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SessionID.GenerateSenderPart() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr {
				t.Logf("SessionID.GenerateSenderPart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			//cannot be hardcoded, because nonce will change during each run
			wantSessionIDSenderPart := keystore.Keccak256(append(tt.args.sid.NonceSender, tt.args.sid.AddrSender.Bytes()...))

			if !bytes.Equal(tt.args.sid.SidSenderPart, wantSessionIDSenderPart) {
				t.Errorf("SessionID.GenerateSenderPart() SidSenderPart = %x, want %x", tt.args.sid.SidSenderPart, tt.wantSessionID.SidSenderPart)
			}
			if !bytes.Equal(tt.args.sid.AddrSender.Bytes(), tt.args.AddrSender.Bytes()) {
				t.Errorf("SessionID.GenerateSenderPart() AddrSender = %x, want %x", tt.args.sid.AddrSender.Bytes(), tt.wantSessionID.AddrSender.Bytes())
			}

		})
	}
}

func Test_SessionID_GenerateReceiverPart(t *testing.T) {
	type args struct {
		AddrReceiver types.Address
		sid          SessionID
	}
	tests := []struct {
		name          string
		args          args
		wantSessionID SessionID
		wantErr       bool
	}{
		{
			name: "valid",
			args: args{
				sid:          SessionID{},
				AddrReceiver: bobID.OnChainID,
			},
			wantErr: false,
		},
		{
			name: "locked",
			args: args{
				sid:          SessionID{Locked: true},
				AddrReceiver: bobID.OnChainID,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.sid.GenerateReceiverPart(tt.args.AddrReceiver)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SessionID.GenerateReceiverPart() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr {
				t.Logf("SessionID.GenerateReceiverPart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			//cannot be hardcoded, because nonce will change during each run
			wantSessionIDReceiverPart := keystore.Keccak256(append(tt.args.sid.NonceReceiver, tt.args.sid.AddrReceiver.Bytes()...))

			if !bytes.Equal(tt.args.sid.SidReceiverPart, wantSessionIDReceiverPart) {
				t.Errorf("SessionID.GenerateReceiverPart() SidReceiverPart = %x, want %x", tt.args.sid.SidReceiverPart, tt.wantSessionID.SidReceiverPart)
			}
			if !bytes.Equal(tt.args.sid.AddrReceiver.Bytes(), tt.args.AddrReceiver.Bytes()) {
				t.Errorf("SessionID.GenerateReceiverPart() AddrReceiver = %x, want %x", tt.args.sid.AddrReceiver.Bytes(), tt.wantSessionID.AddrReceiver.Bytes())
			}

		})
	}
}

func Test_SessionID_GenerateCompleteSid(t *testing.T) {
	tests := []struct {
		name    string
		sid     SessionID
		wantErr bool
	}{
		{
			name: "err_locked_true",
			sid: SessionID{
				Locked: true,
			},
			wantErr: true,
		}, {
			name: "err_senderPart_nil",
			sid: SessionID{
				SidSenderPart: nil,
				Locked:        false,
			},
			wantErr: true,
		},
		{
			name: "err_receiverPart_nil",
			sid: SessionID{
				SidSenderPart:   types.Hex2Bytes("fa5a62c4471e12cb2390d3cf31b40a3dc3ba89a5da69203d0064fca83e14fafa"),
				SidReceiverPart: nil,
				Locked:          false,
			},
			wantErr: true,
		},
		{
			name: "valid",
			sid: SessionID{
				SidSenderPart:   types.Hex2Bytes("fa5a62c4471e12cb2390d3cf31b40a3dc3ba89a5da69203d0064fca83e14fafa"),
				SidReceiverPart: types.Hex2Bytes("4e1f45d6f62c11356dd567d018b40e2e64056c5d0e0b1bdcfa0aa8615e853530"),
				AddrSender:      aliceID.OnChainID,
				AddrReceiver:    bobID.OnChainID,
				NonceSender:     types.Hex2Bytes("2041bb9fe4740252f442c99d8b1f100e8294978e10f10ccfa1409fd270ebb326"),
				NonceReceiver:   types.Hex2Bytes("249613dbc237c4b3a8f4ab3ec52047dfa53d33a4b068b6c98e88abfa4170b18f"),
				Locked:          false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sid.GenerateCompleteSid()
			if (err != nil) != tt.wantErr {
				t.Fatalf("SessionID.GenerateCompleteSid() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr {
				t.Logf("SessionID.GenerateCompleteSid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_SessionID_Validate(t *testing.T) {

	tests := []struct {
		name      string
		sid       SessionID
		wantValid bool
		wantErr   bool
	}{
		{
			name: "valid",
			sid: SessionID{
				SidComplete:     big.NewInt(0).SetBytes(types.Hex2Bytes("e7b9f1350657e7c272508a7cc9451766473bc4c00ece53ea71ac1485e7a7769c")),
				SidSenderPart:   types.Hex2Bytes("fa5a62c4471e12cb2390d3cf31b40a3dc3ba89a5da69203d0064fca83e14fafa"),
				SidReceiverPart: types.Hex2Bytes("4e1f45d6f62c11356dd567d018b40e2e64056c5d0e0b1bdcfa0aa8615e853530"),
				AddrSender:      aliceID.OnChainID,
				AddrReceiver:    bobID.OnChainID,
				NonceSender:     types.Hex2Bytes("2041bb9fe4740252f442c99d8b1f100e8294978e10f10ccfa1409fd270ebb326"),
				NonceReceiver:   types.Hex2Bytes("249613dbc237c4b3a8f4ab3ec52047dfa53d33a4b068b6c98e88abfa4170b18f"),
				Locked:          true,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "locked",
			sid: SessionID{
				Locked: false,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "invalid_senderPartNil",
			sid: SessionID{
				SidSenderPart: nil,
				Locked:        true,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "invalid_receiverPartNil",
			sid: SessionID{
				SidSenderPart:   types.Hex2Bytes("fa5a62c4471e12cb2390d3cf31b40a3dc3ba89a5da69203d0064fca83e14fafa"),
				SidReceiverPart: nil,
				Locked:          true,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			//SID SenderPart was created for senderid - aliceID
			name: "invalid_senderPart_wrongSenderID",
			sid: SessionID{
				SidComplete:     big.NewInt(0).SetBytes(types.Hex2Bytes("e7b9f1350657e7c272508a7cc9451766473bc4c00ece53ea71ac1485e7a7769c")),
				SidSenderPart:   types.Hex2Bytes("fa5a62c4471e12cb2390d3cf31b40a3dc3ba89a5da69203d0064fca83e14fafa"),
				SidReceiverPart: types.Hex2Bytes("4e1f45d6f62c11356dd567d018b40e2e64056c5d0e0b1bdcfa0aa8615e853530"),
				AddrSender:      bobID.OnChainID, //Sender is actually aliceID
				AddrReceiver:    bobID.OnChainID,
				NonceSender:     types.Hex2Bytes("2041bb9fe4740252f442c99d8b1f100e8294978e10f10ccfa1409fd270ebb326"),
				NonceReceiver:   types.Hex2Bytes("249613dbc237c4b3a8f4ab3ec52047dfa53d33a4b068b6c98e88abfa4170b18f"),
				Locked:          true,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			//SID ReceiverPart was created for receiverid - bobID
			name: "invalid_receiverPart_wrongReceiverID",
			sid: SessionID{
				SidComplete:     big.NewInt(0).SetBytes(types.Hex2Bytes("e7b9f1350657e7c272508a7cc9451766473bc4c00ece53ea71ac1485e7a7769c")),
				SidSenderPart:   types.Hex2Bytes("fa5a62c4471e12cb2390d3cf31b40a3dc3ba89a5da69203d0064fca83e14fafa"),
				SidReceiverPart: types.Hex2Bytes("4e1f45d6f62c11356dd567d018b40e2e64056c5d0e0b1bdcfa0aa8615e853530"),
				AddrSender:      aliceID.OnChainID,
				AddrReceiver:    aliceID.OnChainID, //ReceiverID is actually bobID
				NonceSender:     types.Hex2Bytes("2041bb9fe4740252f442c99d8b1f100e8294978e10f10ccfa1409fd270ebb326"),
				NonceReceiver:   types.Hex2Bytes("249613dbc237c4b3a8f4ab3ec52047dfa53d33a4b068b6c98e88abfa4170b18f"),
				Locked:          true,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "invalid_sidcomplete",
			sid: SessionID{
				// Correct SidComplete is big.NewInt(0).SetBytes(types.Hex2Bytes("e7b9f1350657e7c272508a7cc9451766473bc4c00ece53ea71ac1485e7a7769c")),
				SidComplete:     big.NewInt(100),
				SidSenderPart:   types.Hex2Bytes("fa5a62c4471e12cb2390d3cf31b40a3dc3ba89a5da69203d0064fca83e14fafa"),
				SidReceiverPart: types.Hex2Bytes("4e1f45d6f62c11356dd567d018b40e2e64056c5d0e0b1bdcfa0aa8615e853530"),
				AddrSender:      aliceID.OnChainID,
				AddrReceiver:    bobID.OnChainID,
				NonceSender:     types.Hex2Bytes("2041bb9fe4740252f442c99d8b1f100e8294978e10f10ccfa1409fd270ebb326"),
				NonceReceiver:   types.Hex2Bytes("249613dbc237c4b3a8f4ab3ec52047dfa53d33a4b068b6c98e88abfa4170b18f"),
				Locked:          true,
			},
			wantValid: false,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, err := tt.sid.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("SessionID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotValid != tt.wantValid {
				t.Errorf("SessionID.Validate() = %v, want %v", gotValid, tt.wantValid)
			}
		})
	}
}

func Test_SessionID_SoliditySHA3(t *testing.T) {
	tests := []struct {
		name string
		sid  SessionID
		want []byte
	}{
		{
			name: "valid",
			sid: SessionID{
				SidComplete:     testSessionID.SidComplete,
				SidSenderPart:   testSessionID.SidSenderPart,
				SidReceiverPart: testSessionID.SidReceiverPart,
			},
			want: types.Hex2Bytes("1a45efad199a6ca82310df0f4fddff37a7b29a2223afec4a5b024ceea20fa3cd"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sid.SoliditySHA3()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SessionID.SoliditySHA3() = %x, want %x", got, tt.want)
			}
		})
	}
}

func Test_MSCBaseState_String(t *testing.T) {

	tests := []struct {
		name         string
		mscBaseState MSCBaseState
		want         string
	}{
		{
			name: "valid_empty",
			mscBaseState: MSCBaseState{
				VpcAddress:      types.Address{},
				Sid:             &big.Int{},
				BlockedSender:   &big.Int{},
				BlockedReceiver: &big.Int{},
				Version:         &big.Int{},
			},
			want: "{VpcAddress:0x0000000000000000000000000000000000000000 Sid:0x BlockedSender:0 BlockedReceiver:0 Version:0}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mscBaseState.String()
			if got != tt.want {
				t.Errorf("MSCBaseState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_MSCBaseState_Equal(t *testing.T) {
	type args struct {
		a, b MSCBaseState
	}
	tests := []struct {
		name       string
		args       args
		wantResult bool
	}{
		{
			name: "equal_states",
			args: args{
				a: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
				b: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
			},
			wantResult: true,
		},
		{
			name: "unequal_vpcAddress",
			args: args{
				a: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
				b: MSCBaseState{
					VpcAddress:      bobID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
			},
			wantResult: false,
		},
		{
			name: "unequal_sid",
			args: args{
				a: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
				b: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(654321),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
			},
			wantResult: false,
		},
		{
			name: "unequal_blockedSender",
			args: args{
				a: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
				b: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(30),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
			},
			wantResult: false,
		},
		{
			name: "unequal_blockedReceiver",
			args: args{
				a: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
				b: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(30),
					Version:         big.NewInt(1),
				},
			},
			wantResult: false,
		},
		{
			name: "unequal_version",
			args: args{
				a: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
				b: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(5),
				},
			},
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := tt.args.a.Equal(tt.args.b)
			if gotResult != tt.wantResult {
				t.Errorf("MSCBaseState.Equal() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
func Test_MSCBaseState_SoliditySHA3(t *testing.T) {
	tests := []struct {
		name         string
		mscBaseState MSCBaseState
		want         []byte
	}{
		{
			name: "valid_empty",
			mscBaseState: MSCBaseState{
				VpcAddress:      types.Address{},
				Sid:             &big.Int{},
				BlockedSender:   &big.Int{},
				BlockedReceiver: &big.Int{},
				Version:         &big.Int{},
			},
			want: types.Hex2Bytes("df2dce66784a9bb231911f54817836d5568a99bb60aeb5bed0eadf2f3a6585b8"),
		},
		{
			name: "valid_non_empty",
			mscBaseState: MSCBaseState{
				VpcAddress:      aliceID.OnChainID,
				Sid:             big.NewInt(123456),
				BlockedSender:   big.NewInt(10),
				BlockedReceiver: big.NewInt(20),
				Version:         big.NewInt(1),
			},
			want: types.Hex2Bytes("9cc6fb30c9f49ae7f9aa717b6fe9d1563f9ce8652a491e34db44576e110f4c09"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mscBaseState.SoliditySHA3()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MSCBaseState.SoliditySHA3() = %x, want %x", got, tt.want)
			}
		})
	}
}
func Test_MSCBaseStateSigned_String(t *testing.T) {
	tests := []struct {
		name               string
		mscBaseStateSigned MSCBaseStateSigned
		want               string
	}{
		{
			name: "valid_empty",
			mscBaseStateSigned: MSCBaseStateSigned{
				MSContractBaseState: MSCBaseState{
					VpcAddress:      types.Address{},
					Sid:             &big.Int{},
					BlockedSender:   &big.Int{},
					BlockedReceiver: &big.Int{},
					Version:         &big.Int{},
				},
				SignSender:   nil,
				SignReceiver: nil,
			},
			want: "{MSCBaseState:{VpcAddress:0x0000000000000000000000000000000000000000 Sid:0x BlockedSender:0 BlockedReceiver:0 Version:0} SignSender:0x SignReceiver:0x}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mscBaseStateSigned.String()
			if got != tt.want {
				t.Errorf("MSCBaseStateSigned.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_MSCBaseStateSigned_Equal(t *testing.T) {
	type args struct {
		a, b MSCBaseStateSigned
	}
	tests := []struct {
		name       string
		args       args
		wantResult bool
	}{
		{
			name: "equal_states",
			args: args{
				a: MSCBaseStateSigned{
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
				b: MSCBaseStateSigned{
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
			wantResult: true,
		},
		{
			name: "unequal_mscBaseState_sid",
			args: args{
				a: MSCBaseStateSigned{
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
				b: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      aliceID.OnChainID,
						Sid:             big.NewInt(654321),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
						Version:         big.NewInt(1),
					},
					SignSender:   []byte("sign-sender"),
					SignReceiver: []byte("sign-receiver"),
				},
			},
			wantResult: false,
		},
		{
			name: "unequal_signsender",
			args: args{
				a: MSCBaseStateSigned{
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
				b: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      aliceID.OnChainID,
						Sid:             big.NewInt(123456),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
						Version:         big.NewInt(1),
					},
					SignSender:   []byte("sign-sender-mismatch"),
					SignReceiver: []byte("sign-receiver"),
				},
			},
			wantResult: false,
		},
		{
			name: "unequal_signreceiver",
			args: args{
				a: MSCBaseStateSigned{
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
				b: MSCBaseStateSigned{
					MSContractBaseState: MSCBaseState{
						VpcAddress:      aliceID.OnChainID,
						Sid:             big.NewInt(123456),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
						Version:         big.NewInt(1),
					},
					SignSender:   []byte("sign-sender"),
					SignReceiver: []byte("sign-receiver-mismatch"),
				},
			},
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := tt.args.a.Equal(tt.args.b)
			if gotResult != tt.wantResult {
				t.Errorf("MSCBaseStateSigned.Equal() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
func Test_MSCBaseStateSigned_AddSign(t *testing.T) {
	type args struct {
		idWithCreds identity.OffChainID
		role        Role
	}
	tests := []struct {
		name               string
		mscBaseStateSigned MSCBaseStateSigned
		args               args
		wantSignature      []byte
		wantErr            bool
	}{
		{
			name: "valid_addSignSender",
			mscBaseStateSigned: MSCBaseStateSigned{
				MSContractBaseState: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				}},
			args: args{
				idWithCreds: identity.OffChainID{
					OnChainID:        aliceID.OnChainID,
					ListenerIPAddr:   aliceID.ListenerIPAddr,
					ListenerEndpoint: aliceID.ListenerEndpoint,
					KeyStore:         testKeyStore,
					Password:         alicePassword,
				},
				role: Sender,
			},
			wantSignature: types.Hex2Bytes("2be8ccd7928d47e38a2bc04377817d2f7364391ef0d39b516f64f4d036b499a00a4f50da3b87d12fed65532e51b401c507713507cea97b9d3e68e07defd531241c"),
			wantErr:       false,
		},
		{
			name: "valid_addSignReceiver",
			mscBaseStateSigned: MSCBaseStateSigned{
				MSContractBaseState: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				}},
			args: args{
				idWithCreds: identity.OffChainID{
					OnChainID:        bobID.OnChainID,
					ListenerIPAddr:   bobID.ListenerIPAddr,
					ListenerEndpoint: bobID.ListenerEndpoint,
					KeyStore:         testKeyStore,
					Password:         bobPassword,
				},
				role: Receiver,
			},
			wantSignature: types.Hex2Bytes("f53a6a46f3d6a94b035b9bc5fb17d0fc762cbf5119e91d3c86515e9db2499d9d5676e333069ebc7b7c9edc2c216cc3dfc15f3b480ba6aa1a5157dd7c28d7a9c61c"),
			wantErr:       false,
		},
		{
			name: "invalid_signError_wrongPassword",
			mscBaseStateSigned: MSCBaseStateSigned{
				MSContractBaseState: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				}},
			args: args{
				idWithCreds: identity.OffChainID{
					OnChainID:        aliceID.OnChainID,
					ListenerIPAddr:   aliceID.ListenerIPAddr,
					ListenerEndpoint: aliceID.ListenerEndpoint,
					KeyStore:         testKeyStore,
					Password:         alicePassword + "xyz",
				},
				role: Sender,
			},
			wantSignature: types.Hex2Bytes("2be8ccd7928d47e38a2bc04377817d2f7364391ef0d39b516f64f4d036b499a00a4f50da3b87d12fed65532e51b401c507713507cea97b9d3e68e07defd531241c"),
			wantErr:       true,
		},
		{
			name: "invalid_role",
			mscBaseStateSigned: MSCBaseStateSigned{
				MSContractBaseState: MSCBaseState{}},
			args: args{
				idWithCreds: identity.OffChainID{},
				role:        Role("invalid-role"),
			},
			wantSignature: types.Hex2Bytes("2be8ccd7928d47e38a2bc04377817d2f7364391ef0d39b516f64f4d036b499a00a4f50da3b87d12fed65532e51b401c507713507cea97b9d3e68e07defd531241c"),
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mscBaseStateSigned.AddSign(tt.args.idWithCreds, tt.args.role)
			if (err != nil) != tt.wantErr {
				t.Fatalf("MSCBaseStateSigned.AddSign() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				t.Logf("MSCBaseStateSigned.AddSign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.args.role == Sender {
				if !bytes.Equal(tt.mscBaseStateSigned.SignSender, tt.wantSignature) {
					t.Errorf("MSCBaseStateSigned.AddSign() signature = %x, want %x", tt.mscBaseStateSigned.SignSender, tt.wantSignature)
				}
			} else if tt.args.role == Receiver {
				if !bytes.Equal(tt.mscBaseStateSigned.SignReceiver, tt.wantSignature) {
					t.Errorf("MSCBaseStateSigned.AddSign() signature = %x, want %x", tt.mscBaseStateSigned.SignReceiver, tt.wantSignature)
				}
			}

		})
	}
}

func Test_MSCBaseStateSigned_VerifySign(t *testing.T) {
	type args struct {
		id   identity.OffChainID
		role Role
	}
	tests := []struct {
		name               string
		mscBaseStateSigned MSCBaseStateSigned
		args               args
		wantIsValid        bool
		wantErr            bool
	}{
		{
			name: "valid_roleSender",
			mscBaseStateSigned: MSCBaseStateSigned{
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
			args: args{
				id: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
				},
				role: Sender,
			},
			wantIsValid: true,
			wantErr:     false,
		},
		{
			name: "valid_roleReceiver",
			mscBaseStateSigned: MSCBaseStateSigned{
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
			args: args{
				id: identity.OffChainID{
					OnChainID: bobID.OnChainID,
				},
				role: Receiver,
			},
			wantIsValid: true,
			wantErr:     false,
		},
		{
			name: "valid_roleReceiver_wrongID",
			mscBaseStateSigned: MSCBaseStateSigned{
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
			args: args{
				id: identity.OffChainID{
					OnChainID: aliceID.OnChainID, //Actual signer id is bobID.OnChainID
				},
				role: Receiver,
			},
			wantIsValid: false,
			wantErr:     false,
		},
		{
			name: "invalid_SignSender",
			mscBaseStateSigned: MSCBaseStateSigned{
				MSContractBaseState: MSCBaseState{
					VpcAddress:      aliceID.OnChainID,
					Sid:             big.NewInt(123456),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
					Version:         big.NewInt(1),
				},
				//For valid signature, Last byte should be 0x1B (27) or 0x1C (28), here it is 01
				SignSender: types.Hex2Bytes("2be8ccd7928d47e38a2bc04377817d2f7364391ef0d39b516f64f4d036b499a00a4f50da3b87d12fed65532e51b401c507713507cea97b9d3e68e07defd5312401"),
			},
			args: args{
				id: identity.OffChainID{
					OnChainID: aliceID.OnChainID, //Actual signer id is bobID.OnChainID
				},
				role: Sender,
			},
			wantIsValid: false,
			wantErr:     true,
		},
		{
			name:               "invalid_role",
			mscBaseStateSigned: MSCBaseStateSigned{},
			args: args{
				role: Role("invalid-role"),
			},
			wantIsValid: false,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsValid, err := tt.mscBaseStateSigned.VerifySign(tt.args.id, tt.args.role)
			if (err != nil) != tt.wantErr {
				t.Fatalf("MSCBaseStateSigned.VerifySign() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotIsValid != tt.wantIsValid {
				t.Errorf("MSCBaseStateSigned.VerifySign() = %v, want %v", gotIsValid, tt.wantIsValid)
			}
		})
	}
}
func Test_VPCStateID_SoliditySHA3(t *testing.T) {

	tests := []struct {
		name       string
		vpcStateID VPCStateID
		want       []byte
	}{
		{
			name: "valid_nil",
			vpcStateID: VPCStateID{
				AddSender:    types.Address{},
				AddrReceiver: types.Address{},
				SID:          &big.Int{},
			},
			want: types.Hex2Bytes("3cac317908c699fe873a7f6ee4e8cd63fbe9918b2315c97be91585590168e301"),
		},
		{
			name: "valid_non_empty",
			vpcStateID: VPCStateID{
				AddSender:    aliceID.OnChainID,
				AddrReceiver: bobID.OnChainID,
				SID:          big.NewInt(10),
			},
			want: types.Hex2Bytes("98c16a46f27453abd094f7a73c77a755840bc1ec2e9047a26d5fa5a2ac9b1b40"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.vpcStateID.SoliditySHA3()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VPCStateID.SoliditySHA3() = 0x%x, want 0x%x", got, tt.want)
			}
		})
	}
}
func Test_VPCState_String(t *testing.T) {
	tests := []struct {
		name     string
		vpcState VPCState
		want     string
	}{
		{
			name: "valid_empty",
			vpcState: VPCState{
				ID:              nil,
				Version:         &big.Int{},
				BlockedSender:   &big.Int{},
				BlockedReceiver: &big.Int{},
			},
			want: "{Id:0x Version:0 BlockedSender:0 BlockedReceiver:0}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.vpcState.String()
			if got != tt.want {
				t.Errorf("VPCState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_VPCState_SoliditySHA3(t *testing.T) {
	tests := []struct {
		name     string
		vpcState VPCState
		want     []byte
	}{
		{
			name: "valid_empty",
			vpcState: VPCState{
				ID:              []byte{},
				Version:         &big.Int{},
				BlockedSender:   &big.Int{},
				BlockedReceiver: &big.Int{},
			},
			want: types.Hex2Bytes("012893657d8eb2efad4de0a91bcd0e39ad9837745dec3ea923737ea803fc8e3d"),
		},
		{
			name: "valid_non_empty",
			vpcState: VPCState{
				ID:              types.Hex2Bytes("ab12cd34"),
				Version:         big.NewInt(1),
				BlockedSender:   big.NewInt(10),
				BlockedReceiver: big.NewInt(20),
			},
			want: types.Hex2Bytes("1ad724f1b659829b64e51a9789af3a1175194857a17a3f6d2061172dea2e6a77"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vpcState.SoliditySHA3(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VPCState.SoliditySHA3() = 0x%x, want 0x%x", got, tt.want)
			}
		})
	}
}
func Test_VPCState_Equal(t *testing.T) {
	type args struct {
		a, b VPCState
	}
	tests := []struct {
		name       string
		args       args
		wantResult bool
	}{
		{
			name: "valid_equal",
			args: args{
				a: VPCState{
					ID:              types.Hex2Bytes("ab12cd34"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				b: VPCState{
					ID:              types.Hex2Bytes("ab12cd34"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
			},
			wantResult: true,
		},
		{
			name: "valid_unequal_id",
			args: args{
				a: VPCState{
					ID:              types.Hex2Bytes("ab12cd34"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				b: VPCState{
					ID:              types.Hex2Bytes("fe00abdc"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
			},
			wantResult: false,
		},
		{
			name: "valid_unequal_version",
			args: args{
				a: VPCState{
					ID:              types.Hex2Bytes("ab12cd34"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				b: VPCState{
					ID:              types.Hex2Bytes("ab12cd34"),
					Version:         big.NewInt(2),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
			},
			wantResult: false,
		},
		{
			name: "valid_unequal_blockedSender",
			args: args{
				a: VPCState{
					ID:              types.Hex2Bytes("ab12cd34"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				b: VPCState{
					ID:              types.Hex2Bytes("ab12cd34"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(11),
					BlockedReceiver: big.NewInt(20),
				},
			},
			wantResult: false,
		},
		{
			name: "valid_unequal_blockedReceiver",
			args: args{
				a: VPCState{
					ID:              types.Hex2Bytes("ab12cd34"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				b: VPCState{
					ID:              types.Hex2Bytes("ab12cd34"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(21),
				},
			},
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := tt.args.a.Equal(tt.args.b)
			if gotResult != tt.wantResult {
				t.Errorf("VPCState.Equal() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_VPCStateSigned_String(t *testing.T) {
	tests := []struct {
		name           string
		vpcStateSigned VPCStateSigned
		want           string
	}{
		{
			name: "valid_empty",
			vpcStateSigned: VPCStateSigned{
				VPCState: VPCState{
					ID:              nil,
					Version:         &big.Int{},
					BlockedSender:   &big.Int{},
					BlockedReceiver: &big.Int{},
				},
				SignSender:   nil,
				SignReceiver: nil,
			},
			want: "{VPCState:{Id:0x Version:0 BlockedSender:0 BlockedReceiver:0} SignSender:0x SignReceiver:0x}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.vpcStateSigned.String()
			if got != tt.want {
				t.Errorf("VPCStateSigned.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_VPCStateSigned_Equal(t *testing.T) {
	type args struct {
		a, b VPCStateSigned
	}
	tests := []struct {
		name       string
		args       args
		wantResult bool
	}{
		{
			name: "valid_equal",
			args: args{
				a: VPCStateSigned{
					VPCState: VPCState{
						ID:              types.Hex2Bytes("ab12cd34"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   []byte("sign-sender"),
					SignReceiver: []byte("sign-receiver"),
				},
				b: VPCStateSigned{
					VPCState: VPCState{
						ID:              types.Hex2Bytes("ab12cd34"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   []byte("sign-sender"),
					SignReceiver: []byte("sign-receiver"),
				},
			},
			wantResult: true,
		},
		{
			name: "valid_unequal_vpcstate_version",
			args: args{
				a: VPCStateSigned{
					VPCState: VPCState{
						ID:              types.Hex2Bytes("ab12cd34"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   []byte("sign-sender"),
					SignReceiver: []byte("sign-receiver"),
				},
				b: VPCStateSigned{
					VPCState: VPCState{
						ID:              types.Hex2Bytes("ab12cd34"),
						Version:         big.NewInt(2),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   []byte("sign-sender"),
					SignReceiver: []byte("sign-receiver"),
				},
			},
			wantResult: false,
		},
		{
			name: "valid_unequal_signsender",
			args: args{
				a: VPCStateSigned{
					VPCState: VPCState{
						ID:              types.Hex2Bytes("ab12cd34"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   []byte("sign-sender"),
					SignReceiver: []byte("sign-receiver"),
				},
				b: VPCStateSigned{
					VPCState: VPCState{
						ID:              types.Hex2Bytes("ab12cd34"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   []byte("sign-sender-mismatch"),
					SignReceiver: []byte("sign-receiver"),
				},
			},
			wantResult: false,
		},
		{
			name: "valid_unequal_signreceiver",
			args: args{
				a: VPCStateSigned{
					VPCState: VPCState{
						ID:              types.Hex2Bytes("ab12cd34"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   []byte("sign-sender"),
					SignReceiver: []byte("sign-receiver"),
				},
				b: VPCStateSigned{
					VPCState: VPCState{
						ID:              types.Hex2Bytes("ab12cd34"),
						Version:         big.NewInt(1),
						BlockedSender:   big.NewInt(10),
						BlockedReceiver: big.NewInt(20),
					},
					SignSender:   []byte("sign-sender"),
					SignReceiver: []byte("sign-receiver-mismatch"),
				},
			},
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := tt.args.a.Equal(tt.args.b)
			if gotResult != tt.wantResult {
				t.Errorf("VPCStateSigned.Equal() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_VPCStateSigned_AddSign(t *testing.T) {

	type args struct {
		idWithCreds identity.OffChainID
		role        Role
	}
	tests := []struct {
		name           string
		vpcStateSigned VPCStateSigned
		args           args
		wantSignature  []byte
		wantErr        bool
	}{
		{
			name: "valid_addSignSender",
			vpcStateSigned: VPCStateSigned{
				VPCState: VPCState{
					ID:              []byte("sample-id"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				SignSender:   nil,
				SignReceiver: nil,
			},
			args: args{
				idWithCreds: identity.OffChainID{
					OnChainID:        aliceID.OnChainID,
					ListenerIPAddr:   aliceID.ListenerIPAddr,
					ListenerEndpoint: aliceID.ListenerEndpoint,
					KeyStore:         testKeyStore,
					Password:         alicePassword,
				},
				role: Sender,
			},
			wantSignature: types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
			wantErr:       false,
		},
		{
			name: "valid_addSignReceiver",
			vpcStateSigned: VPCStateSigned{
				VPCState: VPCState{
					ID:              []byte("sample-id"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				SignSender:   nil,
				SignReceiver: nil,
			},
			args: args{
				idWithCreds: identity.OffChainID{
					OnChainID:        bobID.OnChainID,
					ListenerIPAddr:   bobID.ListenerIPAddr,
					ListenerEndpoint: bobID.ListenerEndpoint,
					KeyStore:         testKeyStore,
					Password:         bobPassword,
				},
				role: Receiver,
			},
			wantSignature: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
			wantErr:       false,
		},
		{
			name: "invalid-sender-password",
			vpcStateSigned: VPCStateSigned{
				VPCState: VPCState{
					ID:              []byte("sample-id"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				SignSender:   nil,
				SignReceiver: nil,
			},
			args: args{
				idWithCreds: identity.OffChainID{
					OnChainID:        aliceID.OnChainID,
					ListenerIPAddr:   aliceID.ListenerIPAddr,
					ListenerEndpoint: aliceID.ListenerEndpoint,
					KeyStore:         testKeyStore,
					Password:         alicePassword + "xyz",
				},
				role: Sender,
			},
			wantSignature: types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
			wantErr:       true,
		},
		{
			name: "invalid-role",
			vpcStateSigned: VPCStateSigned{
				VPCState:     VPCState{},
				SignSender:   nil,
				SignReceiver: nil,
			},
			args: args{
				idWithCreds: identity.OffChainID{},
				role:        Role("invalid-role"),
			},
			wantSignature: types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vpcStateSigned.AddSign(tt.args.idWithCreds, tt.args.role)
			if (err != nil) != tt.wantErr {
				t.Fatalf("VPCStateSigned.AddSign() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr {
				t.Logf("VPCStateSigned.AddSign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.args.role == Sender {
				if !bytes.Equal(tt.vpcStateSigned.SignSender, tt.wantSignature) {
					t.Errorf("vpcStateSigned.AddSign() signature = %x, want %x", tt.vpcStateSigned.SignSender, tt.wantSignature)
				}
			} else if tt.args.role == Receiver {
				if !bytes.Equal(tt.vpcStateSigned.SignReceiver, tt.wantSignature) {
					t.Errorf("vpcStateSigned.AddSign() signature = %x, want %x", tt.vpcStateSigned.SignReceiver, tt.wantSignature)
				}
			}

		})
	}
}

func Test_VPCStateSigned_VerifySign(t *testing.T) {
	type args struct {
		id   identity.OffChainID
		role Role
	}
	tests := []struct {
		name           string
		vpcStateSigned VPCStateSigned
		args           args
		wantIsValid    bool
		wantErr        bool
	}{
		{
			name: "valid_roleSender",
			vpcStateSigned: VPCStateSigned{
				VPCState: VPCState{
					ID:              []byte("sample-id"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				SignSender:   types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f61c"),
				SignReceiver: nil,
			},
			args: args{
				id: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
				},
				role: Sender,
			},
			wantIsValid: true,
			wantErr:     false,
		},
		{
			name: "valid_roleReceiver",
			vpcStateSigned: VPCStateSigned{
				VPCState: VPCState{
					ID:              []byte("sample-id"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				SignSender:   nil,
				SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
			},
			args: args{
				id: identity.OffChainID{
					OnChainID: bobID.OnChainID,
				},
				role: Receiver,
			},
			wantIsValid: true,
			wantErr:     false,
		},
		{
			name: "valid_roleReceiver_wrongID",
			vpcStateSigned: VPCStateSigned{
				VPCState: VPCState{
					ID:              []byte("sample-id"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				SignSender:   nil,
				SignReceiver: types.Hex2Bytes("1172ed9142808860222ae220d99203a14ab6756ee65de9f091868905dafb195606239f0962886d54b029c08ee26551f057e97254ed3c78a728ac38bf1865ea1f1b"),
			},
			args: args{
				id: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
				},
				role: Receiver,
			},
			wantIsValid: false,
			wantErr:     false,
		},
		{
			name: "invalid_SignSender",
			vpcStateSigned: VPCStateSigned{
				VPCState: VPCState{
					ID:              []byte("sample-id"),
					Version:         big.NewInt(1),
					BlockedSender:   big.NewInt(10),
					BlockedReceiver: big.NewInt(20),
				},
				//For valid signature, Last byte should be 0x1B (27) or 0x1C (28), here it is 01
				SignSender:   types.Hex2Bytes("14c5811728aed81bc8f7eb1fb71f45d7e7364561f88987da73b9c209b3ec240f2f56f83ad507b425ed096dfbbcacdc2558b6f47c13d91da413ab84f801ee49f601"),
				SignReceiver: nil,
			},
			args: args{
				id: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
				},
				role: Sender,
			},
			wantIsValid: false,
			wantErr:     true,
		},
		{
			name:           "invalid_role",
			vpcStateSigned: VPCStateSigned{},
			args: args{
				id: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
				},
				role: Role("invalid-role"),
			},
			wantIsValid: false,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsValid, err := tt.vpcStateSigned.VerifySign(tt.args.id, tt.args.role)
			if (err != nil) != tt.wantErr {
				t.Fatalf("VPCStateSigned.VerifySign() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotIsValid != tt.wantIsValid {
				t.Errorf("VPCStateSigned.VerifySign() = %v, want %v", gotIsValid, tt.wantIsValid)
			}
		})
	}
}

func Test_GenerateRandomNumber(t *testing.T) {

	randomNumberSize := 32

	randomNumber, err := GenerateRandomNumber(randomNumberSize)
	if err != nil {
		t.Fatalf("GenerateRandomNumber() err = %v, want nil", nil)
	}

	gotRandomNumberSize := len(randomNumber)
	if gotRandomNumberSize != randomNumberSize {
		t.Errorf("GenerateRandomNumber() randomNumber of length = %v, want length %v", gotRandomNumberSize, randomNumberSize)
	}
}
