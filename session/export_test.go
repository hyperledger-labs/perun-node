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
	"github.com/hyperledger-labs/perun-node/log"
)

// SetWalletBackend is used to set a test wallet backend during tests.
func SetWalletBackend(wb perun.WalletBackend) {
	walletBackend = wb
}

func NewSessionForTest(cfg Config, isOpen bool, chClient perun.ChClient) (*Session, error) {
	user, err := NewUnlockedUser(walletBackend, cfg.User)
	if err != nil {
		return nil, err
	}

	chAsset, err := walletBackend.ParseAddr(cfg.Asset)
	if err != nil {
		return nil, err
	}

	idProvider, err := initIDProvider(cfg.IDProviderType, cfg.IDProviderURL, walletBackend, user.PeerID)
	if err != nil {
		return nil, err
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
		timeoutCfg:           timeoutCfg,
		user:                 user,
		chAsset:              chAsset,
		chClient:             chClient,
		idProvider:           idProvider,
		chs:                  make(map[string]*Channel),
		chProposalResponders: make(map[string]chProposalResponderEntry),
	}, nil
}

func NewChForTest(pch perun.Channel, currency string, parts []string, challengeDurSecs uint64, isOpen bool) *Channel {
	timeoutCfg := timeoutConfig{response: 1 * time.Second}
	ch := newCh(pch, currency, parts, timeoutCfg, challengeDurSecs)
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
