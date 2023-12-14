// Copyright (c) 2020 - for information on the respective copyright owner
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
	"net"

	"github.com/pkg/errors"
	grpclib "google.golang.org/grpc"
	pchannel "perun.network/go-perun/channel"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/api/handlers"
)

// ServePaymentAPI starts a payment channel API server that listens for incoming grpc
// requests at the specified address and serves those requests using the node API instance.
func ServePaymentAPI(n perun.NodeAPI, grpcPort string) error {
	paymentChServer := &payChAPIServer{
		n:                n,
		chProposalsNotif: make(map[string]chan bool),
		chUpdatesNotif:   make(map[string]map[string]chan bool),
	}

	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return errors.Wrap(err, "starting listener")
	}
	grpcServer := grpclib.NewServer()
	pb.RegisterPayment_APIServer(grpcServer, paymentChServer)

	return grpcServer.Serve(listener)
}

// ServeFundingWatchingAPI starts a payment channel API server that listens for incoming grpc
// requests at the specified address and serves those requests using the node API instance.
func ServeFundingWatchingAPI(n perun.NodeAPI, grpcPort string) error {
	paymentChServer := &payChAPIServer{
		n:                n,
		chProposalsNotif: make(map[string]chan bool),
		chUpdatesNotif:   make(map[string]map[string]chan bool),
	}
	fundingServer := &fundingServer{
		FundingHandler: &handlers.FundingHandler{
			N:          n,
			Subscribes: make(map[string]map[pchannel.ID]pchannel.AdjudicatorSubscription),
		},
	}
	watchingServer := &watchingServer{
		WatchingHandler: &handlers.WatchingHandler{
			N:          n,
			Subscribes: make(map[string]map[pchannel.ID]pchannel.AdjudicatorSubscription),
		},
	}

	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return errors.Wrap(err, "starting listener")
	}
	grpcServer := grpclib.NewServer()
	pb.RegisterFunding_APIServer(grpcServer, fundingServer)
	pb.RegisterWatching_APIServer(grpcServer, watchingServer)
	pb.RegisterPayment_APIServer(grpcServer, paymentChServer)

	return grpcServer.Serve(listener)
}
