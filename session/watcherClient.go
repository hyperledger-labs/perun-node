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
	"sync"

	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	pwatcher "perun.network/go-perun/watcher"

	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

type grpcWatcher struct {
	apiKey string
	client pb.Watching_APIClient
}

func (w *grpcWatcher) StartWatchingLedgerChannel( //nolint:funlen, gocognit
	_ context.Context,
	signedState pchannel.SignedState,
) (pwatcher.StatesPub, pwatcher.AdjudicatorSub, error) {
	stream, err := w.client.StartWatchingLedgerChannel(context.Background())
	if err != nil {
		return nil, nil, errors.Wrap(err, "connecting to the server")
	}
	protoReq, err := signedStateToLedgerChReq(signedState)
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

	var tx pchannel.Transaction
	var isOpen bool
	go func() {
		statesStream := statesPubSub.statesStream()
		for {
			tx, isOpen = <-statesStream
			if !isOpen {
				return
			}
			protoReq, err = txToProtoLedgerChReq(tx)
			if err != nil {
				return
			}

			// Handle error while sending notification.
			err = stream.Send(protoReq)
			if err != nil {
				return
			}
		}
	}()

	var req *pb.StartWatchingLedgerChannelResp
	var adjEvent pchannel.AdjudicatorEvent

	go func() {
	AdjEventsSubLoop:
		for {
			req, err = stream.Recv()
			if err != nil {
				err = errors.WithMessage(err, "reading published adj event data")
				break AdjEventsSubLoop
			}
			adjEvent, err = adjEventFromLedgerChResp(req)
			if err != nil {
				err = errors.WithMessage(err, "parsing published adj event data")
				break AdjEventsSubLoop
			}
			adjEventsPubSub.publish(adjEvent)
		}
	}()

	return statesPubSub, adjEventsPubSub, err
}

func (w *grpcWatcher) StartWatchingSubChannel(
	_ context.Context,
	_ pchannel.ID,
	_ pchannel.SignedState,
) (pwatcher.StatesPub, pwatcher.AdjudicatorSub, error) {
	return nil, nil, nil
}

func (w *grpcWatcher) StopWatching(_ context.Context, _ pchannel.ID) error {
	return nil
}

func txToProtoLedgerChReq(req pchannel.Transaction) (protoReq *pb.StartWatchingLedgerChannelReq, err error) {
	protoReq = &pb.StartWatchingLedgerChannelReq{}

	if protoReq.State, err = pb.FromState(req.State); err != nil {
		return protoReq, err
	}
	sigs := make([][]byte, len(req.Sigs))
	for i := range sigs {
		copy(sigs[i], req.Sigs[i])
	}
	return protoReq, nil
}

func signedStateToLedgerChReq(req pchannel.SignedState) (protoReq *pb.StartWatchingLedgerChannelReq, err error) {
	protoReq = &pb.StartWatchingLedgerChannelReq{}

	if protoReq.Params, err = pb.FromParams(req.Params); err != nil {
		return protoReq, err
	}
	if protoReq.State, err = pb.FromState(req.State); err != nil {
		return protoReq, err
	}
	sigs := make([][]byte, len(req.Sigs))
	for i := range sigs {
		copy(sigs[i], req.Sigs[i])
	}
	return protoReq, nil
}

func adjEventFromLedgerChResp(protoResponse *pb.StartWatchingLedgerChannelResp,
) (adjEvent pchannel.AdjudicatorEvent, err error) {
	switch e := protoResponse.Response.(type) {
	case *pb.StartWatchingLedgerChannelResp_RegisteredEvent:
		adjEvent, err = pb.ToRegisteredEvent(e.RegisteredEvent)
	case *pb.StartWatchingLedgerChannelResp_ProgressedEvent:
		adjEvent, err = pb.ToProgressedEvent(e.ProgressedEvent)
	case *pb.StartWatchingLedgerChannelResp_ConcludedEvent:
		adjEvent = pb.ToConcludedEvent(e.ConcludedEvent)
	case *pb.StartWatchingLedgerChannelResp_Error:
		return nil, err
	default:
		return nil, errors.New("unknown even type")
	}
	return adjEvent, err
}

var _ pwatcher.AdjudicatorSub = &adjudicatorPubSub{}

const adjPubSubBufferSize = 10

type (
	adjudicatorPubSub struct {
		once sync.Once //nolint:unused
		pipe chan pchannel.AdjudicatorEvent
	}
)

func newAdjudicatorEventsPubSub() *adjudicatorPubSub {
	return &adjudicatorPubSub{
		pipe: make(chan pchannel.AdjudicatorEvent, adjPubSubBufferSize),
	}
}

// publish publishes the given adjudicator event to the subscriber.
//
// Panics if the pub-sub instance is already closed. It is implemented this
// way, because
//  1. The watcher will publish on this pub-sub only when it receives an
//     adjudicator event from the blockchain.
//  2. When de-registering a channel from the watcher, watcher will close the
//     subscription for adjudicator events from blockchain, before closing this
//     pub-sub.
//  3. This way, it can be guaranteed that, this method will never be called
//     after the pub-sub instance is closed.
func (a *adjudicatorPubSub) publish(e pchannel.AdjudicatorEvent) {
	a.pipe <- e
}

// close closes the publisher instance and the associated subscription. Any
// further call to publish, after a pub-sub is closed will panic.
func (a *adjudicatorPubSub) close() { //nolint:unused
	a.once.Do(func() { close(a.pipe) })
}

// EventStream returns a channel for consuming the published adjudicator
// events. It always returns the same channel and does not support
// multiplexing.
//
// The channel will be closed when the pub-sub instance is closed and Err
// should tell the possible error.
func (a *adjudicatorPubSub) EventStream() <-chan pchannel.AdjudicatorEvent {
	return a.pipe
}

// Err always returns nil. Because, there will be no errors when closing a
// local subscription.
func (a *adjudicatorPubSub) Err() error {
	return nil
}

var _ pwatcher.StatesPub = &statesPubSub{}

const statesPubSubBufferSize = 10

type (
	statesPubSub struct {
		once sync.Once //nolint:unused
		pipe chan pchannel.Transaction
	}
)

func newStatesPubSub() *statesPubSub {
	return &statesPubSub{
		pipe: make(chan pchannel.Transaction, statesPubSubBufferSize),
	}
}

// Publish publishes the given transaction (state and signatures on it) to the
// subscriber.
//
// Always returns nil. Error result is for implementing watcher.StatesPub.
//
// Panics if the pub-sub instance is already closed. It is implemented this
// way, because
//  1. Watcher requires that, the Publish method must not be called after stop
//     watching for a channel. See docs of watcher.StatesPub for more details.
//  2. Hence, by properly integrating the watcher into the client, it can be
//     guaranteed that, this method will never be called after the pub-sub
//     instance is closed and that, this method will never panic.
func (s *statesPubSub) Publish(_ context.Context, tx pchannel.Transaction) error {
	s.pipe <- tx
	return nil
}

// close closes the publisher instance and the associated subscription. Any
// further call to Publish, after a pub-sub is closed will panic.
func (s *statesPubSub) close() { //nolint:unused
	s.once.Do(func() { close(s.pipe) })
}

// statesStream returns a channel for consuming the published states. It always
// returns the same channel and does not support multiplexing.
func (s *statesPubSub) statesStream() <-chan pchannel.Transaction {
	return s.pipe
}
