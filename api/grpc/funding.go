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

	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/api/handlers"
)

// fundingServer represents a grpc server that can serve funding API.
type fundingServer struct {
	pb.UnimplementedFunding_APIServer
	*handlers.FundingHandler
}

// Fund wraps session.Fund.
func (a *fundingServer) Fund(ctx context.Context, req *pb.FundReq) (*pb.FundResp, error) {
	return a.FundingHandler.Fund(ctx, req)
}

// RegisterAssetERC20 is a stub that always returns false. Because, the remote
// funder does not support use of assets other than the default ERC20 asset.
//
// TODO: Make actual implementation.
func (a *fundingServer) RegisterAssetERC20(ctx context.Context, req *pb.RegisterAssetERC20Req) (
	*pb.RegisterAssetERC20Resp, error,
) {
	return a.FundingHandler.RegisterAssetERC20(ctx, req)
}

// IsAssetRegistered wraps session.IsAssetRegistered.
func (a *fundingServer) IsAssetRegistered(ctx context.Context, req *pb.IsAssetRegisteredReq) (
	*pb.IsAssetRegisteredResp,
	error,
) {
	return a.FundingHandler.IsAssetRegistered(ctx, req)
}

// Register wraps session.Register.
func (a *fundingServer) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	return a.FundingHandler.Register(ctx, req)
}

// Withdraw wraps session.Withdraw.
func (a *fundingServer) Withdraw(ctx context.Context, req *pb.WithdrawReq) (*pb.WithdrawResp, error) {
	return a.FundingHandler.Withdraw(ctx, req)
}

// Progress wraps session.Progress.
func (a *fundingServer) Progress(ctx context.Context, req *pb.ProgressReq) (*pb.ProgressResp, error) {
	return a.FundingHandler.Progress(ctx, req)
}

// Subscribe wraps session.Subscribe.

func (a *fundingServer) Subscribe(req *pb.SubscribeReq, stream pb.Funding_API_SubscribeServer) error {
	notify := func(notif *pb.SubscribeResp) error {
		return stream.Send(notif)
	}

	return a.FundingHandler.Subscribe(req, notify)
}

func (a *fundingServer) Unsubscribe(ctx context.Context, req *pb.UnsubscribeReq) (*pb.UnsubscribeResp, error) {
	return a.FundingHandler.Unsubscribe(ctx, req)
}
