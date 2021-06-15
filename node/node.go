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

package node

import (
	"time"

	"github.com/pkg/errors"
	psync "perun.network/go-perun/pkg/sync"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/log"
	"github.com/hyperledger-labs/perun-node/session"
)

type node struct {
	log.Logger
	cfg      perun.NodeConfig
	sessions map[string]perun.SessionAPI
	psync.Mutex
}

// Error type is used to define error constants for this package.
type Error string

// Error implements error interface.
func (e Error) Error() string {
	return string(e)
}

// New returns a perun NodeAPI instance initialized using the given config.
// This should be called only once, subsequent calls after the first non error
// response will return an error.
func New(cfg perun.NodeConfig) (perun.NodeAPI, error) {
	// Currently, credentials are required for initializing a blockchain backend that
	// can validate the deployed contracts. Only a session has credentials and not a node.
	// So for now, check only if the addresses can be parsed.
	// TODO: (mano) Implement a read-only blockchain backend that can be initialized without
	// credentials and use it here.
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

// Time returns the time as per perun node's clock. It should be used to check
// the expiry of notifications.
func (n *node) Time() int64 {
	n.Debug("Received request: node.Time")
	return time.Now().UTC().Unix()
}

// GetConfig returns the configuration parameters of the node.
func (n *node) GetConfig() perun.NodeConfig {
	n.Debug("Received request: node.GetConfig")
	return n.cfg
}

// Help returns the list of user APIs served by the node.
func (n *node) Help() []string {
	n.Debug("Received request: node.Help")
	return []string{"payment"}
}

// Initializes a new session with the configuration in the given file. If
// channels were persisted during the previous instance of the session, they
// will be restored and their last known info will be returned.
//
// If there is an error, it will be one of the following codes:
// - ErrInvalidArgument with Name:"configFile" when config file cannot be accessed.
// - ErrInvalidConfig when any of the configuration is invalid.
// - ErrInvalidContracts when the contracts at the addresses in config are invalid.
// - ErrUnknownInternal.
func (n *node) OpenSession(configFile string) (string, []perun.ChInfo, perun.APIError) {
	n.WithField("method", "OpenSession").Infof("\nReceived request with params %+v", configFile)
	n.Lock()
	defer n.Unlock()

	var apiErr perun.APIError
	defer func() {
		if apiErr != nil {
			n.WithFields(perun.APIErrAsMap("OpenSession", apiErr)).Error(apiErr.Message())
		}
	}()

	sessionConfig, err := session.ParseConfig(configFile)
	if err != nil {
		err = errors.WithMessage(err, "parsing config")
		return "", nil, perun.NewAPIErrInvalidArgument(err, session.ArgNameConfigFile, configFile)
	}
	sess, apiErr := session.New(sessionConfig)
	if apiErr != nil {
		return "", nil, apiErr
	}
	n.sessions[sess.ID()] = sess

	n.WithFields(log.Fields{"method": "OpenSession", "sessionID": sess.ID()}).Info("Session opened successfully")
	return sess.ID(), sess.GetChsInfo(), nil
}

// GetSession is an internal API that retreives the session API instance
// corresponding to the given session ID.
//
// The session instance is safe for concurrent user.
//
// If there is an error, it will be one of the following codes:
// - ErrResourceNotFound when the session ID is not known.
func (n *node) GetSession(sessionID string) (perun.SessionAPI, perun.APIError) {
	n.WithField("method", "GetSession").Info("Received request with params:", sessionID)

	n.Lock()
	sess, ok := n.sessions[sessionID]
	n.Unlock()
	if !ok {
		apiErr := perun.NewAPIErrResourceNotFound(session.ResTypeSession, sessionID)
		n.WithFields(perun.APIErrAsMap("GetSession (internal)", apiErr)).Error(apiErr.Message())
		return nil, apiErr
	}
	n.WithField("method", "GetSession").Info("Session retrieved:")
	return sess, nil
}
