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

package grpc_test

import (
	"testing"

	"github.com/hyperledger-labs/perun-node/api/grpc"
)

func Test_InitDashboard(t *testing.T) {
	d := grpc.InitDashboard()
	d.PrintBlank()
	d.SessionOpened()

	d.PrintBlank()
	d.FundingRequest([]string{"car", "charger"}, []string{"3", "0"})
	d.FundingSuccessful()

	d.PrintBlank()
	d.WatchingRequest()
	d.WatchingSuccessful()

	d.PrintBlank()
	d.ChannelUpdated([]string{"car", "charger"}, []string{"2", "1"})
	d.ChannelUpdated([]string{"car", "charger"}, []string{"1", "2"})
	d.ChannelUpdated([]string{"car", "charger"}, []string{"0", "3"})

	d.PrintBlank()
	d.ChannelFinalized([]string{"car", "charger"}, []string{"0", "3"})

	d.PrintBlank()
	d.ChannelRegistered([]string{"car", "charger"}, []string{"2", "1"})

	d.PrintBlank()
	d.ChannelConcluded()

	d.PrintBlank()
	d.ChannelWithdrawn([]string{"car", "charger"}, []string{"2", "1"})
}
