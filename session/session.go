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
	"math/big"
	"strings"
	"time"

	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	pclient "perun.network/go-perun/client"
	psync "perun.network/go-perun/pkg/sync"
	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/client"
	"github.com/hyperledger-labs/perun-node/comm/tcp"
	"github.com/hyperledger-labs/perun-node/contacts/contactsyaml"
	"github.com/hyperledger-labs/perun-node/currency"
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
		chClient   perun.ChClient
		contacts   perun.Contacts

		chs map[string]*channel

		chProposalNotifier    perun.ChProposalNotifier
		chProposalNotifsCache []perun.ChProposalNotif
		chProposalResponders  map[string]chProposalResponderEntry

		chCloseNotifier    perun.ChCloseNotifier
		chCloseNotifsCache []perun.ChCloseNotif
	}

	chProposalResponderEntry struct {
		notif     perun.ChProposalNotif
		responder chProposalResponder
	}

	//go:generate mockery --name chProposalResponder --output ../internal/mocks

	// Proposal Responder defines the methods on proposal responder that will be used by the perun node.
	chProposalResponder interface {
		Accept(context.Context, pclient.ProposalAcc) (*pclient.Channel, error)
		Reject(ctx context.Context, reason string) error
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
		chs:                  make(map[string]*channel),
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
	return s.id
}

func (s *session) AddContact(peer perun.Peer) error {
	s.Debugf("Received request: session.AddContact. Params %+v", peer)
	s.Lock()
	defer s.Unlock()

	err := s.contacts.Write(peer.Alias, peer)
	if err != nil {
		s.Error(err)
	}
	return perun.GetAPIError(err)
}

func (s *session) GetContact(alias string) (perun.Peer, error) {
	s.Debugf("Received request: session.GetContact. Params %+v", alias)
	s.Lock()
	defer s.Unlock()

	peer, isPresent := s.contacts.ReadByAlias(alias)
	if !isPresent {
		s.Error(perun.ErrUnknownAlias)
		return perun.Peer{}, perun.ErrUnknownAlias
	}
	return peer, nil
}

func (s *session) OpenCh(pctx context.Context, openingBalInfo perun.BalInfo, app perun.App, challengeDurSecs uint64) (
	perun.ChInfo, error) {
	s.Debugf("\nReceived request:session.OpenCh Params %+v,%+v,%+v", openingBalInfo, app, challengeDurSecs)
	s.Lock()
	defer s.Unlock()

	sanitizeBalInfo(openingBalInfo)
	parts, err := retrieveParts(openingBalInfo.Parts, s.contacts)
	if err != nil {
		s.Error(err, "retrieving channel participants using session contacts")
		return perun.ChInfo{}, perun.GetAPIError(err)
	}
	registerParts(parts, s.chClient)

	allocations, err := makeAllocation(openingBalInfo, s.chAsset)
	if err != nil {
		s.Error(err, "making allocations")
		return perun.ChInfo{}, perun.GetAPIError(err)
	}

	proposal := pclient.NewLedgerChannelProposal(
		challengeDurSecs,
		s.user.OffChainAddr,
		allocations,
		offChainAddrs(parts),
		pclient.WithApp(app.Def, app.Data),
		pclient.WithRandomNonce())
	ctx, cancel := context.WithTimeout(pctx, s.timeoutCfg.proposeCh(challengeDurSecs))
	defer cancel()
	pch, err := s.chClient.ProposeChannel(ctx, proposal)
	if err != nil {
		s.Error(err)
		// TODO: (mano) Use errors.Is here once a sentinel error value is defined in the SDK.
		if strings.Contains(err.Error(), "channel proposal rejected") {
			err = perun.ErrPeerRejected
		}
		return perun.ChInfo{}, perun.GetAPIError(err)
	}

	ch := newCh(pch, openingBalInfo.Currency, openingBalInfo.Parts, s.timeoutCfg, challengeDurSecs)
	s.addCh(ch)
	return ch.GetChInfo(), nil
}

// sanitizeBalInfo checks if the entry for ownAlias is at index 0,
// if not it rearranges the Aliases & Balance lists to make the index of ownAlias 0.
//
// BalanceInfo will be unchanged if there is no entry for ownAlias.
func sanitizeBalInfo(balInfo perun.BalInfo) {
	ownIdx := 0
	for idx := range balInfo.Parts {
		if balInfo.Parts[idx] == perun.OwnAlias {
			ownIdx = idx
		}
	}
	// Rearrange when ownAlias is not index 0.
	if ownIdx != 0 {
		balInfo.Parts[ownIdx] = balInfo.Parts[0]
		balInfo.Parts[0] = perun.OwnAlias

		ownAmount := balInfo.Bal[ownIdx]
		balInfo.Bal[ownIdx] = balInfo.Bal[0]
		balInfo.Bal[0] = ownAmount
	}
}

// retrieveParts retrieves the peers from corresponding to the aliases from the contacts provider.
// The order of entries for parts list will be same as that of aliases. i.e aliases[i] = parts[i].Alias.
func retrieveParts(aliases []string, contacts perun.ContactsReader) ([]perun.Peer, error) {
	knownParts := make(map[string]perun.Peer, len(aliases))
	parts := make([]perun.Peer, len(aliases))
	missingParts := make([]string, 0, len(aliases))
	repeatedParts := make([]string, 0, len(aliases))
	foundOwnAlias := false
	for idx, alias := range aliases {
		if alias == perun.OwnAlias {
			foundOwnAlias = true
		}
		peer, isPresent := contacts.ReadByAlias(alias)
		if !isPresent {
			missingParts = append(missingParts, alias)
			continue
		}
		if _, isPresent := knownParts[alias]; isPresent {
			repeatedParts = append(repeatedParts, alias)
		}
		knownParts[alias] = peer
		parts[idx] = peer
	}

	if len(missingParts) != 0 {
		return nil, errors.New(fmt.Sprintf("No peers found in contacts for the following alias(es): %v", knownParts))
	}
	if len(repeatedParts) != 0 {
		return nil, errors.New(fmt.Sprintf("Repeated entries in aliases: %v", knownParts))
	}
	if !foundOwnAlias {
		return nil, errors.New("No entry for self found in aliases")
	}

	return parts, nil
}

// registerParts will register the given parts to the passed registry.
func registerParts(parts []perun.Peer, r perun.Registerer) {
	for idx := range parts {
		if parts[idx].Alias != perun.OwnAlias { // Skip own alias.
			r.Register(parts[idx].OffChainAddr, parts[idx].CommAddr)
		}
	}
}

// offChainAddrs returns the list of off-chain addresses corresponding to the given list of peers.
func offChainAddrs(parts []perun.Peer) []pwallet.Address {
	addrs := make([]pwallet.Address, len(parts))
	for i := range parts {
		addrs[i] = parts[i].OffChainAddr
	}
	return addrs
}

// makeAllocation makes an allocation using the BalanceInfo and the chAsset.
// Order of amounts in the balance is same as the order of Aliases in the Balance Info.
// It errors if any of the amounts cannot be parsed using the interpreter corresponding to the currency.
func makeAllocation(balInfo perun.BalInfo, chAsset pchannel.Asset) (*pchannel.Allocation, error) {
	if !currency.IsSupported(balInfo.Currency) {
		return nil, perun.ErrUnsupportedCurrency
	}

	balance := make([]*big.Int, len(balInfo.Bal))
	var err error
	for i := range balInfo.Bal {
		balance[i], err = currency.NewParser(balInfo.Currency).Parse(balInfo.Bal[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "Parsing amount: %v", balInfo.Bal[i])
		}
	}

	return &pchannel.Allocation{
		Assets:   []pchannel.Asset{chAsset},
		Balances: [][]*big.Int{balance},
	}, nil
}

func (s *session) addCh(ch *channel) {
	// TODO: (mano) use logger with multiple fields and use session-id, channel-id.
	ch.Logger = log.NewLoggerWithField("channel-id", ch.id)
	s.chs[ch.id] = ch
	go func(s *session, chID string) {
		ch.Debug("Started channel watcher")
		err := ch.pch.Watch()
		s.HandleClose(chID, err)
	}(s, ch.id)
}

func (s *session) HandleProposal(chProposal pclient.ChannelProposal, responder *pclient.ProposalResponder) {
	s.Debugf("SDK Callback: HandleProposal. Params: %+v", chProposal)
	s.Lock()
	defer s.Unlock()
	expiry := time.Now().UTC().Add(s.timeoutCfg.response).Unix()

	parts := make([]string, len(chProposal.Proposal().PeerAddrs))
	for i := range chProposal.Proposal().PeerAddrs {
		p, ok := s.contacts.ReadByOffChainAddr(chProposal.Proposal().PeerAddrs[i])
		if !ok {
			s.Info("Received channel proposal from unknonwn peer", chProposal.Proposal().PeerAddrs[i].String())
			// nolint: errcheck, gosec		// It is sufficient to just log this error.
			s.rejectChProposal(context.Background(), responder, "peer not found in session contacts")
			expiry = 0
			break
		}
		parts[i] = p.Alias
	}

	notif := chProposalNotif(parts, currency.ETH, chProposal.Proposal(), expiry)
	entry := chProposalResponderEntry{
		notif:     notif,
		responder: responder,
	}
	// Need not store entries for notification with expiry = 0, as these update requests have
	// already been rejected by the perun node. Hence no response is expected for these notifications.
	if expiry != 0 {
		s.chProposalResponders[notif.ProposalID] = entry
	}

	// Set ETH as the currency interpreter for incoming channel.
	// TODO: (mano) Provide an option for user to configure when more currency interpretters are supported.
	if s.chProposalNotifier == nil {
		s.chProposalNotifsCache = append(s.chProposalNotifsCache, notif)
		s.Debugf("HandleProposal: Notification cached", notif)
	} else {
		go s.chProposalNotifier(notif)
		s.Debugf("HandleProposal: Notification sent", notif)
	}
}

func chProposalNotif(parts []string, curr string, chProposal *pclient.BaseChannelProposal,
	expiry int64) perun.ChProposalNotif {
	return perun.ChProposalNotif{
		ProposalID:       fmt.Sprintf("%x", chProposal.ProposalID()),
		OpeningBalInfo:   makeBalInfoFromRawBal(parts, curr, chProposal.InitBals.Balances[0]),
		App:              makeApp(chProposal.Proposal().App, chProposal.InitData),
		ChallengeDurSecs: chProposal.ChallengeDuration,
		Expiry:           expiry,
	}
}

func (s *session) SubChProposals(notifier perun.ChProposalNotifier) error {
	s.Debug("Received request: session.SubChProposals")
	s.Lock()
	defer s.Unlock()

	if s.chProposalNotifier != nil {
		return perun.ErrSubAlreadyExists
	}
	s.chProposalNotifier = notifier

	// Send all cached notifications.
	for i := len(s.chProposalNotifsCache); i > 0; i-- {
		go s.chProposalNotifier(s.chProposalNotifsCache[0])
		s.chProposalNotifsCache = s.chProposalNotifsCache[1:i]
	}
	return nil
}

func (s *session) UnsubChProposals() error {
	s.Debug("Received request: session.UnsubChProposals")
	s.Lock()
	defer s.Unlock()

	if s.chProposalNotifier == nil {
		return perun.ErrNoActiveSub
	}
	s.chProposalNotifier = nil
	return nil
}

func (s *session) RespondChProposal(pctx context.Context, chProposalID string, accept bool) error {
	s.Debugf("Received request: session.RespondChProposal. Params: %+v, %+v", chProposalID, accept)
	s.Lock()
	defer s.Unlock()

	entry, ok := s.chProposalResponders[chProposalID]
	if !ok {
		s.Info(perun.ErrUnknownProposalID)
		return perun.ErrUnknownProposalID
	}
	delete(s.chProposalResponders, chProposalID)

	currTime := time.Now().UTC().Unix()
	if entry.notif.Expiry < currTime {
		s.Info("timeout:", entry.notif.Expiry, "received response at:", currTime)
		return perun.ErrRespTimeoutExpired
	}

	switch accept {
	case true:
		if err := s.acceptChProposal(pctx, entry); err != nil {
			s.Error("Accepting channel proposal", err)
			return perun.GetAPIError(err)
		}
	case false:
		if err := s.rejectChProposal(pctx, entry.responder, "rejected by user"); err != nil {
			s.Error("Rejecting channel proposal", err)
			return perun.GetAPIError(err)
		}
	}
	return nil
}

func (s *session) acceptChProposal(pctx context.Context, entry chProposalResponderEntry) error {
	ctx, cancel := context.WithTimeout(pctx, s.timeoutCfg.respChProposalAccept(entry.notif.ChallengeDurSecs))
	defer cancel()
	pch, err := entry.responder.Accept(ctx, pclient.ProposalAcc{Participant: s.user.OffChainAddr})
	if err != nil {
		s.Error("Accepting channel proposal", err)
		return err
	}

	// Set ETH as the currency interpreter for incoming channel.
	// TODO: (mano) Provide an option for user to configure when more currency interpreters are supported.
	ch := newCh(pch, currency.ETH, entry.notif.OpeningBalInfo.Parts, s.timeoutCfg, entry.notif.ChallengeDurSecs)
	s.addCh(ch)
	return nil
}

func (s *session) rejectChProposal(pctx context.Context, responder chProposalResponder, reason string) error {
	ctx, cancel := context.WithTimeout(pctx, s.timeoutCfg.respChProposalReject())
	defer cancel()
	err := responder.Reject(ctx, reason)
	if err != nil {
		s.Error("Rejecting channel proposal from unknown peer", err)
	}
	return err
}

func (s *session) GetChsInfo() []perun.ChInfo {
	s.Debug("Received request: session.GetChInfos")
	s.Lock()
	defer s.Unlock()

	openChsInfo := make([]perun.ChInfo, len(s.chs))
	i := 0
	for _, ch := range s.chs {
		openChsInfo[i] = ch.GetChInfo()
		i++
	}
	return openChsInfo
}

func (s *session) GetCh(chID string) (perun.ChAPI, error) {
	s.Debugf("Internal call to get channel instance. Params: %+v", chID)
	s.Lock()
	defer s.Unlock()

	ch, ok := s.chs[chID]
	if !ok {
		s.Info(perun.ErrUnknownChID)
		return nil, perun.ErrUnknownChID
	}
	return ch, nil
}

func (s *session) HandleUpdate(chUpdate pclient.ChannelUpdate, responder *pclient.UpdateResponder) {
	s.Debugf("SDK Callback: HandleUpdate. Params: %+v", chUpdate)
	s.Lock()
	defer s.Unlock()
	expiry := time.Now().UTC().Add(s.timeoutCfg.response).Unix()

	chID := fmt.Sprintf("%x", chUpdate.State.ID)
	updateID := fmt.Sprintf("%s_%d", chID, chUpdate.State.Version)

	ch, ok := s.chs[chID]
	if !ok {
		s.Info("Received update for unknown channel", chID)
		err := responder.Reject(context.Background(), "unknown channel for this session")
		s.Info("Error rejecting unknown channel with id %s: %v", chID, err)
		return
	}

	ch.Lock()
	defer ch.Unlock()
	if chUpdate.State.IsFinal {
		ch.Info("Received final update from peer, channel is finalized.")
		ch.lockState = finalized
	}

	entry := chUpdateResponderEntry{
		responder:   responder,
		notifExpiry: expiry,
	}
	// Need not store entries for notification with expiry = 0, as these update requests have
	// already been rejected by the perun node. Hence no response is expected for these notifications.
	if expiry != 0 {
		ch.chUpdateResponders[updateID] = entry
	}
	notif := chUpdateNotif(ch.getChInfo(), chUpdate.State, expiry)
	if ch.chUpdateNotifier == nil {
		ch.chUpdateNotifCache = append(ch.chUpdateNotifCache, notif)
		ch.Debug("HandleUpdate: Notification cached")
	} else {
		go ch.chUpdateNotifier(notif)
		ch.Debug("HandleUpdate: Notification sent")
	}
}

func chUpdateNotif(currChInfo perun.ChInfo, proposedState *pchannel.State, expiry int64) perun.ChUpdateNotif {
	return perun.ChUpdateNotif{
		UpdateID:       fmt.Sprintf("%s_%d", currChInfo.ChID, proposedState.Version),
		CurrChInfo:     currChInfo,
		ProposedChInfo: makeChInfo(currChInfo.ChID, currChInfo.BalInfo.Parts, currChInfo.BalInfo.Currency, proposedState),
		Expiry:         expiry,
	}
}

func (s *session) Close(force bool) error {
	return nil
}

func (s *session) HandleClose(chID string, err error) {
	s.Debug("SDK Callback: Channel watcher returned.")

	ch := s.chs[chID]
	ch.Lock()
	defer ch.Unlock()

	if ch.lockState == open || ch.lockState == finalized {
		ch.lockState = closed
	}

	notif := perun.ChCloseNotif{
		ClosedChInfo: ch.getChInfo(),
	}
	if err != nil {
		notif.Error = err.Error()
	}

	if s.chCloseNotifier == nil {
		s.chCloseNotifsCache = append(s.chCloseNotifsCache, notif)
		s.Debug("HandleClose: Notification cached")
	} else {
		go s.chCloseNotifier(notif)
		s.Debug("HandleClose: Notification sent")
	}
}

func (s *session) SubChCloses(notifier perun.ChCloseNotifier) error {
	s.Debug("Received request: session.SubChCloses")
	s.Lock()
	defer s.Unlock()

	if s.chCloseNotifier != nil {
		return perun.ErrSubAlreadyExists
	}
	s.chCloseNotifier = notifier

	// Send all cached notifications
	for i := len(s.chCloseNotifsCache); i > 0; i-- {
		go s.chCloseNotifier(s.chCloseNotifsCache[0])
		s.chCloseNotifsCache = s.chCloseNotifsCache[1:i]

	}
	return nil
}

func (s *session) UnsubChCloses() error {
	s.Debug("Received request: session.UnsubChCloses")
	s.Lock()
	defer s.Unlock()

	if s.chCloseNotifier == nil {
		return perun.ErrNoActiveSub
	}
	s.chCloseNotifier = nil
	return nil
}
