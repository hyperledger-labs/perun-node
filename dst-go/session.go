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

package main

import (
	"fmt"
	"time"

	"github.com/direct-state-transfer/dst-go/blockchain"
	"github.com/direct-state-transfer/dst-go/channel"
	"github.com/direct-state-transfer/dst-go/ethereum/keystore"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
)

// Session represents a user session that provides single point access to blockchain connection, identity store and channels of a user.
type Session struct {
	owner identity.OffChainID

	keyStore *keystore.KeyStore        //Identity of the user for on chain transactions
	idStore  *identity.OffChainIDStore //Identity of the user and all others for offchain transactions

	idVerified chan *channel.Instance //Incoming offchain connection requests after id verification
	listener   channel.Shutdown       //listener instance to call shutdown

	LibSignAddr types.Address //LibSignatures contract addressed for all owner initiated sessions

	maxConn uint32 //Maximum number of off chain connections
}

// NewSession initialises and returns a user session with ethAddr as owner.
// It also initialises a listener that can simultaneously have maxConn number of active offchain channels,
// and also instances to access keystore at keysdir and idstore in idFile
func NewSession(ethAddr types.Address, keysDir, idFile string, maxConn uint32) (session Session, err error) {

	keyStore, idStore, err := identity.NewSession(keysDir, idFile)
	if err != nil {
		return Session{}, err
	}

	keyPresent := keyStore.HasAddress(ethAddr.Address)
	if !keyPresent {
		err = fmt.Errorf("Address %s not found in specified keystore dir", ethAddr.Hex())
		return Session{}, err
	}

	selfID, idPresent := idStore.OffChainID(ethAddr)
	if !idPresent {
		err = fmt.Errorf("Address %s not found in specified idstore dir", ethAddr.Hex())
		return Session{}, err
	}
	selfID.KeyStore = keyStore
	//TODO - Passsword to be obtained as user input
	selfID.Password = ""

	idVerifiedConn, listener, err := channel.NewSession(selfID, channel.WebSocket, maxConn)
	if err != nil {
		err = fmt.Errorf("Channel store init error - %s", err.Error())
		return Session{}, err
	}

	logger.Debug("Deploying libSig for", selfID.OnChainID.Hex())
	sessionLibSignAddr, err := blockchain.SetupLibSignatures(LibSignAddr, BlockchainConn, selfID)
	if err != nil {
		err = fmt.Errorf("Cannot setup libSign contract %s", err.Error())
		return Session{}, err
	}
	logger.Debug("Deployed libSig for", selfID.OnChainID.Hex())

	session = Session{
		owner:       selfID,
		keyStore:    keyStore,
		idStore:     idStore,
		idVerified:  idVerifiedConn,
		listener:    listener,
		LibSignAddr: sessionLibSignAddr,
		maxConn:     maxConn,
	}

	//TODO : Session handler go routines
	go idVerifiedConnHandler(session.idVerified)

	return session, nil
}

func idVerifiedConnHandler(idVerifiedConn chan *channel.Instance) {

	sessionTimeout := time.Tick(60 * time.Minute)
	for {
		select {
		case newConn := <-idVerifiedConn:
			//TODO : Pass the connection for user confirmation via remote interface
			//For now, it is assumed that user confirms all incoming requests
			logger.Info("New Incoming connection - ", newConn.PeerID())
		case <-sessionTimeout:
			return
		}
	}
}
