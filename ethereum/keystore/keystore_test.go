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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/direct-state-transfer/dst-go/ethereum/types"
)

func Test_KeyStore_GetKey(t *testing.T) {
	type args struct {
		KeyStore *KeyStore
		ethAddr  types.Address
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "validKey1",
			args: args{
				KeyStore: testKeyStore,
				ethAddr:  aliceEthereumAddr,
				password: alicePassword},
			wantErr: false,
		},
		{
			name: "validKey2",
			args: args{
				KeyStore: testKeyStore,
				ethAddr:  bobEthereumAddr,
				password: bobPassword},
			wantErr: false,
		},
		{
			name: "wrongPassphrase",
			args: args{
				KeyStore: testKeyStore,
				ethAddr:  bobEthereumAddr,
				password: bobPassword + "123"},
			wantErr: true,
		},

		{
			name: "invalidFile",
			args: args{
				KeyStore: testKeyStore,
				ethAddr:  types.Address{},
				password: "dummy-password"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := tt.args.KeyStore
			_, err := ks.GetKey(tt.args.ethAddr, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("KeyStore.GetKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

	t.Run("file-removed-by-user", func(t *testing.T) {
		ks := testKeyStore

		//Setup
		ethAccount, err := ks.Find(MakeAccount(aliceEthereumAddr))
		if err != nil {
			t.Errorf("Setup : KeyStore.Find() error = %v, wantErr nil", err)
			return
		}

		keyFileURL := ethAccount.URL.Path
		tempAlternateFileName := filepath.Dir(keyFileURL) + "temp-key-file-name"
		err = os.Rename(keyFileURL, tempAlternateFileName)
		if err != nil {
			t.Errorf("Setup : os.Rename() error = %v, wantErr nil", err)
			return
		}
		//Teardown
		defer func() {
			err = os.Rename(tempAlternateFileName, keyFileURL)
			if err != nil {
				t.Errorf("Teardown : os.Rename() error = %v, wantErr nil", err)
				return
			}
		}()

		//Test
		_, err = ks.GetKey(aliceEthereumAddr, alicePassword)
		if err == nil {
			t.Errorf("KeyStore.GetKey() error = nil, want non nil error")
		} else {
			t.Logf("KeyStore.GetKey() error = %v, want nil", err)
		}
	})

}

func Test_KeyStore_SignHashWithPassphrase(t *testing.T) {

	sampleHash := types.Hex2Bytes("b48d38f93eaa084033fc5970bf96e559c33c4cdc07d889ab00b4d63f9590739d")

	type fields struct {
		KeyStore *KeyStore
	}
	type args struct {
		ethAddr    types.Address
		passphrase string
		hash       []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "valid",
			fields: fields{
				KeyStore: testKeyStore,
			},
			args: args{
				ethAddr:    aliceEthereumAddr,
				passphrase: alicePassword,
				hash:       sampleHash,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			account, err := tt.fields.KeyStore.Find(MakeAccount(tt.args.ethAddr))
			if err != nil {
				t.Errorf("Setup : kesytore.Find() error = %v, want nil", err)
				return
			}

			wantSignature, wantErr := tt.fields.KeyStore.KeyStore.SignHashWithPassphrase(account.Account, tt.args.passphrase, tt.args.hash)
			gotSignature, gotErr := tt.fields.KeyStore.SignHashWithPassphrase(account, tt.args.passphrase, tt.args.hash)
			if wantErr != gotErr {
				t.Fatalf("KeyStore.SignHashWithPassphrase() error = %v, wantErr %v", gotErr, wantErr)
			}
			if !reflect.DeepEqual(gotSignature, wantSignature) {
				t.Errorf("KeyStore.SignHashWithPassphrase() = %v, want %v", gotSignature, wantSignature)
			}
		})
	}
}
