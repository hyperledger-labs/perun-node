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

package perunnode

import (
	psync "perun.network/go-perun/pkg/sync"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/log"
)

type node struct {
	log.Logger
	cfg      perun.NodeConfig
	sessions map[string]perun.SessionAPI
	psync.Mutex
}

// Time returns the current UTC time as per the node's system clock in Unix format.
func (n *node) Time() int64 {
	return 0
}

// GetConfig returns the node level configuration parameters.
func (n *node) GetConfig() perun.NodeConfig {
	return perun.NodeConfig{}
}

// Help returns the set of APIs offered by the node.
func (n *node) Help() []string {
	return nil
}

// OpenSession opens and session on this node uses the given config and returns the
// sessionID that can be used to access this session in subsequent calls.
func (n *node) OpenSession(configFile string) (string, error) {
	return "", nil
}

// GetSession is a special call that should be used internally to get the session instance.
// This should not be exposed to the user.
func (n *node) GetSession(sessionID string) (perun.SessionAPI, error) {
	return nil, nil
}
