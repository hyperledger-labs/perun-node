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
	"fmt"
	"math/big"
	"sync"
	"time"

	pchannel "perun.network/go-perun/channel"
	pclient "perun.network/go-perun/client"
	pwire "perun.network/go-perun/wire"
	psync "polycry.pt/poly-go/sync"

	"github.com/pkg/errors"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/log"
)

const (
	open   chStatus = iota // Open for off-chain tx.
	closed                 // Closed for off-chain tx, settled on-chain and amount withdrawn.
)

type (
	// Channel represents the perun channel established between different
	// parties.
	//
	// It implements the perun.ChAPI interface.
	Channel struct {
		log.Logger
		params

		pch               PChannel
		status            chStatus
		wasCloseInitiated bool

		chUpdateNotifier   perun.ChUpdateNotifier
		chUpdateNotifCache []perun.ChUpdateNotif
		chUpdateResponders map[string]chUpdateResponderEntry

		watcherWg *sync.WaitGroup
		psync.Mutex
	}

	// params represent the parameters of the channel that do not change after
	// it is initialized.
	params struct {
		id               string
		currencies       []perun.Currency
		symbols          map[string]int
		parts            []string
		timeoutCfg       timeoutConfig
		challengeDurSecs uint64
		chainURL         string
	}

	// PChannel represents the methods on the state channel controller defined
	// in go-perun used by the perun node. This interface is introduced for the
	// purpose of mocking during tests.
	PChannel interface {
		Close() error
		ID() pchannel.ID
		Idx() pchannel.Index
		IsClosed() bool
		Params() *pchannel.Params
		Peers() []pwire.Address
		Phase() pchannel.Phase
		State() *pchannel.State
		OnUpdate(cb func(from, to *pchannel.State))
		Update(ctx context.Context, update func(*pchannel.State)) error
		Settle(ctx context.Context, isSecondary bool) error
		Watch(pclient.AdjudicatorEventHandler) error
	}

	chStatus uint8

	chUpdateResponderEntry struct {
		notif       perun.ChUpdateNotif
		responder   ChUpdateResponder
		notifExpiry int64
	}

	// ChUpdateResponder represents the methods on channel update responder that will be used the perun node.
	ChUpdateResponder interface {
		Accept(ctx context.Context) error
		Reject(ctx context.Context, reason string) error
	}
)

//go:generate mockery --name PChannel --output ../internal/mocks

//go:generate mockery --name ChUpdateResponder --output ../internal/mocks

// newCh initializes  a channel instance using the passed pchannel (controller)
// and other channel parameters.
func newCh(pch PChannel, chainURL string, currencies []perun.Currency, parts []string, timeoutCfg timeoutConfig,
	challengeDurSecs uint64) *Channel {
	ch := &Channel{
		params: params{
			id:               fmt.Sprintf("%x", pch.ID()),
			timeoutCfg:       timeoutCfg,
			challengeDurSecs: challengeDurSecs,
			chainURL:         chainURL,
			currencies:       currencies,
			symbols:          make(map[string]int, len(currencies)),
			parts:            parts,
		},
		pch:                pch,
		status:             open,
		wasCloseInitiated:  false,
		chUpdateResponders: make(map[string]chUpdateResponderEntry),
		watcherWg:          &sync.WaitGroup{},
	}
	for i := range currencies {
		ch.symbols[currencies[i].Symbol()] = i
	}

	ch.watcherWg.Add(1)
	go func(ch *Channel) {
		err := ch.pch.Watch(ch)
		ch.watcherWg.Done()
		ch.Errorf("Watcher returned with error: %+v", err)
	}(ch)
	return ch
}

// HandleAdjudicatorEvent is invoked when an on-chain event is received from
// the adjudicator contract.
// It process the event depending upon its type and whether the channel is
// finalized (collaborative close) or not (non-collaborative close).
func (ch *Channel) HandleAdjudicatorEvent(e pchannel.AdjudicatorEvent) {
	ch.Debugf("Received HandleAdjudicatorEvent of type %T: %+v", e, e)
	ch.Lock()
	defer ch.Unlock()

	switch e.(type) {

	// Valid only when that state is Final and close was not initiated by us.
	case *pchannel.ConcludedEvent:
		// For collaborative close, this type of event will be received after one
		// of the parties registered the final state on the chain. The channel is
		// settled on this event.
		//
		// For non-collaborative close, both parties receive this event after
		// channel is concluded. But it should BE IGNORED, as it will be handled by
		// go-perun framework itself.
		// For details, see the ensureConcluded call in adjudicator.Withdraw
		// implementation in go-perun/backend/ethereum/channel package.
		if !ch.wasCloseInitiated {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			apiErr := ch.settle(ctx)
			if apiErr != nil {
				ch.WithField("method", "HandleAdjudicatorEvent").Errorf("Settling the channel: %v", apiErr)
			} else {
				ch.WithField("method", "HandleAdjudicatorEvent").Info("Channel settled")
			}
			ch.closeAndNotify(apiErr)
			return
		}

	default:
		ch.Infof("Ignoring adjudicator event that is not of type ConcludedEvent")
	}
}

// settle concludes the channel on-chain and ensures the funds are withdrawn.
func (ch *Channel) settle(pctx context.Context) perun.APIError {
	ctx, cancel := context.WithTimeout(pctx, ch.timeoutCfg.settle(ch.challengeDurSecs))
	defer cancel()
	err := ch.pch.Settle(ctx, !ch.wasCloseInitiated)
	if err != nil {
		return ch.handleChSettleError(errors.WithMessage(err, "settling channel"))
	}
	return nil
}

// handleChSettleError inspects the passed error, constructs an
// appropriate APIError and returns it.
func (ch *Channel) handleChSettleError(err error) perun.APIError {
	var apiErr perun.APIError
	if apiErr = handleChainError(ch.chainURL, ch.timeoutCfg.onChainTx.String(), err); apiErr != nil {
		return apiErr
	}
	return perun.NewAPIErrUnknownInternal(err)
}

// closeAndNotify marks the channel as closed and sends a channel close
// notification if an active subscription for channel update already exists.
// The notification is dropped otherwise. Because the user will not able to
// subscribe to update notifications for a channel after it is closed.
func (ch *Channel) closeAndNotify(err perun.APIError) {
	ch.close()
	ch.Info("Channel closed")

	if ch.chUpdateNotifier == nil {
		ch.Debug("Channel close notification dropped as there is no active subscription")
		return
	}
	notif := makeChCloseNotif(ch.getChInfo(), err)
	ch.chUpdateNotifier(notif)
	ch.unsubChUpdates()
	ch.Debug("Channel close notification sent")
}

func makeChCloseNotif(currChInfo perun.ChInfo, err perun.APIError) perun.ChUpdateNotif {
	return perun.ChUpdateNotif{
		UpdateID:       fmt.Sprintf("%s_%s_%s", currChInfo.ChID, currChInfo.Version, "closed"),
		CurrChInfo:     currChInfo,
		ProposedChInfo: perun.ChInfo{},
		Type:           perun.ChUpdateTypeClosed,
		Expiry:         0,
		Error:          err,
	}
}

// ID returns the ID of the channel.
//
// Does not require a mutex lock, as the data will remain unchanged throughout
// the lifecycle of the channel.
func (ch *Channel) ID() string {
	return ch.id
}

// Currencies returns the currency interpreters used in the channel.
//
// Does not require a mutex lock, as the data will remain unchanged throughout
// the lifecycle of the channel.
func (ch *Channel) Currencies() []perun.Currency {
	return ch.currencies
}

// Currency returns the index of this currency in channel balances and the
// currency interpreter.  If symbol is unknown, returns an false.
//
// Does not require a mutex lock, as the data will remain unchanged throughout
// the lifecycle of the channel.
func (ch *Channel) Currency(symbol string) (int, perun.Currency, bool) {
	idx, found := ch.symbols[symbol]
	if !found {
		return 0, nil, false
	}
	return idx, ch.currencies[idx], true
}

// Parts returns the list of aliases of the channel participants.
//
// Does not require a mutex lock, as the data will remain unchanged throughout
// the lifecycle of the channel.
func (ch *Channel) Parts() []string {
	return ch.parts
}

// ChallengeDurSecs returns the challenge duration for the channel (in seconds)
// for refuting when an invalid/older state is registered on the blockchain
// closing the channel.
//
// Does not require a mutex lock, as the data will remain unchanged throughout
// the lifecycle of the channel.
func (ch *Channel) ChallengeDurSecs() uint64 {
	return ch.challengeDurSecs
}

// SendChUpdate sends an update on the channel. The state will be passed to the
// updater function which can update it. The updated state will then be
// validated and then sent to other participants for their signature.
//
// If there is an error, it will be one of the following codes:
// - ErrInvalidArgument with Name:"Amount" when any of the amounts is invalid.
// - ErrPeerRequestTimedOut when peer request times out.
// - ErrPeerRejected when peer rejects the request.
// - ErrUnknownInternal.
func (ch *Channel) SendChUpdate(pctx context.Context, updater perun.StateUpdater) (perun.ChInfo, perun.APIError) {
	ch.WithField("method", "SendChUpdate").Infof("\nReceived request with params %+v", updater)
	ch.Lock()
	defer ch.Unlock()

	var apiErr perun.APIError
	defer func() {
		if apiErr != nil {
			ch.WithFields(perun.APIErrAsMap("SendChUpdate", apiErr)).Error(apiErr.Message())
		}
	}()

	if ch.status == closed {
		apiErr = perun.NewAPIErrFailedPreCondition(ErrChClosed)
		return ch.getChInfo(), apiErr
	}

	ctx, cancel := context.WithTimeout(pctx, ch.timeoutCfg.chUpdate())
	defer cancel()
	err := ch.pch.Update(ctx, updater)
	if err != nil {
		apiErr = ch.handleSendChUpdateError(errors.WithMessage(err, "sending channel update"))
		return perun.ChInfo{}, apiErr
	}
	ch.WithField("method", "SendChUpdate").Info("State updated successfully")
	return ch.getChInfo(), nil
}

// handleSendChUpdateError inspects the passed error, constructs an
// appropriate APIError and returns it.
func (ch *Channel) handleSendChUpdateError(err error) perun.APIError {
	peerAlias := ch.parts[ch.pch.Idx()^1] // Logic works only for a two party channel.
	if apiErr := handleProposalError(peerAlias, ch.timeoutCfg.response.String(), err); apiErr != nil {
		return apiErr
	}
	return perun.NewAPIErrUnknownInternal(err)
}

// HandleUpdate handles the incoming updates on an open channel. All updates
// are sent to a centralized update handler defined on the session. The
// centrazlied handler identifies the channel and then invokes this function to
// process the update.
func (ch *Channel) HandleUpdate(
	currState *pchannel.State, chUpdate pclient.ChannelUpdate, responder ChUpdateResponder) {
	ch.Lock()
	defer ch.Unlock()

	if ch.status == closed {
		ch.Error("Unexpected HandleUpdate call for closed channel")
		return
	}

	expiry := time.Now().UTC().Add(ch.timeoutCfg.response).Unix()
	currChInfo := ch.makeChInfo(currState)
	notif := ch.makeChUpdateNotif(currChInfo, chUpdate.State, expiry)
	entry := chUpdateResponderEntry{
		notif:       notif,
		responder:   responder,
		notifExpiry: expiry,
	}

	// Need not store entries for notification with expiry = 0, as these update requests have
	// already been rejected by the perun node. Hence no response is expected for these notifications.
	if expiry != 0 {
		ch.chUpdateResponders[notif.UpdateID] = entry
	}
	ch.sendChUpdateNotif(notif)
}

func (ch *Channel) sendChUpdateNotif(notif perun.ChUpdateNotif) {
	if ch.chUpdateNotifier == nil {
		ch.chUpdateNotifCache = append(ch.chUpdateNotifCache, notif)
		ch.Debug("HandleUpdate: Notification cached")
		return
	}
	go func() {
		ch.chUpdateNotifier(notif)
		ch.Debug("HandleUpdate: Notification sent")
	}()
}

func (ch *Channel) makeChUpdateNotif(
	currChInfo perun.ChInfo, proposedState *pchannel.State, expiry int64) perun.ChUpdateNotif {
	var chUpdateType perun.ChUpdateType
	switch proposedState.IsFinal {
	case true:
		chUpdateType = perun.ChUpdateTypeFinal
	case false:
		chUpdateType = perun.ChUpdateTypeOpen
	}
	return perun.ChUpdateNotif{
		UpdateID:       fmt.Sprintf("%s_%d", ch.ID(), proposedState.Version),
		CurrChInfo:     currChInfo,
		ProposedChInfo: ch.makeChInfo(proposedState),
		Type:           chUpdateType,
		Expiry:         expiry,
		Error:          nil,
	}
}

// SubChUpdates subscribes to notifications on new incoming channel updates for
// the specified channel in the session. Only one subscription can be made at a
// time. Making a new subscription without canceling the previous one will
// return an error.
//
// See perun.ChUpateNotif for the format of notification.
//
// The incoming channel update received when there was no subscription will
// have been cached by the node. Once a new subscription is made, node will
// send these cached requests (if any), as individual notifications. It will
// then continue to send a notification for each new incoming channel update.
//
// Response to the notifications can be sent using the RespondChUpdate API
// before the notification expires.
//
// If there is an error, it will be one of the following codes:
// - ErrResourceExists with ResourceType: "updatesSub" when a subscription already exists.
func (ch *Channel) SubChUpdates(notifier perun.ChUpdateNotifier) perun.APIError {
	ch.WithField("method", "SubChUpdates").Info("Received request with params:", notifier)
	ch.Lock()
	defer ch.Unlock()

	if ch.status == closed {
		apiErr := perun.NewAPIErrFailedPreCondition(ErrChClosed)
		ch.WithFields(perun.APIErrAsMap("SubChUpdates", apiErr)).Error(apiErr.Message())
		return apiErr
	}

	if ch.chUpdateNotifier != nil {
		apiErr := perun.NewAPIErrResourceExists(ResTypeUpdateSub, ch.ID())
		ch.WithFields(perun.APIErrAsMap("SubChUpdates", apiErr)).Error(apiErr.Message())
		return apiErr
	}
	ch.chUpdateNotifier = notifier

	// Send all cached notifications
	for i := len(ch.chUpdateNotifCache); i > 0; i-- {
		go ch.chUpdateNotifier(ch.chUpdateNotifCache[0])
		ch.chUpdateNotifCache = ch.chUpdateNotifCache[1:i]
	}
	return nil
}

// UnsubChUpdates unsubscribes from notifications on new incoming channel
// updates for the specified channel in the specified session.
//
// If there is an error, it will be one of the following codes:
// - ErrResourceNotFound with ResourceType: "updatesSub" when a subscription does not exist.
func (ch *Channel) UnsubChUpdates() perun.APIError {
	ch.WithField("method", "UnsubChUpdates").Info("Received request")
	ch.Lock()
	defer ch.Unlock()

	if ch.status == closed {
		apiErr := perun.NewAPIErrFailedPreCondition(ErrChClosed)
		ch.WithFields(perun.APIErrAsMap("UnsubChUpdates", apiErr)).Error(apiErr.Message())
		return apiErr
	}

	if ch.chUpdateNotifier == nil {
		apiErr := perun.NewAPIErrResourceNotFound(ResTypeUpdateSub, ch.ID())
		ch.WithFields(perun.APIErrAsMap("UnsubChUpdates", apiErr)).Error(apiErr.Message())
		return apiErr
	}
	ch.unsubChUpdates()
	return nil
}

func (ch *Channel) unsubChUpdates() {
	ch.chUpdateNotifier = nil
}

// RespondChUpdate responds to an incoming channel update for which a
// notification had been received. Response should be sent before the
// notification expires. Use the `Time` API to fetch current time of the perun
// node for checking notification expiry.
//
// If there is an error, it will be one of the following codes:
// - ErrResourceNotFound with ResourceType: "update" when update ID is not known.
// - ErrUserResponseTimedOut when user responded after time out expired.
// - ErrUnknownInternal.
func (ch *Channel) RespondChUpdate(pctx context.Context, updateID string, accept bool) (
	perun.ChInfo, perun.APIError) {
	ch.WithField("method", "RespondChUpdate").Infof("\nReceived request with params %+v,%+v", updateID, accept)
	ch.Lock()
	defer ch.Unlock()

	var apiErr perun.APIError
	defer func() {
		if apiErr != nil {
			ch.WithFields(perun.APIErrAsMap("RespondChUpdate", apiErr)).Error(apiErr.Message())
		}
	}()

	if ch.status == closed {
		apiErr = perun.NewAPIErrFailedPreCondition(ErrChClosed)
		return ch.getChInfo(), apiErr
	}

	entry, ok := ch.chUpdateResponders[updateID]
	if !ok {
		apiErr = perun.NewAPIErrResourceNotFound(ResTypeUpdate, updateID)
		return ch.getChInfo(), apiErr
	}
	delete(ch.chUpdateResponders, updateID)

	currTime := time.Now().UTC().Unix()
	if entry.notifExpiry < currTime {
		apiErr = perun.NewAPIErrUserResponseTimedOut(entry.notif.Expiry, currTime)
		return ch.getChInfo(), apiErr
	}

	switch accept {
	case true:
		apiErr = ch.acceptChUpdate(pctx, entry)
		if apiErr == nil {
			ch.WithField("method", "RespondChUpdate").Info("Channel update accepted successfully")
		}
		if apiErr == nil && entry.notif.Type == perun.ChUpdateTypeFinal {
			apiErr = ch.settle(pctx)
			if apiErr != nil {
				ch.Errorf("Settling the channel with the finalized state: %v", apiErr)
			} else {
				ch.Info("Settled the channel with the finalized state")
			}
			ch.closeAndNotify(apiErr)
		}
	case false:
		apiErr = ch.rejectChUpdate(pctx, entry, "rejected by user")
		if apiErr == nil {
			ch.WithField("method", "RespondChUpdate").Info("Channel update rejected successfully")
		}
	}
	return ch.getChInfo(), apiErr
}

func (ch *Channel) acceptChUpdate(pctx context.Context, entry chUpdateResponderEntry) perun.APIError {
	ctx, cancel := context.WithTimeout(pctx, ch.timeoutCfg.respChUpdate())
	defer cancel()
	err := entry.responder.Accept(ctx)
	if err != nil {
		ch.Error("Accepting channel update", err)
		return perun.NewAPIErrUnknownInternal(errors.Wrap(err, "accepting update"))
	}
	return nil
}

func (ch *Channel) rejectChUpdate(pctx context.Context, entry chUpdateResponderEntry, reason string) perun.APIError {
	ctx, cancel := context.WithTimeout(pctx, ch.timeoutCfg.respChUpdate())
	defer cancel()
	err := entry.responder.Reject(ctx, reason)
	if err != nil {
		ch.Error("Rejecting channel update", err)
		return perun.NewAPIErrUnknownInternal(errors.Wrap(err, "rejecting update"))
	}
	return nil
}

// GetChInfo gets the last agreed state of the specified payment channel.
func (ch *Channel) GetChInfo() perun.ChInfo {
	ch.WithField("method", "GetChInfo").Info("Received request")
	ch.Lock()
	chInfo := ch.getChInfo()
	ch.Unlock()
	return chInfo
}

// This function assumes that caller has already locked the channel.
func (ch *Channel) getChInfo() perun.ChInfo {
	return ch.makeChInfo(ch.pch.State().Clone())
}

func (ch *Channel) makeChInfo(state *pchannel.State) perun.ChInfo {
	if state == nil {
		return perun.ChInfo{}
	}
	return perun.ChInfo{
		ChID:    ch.id,
		BalInfo: makeBalInfoFromState(ch.parts, ch.currencies, state),
		App:     makeApp(state.App, state.Data),
		Version: fmt.Sprintf("%d", state.Version),
	}
}

// makeApp returns perun.makeApp formed from the given app def and app data.
func makeApp(def pchannel.App, data pchannel.Data) perun.App {
	return perun.App{
		Def:  def,
		Data: data,
	}
}

// makeBalInfoFromState retrieves balance information from the channel state.
func makeBalInfoFromState(parts []string, currencies []perun.Currency, state *pchannel.State) perun.BalInfo {
	return makeBalInfoFromRawBal(parts, currencies, state.Balances)
}

// makeBalInfoFromRawBal retrieves balance information from the raw balance.
func makeBalInfoFromRawBal(parts []string, currencies []perun.Currency, rawBal [][]*big.Int) perun.BalInfo {
	balInfo := perun.BalInfo{
		Currencies: make([]string, len(currencies)),
		Parts:      parts,
		Bals:       make([][]string, len(rawBal)),
	}
	for i := range rawBal {
		balInfo.Currencies[i] = currencies[i].Symbol()
		balInfo.Bals[i] = make([]string, len(rawBal[i]))
		for j := range rawBal[i] {
			balInfo.Bals[i][j] = currencies[i].Print(rawBal[i][j])
		}
	}

	return balInfo
}

// Close closes the channel. First it tries to finalize the last agreed state
// of the payment channel off-chain (by sending a finalizing update) and then
// settling it on the blockchain. If the channel participants reject/not
// respond to the finalizing update, the last agreed state will be finalized
// directly on the blockchain. The call will return after this.
//
// The node will then wait for the challenge duration to pass (if the channel was
// directly settled on the blockchain) and the withdraw the balance as per the
// settled state to the user's account. It then sends a channel update
// notification with update types as `Closed`.
//
// If there is an error in the closing update, it will be one of the following codes:
// - ErrTxTimedOut with TxType: "Conclude" or "ConcludeFinal" when on-chain finalizing tx times out.
// - ErrTxTimedOut with TxType: "Withdraw"  when withdrawing tx times out.
// - ErrChainNotReachable when connection to blockchain drops while finalizing on-chain or withdrawing.
// - ErrUnknownInternal
//
// If there is an error returned by this API, it will be one of the following codes:
// - ErrTxTimedOut with TxType: "Register" when register tx times out.
// - ErrChainNotReachable when connection to blockchain drops while register.
// - ErrUnknownInternal.
func (ch *Channel) Close(pctx context.Context) (perun.ChInfo, perun.APIError) {
	ch.WithField("method", "ChClose").Infof("\nReceived request")
	ch.Lock()
	defer ch.Unlock()

	var apiErr perun.APIError
	defer func() {
		if apiErr != nil {
			ch.WithFields(perun.APIErrAsMap("ChClose", apiErr)).Error(apiErr.Message())
		}
	}()

	if ch.status == closed {
		apiErr = perun.NewAPIErrFailedPreCondition(ErrChClosed)
		return ch.getChInfo(), apiErr
	}

	ch.wasCloseInitiated = true
	ch.finalize(pctx)
	apiErr = ch.settle(pctx)
	if apiErr != nil {
		ch.WithField("method", "HandleAdjudicatorEvent").Errorf("Settling the channel: %v", apiErr)
	} else {
		ch.WithField("method", "HandleAdjudicatorEvent").Info("Channel settled")
	}
	ch.closeAndNotify(apiErr)
	ch.WithField("method", "ChClose").Info("Channel closed")
	return ch.getChInfo(), apiErr
}

// finalize tries to finalize the channel offchain by sending an update with isFinal = true
// to all channel participants.
//
// If this succeeds, calling Settle consequently will close the channel collaboratively by directly settling
// the channel on the blockchain without registering or waiting for challenge duration to expire.
// If this fails, calling Settle consequently will close the channel non-collaboratively, by registering
// the state on-chain and waiting for challenge duration to expire.
func (ch *Channel) finalize(pctx context.Context) {
	chFinalizer := func(state *pchannel.State) {
		state.IsFinal = true
	}
	ctx, cancel := context.WithTimeout(pctx, ch.timeoutCfg.chUpdate())
	defer cancel()
	err := ch.pch.Update(ctx, chFinalizer)
	if err != nil {
		apiErr := ch.handleSendChUpdateError(err)
		ch.WithFields(perun.APIErrAsMap("ChClose", apiErr)).Error(apiErr.Message())
		ch.Info("Channel not finalized. Proceeding with non-collaborative close")
		return
	}

	ch.Info("Channel finalized. Proceeding with collaborative close")
}

// Close the computing resources (listeners, subscriptions etc.,) of the channel.
// If it fails, this error can be ignored.
// It also removes the channel from the session.
func (ch *Channel) close() {
	if err := ch.pch.Close(); err != nil {
		ch.WithField("method", "ChClose").Errorf("\nClosing channel %v", err)
	}
	ch.watcherWg.Wait()
	ch.status = closed
}
