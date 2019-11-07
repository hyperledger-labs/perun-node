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

package identity

import (
	"reflect"
	"testing"

	"github.com/direct-state-transfer/dst-go/ethereum/keystore"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/log"
)

func Test_NewSession(t *testing.T) {
	type args struct {
		keysDir string
		idFile  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid_args",
			args: args{
				keysDir: testKeyStorePath,
				idFile:  knownIdsFile,
			},
			wantErr: false,
		},
		{
			name: "invalid_keystore_dir",
			args: args{
				keysDir: "./invalid-keystore",
				idFile:  knownIdsFile,
			},
			wantErr: true,
		},
		{
			name: "invalid_idFile",
			args: args{
				keysDir: testKeyStorePath,
				idFile:  "invalid-idfile",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := NewSession(tt.args.keysDir, tt.args.idFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewSession() got %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_GetKey(t *testing.T) {
	type args struct {
		keyStore  *keystore.KeyStore
		onChainID types.Address
		password  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "validKey1",
			args: args{
				onChainID: aliceID.OnChainID,
				password:  alicePassword},
			wantErr: false,
		},
		{
			name: "validKey2",
			args: args{
				onChainID: bobID.OnChainID,
				password:  bobPassword},
			wantErr: false,
		},
		{
			name: "wrongPassphrase",
			args: args{
				onChainID: bobID.OnChainID,
				password:  bobPassword + "123"},
			wantErr: true,
		},

		{
			name: "invalidFile",
			args: args{
				onChainID: types.Address{},
				password:  "dummy-password"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.keyStore = NewKeystore(testKeyStorePath)
			_, err := GetKey(tt.args.keyStore, tt.args.onChainID, tt.args.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_GetKey_InGoRoutine(t *testing.T) {
	type args struct {
		keyStore  *keystore.KeyStore
		onChainID types.Address
		password  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "validKey1",
			args: args{
				onChainID: aliceID.OnChainID,
				password:  alicePassword},
			wantErr: false,
		},
		{
			name: "validKey2",
			args: args{
				onChainID: bobID.OnChainID,
				password:  bobPassword},
			wantErr: false,
		},
		{
			name: "wrongPassphrase",
			args: args{
				onChainID: bobID.OnChainID,
				password:  bobPassword + "123"},
			wantErr: true,
		},

		{
			name: "invalidFile",
			args: args{
				onChainID: types.Address{},
				password:  "dummy-password"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.keyStore = NewKeystore(testKeyStorePath)
			gotKeyChan := make(chan *keystore.Key)
			errChan := make(chan error)

			go func(gotKeyChan chan *keystore.Key, errChan chan error) {
				gotKey, err := GetKey(tt.args.keyStore, tt.args.onChainID, tt.args.password)
				gotKeyChan <- gotKey
				errChan <- err
			}(gotKeyChan, errChan)
			<-gotKeyChan
			err := <-errChan

			if (err != nil) != tt.wantErr {
				t.Errorf("GetKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_OffchainID_SetCredentials(t *testing.T) {
	type fields struct {
		OnChainID        types.Address
		ListenerIPAddr   string
		ListenerEndpoint string
		Keystore         *keystore.KeyStore
		Password         string
	}
	type args struct {
		ks       *keystore.KeyStore
		password string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantSuccess bool
	}{
		{
			name: "validKeystore-emptyPassword",
			args: args{
				ks:       dummyKeyStore,
				password: "",
			},
			wantSuccess: true,
		},
		{
			name: "validKeystore-nonEmptyPassword",
			args: args{
				ks:       dummyKeyStore,
				password: "some-non-empty-password",
			},
			wantSuccess: true,
		},
		{
			name: "nilKeystore-emptyPassword",
			args: args{
				password: "",
			},
			wantSuccess: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := &OffChainID{
				OnChainID:        tt.fields.OnChainID,
				ListenerIPAddr:   tt.fields.ListenerIPAddr,
				ListenerEndpoint: tt.fields.ListenerEndpoint,
				KeyStore:         tt.fields.Keystore,
				Password:         tt.fields.Password,
			}
			if gotSuccess := id.SetCredentials(tt.args.ks, tt.args.password); gotSuccess != tt.wantSuccess {
				t.Errorf("OffChainID.SetCredentials() = %v, want %v", gotSuccess, tt.wantSuccess)
			}

			if !reflect.DeepEqual(id.KeyStore, tt.args.ks) {
				t.Errorf("OffChainID.SetCredentials() has set Keystore %v, want %v", *id.KeyStore, *tt.args.ks)
			}
			if !reflect.DeepEqual(id.Password, tt.args.password) {
				t.Errorf("OffChainID.SetCredentials() has set password %v, want %v", id.Password, tt.args.password)
			}
		})
	}
}

func Test_OffchainID_GetCredentials(t *testing.T) {
	type fields struct {
		OnChainID        types.Address
		ListenerIPAddr   string
		ListenerEndpoint string
		Keystore         *keystore.KeyStore
		Password         string
	}
	tests := []struct {
		name          string
		fields        fields
		wantKs        *keystore.KeyStore
		wantPassword  string
		wantAvailable bool
	}{
		{
			name: "validKeystore-emptyPassword",
			fields: fields{
				Keystore: dummyKeyStore,
				Password: "",
			},
			wantKs:        dummyKeyStore,
			wantPassword:  "",
			wantAvailable: true,
		},
		{
			name: "validKeystore-nonEmptyPassword",
			fields: fields{
				Keystore: dummyKeyStore,
				Password: "some-non-empty-password",
			},
			wantKs:        dummyKeyStore,
			wantPassword:  "some-non-empty-password",
			wantAvailable: true,
		},
		{
			name: "nilKeystore-emptyPassword",
			fields: fields{
				Keystore: nil,
				Password: "",
			},
			wantKs:        nil,
			wantPassword:  "",
			wantAvailable: false,
		},
		{
			name: "nilKeystore-nonEmptyPassword",
			fields: fields{
				Keystore: nil,
				Password: "some-non-empty-password",
			},
			wantKs:        nil,
			wantPassword:  "",
			wantAvailable: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := &OffChainID{
				OnChainID:        tt.fields.OnChainID,
				ListenerIPAddr:   tt.fields.ListenerIPAddr,
				ListenerEndpoint: tt.fields.ListenerEndpoint,
				KeyStore:         tt.fields.Keystore,
				Password:         tt.fields.Password,
			}
			gotKs, gotPassword, gotAvailable := id.GetCredentials()

			if !reflect.DeepEqual(gotKs, tt.wantKs) {
				t.Errorf("OffChainID.GetCredentials() gotKs = %v, want %v", gotKs, tt.wantKs)
			}
			if gotPassword != tt.wantPassword {
				t.Errorf("OffChainID.GetCredentials() gotPassword = %v, want %v", gotPassword, tt.wantPassword)
			}
			if gotAvailable != tt.wantAvailable {
				t.Errorf("OffChainID.GetCredentials() gotAvailable = %v, want %v", gotAvailable, tt.wantAvailable)
			}
		})
	}
}

func Test_OffchainID_ClearCredentials(t *testing.T) {
	type fields struct {
		OnChainID        types.Address
		ListenerIPAddr   string
		ListenerEndpoint string
		Keystore         *keystore.KeyStore
		Password         string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "validKeystore-emptyPassword",
			fields: fields{
				Keystore: dummyKeyStore,
				Password: "",
			},
		},
		{
			name: "validKeystore-nonEmptyPassword",
			fields: fields{
				Keystore: dummyKeyStore,
				Password: "some-non-empty-password",
			},
		},
		{
			name: "nilKeystore-emptyPassword",
			fields: fields{
				Keystore: nil,
				Password: "",
			},
		},
		{
			name: "nilKeystore-nonEmptyPassword",
			fields: fields{
				Keystore: nil,
				Password: "some-non-empty-password",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := &OffChainID{
				OnChainID:        tt.fields.OnChainID,
				ListenerIPAddr:   tt.fields.ListenerIPAddr,
				ListenerEndpoint: tt.fields.ListenerEndpoint,
				KeyStore:         tt.fields.Keystore,
				Password:         tt.fields.Password,
			}
			id.ClearCredentials()

			if id.KeyStore != nil {
				t.Errorf("OffChainID.ClearCredentials() id.KeyStore = %v, want nil", id.KeyStore)
			}
			if id.Password != "" {
				t.Errorf("OffChainID.ClearCredentials() id.Password = %v, want \"\"", id.Password)
			}
		})
	}
}

func Test_IsKeysPresent(t *testing.T) {
	type args struct {
		ks               *keystore.KeyStore
		requiredAccounts []types.Address
	}
	tests := []struct {
		name                string
		args                args
		wantMissingAccounts []types.Address
	}{
		{
			name: "all_keys_present",
			args: args{
				ks:               testKeyStore,
				requiredAccounts: []types.Address{aliceID.OnChainID},
			},
			wantMissingAccounts: []types.Address{},
		},
		{
			name: "keys_not_present",
			args: args{
				ks:               dummyKeyStore,
				requiredAccounts: []types.Address{aliceID.OnChainID},
			},
			wantMissingAccounts: []types.Address{aliceID.OnChainID},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMissingAccounts := IsKeysPresent(tt.args.ks, tt.args.requiredAccounts)

			if !reflect.DeepEqual(gotMissingAccounts, tt.wantMissingAccounts) {
				t.Errorf("IsKeysPresent() = %v, want %v", gotMissingAccounts, tt.wantMissingAccounts)
			}
		})
	}
}

func Test_Equal(t *testing.T) {
	type args struct {
		a OffChainID
		b OffChainID
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "match",
			args: args{
				a: aliceID,
				b: aliceID,
			},
			want: true,
		},
		{
			name: "no-match",
			args: args{
				a: aliceID,
				b: bobID,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Equal(tt.args.a, tt.args.b)

			if got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_OffchainID_String(t *testing.T) {
	type fields struct {
		OnChainID      types.Address
		ListenerIPAddr string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "valid-1",
			fields: fields{
				OnChainID:      aliceID.OnChainID,
				ListenerIPAddr: "192.168.1.1:3125",
			},
			want: "OnChainID :0x932a74da117eb9288ea759487360cd700e7777e1, ListenerIPAddr 192.168.1.1:3125",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := OffChainID{
				OnChainID:      tt.fields.OnChainID,
				ListenerIPAddr: tt.fields.ListenerIPAddr,
			}
			got := id.String()

			if got != tt.want {
				t.Errorf("OffChainID.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_OffchainID_ListenerLocalAddr(t *testing.T) {
	type fields struct {
		ListenerIPAddr string
	}
	tests := []struct {
		name                  string
		fields                fields
		wantListenerLocalAddr string
		wantErr               bool
	}{
		{
			name: "valid-1",
			fields: fields{
				ListenerIPAddr: "192.168.1.1:3125",
			},
			wantListenerLocalAddr: "localhost:3125",
			wantErr:               false,
		},
		{
			name: "invalid-1",
			fields: fields{
				ListenerIPAddr: "192.168.1.1",
			},
			wantErr: true,
		},
		{
			name: "invalid-2",
			fields: fields{
				ListenerIPAddr: "192.168.1.1:3126:323",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := OffChainID{
				ListenerIPAddr: tt.fields.ListenerIPAddr,
			}
			gotListenerLocalAddr, err := id.ListenerLocalAddr()

			if (err != nil) != tt.wantErr {
				t.Errorf("OffChainID.ListenerLocalAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotListenerLocalAddr != tt.wantListenerLocalAddr {
				t.Errorf("OffChainID.ListenerLocalAddr() = %v, want %v", gotListenerLocalAddr, tt.wantListenerLocalAddr)
			}
		})
	}
}

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
			name: "valid",
			args: args{
				cfg: &Config{
					Logger: log.Config{
						Level:   log.DebugLevel,
						Backend: log.StdoutBackend,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitModule(tt.args.cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("InitModule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_OffChainIDStore_Update(t *testing.T) {
	type fields struct {
		Filename string
		IDList   []OffChainID
		IDMap    map[types.Address]OffChainID
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				Filename: "testdata/known_ids.json",
			},
			wantErr: false,
		},
		{
			name: "invalid_json",
			fields: fields{
				Filename: "testdata/invalid_known_ids.json",
			},
			wantErr: true,
		},
		{
			name: "non_existent_file",
			fields: fields{
				Filename: "testdata/no-such-file",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idStore := &OffChainIDStore{
				Filename: tt.fields.Filename,
				IDMap:    make(map[types.Address]OffChainID),
			}
			err := idStore.Update()

			if (err != nil) != tt.wantErr {
				t.Errorf("OffChainIDStore.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_OffChainIDStore_OffChainID(t *testing.T) {
	type fields struct {
		IDMap map[types.Address]OffChainID
	}
	type args struct {
		OnChainID types.Address
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantID      OffChainID
		wantPresent bool
	}{
		{
			name: "present",
			fields: fields{
				IDMap: map[types.Address]OffChainID{
					aliceID.OnChainID: aliceID},
			},
			args: args{
				OnChainID: aliceID.OnChainID},
			wantID:      aliceID,
			wantPresent: true,
		},
		{
			name: "not_present",
			fields: fields{
				IDMap: make(map[types.Address]OffChainID)},
			args: args{
				OnChainID: aliceID.OnChainID},
			wantID:      OffChainID{},
			wantPresent: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idStore := &OffChainIDStore{
				IDMap: tt.fields.IDMap,
			}
			gotID, gotPresent := idStore.OffChainID(tt.args.OnChainID)

			if !reflect.DeepEqual(gotID, tt.wantID) {
				t.Errorf("OffChainIDStore.OffChainID() gotId = %v, want %v", gotID, tt.wantID)
			}
			if gotPresent != tt.wantPresent {
				t.Errorf("OffChainIDStore.OffChainID() gotPresent = %v, want %v", gotPresent, tt.wantPresent)
			}
		})
	}
}
