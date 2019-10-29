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

package blockchain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"

	"github.com/direct-state-transfer/dst-go/ethereum/adapter"
	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
	"github.com/direct-state-transfer/dst-go/log"
)

type dummySubscription struct {
}

func (s *dummySubscription) Unsubscribe() {}
func (s *dummySubscription) Err() <-chan error {
	errChan := make(chan error, 1)
	return errChan
}

func Test_InitModule_Integration(t *testing.T) {

	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	//Setup
	conn, err := adapter.NewRealBackend(ethereumNodeURL)
	if err != nil {
		t.Fatalf("Error setting up connection to blockchain node.\n%v", err)
	}

	ownerID := aliceID
	if !ownerID.SetCredentials(testKeyStore, alicePassword) {
		t.Fatalf("Test case setup : ownerId.SetCredentials() failed")
	}

	t.Run("SuccessfulInit_LibSigAddrGiven", func(t *testing.T) {

		libSignaturesAddr, err := setupContract(contract.Store.LibSignatures(), conn, ownerID)
		if err != nil {
			t.Fatalf("Test case setup : setupContract() Error - %v", err)
		}

		cfg := &Config{
			Logger: log.Config{
				Level:   log.ErrorLevel,
				Backend: log.StdoutBackend,
			},
			libSignaturesAddr: libSignaturesAddr,
			gethURL:           ethereumNodeURL,
		}

		gotConn, gotLibSignaturesAddr, err := InitModule(cfg)
		if err != nil {
			t.Fatalf("InitModule() Error - %v", err)
		}

		if 0 != bytes.Compare(gotLibSignaturesAddr.Bytes(), libSignaturesAddr.Bytes()) {
			t.Errorf("InitModule() gotLibSignaturesAddr - %s, want - %s", gotLibSignaturesAddr.Hex(), libSignaturesAddr.Hex())
		}

		if logger == nil {
			t.Errorf("InitModule() logger is nil, want not nil. logger config - %+v ", cfg.Logger)
		}
		_, _ = gotConn, gotLibSignaturesAddr

	})

	t.Run("SuccessfulInit_LibSigAddrNotGiven", func(t *testing.T) {
		cfg := &Config{
			Logger: log.Config{
				Level:   log.ErrorLevel,
				Backend: log.StdoutBackend,
			},
			gethURL: ethereumNodeURL,
		}

		_, _, err := InitModule(cfg)
		if err != nil {
			t.Fatalf("InitModule() Error - %v", err)
		}

		if logger == nil {
			t.Errorf("InitModule() logger is nil, want not nil. logger config - %+v ", cfg.Logger)
		}

	})

	t.Run("SuccessfulInit_InvalidContractAddrGiven", func(t *testing.T) {

		libSignaturesAddr, err := setupContract(contract.Store.MSContract(), conn, ownerID)
		if err != nil {
			t.Fatalf("Test case setup : setupContract() Error - %v", err)
		}

		cfg := &Config{
			Logger: log.Config{
				Level:   log.ErrorLevel,
				Backend: log.StdoutBackend,
			},
			libSignaturesAddr: libSignaturesAddr,
			gethURL:           ethereumNodeURL,
		}

		_, _, err = InitModule(cfg)
		if err != nil {
			t.Fatalf("InitModule() Error - %v", err)
		}

		if logger == nil {
			t.Errorf("InitModule() logger is nil, want not nil. logger config - %+v ", cfg.Logger)
		}
	})

	t.Run("InvalidEthereumNodeUrl", func(t *testing.T) {
		cfg := &Config{
			Logger: log.Config{
				Level:   log.ErrorLevel,
				Backend: log.StdoutBackend,
			},
			gethURL: "invalid-url",
		}

		_, _, err := InitModule(cfg)
		if err == nil {
			t.Errorf("InitModule() want error, got nil")
		}
	})

}

func Test_SetupLibSignatures(t *testing.T) {
	type args struct {
		conn         adapter.ContractBackend
		sessionOwner identity.OffChainID
	}
	tests := []struct {
		name string
		args args

		mockContractToSetup contract.Handler

		wantErr   bool
		wantMatch bool
	}{
		{
			name: "noLibSignAddr",
			args: args{
				sessionOwner: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword,
					KeyStore:  testKeyStore,
				},
			},
			mockContractToSetup: contract.Handler{},
			wantErr:             false,
			wantMatch:           true,
		},
		{
			name: "validMSContract",
			args: args{
				sessionOwner: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword,
					KeyStore:  testKeyStore,
				},
			},
			mockContractToSetup: contract.Store.MSContract(),
			wantErr:             false,
			wantMatch:           false,
		},
		{
			name: "validLibSignAddr",
			args: args{
				sessionOwner: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword,
					KeyStore:  testKeyStore,
				},
			},
			mockContractToSetup: contract.Store.LibSignatures(),
			wantErr:             false,
			wantMatch:           true,
		},
	}

	conn := adapter.NewSimulatedBackend(balanceList)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.args.conn = conn

			var contractAddr types.Address
			var err error
			if !reflect.DeepEqual(tt.mockContractToSetup, contract.Handler{}) {
				contractAddr, err = setupContract(tt.mockContractToSetup, conn, tt.args.sessionOwner)
				if err != nil {
					t.Fatalf("setupContract() error = %v, wantErr %v", err, tt.wantErr)
				}
				conn.Commit()
			}

			_, err = SetupLibSignatures(contractAddr, tt.args.conn, tt.args.sessionOwner)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetupLibSignatures() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func Test_SetupLibSignatures_Integration(t *testing.T) {
	type args struct {
		conn         adapter.ContractBackend
		sessionOwner identity.OffChainID
	}
	tests := []struct {
		name string
		args args

		mockContractToSetup contract.Handler

		wantErr   bool
		wantMatch bool
	}{
		{
			name: "noLibSignAddr",
			args: args{
				sessionOwner: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword,
					KeyStore:  testKeyStore,
				},
			},
			mockContractToSetup: contract.Handler{},
			wantErr:             false,
			wantMatch:           true,
		},
		{
			name: "validMSContract",
			args: args{
				sessionOwner: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword,
					KeyStore:  testKeyStore,
				},
			},
			mockContractToSetup: contract.Store.MSContract(),
			wantErr:             false,
			wantMatch:           false,
		},
		{
			name: "validLibSignAddr",
			args: args{
				sessionOwner: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					Password:  alicePassword,
					KeyStore:  testKeyStore,
				},
			},
			mockContractToSetup: contract.Store.LibSignatures(),
			wantErr:             false,
			wantMatch:           true,
		},
	}

	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	//Setup
	conn, err := adapter.NewRealBackend(ethereumNodeURL)
	if err != nil {
		t.Fatalf("Error setting up connection to blockchain node.\n%v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.args.conn = conn

			var contractAddr types.Address
			var err error
			if !reflect.DeepEqual(tt.mockContractToSetup, contract.Handler{}) {
				contractAddr, err = setupContract(tt.mockContractToSetup, conn, tt.args.sessionOwner)
				if err != nil {
					t.Fatalf("setupContract() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			_, err = SetupLibSignatures(contractAddr, tt.args.conn, tt.args.sessionOwner)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetupLibSignatures() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func setupContract(handler contract.Handler, conn adapter.ContractBackend, defaultID identity.OffChainID) (
	deployedContractAddr types.Address, err error) {

	switch handler {

	case contract.Store.LibSignatures():
		var params []interface{}
		libSignaturesAddr, _, _, err := adapter.DeployContract(handler, conn, params, defaultID)
		if err != nil {
			return types.Address{}, fmt.Errorf("Deploy libSignatures - %v", err)
		}
		deployedContractAddr = libSignaturesAddr

	case contract.Store.VPC():
		var params []interface{}
		libSignaturesAddr, _, _, err := adapter.DeployContract(contract.Store.LibSignatures(), conn, params, defaultID)
		if err != nil {
			return types.Address{}, fmt.Errorf("Deploy libSignatures - %v", err)
		}

		params = append(params, libSignaturesAddr)
		vpcAddr, _, _, err := adapter.DeployContract(handler, conn, params, defaultID)
		if err != nil {
			return types.Address{}, fmt.Errorf("Deploy vpc - %v", err)
		}

		deployedContractAddr = vpcAddr
	case contract.Store.MSContract():
		var params []interface{}
		libSignaturesAddr, _, _, err := adapter.DeployContract(contract.Store.LibSignatures(), conn, params, defaultID)
		if err != nil {
			return types.Address{}, fmt.Errorf("Deploy libSignatures - %v", err)
		}

		params = append(params, libSignaturesAddr, aliceID.OnChainID, bobID.OnChainID)
		msContractAddr, _, _, err := adapter.DeployContract(handler, conn, params, defaultID)
		if err != nil {
			return types.Address{}, fmt.Errorf("Deploy msc - %v", err)
		}

		deployedContractAddr = msContractAddr
	}

	return deployedContractAddr, nil
}

//mock based tests
func Test_Instance_Libsignatures_mock(t *testing.T) {

	t.Run("ValidAddress", func(t *testing.T) {

		contractAddr := types.HexToAddress("de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
		inst := Instance{
			libSignaturesAddr: contractAddr,
		}

		gotContractAddr := inst.LibSignatures()

		if !bytes.Equal(contractAddr.Bytes(), gotContractAddr.Bytes()) {
			t.Errorf("Instance.LibSignatures() = %s, want %s", gotContractAddr.Hex(), contractAddr.Hex())
		}
	})

	t.Run("Empty", func(t *testing.T) {

		inst := Instance{}

		gotContractAddr := inst.LibSignatures()

		if !bytes.Equal(types.Address{}.Bytes(), gotContractAddr.Bytes()) {
			t.Errorf("Instance.LibSignatures() = %s, want empty", gotContractAddr.Hex())
		}
	})
}

func Test_Instance_SetLibsignatures_mock(t *testing.T) {

	contractAddr := types.HexToAddress("de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")

	t.Run("success", func(t *testing.T) {

		conn := &MockContractBackend{}
		inst := Instance{
			Conn: conn,
		}

		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(libSignaturesRuntimeBin, nil)
		err := inst.SetLibSignatures(contractAddr)
		if err != nil {
			t.Fatalf("Instance.LibSignatures() error=%s, want nil", err.Error())
		}

		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("Instance.LibSignatures() did not call conn.CodeAt with expected params")
		}
	})

	t.Run("VerifyCodeAt_Error", func(t *testing.T) {

		conn := &MockContractBackend{}
		inst := Instance{
			Conn: conn,
		}

		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(dummyRuntimeBin, fmt.Errorf(""))
		err := inst.SetLibSignatures(contractAddr)
		if err == nil {
			t.Errorf("Instance.LibSignatures() error=nil, want %s", err.Error())
		}

		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("Instance.LibSignatures() did not call conn.CodeAt with expected params")
		}
	})

	t.Run("VerifyCodeAt_NoMatch", func(t *testing.T) {

		conn := &MockContractBackend{}
		inst := Instance{
			Conn: conn,
		}

		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(dummyRuntimeBin, nil)
		err := inst.SetLibSignatures(contractAddr)
		if err == nil {
			t.Errorf("Instance.LibSignatures() error=nil, want %s", err.Error())
		}

		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("Instance.LibSignatures() did not call conn.CodeAt with expected params")
		}
	})
}

func Test_Instance_MSContractAddr_mock(t *testing.T) {

	t.Run("ValidAddress", func(t *testing.T) {

		contractAddr := types.HexToAddress("de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
		inst := Instance{
			msContractAddr: contractAddr,
		}

		gotContractAddr := inst.MSContractAddr()

		if !bytes.Equal(contractAddr.Bytes(), gotContractAddr.Bytes()) {
			t.Errorf("Instance.MSContractAddr() = %s, want %s", gotContractAddr.Hex(), contractAddr.Hex())
		}
	})

	t.Run("Empty", func(t *testing.T) {

		inst := Instance{}

		gotContractAddr := inst.MSContractAddr()

		if !bytes.Equal(types.Address{}.Bytes(), gotContractAddr.Bytes()) {
			t.Errorf("Instance.MSContractAddr() = %s, want empty", gotContractAddr.Hex())
		}
	})
}
func Test_Instance_SetMSContractAddr_mock(t *testing.T) {

	contractAddr := types.HexToAddress("de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")

	t.Run("success", func(t *testing.T) {

		conn := &MockContractBackend{}
		inst := Instance{
			Conn: conn,
		}

		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(msContractRuntimeBin, nil)
		err := inst.SetMSContractAddr(contractAddr)
		if err != nil {
			t.Fatalf("Instance.SetMSContractAddr() error=%s, want nil", err.Error())
		}

		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("Instance.SetMSContractAddr() did not call conn.CodeAt with expected params")
		}
	})

	t.Run("VerifyCodeAt_Error", func(t *testing.T) {

		conn := &MockContractBackend{}
		inst := Instance{
			Conn: conn,
		}

		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(dummyRuntimeBin, fmt.Errorf(""))
		err := inst.SetMSContractAddr(contractAddr)
		if err == nil {
			t.Errorf("Instance.SetMSContractAddr() error=nil, want %s", err.Error())
		}

		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("Instance.SetMSContractAddr() did not call conn.CodeAt with expected params")
		}
	})

	t.Run("VerifyCodeAt_NoMatch", func(t *testing.T) {

		conn := &MockContractBackend{}
		inst := Instance{
			Conn: conn,
		}

		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(dummyRuntimeBin, nil)
		err := inst.SetMSContractAddr(contractAddr)
		if err == nil {
			t.Errorf("Instance.SetMSContractAddr() error=nil, want %s", err.Error())
		}

		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("Instance.SetMSContractAddr() did not call conn.CodeAt with expected params")
		}
	})
}
func Test_Instance_VPCAddr_mock(t *testing.T) {

	t.Run("ValidAddress", func(t *testing.T) {

		contractAddr := types.HexToAddress("de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
		inst := Instance{
			vpcAddr: contractAddr,
		}

		gotContractAddr := inst.VPCAddr()

		if !bytes.Equal(contractAddr.Bytes(), gotContractAddr.Bytes()) {
			t.Errorf("Instance.VPCAddr() = %s, want %s", gotContractAddr.Hex(), contractAddr.Hex())
		}
	})

	t.Run("Empty", func(t *testing.T) {

		inst := Instance{}

		gotContractAddr := inst.VPCAddr()

		if !bytes.Equal(types.Address{}.Bytes(), gotContractAddr.Bytes()) {
			t.Errorf("Instance.VPCAddr() = %s, want empty", gotContractAddr.Hex())
		}
	})
}
func Test_Instance_SetVPCAddr_mock(t *testing.T) {

	contractAddr := types.HexToAddress("de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")

	t.Run("success", func(t *testing.T) {

		conn := &MockContractBackend{}
		inst := Instance{
			Conn: conn,
		}

		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(vpcRuntimeBin, nil)
		err := inst.SetVPCAddr(contractAddr)
		if err != nil {
			t.Errorf("Instance.VPCAddr() error=%s, want nil", err.Error())
		}

		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("Instance.VPCAddr() did not call conn.CodeAt with expected params")
		}
	})

	t.Run("VerifyCodeAt_Error", func(t *testing.T) {

		conn := &MockContractBackend{}
		inst := Instance{
			Conn: conn,
		}

		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(dummyRuntimeBin, fmt.Errorf(""))
		err := inst.SetVPCAddr(contractAddr)
		if err == nil {
			t.Errorf("Instance.VPCAddr() error=nil, want %s", err.Error())
		}

		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("Instance.VPCAddr() did not call conn.CodeAt with expected params")
		}
	})

	t.Run("VerifyCodeAt_NoMatch", func(t *testing.T) {

		conn := &MockContractBackend{}
		inst := Instance{
			Conn: conn,
		}

		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(dummyRuntimeBin, nil)
		err := inst.SetVPCAddr(contractAddr)
		if err == nil {
			t.Errorf("Instance.VPCAddr() error=nil, want %s", err.Error())
		}

		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("Instance.VPCAddr() did not call conn.CodeAt with expected params")
		}
	})
}

func Test_Instance_DeployVPC_mock(t *testing.T) {

	t.Run("VPCValid", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)

		txHash := types.HexToHash("0x6b8390d5866e311a0f2610ee8e633e9bb97b8bb0cd0f3ffaf7de185669fc5bad")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()
		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(vpcRuntimeBin, nil)
		conn.On("Commit").Return()

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
		}

		err := inst.DeployVPC()

		//Assert on results
		if err != nil {
			t.Fatalf("DeployVPC() error = %v, want nil", err)
		}
		if !bytes.Equal(inst.VPCAddr().Bytes(), contractAddr.Bytes()) {
			t.Errorf("Instance.DeployVPC() Instance.VPCAddr is set to %s, expected %s", inst.VPCAddr().String(), contractAddr.String())
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.DeployVPC() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.DeployVPC() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.DeployVPC() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.DeployVPC() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.DeployVPC() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.DeployVPC() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "CodeAt", 1) {
			t.Errorf("Instance.DeployVPC() - CodeAt() was not called")
		}

	})

	t.Run("Deploy_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(fmt.Errorf(""))

		conn.On("Commit").Return()

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
		}

		err := inst.DeployVPC()

		//Assert on results
		if err == nil {
			t.Errorf("Instance.DeployVPC() error = nil, want %v", err)
		}
		if !bytes.Equal(inst.VPCAddr().Bytes(), types.Address{}.Bytes()) {
			t.Errorf("Instance.DeployVPC() Instance.VPCAddr is set to %s, expected %s", inst.VPCAddr().String(), contractAddr.String())
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.DeployVPC() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.DeployVPC() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.DeployVPC() - SendTransaction() was not called")
		}
	})

	t.Run("SetVPCAddr_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)

		txHash := types.HexToHash("0x6b8390d5866e311a0f2610ee8e633e9bb97b8bb0cd0f3ffaf7de185669fc5bad")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()
		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(dummyRuntimeBin, nil)
		conn.On("Commit").Return()

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
		}

		err := inst.DeployVPC()

		//Assert on results
		if err == nil {
			t.Errorf("Instance.DeployVPC() error = nil, want %v", err)
		}
		if !bytes.Equal(inst.VPCAddr().Bytes(), types.Address{}.Bytes()) {
			t.Errorf("Instance.DeployVPC() Instance.VPCAddr is set to %s, expected %s", inst.VPCAddr().String(), contractAddr.String())
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.DeployVPC() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.DeployVPC() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.DeployVPC() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.DeployVPC() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.DeployVPC() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.DeployVPC() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "CodeAt", 1) {
			t.Errorf("Instance.DeployVPC() - CodeAt() was not called")
		}

	})
}

func Test_Instance_DeployMSContract_mock(t *testing.T) {

	t.Run("MSContractValid", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)

		txHash := types.HexToHash("0x07ea556186d043aa1c086b69b7e5a97939a7d10651ba33299b1913ee4ccac667")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()
		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(msContractRuntimeBin, nil)
		conn.On("Commit").Return()

		conn.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).Return(&dummySubscription{}, nil)
		conn.On("FilterLogs", mock.Anything, mock.Anything).Return([]ethereumTypes.Log{}, nil)

		vpcInst, err := contract.NewVPC(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate vpc error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			VPCInst:           vpcInst,
		}

		err = inst.DeployMSContract(aliceID.OnChainID, bobID.OnChainID)
		time.Sleep(100 * time.Millisecond) //Filter event occurs in go-routine and hence may take some time

		//Assert on results
		if err != nil {
			t.Fatalf("Instance.DeployMSContract() error = %v, want nil", err)
		}
		if !bytes.Equal(inst.MSContractAddr().Bytes(), contractAddr.Bytes()) {
			t.Errorf("Instance.DeployMSContract() Instance.MSContractAddr is set to %s, expected %s", inst.VPCAddr().String(), contractAddr.String())
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.DeployMSContract() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.DeployMSContract() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.DeployMSContract() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.DeployMSContract() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.DeployMSContract() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.DeployMSContract() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "CodeAt", 1) {
			t.Errorf("Instance.DeployMSContract() - CodeAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SubscribeFilterLogs", 7) {
			t.Errorf("Instance.DeployMSContract() - SubscribeFilterLogs() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "FilterLogs", 1) {
			t.Errorf("Instance.DeployMSContract() - FilterLogs() was not called")
		}

	})

	t.Run("Deploy_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(fmt.Errorf(""))

		conn.On("Commit").Return()

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
		}

		err := inst.DeployMSContract(aliceID.OnChainID, bobID.OnChainID)

		//Assert on results
		if err == nil {
			t.Errorf("Instance.DeployMSContract() error = nil, want %v", err)
		}
		if !bytes.Equal(inst.MSContractAddr().Bytes(), types.Address{}.Bytes()) {
			t.Errorf("Instance.DeployMSContract() Instance.MSContractAddr is set to %s, expected %s", inst.VPCAddr().String(), contractAddr.String())
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.DeployMSContract() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.DeployMSContract() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.DeployMSContract() - SendTransaction() was not called")
		}
	})

	t.Run("SetVPCAddr_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)

		txHash := types.HexToHash("0x07ea556186d043aa1c086b69b7e5a97939a7d10651ba33299b1913ee4ccac667")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()
		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(dummyRuntimeBin, nil)
		conn.On("Commit").Return()

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
		}

		err := inst.DeployMSContract(aliceID.OnChainID, bobID.OnChainID)

		//Assert on results
		if err == nil {
			t.Errorf("Instance.DeployMSContract() error = nil, want %v", err)
		}
		if !bytes.Equal(inst.VPCAddr().Bytes(), types.Address{}.Bytes()) {
			t.Errorf("Instance.DeployMSContract() Instance.VPCAddr is set to %s, expected %s", inst.VPCAddr().String(), contractAddr.String())
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.DeployMSContract() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.DeployMSContract() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.DeployMSContract() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.DeployMSContract() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.DeployMSContract() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.DeployMSContract() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "CodeAt", 1) {
			t.Errorf("Instance.DeployMSContract() - CodeAt() was not called")
		}

	})
	t.Run("InitializeEvents_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)

		txHash := types.HexToHash("0x07ea556186d043aa1c086b69b7e5a97939a7d10651ba33299b1913ee4ccac667")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()
		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(msContractRuntimeBin, nil)
		conn.On("Commit").Return()

		conn.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).Return(&dummySubscription{}, fmt.Errorf(""))

		vpcInst, err := contract.NewVPC(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate vpc error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			VPCInst:           vpcInst,
		}

		err = inst.DeployMSContract(aliceID.OnChainID, bobID.OnChainID)

		//Assert on results
		if err == nil {
			t.Errorf("Instance.DeployMSContract() error = nil, want not nil")
		}
		if !bytes.Equal(inst.MSContractAddr().Bytes(), contractAddr.Bytes()) {
			t.Errorf("Instance.DeployMSContract() Instance.MSContractAddr is set to %s, expected %s", inst.VPCAddr().String(), contractAddr.String())
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.DeployMSContract() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.DeployMSContract() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.DeployMSContract() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.DeployMSContract() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.DeployMSContract() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.DeployMSContract() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "CodeAt", 1) {
			t.Errorf("Instance.DeployMSContract() - CodeAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SubscribeFilterLogs", 1) {
			t.Errorf("Instance.DeployMSContract() - SubscribeFilterLogs() was not called")
		}

	})
}

func Test_Instance_Confirm_mock(t *testing.T) {

	t.Run("Valid", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusSuccessful}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x03441d4f78a0359d5a72cf9eb1c01bbb240af9113b78e0c7b9e1f29d479014df")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, nil)

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Confirm(big.NewInt(10))

		//Assert on results
		if err != nil {
			t.Fatalf("Instance.Confirm() error = %v, want nil", err)
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Confirm() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.Confirm() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.Confirm() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.Confirm() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.Confirm() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.Confirm() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.Confirm() - TransactionReceipt() was not called")
		}

	})
	t.Run("MakeTransaction_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), fmt.Errorf(""))

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Confirm(big.NewInt(10))

		//Assert on results
		if err == nil {
			t.Errorf("Instance.Confirm() error = nil, want non nil")
			return
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Confirm() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 0) {
			t.Errorf("Instance.Confirm() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 0) {
			t.Errorf("Instance.Confirm() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 0) {
			t.Errorf("Instance.Confirm() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 0) {
			t.Errorf("Instance.Confirm() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 0) {
			t.Errorf("Instance.Confirm() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.Confirm() - TransactionReceipt() was not called")
		}

	})
	t.Run("SendTransaction_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(fmt.Errorf(""))
		conn.On("Commit").Return()

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Confirm(big.NewInt(10))

		//Assert on results
		if err == nil {
			t.Errorf("Instance.Confirm() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Confirm() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.Confirm() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.Confirm() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 0) {
			t.Errorf("Instance.Confirm() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 0) {
			t.Errorf("Instance.Confirm() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 0) {
			t.Errorf("Instance.Confirm() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.Confirm() - TransactionReceipt() was not called")
		}

	})

	t.Run("WaitTillTxMined_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x03441d4f78a0359d5a72cf9eb1c01bbb240af9113b78e0c7b9e1f29d479014df")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, fmt.Errorf("")).Once()

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Confirm(big.NewInt(10))

		//Assert on results
		if err == nil {
			t.Errorf("Instance.Confirm() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Confirm() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.Confirm() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.Confirm() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.Confirm() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.Confirm() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.Confirm() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.Confirm() - TransactionReceipt() was not called")
		}

	})
	t.Run("TransactionReceipt_Status_Failure", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusFailed}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x03441d4f78a0359d5a72cf9eb1c01bbb240af9113b78e0c7b9e1f29d479014df")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, nil)

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Confirm(big.NewInt(10))

		//Assert on results
		if err == nil {
			t.Errorf("Instance.Confirm() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Confirm() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.Confirm() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.Confirm() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.Confirm() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.Confirm() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.Confirm() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.Confirm() - TransactionReceipt() was not called")
		}

	})
	t.Run("TransactionReceipt_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusSuccessful}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x03441d4f78a0359d5a72cf9eb1c01bbb240af9113b78e0c7b9e1f29d479014df")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, fmt.Errorf(""))

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Confirm(big.NewInt(10))

		//Assert on results
		if err == nil {
			t.Errorf("Instance.Confirm() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Confirm() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.Confirm() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.Confirm() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.Confirm() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.Confirm() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.Confirm() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.Confirm() - TransactionReceipt() was not called")
		}

	})
}

func Test_Instance_StateRegister_mock(t *testing.T) {

	t.Run("Valid", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusSuccessful}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x056812ac01cc5faf3825a6446be0a71b39305f364abfee76bea6797e9ad0921a")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, nil)

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.StateRegister(big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err != nil {
			t.Fatalf("Instance.StateRegister() error = %v, want nil", err)
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.StateRegister() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.StateRegister() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.StateRegister() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.StateRegister() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.StateRegister() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.StateRegister() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.StateRegister() - TransactionReceipt() was not called")
		}

	})
	t.Run("MakeTransaction_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), fmt.Errorf(""))

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.StateRegister(big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.StateRegister() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.StateRegister() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 0) {
			t.Errorf("Instance.StateRegister() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 0) {
			t.Errorf("Instance.StateRegister() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 0) {
			t.Errorf("Instance.StateRegister() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 0) {
			t.Errorf("Instance.StateRegister() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 0) {
			t.Errorf("Instance.StateRegister() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.StateRegister() - TransactionReceipt() was not called")
		}

	})
	t.Run("SendTransaction_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(fmt.Errorf(""))
		conn.On("Commit").Return()

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.StateRegister(big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.StateRegister() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.StateRegister() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.StateRegister() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.StateRegister() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 0) {
			t.Errorf("Instance.StateRegister() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 0) {
			t.Errorf("Instance.StateRegister() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 0) {
			t.Errorf("Instance.StateRegister() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.StateRegister() - TransactionReceipt() was not called")
		}

	})

	t.Run("WaitTillTxMined_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x056812ac01cc5faf3825a6446be0a71b39305f364abfee76bea6797e9ad0921a")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, fmt.Errorf("")).Once()

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.StateRegister(big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.StateRegister() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.StateRegister() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.StateRegister() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.StateRegister() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.StateRegister() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.StateRegister() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.StateRegister() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.StateRegister() - TransactionReceipt() was not called")
		}

	})
	t.Run("TransactionReceipt_Status_Failure", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusFailed}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x056812ac01cc5faf3825a6446be0a71b39305f364abfee76bea6797e9ad0921a")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, nil)

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.StateRegister(big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.StateRegister() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.StateRegister() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.StateRegister() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.StateRegister() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.StateRegister() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.StateRegister() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.StateRegister() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.StateRegister() - TransactionReceipt() was not called")
		}

	})
	t.Run("TransactionReceipt_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusSuccessful}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x056812ac01cc5faf3825a6446be0a71b39305f364abfee76bea6797e9ad0921a")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, fmt.Errorf(""))

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.StateRegister(big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.StateRegister() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.StateRegister() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.StateRegister() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.StateRegister() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.StateRegister() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.StateRegister() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.StateRegister() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.StateRegister() - TransactionReceipt() was not called")
		}

	})
}

func Test_Instance_Execute_mock(t *testing.T) {

	t.Run("Valid", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusSuccessful}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0xe7a42468aba16b0959071b82c30561639b8747510c7b210e0f89c907b37d8b27")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, nil)

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Execute(types.Address{}, types.Address{})

		//Assert on results
		if err != nil {
			t.Fatalf("Instance.Execute() error = %v, want nil", err)
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Execute() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.Execute() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.Execute() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.Execute() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.Execute() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.Execute() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.Execute() - TransactionReceipt() was not called")
		}

	})
	t.Run("MakeTransaction_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), fmt.Errorf(""))

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Execute(types.Address{}, types.Address{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.Execute() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Execute() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 0) {
			t.Errorf("Instance.Execute() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 0) {
			t.Errorf("Instance.Execute() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 0) {
			t.Errorf("Instance.Execute() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 0) {
			t.Errorf("Instance.Execute() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 0) {
			t.Errorf("Instance.Execute() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.Execute() - TransactionReceipt() was not called")
		}

	})
	t.Run("SendTransaction_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(fmt.Errorf(""))
		conn.On("Commit").Return()

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Execute(types.Address{}, types.Address{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.Execute() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Execute() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.Execute() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.Execute() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 0) {
			t.Errorf("Instance.Execute() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 0) {
			t.Errorf("Instance.Execute() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 0) {
			t.Errorf("Instance.Execute() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.Execute() - TransactionReceipt() was not called")
		}

	})

	t.Run("WaitTillTxMined_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0xe7a42468aba16b0959071b82c30561639b8747510c7b210e0f89c907b37d8b27")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, fmt.Errorf("")).Once()

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Execute(types.Address{}, types.Address{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.Execute() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Execute() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.Execute() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.Execute() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.Execute() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.Execute() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.Execute() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.Execute() - TransactionReceipt() was not called")
		}

	})
	t.Run("TransactionReceipt_Status_Failure", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusFailed}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0xe7a42468aba16b0959071b82c30561639b8747510c7b210e0f89c907b37d8b27")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, nil)

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Execute(types.Address{}, types.Address{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.Execute() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Execute() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.Execute() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.Execute() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.Execute() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.Execute() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.Execute() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.Execute() - TransactionReceipt() was not called")
		}

	})
	t.Run("TransactionReceipt_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusSuccessful}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0xe7a42468aba16b0959071b82c30561639b8747510c7b210e0f89c907b37d8b27")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, fmt.Errorf(""))

		msContractInst, err := contract.NewMSContract(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate msContract error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
		}

		err = inst.Execute(types.Address{}, types.Address{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.Execute() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.Execute() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.Execute() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.Execute() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.Execute() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.Execute() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.Execute() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.Execute() - TransactionReceipt() was not called")
		}

	})
}

func Test_Instance_VPCClose_mock(t *testing.T) {

	t.Run("Valid", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusSuccessful}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x8e6c7f9ac64544c6a81ca191a50781d472b9aa3a5ce667ae51681441fd2269b7")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, nil)

		vpcInst, err := contract.NewVPC(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate vpc error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			VPCInst:           vpcInst,
		}

		err = inst.VPCClose(big.NewInt(0), big.NewInt(0), types.Address{}, types.Address{}, big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err != nil {
			t.Fatalf("Instance.VPCClose() error = %v, want nil", err)
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.VPCClose() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.VPCClose() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.VPCClose() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.VPCClose() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.VPCClose() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.VPCClose() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.VPCClose() - TransactionReceipt() was not called")
		}

	})
	t.Run("MakeTransaction_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), fmt.Errorf(""))

		vpcInst, err := contract.NewVPC(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate vpc error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			VPCInst:           vpcInst,
		}

		err = inst.VPCClose(big.NewInt(0), big.NewInt(0), types.Address{}, types.Address{}, big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.VPCClose() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.VPCClose() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 0) {
			t.Errorf("Instance.VPCClose() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 0) {
			t.Errorf("Instance.VPCClose() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 0) {
			t.Errorf("Instance.VPCClose() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 0) {
			t.Errorf("Instance.VPCClose() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 0) {
			t.Errorf("Instance.VPCClose() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.VPCClose() - TransactionReceipt() was not called")
		}

	})
	t.Run("SendTransaction_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(fmt.Errorf(""))
		conn.On("Commit").Return()

		vpcInst, err := contract.NewVPC(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate vpc error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			VPCInst:           vpcInst,
		}

		err = inst.VPCClose(big.NewInt(0), big.NewInt(0), types.Address{}, types.Address{}, big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.VPCClose() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.VPCClose() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.VPCClose() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.VPCClose() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 0) {
			t.Errorf("Instance.VPCClose() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 0) {
			t.Errorf("Instance.VPCClose() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 0) {
			t.Errorf("Instance.VPCClose() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.VPCClose() - TransactionReceipt() was not called")
		}

	})

	t.Run("WaitTillTxMined_Error", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x8e6c7f9ac64544c6a81ca191a50781d472b9aa3a5ce667ae51681441fd2269b7")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, fmt.Errorf("")).Once()

		vpcInst, err := contract.NewVPC(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate vpc error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			VPCInst:           vpcInst,
		}

		err = inst.VPCClose(big.NewInt(0), big.NewInt(0), types.Address{}, types.Address{}, big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.VPCClose() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.VPCClose() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.VPCClose() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.VPCClose() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.VPCClose() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.VPCClose() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.VPCClose() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 0) {
			t.Errorf("Instance.VPCClose() - TransactionReceipt() was not called")
		}

	})
	t.Run("TransactionReceipt_Status_Failure", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusFailed}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x8e6c7f9ac64544c6a81ca191a50781d472b9aa3a5ce667ae51681441fd2269b7")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, nil)

		vpcInst, err := contract.NewVPC(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate vpc error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			VPCInst:           vpcInst,
		}

		err = inst.VPCClose(big.NewInt(0), big.NewInt(0), types.Address{}, types.Address{}, big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.VPCClose() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.VPCClose() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.VPCClose() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.VPCClose() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.VPCClose() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.VPCClose() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.VPCClose() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.VPCClose() - TransactionReceipt() was not called")
		}

	})
	t.Run("TransactionReceipt_Error", func(t *testing.T) {
		idWithCredentials := identity.OffChainID{
			OnChainID: aliceID.OnChainID,
			KeyStore:  testKeyStore,
			Password:  alicePassword,
		}
		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")
		receipt := &ethereumTypes.Receipt{Status: ethereumTypes.ReceiptStatusSuccessful}

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
		conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
		conn.On("Commit").Return()

		txHash := types.HexToHash("0x8e6c7f9ac64544c6a81ca191a50781d472b9aa3a5ce667ae51681441fd2269b7")
		conn.On("BackendType").Return(adapter.Real)
		conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

		conn.On("TransactionReceipt", context.Background(), txHash.Hash).Return(receipt, fmt.Errorf(""))

		vpcInst, err := contract.NewVPC(contractAddr.Address, conn)
		if err != nil {
			t.Errorf("Instantiate vpc error -" + err.Error())
		}

		inst := Instance{
			Conn:              conn,
			OwnerID:           idWithCredentials,
			libSignaturesAddr: contractAddr,
			VPCInst:           vpcInst,
		}

		err = inst.VPCClose(big.NewInt(0), big.NewInt(0), types.Address{}, types.Address{}, big.NewInt(0), big.NewInt(0), []byte{}, []byte{})

		//Assert on results
		if err == nil {
			t.Errorf("Instance.VPCClose() error = nil, want non nil")
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("Instance.VPCClose() - PendingNonceAt() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("Instance.VPCClose() - SuggestGasPrice() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
			t.Errorf("Instance.VPCClose() - SendTransaction() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "Commit", 1) {
			t.Errorf("Instance.VPCClose() - Commit() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "BackendType", 1) {
			t.Errorf("Instance.VPCClose() - BackendType() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("Instance.VPCClose() - TransactionByHash() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionReceipt", 1) {
			t.Errorf("Instance.VPCClose() - TransactionReceipt() was not called")
		}

	})
}

func Test_Instance_States_mock(t *testing.T) {

	tests := []struct {
		name         string
		ABIHexString string
		wantErr      bool
	}{
		{
			name: "Success",
			//data obtained from logging an actual call
			ABIHexString: "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001158e460913d00000000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000002a800000000000000000000000000000000000000000000000000000000000005000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001",
			wantErr:      false,
		},
		{
			name: "ABIHexParseError",
			//data obtained from logging an actual call and changing last byte from 1 to 3
			//this byte represents boolean value and hence can take only 0/1.
			//hence setting it to 3 will cause a parse error
			ABIHexString: "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001158e460913d00000000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000002a800000000000000000000000000000000000000000000000000000000000005000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000003",
			wantErr:      true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			d, err := hex.DecodeString(tt.ABIHexString)
			if err != nil {
				t.Fatalf("Setup : ABIHex DecodeString err = %v", err)
			}

			//Setup mock
			conn := &MockContractBackend{}
			conn.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return(d, nil)

			vpcInst, err := contract.NewVPC(types.Address{}.Address, conn)
			if err != nil {
				t.Fatalf("Setup : Instantiate vpc error -" + err.Error())
			}

			inst := Instance{
				Conn:    conn,
				VPCInst: vpcInst,
			}

			_, err = inst.States()

			//Assert on results
			if (err != nil) != tt.wantErr {
				t.Fatalf("State() error = %v, want nil", err)
			}
			//Assert on mock calls
			if !conn.AssertNumberOfCalls(t, "CallContract", 1) {
				t.Errorf("Instance.States() - CallContract() was not called")
			}

		})
	}

}

func Test_Instance_InitializeEventsChan_mock(t *testing.T) {

	t.Run("filter_error", func(t *testing.T) {

		contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

		//Setup mock
		conn := &MockContractBackend{}
		conn.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).Return(&dummySubscription{}, nil).Times(7)
		conn.On("FilterLogs", mock.Anything, mock.Anything).Return([]ethereumTypes.Log{}, fmt.Errorf("")).Times(1)

		vpcInst, _ := contract.NewVPC(contractAddr.Address, conn)
		msContractInst, _ := contract.NewMSContract(contractAddr.Address, conn)

		inst := Instance{
			Conn:              conn,
			libSignaturesAddr: contractAddr,
			MSContractInst:    msContractInst,
			VPCInst:           vpcInst,
		}

		_, err := inst.InitializeEventsChan()
		time.Sleep(100 * time.Millisecond) //Filter event occurs in go-routine and hence may take some time

		//Assert on results
		if err != nil {
			t.Fatalf("Instance.InitializeEventsChan() error = %v, want nil", err)
		}

		//Assert on mock calls
		if !conn.AssertNumberOfCalls(t, "SubscribeFilterLogs", 7) {
			t.Errorf("Instance.InitializeEventsChan() - SubscribeFilterLogs() was not called")
		}
		if !conn.AssertNumberOfCalls(t, "FilterLogs", 1) {
			t.Errorf("Instance.InitializeEventsChan() - FilterLogs() was not called")
		}

	})

	tests := []struct {
		name                string
		countSuccessfulSub  int
		countFailingSub     int
		countSubCalls       int
		countFilterLogCalls int
		wantErr             bool
	}{
		{"valid", 7, 0, 7, 1, false},
		{"subscribe_event_1_error", 0, 1, 1, 0, true},
		{"subscribe_event_2_error", 1, 1, 2, 1, true},
		{"subscribe_event_3_error", 2, 1, 3, 1, true},
		{"subscribe_event_4_error", 3, 1, 4, 1, true},
		{"subscribe_event_5_error", 4, 1, 5, 1, true},
		{"subscribe_event_6_error", 5, 1, 6, 1, true},
		{"subscribe_event_7_error", 6, 1, 7, 1, true},
	}

	for _, tt := range tests {
		t.Run("subscribe_event_2_error", func(t *testing.T) {

			contractAddr := types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D")

			//Setup mock
			conn := &MockContractBackend{}
			if tt.countSuccessfulSub != 0 {
				conn.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).Return(&dummySubscription{}, nil).Times(tt.countSuccessfulSub)
			}
			if tt.countFailingSub != 0 {
				conn.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).Return(&dummySubscription{}, fmt.Errorf("")).Times(tt.countFailingSub)
			}
			if tt.countFilterLogCalls != 0 {
				conn.On("FilterLogs", mock.Anything, mock.Anything).Return([]ethereumTypes.Log{}, nil).Times(tt.countFilterLogCalls)
			}

			vpcInst, _ := contract.NewVPC(contractAddr.Address, conn)
			msContractInst, _ := contract.NewMSContract(contractAddr.Address, conn)

			inst := Instance{
				Conn:              conn,
				libSignaturesAddr: contractAddr,
				MSContractInst:    msContractInst,
				VPCInst:           vpcInst,
			}

			_, err := inst.InitializeEventsChan()
			time.Sleep(100 * time.Millisecond) //Filter event occurs in go-routine and hence may take some time

			//Assert on results
			if tt.wantErr != (err != nil) {
				t.Fatalf("Instance.InitializeEventsChan() error = %v, wantErr=%t", err, tt.wantErr)
			}

			//Assert on mock calls
			if !conn.AssertNumberOfCalls(t, "SubscribeFilterLogs", tt.countSubCalls) {
				t.Errorf("Instance.InitializeEventsChan() - SubscribeFilterLogs() was not called")
			}
			if !conn.AssertNumberOfCalls(t, "FilterLogs", tt.countFilterLogCalls) {
				t.Errorf("Instance.InitializeEventsChan() - FilterLogs() was not called")
			}
		})
	}
}

func Test_SetupLibSignatures_mock(t *testing.T) {
	tests := []struct {
		name                     string
		libSigAddr               types.Address
		sessionOwner             identity.OffChainID
		mockContractToSetup      []byte
		wantDeployLibSig         bool
		verifyCodeAtReturnsError error
		wantErr                  bool
		wantMatch                bool
	}{
		{
			name: "validLibSigAddr",
			sessionOwner: identity.OffChainID{
				OnChainID: aliceID.OnChainID,
				Password:  alicePassword,
				KeyStore:  testKeyStore,
			},
			libSigAddr:               types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D"),
			mockContractToSetup:      libSignaturesRuntimeBin,
			verifyCodeAtReturnsError: nil,
			wantDeployLibSig:         true,
			wantErr:                  false,
			wantMatch:                true,
		},
		{
			name: "invalidLibSigAddr",
			sessionOwner: identity.OffChainID{
				OnChainID: aliceID.OnChainID,
				Password:  alicePassword,
				KeyStore:  testKeyStore,
			},
			libSigAddr:               types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D"),
			mockContractToSetup:      vpcRuntimeBin,
			verifyCodeAtReturnsError: nil,
			wantDeployLibSig:         true,
			wantErr:                  false,
			wantMatch:                true,
		},
		{
			name: "verifyCodeAt_Error",
			sessionOwner: identity.OffChainID{
				OnChainID: aliceID.OnChainID,
				Password:  alicePassword,
				KeyStore:  testKeyStore,
			},
			libSigAddr:               types.HexToAddress("0x21c7c9b5aC63D9930d6410Ed29499CBD6AFcee4D"),
			mockContractToSetup:      vpcRuntimeBin,
			verifyCodeAtReturnsError: fmt.Errorf(""),
			wantDeployLibSig:         true,
			wantErr:                  false,
			wantMatch:                true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			conn := &MockContractBackend{}
			conn.On("CodeAt", context.Background(), tt.libSigAddr.Address, (*big.Int)(nil)).Return(tt.mockContractToSetup, tt.verifyCodeAtReturnsError)

			if tt.wantDeployLibSig {

				conn.On("PendingNonceAt", context.Background(), tt.sessionOwner.OnChainID.Address).Return(uint64(0), nil)
				conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
				conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)

				conn.On("BackendType").Return(adapter.Real)
				conn.On("TransactionByHash", context.Background(), mock.AnythingOfType("common.Hash")).Return(nil, false, nil).Once()
				conn.On("Commit").Return()
			}

			_, err := SetupLibSignatures(tt.libSigAddr, conn, tt.sessionOwner)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SetupLibSignatures() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_NewSimulatedInstance(t *testing.T) {

	conn := &MockContractBackend{}
	ownerID := aliceID

	gotInstance := NewInstance(conn, ownerID)

	if !reflect.DeepEqual(gotInstance.Conn, conn) {
		t.Errorf("NewInstance() Instance.Conn not assigned properly")
	}
	if !identity.Equal(gotInstance.OwnerID, ownerID) {
		t.Errorf("NewInstance() Instance.Conn not assigned properly")
	}
}
