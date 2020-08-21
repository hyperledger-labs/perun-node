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
	"time"

	psync "perun.network/go-perun/pkg/sync"

	"github.com/pkg/errors"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/log"
)

type node struct {
	log.Logger
	cfg      perun.NodeConfig
	sessions map[string]perun.SessionAPI
	psync.Mutex
}

// New returns a perun NodeAPI instance initialized using the given config.
// This should be called only once, subsequent calls after the first non error
// response will return an error.
func New(cfg perun.NodeConfig) (*node, error) { // nolint: golint
	// it is okay to return an exported function as it implements an exported interface.

	// To validate the contracts, credentials are required for connecting to the
	// blockchain, which only a session has.
	// For now, just check if the addresses are valid.
	wb := ethereum.NewWalletBackend()
	_, err := wb.ParseAddr(cfg.Adjudicator)
	if err != nil {
		return nil, errors.WithMessage(err, "default adjudicator address")
	}
	_, err = wb.ParseAddr(cfg.Asset)
	if err != nil {
		return nil, errors.WithMessage(err, "default adjudicator address")
	}

	err = log.InitLogger(cfg.LogLevel, cfg.LogFile)
	if err != nil {
		return nil, errors.WithMessage(err, "initializing logger for node")
	}
	return &node{
		Logger:   log.NewLoggerWithField("node", 1), // ID of the node is always 1.
		cfg:      cfg,
		sessions: make(map[string]perun.SessionAPI),
	}, nil
}

// Time returns the current UTC time as per the node's system clock in Unix format.
func (n *node) Time() int64 {
	n.Debug("Received request: node.Time")
	return time.Now().UTC().Unix()
}

// GetConfig returns the node level configuration parameters.
func (n *node) GetConfig() perun.NodeConfig {
	n.Logger.Debug("Received request: node.GetConfig")
	return n.cfg
}

// Help returns the set of APIs offered by the node.
func (n *node) Help() []string {
	return []string{"payment"}
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
