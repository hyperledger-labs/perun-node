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

package grpc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

func Test_ChUpdateType(t *testing.T) {
	assert.EqualValues(t, perun.ChUpdateTypeOpen, pb.SubPayChUpdatesResp_Notify_open)
	assert.EqualValues(t, perun.ChUpdateTypeFinal, pb.SubPayChUpdatesResp_Notify_final)
	assert.EqualValues(t, perun.ChUpdateTypeClosed, pb.SubPayChUpdatesResp_Notify_closed)
}

func Test_ErrorCategory(t *testing.T) {
	assert.Equal(t, pb.ErrorCategory_ParticipantError, pb.ErrorCategory(perun.ParticipantError))
	assert.Equal(t, pb.ErrorCategory_ClientError, pb.ErrorCategory(perun.ClientError))
	assert.Equal(t, pb.ErrorCategory_ProtocolError, pb.ErrorCategory(perun.ProtocolFatalError))
	assert.Equal(t, pb.ErrorCategory_InternalError, pb.ErrorCategory(perun.InternalError))
}

func Test_ErrorCode(t *testing.T) {
	assert.Equal(t, pb.ErrorCode_ErrV2PeerResponseTimedOut, pb.ErrorCode(perun.ErrV2PeerResponseTimedOut))
	assert.Equal(t, pb.ErrorCode_ErrV2RejectedByPeer, pb.ErrorCode(perun.ErrV2RejectedByPeer))
	assert.Equal(t, pb.ErrorCode_ErrV2PeerNotFunded, pb.ErrorCode(perun.ErrV2PeerNotFunded))
	assert.Equal(t, pb.ErrorCode_ErrV2UserResponseTimedOut, pb.ErrorCode(perun.ErrV2UserResponseTimedOut))
	assert.Equal(t, pb.ErrorCode_ErrV2ResourceNotFound, pb.ErrorCode(perun.ErrV2ResourceNotFound))
	assert.Equal(t, pb.ErrorCode_ErrV2ResourceExists, pb.ErrorCode(perun.ErrV2ResourceExists))
	assert.Equal(t, pb.ErrorCode_ErrV2InvalidArgument, pb.ErrorCode(perun.ErrV2InvalidArgument))
	assert.Equal(t, pb.ErrorCode_ErrV2FailedPreCondition, pb.ErrorCode(perun.ErrV2FailedPreCondition))
	assert.Equal(t, pb.ErrorCode_ErrV2InvalidConfig, pb.ErrorCode(perun.ErrV2InvalidConfig))
	assert.Equal(t, pb.ErrorCode_ErrV2ChainNodeNotReachable, pb.ErrorCode(perun.ErrV2ChainNodeNotReachable))
	assert.Equal(t, pb.ErrorCode_ErrV2InvalidContracts, pb.ErrorCode(perun.ErrV2InvalidContracts))
	assert.Equal(t, pb.ErrorCode_ErrV2TxTimedOut, pb.ErrorCode(perun.ErrV2TxTimedOut))
	assert.Equal(t, pb.ErrorCode_ErrV2InsufficientBalForTx, pb.ErrorCode(perun.ErrV2InsufficientBalForTx))
	assert.Equal(t, pb.ErrorCode_ErrV2ChainNodeDisconnected, pb.ErrorCode(perun.ErrV2ChainNodeDisconnected))
	assert.Equal(t, pb.ErrorCode_ErrV2InsufficientBalForDeposit, pb.ErrorCode(perun.ErrV2InsufficientBalForDeposit))
	assert.Equal(t, pb.ErrorCode_ErrV2UnknownInternal, pb.ErrorCode(perun.ErrV2UnknownInternal))
	assert.Equal(t, pb.ErrorCode_ErrV2OffChainComm, pb.ErrorCode(perun.ErrV2OffChainComm))
}
