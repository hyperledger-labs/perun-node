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

import (
	"time"

	pchannel "perun.network/go-perun/channel"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/log"
)

// SetWalletBackend is used to set a test wallet backend during tests.
func SetWalletBackend(wb perun.WalletBackend) {
	walletBackend = wb
}

func NewClientForTest(pClient pClient,
	bus perun.WireBus, msgBusRegistry perun.Registerer, dbConn Closer) ChClient {
	return &client{
		pClient:        pClient,
		msgBus:         bus,
		msgBusRegistry: msgBusRegistry,
		dbConn:         dbConn,
	}
}

func NewSessionForTest(cfg Config, isOpen bool, chClient ChClient) (*Session, error) {
	user, apiErr := NewUnlockedUser(walletBackend, cfg.User)
	if apiErr != nil {
		return nil, apiErr
	}

	idProvider, apiErr := initIDProvider(cfg.IDProviderType, cfg.IDProviderURL, walletBackend, user.PeerID)
	if apiErr != nil {
		return nil, apiErr
	}

	sessionID := calcSessionID(user.OffChainAddr.Bytes())
	timeoutCfg := timeoutConfig{
		onChainTx: cfg.OnChainTxTimeout,
		response:  cfg.ResponseTimeout,
	}

	return &Session{
		Logger:               log.NewLoggerWithField("session-id", sessionID),
		id:                   sessionID,
		isOpen:               isOpen,
		chainURL:             cfg.ChainURL,
		timeoutCfg:           timeoutCfg,
		user:                 user,
		chAsset:              cfg.Asset,
		chClient:             chClient,
		idProvider:           idProvider,
		chs:                  newChRegistry(initialChRegistrySize),
		chProposalResponders: make(map[string]chProposalResponderEntry),
	}, nil
}

func NewChForTest(pch PChannel,
	currency string, parts []string, responseTimeout time.Duration, challengeDurSecs uint64, isOpen bool) *Channel {
	chainURL := ethereumtest.ChainURL
	onChainTxTimeout := ethereumtest.OnChainTxTimeout
	timeoutCfg := timeoutConfig{
		response:  responseTimeout,
		onChainTx: onChainTxTimeout,
	}
	ch := newCh(pch, chainURL, currency, parts, timeoutCfg, challengeDurSecs)
	if isOpen {
		ch.status = open
	} else {
		ch.status = closed
	}
	ch.Logger = log.NewLoggerWithField("channel-id", ch.id)
	return ch
}

func MakeAllocation(openingBalInfo perun.BalInfo, chAsset pchannel.Asset) (*pchannel.Allocation, error) {
	return makeAllocation(openingBalInfo, chAsset)
}
