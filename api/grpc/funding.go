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
	"fmt"

	pchannel "perun.network/go-perun/channel"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

// fundingServer represents a grpc server that can serve funding API.
type fundingServer struct {
	pb.UnimplementedFunding_APIServer
	n perun.NodeAPI
}

// Fund wraps session.Fund.
func (a *fundingServer) Fund(ctx context.Context, grpcReq *pb.FundReq) (*pb.FundResp, error) {
	errResponse := func(err perun.APIError) *pb.FundResp {
		return &pb.FundResp{
			Error: pb.FromError(err),
		}
	}

	sess, apiErr := a.n.GetSession(grpcReq.SessionID)
	if apiErr != nil {
		return errResponse(apiErr), nil
	}
	req, err := pb.ToFundingReq(grpcReq)
	if err != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err)), nil
	}

	err = sess.Fund(ctx, req)
	if err != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err)), nil
	}

	return &pb.FundResp{
		Error: nil,
	}, nil
}

// RegisterAssetERC20 is a stub that always returns false. Because, the remote
// funder does not support use of assets other than the default ERC20 asset.
//
// TODO: Make actual implementation.
func (a *payChAPIServer) RegisterAssetERC20(_ context.Context, _ *pb.RegisterAssetERC20Req) (
	*pb.RegisterAssetERC20Resp, error,
) {
	return &pb.RegisterAssetERC20Resp{
		MsgSuccess: false,
	}, nil
}

// IsAssetRegistered wraps session.IsAssetRegistered.
func (a *payChAPIServer) IsAssetRegistered(_ context.Context, req *pb.IsAssetRegisteredReq) (
	*pb.IsAssetRegisteredResp,
	error,
) {
	errResponse := func(err perun.APIError) *pb.IsAssetRegisteredResp {
		return &pb.IsAssetRegisteredResp{
			Response: &pb.IsAssetRegisteredResp_Error{
				Error: pb.FromError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	asset := pchannel.NewAsset()
	err2 := asset.UnmarshalBinary(req.Asset)
	if err2 != nil {
		err = perun.NewAPIErrInvalidArgument(err2, "asset", fmt.Sprintf("%x", req.Asset))
		return errResponse(err), nil
	}

	isRegistered := sess.IsAssetRegistered(asset)

	return &pb.IsAssetRegisteredResp{
		Response: &pb.IsAssetRegisteredResp_MsgSuccess_{
			MsgSuccess: &pb.IsAssetRegisteredResp_MsgSuccess{
				IsRegistered: isRegistered,
			},
		},
	}, nil
}
