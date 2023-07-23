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

package session

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"

	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
)

// Fund provides a wrapper to call the fund method on session funder.
// On-chain wallet can be direcly used without additional locks,
// because the methods on wallet are concurrency safe by themselves.
func (s *Session) Fund(ctx context.Context, req pchannel.FundingReq) error {
	err := s.funder.Fund(ctx, req)
	if err == nil {
		s.user.OnChain.Wallet.IncrementUsage(s.user.OnChain.Addr)
	}
	return err
}

// RegisterAssetERC20 is a stub that always returns false. Because, the remote
// funder does not support use of assets other than the default ERC20 asset.
//
// TODO: Make actual implementation.
func (s *Session) RegisterAssetERC20(asset pchannel.Asset, token, acc pwallet.Address) bool {
	s.WithField("method", "RegisterAssetERC20").
		Infof("\nReceived request with params %+v, %+v, %+v", asset, token, acc)
	s.WithField("method", "RegisterAssetERC20").Infof("Unimplemented method. Hence returning false")
	return false
}

// IsAssetRegistered provides a wrapper to call the IsAssetRegistered method
// on session funder.
func (s *Session) IsAssetRegistered(asset pchannel.Asset) bool {
	s.WithField("method", "IsAssetRegistered").Infof("\nReceived request with params %+v", asset)
	isAssetRegistered := s.funder.IsAssetRegistered(asset)
	s.WithField("method", "IsAssetRegistered").Infof("Response: %v", isAssetRegistered)
	return isAssetRegistered
}

type grpcFunder struct {
	apiKey string
	client pb.Funding_APIClient
}

func (f *grpcFunder) Fund(_ context.Context, fundingReq pchannel.FundingReq) error {
	protoReq, err := pb.FromFundingReq(fundingReq)
	if err != nil {
		err = errors.WithMessage(err, "constructing grpc funding request")
		return perun.NewAPIErrUnknownInternal(err)
	}
	protoReq.SessionID = f.apiKey
	resp, err := f.client.Fund(context.Background(), protoReq)
	if err != nil {
		err = errors.WithMessage(err, "sending the funding request")
		return perun.NewAPIErrUnknownInternal(err)
	}
	if resp.Error != nil && resp.Error.Message != "" {
		// TODO: Proper error handling
		err = errors.Errorf("funding the channel: %s", resp.Error.Message)
		return perun.NewAPIErrUnknownInternal(err)
	}
	return nil
}

func (f *grpcFunder) RegisterAssetERC20(asset pchannel.Asset, token, acc pwallet.Address) bool {
	protoAsset, err := asset.MarshalBinary()
	if err != nil {
		return false
	}
	protoToken, err := token.MarshalBinary()
	if err != nil {
		return false
	}
	protoAcc, err := acc.MarshalBinary()
	if err != nil {
		return false
	}
	registerAssetERC20Req := &pb.RegisterAssetERC20Req{
		SessionID:   f.apiKey,
		Asset:       protoAsset,
		TokenAddr:   fmt.Sprintf("%x", protoToken),
		DeposiorAcc: fmt.Sprintf("%x", protoAcc),
	}

	resp, err := f.client.RegisterAssetERC20(context.Background(), registerAssetERC20Req)
	if err != nil {
		return false
	}

	return resp.MsgSuccess
}

func (f *grpcFunder) IsAssetRegistered(asset pchannel.Asset) bool {
	protoAsset, err := asset.MarshalBinary()
	if err != nil {
		return false
	}
	isAssetRegisteredReq := &pb.IsAssetRegisteredReq{
		SessionID: f.apiKey,
		Asset:     protoAsset,
	}

	resp, err := f.client.IsAssetRegistered(context.Background(), isAssetRegisteredReq)
	if err != nil {
		return false
	}

	_, ok := resp.Response.(*pb.IsAssetRegisteredResp_Error)
	if ok {
		// TODO: Proper error handling and return it.
		return false
	}
	return resp.Response.(*pb.IsAssetRegisteredResp_MsgSuccess_).MsgSuccess.IsRegistered
}
