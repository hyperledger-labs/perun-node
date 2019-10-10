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

package keystore

import (
	"bytes"
	"crypto/ecdsa"
	"math/big"
	"reflect"
	"testing"

	"github.com/direct-state-transfer/dst-go/ethereum/keystore/mocks"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type hashSignaturePairs struct {
	hash      []byte
	signature []byte
}

//Reference values generated online and validated with multiple tools
var (
	aliceValid1 = hashSignaturePairs{
		hash:      types.Hex2Bytes("b48d38f93eaa084033fc5970bf96e559c33c4cdc07d889ab00b4d63f9590739d"),
		signature: types.Hex2Bytes("343bdef3d12533ca7236c1f402e867252aeb22e4b4059ff8caa4012bcb5d433e1397e2b04f120ca050d8cfc9e997ffa03a53b687b69ba86d7e4b5511ebe337d600"),
	}
	alicePubKey  = types.Hex2Bytes("0417e8bb87d94d2da9748bb666981335c08e4a545812984a5d5afa29471708d3a0e843637c7ddd6a09c1d227b866b387e3c371f95fa22bf1ffa737cf29b19f6a1b")
	alicePubKeyX = types.Hex2Bytes("17e8bb87d94d2da9748bb666981335c08e4a545812984a5d5afa29471708d3a0")
	alicePubKeyY = types.Hex2Bytes("e843637c7ddd6a09c1d227b866b387e3c371f95fa22bf1ffa737cf29b19f6a1b")
)

func Test_EthereumCryptoWrapper_SigToPub(t *testing.T) {

	type args struct {
		hash []byte
		sig  []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "valid",
			args: args{
				hash: aliceValid1.hash,
				sig:  aliceValid1.signature},
		},
	}
	wrapper := EthereumCryptoWrapper{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantPublicKey, wantErr := crypto.SigToPub(tt.args.hash, tt.args.sig)
			gotPublicKey, gotErr := wrapper.SigToPub(tt.args.hash, tt.args.sig)

			if !reflect.DeepEqual(wantPublicKey, gotPublicKey) {
				t.Errorf("SigToPub() want public key=%+v, got %+v", wantPublicKey, gotPublicKey)
			}
			if !reflect.DeepEqual(wantErr, gotErr) {
				t.Errorf("SigToPub() want err=%+v, got %+v", wantErr, gotErr)
			}
		})
	}

}

func Test_EthereumCryptoWrapper_PubkeyToAddress(t *testing.T) {
	type args struct {
		pubkey ecdsa.PublicKey
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "valid",
			args: args{
				pubkey: ecdsa.PublicKey{
					Curve: crypto.S256(),
					X:     big.NewInt(0).SetBytes(alicePubKeyX),
					Y:     big.NewInt(0).SetBytes(alicePubKeyY)}},
		},
	}

	wrapper := EthereumCryptoWrapper{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantAddress := crypto.PubkeyToAddress(tt.args.pubkey)
			gotAddress := wrapper.PubkeyToAddress(tt.args.pubkey)

			if !bytes.Equal(wantAddress.Bytes(), gotAddress.Bytes()) {
				t.Errorf("PubkeyToAddress() = %v, want %v", gotAddress, wantAddress)
			}
		})
	}
}

func Test_EthereumCryptoWrapper_FromECDSAPub(t *testing.T) {
	type args struct {
		pubkey *ecdsa.PublicKey
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "valid",
			args: args{
				pubkey: &ecdsa.PublicKey{
					Curve: crypto.S256(),
					X:     big.NewInt(0).SetBytes(alicePubKeyX),
					Y:     big.NewInt(0).SetBytes(alicePubKeyY)}},
		},
	}
	wrapper := EthereumCryptoWrapper{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := crypto.FromECDSAPub(tt.args.pubkey)
			got := wrapper.FromECDSAPub(tt.args.pubkey)
			if !bytes.Equal(got, want) {
				t.Errorf("EthereumCryptoWrapper.FromECDSAPub() = %v, want %v", got, want)
			}
		})
	}
}

func Test_EthereumCryptoWrapper_VerifySignature(t *testing.T) {
	type args struct {
		pubkey    []byte
		hash      []byte
		signature []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "valid",
			args: args{
				pubkey:    alicePubKey,
				hash:      aliceValid1.hash,
				signature: aliceValid1.signature,
			}},
	}
	wrapper := EthereumCryptoWrapper{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := crypto.VerifySignature(tt.args.pubkey, tt.args.hash, tt.args.signature)
			got := wrapper.VerifySignature(tt.args.pubkey, tt.args.hash, tt.args.signature)
			if got != want {
				t.Errorf("EthereumCryptoWrapper.VerifySignature() = %v, want %v", got, want)
			}
		})
	}
}

func Test_EthereumCryptoWrapper_Keccak256(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "valid",
			args: args{
				data: types.Hex2Bytes("b48d38f93eaa084033fc5970bf96e559c3"), //some randomn data
			},
		},
	}
	wrapper := EthereumCryptoWrapper{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := crypto.Keccak256(tt.args.data)
			got := wrapper.Keccak256(tt.args.data)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("EthereumCryptoWrapper.Keccak256() = %v, want %v", got, want)
			}
		})
	}
}

var CryptoWrapperObj = &mocks.CryptoWrapperInterface{}

func Test_SigToPub(t *testing.T) {

	//Setup
	ActualCryptoWrapper := CryptoWrapperInstance
	CryptoWrapperInstance = CryptoWrapperObj
	//Teardown
	defer func() { CryptoWrapperInstance = ActualCryptoWrapper }()

	hash := []byte{}
	signature := []byte{}
	CryptoWrapperObj.On("SigToPub", hash, signature).Return(nil, nil)
	_, _ = SigToPub(hash, signature)

	if !CryptoWrapperObj.AssertCalled(t, "SigToPub", hash, signature) {
		t.Errorf("SigToPub() was not called as expected")
	}
}

func Test_PubkeyToAddress(t *testing.T) {

	//Setup
	ActualCryptoWrapper := CryptoWrapperInstance
	CryptoWrapperInstance = CryptoWrapperObj
	//Teardown
	defer func() { CryptoWrapperInstance = ActualCryptoWrapper }()

	publicKey := ecdsa.PublicKey{}
	CryptoWrapperObj.On("PubkeyToAddress", publicKey).Return(nil)
	_ = PubkeyToAddress(publicKey)

	if !CryptoWrapperObj.AssertCalled(t, "PubkeyToAddress", publicKey) {
		t.Errorf("PubkeyToAddress() was not called as expected")
	}
}

func Test_FromECDSAPub(t *testing.T) {

	//Setup
	ActualCryptoWrapper := CryptoWrapperInstance
	CryptoWrapperInstance = CryptoWrapperObj
	//Teardown
	defer func() { CryptoWrapperInstance = ActualCryptoWrapper }()

	publicKey := &ecdsa.PublicKey{}
	CryptoWrapperObj.On("FromECDSAPub", publicKey).Return(nil)
	_ = FromECDSAPub(publicKey)

	if !CryptoWrapperObj.AssertCalled(t, "FromECDSAPub", publicKey) {
		t.Errorf("FromECDSAPub() was not called as expected")
	}
}

func Test_VerifySignature(t *testing.T) {

	//Setup
	ActualCryptoWrapper := CryptoWrapperInstance
	CryptoWrapperInstance = CryptoWrapperObj
	//Teardown
	defer func() { CryptoWrapperInstance = ActualCryptoWrapper }()

	pubkey := []byte{}
	hash := []byte{}
	signature := []byte{}

	CryptoWrapperObj.On("VerifySignature", pubkey, hash, signature).Return(false)
	_ = VerifySignature(pubkey, hash, signature)

	if !CryptoWrapperObj.AssertCalled(t, "VerifySignature", pubkey, hash, signature) {
		t.Errorf("VerifySignature() was not called as expected")
	}
}

func Test_Keccak256(t *testing.T) {

	//Setup
	ActualCryptoWrapper := CryptoWrapperInstance
	CryptoWrapperInstance = CryptoWrapperObj
	//Teardown
	defer func() { CryptoWrapperInstance = ActualCryptoWrapper }()

	data := []byte{}
	CryptoWrapperObj.On("Keccak256", data).Return(nil)
	_ = Keccak256(data)

	if !CryptoWrapperObj.AssertCalled(t, "Keccak256", data) {
		t.Errorf("Keccak256() was not called as expected")
	}
}
