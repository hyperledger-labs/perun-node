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

package session

import "time"

// processingTime is to accommodate for computational and communications delays.
// This also includes the timespent spent waiting for a mutex.
var processingTime = 5 * time.Second

type timeoutConfig struct {
	onChainTx time.Duration
	response  time.Duration
}

func (t timeoutConfig) proposeCh(challegeDurSecs uint64) time.Duration {
	challegeDur := time.Duration(challegeDurSecs) * time.Second
	// The worst case path considered is
	// 1. Connect to the peer and expect acknowledgement,
	// 2. Send channel proposal and wait for response.
	// 3. Fund the channel and wait for challengeDurSecs for other to fund.
	// 4. Somebody does not fund, so register the initial state and wait for challengeDurSecs.
	// 5. Withdraw funds after challenge duration.
	return 2*t.response + 3*t.onChainTx + 2*challegeDur + processingTime
}

func (t timeoutConfig) respChProposalAccept(challegeDurSecs uint64) time.Duration {
	// The worst case path for accept is almost same as that for propose channel.
	return t.proposeCh(challegeDurSecs)
}

func (t timeoutConfig) respChProposalReject() time.Duration {
	// The time taken to send response.
	return t.response + processingTime
}

func (t timeoutConfig) chUpdate() time.Duration {
	// The time taken to receive the response.
	return t.response + processingTime
}

func (t timeoutConfig) respChUpdate() time.Duration {
	// The time taken to send response.
	return t.response + processingTime
}

func (t timeoutConfig) settle(challegeDurSecs uint64) time.Duration {
	challegeDur := time.Duration(challegeDurSecs) * time.Second

	// Register is implicitly called when settling. Hence, include a timeout
	// for the same.
	// The worst case path considered for register is
	// 1. Register state on blockchain
	// 2. and wait for challenge duration to expire.
	registerTimeout := 1*t.onChainTx + 1*challegeDur

	// The worst case path considered for settle is
	// 3. Conclude the final state on blockchain.
	// 2. Wait for challenge duration to expire.
	// 4. Withdraw amount.
	settleTimeout := 2*t.onChainTx + 1*challegeDur

	return registerTimeout + settleTimeout + processingTime
}
