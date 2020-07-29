// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/perun-node
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

package client

import "time"

// Config represents the configuration parameters for state channel client.
type Config struct {
	Chain ChainConfig

	// Path to directory containing persistence database.
	DatabaseDir string
	// Timeout for re-establishing all open channels (if any) that was persisted during the
	// previous running instance of the node.
	PeerReconnTimeout time.Duration
}

// ChainConfig represents the configuration parameters for connecting to blockchain.
type ChainConfig struct {
	// Addresses of on-chain contracts used for establishing state channel network.
	Adjudicator string
	Asset       string

	// URL for connecting to the blockchain node.
	URL string
	// ConnTimeout is the timeout used when dialing for new connections to the on-chain node.
	ConnTimeout time.Duration
}
