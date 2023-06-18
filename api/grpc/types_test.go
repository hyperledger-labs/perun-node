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
