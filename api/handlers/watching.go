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

package handlers

import (
	"context"

	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
	psync "polycry.pt/poly-go/sync"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

// WatchingHandler represents a grpc server that can serve watching API.
type WatchingHandler struct {
	N perun.NodeAPI

	// The mutex should be used when accessing the map data structures.
	psync.Mutex
	Subscribes map[string]map[pchannel.ID]pchannel.AdjudicatorSubscription
}

// StartWatchingLedgerChannel wraps session.StartWatchingLedgerChannel.
func (a *WatchingHandler) StartWatchingLedgerChannel( //nolint: funlen, gocognit
	req *pb.StartWatchingLedgerChannelReq,
	sendAdjEvent func(resp *pb.StartWatchingLedgerChannelResp) error,
	receiveState func() (req *pb.StartWatchingLedgerChannelReq, err error),
) error {
	var err error
	sess, err := a.N.GetSession(req.SessionID)
	if err != nil {
		return errors.WithMessage(err, "retrieving session")
	}

	signedState, err := signedStateFromProtoLedgerChReq(req)
	if err != nil {
		return errors.WithMessage(err, "parsing signed state")
	}

	statesPub, adjSub, err := sess.StartWatchingLedgerChannel(context.TODO(), *signedState)
	if err != nil {
		return errors.WithMessage(err, "start watching")
	}

	// This stream is anyways closed when StopWatching is called for.
	// Hence, that will act as the exit condition for the loop.
	go func() {
		adjEventStream := adjSub.EventStream()
		var protoResponse *pb.StartWatchingLedgerChannelResp
		for {
			adjEvent, isOpen := <-adjEventStream
			if !isOpen {
				return
			}
			protoResponse, err = adjEventToProtoLedgerChResp(adjEvent)
			if err != nil {
				return
			}

			// Handle error while sending notification.
			err = sendAdjEvent(protoResponse)
			if err != nil {
				return
			}
		}
	}()

	// It should be the responsibility of the streamer to close things.
	// Hence, the client should be closing this stream, which will cause srv.Recv to return an error.
	// The error will act as the exit condition for this for{} loop.
	var tx *pchannel.Transaction
StatesPubLoop:
	for {

		req, err = receiveState()
		if err != nil {
			err = errors.WithMessage(err, "reading published states pub data")
			break StatesPubLoop
		}
		tx, err = transactionFromProtoLedgerChReq(req)
		if err != nil {
			err = errors.WithMessage(err, "parsing published states pub data")
			break StatesPubLoop
		}

		err = statesPub.Publish(context.TODO(), *tx)
		if err != nil {
			err = errors.WithMessage(err, "locally relaying published states pub data")
			break StatesPubLoop
		}
	}

	// TODO: Ensure adjEventSteam go-routine is killed. It should not be leaking.
	// It is to allow for that, label breaks are used instead of return statements in the above for-select.
	return err
}

// StopWatching wraps session.StopWatching.
func (a *WatchingHandler) StopWatching(ctx context.Context, req *pb.StopWatchingReq) (*pb.StopWatchingResp, error) {
	errResponse := func(err perun.APIError) *pb.StopWatchingResp {
		return &pb.StopWatchingResp{
			Error: pb.FromError(err),
		}
	}

	sess, err := a.N.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	var chID pchannel.ID
	copy(chID[:], req.ChID)
	err2 := sess.StopWatching(ctx, chID)
	if err2 != nil {
		return errResponse(err), nil
	}

	return &pb.StopWatchingResp{Error: nil}, nil
}

func adjEventToProtoLedgerChResp(adjEvent pchannel.AdjudicatorEvent) (*pb.StartWatchingLedgerChannelResp, error) {
	protoResponse := &pb.StartWatchingLedgerChannelResp{}
	switch e := adjEvent.(type) {
	case *pchannel.RegisteredEvent:
		registeredEvent, err := pb.FromRegisteredEvent(e)
		protoResponse.Response = &pb.StartWatchingLedgerChannelResp_RegisteredEvent{
			RegisteredEvent: registeredEvent,
		}
		return protoResponse, err
	case *pchannel.ProgressedEvent:
		progressedEvent, err := pb.FromProgressedEvent(e)
		protoResponse.Response = &pb.StartWatchingLedgerChannelResp_ProgressedEvent{
			ProgressedEvent: progressedEvent,
		}
		return protoResponse, err
	case *pchannel.ConcludedEvent:
		concludedEvent, err := pb.FromConcludedEvent(e)
		protoResponse.Response = &pb.StartWatchingLedgerChannelResp_ConcludedEvent{
			ConcludedEvent: concludedEvent,
		}
		return protoResponse, err
	default:
		apiErr := perun.NewAPIErrUnknownInternal(errors.New("unknown even type"))
		protoResponse.Response = &pb.StartWatchingLedgerChannelResp_Error{
			Error: pb.FromError(apiErr),
		}
		return protoResponse, nil
	}
}

func signedStateFromProtoLedgerChReq(req *pb.StartWatchingLedgerChannelReq) (
	signedState *pchannel.SignedState, err error,
) {
	signedState = &pchannel.SignedState{}
	signedState.Params, err = pb.ToParams(req.Params)
	if err != nil {
		return nil, err
	}
	signedState.State, err = pb.ToState(req.State)
	if err != nil {
		return nil, err
	}
	sigs := make([]pwallet.Sig, len(req.Sigs))
	for i := range sigs {
		copy(sigs[i], req.Sigs[i])
	}
	return signedState, nil
}

func transactionFromProtoLedgerChReq(req *pb.StartWatchingLedgerChannelReq) (
	transaction *pchannel.Transaction, err error,
) {
	transaction = &pchannel.Transaction{}
	transaction.State, err = pb.ToState(req.State)
	if err != nil {
		return nil, err
	}
	transaction.Sigs = make([]pwallet.Sig, len(req.Sigs))
	for i := range transaction.Sigs {
		copy(transaction.Sigs[i], req.Sigs[i])
	}
	return transaction, nil
}
