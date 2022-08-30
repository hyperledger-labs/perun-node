// Copyright (c) 2022 - for information on the respective copyright owner
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

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
)

// Fund provides a wrapper to call the fund method on session funder.
// On-chain wallet can be direcly used without additional locks,
// because the methods on wallet are concurrency safe by themselves.
func (s *Session) Fund(ctx context.Context, req pchannel.FundingReq) perun.APIError {
	// s.WithField("method", "Fund").Infof("\nReceived request with params %+v", req)
	s.Infof("\ncar: funding request for channel with charger, balance: %+v", req.Agreement)
	err := s.funder.Fund(ctx, req)
	// TODO: Proper error handling with specific, actionable error codes.
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("Fund", apiErr)).Error(apiErr.Message())
		return apiErr
	}
	s.user.OnChain.Wallet.IncrementUsage(s.user.OnChain.Addr)
	// s.WithField("method", "Fund").Infof("Funded channel successfully: %+v", req.State.ID)
	s.Infof("\nfunded successfully", req.Agreement)
	return nil
}

// RegisterAssetERC20 is a stub that always returns false. Because, the remote
// funder does not support use of assets other than the default ERC20 asset.
//
// TODO: Make actual implementation.
func (s *Session) RegisterAssetERC20(asset pchannel.Asset, token, acc pwallet.Address) bool {
	s.WithField("method", "RegisterAssetERC20").Infof("\nReceived request with params %+v, %+v, %+v",
		asset, token, acc)
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
	client pb.Payment_APIClient
}

func (f *grpcFunder) Fund(ctx context.Context, fundingReq pchannel.FundingReq) error {
	protoReq, err := fromFundReq(fundingReq)
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
		err = errors.WithMessage(err, "funding the channel")
		return perun.NewAPIErrUnknownInternal(errors.New(resp.Error.Message))
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

func fromFundReq(req pchannel.FundingReq) (protoReq *pb.FundReq, err error) {
	protoReq = &pb.FundReq{}

	if protoReq.Params, err = fromParams(req.Params); err != nil {
		return protoReq, err
	}
	if protoReq.State, err = fromState(req.State); err != nil {
		return protoReq, err
	}

	protoReq.Idx = uint32(protoReq.Idx)
	protoReq.Agreement, err = fromBalances(req.Agreement)
	if err != nil {
		return nil, errors.WithMessage(err, "agreement")
	}
	return protoReq, nil
}

func fromParams(params *pchannel.Params) (protoParams *pb.Params, err error) {
	protoParams = &pb.Params{}

	protoParams.Nonce = params.Nonce.Bytes()
	protoParams.ChallengeDuration = params.ChallengeDuration
	protoParams.LedgerChannel = params.LedgerChannel
	protoParams.VirtualChannel = params.VirtualChannel
	protoParams.Parts, err = fromWalletAddrs(params.Parts)
	if err != nil {
		return nil, errors.WithMessage(err, "parts")
	}
	protoParams.App, err = fromApp(params.App)
	return protoParams, err
}

func fromWalletAddrs(addrs []pwallet.Address) (protoAddrs [][]byte, err error) {
	protoAddrs = make([][]byte, len(addrs))
	for i := range addrs {
		protoAddrs[i], err = addrs[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return protoAddrs, nil
}

func fromApp(app pchannel.App) (protoApp []byte, err error) {
	if pchannel.IsNoApp(app) {
		return []byte{}, nil
	}
	protoApp, err = app.Def().MarshalBinary()
	return protoApp, err
}

func fromState(state *pchannel.State) (protoState *pb.State, err error) {
	protoState = &pb.State{}

	protoState.Id = make([]byte, len(state.ID))
	copy(protoState.Id, state.ID[:])
	protoState.Version = state.Version
	protoState.IsFinal = state.IsFinal
	protoState.Allocation, err = fromAllocation(state.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "allocation")
	}
	protoState.App, protoState.Data, err = fromAppAndData(state.App, state.Data)
	return protoState, err
}

func fromAllocation(alloc pchannel.Allocation) (protoAlloc *pb.Allocation, err error) {
	protoAlloc = &pb.Allocation{}
	protoAlloc.Assets = make([][]byte, len(alloc.Assets))
	for i := range alloc.Assets {
		protoAlloc.Assets[i], err = alloc.Assets[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th asset", i)
		}
	}
	locked := make([]*pb.SubAlloc, len(alloc.Locked))
	for i := range alloc.Locked {
		locked[i], err = fromSubAlloc(alloc.Locked[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th sub alloc", i)
		}
	}
	protoAlloc.Balances, err = fromBalances(alloc.Balances)
	return protoAlloc, err
}

func fromBalances(balances pchannel.Balances) (protoBalances *pb.Balances, err error) {
	protoBalances = &pb.Balances{
		Balances: make([]*pb.Balance, len(balances)),
	}
	for i := range balances {
		protoBalances.Balances[i], err = fromBalance(balances[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th balance", i)
		}
	}
	return protoBalances, nil
}

func fromBalance(balance []pchannel.Bal) (protoBalance *pb.Balance, err error) {
	protoBalance = &pb.Balance{
		Balance: make([][]byte, len(balance)),
	}
	for i := range balance {
		if balance[i] == nil {
			return nil, fmt.Errorf("%d'th amount is nil", i) //nolint:goerr113  // We do not want to define this as constant error.
		}
		if balance[i].Sign() == -1 {
			return nil, fmt.Errorf("%d'th amount is negative", i) //nolint:goerr113  // We do not want to define this as constant error.
		}
		protoBalance.Balance[i] = balance[i].Bytes()
	}
	return protoBalance, nil
}

func fromAppAndData(app pchannel.App, data pchannel.Data) (protoApp, protoData []byte, err error) {
	if channel.IsNoApp(app) {
		return []byte{}, []byte{}, nil
	}
	protoApp, err = app.Def().MarshalBinary()
	if err != nil {
		return []byte{}, []byte{}, err
	}
	protoData, err = data.MarshalBinary()
	return protoApp, protoData, err
}

func fromSubAlloc(subAlloc channel.SubAlloc) (protoSubAlloc *pb.SubAlloc, err error) {
	protoSubAlloc = &pb.SubAlloc{}
	protoSubAlloc.Id = make([]byte, len(subAlloc.ID))
	copy(protoSubAlloc.Id, subAlloc.ID[:])
	protoSubAlloc.IndexMap = &pb.IndexMap{IndexMap: fromIndexMap(subAlloc.IndexMap)}
	protoSubAlloc.Bals, err = fromBalance(subAlloc.Bals)
	return protoSubAlloc, err
}

func fromIndexMap(indexMap []channel.Index) (protoIndexMap []uint32) {
	protoIndexMap = make([]uint32, len(indexMap))
	for i := range indexMap {
		protoIndexMap[i] = uint32(indexMap[i])
	}
	return protoIndexMap
}
