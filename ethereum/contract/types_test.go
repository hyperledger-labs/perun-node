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

package contract

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/direct-state-transfer/dst-go/ethereum/types"
)

func Test_StoreType(t *testing.T) {

	var (
		testLibsignatures = Handler{
			Name:               "LibSignatures",
			HashSolFile:        strings.ToLower("359e2e9f7bacdcefc6962c46182aba7f16b8b0a8314468ca8dd88edd25299209"),
			HashGoFile:         strings.ToLower("1b6a7102bd6726168ea973b0bb2107655b638bc09b702cffe05a9d90e67373d7"),
			HashBinRuntimeFile: strings.ToLower("3c0f29dfe76fd55ab0b023b26c97d2a306805a03788277cd0c7d2817cb7a9bf9"),
			GasUnits:           uint64(40e5),
			Version:            "0.0.1",
		}
		testMscontract = Handler{
			Name:               "MSContract",
			HashSolFile:        strings.ToLower("d2f7c0055a445f4823a2a4312df1bb7602fb62b022d99047634fe0cb0a941938"),
			HashGoFile:         strings.ToLower("8a7f2d750db8bb1bf7c1a8fedae204a828a1e81e7ac7a252737f3528dea29ceb"),
			HashBinRuntimeFile: strings.ToLower("4fb304c42b1bad1b03417c72e925d87488dea1d22e13ed0601abbe8d86c8e8ad"),
			GasUnits:           uint64(40e5),
			Version:            "0.0.1",
		}
		testVpc = Handler{
			Name:               "VPC",
			HashSolFile:        strings.ToLower("c2195856c9d206c18d3ec49eb8b4b38a2b022ec6ef155478fe1bb3e8fd78b494"),
			HashGoFile:         strings.ToLower("1b6a7102bd6726168ea973b0bb2107655b638bc09b702cffe05a9d90e67373d7"),
			HashBinRuntimeFile: strings.ToLower("978bc824e314529d620afb0c1f07770dad5e36b11ad0060102f67fc29e3a60fe"),
			GasUnits:           uint64(40e5),
			Version:            "0.0.1",
		}
		testTimeoutMscontract          = 100 * time.Minute
		testTimeoutVpcValidity         = 10 * time.Minute
		testTimeoutVpcExtendedValidity = 20 * time.Minute
		testSHA256Sum                  = types.Hex2Bytes("8b9df51d7a7343a92133f24aaa1ba544d92eac1ccd66c731aa243e3273cb0f69")
	)

	testStore := StoreType{
		libSignatures:              testLibsignatures,
		msContract:                 testMscontract,
		vpc:                        testVpc,
		timeoutMSContract:          testTimeoutMscontract,
		timeoutVPCValidity:         testTimeoutVpcValidity,
		timeoutVPCExtendedValidity: testTimeoutVpcExtendedValidity,
	}

	t.Run("StoreType_Libsignatures", func(t *testing.T) {
		got := testStore.LibSignatures()
		if !reflect.DeepEqual(testLibsignatures, got) {
			t.Errorf("StoreType_Libsignatures() got %v, want %v", got, testLibsignatures)
		}
	})
	t.Run("StoreType_Mscontract", func(t *testing.T) {
		got := testStore.MSContract()
		if !reflect.DeepEqual(testMscontract, got) {
			t.Errorf("StoreType_Mscontract() got %v, want %v", got, testMscontract)
		}
	})
	t.Run("StoreType_Vpc", func(t *testing.T) {
		got := testStore.VPC()
		if !reflect.DeepEqual(testVpc, got) {
			t.Errorf("StoreType_Vpc() got %v, want %v", got, testVpc)
		}
	})
	t.Run("StoreType_TimeoutMscontract", func(t *testing.T) {
		got := testStore.TimeoutMSContract()
		if !reflect.DeepEqual(testTimeoutMscontract, got) {
			t.Errorf("StoreType_TimeoutMscontract() got %v, want %v", got, testTimeoutMscontract)
		}
	})
	t.Run("StoreType_TimeoutVpcValidity", func(t *testing.T) {
		got := testStore.TimeoutVPCValidity()
		if !reflect.DeepEqual(testTimeoutVpcValidity, got) {
			t.Errorf("StoreType_TimeoutVpcValidity() got %v, want %v", got, testTimeoutVpcValidity)
		}
	})
	t.Run("StoreType_TimeoutVpcExtendedValidity", func(t *testing.T) {
		got := testStore.TimeoutVPCExtendedValidity()
		if !reflect.DeepEqual(testTimeoutVpcExtendedValidity, got) {
			t.Errorf("StoreType_TimeoutVpcExtendedValidity() got %v, want %v", got, testTimeoutVpcExtendedValidity)
		}
	})

	t.Run("StoreType_SHA256Sum", func(t *testing.T) {
		got := testStore.SHA256Sum()
		if !reflect.DeepEqual(testSHA256Sum, got) {
			t.Errorf("StoreType_SHA256Sum() got %x, want %v", got, testSHA256Sum)
		}
	})
}
