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
