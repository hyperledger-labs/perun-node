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

	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	psync "polycry.pt/poly-go/sync"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

// fundingServer represents a grpc server that can serve funding API.
type fundingServer struct {
	pb.UnimplementedFunding_APIServer
	n perun.NodeAPI

	// The mutex should be used when accessing the map data structures.
	psync.Mutex
	subscribes map[string]map[pchannel.ID]pchannel.AdjudicatorSubscription
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

// Register wraps session.Register.
func (a *fundingServer) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	errResponse := func(err perun.APIError) *pb.RegisterResp {
		return &pb.RegisterResp{
			Error: pb.FromError(err),
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	adjReq, err2 := pb.ToAdjReq(req.AdjReq)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}
	signedStates := make([]pchannel.SignedState, len(req.SignedStates))
	for i := range signedStates {
		signedStates[i], err2 = pb.ToSignedState(req.SignedStates[i])
		if err2 != nil {
			return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
		}
	}

	err2 = sess.Register(ctx, adjReq, signedStates)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}

	return &pb.RegisterResp{
		Error: nil,
	}, nil
}

// Withdraw wraps session.Withdraw.
func (a *fundingServer) Withdraw(ctx context.Context, req *pb.WithdrawReq) (*pb.WithdrawResp, error) {
	errResponse := func(err perun.APIError) *pb.WithdrawResp {
		return &pb.WithdrawResp{
			Error: pb.FromError(err),
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	adjReq, err2 := pb.ToAdjReq(req.AdjReq)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}
	stateMap := pchannel.StateMap(make(map[pchannel.ID]*pchannel.State))

	for i := range req.StateMap {
		var id pchannel.ID
		copy(id[:], req.StateMap[i].Id)
		stateMap[id], err2 = pb.ToState(req.StateMap[i].State)
		if err2 != nil {
			return errResponse(err), nil
		}
	}

	err2 = sess.Withdraw(ctx, adjReq, stateMap)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}

	return &pb.WithdrawResp{
		Error: nil,
	}, nil
}

// Progress wraps session.Progress.
func (a *fundingServer) Progress(ctx context.Context, req *pb.ProgressReq) (*pb.ProgressResp, error) {
	errResponse := func(err perun.APIError) *pb.ProgressResp {
		return &pb.ProgressResp{
			Error: pb.FromError(err),
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	var progReq perun.ProgressReq
	var err2 error
	progReq.AdjudicatorReq, err2 = pb.ToAdjReq(req.AdjReq)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}
	progReq.NewState, err2 = pb.ToState(req.NewState)
	if err2 != nil {
		return errResponse(err), nil
	}
	copy(progReq.Sig, req.Sig)

	err2 = sess.Progress(ctx, progReq)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}

	return &pb.ProgressResp{
		Error: nil,
	}, nil
}

// Subscribe wraps session.Subscribe.

func (a *fundingServer) Subscribe(req *pb.SubscribeReq, stream pb.Funding_API_SubscribeServer) error {
	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errors.WithMessage(err, "retrieving session")
	}

	var chID pchannel.ID
	copy(chID[:], req.ChID)

	adjSub, err := sess.Subscribe(context.Background(), chID)
	if err != nil {
		return errors.WithMessage(err, "setting up subscription")
	}

	a.Lock()
	a.subscribes[req.SessionID][chID] = adjSub
	a.Unlock()

	// This stream is anyways closed when StopWatching is called for.
	// Hence, that will act as the exit condition for the loop.
	go func() {
		// will return nil, when the sub is closed.
		// so, we need a mechanism to call close on the server side.
		// so, add a call Unsubscribe, which simply calls close.
		for {
			adjEvent := adjSub.Next()
			if adjEvent == nil {
				err := errors.WithMessage(adjSub.Err(), "sub closed with error")
				notif := &pb.SubscribeResp_Error{
					Error: pb.FromError(perun.NewAPIErrUnknownInternal(err)),
				}
				// TODO: Proper error handling. For now, ignore this error.
				_ = stream.Send(&pb.SubscribeResp{Response: notif}) //nolint: errcheck
				return
			}
			notif, err := pb.SubscribeResponseFromAdjEvent(adjEvent)
			if err != nil {
				return
			}
			err = stream.Send(notif)
			if err != nil {
				return
			}
		}
	}()

	return nil
}

func (a *fundingServer) Unsubscribe(_ context.Context, req *pb.UnsubscribeReq) (*pb.UnsubscribeResp, error) {
	errResponse := func(err perun.APIError) *pb.UnsubscribeResp {
		return &pb.UnsubscribeResp{
			Error: pb.FromError(err),
		}
	}

	var chID pchannel.ID
	copy(chID[:], req.ChID)

	a.Lock()
	adjSub := a.subscribes[req.SessionID][chID]
	delete(a.subscribes[req.SessionID], chID)
	a.Unlock()

	if err := adjSub.Close(); err != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(errors.WithMessage(err, "retrieving session"))), nil
	}

	return &pb.UnsubscribeResp{
		Error: nil,
	}, nil
}
