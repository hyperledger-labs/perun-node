// Copyright (c) 2023 - for information on the respective copyright owner
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

package grpc

import (
	"context"

	"github.com/pkg/errors"

	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/api/handlers"
)

// watchingServer represents a grpc server that can serve watching API.
type watchingServer struct {
	pb.UnimplementedWatching_APIServer
	*handlers.WatchingHandler
}

// StartWatchingLedgerChannel wraps session.StartWatchingLedgerChannel.
func (a *watchingServer) StartWatchingLedgerChannel(
	srv pb.Watching_API_StartWatchingLedgerChannelServer,
) error {
	req, err := srv.Recv()
	if err != nil {
		return errors.WithMessage(err, "reading request data")
	}

	sendAdjEvent := func(resp *pb.StartWatchingLedgerChannelResp) error {
		return srv.Send(resp)
	}

	receiveState := func() (req *pb.StartWatchingLedgerChannelReq, err error) {
		return srv.Recv()
	}

	return a.WatchingHandler.StartWatchingLedgerChannel(req, sendAdjEvent, receiveState)
}

// StopWatching wraps session.StopWatching.
func (a *watchingServer) StopWatching(ctx context.Context, req *pb.StopWatchingReq) (*pb.StopWatchingResp, error) {
	return a.WatchingHandler.StopWatching(ctx, req)
}
