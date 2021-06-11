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
	"github.com/hyperledger-labs/perun-node/comm/tcp"
	"github.com/hyperledger-labs/perun-node/comm/tcp/tcptest"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/idprovider"
	"github.com/hyperledger-labs/perun-node/idprovider/local"
	"github.com/hyperledger-labs/perun-node/log"
)

// initialChRegistrySize is the initial size of channel registry. The
// registry size will automatically be increased when more channels are added.
const initialChRegistrySize = 10

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
	// For failed pre-condition.
	ErrChClosed      Error = "action not allowed on a closed channel"
	ErrSessionClosed Error = "action not allowed on a closed session"

	// For invalid config.
	ErrUnsupportedType      Error = "type not supported, see node config for supported types"
	ErrRepeatedPeerAlias    Error = "found repeated entries but each value should be unique"
	ErrEntryForSelfNotFound Error = "own peer alias (self) not found"
)

// Enumeration of valid resource types for used in ResourceNotFound and
// ResourceExists errors.
const (
	ResTypeUpdate      perun.ResourceType = "update"
	ResTypeUpdateSub   perun.ResourceType = "updatesSub"
	ResTypeChannel     perun.ResourceType = "channel"
	ResTypeProposal    perun.ResourceType = "proposal"
	ResTypeProposalSub perun.ResourceType = "proposalsSub"
	ResTypePeerID      perun.ResourceType = "peerID"
	ResTypeSession     perun.ResourceType = "session"
	ResTypeCurrency    perun.ResourceType = "currency"
)

// Enumeration of valid argument names for using in InvalidArgument error.
const (
	ArgNameAmount       perun.ArgumentName = "amount"
	ArgNameCurrency     perun.ArgumentName = "currency"
	ArgNamePeerAlias    perun.ArgumentName = "peerAlias"
	ArgNameOffChainAddr perun.ArgumentName = "offChainAddress"
	ArgNameConfigFile   perun.ArgumentName = "configFile"
	ArgNamePayee        perun.ArgumentName = "payee"
)

type (
	// Session provides a context for the user to interact with a node. It manages
	// user data (such as keys, peer IDs), and channel client.
	//
	// It implements the perun.SessionAPI interface. Once established, a user can
	// establish and transact on state channels.
	Session struct {
		log.Logger
		psync.Mutex

		id         string
		isOpen     bool
		user       User
		chAsset    pchannel.Asset
		chClient   ChClient
		idProvider perun.IDProvider

		timeoutCfg timeoutConfig
		chainURL   string // chain URL is stored for retrieval when annotating errors

		chs *chRegistry

		chProposalNotifier    perun.ChProposalNotifier
		chProposalNotifsCache []perun.ChProposalNotif
		chProposalResponders  map[string]chProposalResponderEntry
	}

	chProposalResponderEntry struct {
		proposal  pclient.LedgerChannelProposal
		notif     perun.ChProposalNotif
		responder ChProposalResponder
	}

	// ChProposalResponder defines the methods on proposal responder that will be used by the perun node.
	ChProposalResponder interface {
		Accept(context.Context, *pclient.LedgerChannelProposalAcc) (PChannel, error)
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
func (r *chProposalResponderWrapped) Accept(ctx context.Context, proposalAcc *pclient.LedgerChannelProposalAcc) (
	PChannel, error) {
	return r.ProposalResponder.Accept(ctx, proposalAcc)
}

// New initializes a SessionAPI instance for the given configuration and returns an
// instance of it. All methods on it are safe for concurrent use.
func New(cfg Config) (*Session, perun.APIError) {
	user, apiErr := NewUnlockedUser(walletBackend, cfg.User)
	if apiErr != nil {
		return nil, apiErr
	}

	if cfg.User.CommType != "tcp" {
		return nil, perun.NewAPIErrInvalidConfig(ErrUnsupportedType, "commType", cfg.User.CommType)
	}
	commBackend := tcp.NewTCPBackend(tcptest.DialerTimeout)
	chAsset, err := walletBackend.ParseAddr(cfg.Asset)
	if err != nil {
		return nil, perun.NewAPIErrInvalidConfig(err, "asset", cfg.Asset)
	}
	idProvider, apiErr := initIDProvider(cfg.IDProviderType, cfg.IDProviderURL, walletBackend, user.PeerID)
	if apiErr != nil {
		return nil, apiErr
	}

	chClientCfg := clientConfig{
		Chain: chainConfig{
			Adjudicator:      cfg.Adjudicator,
			Asset:            cfg.Asset,
			URL:              cfg.ChainURL,
			ChainID:          cfg.ChainID,
			ConnTimeout:      cfg.ChainConnTimeout,
			OnChainTxTimeout: cfg.OnChainTxTimeout,
		},
		DatabaseDir:       cfg.DatabaseDir,
		PeerReconnTimeout: cfg.PeerReconnTimeout,
	}
	chClient, apiErr := newEthereumPaymentClient(chClientCfg, user, commBackend)
	if apiErr != nil {
		return nil, apiErr
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
		chainURL:             cfg.ChainURL,
		timeoutCfg:           timeoutCfg,
		user:                 user,
		chAsset:              chAsset,
		chClient:             chClient,
		idProvider:           idProvider,
		chs:                  newChRegistry(initialChRegistrySize),
		chProposalResponders: make(map[string]chProposalResponderEntry),
	}
	err = sess.chClient.RestoreChs(sess.handleRestoredCh)
	if err != nil {
		err = errors.WithMessage(err, "restoring channels")
		return nil, perun.NewAPIErrInvalidConfig(err, "databaseDir", cfg.DatabaseDir)
	}
	chClient.Handle(sess, sess) // Init handlers
	return sess, nil
}

func initIDProvider(idProviderType, idProviderURL string, wb perun.WalletBackend, own perun.PeerID) (
	perun.IDProvider, perun.APIError) {
	if idProviderType != "local" {
		return nil, perun.NewAPIErrInvalidConfig(ErrUnsupportedType, "idProviderType", idProviderType)
	}
	idProvider, err := local.NewIDprovider(idProviderURL, wb)
	if err != nil {
		return nil, perun.NewAPIErrInvalidConfig(err, "idProviderURL", idProviderURL)
	}

	own.Alias = perun.OwnAlias
	err = idProvider.Write(perun.OwnAlias, own)
	if err != nil {
		err = errors.Wrap(err, "registering own user in ID Provider")
		return nil, perun.NewAPIErrInvalidConfig(err, "idProviderURL", idProviderURL)
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
	_, _ = h.Write(userOffChainAddr)
	_, _ = h.Write([]byte(time.Now().UTC().String()))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ID implements sessionAPI.ID.
func (s *Session) ID() string {
	return s.id
}

func (s *Session) handleRestoredCh(pch PChannel) {
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

	ch := newCh(pch, s.chainURL, currency.ETH, aliases, s.timeoutCfg, pch.Params().ChallengeDuration)
	s.addCh(ch)
	s.Debugf("restored channel from persistence: %v", ch.getChInfo())
}

// AddPeerID adds the peer ID to the ID provider instance of the session.
//
// If there is errors, it will be one of the following codes:
// - ErrResourceExists with ResourceType: "peerID" when peer ID is already registered
// - ErrInvalidArgument with Name:"peerAlias" when peer alias is used for another peer,
// - ErrInvalidArgument with Name:"offChainAddress" when off-chain address is invalid.
// - ErrUnknownInternal
func (s *Session) AddPeerID(peerID perun.PeerID) perun.APIError {
	s.WithField("method", "AddPeerID").Info("Received request with params:", peerID)
	s.Lock()
	defer s.Unlock()

	var apiErr perun.APIError
	if !s.isOpen {
		apiErr = perun.NewAPIErrFailedPreCondition(ErrSessionClosed)
		s.WithFields(perun.APIErrAsMap("AddPeerID", apiErr)).Error(apiErr.Message())
		return apiErr
	}

	err := s.idProvider.Write(peerID.Alias, peerID)
	if err != nil {
		// The error should be one of these following errors.
		switch {
		case errors.Is(err, idprovider.ErrPeerAliasAlreadyUsed):
			apiErr = perun.NewAPIErrInvalidArgument(err, ArgNamePeerAlias, peerID.Alias)
		case errors.Is(err, idprovider.ErrPeerIDAlreadyRegistered):
			apiErr = perun.NewAPIErrResourceExists(ResTypePeerID, peerID.Alias)
		case errors.Is(err, idprovider.ErrParsingOffChainAddress):
			apiErr = perun.NewAPIErrInvalidArgument(err, ArgNameOffChainAddr, peerID.OffChainAddrString)
		default:
			apiErr = perun.NewAPIErrUnknownInternal(err)
		}
		s.WithFields(perun.APIErrAsMap("AddPeerID", apiErr)).Error(apiErr.Message())
		return apiErr
	}
	s.WithField("method", "AddPeerID").Info("Peer ID added successfully")
	return nil
}

// GetPeerID gets the peer ID for the given alias from the ID provider instance
// of the session.
//
// If there is errors, it will be one of the following codes:
// - ErrResourceNotFound with ResourceType: "peerID" when peer alias is not known,
func (s *Session) GetPeerID(alias string) (perun.PeerID, perun.APIError) {
	s.WithField("method", "GetPeerID").Info("Received request with params:", alias)
	s.Lock()
	defer s.Unlock()

	if !s.isOpen {
		apiErr := perun.NewAPIErrFailedPreCondition(ErrSessionClosed)
		s.WithFields(perun.APIErrAsMap("GetPeerID", apiErr)).Error(apiErr.Message())
		return perun.PeerID{}, apiErr
	}

	peerID, isPresent := s.idProvider.ReadByAlias(alias)
	if !isPresent {
		apiErr := perun.NewAPIErrResourceNotFound(ResTypePeerID, alias)
		s.WithFields(perun.APIErrAsMap("GetPeerID", apiErr)).Error(apiErr.Message())
		return perun.PeerID{}, apiErr
	}
	s.WithField("method", "GetPeerID").Info("Peer ID retreived successfully")
	return peerID, nil
}

// OpenCh proposes a channel to the participants with the specified opening
// balance and app, funds it on the blockchain when the proposal is accepted
// and sets it up for off-chain transactions when all the participants have
// funded the channel on the blockchain.
//
// `Challenge duration` is the time available for the node to refute in case of
// disputes when a state is registered on the blockchain.
//
// If there is errors, it will be one of the following codes:
// - ErrResourceNotFound with ResourceType: "peerID" when any of the peer aliases are not known.
// - ErrResourceNotFound with ResourceType: "currency" when the currency is not known.
// - ErrInvalidArgument with Name:"Amount" when any of the amounts is invalid.
// - ErrPeerRequestTimedOut when peer request times out.
// - ErrPeerRejected when peer rejects the request.
// - ErrPeerNotFunded when peer did not fund the channel in time.
// - ErrTxTimedOut with TxType: "Fund" when funding tx times out.
// - ErrChainNotReachable when connection to blockchain drops while funding.
// - ErrUnknownInternal
func (s *Session) OpenCh(pctx context.Context, openingBalInfo perun.BalInfo, app perun.App, challengeDurSecs uint64) (
	perun.ChInfo, perun.APIError) {
	s.WithField("method", "OpenCh").Infof(
		"\nReceived request with params %+v,%+v,%+v", openingBalInfo, app, challengeDurSecs)
	// Session lock is not acquired at the beginning, but only when adding the channel to session.

	var apiErr perun.APIError
	defer func() {
		if apiErr != nil {
			s.WithFields(perun.APIErrAsMap("OpenCh", apiErr)).Error(apiErr.Message())
		}
	}()

	if !s.isOpen {
		apiErr = perun.NewAPIErrFailedPreCondition(ErrSessionClosed)
		return perun.ChInfo{}, apiErr
	}

	sanitizeBalInfo(openingBalInfo)
	var parts []perun.PeerID
	parts, apiErr = retrievePartIDs(openingBalInfo.Parts, s.idProvider)
	if apiErr != nil {
		return perun.ChInfo{}, apiErr
	}
	registerParts(parts, s.chClient)

	var allocations *pchannel.Allocation
	allocations, apiErr = makeAllocation(openingBalInfo, s.chAsset)
	if apiErr != nil {
		return perun.ChInfo{}, apiErr
	}

	proposal, err := pclient.NewLedgerChannelProposal(challengeDurSecs, s.user.OffChainAddr, allocations,
		makeOffChainAddrs(parts), pclient.WithApp(app.Def, app.Data), pclient.WithRandomNonce())
	if err != nil {
		apiErr = perun.NewAPIErrUnknownInternal(errors.WithMessage(err, "constructing channel proposal"))
		return perun.ChInfo{}, apiErr
	}
	ctx, cancel := context.WithTimeout(pctx, s.timeoutCfg.proposeCh(challengeDurSecs))
	defer cancel()
	pch, err := s.chClient.ProposeChannel(ctx, proposal)
	if err != nil {
		apiErr = s.handleProposeChError(openingBalInfo.Parts, errors.WithMessage(err, "proposing channel"))
		return perun.ChInfo{}, apiErr
	}

	ch := newCh(pch, s.chainURL, openingBalInfo.Currency, openingBalInfo.Parts, s.timeoutCfg, challengeDurSecs)
	s.addCh(ch)
	s.WithFields(log.Fields{"method": "OpenCh", "channelID": ch.ID()}).Info("Channel opened successfully")
	return ch.GetChInfo(), nil
}

// handleProposeChError inspects the passed error, constructs an
// appropriate APIError and returns it.
//
// Passed error must be non-nil.
func (s *Session) handleProposeChError(parts []string, err error) perun.APIError {
	var peerIdx uint16 = 1 // In a sanitized openingBalInfo, peer (proposee) is at index 1.

	var apiErr perun.APIError
	if apiErr = handleChainError(s.chainURL, s.timeoutCfg.onChainTx.String(), err); apiErr != nil {
		return apiErr
	} else if apiErr = handleFundingTimeoutError(parts[peerIdx], peerIdx, err); apiErr != nil {
		return apiErr
	} else if apiErr = handleProposalError(parts[peerIdx], s.timeoutCfg.response.String(), err); apiErr != nil {
		return apiErr
	}
	return perun.NewAPIErrUnknownInternal(err)
}

// handleProposalError inspects if the passed error is a proposal error.
// If yes, it constructs & returns an APIError. If not, returns nil
//
// Passed error must be non-nil.
func handleProposalError(peerAlias, responseTimeout string, err error) perun.APIError {
	peerResponseTimedOutError := pclient.RequestTimedOutError("")
	peerRejectedError := pclient.PeerRejectedError{}

	switch {
	case errors.As(err, &peerResponseTimedOutError):
		return perun.NewAPIErrPeerRequestTimedOut(err, peerAlias, responseTimeout)

	case errors.As(err, &peerRejectedError):
		return perun.NewAPIErrPeerRejected(err, peerAlias, peerRejectedError.Reason)

	default:
		return nil
	}
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
func retrievePartIDs(aliases []string, idProvider perun.IDReader) ([]perun.PeerID, perun.APIError) {
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
		return nil, perun.NewAPIErrResourceNotFound(ResTypePeerID, strings.Join(missingParts, ","))
	}
	if len(repeatedParts) != 0 {
		aliasesValue := strings.Join(aliases, ",")
		return nil, perun.NewAPIErrInvalidArgument(ErrRepeatedPeerAlias, ArgNamePeerAlias, aliasesValue)
	}
	if !foundOwnAlias {
		aliasesValue := strings.Join(aliases, ",")
		return nil, perun.NewAPIErrInvalidArgument(ErrEntryForSelfNotFound, ArgNamePeerAlias, aliasesValue)
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
func makeAllocation(balInfo perun.BalInfo, chAsset pchannel.Asset) (*pchannel.Allocation, perun.APIError) {
	if !currency.IsSupported(balInfo.Currency) {
		return nil, perun.NewAPIErrResourceNotFound(ResTypeCurrency, balInfo.Currency)
	}

	balance := make([]*big.Int, len(balInfo.Bal))
	var err error
	for i := range balInfo.Bal {
		balance[i], err = currency.NewParser(balInfo.Currency).Parse(balInfo.Bal[i])
		if err != nil {
			err = errors.WithMessage(err, "parsing amount")
			return nil, perun.NewAPIErrInvalidArgument(err, ArgNameAmount, balInfo.Bal[i])
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
	s.chs.put(ch)
	s.Unlock()
}

// HandleProposal is a handler to be registered on the channel client for
// processing incoming channel proposals.
func (s *Session) HandleProposal(chProposal pclient.ChannelProposal, responder *pclient.ProposalResponder) {
	s.HandleProposalWInterface(chProposal, &chProposalResponderWrapped{responder})
}

// HandleProposalWInterface is the actual implemention of HandleProposal that
// takes arguments as interface types.
//
// It is implemented this way to enable easier testing.
func (s *Session) HandleProposalWInterface(chProposal pclient.ChannelProposal, responder ChProposalResponder) {
	ledgerChProposal, ok := chProposal.(*pclient.LedgerChannelProposal)
	if !ok {
		// Our handler is expected to handle only ledger channel proposals,
		// if it is anything else (sub-channel proposals), simply drop it.
		return
	}

	s.Debugf("SDK Callback: HandleProposal. Params: %+v", ledgerChProposal)
	expiry := time.Now().UTC().Add(s.timeoutCfg.response).Unix()

	if !s.isOpen {
		// Code will not reach here during runtime as chClient is closed when closing a session.
		s.Error("Unexpected HandleProposal callback invoked on a closed session")
		return
	}

	parts := make([]string, len(ledgerChProposal.Peers))
	for i := range ledgerChProposal.Peers {
		p, ok := s.idProvider.ReadByOffChainAddr(ledgerChProposal.Peers[i])
		if !ok {
			s.Info("Received channel proposal from unknonwn peer ID", ledgerChProposal.Peers[i].String())
			// nolint: errcheck              // It is sufficient to just log this error.
			s.rejectChProposal(context.Background(), responder, "peer ID not found in session ID Provider")
			expiry = 0
			break
		}
		parts[i] = p.Alias
	}

	notif := chProposalNotif(parts, currency.ETH, ledgerChProposal, expiry)
	entry := chProposalResponderEntry{
		proposal:  *ledgerChProposal,
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

func chProposalNotif(parts []string, curr string, chProposal *pclient.LedgerChannelProposal,
	expiry int64) perun.ChProposalNotif {
	return perun.ChProposalNotif{
		ProposalID:       fmt.Sprintf("%x", chProposal.ProposalID()),
		OpeningBalInfo:   makeBalInfoFromRawBal(parts, curr, chProposal.InitBals.Balances[0]),
		App:              makeApp(chProposal.App, chProposal.InitData),
		ChallengeDurSecs: chProposal.ChallengeDuration,
		Expiry:           expiry,
	}
}

// SubChProposals subscribes to notifications on new incoming channel proposals
// in the session. Only one subscription can be made at a time.  Making a new
// subscription without canceling the previous one will return an error.
//
// See perun.ChProposalNotif for the format of notification.
//
// The incoming channel proposal received when there was no subscription will
// have been cached by the node. Once a new subscription is made, node will
// send these cached requests (if any), as individual notifications. It will
// then continue to send a notification for each new incoming channel proposal.
//
// Response to the notifications can be sent using the RespondChProposal API
// before the notification expires.
//
// If the proposal was received from a `Peer ID` that is not found in the ID
// provider of the session, the proposal will be automatically rejected by the
// node. User will still receive a notification of this proposal with the
// `Alias` of the peer set to the hex representation of its off-chain address
// in the `Opening Balance` and the app. These notifications should not be
// responded to. If the user till responds to it, a ErrResourceNotFound error
// will be returned.
//
// If there is errors, it will be one of the following codes:
// - ErrResourceExists with ResourceType: "proposalsSub" when a subscription already exists.
func (s *Session) SubChProposals(notifier perun.ChProposalNotifier) perun.APIError {
	s.WithField("method", "SubChProposals").Info("Received request with params:", notifier)
	s.Lock()
	defer s.Unlock()

	var apiErr perun.APIError
	if !s.isOpen {
		apiErr = perun.NewAPIErrFailedPreCondition(ErrSessionClosed)
		s.WithFields(perun.APIErrAsMap("SubChProposals", apiErr)).Error(apiErr.Message())
		return apiErr
	}

	if s.chProposalNotifier != nil {
		apiErr = perun.NewAPIErrResourceExists(ResTypeProposalSub, s.ID())
		s.WithFields(perun.APIErrAsMap("SubChProposals", apiErr)).Error(apiErr.Message())
		return apiErr
	}
	s.chProposalNotifier = notifier

	// Send all cached notifications.
	for i := len(s.chProposalNotifsCache); i > 0; i-- {
		go s.chProposalNotifier(s.chProposalNotifsCache[0])
		s.chProposalNotifsCache = s.chProposalNotifsCache[1:i]
	}
	s.WithField("method", "SubChProposals").Info("Subscribed successfully")
	return nil
}

// UnsubChProposals unsubscribes from notifications on new incoming channel
// proposals in the specified session.
//
// If there is errors, it will be one of the following codes:
// - ErrResourceNotFound with ResourceType: "proposalsSub" when a subscription does not exist.
func (s *Session) UnsubChProposals() perun.APIError {
	s.WithField("method", "UnsubChProposals").Info("Received request")
	s.Lock()
	defer s.Unlock()

	var apiErr perun.APIError
	if !s.isOpen {
		apiErr = perun.NewAPIErrFailedPreCondition(ErrSessionClosed)
		s.WithFields(perun.APIErrAsMap("UnsubChProposals", apiErr)).Error(apiErr.Message())
		return apiErr
	}

	if s.chProposalNotifier == nil {
		apiErr = perun.NewAPIErrResourceNotFound(ResTypeProposalSub, s.ID())
		s.WithFields(perun.APIErrAsMap("UnsubChProposals", apiErr)).Error(apiErr.Message())
		return apiErr
	}
	s.chProposalNotifier = nil
	s.WithField("method", "UnsubChProposals").Info("Unsubscribed successfully")
	return nil
}

// RespondChProposal responds to the specified channel proposal for which a
// notification had been received. Response should be sent before the
// notification expires. Use the `Time` API to fetch current time of the perun
// node as as reference for checking notification expiry.
//
// If there is errors, it will be one of the following codes:
// - ErrResourceNotFound with ResourceType: "proposal" when proposal ID is not known.
// - ErrFailedPreCondition when session is closed.
// - ErrPeerNotFunded when peer did not fund the channel in time.
// - ErrUserResponseTimedOut when user responded after time out expired.
// - ErrTxTimedOut with TxType: "Fund" when there is tx timed error while funding.
// - ErrChainNotReachable when connection to blockchain drops while funding.
// - ErrUnknownInternal
func (s *Session) RespondChProposal(pctx context.Context, chProposalID string, accept bool) (
	perun.ChInfo, perun.APIError) {
	s.WithField("method", "RespondChProposal").Infof("\nReceived request with Params %+v,%+v", chProposalID, accept)
	// Session lock is not acquired at the beginning, but only when retrieving
	// the proposal and when adding the opened channel to session.

	var apiErr perun.APIError
	defer func() {
		if apiErr != nil {
			s.WithFields(perun.APIErrAsMap("RespondChProposal", apiErr)).Error(apiErr.Message())
		}
	}()

	if !s.isOpen {
		apiErr = perun.NewAPIErrFailedPreCondition(ErrSessionClosed)
		return perun.ChInfo{}, apiErr
	}

	// Lock the session mutex only when retrieving the channel responder and deleting it.
	// It will again be locked when adding the channel to the session.
	s.Lock()
	entry, ok := s.chProposalResponders[chProposalID]
	if !ok {
		s.Unlock()
		apiErr = perun.NewAPIErrResourceNotFound(ResTypeProposal, chProposalID)
		return perun.ChInfo{}, apiErr
	}
	delete(s.chProposalResponders, chProposalID)
	s.Unlock()

	currTime := time.Now().UTC().Unix()
	if entry.notif.Expiry < currTime {
		apiErr = perun.NewAPIErrUserResponseTimedOut(entry.notif.Expiry, currTime)
		return perun.ChInfo{}, apiErr
	}

	var openedChInfo perun.ChInfo
	switch accept {
	case true:
		openedChInfo, apiErr = s.acceptChProposal(pctx, entry)
	case false:
		apiErr = s.rejectChProposal(pctx, entry.responder, "rejected by user")
	}
	return openedChInfo, apiErr
}

func (s *Session) acceptChProposal(pctx context.Context, entry chProposalResponderEntry) (
	perun.ChInfo, perun.APIError) {
	ctx, cancel := context.WithTimeout(pctx, s.timeoutCfg.respChProposalAccept(entry.notif.ChallengeDurSecs))
	defer cancel()

	proposal := entry.proposal
	resp := proposal.Accept(s.user.OffChainAddr, pclient.WithRandomNonce())

	pch, err := entry.responder.Accept(ctx, resp)
	if err != nil {
		err = errors.WithMessage(err, "accepting channel proposal")
		return perun.ChInfo{}, s.handleChProposalAcceptError(entry.notif.OpeningBalInfo.Parts, err)
	}

	// Set ETH as the currency interpreter for incoming channel.
	// TODO: (mano) Provide an option for user to configure when more currency interpreters are supported.
	parts := entry.notif.OpeningBalInfo.Parts
	ch := newCh(pch, s.chainURL, currency.ETH, parts, s.timeoutCfg, entry.notif.ChallengeDurSecs)
	s.addCh(ch)
	s.WithFields(log.Fields{"method": "RespondChProposal", "channelID": ch.ID()}).Info("Channel opened successfully")
	return ch.getChInfo(), nil
}

// handleProposeChannelError inspects the passed error, constructs an
// appropriate APIError and returns it.
//
// Passed error must be non-nil.
func (s *Session) handleChProposalAcceptError(parts []string, err error) perun.APIError {
	var peerIdx uint16 // In a sanitized openingBalInfo, peer (proposer) is at index 0.

	var apiErr perun.APIError
	if apiErr = handleChainError(s.chainURL, s.timeoutCfg.onChainTx.String(), err); apiErr != nil {
		return apiErr
	} else if apiErr = handleFundingTimeoutError(parts[peerIdx], peerIdx, err); apiErr != nil {
		return apiErr
	}
	return perun.NewAPIErrUnknownInternal(err)
}

func (s *Session) rejectChProposal(pctx context.Context, responder ChProposalResponder,
	reason string) perun.APIError {
	ctx, cancel := context.WithTimeout(pctx, s.timeoutCfg.respChProposalReject())
	defer cancel()
	err := responder.Reject(ctx, reason)
	if err != nil {
		return perun.NewAPIErrUnknownInternal(err)
	}
	return nil
}

// GetChsInfo gets the list of all channels in the session with their latest
// agreed state.
func (s *Session) GetChsInfo() []perun.ChInfo {
	s.WithField("method", "GetChsInfo").Info("Received request")
	s.Lock()
	defer s.Unlock()

	chsInfo := make([]perun.ChInfo, s.chs.count())
	s.chs.forEach(func(i int, ch *Channel) {
		chsInfo[i] = ch.GetChInfo()
	})

	return chsInfo
}

// GetCh is an internal API that retreives the channel API instance
// corresponding to the given channel ID.
//
// The channel instance is safe for concurrent user.
//
// If there is errors, it will be one of the following codes:
// - perun.ErrResourceNotFound when the channel ID is not known.
func (s *Session) GetCh(chID string) (perun.ChAPI, perun.APIError) {
	s.WithField("method", "GetCh").Info("Received request with params:", chID)

	s.Lock()
	ch := s.chs.get(chID)
	s.Unlock()

	if ch == nil {
		apiErr := perun.NewAPIErrResourceNotFound(ResTypeChannel, chID)
		s.WithFields(perun.APIErrAsMap("GetCh (internal)", apiErr)).Error(apiErr.Message())
		return nil, apiErr
	}

	s.WithField("method", "GetCh").Info("Channel retrieved:")
	return ch, nil
}

// HandleUpdate is a handler to be registered on the channel client for
// processing incoming channel updates.  This function just identifies the
// channel to which update is received and invokes the handler for that
// channel.
func (s *Session) HandleUpdate(
	currState *pchannel.State, chUpdate pclient.ChannelUpdate, responder *pclient.UpdateResponder) {
	s.HandleUpdateWInterface(currState, chUpdate, responder)
}

// HandleUpdateWInterface is the actual implemention of HandleUpdate that takes
// arguments as interface types.
// It is implemented this way to enable easier testing.
func (s *Session) HandleUpdateWInterface(
	currState *pchannel.State, chUpdate pclient.ChannelUpdate, responder ChUpdateResponder) {
	s.Debugf("SDK Callback: HandleUpdate. Params: %+v", chUpdate)
	s.Lock()
	defer s.Unlock()

	if !s.isOpen {
		// Code will not reach here during runtime as chClient is closed when closing a session.
		s.Error("Unexpected HandleUpdate callback invoked on a closed session")
		return
	}

	chID := fmt.Sprintf("%x", chUpdate.State.ID)
	ch := s.chs.get(chID)
	if ch == nil {
		s.Info("Received update for unknown channel", chID)
		err := responder.Reject(context.Background(), "unknown channel for this session")
		s.Info("Error rejecting incoming update for unknown channel with id %s: %v", chID, err)
		return
	}
	go ch.HandleUpdate(currState, chUpdate, responder)
}

// Close closes the specified session. All session data will be persisted to
// disk.
//
// `Force` parameter determines what happens when there are open channels in the
// session.
//   * If `False` the API returns an error when there are open channels. This
//     should be used by default.
//   * If `True`, the session is forcibly closed and the API returns list of open
//     channels that were persisted. When a session is re-opened with the same
//     config file, these channels can be restored in open state. However, use this
//     with caution, as closing a session with open channels creates a possibility
//     for channel participants in any of the those open open channels to register
//     an older, invalid state on the blockchain and finalize it.
//
// If there is an error, it will be one of the following codes:
// - ErrFailedPreCondition when session is closed with force=false and unclosed channels
//   exists. Additional Info will contain an extra field: OpenChannelsInfo
//   that contains a list of Channel Info.
// - ErrUnknownInternal
func (s *Session) Close(force bool) ([]perun.ChInfo, perun.APIError) {
	s.WithField("method", "Close").Infof("\nReceived request with params %+v", force)
	s.Debug("Received request: session.Close")
	s.Lock()
	defer s.Unlock()

	var apiErr perun.APIError
	defer func() {
		if apiErr != nil {
			s.WithFields(perun.APIErrAsMap("Close", apiErr)).Error(apiErr.Message())
		}
	}()

	if !s.isOpen {
		apiErr = perun.NewAPIErrFailedPreCondition(ErrSessionClosed)
		return nil, apiErr
	}

	openChsInfo := []perun.ChInfo{}
	unexpectedPhaseChIDs := []perun.ChInfo{}
	s.chs.forEach(func(i int, ch *Channel) {
		// Acquire channel mutex to ensure any ongoing operation on the channel is finished.
		ch.Lock()

		// Calling Phase() also waits for the mutex on pchannel that ensures any handling of Registered event
		// in the Watch routine is also completed. But if the event was received after acquiring channel mutex
		// and completed before pc.Phase() returned, this event will not yet be serviced by perun-node.
		// A solution to this is to add a provision (that is currently missing) to suspend the Watcher (only
		// for open channels) before acquiring channel mutex and restoring it later if force option is false.
		//
		// TODO (mano): Add a provision in go-perun to suspend the watcher and use it here.
		//
		// Since there will be no ongoing operations in perun-node, the pchannel should be in one of the two
		// stable phases known to perun node (see state diagram in the docs for details) : Acting or Withdrawn.
		phase := ch.pch.Phase()
		if phase != pchannel.Acting && phase != pchannel.Withdrawn {
			unexpectedPhaseChIDs = append(unexpectedPhaseChIDs, ch.getChInfo())
		}
		if ch.status == open {
			openChsInfo = append(openChsInfo, ch.getChInfo())
		}
	})

	if len(unexpectedPhaseChIDs) != 0 {
		s.unlockAllChs()
		err := errors.New("session cannot be closed with channels in unexpected phase")
		apiErr = perun.NewAPIErrFailedPreConditionUnclosedChs(err, unexpectedPhaseChIDs)
		return nil, apiErr
	}
	if !force && len(openChsInfo) != 0 {
		s.unlockAllChs()
		err := errors.New("session cannot be closed with channels in open phase without force option")
		apiErr = perun.NewAPIErrFailedPreConditionUnclosedChs(err, openChsInfo)
		return nil, apiErr
	}

	s.isOpen = false
	apiErr = s.close()
	return openChsInfo, apiErr
}

func (s *Session) unlockAllChs() {
	s.chs.forEach(func(i int, ch *Channel) {
		ch.Unlock()
	})
}

func (s *Session) close() perun.APIError {
	s.user.OnChain.Wallet.LockAll()
	s.user.OffChain.Wallet.LockAll()
	err := errors.WithMessage(s.chClient.Close(), "closing session")
	if err != nil {
		return perun.NewAPIErrUnknownInternal(err)
	}
	return nil
}

// handleFundingTimeoutError inspects if the passed error is an funding timeout error.
// If yes, it constructs & returns an APIError. If not, returns nil
//
// Passed error must be non-nil.
func handleFundingTimeoutError(peerAlias string, peerIdx uint16, err error) perun.APIError {
	fundingTimeoutError := pchannel.FundingTimeoutError{}
	ok := errors.As(err, &fundingTimeoutError)
	if !ok {
		return nil
	}
	if len(fundingTimeoutError.Errors) != 1 {
		err = errors.WithMessage(err, "channel can contain only one asset")
		return perun.NewAPIErrUnknownInternal(err)
	}
	if len(fundingTimeoutError.Errors[0].TimedOutPeers) != 1 {
		err = errors.WithMessage(err, "channel can contain only one participant other than self")
		return perun.NewAPIErrUnknownInternal(err)
	}
	if fundingTimeoutError.Errors[0].TimedOutPeers[0] != peerIdx {
		err = errors.WithMessage(err, fmt.Sprintf("index of peer must be %d", peerIdx))
		return perun.NewAPIErrUnknownInternal(err)
	}
	return perun.NewAPIErrPeerNotFunded(err, peerAlias)
}

// handleChainError inspects if the passed error is an on-chain error.
// If yes, it constructs & returns an APIError. If not, returns nil
//
// Passed error must be non-nil.
func handleChainError(chainURL, onChainTxTimeout string, err error) perun.APIError {
	txTimedOutError := pclient.TxTimedoutError{}
	chainNotReachableError := pclient.ChainNotReachableError{}

	switch {
	case errors.As(err, &txTimedOutError):
		txType := txTimedOutError.TxType
		txID := txTimedOutError.TxID
		return perun.NewAPIErrTxTimedOut(err, txType, txID, onChainTxTimeout)

	case errors.As(err, &chainNotReachableError):
		return perun.NewAPIErrChainNotReachable(err, chainURL)

	default:
		return nil
	}
}
