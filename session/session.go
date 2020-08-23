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
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	pclient "perun.network/go-perun/client"
	psync "perun.network/go-perun/pkg/sync"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/client"
	"github.com/hyperledger-labs/perun-node/comm/tcp"
	"github.com/hyperledger-labs/perun-node/contacts/contactsyaml"
	"github.com/hyperledger-labs/perun-node/log"
)

// walletBackend for initializing user wallets and parsing off-chain addresses
// in incoming contacts. A package level unexported variable is used so that a
// test wallet backend can be set using a function defined in export_test.go.
// Because real backend have large unlocking times and hence tests take very long.
var walletBackend perun.WalletBackend

func init() {
	// This can be overridden (only) in tests by calling the SetWalletBackend function.
	walletBackend = ethereum.NewWalletBackend()
}

type (
	session struct {
		log.Logger
		psync.Mutex

		id         string
		timeoutCfg timeoutConfig
		user       perun.User
		chAsset    pchannel.Asset
		chClient   perun.ChannelClient
		contacts   perun.Contacts

		channels map[string]*channel

		chProposalResponders map[string]chProposalResponderEntry
	}

	// the fields in the type will be defined later.
	// for now it is just a stub.
	chProposalResponderEntry struct {
	}
)

// New initializes a SessionAPI instance for the given configuration and returns an
// instance of it. All methods on it are safe for concurrent use.
func New(cfg Config) (*session, error) {
	user, err := NewUnlockedUser(walletBackend, cfg.User)
	if err != nil {
		return nil, err
	}

	if cfg.User.CommType != "tcp" {
		return nil, perun.ErrUnsupportedCommType
	}
	commBackend := tcp.NewTCPBackend(30 * time.Second)

	chAsset, err := walletBackend.ParseAddr(cfg.Asset)
	if err != nil {
		return nil, err
	}

	contacts, err := initContacts(cfg.ContactsType, cfg.ContactsURL, walletBackend, user.Peer)
	if err != nil {
		return nil, err
	}

	chClientCfg := client.Config{
		Chain: client.ChainConfig{
			Adjudicator:      cfg.Adjudicator,
			Asset:            cfg.Asset,
			URL:              cfg.ChainURL,
			ConnTimeout:      cfg.ChainConnTimeout,
			OnChainTxTimeout: cfg.OnChainTxTimeout,
		},
		DatabaseDir:       cfg.DatabaseDir,
		PeerReconnTimeout: cfg.PeerReconnTimeout,
	}
	chClient, err := client.NewEthereumPaymentClient(chClientCfg, user, commBackend)
	if err != nil {
		return nil, err
	}

	sessionID := calcSessionID(user.OffChainAddr.Bytes())
	timeoutCfg := timeoutConfig{
		onChainTx: cfg.OnChainTxTimeout,
		response:  cfg.ResponseTimeout,
	}
	sess := &session{
		Logger:               log.NewLoggerWithField("session-id", sessionID),
		id:                   sessionID,
		timeoutCfg:           timeoutCfg,
		user:                 user,
		chAsset:              chAsset,
		chClient:             chClient,
		contacts:             contacts,
		channels:             make(map[string]*channel),
		chProposalResponders: make(map[string]chProposalResponderEntry),
	}
	chClient.Handle(sess, sess) // Init handlers
	return sess, nil
}

func initContacts(contactsType, contactsURL string, wb perun.WalletBackend, own perun.Peer) (perun.Contacts, error) {
	if contactsType != "yaml" {
		return nil, perun.ErrUnsupportedContactsType
	}
	contacts, err := contactsyaml.New(contactsURL, wb)
	if err != nil {
		return nil, err
	}

	own.Alias = perun.OwnAlias
	err = contacts.Write(perun.OwnAlias, own)
	if err != nil && !errors.Is(err, perun.ErrPeerExists) {
		return nil, errors.Wrap(err, "registering own user in contacts")
	}
	return contacts, nil
}

// calcSessionID calculates the sessionID as sha256 hash over the off-chain address of the user and
// the current UTC time.
//
// A time dependant parameter is required to ensure the same user is able to open multiple sessions
// with the same node and have unique session id for each.
func calcSessionID(userOffChainAddr []byte) string {
	h := sha256.New()
	_, _ = h.Write(userOffChainAddr)                  // nolint:errcheck		// this func does not err
	_, _ = h.Write([]byte(time.Now().UTC().String())) // nolint:errcheck		// this func does not err
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s *session) ID() string {
	return ""
}

func (s *session) AddContact(peer perun.Peer) error {
	return nil
}

func (s *session) GetContact(alias string) (perun.Peer, error) {
	return perun.Peer{}, nil
}

func (s *session) OpenCh(
	pctx context.Context,
	peerAlias string,
	openingBals perun.BalInfo,
	app perun.App,
	challengeDurSecs uint64) (perun.ChannelInfo, error) {
	return perun.ChannelInfo{}, nil
}

func (s *session) HandleProposal(chProposal *pclient.ChannelProposal, responder *pclient.ProposalResponder) {
}

func (s *session) SubChProposals(notifier perun.ChProposalNotifier) error {
	return nil
}

func (s *session) UnsubChProposals() error {
	return nil
}

func (s *session) RespondChProposal(pctx context.Context, chProposalID string, accept bool) error {
	return nil
}

func (s *session) GetChInfos() []perun.ChannelInfo {
	return nil
}

func (s *session) GetCh(channelID string) (perun.ChannelAPI, error) {
	return nil, nil
}

func (s *session) HandleUpdate(chUpdate pclient.ChannelUpdate, responder *pclient.UpdateResponder) {
}

func (s *session) Close(force bool) error {
	return nil
}

func (s *session) HandleClose(chID string, err error) {
}

func (s *session) SubChCloses(notifier perun.ChCloseNotifier) error {
	return nil
}

func (s *session) UnsubChCloses() error {
	return nil
}
