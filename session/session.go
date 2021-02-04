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
	pwire "perun.network/go-perun/wire"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/client"
	"github.com/hyperledger-labs/perun-node/comm/tcp"
	"github.com/hyperledger-labs/perun-node/comm/tcp/tcptest"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/idprovider"
	"github.com/hyperledger-labs/perun-node/idprovider/local"
	"github.com/hyperledger-labs/perun-node/log"
)

// walletBackend for initializing user wallets and parsing off-chain addresses
// in incoming peer IDs. A package level unexported variable is used so that a
// test wallet backend can be set using a function defined in export_test.go.
// Because real backend have large unlocking times and hence tests take very long.
var walletBackend perun.WalletBackend

func init() {
	// This can be overridden (only) in tests by calling the SetWalletBackend function.
	walletBackend = ethereum.NewWalletBackend()
}

// Error type is used to define error constants for this package.
type Error string

// Error implements error interface.
func (e Error) Error() string {
	return string(e)
}

// Definition of error constants for this package.
const (
	ErrSessionClosed Error = "operation not allowed on a closed session"
)

type (
	// Session implements perun.SessionAPI.
	Session struct {
		log.Logger
		psync.Mutex

		id         string
		isOpen     bool
		timeoutCfg timeoutConfig
		user       perun.User
		chAsset    pchannel.Asset
		chClient   perun.ChClient
		idProvider perun.IDProvider

		chs map[string]*Channel

		chProposalNotifier    perun.ChProposalNotifier
		chProposalNotifsCache []perun.ChProposalNotif
		chProposalResponders  map[string]chProposalResponderEntry
	}

	chProposalResponderEntry struct {
		proposal  pclient.ChannelProposal
		notif     perun.ChProposalNotif
		responder ChProposalResponder
	}

	// ChProposalResponder defines the methods on proposal responder that will be used by the perun node.
	ChProposalResponder interface {
		Accept(context.Context, *pclient.ChannelProposalAcc) (perun.Channel, error)
		Reject(ctx context.Context, reason string) error
	}
)

//go:generate mockery --name ChProposalResponder --output ../internal/mocks

// chProposalResponderWrapped is a wrapper around pclient.ProposalResponder that returns a channel of
// interface type instead of struct type. This enables easier mocking of the returned value in tests.
type chProposalResponderWrapped struct {
	*pclient.ProposalResponder
}

// Accept is a wrapper around the original function, that returns a channel of interface type instead of struct type.
func (r *chProposalResponderWrapped) Accept(ctx context.Context, proposalAcc *pclient.ChannelProposalAcc) (
	perun.Channel, error) {
	return r.ProposalResponder.Accept(ctx, proposalAcc)
}

// New initializes a SessionAPI instance for the given configuration and returns an
// instance of it. All methods on it are safe for concurrent use.
func New(cfg Config) (*Session, error) {
	user, err := NewUnlockedUser(walletBackend, cfg.User)
	if err != nil {
		return nil, err
	}

	if cfg.User.CommType != "tcp" {
		return nil, perun.ErrUnsupportedCommType
	}
	commBackend := tcp.NewTCPBackend(tcptest.DialerTimeout)

	chAsset, err := walletBackend.ParseAddr(cfg.Asset)
	if err != nil {
		return nil, err
	}

	idProvider, err := initIDProvider(cfg.IDProviderType, cfg.IDProviderURL, walletBackend, user.PeerID)
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
	sess := &Session{
		Logger:               log.NewLoggerWithField("session-id", sessionID),
		id:                   sessionID,
		isOpen:               true,
		timeoutCfg:           timeoutCfg,
		user:                 user,
		chAsset:              chAsset,
		chClient:             chClient,
		idProvider:           idProvider,
		chs:                  make(map[string]*Channel),
		chProposalResponders: make(map[string]chProposalResponderEntry),
	}
	err = sess.chClient.RestoreChs(sess.handleRestoredCh)
	if err != nil {
		return nil, errors.WithMessage(err, "restoring channels")
	}
	chClient.Handle(sess, sess) // Init handlers
	return sess, nil
}

func initIDProvider(idProviderType, idProviderURL string, wb perun.WalletBackend, own perun.PeerID) (
	perun.IDProvider, error) {
	if idProviderType != "local" {
		return nil, perun.ErrUnsupportedIDProviderType
	}
	idProvider, err := local.NewIDprovider(idProviderURL, wb)
	if err != nil {
		return nil, err
	}

	own.Alias = perun.OwnAlias
	err = idProvider.Write(perun.OwnAlias, own)
	if err != nil && !errors.Is(err, perun.ErrPeerExists) {
		return nil, errors.Wrap(err, "registering own user in ID Provider")
	}
	return idProvider, nil
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

// ID implements sessionAPI.ID.
func (s *Session) ID() string {
	return s.id
}

func (s *Session) handleRestoredCh(pch perun.Channel) {
	s.Debugf("found channel in persistence: 0x%x", pch.ID())

	// Restore only those channels that are in acting phase.
	if pch.Phase() != pchannel.Acting {
		return
	}
	partOffChainAddrs := pch.Peers()
	partIDs := make([]perun.PeerID, len(partOffChainAddrs))
	aliases := make([]string, len(partOffChainAddrs))
	for i := range pch.Peers() {
		p, ok := s.idProvider.ReadByOffChainAddr(partOffChainAddrs[i])
		if !ok {
			s.Info("Unknown peer address in a persisted channel, will not be restored", pch.Peers()[i].String())
			return
		}
		partIDs[i] = p
		aliases[i] = p.Alias
	}

	registerParts(partIDs, s.chClient)

	ch := newCh(pch, currency.ETH, aliases, s.timeoutCfg, pch.Params().ChallengeDuration)
	s.addCh(ch)
	s.Debugf("restored channel from persistence: %v", ch.getChInfo())
}

// AddPeerID implements sessionAPI.AddPeerID.
func (s *Session) AddPeerID(peerID perun.PeerID) perun.APIErrorV2 {
	s.WithFields(log.Fields{"method": "AddPeerID"}).Info("Received request with params:", peerID)
	s.Lock()
	defer s.Unlock()

	if !s.isOpen {
		err := ErrSessionClosed
		s.WithFields(log.Fields{"method": "AddPeerID"}).Error(err)
		return perun.NewAPIErrV2FailedPreCondition(err.Error())
	}

	err := s.idProvider.Write(peerID.Alias, peerID)
	if err != nil {
		var apiErr perun.APIErrorV2
		switch {
		case errors.Is(err, idprovider.ErrPeerAliasAlreadyUsed):
			requirement := "peer alias should be unique for each peer ID"
			apiErr = perun.NewAPIErrV2InvalidArgument("peer alias", peerID.Alias, requirement, err.Error())
		case errors.Is(err, idprovider.ErrPeerIDAlreadyRegistered):
			apiErr = perun.NewAPIErrV2ResourceExists("peer alias", peerID.Alias, err.Error())
		case errors.Is(err, idprovider.ErrParsingOffChainAddress):
			apiErr = perun.NewAPIErrV2InvalidArgument("off-chain address string", peerID.OffChainAddrString, "", err.Error())
		default:
			apiErr = perun.NewAPIErrV2UnknownInternal(err)
		}
		s.WithFields(perun.APIErrV2AsMap(apiErr)).Error(apiErr.Message())
		return apiErr
	}
	s.Info("Peer ID successfully added")
	return nil
}

// GetPeerID implements sessionAPI.GetPeerID.
func (s *Session) GetPeerID(alias string) (perun.PeerID, error) {
	s.Debugf("Received request: session.GetPeerID. Params %+v", alias)
	s.Lock()
	defer s.Unlock()

	if !s.isOpen {
		return perun.PeerID{}, perun.ErrSessionClosed
	}

	peerID, isPresent := s.idProvider.ReadByAlias(alias)
	if !isPresent {
		s.Error(perun.ErrUnknownAlias)
		return perun.PeerID{}, perun.ErrUnknownAlias
	}
	return peerID, nil
}

// OpenCh implements sessionAPI.OpenCh.
func (s *Session) OpenCh(pctx context.Context, openingBalInfo perun.BalInfo, app perun.App, challengeDurSecs uint64) (
	perun.ChInfo, error) {
	s.Debugf("\nReceived request:session.OpenCh Params %+v,%+v,%+v", openingBalInfo, app, challengeDurSecs)
	// Session is locked only when adding the channel to session.

	if !s.isOpen {
		return perun.ChInfo{}, perun.ErrSessionClosed
	}

	sanitizeBalInfo(openingBalInfo)
	parts, err := retrievePartIDs(openingBalInfo.Parts, s.idProvider)
	if err != nil {
		s.Error(err, "retrieving channel participant IDs using session ID Provider")
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
		makeOffChainAddrs(parts),
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

// retrievePartIDs retrieves the peer IDs corresponding to the aliases from the ID provider.
// The order of entries for parts list will be same as that of aliases. i.e aliases[i] = parts[i].Alias.
func retrievePartIDs(aliases []string, idProvider perun.IDReader) ([]perun.PeerID, error) {
	knownParts := make(map[string]perun.PeerID, len(aliases))
	partIDs := make([]perun.PeerID, len(aliases))
	missingParts := make([]string, 0, len(aliases))
	repeatedParts := make([]string, 0, len(aliases))
	foundOwnAlias := false
	for idx, alias := range aliases {
		if alias == perun.OwnAlias {
			foundOwnAlias = true
		}
		peerID, isPresent := idProvider.ReadByAlias(alias)
		if !isPresent {
			missingParts = append(missingParts, alias)
			continue
		}
		if _, isPresent := knownParts[alias]; isPresent {
			repeatedParts = append(repeatedParts, alias)
		}
		knownParts[alias] = peerID
		partIDs[idx] = peerID
	}

	if len(missingParts) != 0 {
		return nil, errors.New(fmt.Sprintf("No peer IDs found in ID Provider for the following alias(es): %v", missingParts))
	}
	if len(repeatedParts) != 0 {
		return nil, errors.New(fmt.Sprintf("Repeated entries in aliases: %v", repeatedParts))
	}
	if !foundOwnAlias {
		return nil, errors.New("No entry for self found in aliases")
	}

	return partIDs, nil
}

// registerParts will register the given parts to the passed registry.
func registerParts(parts []perun.PeerID, r perun.Registerer) {
	for idx := range parts {
		if parts[idx].Alias != perun.OwnAlias { // Skip own alias.
			r.Register(parts[idx].OffChainAddr, parts[idx].CommAddr)
		}
	}
}

// makeOffChainAddrs returns the list of off-chain addresses corresponding to the given list of peer IDs.
func makeOffChainAddrs(partIDs []perun.PeerID) []pwire.Address {
	addrs := make([]pwire.Address, len(partIDs))
	for i := range partIDs {
		addrs[i] = partIDs[i].OffChainAddr
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

// addCh adds the channel to session. It locks the session mutex during the operation.
func (s *Session) addCh(ch *Channel) {
	ch.Logger = log.NewDerivedLoggerWithField(s.Logger, "channel-id", ch.id)
	s.Lock()
	s.chs[ch.id] = ch
	s.Unlock()
}

// HandleProposal is a handler to be registered on the channel client for processing incoming channel proposals.
func (s *Session) HandleProposal(chProposal pclient.ChannelProposal, responder *pclient.ProposalResponder) {
	s.HandleProposalWInterface(chProposal, &chProposalResponderWrapped{responder})
}

// HandleProposalWInterface is the actual implemention of HandleProposal that takes arguments as interface types.
// It is implemented this way to enable easier testing.
func (s *Session) HandleProposalWInterface(chProposal pclient.ChannelProposal, responder ChProposalResponder) {
	s.Debugf("SDK Callback: HandleProposal. Params: %+v", chProposal)
	expiry := time.Now().UTC().Add(s.timeoutCfg.response).Unix()

	if !s.isOpen {
		// Code will not reach here during runtime as chClient is closed when closing a session.
		s.Error("Unexpected HandleProposal callback invoked on a closed session")
		return
	}

	parts := make([]string, len(chProposal.Base().PeerAddrs))
	for i := range chProposal.Base().PeerAddrs {
		p, ok := s.idProvider.ReadByOffChainAddr(chProposal.Base().PeerAddrs[i])
		if !ok {
			s.Info("Received channel proposal from unknonwn peer ID", chProposal.Base().PeerAddrs[i].String())
			// nolint: errcheck, gosec		// It is sufficient to just log this error.
			s.rejectChProposal(context.Background(), responder, "peer ID not found in session ID Provider")
			expiry = 0
			break
		}
		parts[i] = p.Alias
	}

	notif := chProposalNotif(parts, currency.ETH, chProposal.Base(), expiry)
	entry := chProposalResponderEntry{
		proposal:  chProposal,
		notif:     notif,
		responder: responder,
	}

	s.Lock()
	defer s.Unlock()
	// Need not store entries for notification with expiry = 0, as these update requests have
	// already been rejected by the perun node. Hence no response is expected for these notifications.
	if expiry != 0 {
		s.chProposalResponders[notif.ProposalID] = entry
	}

	// Set ETH as the currency interpreter for incoming channel.
	// TODO: (mano) Provide an option for user to configure when more currency interpretters are supported.
	if s.chProposalNotifier == nil {
		s.chProposalNotifsCache = append(s.chProposalNotifsCache, notif)
		s.Debug("HandleProposal: Notification cached", notif)
	} else {
		go s.chProposalNotifier(notif)
		s.Debug("HandleProposal: Notification sent", notif)
	}
}

func chProposalNotif(parts []string, curr string, chProposal *pclient.BaseChannelProposal,
	expiry int64) perun.ChProposalNotif {
	return perun.ChProposalNotif{
		ProposalID:       fmt.Sprintf("%x", chProposal.ProposalID()),
		OpeningBalInfo:   makeBalInfoFromRawBal(parts, curr, chProposal.InitBals.Balances[0]),
		App:              makeApp(chProposal.Base().App, chProposal.InitData),
		ChallengeDurSecs: chProposal.ChallengeDuration,
		Expiry:           expiry,
	}
}

// SubChProposals implements sessionAPI.SubChProposals.
func (s *Session) SubChProposals(notifier perun.ChProposalNotifier) error {
	s.Debug("Received request: session.SubChProposals")
	s.Lock()
	defer s.Unlock()

	if !s.isOpen {
		return perun.ErrSessionClosed
	}

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

// UnsubChProposals implements sessionAPI.UnsubChProposals.
func (s *Session) UnsubChProposals() error {
	s.Debug("Received request: session.UnsubChProposals")
	s.Lock()
	defer s.Unlock()

	if !s.isOpen {
		return perun.ErrSessionClosed
	}

	if s.chProposalNotifier == nil {
		return perun.ErrNoActiveSub
	}
	s.chProposalNotifier = nil
	return nil
}

// RespondChProposal implements sessionAPI.RespondChProposal.
func (s *Session) RespondChProposal(pctx context.Context, chProposalID string, accept bool) (perun.ChInfo, error) {
	s.Debugf("Received request: session.RespondChProposal. Params: %+v, %+v", chProposalID, accept)

	if !s.isOpen {
		return perun.ChInfo{}, perun.ErrSessionClosed
	}

	// Lock the session mutex only when retrieving the channel responder and deleting it.
	// It will again be locked when adding the channel to the session.
	s.Lock()
	entry, ok := s.chProposalResponders[chProposalID]
	if !ok {
		s.Info(perun.ErrUnknownProposalID)
		s.Unlock()
		return perun.ChInfo{}, perun.ErrUnknownProposalID
	}
	delete(s.chProposalResponders, chProposalID)
	s.Unlock()

	currTime := time.Now().UTC().Unix()
	if entry.notif.Expiry < currTime {
		s.Info("timeout:", entry.notif.Expiry, "received response at:", currTime)
		return perun.ChInfo{}, perun.ErrRespTimeoutExpired
	}

	var openedChInfo perun.ChInfo
	var err error
	switch accept {
	case true:
		openedChInfo, err = s.acceptChProposal(pctx, entry)
	case false:
		err = s.rejectChProposal(pctx, entry.responder, "rejected by user")
	}
	return openedChInfo, perun.GetAPIError(err)
}

func (s *Session) acceptChProposal(pctx context.Context, entry chProposalResponderEntry) (perun.ChInfo, error) {
	ctx, cancel := context.WithTimeout(pctx, s.timeoutCfg.respChProposalAccept(entry.notif.ChallengeDurSecs))
	defer cancel()

	proposal := entry.proposal.Base()
	resp := proposal.NewChannelProposalAcc(s.user.OffChainAddr, pclient.WithRandomNonce())
	pch, err := entry.responder.Accept(ctx, resp)
	if err != nil {
		s.Error("Accepting channel proposal", err)
		return perun.ChInfo{}, err
	}

	// Set ETH as the currency interpreter for incoming channel.
	// TODO: (mano) Provide an option for user to configure when more currency interpreters are supported.
	ch := newCh(pch, currency.ETH, entry.notif.OpeningBalInfo.Parts, s.timeoutCfg, entry.notif.ChallengeDurSecs)
	s.addCh(ch)
	return ch.getChInfo(), nil
}

func (s *Session) rejectChProposal(pctx context.Context, responder ChProposalResponder, reason string) error {
	ctx, cancel := context.WithTimeout(pctx, s.timeoutCfg.respChProposalReject())
	defer cancel()
	err := responder.Reject(ctx, reason)
	if err != nil {
		s.Error("Rejecting channel proposal", err)
	}
	return err
}

// GetChsInfo implements sessionAPI.GetChsInfo.
func (s *Session) GetChsInfo() []perun.ChInfo {
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

// GetCh implements sessionAPI.GetCh.
func (s *Session) GetCh(chID string) (perun.ChAPI, error) {
	s.Debugf("Internal call to get channel instance. Params: %+v", chID)

	s.Lock()
	ch, ok := s.chs[chID]
	s.Unlock()
	if !ok {
		s.Info(perun.ErrUnknownChID, "not found in session")
		return nil, perun.ErrUnknownChID
	}

	return ch, nil
}

// HandleUpdate is a handler to be registered on the channel client for processing incoming channel updates.
// This function just identifies the channel to which update is received and invokes the handler for that
// channel.
func (s *Session) HandleUpdate(chUpdate pclient.ChannelUpdate, responder *pclient.UpdateResponder) {
	s.HandleUpdateWInterface(chUpdate, responder)
}

// HandleUpdateWInterface is the actual implemention of HandleUpdate that takes arguments as interface types.
// It is implemented this way to enable easier testing.
func (s *Session) HandleUpdateWInterface(chUpdate pclient.ChannelUpdate, responder ChUpdateResponder) {
	s.Debugf("SDK Callback: HandleUpdate. Params: %+v", chUpdate)
	s.Lock()
	defer s.Unlock()

	if !s.isOpen {
		// Code will not reach here during runtime as chClient is closed when closing a session.
		s.Error("Unexpected HandleUpdate callback invoked on a closed session")
		return
	}

	chID := fmt.Sprintf("%x", chUpdate.State.ID)
	ch, ok := s.chs[chID]
	if !ok {
		s.Info("Received update for unknown channel", chID)
		err := responder.Reject(context.Background(), "unknown channel for this session")
		s.Info("Error rejecting incoming update for unknown channel with id %s: %v", chID, err)
		return
	}
	go ch.HandleUpdate(chUpdate, responder)
}

// Close implements sessionAPI.Close.
func (s *Session) Close(force bool) ([]perun.ChInfo, error) {
	s.Debug("Received request: session.Close")
	s.Lock()
	defer s.Unlock()

	if !s.isOpen {
		return nil, perun.ErrSessionClosed
	}

	openChsInfo := []perun.ChInfo{}
	unexpectedPhaseChIDs := []string{}

	for _, ch := range s.chs {
		// Acquire channel mutex to ensure any ongoing operation on the channel is finished.
		ch.Lock()

		// Calling Phase() also waits for the mutex on pchannel that ensures any handling of Registered event
		// in the Watch routine is also completed. But if the event was received after acquiring channel mutex
		// and completed before pc.Phase() returned, this event will not yet be serviced by perun-node.
		// A solution to this is to add a provision (that is currenlt missing) to suspend the Watcher (only
		// for open channels) before acquiring channel mutex and restoring it later if force option is false.
		//
		// TODO (mano): Add a provision in go-perun to suspend the watcher and it use it here.
		//
		// Since there will be no ongoing operations in perun-node, the pchannel is should be in one of the two
		// stable phases knonwn to perun node (see state diagram in the docs for details) : Acting or Withdrawn.
		phase := ch.pch.Phase()
		if phase != pchannel.Acting && phase != pchannel.Withdrawn {
			unexpectedPhaseChIDs = append(unexpectedPhaseChIDs, ch.ID())
		}
		if ch.status == open {
			openChsInfo = append(openChsInfo, ch.getChInfo())
		}
	}
	if len(unexpectedPhaseChIDs) != 0 {
		err := fmt.Errorf("chs in unexpected phase during session close: %v", unexpectedPhaseChIDs)
		s.Error(err.Error())
		s.unlockAllChs()
		return nil, perun.GetAPIError(errors.WithStack(err))
	}
	if !force && len(openChsInfo) != 0 {
		err := fmt.Errorf("%w: %v", perun.ErrOpenCh, openChsInfo)
		s.Error(err.Error())
		s.unlockAllChs()
		return openChsInfo, perun.GetAPIError(errors.WithStack(err))
	}

	s.isOpen = false
	return openChsInfo, s.close()
}

func (s *Session) unlockAllChs() {
	for _, ch := range s.chs {
		ch.Unlock()
	}
}

func (s *Session) close() error {
	s.user.OnChain.Wallet.LockAll()
	s.user.OffChain.Wallet.LockAll()
	return errors.WithMessage(s.chClient.Close(), "closing session")
}
