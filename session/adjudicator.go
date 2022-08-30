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

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
)

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

type grpcAdjudicator struct {
	apiKey string
	client pb.Payment_APIClient
}

func (a *grpcAdjudicator) Register(
	ctx context.Context,
	adjReq pchannel.AdjudicatorReq,
	signedStates []pchannel.SignedState,
) (err error) {
	protoReq := pb.RegisterReq{}
	protoReq.AdjReq, err = fromAdjReq(adjReq)
	if err != nil {
		err = errors.WithMessage(err, "parsing grpc adjudicator request")
		return perun.NewAPIErrUnknownInternal(err)
	}
	protoReq.SignedStates = make([]*pb.SignedState, len(signedStates))
	for i := range signedStates {
		protoReq.SignedStates[i], err = toSignedState(&signedStates[i])
		if err != nil {
			err = errors.WithMessagef(err, "parsing %d'th signed state", i)
			return perun.NewAPIErrUnknownInternal(err)
		}
	}
	protoReq.SessionID = a.apiKey

	resp, err := a.client.Register(context.Background(), &protoReq)
	if err != nil {
		err = errors.WithMessage(err, "sending the funding request")
		return perun.NewAPIErrUnknownInternal(err)
	}
	if resp.Error != nil && resp.Error.Message != "" {
		// TODO: Proper error handling
		err = errors.WithMessage(err, "registering the channel the channel")
		return perun.NewAPIErrUnknownInternal(errors.New(resp.Error.Message))
	}
	return nil
}

func (a *grpcAdjudicator) Withdraw(
	ctx context.Context,
	adjReq pchannel.AdjudicatorReq,
	stateMap pchannel.StateMap,
) (err error) {
	protoReq := pb.WithdrawReq{}
	protoReq.AdjReq, err = fromAdjReq(adjReq)
	if err != nil {
		err = errors.WithMessage(err, "parsing grpc adjudicator request")
		return perun.NewAPIErrUnknownInternal(err)
	}
	protoReq.StateMap, err = fromStateMap(stateMap)
	if err != nil {
		err = errors.WithMessage(err, "parsing grpc adjudicator request")
		return perun.NewAPIErrUnknownInternal(err)
	}
	protoReq.SessionID = a.apiKey

	resp, err := a.client.Withdraw(context.Background(), &protoReq)
	if err != nil {
		err = errors.WithMessage(err, "sending the funding request")
		return perun.NewAPIErrUnknownInternal(err)
	}
	if resp.Error != nil && resp.Error.Message != "" {
		// TODO: Proper error handling
		err = errors.WithMessage(err, "withdrawing the channel the channel")
		return perun.NewAPIErrUnknownInternal(errors.New(resp.Error.Message))
	}
	return nil
}

func (a *grpcAdjudicator) Progress(
	ctx context.Context,
	progReq pchannel.ProgressReq,
) (err error) {
	protoReq := pb.ProgressReq{}
	protoReq.AdjReq, err = fromAdjReq(progReq.AdjudicatorReq)
	if err != nil {
		err = errors.WithMessage(err, "parsing grpc adjudicator request")
		return perun.NewAPIErrUnknownInternal(err)
	}
	protoReq.NewState, err = fromState(progReq.NewState)
	if err != nil {
		err = errors.WithMessage(err, "parsing grpc adjudicator request")
		return perun.NewAPIErrUnknownInternal(err)
	}
	protoReq.Sig = progReq.Sig
	protoReq.SessionID = a.apiKey

	resp, err := a.client.Progress(context.Background(), &protoReq)
	if err != nil {
		err = errors.WithMessage(err, "sending the funding request")
		return perun.NewAPIErrUnknownInternal(err)
	}
	if resp.Error != nil && resp.Error.Message != "" {
		// TODO: Proper error handling
		err = errors.WithMessage(err, "progressing the channel the channel")
		return perun.NewAPIErrUnknownInternal(errors.New(resp.Error.Message))
	}
	return nil
}

func (a *grpcAdjudicator) Subscribe(
	ctx context.Context,
	chID pchannel.ID,
) (pchannel.AdjudicatorSubscription, error) {
	adjSubReq := &pb.SubscribeReq{
		SessionID: a.apiKey,
		ChID:      chID[:],
	}
	stream, err := a.client.Subscribe(ctx, adjSubReq)
	if err != nil {
		return nil, err
	}
	adjSubRelay := newAdjudicatorEventsSub(chID, a)
	func() {
		subscribeResp, err := stream.Recv()
		if err != nil {
			return
		}
		adjEvent, err := fromProtoSubscribeResp(subscribeResp)
		if err != nil {
			return
		}
		adjSubRelay.publish(adjEvent)

	}()
	return adjSubRelay, nil
}

// AdjudicatorSubscription interface {
// 	// Next returns the most recent past or next future event. If the subscription is
// 	// closed or any other error occurs, it should return nil.
// 	Next() AdjudicatorEvent

// 	// Err returns the error status of the subscription. After Next returns nil,
// 	// Err should be checked for an error.
// 	Err() error

// 	// Close closes the subscription. Any call to Next should immediately return
// 	// nil.
// 	Close() error
// }

func fromProtoSubscribeResp(protoResponse *pb.SubscribeResp,
) (adjEvent pchannel.AdjudicatorEvent, err error) {
	switch e := protoResponse.Response.(type) {
	case *pb.SubscribeResp_RegisteredEvent:
		adjEvent, err = toGrpcRegisteredEvent(e.RegisteredEvent)
	case *pb.SubscribeResp_ProgressedEvent:
		adjEvent, err = toGrpcProgressedEvent(e.ProgressedEvent)
	case *pb.SubscribeResp_ConcludedEvent:
		adjEvent, err = toGrpcConcludedEvent(e.ConcludedEvent)
	case *pb.SubscribeResp_Error:
		return nil, err
	default:
		return nil, errors.New("unknown even type")
	}
	return adjEvent, nil
}

func fromAdjReq(req pchannel.AdjudicatorReq) (protoReq *pb.AdjudicatorReq, err error) {
	protoReq = &pb.AdjudicatorReq{}

	if protoReq.Params, err = fromParams(req.Params); err != nil {
		return protoReq, err
	}
	if protoReq.Tx, err = fromTx(req.Tx); err != nil {
		return protoReq, err
	}
	if protoReq.Acc, err = req.Acc.Address().MarshalBinary(); err != nil {
		return protoReq, err
	}

	protoReq.Idx = uint32(protoReq.Idx)
	protoReq.Secondary = req.Secondary
	return protoReq, nil
}

func fromStateMap(req pchannel.StateMap) (protoReq []*pb.StateMap, err error) {
	protoReq = make([]*pb.StateMap, len(req))

	i := 0
	for id, state := range req {
		protoReq[i].Id = id[:]
		protoReq[i].State, err = fromState(state)
		if err != nil {
			return nil, err
		}
		i++
	}
	return protoReq, nil
}

func fromTx(req pchannel.Transaction) (protoReq *pb.Transaction, err error) {
	protoReq = &pb.Transaction{}

	if protoReq.State, err = fromState(req.State); err != nil {
		return protoReq, err
	}
	protoReq.Sigs = make([][]byte, len(req.Sigs))
	for i := range req.Sigs {
		protoReq.Sigs[i] = []byte(req.Sigs[i])
	}
	return protoReq, nil
}

func toSignedState(signedState *pchannel.SignedState) (protoSignedState *pb.SignedState, err error) {
	protoSignedState = &pb.SignedState{}
	protoSignedState.Params, err = fromParams(signedState.Params)
	if err != nil {
		return nil, err
	}
	protoSignedState.State, err = fromState(signedState.State)
	if err != nil {
		return nil, err
	}
	protoSignedState.Sigs = make([][]byte, len(signedState.Sigs))
	for i := range protoSignedState.Sigs {
		protoSignedState.Sigs[i] = []byte(signedState.Sigs[i])
	}
	return protoSignedState, nil

}
