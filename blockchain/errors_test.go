// Copyright (c) 2021 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/hyperledger-labs/perun-node
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

package blockchain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node/blockchain"
)

func Test_NewInvalidContractError(t *testing.T) {
	name := "some-name"
	address := "some-address"
	err := assert.AnError

	gotErr := blockchain.NewInvalidContractError(name, address, err)
	require.Error(t, gotErr)

	invalidContractErr := blockchain.InvalidContractError{}
	require.True(t, errors.As(gotErr, &invalidContractErr))

	assert.Equal(t, name, invalidContractErr.Name)
	assert.Equal(t, address, invalidContractErr.Address)
	assert.True(t, errors.Is(gotErr, err), "should return the underlying error for comparison")
}

func Test_NewAssetERC20RegisteredError(t *testing.T) {
	asset := "some-asset"
	symbol := "some-symbol"

	gotErr := blockchain.NewAssetERC20RegisteredError(asset, symbol)
	require.Error(t, gotErr)

	assetERC20RegisteredError := blockchain.AssetERC20RegisteredError{}
	require.True(t, errors.As(gotErr, &assetERC20RegisteredError))

	assert.Equal(t, asset, assetERC20RegisteredError.Asset)
	assert.Equal(t, symbol, assetERC20RegisteredError.Symbol)
}
