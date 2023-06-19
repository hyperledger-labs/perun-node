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

package pb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

func Test_ErrorCategory(t *testing.T) {
	assert.Equal(t, pb.ErrorCategory_ParticipantError, pb.ErrorCategory(perun.ParticipantError))
	assert.Equal(t, pb.ErrorCategory_ClientError, pb.ErrorCategory(perun.ClientError))
	assert.Equal(t, pb.ErrorCategory_ProtocolError, pb.ErrorCategory(perun.ProtocolFatalError))
	assert.Equal(t, pb.ErrorCategory_InternalError, pb.ErrorCategory(perun.InternalError))
}

func Test_ErrorCode(t *testing.T) {
	assert.Equal(t, pb.ErrorCode_ErrPeerRequestTimedOut, pb.ErrorCode(perun.ErrPeerRequestTimedOut))
	assert.Equal(t, pb.ErrorCode_ErrPeerRejected, pb.ErrorCode(perun.ErrPeerRejected))
	assert.Equal(t, pb.ErrorCode_ErrPeerNotFunded, pb.ErrorCode(perun.ErrPeerNotFunded))
	assert.Equal(t, pb.ErrorCode_ErrUserResponseTimedOut, pb.ErrorCode(perun.ErrUserResponseTimedOut))
	assert.Equal(t, pb.ErrorCode_ErrResourceNotFound, pb.ErrorCode(perun.ErrResourceNotFound))
	assert.Equal(t, pb.ErrorCode_ErrResourceExists, pb.ErrorCode(perun.ErrResourceExists))
	assert.Equal(t, pb.ErrorCode_ErrInvalidArgument, pb.ErrorCode(perun.ErrInvalidArgument))
	assert.Equal(t, pb.ErrorCode_ErrFailedPreCondition, pb.ErrorCode(perun.ErrFailedPreCondition))
	assert.Equal(t, pb.ErrorCode_ErrInvalidConfig, pb.ErrorCode(perun.ErrInvalidConfig))
	assert.Equal(t, pb.ErrorCode_ErrInvalidContracts, pb.ErrorCode(perun.ErrInvalidContracts))
	assert.Equal(t, pb.ErrorCode_ErrTxTimedOut, pb.ErrorCode(perun.ErrTxTimedOut))
	assert.Equal(t, pb.ErrorCode_ErrChainNotReachable, pb.ErrorCode(perun.ErrChainNotReachable))
	assert.Equal(t, pb.ErrorCode_ErrUnknownInternal, pb.ErrorCode(perun.ErrUnknownInternal))
}
