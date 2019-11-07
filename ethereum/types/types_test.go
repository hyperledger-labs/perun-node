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

package types

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func Test_Hex2Bytes(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "valid-1",
			args: args{
				str: "932a74da117eb9288ea759487360cd700e7777e1",
			},
			want: []byte{147, 42, 116, 218, 17, 126, 185, 40, 142, 167, 89, 72, 115, 96, 205, 112, 14, 119, 119, 225},
		},
		{
			name: "invalid-prefix0x",
			args: args{
				str: "0x932a74da117eb9288ea759487360cd700e7777e1",
			},
			want: nil,
		},
		{
			name: "error_oddNoOfChars_inHexString",
			args: args{
				str: "932a74da117eb9288ea759487360cd700e7777e",
			},
			want: []byte{147, 42, 116, 218, 17, 126, 185, 40, 142, 167, 89, 72, 115, 96, 205, 112, 14, 119, 119},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Hex2Bytes(tt.args.str); !bytes.Equal(got, tt.want) {
				t.Errorf("Hex2Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_HexToAddress(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		args args
		want Address
	}{
		{
			name: "valid1",
			args: args{
				addr: "0x932a74da117eb9288ea759487360cd700e7777e1",
			},
			want: Address{
				Address: [common.AddressLength]byte{147, 42, 116, 218, 17, 126, 185, 40, 142, 167, 89, 72, 115, 96, 205, 112, 14, 119, 119, 225},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HexToAddress(tt.args.addr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HexToAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_HexToHash(t *testing.T) {
	type args struct {
		hash string
	}
	tests := []struct {
		name string
		args args
		want Hash
	}{
		{
			name: "valid1",
			args: args{
				hash: "0xffcd919661cc14ab71d9d80c3c3481b6b8015c554fe1212a55bbfd138b70fc61",
			},
			want: Hash{
				Hash: [common.HashLength]byte{255, 205, 145, 150, 97, 204, 20, 171, 113, 217, 216, 12, 60, 52, 129, 182, 184, 1, 92, 85, 79, 225, 33, 42, 85, 187, 253, 19, 139, 112, 252, 97},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HexToHash(tt.args.hash); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HexToHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_WeiToEther(t *testing.T) {
	type args struct {
		val *big.Int
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "valid1",
			args: args{
				val: big.NewInt(1),
			},
			want: new(big.Int).Div(big.NewInt(1), big.NewInt(1e18)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WeiToEther(tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WeiToEther() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_EtherToWei(t *testing.T) {
	type args struct {
		val *big.Int
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "valid1",
			args: args{
				val: big.NewInt(1),
			},
			want: new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EtherToWei(tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EtherToWei() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_WeiToGwei(t *testing.T) {
	type args struct {
		val *big.Int
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "valid1",
			args: args{
				val: big.NewInt(1),
			},
			want: new(big.Int).Div(big.NewInt(1), big.NewInt(1e9)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WeiToGwei(tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WeiToGwei() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_GweiToWei(t *testing.T) {
	type args struct {
		val *big.Int
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "valid1",
			args: args{
				val: big.NewInt(1),
			},
			want: new(big.Int).Mul(big.NewInt(1), big.NewInt(1e9)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GweiToWei(tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GweiToWei() = %v, want %v", got, tt.want)
			}
		})
	}
}
