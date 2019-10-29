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

package adapter

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/mock"

	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
)

func Test_MakeFilterOpts(t *testing.T) {
	type args struct {
		context    context.Context
		startBlock *uint64
		endBlock   *uint64
	}
	tests := []struct {
		name string
		args args
		want *bind.FilterOpts
	}{
		{
			name: "valid-1",
			args: args{
				context:    nil,
				startBlock: new(uint64),
				endBlock:   new(uint64),
			},
			want: &bind.FilterOpts{
				Start:   0,
				End:     new(uint64),
				Context: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeFilterOpts(tt.args.context, tt.args.startBlock, tt.args.endBlock)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeFilterOpts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_MakeCallOpts(t *testing.T) {
	type args struct {
		context        context.Context
		includePending bool
		address        types.Address
	}
	tests := []struct {
		name string
		args args
		want *bind.CallOpts
	}{
		{
			name: "valid-1",
			args: args{
				context:        nil,
				includePending: true,
				address:        aliceID.OnChainID,
			},
			want: &bind.CallOpts{
				Pending: true,
				From:    aliceID.OnChainID.Address,
				Context: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeCallOpts(tt.args.context, tt.args.includePending, tt.args.address)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeCallOpts() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_MakeTransactOpts(t *testing.T) {
	type args struct {
		idWithCredentials identity.OffChainID
		valueInWei        *big.Int
		gasLimit          uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "validOpts-1",
			args: args{
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
				valueInWei: big.NewInt(0),
				gasLimit:   uint64(40e5)},
			wantErr: false,
		},
		{
			name: "noCredentialsInId",
			args: args{
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
				},
				valueInWei: big.NewInt(0),
				gasLimit:   uint64(40e5)},
			wantErr: true,
		},
		{
			name: "wrongPassword",
			args: args{
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword + "xyz",
				},
				valueInWei: big.NewInt(0),
				gasLimit:   uint64(40e5)},
			wantErr: true,
		},
	}

	conn := NewSimulatedBackend(balanceList)

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			_, err := MakeTransactOpts(conn, tt.args.idWithCredentials, tt.args.valueInWei, tt.args.gasLimit)

			if (err != nil) != tt.wantErr {
				t.Errorf("MakeTransactOpts() error = %+v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_MakeTransactOpts_Integration(t *testing.T) {
	type args struct {
		idWithCredentials identity.OffChainID
		valueInWei        *big.Int
		gasLimit          uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "validOpts-1",
			args: args{
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
				valueInWei: big.NewInt(0),
				gasLimit:   uint64(40e5)},
			wantErr: false,
		},
		{
			name: "noCredentialsInId",
			args: args{
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
				},
				valueInWei: big.NewInt(0),
				gasLimit:   uint64(40e5)},
			wantErr: true,
		},
		{
			name: "wrongPassword",
			args: args{
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword + "xyz",
				},
				valueInWei: big.NewInt(0),
				gasLimit:   uint64(40e5)},
			wantErr: true,
		},
	}

	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	//Setup
	conn, err := NewRealBackend(ethereumNodeURL)
	if err != nil {
		t.Fatalf("Error setting up connection to blockchain node.\n%v", err)
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			_, err := MakeTransactOpts(conn, tt.args.idWithCredentials, tt.args.valueInWei, tt.args.gasLimit)

			if (err != nil) != tt.wantErr {
				t.Errorf("MakeTransactOpts() error = %+v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func Test_DeployContract(t *testing.T) {
	type args struct {
		contract          contract.Handler
		params            []interface{}
		idWithCredentials identity.OffChainID
	}

	tests := []struct {
		name string
		args args

		wantErr  bool
		wantType reflect.Type
	}{
		{
			name: "libSignatures",
			args: args{
				contract: contract.Store.LibSignatures(),
				params:   nil,
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			wantErr:  false,
			wantType: reflect.TypeOf((*contract.LibSignatures)(nil)),
		},
		{
			name: "VPC",
			args: args{
				contract: contract.Store.VPC(),
				params:   nil, //params are defined before deploy contract call
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			wantErr:  false,
			wantType: reflect.TypeOf((*contract.VPC)(nil)),
		},
		{
			name: "MSContract",
			args: args{
				contract: contract.Store.MSContract(),
				params:   nil, //params are defined before deploy contract call
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			wantErr:  false,
			wantType: reflect.TypeOf((*contract.MSContract)(nil)),
		},
	}
	conn := NewSimulatedBackend(balanceList)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.args.contract == contract.Store.VPC() {
				libSignAddr, _, _, err := DeployContract(contract.Store.LibSignatures(), conn, nil, tt.args.idWithCredentials)
				if err != nil {
					t.Fatalf("Error setting up libSign for vpc -%s", err)
				}
				conn.Commit()
				tt.args.params = nil
				tt.args.params = append(tt.args.params, libSignAddr)
			}
			if tt.args.contract == contract.Store.MSContract() {
				libSignAddr, _, _, err := DeployContract(contract.Store.LibSignatures(), conn, nil, tt.args.idWithCredentials)
				if err != nil {
					t.Fatalf("Error setting up libSign for vpc -%s", err)
				}
				conn.Commit()
				tt.args.params = nil
				tt.args.params = append(tt.args.params, libSignAddr, aliceID.OnChainID, bobID.OnChainID)
			}
			_, tx, gotHandler, err := DeployContract(tt.args.contract, conn, tt.args.params, tt.args.idWithCredentials)
			conn.Commit()

			if (err != nil) != tt.wantErr {
				t.Fatalf("DeployContract() error = %v, wantErr %v", err, tt.wantErr)
			}

			if reflect.TypeOf(gotHandler) != tt.wantType {
				t.Errorf("DeployContract() %s gotHandlerType = %v, want %v", tt.name, reflect.TypeOf(gotHandler), tt.wantType)
			}
			_, err = conn.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				t.Errorf("DeployContract() %s error reading transaction receipt = %v", tt.name, err)
			}

		})
	}
}

func Test_DeployContract_Integration(t *testing.T) {
	type args struct {
		contract          contract.Handler
		params            []interface{}
		idWithCredentials identity.OffChainID
	}

	tests := []struct {
		name string
		args args

		wantHandlerTypeLib        *contract.LibSignatures
		wantHandlerTypeVpc        *contract.VPC
		wantHandlerTypeMSContract *contract.MSContract
		wantErr                   bool
	}{
		{
			name: "libSignatures",
			args: args{
				contract: contract.Store.LibSignatures(),
				params:   nil,
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			wantErr: false,
		},
		{
			name: "VPC",
			args: args{
				contract: contract.Store.VPC(),
				params:   nil, //params are defined before deploy contract call
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			wantErr: false,
		},
		{
			name: "MSContract",
			args: args{
				contract: contract.Store.MSContract(),
				params:   nil, //params are defined before deploy contract call
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			wantErr: false,
		},
	}
	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	conn, err := NewRealBackend(ethereumNodeURL)
	if err != nil {
		t.Fatalf("DeployContract() Error in setting up connection %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.args.contract == contract.Store.VPC() {
				libSignAddr := types.HexToAddress("0x847b3655B5bEB829cB3cD41C00A27648de737C39")
				tt.args.params = append(tt.args.params, libSignAddr)
			}
			if tt.args.contract == contract.Store.MSContract() {
				libSignAddr := types.HexToAddress("0x847b3655B5bEB829cB3cD41C00A27648de737C39")
				tt.args.params = append(tt.args.params, libSignAddr, aliceID.OnChainID, bobID.OnChainID)
			}
			_, tx, gotHandler, err := DeployContract(tt.args.contract, conn, tt.args.params, tt.args.idWithCredentials)

			if (err != nil) != tt.wantErr {
				t.Fatalf("DeployContract() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.args.contract == contract.Store.LibSignatures() {
				if reflect.TypeOf(gotHandler) != reflect.TypeOf(tt.wantHandlerTypeLib) {
					t.Errorf("DeployContract() %s gotHandlerType = %v, want %v", tt.name, reflect.TypeOf(gotHandler), tt.wantHandlerTypeLib)
				}
			}
			if tt.args.contract == contract.Store.VPC() {
				if reflect.TypeOf(gotHandler) != reflect.TypeOf(tt.wantHandlerTypeVpc) {
					t.Errorf("DeployContract() %s gotHandlerType = %v, want %v", tt.name, reflect.TypeOf(gotHandler), tt.wantHandlerTypeVpc)
				}
			}
			if tt.args.contract == contract.Store.MSContract() {
				if reflect.TypeOf(gotHandler) != reflect.TypeOf(tt.wantHandlerTypeMSContract) {
					t.Errorf("DeployContract() %s gotHandlerType = %v, want %v", tt.name, reflect.TypeOf(gotHandler), tt.wantHandlerTypeMSContract)
				}
			}
			_, err = conn.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				t.Errorf("DeployContract() %s error reading transaction receipt = %v", tt.name, err)
			}

		})
	}
}

func Test_VerifyCodeAt(t *testing.T) {
	type args struct {
		contractAddr types.Address
		handler      contract.Handler
	}
	tests := []struct {
		name            string
		args            args
		wantMatchStatus contract.MatchStatus
		wantErr         bool
	}{
		{
			name: "libSignatures", wantMatchStatus: contract.Match, wantErr: false,
			args: args{handler: contract.Store.LibSignatures()},
		},
		{
			name: "VPC", wantMatchStatus: contract.Match, wantErr: false,
			args: args{handler: contract.Store.VPC()},
		},
		{
			name: "MSContract", wantMatchStatus: contract.Match, wantErr: false,
			args: args{handler: contract.Store.MSContract()},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			//Setup
			defaultID := identity.OffChainID{
				OnChainID: aliceID.OnChainID,
				Password:  alicePassword,
				KeyStore:  testKeyStore,
			}

			conn := NewSimulatedBackend(balanceList)

			var err error
			tt.args.contractAddr, err = setupContract(tt.args.handler, conn, defaultID)
			if err != nil {
				t.Fatalf("Error in setup : deployContract - %s - %v", tt.name, err)
			}
			conn.Commit()

			gotMatchStatus, err := VerifyCodeAt(tt.args.contractAddr, tt.args.handler.HashBinRuntimeFile, false, conn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifyCodeAt() %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if !reflect.DeepEqual(gotMatchStatus, tt.wantMatchStatus) {
				t.Errorf("VerifyCodeAt() %s = %v, want %v", tt.name, gotMatchStatus, tt.wantMatchStatus)
			}
		})
	}
}

func Test_VerifyCodeAt_Integration(t *testing.T) {
	type args struct {
		contractAddr types.Address
		handler      contract.Handler
	}
	tests := []struct {
		name            string
		args            args
		wantMatchStatus contract.MatchStatus
		wantErr         bool
	}{
		{
			name: "libSignatures", wantMatchStatus: contract.Match, wantErr: false,
			args: args{handler: contract.Store.LibSignatures()},
		},
		{
			name: "VPC", wantMatchStatus: contract.Match, wantErr: false,
			args: args{handler: contract.Store.VPC()},
		},
		{
			name: "MSContract", wantMatchStatus: contract.Match, wantErr: false,
			args: args{handler: contract.Store.MSContract()},
		},
	}

	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	//Setup
	conn, err := NewRealBackend(ethereumNodeURL)
	if err != nil {
		t.Errorf("DeployContract() Error in setting up connection %v", err)
		return
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			defaultID := identity.OffChainID{
				OnChainID: aliceID.OnChainID,
				Password:  alicePassword,
				KeyStore:  testKeyStore,
			}

			var err error
			tt.args.contractAddr, err = setupContract(tt.args.handler, conn, defaultID)
			if err != nil {
				t.Fatalf("Error in setup : deployContract - %s - %v", tt.name, err)
			}

			gotMatchStatus, err := VerifyCodeAt(tt.args.contractAddr, tt.args.handler.HashBinRuntimeFile, false, conn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifyCodeAt() %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if !reflect.DeepEqual(gotMatchStatus, tt.wantMatchStatus) {
				t.Errorf("VerifyCodeAt() %s = %v, want %v", tt.name, gotMatchStatus, tt.wantMatchStatus)
			}
		})
	}
}

func setupContract(handler contract.Handler, conn ContractBackend, defaultID identity.OffChainID) (
	deployedContractAddr types.Address, err error) {

	switch handler {

	case contract.Store.LibSignatures():
		var params []interface{}
		libSignaturesAddr, _, _, err := DeployContract(handler, conn, params, defaultID)
		if err != nil {
			return types.Address{}, fmt.Errorf("Deploy libSignatures - %v", err)
		}
		deployedContractAddr = libSignaturesAddr

	case contract.Store.VPC():
		var params []interface{}
		libSignaturesAddr, _, _, err := DeployContract(contract.Store.LibSignatures(), conn, params, defaultID)
		if err != nil {
			return types.Address{}, fmt.Errorf("Deploy libSignatures - %v", err)
		}

		params = append(params, libSignaturesAddr)
		vpcAddr, _, _, err := DeployContract(handler, conn, params, defaultID)
		if err != nil {
			return types.Address{}, fmt.Errorf("Deploy vpc - %v", err)
		}

		deployedContractAddr = vpcAddr
	case contract.Store.MSContract():
		var params []interface{}
		libSignaturesAddr, _, _, err := DeployContract(contract.Store.LibSignatures(), conn, params, defaultID)
		if err != nil {
			return types.Address{}, fmt.Errorf("Deploy libSignatures - %v", err)
		}

		params = append(params, libSignaturesAddr, aliceID.OnChainID, bobID.OnChainID)
		msContractAddr, _, _, err := DeployContract(handler, conn, params, defaultID)
		if err != nil {
			return types.Address{}, fmt.Errorf("Deploy msc - %v", err)
		}

		deployedContractAddr = msContractAddr
	}

	return deployedContractAddr, nil
}

func Test_NewRealBackend_Integration(t *testing.T) {

	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	t.Run("valid", func(t *testing.T) {
		rawURL := ethereumNodeURL
		conn, err := NewRealBackend(rawURL)

		if err != nil {
			t.Errorf("NewRealBackend() Error=%s, want nil", err.Error())
		}
		if conn == nil {
			t.Errorf("NewRealBackend() *RealBackend is nil, want non nil")
		}
	})

	t.Run("invalid_protocol", func(t *testing.T) {
		rawURL := "abc://localhost:8546"
		conn, err := NewRealBackend(rawURL)

		if err == nil {
			t.Errorf("NewRealBackend() Error is nil, want %s", err.Error())
		}
		if conn != nil {
			t.Errorf("NewRealBackend() *RealBackend is not nil, want nil")
		}
	})
}

//mock based tests
func Test_deployContract_mock(t *testing.T) {

	dummyLibSigAddress := types.Address{}
	emptyInterfaceSlice := make([]interface{}, 0)

	type args struct {
		contract          contract.Handler
		params            []interface{}
		idWithCredentials identity.OffChainID
	}

	tests := []struct {
		name string
		args args

		assertDeploy bool
		wantErr      bool
		wantType     reflect.Type
	}{
		{
			name: "libSignatures_valid",
			args: args{
				contract: contract.Store.LibSignatures(),
				params:   nil,
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: true,
			wantErr:      false,
			wantType:     reflect.TypeOf((*contract.LibSignatures)(nil)),
		},
		{
			name: "VPC_valid",
			args: args{
				contract: contract.Store.VPC(),
				params:   append(emptyInterfaceSlice, dummyLibSigAddress),
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: true,
			wantErr:      false,
			wantType:     reflect.TypeOf((*contract.VPC)(nil)),
		},
		{
			name: "VPC_no_params",
			args: args{
				contract: contract.Store.VPC(),
				params:   nil,
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: false,
			wantErr:      true,
		},
		{
			name: "VPC_param_1_not_address",
			args: args{
				contract: contract.Store.VPC(),
				params:   append(emptyInterfaceSlice, 1),
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: false,
			wantErr:      true,
		},
		{
			name: "MSContract_valid",
			args: args{
				contract: contract.Store.MSContract(),
				params:   append(emptyInterfaceSlice, dummyLibSigAddress, aliceID.OnChainID, bobID.OnChainID),
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: true,
			wantErr:      false,
			wantType:     reflect.TypeOf((*contract.MSContract)(nil)),
		},
		{
			name: "MSContract_no_params",
			args: args{
				contract: contract.Store.MSContract(),
				params:   nil,
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: false,
			wantErr:      true,
		},
		{
			name: "MSContract_params_1_not_Address",
			args: args{
				contract: contract.Store.MSContract(),
				params:   append(emptyInterfaceSlice, 1, aliceID.OnChainID, bobID.OnChainID),
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: false,
			wantErr:      true,
		},
		{
			name: "MSContract_params_2_not_Address",
			args: args{
				contract: contract.Store.MSContract(),
				params:   append(emptyInterfaceSlice, dummyLibSigAddress, 1, bobID.OnChainID),
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: false,
			wantErr:      true,
		},
		{
			name: "MSContract_params_3_not_Address",
			args: args{
				contract: contract.Store.MSContract(),
				params:   append(emptyInterfaceSlice, dummyLibSigAddress, aliceID.OnChainID, 1),
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: false,
			wantErr:      true,
		},
		{
			name: "invalid_handler",
			args: args{
				contract: contract.Handler{},
				params:   append(emptyInterfaceSlice, dummyLibSigAddress, aliceID.OnChainID, 1),
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: false,
			wantErr:      true,
		}, {
			name: "credentials_unsert",
			args: args{
				contract: contract.Handler{},
				params:   nil,
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
				},
			},
			assertDeploy: false,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			conn := &MockContractBackend{}

			//Make transaction opts
			conn.On("PendingNonceAt", context.Background(), tt.args.idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
			conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
			conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)

			_, _, gotHandler, err := deployContract(tt.args.contract, conn, tt.args.params, tt.args.idWithCredentials)

			if (err != nil) != tt.wantErr {
				t.Fatalf("deployContract() error = %v, wantErr %v", err, tt.wantErr)
			}

			if reflect.TypeOf(gotHandler) != tt.wantType {
				t.Errorf("deployContract() %s gotHandlerType = %v, want %v", tt.name, reflect.TypeOf(gotHandler), tt.wantType)
			}

			if tt.assertDeploy && !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
				t.Errorf("deployContract() - SendTransaction() was not called")
			}
		})
	}
}
func Test_DeployContract_mock(t *testing.T) {

	dummyLibSigAddress := types.Address{}
	emptyInterfaceSlice := make([]interface{}, 0)

	type args struct {
		contract          contract.Handler
		params            []interface{}
		idWithCredentials identity.OffChainID
	}

	tests := []struct {
		name string
		args args

		assertDeploy bool
		wantErr      bool
		wantType     reflect.Type
	}{
		{
			name: "valid",
			args: args{
				contract: contract.Store.VPC(),
				params:   append(emptyInterfaceSlice, dummyLibSigAddress),
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: true,
			wantErr:      false,
			wantType:     reflect.TypeOf((*contract.VPC)(nil)),
		},
		{
			name: "deployContract_Error",
			args: args{
				contract: contract.Store.VPC(),
				params:   nil,
				idWithCredentials: identity.OffChainID{
					OnChainID: aliceID.OnChainID,
					KeyStore:  testKeyStore,
					Password:  alicePassword,
				},
			},
			assertDeploy: false,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			conn := &MockContractBackend{}

			//deployContract
			accountAddress := tt.args.idWithCredentials.OnChainID.Address
			accountNonce := uint64(1)
			txHash := types.HexToHash("ffcd919661cc14ab71d9d80c3c3481b6b8015c554fe1212a55bbfd138b70fc61")

			conn.On("PendingNonceAt", context.Background(), accountAddress).Return(accountNonce, nil)
			conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)
			conn.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)

			//WaitTillTxMined
			conn.On("BackendType").Return(Real)
			conn.On("TransactionByHash", context.Background(), txHash.Hash).Return(nil, false, nil).Once()

			_, _, gotHandler, err := DeployContract(tt.args.contract, conn, tt.args.params, tt.args.idWithCredentials)

			if (err != nil) != tt.wantErr {
				t.Fatalf("DeployContract() error = %v, wantErr %v", err.Error(), tt.wantErr)
			}

			if reflect.TypeOf(gotHandler) != tt.wantType {
				t.Errorf("DeployContract() %s gotHandlerType = %v, want %v", tt.name, reflect.TypeOf(gotHandler), tt.wantType)
			}

			if tt.assertDeploy {
				if !conn.AssertNumberOfCalls(t, "SendTransaction", 1) {
					t.Errorf("DeployContract() - SendTransaction() was not called")
				}
			}

		})
	}
}
func Test_WaitTillTxMined_mock(t *testing.T) {

	t.Run("Real_Success", func(t *testing.T) {

		conn := &MockContractBackend{}
		conn.On("BackendType").Return(Real)
		conn.On("TransactionByHash", context.Background(), types.Hash{}.Hash).Return(nil, true, nil).Once()
		conn.On("TransactionByHash", context.Background(), types.Hash{}.Hash).Return(nil, false, nil).Once()

		_, _ = WaitTillTxMined(conn, types.Hash{})

		if !conn.AssertCalled(t, "BackendType") {
			t.Errorf("WaitTillTxMined() did not call BackendType")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 2) {
			t.Errorf("WaitTillTxMined() did not call TransactionByHash twice")
		}
	})

	t.Run("Real_Error", func(t *testing.T) {

		conn := &MockContractBackend{}
		conn.On("BackendType").Return(Real)
		conn.On("TransactionByHash", context.Background(), types.Hash{}.Hash).Return(nil, true, fmt.Errorf("")).Once()

		_, _ = WaitTillTxMined(conn, types.Hash{})

		if !conn.AssertCalled(t, "BackendType") {
			t.Errorf("WaitTillTxMined() did not call BackendType")
		}
		if !conn.AssertNumberOfCalls(t, "TransactionByHash", 1) {
			t.Errorf("WaitTillTxMined() did not call TransactionByHash once")
		}
	})

	t.Run("Simulated", func(t *testing.T) {

		conn := &MockContractBackend{}
		conn.On("BackendType").Return(Simulated)

		_, _ = WaitTillTxMined(conn, types.Hash{})

		if !conn.AssertCalled(t, "BackendType") {
			t.Errorf("WaitTillTxMined() did not call BackendType")
		}
	})

	t.Run("UnknownBackend", func(t *testing.T) {

		conn := &MockContractBackend{}
		conn.On("BackendType").Return(BackendType("unknown-backend"))

		_, _ = WaitTillTxMined(conn, types.Hash{})

		if !conn.AssertCalled(t, "BackendType") {
			t.Errorf("WaitTillTxMined() did not call BackendType")
		}
	})
}

func Test_MakeTransactOpts_mock(t *testing.T) {

	t.Run("Valid", func(t *testing.T) {
		idWithCredentials := identity.OffChainID{
			OnChainID:        aliceID.OnChainID,
			ListenerIPAddr:   aliceID.ListenerIPAddr,
			ListenerEndpoint: aliceID.ListenerEndpoint,
			KeyStore:         testKeyStore,
			Password:         alicePassword,
		}
		valueInWei := big.NewInt(0)
		gasLimit := uint64(40e5)

		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)

		_, err := MakeTransactOpts(conn, idWithCredentials, valueInWei, gasLimit)

		if err != nil {
			t.Fatalf("MakeTransactionOpts() error=%s, want nil", err.Error())
		}
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("MakeTransactionOpts() did not call PendingNonceAt")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("MakeTransactionOpts() did not call SuggestGasPrice")
		}

	})

	t.Run("Credentials_not_set", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID:        aliceID.OnChainID,
			ListenerIPAddr:   aliceID.ListenerIPAddr,
			ListenerEndpoint: aliceID.ListenerEndpoint,
		}
		valueInWei := big.NewInt(0)
		gasLimit := uint64(40e5)

		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)

		_, err := MakeTransactOpts(conn, idWithCredentials, valueInWei, gasLimit)

		if err == nil {
			t.Errorf("MakeTransactionOpts() error=nil, want %s", err.Error())
		}
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 0) {
			t.Errorf("MakeTransactionOpts() called PendingNonceAt when credentials not set")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 0) {
			t.Errorf("MakeTransactionOpts() called SuggestGasPrice when credentials not set")
		}
	})
	t.Run("Key_not_in_keystore", func(t *testing.T) {

		idWithCredentials := identity.OffChainID{
			OnChainID:        aliceID.OnChainID,
			ListenerIPAddr:   aliceID.ListenerIPAddr,
			ListenerEndpoint: aliceID.ListenerEndpoint,
			KeyStore:         dummyKeyStore,
			Password:         alicePassword,
		}
		valueInWei := big.NewInt(0)
		gasLimit := uint64(40e5)

		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)

		_, err := MakeTransactOpts(conn, idWithCredentials, valueInWei, gasLimit)

		if err == nil {
			t.Errorf("MakeTransactionOpts() error=nil, want %s", err.Error())
		}
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 0) {
			t.Errorf("MakeTransactionOpts() called PendingNonceAt when credentials not set")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 0) {
			t.Errorf("MakeTransactionOpts() called SuggestGasPrice when credentials not set")
		}
	})

	t.Run("Pending_Nonce_Error", func(t *testing.T) {
		idWithCredentials := identity.OffChainID{
			OnChainID:        aliceID.OnChainID,
			ListenerIPAddr:   aliceID.ListenerIPAddr,
			ListenerEndpoint: aliceID.ListenerEndpoint,
			KeyStore:         testKeyStore,
			Password:         alicePassword,
		}
		valueInWei := big.NewInt(0)
		gasLimit := uint64(40e5)

		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), fmt.Errorf(""))
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), nil)

		_, err := MakeTransactOpts(conn, idWithCredentials, valueInWei, gasLimit)

		if err == nil {
			t.Errorf("MakeTransactionOpts() error=nil, want %s", err.Error())
		}
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("MakeTransactionOpts() did not call PendingNonceAt")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 0) {
			t.Errorf("MakeTransactionOpts() did not call SuggestGasPrice")
		}

	})

	t.Run("SuggestGasPrice_Error", func(t *testing.T) {
		idWithCredentials := identity.OffChainID{
			OnChainID:        aliceID.OnChainID,
			ListenerIPAddr:   aliceID.ListenerIPAddr,
			ListenerEndpoint: aliceID.ListenerEndpoint,
			KeyStore:         testKeyStore,
			Password:         alicePassword,
		}
		valueInWei := big.NewInt(0)
		gasLimit := uint64(40e5)

		conn := &MockContractBackend{}
		conn.On("PendingNonceAt", context.Background(), idWithCredentials.OnChainID.Address).Return(uint64(0), nil)
		conn.On("SuggestGasPrice", context.Background()).Return(big.NewInt(0), fmt.Errorf(""))

		_, err := MakeTransactOpts(conn, idWithCredentials, valueInWei, gasLimit)

		if err == nil {
			t.Errorf("MakeTransactionOpts() error=nil, want %s", err.Error())
		}
		if !conn.AssertNumberOfCalls(t, "PendingNonceAt", 1) {
			t.Errorf("MakeTransactionOpts() did not call PendingNonceAt")
		}
		if !conn.AssertNumberOfCalls(t, "SuggestGasPrice", 1) {
			t.Errorf("MakeTransactionOpts() did not call SuggestGasPrice")
		}

	})
}

func Test_VerifyCodeAt_mock(t *testing.T) {

	t.Run("Regular_Contract_valid", func(t *testing.T) {

		RegularContractRuntimeBin := types.Hex2Bytes("6080604052600436106100405763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a86b5508114610045575b600080fd5b34801561005157600080fd5b50604080516020600460443581810135601f81018490048402850184019095528484526100bb94823573ffffffffffffffffffffffffffffffffffffffff169460248035953695946064949201919081908401838280828437509497506100cf9650505050505050565b604080519115158252519081900360200190f35b600080600080845160411415156100e957600093506101c9565b50505060208201516040830151606084015160001a601b60ff8216101561010e57601b015b8060ff16601b1415801561012657508060ff16601c14155b1561013457600093506101c9565b60408051600080825260208083018085528a905260ff8516838501526060830187905260808301869052925173ffffffffffffffffffffffffffffffffffffffff8b169360019360a0808201949293601f198101939281900390910191865af11580156101a5573d6000803e3d6000fd5b5050506020604051035173ffffffffffffffffffffffffffffffffffffffff161493505b50505093925050505600a165627a7a72305820e6be47295531dbab9d03e72c0105a751b261bd915c5eee2abb07c427707e57b30029")
		RegularContractRuntimeBinHash := strings.ToLower("3c0f29dfe76fd55ab0b023b26c97d2a306805a03788277cd0c7d2817cb7a9bf9")
		contractAddr := types.HexToAddress("de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
		wantMatchStatus := contract.Match

		conn := &MockContractBackend{}
		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(RegularContractRuntimeBin, nil)

		matchStatus, err := VerifyCodeAt(contractAddr, RegularContractRuntimeBinHash, false, conn)

		if err != nil {
			t.Fatalf("VerifyCodeAt() error=%s, want nil", err.Error())
		}
		if matchStatus != wantMatchStatus {
			t.Errorf("VerifyCodeAt() matchstatus=%v, want %v", matchStatus, wantMatchStatus)
		}
		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("VerifyCodeAt() did not call conn.CodeAt with expected params")
		}
	})

	t.Run("Library_Contract_valid", func(t *testing.T) {

		RegularContractRuntimeBin := types.Hex2Bytes("6080604052600436106100405763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a86b5508114610045575b600080fd5b34801561005157600080fd5b50604080516020600460443581810135601f81018490048402850184019095528484526100bb94823573ffffffffffffffffffffffffffffffffffffffff169460248035953695946064949201919081908401838280828437509497506100cf9650505050505050565b604080519115158252519081900360200190f35b600080600080845160411415156100e957600093506101c9565b50505060208201516040830151606084015160001a601b60ff8216101561010e57601b015b8060ff16601b1415801561012657508060ff16601c14155b1561013457600093506101c9565b60408051600080825260208083018085528a905260ff8516838501526060830187905260808301869052925173ffffffffffffffffffffffffffffffffffffffff8b169360019360a0808201949293601f198101939281900390910191865af11580156101a5573d6000803e3d6000fd5b5050506020604051035173ffffffffffffffffffffffffffffffffffffffff161493505b50505093925050505600a165627a7a72305820e6be47295531dbab9d03e72c0105a751b261bd915c5eee2abb07c427707e57b30029")
		RegularContractRuntimeBinHash := strings.ToLower("f5051d349c36270f476012d30d7b10d7ebf144fa30964178700f15b467488273")
		contractAddr := types.HexToAddress("de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
		wantMatchStatus := contract.Match

		conn := &MockContractBackend{}
		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(RegularContractRuntimeBin, nil)

		matchStatus, err := VerifyCodeAt(contractAddr, RegularContractRuntimeBinHash, true, conn)

		if err != nil {
			t.Fatalf("VerifyCodeAt() error=%s, want nil", err.Error())
		}
		if matchStatus != wantMatchStatus {
			t.Errorf("VerifyCodeAt() matchstatus=%v, want %v", matchStatus, wantMatchStatus)
		}
		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("VerifyCodeAt() did not call conn.CodeAt with expected params")
		}
	})

	t.Run("EmptyCode", func(t *testing.T) {

		RegularContractRuntimeBin := []byte{}
		RegularContractRuntimeBinHash := strings.ToLower("3c0f29dfe76fd55ab0b023b26c97d2a306805a03788277cd0c7d2817cb7a9bf9")
		contractAddr := types.HexToAddress("de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
		wantMatchStatus := contract.Unknown

		conn := &MockContractBackend{}
		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(RegularContractRuntimeBin, nil)

		matchStatus, err := VerifyCodeAt(contractAddr, RegularContractRuntimeBinHash, false, conn)

		if err == nil {
			t.Errorf("VerifyCodeAt() error=nil, want %s", err.Error())
		}
		if matchStatus != wantMatchStatus {
			t.Errorf("VerifyCodeAt() matchstatus=%v, want %v", matchStatus, wantMatchStatus)
		}
		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("VerifyCodeAt() did not call conn.CodeAt with expected params")
		}
	})

	t.Run("Checksum_mismatch", func(t *testing.T) {

		RegularContractRuntimeBin := types.Hex2Bytes("6080604052600436106100405763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a86b5508114610045575b600080fd5b34801561005157600080fd5b50604080516020600460443581810135601f81018490048402850184019095528484526100bb94823573ffffffffffffffffffffffffffffffffffffffff169460248035953695946064949201919081908401838280828437509497506100cf9650505050505050565b604080519115158252519081900360200190f35b600080600080845160411415156100e957600093506101c9565b50505060208201516040830151606084015160001a601b60ff8216101561010e57601b015b8060ff16601b1415801561012657508060ff16601c14155b1561013457600093506101c9565b60408051600080825260208083018085528a905260ff8516838501526060830187905260808301869052925173ffffffffffffffffffffffffffffffffffffffff8b169360019360a0808201949293601f198101939281900390910191865af11580156101a5573d6000803e3d6000fd5b5050506020604051035173ffffffffffffffffffffffffffffffffffffffff161493505b50505093925050505600a165627a7a72305820e6be47295531dbab9d03e72c0105a751b261bd915c5eee2abb07c427707e57b30029")
		RegularContractRuntimeBinHash := strings.ToLower("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
		contractAddr := types.HexToAddress("de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
		wantMatchStatus := contract.NoMatch

		conn := &MockContractBackend{}
		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(RegularContractRuntimeBin, nil)

		matchStatus, err := VerifyCodeAt(contractAddr, RegularContractRuntimeBinHash, false, conn)

		if err != nil {
			t.Fatalf("VerifyCodeAt() error=%s, want nil", err.Error())
		}
		if matchStatus != wantMatchStatus {
			t.Errorf("VerifyCodeAt() matchstatus=%v, want %v", matchStatus, wantMatchStatus)
		}
		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("VerifyCodeAt() did not call conn.CodeAt with expected params")
		}
	})

	t.Run("CodeAt_Error", func(t *testing.T) {

		RegularContractRuntimeBin := []byte{}
		RegularContractRuntimeBinHash := strings.ToLower("3c0f29dfe76fd55ab0b023b26c97d2a306805a03788277cd0c7d2817cb7a9bf9")
		contractAddr := types.HexToAddress("0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
		wantMatchStatus := contract.Unknown

		conn := &MockContractBackend{}
		conn.On("CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)).Return(RegularContractRuntimeBin, fmt.Errorf(""))

		matchStatus, err := VerifyCodeAt(contractAddr, RegularContractRuntimeBinHash, false, conn)

		if err == nil {
			t.Errorf("VerifyCodeAt() error=nil, want %s", err.Error())
		}
		if matchStatus != wantMatchStatus {
			t.Errorf("VerifyCodeAt() matchstatus=%v, want %v", matchStatus, wantMatchStatus)
		}
		if !conn.AssertCalled(t, "CodeAt", context.Background(), contractAddr.Address, (*big.Int)(nil)) {
			t.Errorf("VerifyCodeAt() did not call conn.CodeAt with expected params")
		}
	})
}
