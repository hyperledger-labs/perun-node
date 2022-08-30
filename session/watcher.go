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
	"math"
	"math/big"
	"time"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/pkg/errors"
	pethchannel "perun.network/go-perun/backend/ethereum/channel"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
	pwatcher "perun.network/go-perun/watcher"
)

func (s *Session) StartWatchingLedgerChannel(
	ctx context.Context,
	signedState pchannel.SignedState,
) (pwatcher.StatesPub, pwatcher.AdjudicatorSub, perun.APIError) {
	s.Infof("\ncar: start watching request for channel with charger")
	// s.WithField("method", "StartWatchingLedgerChannel").Infof("\nReceived request with params %+v", signedState)
	statesPub, adjSub, err := s.watcher.StartWatchingLedgerChannel(ctx, signedState)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("StartWatchingLedgerChannel", apiErr)).Error(apiErr.Message())
		return nil, nil, apiErr
	}
	// s.WithField("method", "StartWatchingLedgerChannel").Infof("Started watching for ledger channel %+v",
	// 	signedState.State.ID)
	s.Infof("\nwatcher started successfully")
	return statesPub, adjSub, nil
}

func (s *Session) StartWatchingSubChannel(
	ctx context.Context,
	parent pchannel.ID,
	signedState pchannel.SignedState,
) (pwatcher.StatesPub, pwatcher.AdjudicatorSub, perun.APIError) {
	s.WithField("method", "StartWatchingSubChannel").Infof("\nReceived request with params %+v, %+v", parent, signedState)
	statesPub, adjSub, err := s.watcher.StartWatchingSubChannel(ctx, parent, signedState)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("StartWatchingSubChannel", apiErr)).Error(apiErr.Message())
		return nil, nil, apiErr
	}
	s.WithField("method", "StartWatchingSubChannel").Infof("Started watching for sub channel %+v",
		signedState.State.ID)
	return statesPub, adjSub, nil
}

func (s *Session) StopWatching(ctx context.Context, chID pchannel.ID) perun.APIError {
	s.Infof("\ncar: stop watching request for channel with charger")
	// s.WithField("method", "StopWatching").Infof("\nReceived request with params %+v", chID)
	err := s.watcher.StopWatching(ctx, chID)
	if err != nil {
		apiErr := perun.NewAPIErrUnknownInternal(err)
		s.WithFields(perun.APIErrAsMap("UnsubChProposals", apiErr)).Error(apiErr.Message())
		return apiErr
	}
	// s.WithField("method", "StopWatching").Infof("Stopped watching for channel %+v", chID)
	s.Infof("\nwatcher stopped successfully")
	return nil
}

type grpcWatcher struct {
	apiKey string
	client pb.Payment_APIClient
}

func (w *grpcWatcher) StartWatchingLedgerChannel(
	ctx context.Context,
	signedState pchannel.SignedState,
) (pwatcher.StatesPub, pwatcher.AdjudicatorSub, error) {
	stream, err := w.client.StartWatchingLedgerChannel(context.Background())
	if err != nil {
		return nil, nil, errors.Wrap(err, "connecting to the server")
	}
	protoReq, err := toProtoSignedState(signedState)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parsing to proto request")
	}
	protoReq.SessionID = w.apiKey

	// Parameter for start watching call is sent as first client stream.
	err = stream.Send(protoReq)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parsing to proto request")
	}

	statesPubSub := newStatesPubSub()
	adjEventsPubSub := newAdjudicatorEventsPubSub()

	// CLOSE SHOULD BE CALLED ON THIS ???? SO STORE THEM IN WATCHER ???

	// This stream is anyways closed when StopWatching is called for.
	// Hence, that will act as the exit condition for the loop.
	go func() {
		statesStream := statesPubSub.statesStream()
		for {
			select {
			case tx, isOpen := <-statesStream:
				if !isOpen {
					return
				}
				protoReq, err := toProtoTx(tx)
				if err != nil {
					return
				}

				// Handle error while sending notification.
				err = stream.Send(protoReq)
				if err != nil {
					return
				}
			}
		}
	}()

	go func() {
	AdjEventsSubLoop:
		for {
			req, err := stream.Recv()
			if err != nil {
				err = errors.WithMessage(err, "reading published adj event data")
				break AdjEventsSubLoop
			}
			adjEvent, err := fromProtoResponse(req)
			if err != nil {
				err = errors.WithMessage(err, "parsing published adj event data")
				break AdjEventsSubLoop
			}
			adjEventsPubSub.publish(adjEvent)
		}
	}()

	return statesPubSub, adjEventsPubSub, nil
}

func (w *grpcWatcher) StartWatchingSubChannel(
	ctx context.Context,
	parent pchannel.ID,
	signedState pchannel.SignedState,
) (pwatcher.StatesPub, pwatcher.AdjudicatorSub, error) {
	return nil, nil, nil
}

func (w *grpcWatcher) StopWatching(ctx context.Context, chID pchannel.ID) error {
	return nil
}

func toProtoTx(req pchannel.Transaction) (protoReq *pb.StartWatchingLedgerChannelReq, err error) {
	protoReq = &pb.StartWatchingLedgerChannelReq{}

	if protoReq.State, err = fromState(req.State); err != nil {
		return protoReq, err
	}
	sigs := make([][]byte, len(req.Sigs))
	for i := range sigs {
		sigs[i] = []byte(req.Sigs[i])
	}
	return protoReq, nil
}

func toProtoSignedState(req pchannel.SignedState) (protoReq *pb.StartWatchingLedgerChannelReq, err error) {
	protoReq = &pb.StartWatchingLedgerChannelReq{}

	if protoReq.Params, err = fromParams(req.Params); err != nil {
		return protoReq, err
	}
	if protoReq.State, err = fromState(req.State); err != nil {
		return protoReq, err
	}
	sigs := make([][]byte, len(req.Sigs))
	for i := range sigs {
		sigs[i] = []byte(req.Sigs[i])
	}
	return protoReq, nil
}

func fromProtoResponse(protoResponse *pb.StartWatchingLedgerChannelResp,
) (adjEvent pchannel.AdjudicatorEvent, err error) {
	switch e := protoResponse.Response.(type) {
	case *pb.StartWatchingLedgerChannelResp_RegisteredEvent:
		adjEvent, err = toGrpcRegisteredEvent(e.RegisteredEvent)
	case *pb.StartWatchingLedgerChannelResp_ProgressedEvent:
		adjEvent, err = toGrpcProgressedEvent(e.ProgressedEvent)
	case *pb.StartWatchingLedgerChannelResp_ConcludedEvent:
		adjEvent, err = toGrpcConcludedEvent(e.ConcludedEvent)
	case *pb.StartWatchingLedgerChannelResp_Error:
		return nil, err
	default:
		return nil, errors.New("unknown even type")
	}
	return adjEvent, nil
}

func toGrpcRegisteredEvent(protoEvent *pb.RegisteredEvent) (event *pchannel.RegisteredEvent, err error) {
	event = &pchannel.RegisteredEvent{}
	event.AdjudicatorEventBase = toGrpcAdjudicatorEventBase(protoEvent.AdjudicatorEventBase)
	event.State, err = toState(protoEvent.State)
	if err != nil {
		return nil, errors.WithMessage(err, "parsing state")
	}
	event.Sigs = make([]pwallet.Sig, len(protoEvent.Sigs))
	for i := range protoEvent.Sigs {
		event.Sigs[i] = pwallet.Sig(protoEvent.Sigs[i])
	}
	return event, nil
}

func toGrpcProgressedEvent(protoEvent *pb.ProgressedEvent) (event *pchannel.ProgressedEvent, err error) {
	event = &pchannel.ProgressedEvent{}
	event.AdjudicatorEventBase = toGrpcAdjudicatorEventBase(protoEvent.AdjudicatorEventBase)
	event.State, err = toState(protoEvent.State)
	if err != nil {
		return nil, errors.WithMessage(err, "parsing state")
	}
	event.Idx = pchannel.Index(protoEvent.Idx)
	return event, nil
}

func toGrpcConcludedEvent(protoEvent *pb.ConcludedEvent) (event *pchannel.ConcludedEvent, err error) {
	event = &pchannel.ConcludedEvent{}
	event.AdjudicatorEventBase = toGrpcAdjudicatorEventBase(protoEvent.AdjudicatorEventBase)
	return event, nil
}

func toGrpcAdjudicatorEventBase(protoEvent *pb.AdjudicatorEventBase,
) (event pchannel.AdjudicatorEventBase) {
	copy(event.IDV[:], protoEvent.ChID)
	event.VersionV = protoEvent.Version
	switch protoEvent.Timeout.Type {
	case pb.AdjudicatorEventBase_elapsed:
		event.TimeoutV = &pchannel.ElapsedTimeout{}
	case pb.AdjudicatorEventBase_time:
		event.TimeoutV = &pchannel.TimeTimeout{time.Unix(protoEvent.Timeout.Sec, 0)}
	case pb.AdjudicatorEventBase_ethBlock:
		// TODO: Use proper value. this is uint64 by default.
		event.TimeoutV = &pethchannel.BlockTimeout{Time: uint64(protoEvent.Timeout.Sec)}
	}
	return event
}

func toState(protoState *pb.State) (state *pchannel.State, err error) {
	state = &pchannel.State{}
	copy(state.ID[:], protoState.Id)
	state.Version = protoState.Version
	state.IsFinal = protoState.IsFinal
	allocation, err := toAllocation(protoState.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "allocation")
	}
	state.Allocation = *allocation
	state.App, state.Data, err = toAppAndData(protoState.App, protoState.Data)
	return state, err
}

func toAppAndData(protoApp, protoData []byte) (app pchannel.App, data pchannel.Data, err error) {
	if len(protoApp) == 0 {
		app = pchannel.NoApp()
		data = pchannel.NoData()
		return app, data, nil
	}
	appDef := pwallet.NewAddress()
	err = appDef.UnmarshalBinary(protoApp)
	if err != nil {
		return nil, nil, err
	}
	app, err = pchannel.Resolve(appDef)
	if err != nil {
		return
	}
	data = app.NewData()
	return app, data, data.UnmarshalBinary(protoData)
}

func toAllocation(protoAlloc *pb.Allocation) (alloc *pchannel.Allocation, err error) {
	alloc = &pchannel.Allocation{}
	alloc.Assets = make([]pchannel.Asset, len(protoAlloc.Assets))
	for i := range protoAlloc.Assets {
		alloc.Assets[i] = pchannel.NewAsset()
		err = alloc.Assets[i].UnmarshalBinary(protoAlloc.Assets[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th asset", i)
		}
	}
	alloc.Locked = make([]pchannel.SubAlloc, len(protoAlloc.Locked))
	for i := range protoAlloc.Locked {
		alloc.Locked[i], err = toSubAlloc(protoAlloc.Locked[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th sub alloc", i)
		}
	}
	alloc.Balances = toBalances(protoAlloc.Balances)
	return alloc, nil
}

func toBalances(protoBalances *pb.Balances) (balances pchannel.Balances) {
	balances = make([][]pchannel.Bal, len(protoBalances.Balances))
	for i := range protoBalances.Balances {
		balances[i] = toBalance(protoBalances.Balances[i])
	}
	return balances
}

func toBalance(protoBalance *pb.Balance) (balance []pchannel.Bal) {
	balance = make([]pchannel.Bal, len(protoBalance.Balance))
	for j := range protoBalance.Balance {
		balance[j] = new(big.Int).SetBytes(protoBalance.Balance[j])
	}
	return balance
}

func toSubAlloc(protoSubAlloc *pb.SubAlloc) (subAlloc pchannel.SubAlloc, err error) {
	subAlloc = pchannel.SubAlloc{}
	subAlloc.Bals = toBalance(protoSubAlloc.Bals)
	if len(protoSubAlloc.Id) != len(subAlloc.ID) {
		return subAlloc, errors.New("sub alloc id has incorrect length")
	}
	copy(subAlloc.ID[:], protoSubAlloc.Id)
	subAlloc.IndexMap, err = toIndexMap(protoSubAlloc.IndexMap.IndexMap)
	return subAlloc, err
}

func toIndexMap(protoIndexMap []uint32) (indexMap []pchannel.Index, err error) {
	indexMap = make([]pchannel.Index, len(protoIndexMap))
	for i := range protoIndexMap {
		if protoIndexMap[i] > math.MaxUint16 {
			return nil, fmt.Errorf("%d'th index is invalid", i) //nolint:goerr113  // We do not want to define this as constant error.
		}
		indexMap[i] = pchannel.Index(uint16(protoIndexMap[i]))
	}
	return indexMap, nil
}
