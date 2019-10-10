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

package identity

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/direct-state-transfer/dst-go/ethereum/keystore"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
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
	bobValid1 = hashSignaturePairs{
		hash:      types.Hex2Bytes("b48d38f93eaa084033fc5970bf96e559c33c4cdc07d889ab00b4d63f9590739d"),
		signature: types.Hex2Bytes("f420c36ee23594a00e38cd74a773ff481658c03b32c81f2961698af8ff5672ff58d02f49fbbc3d605fb3c7d7ce920d756ad0c4719ae41598d51fd468349f787100"),
	}
	aliceValid2 = hashSignaturePairs{
		hash:      types.Hex2Bytes("a416544422f6b92e460eb5de02d5121ff40767ba236073fc581d3cb9831203e2"),
		signature: types.Hex2Bytes("fe4bb7aa6623a2c51db102b0f4e33f51e8b6cf3a9179f1ee68a109b95c5cd71b5df7c50c446d1ad4c86abeb8994a918bbfc642f837358183c5a1c746cf6b5f9001"),
	}
	invalidRecoveryID = hashSignaturePairs{
		hash:      types.Hex2Bytes("a416544422f6b92e460eb5de02d5121ff40767ba236073fc581d3cb9831203e2"),
		signature: types.Hex2Bytes("fe4bb7aa6623a2c51db102b0f4e33f51e8b6cf3a9179f1ee68a109b95c5cd71b5df7c50c446d1ad4c86abeb8994a918bbfc642f837358183c5a1c746cf6b5f9008"),
	}
)

//Reference values generated from node-js console
var (
	aliceValidEth1 = hashSignaturePairs{
		hash:      types.Hex2Bytes("a416544422f6b92e460eb5de02d5121ff40767ba236073fc581d3cb9831203e2"),
		signature: types.Hex2Bytes("2c7a54d60745ca35eb4c970e8ca7b9d532d2063a2b64332fbefdbf34d57eaa9722d93a54a11084ac8ba312465cbed5b427b0c8465cc5ef9b7e1018ac0a0d2ea71b"),
	}
	bobValidEth1 = hashSignaturePairs{
		hash:      types.Hex2Bytes("a416544422f6b92e460eb5de02d5121ff40767ba236073fc581d3cb9831203e2"),
		signature: types.Hex2Bytes("c143aadf817683f26ba1ec2f252d503b3156b7c1f44f6a7c8cf0261d48c6afee58fdfd3db2460bf57e073a7935cf2f713f72c535c91671beea01286b908447841b"),
	}
	invalidRecoveryIDEth = hashSignaturePairs{
		hash:      types.Hex2Bytes("b48d38f93eaa084033fc5970bf96e559c33c4cdc07d889ab00b4d63f9590739d"),
		signature: types.Hex2Bytes("343bdef3d12533ca7236c1f402e867252aeb22e4b4059ff8caa4012bcb5d433e1397e2b04f120ca050d8cfc9e997ffa03a53b687b69ba86d7e4b5511ebe337d600"),
	}
)

func Test_SignHashWithPassword(t *testing.T) {

	type args struct {
		idWithCredentials OffChainID
		hash              []byte
	}
	tests := []struct {
		name          string
		args          args
		wantSignature []byte
		wantErr       bool
	}{
		{
			name: "valid-1",
			args: args{
				idWithCredentials: OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword,
					KeyStore:  testKeyStore,
				},
				hash: aliceValid1.hash,
			},
			wantSignature: aliceValid1.signature,
			wantErr:       false,
		},
		{
			name: "valid-2",
			args: args{
				idWithCredentials: OffChainID{
					OnChainID: bobID.OnChainID,
					Password:  bobPassword,
					KeyStore:  testKeyStore,
				},
				hash: bobValid1.hash,
			},
			wantSignature: bobValid1.signature,
			wantErr:       false,
		},
		{
			name: "missing-credentials",
			args: args{
				idWithCredentials: OffChainID{
					OnChainID: aliceID.OnChainID,
				},
				hash: aliceValid1.hash,
			},
			wantErr: true,
		}, {
			name: "wrong-password",
			args: args{
				idWithCredentials: OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword + "rand",
					KeyStore:  testKeyStore,
				},
				hash: aliceValid1.hash,
			},
			wantErr: true,
		}, {
			name: "key-missing-in-keystore",
			args: args{
				idWithCredentials: OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword,
					KeyStore:  dummyKeyStore,
				},
				hash: aliceValid1.hash,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSignature, err := SignHashWithPassword(tt.args.idWithCredentials, tt.args.hash)

			if (err != nil) != tt.wantErr {
				t.Fatalf("SignHashWithPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(gotSignature, tt.wantSignature) {
				t.Errorf("SignHashWithPassword() = %x, want %x", gotSignature, tt.wantSignature)
			}
		})
	}
}

func Test_VerifySignature(t *testing.T) {
	type args struct {
		hash    []byte
		sign    []byte
		ethAddr []byte
	}
	tests := []struct {
		name          string
		args          args
		wantIsSuccess bool
		wantErr       bool
	}{
		{
			name: "valid",
			args: args{
				hash:    aliceValid1.hash,
				sign:    aliceValid1.signature,
				ethAddr: aliceID.OnChainID.Bytes(),
			},
			wantIsSuccess: true,
			wantErr:       false,
		},
		{
			name: "valid-2",
			args: args{
				hash:    bobValid1.hash,
				sign:    bobValid1.signature,
				ethAddr: bobID.OnChainID.Bytes(),
			},
			wantIsSuccess: true,
			wantErr:       false,
		},
		{
			name: "valid-3",
			args: args{
				hash:    aliceValid2.hash,
				sign:    aliceValid2.signature,
				ethAddr: aliceID.OnChainID.Bytes(),
			},
			wantIsSuccess: true,
			wantErr:       false,
		},
		{
			name: "hash-signature-mismatch",
			args: args{
				hash:    aliceValid2.hash,
				sign:    aliceValid1.signature,
				ethAddr: aliceID.OnChainID.Bytes(),
			},
			wantIsSuccess: false,
			wantErr:       false,
		},
		{
			name: "wrong-id",
			args: args{
				hash:    aliceValid1.hash,
				sign:    aliceValid1.signature,
				ethAddr: bobID.OnChainID.Bytes(),
			},
			wantIsSuccess: false,
			wantErr:       false,
		},
		{
			name: "invalid-signature",
			args: args{
				hash:    invalidRecoveryID.hash,
				sign:    invalidRecoveryID.signature,
				ethAddr: bobID.OnChainID.Bytes(),
			},
			wantIsSuccess: false,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsSuccess, err := VerifySignature(tt.args.hash, tt.args.sign, tt.args.ethAddr)

			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotIsSuccess != tt.wantIsSuccess {
				t.Errorf("VerifySignature() = %v, want %v", gotIsSuccess, tt.wantIsSuccess)
			}
		})
	}
}
func Test_SignHashWithPasswordEth(t *testing.T) {

	type args struct {
		idWithCredentials OffChainID
		hash              []byte
	}
	tests := []struct {
		name          string
		args          args
		wantSignature []byte
		wantErr       bool
	}{
		{
			name: "valid1",
			args: args{
				idWithCredentials: OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword,
					KeyStore:  testKeyStore,
				},
				hash: aliceValidEth1.hash,
			},
			wantSignature: aliceValidEth1.signature,
			wantErr:       false,
		},
		{
			name: "valid2",
			args: args{
				idWithCredentials: OffChainID{
					OnChainID: bobID.OnChainID,
					Password:  bobPassword,
					KeyStore:  testKeyStore,
				},
				hash: bobValidEth1.hash,
			},
			wantSignature: bobValidEth1.signature,
			wantErr:       false,
		},
		{
			name: "missing-credentials",
			args: args{
				idWithCredentials: OffChainID{
					OnChainID: aliceID.OnChainID,
				},
				hash: aliceValidEth1.hash,
			},
			wantErr: true,
		}, {
			name: "wrong-password",
			args: args{
				idWithCredentials: OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword + "rand",
					KeyStore:  testKeyStore,
				},
				hash: aliceValidEth1.hash,
			},
			wantErr: true,
		}, {
			name: "key-missing-in-keystore",
			args: args{
				idWithCredentials: OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword,
					KeyStore:  dummyKeyStore,
				},
				hash: aliceValidEth1.hash,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSignature, err := SignHashWithPasswordEth(tt.args.idWithCredentials, tt.args.hash)

			if (err != nil) != tt.wantErr {
				t.Fatalf("SignHashWithPasswordEth() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(gotSignature, tt.wantSignature) {
				t.Errorf("SignHashWithPasswordEth() = %x, want %x", gotSignature, tt.wantSignature)
			}
		})
	}
}
func Test_VerifySignatureEth(t *testing.T) {
	type args struct {
		hash    []byte
		sign    []byte
		ethAddr []byte
	}
	tests := []struct {
		name          string
		args          args
		wantIsSuccess bool
		wantErr       bool
	}{
		{
			name: "valid-1",
			args: args{
				hash:    aliceValidEth1.hash,
				sign:    aliceValidEth1.signature,
				ethAddr: aliceID.OnChainID.Bytes(),
			},
			wantIsSuccess: true,
			wantErr:       false,
		},
		{
			name: "valid-2",
			args: args{
				hash:    bobValidEth1.hash,
				sign:    bobValidEth1.signature,
				ethAddr: bobID.OnChainID.Bytes(),
			},
			wantIsSuccess: true,
			wantErr:       false,
		},
		{
			name: "invalid",
			args: args{
				hash:    invalidRecoveryIDEth.hash,
				sign:    invalidRecoveryIDEth.signature,
				ethAddr: aliceID.OnChainID.Bytes(),
			},
			wantIsSuccess: false,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsSuccess, err := VerifySignatureEth(tt.args.hash, tt.args.sign, tt.args.ethAddr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotIsSuccess != tt.wantIsSuccess {
				t.Errorf("VerifySignature() = %v, want %v", gotIsSuccess, tt.wantIsSuccess)
			}
		})
	}
}

type msg struct {
	MyName      string
	PeerName    string
	MyBalance   int
	PeerBalance int
}

func Test_keyHandling(t *testing.T) {

	Msg := msg{"alice", "bob", 10, 20}
	MsgByteArr, _ := json.Marshal(Msg)
	MsgHash := keystore.Keccak256(MsgByteArr)

	aliceID.Password = alicePassword
	aliceID.KeyStore = testKeyStore
	sign, err := SignHashWithPassword(aliceID, MsgHash)
	if err != nil {
		t.Error(err)
	}

	verified, err := VerifySignature(MsgHash, sign, aliceID.OnChainID.Bytes())
	if err != nil {
		t.Error(err)
	}

	t.Log("Message signed by alice. Verification -", verified)

	//Sign and verify with bob key
	bobID.Password = bobPassword
	bobID.KeyStore = testKeyStore
	sign, err = SignHashWithPassword(bobID, MsgHash)
	if err != nil {
		t.Error(err)
	}

	verified, err = VerifySignature(MsgHash, sign, bobID.OnChainID.Bytes())
	if err != nil {
		t.Error(err)
	}

	t.Log("Message signed by bob. Verification -", verified)
}
