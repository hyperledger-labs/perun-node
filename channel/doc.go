// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/dst-go
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

// Package channel implements node to node (offchain) communication functionality for the node software.
//
// It manages the lifecycle of a channel, functions to send and receive different kinds of message over the channel,
// adapters to support different transport layer protocols such as websocket. It also includes functions for initializing
// listeners for handling new incoming connections and to initiate outgoing connections.
// The primitives sub package defines message formats and other structures (such as state, session)
// required for offchain communication., message packets & its parsers as well the adapter implementations.
package channel
