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

	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node"
)

// Fund provides a wrapper to call the fund method on session.
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

// Register provides a wrapper to call the register method on session.
// On-chain wallet can be direcly used without additional locks,
// because the methods on wallet are concurrency safe by themselves.
func (s *Session) Register(
	ctx context.Context,
	adjReq perun.AdjudicatorReq,
	signedStates []pchannel.SignedState,
) perun.APIError {
	s.Infof("\ncar: register request for channel with charger, balance: %+v", adjReq.Tx.Balances)
	// s.WithField("method", "Register").Infof("\nReceived request with params %+v, %+v", adjReq, signedStates)

	pAdjReq, err := toPChannelAdjudicatorReq(adjReq, s.user.OffChain.Wallet)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("Register", apiErr)).Error(apiErr.Message())
		return apiErr
	}

	err = s.adjudicator.Register(ctx, pAdjReq, signedStates)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("Register", apiErr)).Error(apiErr.Message())
		return apiErr
	}
	s.WithField("method", "Register").Infof("Registered successfully: %+v ", adjReq.Params.ID())
	s.Infof("\nregistered successfully")
	return nil
}

// Withdraw provides a wrapper to call the withdraw method on session.
// On-chain wallet can be direcly used without additional locks,
// because the methods on wallet are concurrency safe by themselves.
func (s *Session) Withdraw(
	ctx context.Context,
	adjReq perun.AdjudicatorReq,
	stateMap pchannel.StateMap,
) perun.APIError {
	s.Infof("\ncar: withdraw request for channel with charger, balance: %+v", adjReq.Tx.Balances)
	// s.WithField("method", "Withdraw").Infof("\nReceived request with params %+v, %+v", adjReq, stateMap)

	pAdjReq, err := toPChannelAdjudicatorReq(adjReq, s.user.OffChain.Wallet)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("Withdraw", apiErr)).Error(apiErr.Message())
		return apiErr
	}

	err = s.adjudicator.Withdraw(ctx, pAdjReq, stateMap)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("Withdraw", apiErr)).Error(apiErr.Message())
		return apiErr
	}
	s.user.OnChain.Wallet.DecrementUsage(s.user.OnChain.Addr)
	// s.WithField("method", "Withdraw").Infof("Withdrawn successfully: %+v ", adjReq.Params.ID())
	s.Infof("\nwithdrawn successfully")
	return nil
}

// Progress provides a wrapper to call the progress method on session.
// On-chain wallet can be direcly used without additional locks,
// because the methods on wallet are concurrency safe by themselves.
func (s *Session) Progress(ctx context.Context, progReq perun.ProgressReq) perun.APIError {
	s.WithField("method", "Progress").Infof("\nReceived request with params %+v", progReq)

	pProgReq, err := toPChannelProgressReq(progReq, s.user.OffChain.Wallet)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("Progress", apiErr)).Error(apiErr.Message())
		return apiErr
	}

	err = s.adjudicator.Progress(ctx, pProgReq)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("Progress", apiErr)).Error(apiErr.Message())
		return apiErr
	}
	s.WithField("method", "Progress").Infof("Progressed successfully: %+v ", progReq.Params.ID())
	return nil
}

// Subscribe provides a wrapper to call the subscribe method on session.
func (s *Session) Subscribe(
	ctx context.Context,
	chID pchannel.ID,
) (pchannel.AdjudicatorSubscription, perun.APIError) {
	s.Infof("\ncar: subscribe request for channel with charger")
	// s.WithField("method", "Subscribe").Infof("\nReceived request with params %+v", chID)

	adjSub, err := s.adjudicator.Subscribe(ctx, chID)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("Progress", apiErr)).Error(apiErr.Message())
		return nil, apiErr
	}
	s.Infof("\nsubscribed successfully")
	return adjSub, nil
}

func toPChannelAdjudicatorReq(in perun.AdjudicatorReq, w pwallet.Wallet) (out pchannel.AdjudicatorReq, err error) {
	out.Acc, err = w.Unlock(in.Acc)
	if err != nil {
		return out, err
	}
	out.Params = in.Params
	out.Tx = in.Tx
	out.Idx = in.Idx
	out.Secondary = in.Secondary
	return out, nil
}

func toPChannelProgressReq(in perun.ProgressReq, w pwallet.Wallet) (out pchannel.ProgressReq, err error) {
	out.AdjudicatorReq, err = toPChannelAdjudicatorReq(in.AdjudicatorReq, w)
	if err != nil {
		return out, err
	}
	out.NewState = in.NewState
	out.Sig = in.Sig
	return out, nil
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
