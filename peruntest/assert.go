// Copyright (c) 2020 - for information on the respective copyright owner
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

package peruntest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
)

// AssertAPIError tests if the passed error contains expected category, code
// and phrases in the message.
func AssertAPIError(t *testing.T, e perun.APIError, categ perun.ErrorCategory, code perun.ErrorCode, msgs ...string) {
	t.Helper()

	require.Error(t, e)
	assert.Equal(t, categ, e.Category())
	assert.Equal(t, code, e.Code())
	for _, msg := range msgs {
		assert.Contains(t, e.Message(), msg)
	}
}

// AssertErrInfoPeerRequestTimedOut tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoPeerRequestTimedOut(t *testing.T, info interface{}, peerAlias, timeout string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrInfoPeerRequestTimedOut)
	require.True(t, ok)
	assert.Equal(t, peerAlias, addInfo.PeerAlias)
	assert.Equal(t, timeout, addInfo.Timeout)
}

// AssertErrInfoPeerRejected tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoPeerRejected(t *testing.T, info interface{}, peerAlias, reason string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrInfoPeerRejected)
	require.True(t, ok)
	assert.Equal(t, peerAlias, addInfo.PeerAlias)
	assert.Equal(t, reason, addInfo.Reason)
}

// AssertErrInfoPeerNotFunded tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoPeerNotFunded(t *testing.T, info interface{}, peerAlias string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrInfoPeerNotFunded)
	require.True(t, ok)
	assert.Equal(t, peerAlias, addInfo.PeerAlias)
}

// AssertErrInfoUserResponseTimedOut tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoUserResponseTimedOut(t *testing.T, info interface{}) {
	t.Helper()

	addInfo, ok := info.(perun.ErrInfoUserResponseTimedOut)
	require.True(t, ok)
	assert.Less(t, addInfo.Expiry, addInfo.ReceivedAt)
}

// AssertErrInfoResourceNotFound tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoResourceNotFound(t *testing.T, info interface{}, resourceType perun.ResourceType, id string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrInfoResourceNotFound)
	require.True(t, ok)
	assert.Equal(t, string(resourceType), addInfo.Type)
	assert.Equal(t, id, addInfo.ID)
}

// AssertErrInfoResourceExists tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoResourceExists(t *testing.T, info interface{}, resourceType perun.ResourceType, id string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrInfoResourceExists)
	require.True(t, ok)
	assert.Equal(t, string(resourceType), addInfo.Type)
	assert.Equal(t, id, addInfo.ID)
}

// AssertErrInfoInvalidArgument tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoInvalidArgument(t *testing.T, info interface{}, name perun.ArgumentName, value string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrInfoInvalidArgument)
	require.True(t, ok)
	assert.Equal(t, string(name), addInfo.Name)
	assert.Equal(t, value, addInfo.Value)
	t.Log("requirement:", addInfo.Requirement)
}

// AssertErrInfoFailedPreCondUnclosedChs tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoFailedPreCondUnclosedChs(t *testing.T, info interface{}, chInfos []perun.ChInfo) {
	t.Helper()

	addInfo, ok := info.(perun.ErrInfoFailedPreCondUnclosedChs)
	require.True(t, ok)
	assert.Equal(t, chInfos, addInfo.ChInfos)
}

// AssertErrInfoInvalidConfig tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoInvalidConfig(t *testing.T, info interface{}, name, value string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrInfoInvalidConfig)
	require.True(t, ok)
	assert.Equal(t, name, addInfo.Name)
	assert.Equal(t, value, addInfo.Value)
}

// AssertErrInfoInvalidContracts tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoInvalidContracts(t *testing.T, info interface{}, contractErrInfos []perun.ContractErrInfo) {
	t.Helper()

	_, ok := info.(perun.ErrInfoInvalidContracts)
	require.True(t, ok)
	// TODO: compare the two list based on matching keys.
	// assert.Len(t, len(contractErrInfos), len(addInfo.ContractErrInfos))
}

// AssertErrInfoTxTimedOut tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoTxTimedOut(t *testing.T, info interface{}, txType, txID, txTimeout string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrInfoTxTimedOut)
	require.True(t, ok)
	assert.Equal(t, txType, addInfo.TxType)
	assert.Equal(t, txID, addInfo.TxID)
	assert.Equal(t, txTimeout, addInfo.TxTimeout)
}

// AssertErrInfoChainNotReachable tests if additional info field is of
// correct type and has expected values.
func AssertErrInfoChainNotReachable(t *testing.T, info interface{}, chainURL string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrInfoChainNotReachable)
	require.True(t, ok)
	assert.Equal(t, chainURL, addInfo.ChainURL)
}
