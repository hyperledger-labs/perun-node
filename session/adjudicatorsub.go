// Copyright 2022 - See NOTICE file for copyright holders.
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

	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"perun.network/go-perun/channel"
	pchannel "perun.network/go-perun/channel"
)

var _ pchannel.AdjudicatorSubscription = &adjudicatorSub{}

type (
	adjudicatorSub struct {
		isOpen bool
		once   sync.Mutex

		pipe    chan channel.AdjudicatorEvent
		chID    pchannel.ID
		grpcAdj *grpcAdjudicator
	}
)

func newAdjudicatorEventsSub(chID pchannel.ID, grpcAdj *grpcAdjudicator) *adjudicatorSub {
	return &adjudicatorSub{
		isOpen:  true,
		pipe:    make(chan channel.AdjudicatorEvent, adjPubSubBufferSize),
		chID:    chID,
		grpcAdj: grpcAdj,
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
func (a *adjudicatorSub) publish(e channel.AdjudicatorEvent) {
	a.once.Lock()
	a.pipe <- e
	a.once.Unlock()
}

// close closes the publisher instance and the associated subscription. Any
// further call to publish, after a pub-sub is closed will panic.
func (a *adjudicatorSub) Close() error {
	a.once.Lock()
	defer a.once.Unlock()

	unsubReq := &pb.UnsubscribeReq{
		SessionID: a.grpcAdj.apiKey,
		ChID:      a.chID[:],
	}
	a.grpcAdj.client.Unsubscribe(context.Background(), unsubReq)
	// TODO: Handle error.

	return nil
}

// EventStream returns a channel for consuming the published adjudicator
// events. It always returns the same channel and does not support
// multiplexing.
//
// The channel will be closed when the pub-sub instance is closed and Err
// should tell the possible error.
func (a *adjudicatorSub) Next() pchannel.AdjudicatorEvent {
	return <-a.pipe
}

// Err always returns nil. Because, there will be no errors when closing a
// local subscription.
func (a *adjudicatorSub) Err() error {
	return nil // Getting an error is not implemented.
}
