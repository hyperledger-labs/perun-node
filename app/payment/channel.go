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

package payment

import (
	"context"
	"math/big"

	pchannel "perun.network/go-perun/channel"

	"github.com/pkg/errors"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/session"
)

// Error type is used to define error constants for this package.
type Error string

// Error implements error interface.
func (e Error) Error() string {
	return string(e)
}

// Definition of error constants for this package.
const (
	ErrInvalidAmount Error = "invalid amount"
	ErrInvalidPayee  Error = "invalid payee"
)

type (
	// Payment contains the info required for making a single payment.
	Payment struct {
		Currency string
		Payee    string
		Amount   string
	}

	// PayChInfo represents the interpretation of channelInfo for payment app.
	PayChInfo struct {
		ChID    string
		BalInfo perun.BalInfo
		Version string
	}

	// PayChUpdateNotifier represents the interpretation of channel update notifier for payment app.
	PayChUpdateNotifier func(PayChUpdateNotif)

	// PayChUpdateNotif represents the interpretation of channel update notification for payment app.
	// ProposedChInfo (of ChUpdateNotif) is sent in the ChInfo field for regular updates and
	// CurrChInfo (of ChCloseNotif) is sent in the ChInfo field for channel close update.
	// See perun.ChUpdateNotif for documentation on the other struct fields.
	PayChUpdateNotif struct {
		UpdateID          string
		ProposedPayChInfo PayChInfo
		Type              perun.ChUpdateType
		Expiry            int64
		Error             perun.APIError
	}
)

// SendPayChUpdate sends a payment update on the channel that can send or
// request funds. Use `self` in the `payee` field to pay the user itself and
// "alias of the peer" to pay the peer.
//
// If there is an error, it will be one of the following codes:
// - ErrResourceNotFound with ResourceType: "peerID" when any of the peer aliases are not known.
// - ErrInvalidArgument with Name:"Amount" when the amount is invalid.
// or any of the errors returned by the session.SendChUpdate API.
func SendPayChUpdate(pctx context.Context, ch perun.ChAPI, payments []Payment) (PayChInfo, perun.APIError) {
	updates := make([]func(state *pchannel.State), len(payments))

	for i := range payments {
		idxOfCurrencyInBals, currency, found := ch.Currency(payments[i].Currency)
		if !found {
			return PayChInfo{}, perun.NewAPIErrResourceNotFound(session.ResTypeCurrency, payments[i].Currency)
		}
		parsedAmount, err := currency.Parse(payments[i].Amount)
		if err != nil {
			err = errors.WithMessage(err, ErrInvalidAmount.Error())
			return PayChInfo{}, perun.NewAPIErrInvalidArgument(err, session.ArgNameAmount, payments[i].Amount)
		}
		payerIdx, payeeIdx, err := getPayerPayeeIdx(ch.Parts(), payments[i].Payee)
		if err != nil {
			return PayChInfo{}, perun.NewAPIErrInvalidArgument(err, session.ArgNamePayee, payments[i].Payee)
		}
		updates[i] = newUpdate(payerIdx, payeeIdx, idxOfCurrencyInBals, parsedAmount)
	}

	chInfo, apiErr := ch.SendChUpdate(pctx, func(state *pchannel.State) error {
		for i := range updates {
			updates[i](state)
		}
		return nil
	})
	return toPayChInfo(chInfo), apiErr
}

func getPayerPayeeIdx(parts []string, payee string) (payerIdx, payeeIdx int, _ error) {
	for i := range parts {
		if parts[i] == payee {
			payeeIdx = i
			payerIdx = payeeIdx ^ 1
			return payerIdx, payeeIdx, nil
		}
	}
	return 0, 0, ErrInvalidPayee
}

func newUpdate(payerIdx, payeeIdx, idxOfCurrencyInBals int, parsedAmount *big.Int) func(state *pchannel.State) {
	return func(state *pchannel.State) {
		bal := state.Allocation.Balances[idxOfCurrencyInBals]
		bal[payerIdx].Sub(bal[payerIdx], parsedAmount)
		bal[payeeIdx].Add(bal[payeeIdx], parsedAmount)
	}
}

// GetPayChInfo fetches the channel info for this channel and interprets it as
// payment channel info.
func GetPayChInfo(ch perun.ChAPI) PayChInfo {
	return toPayChInfo(ch.GetChInfo())
}

// SubPayChUpdates sets up a subscription for incoming channel updates and
// interprets the notifications as payment update notifiations.
//
// See session.SubChUpdates for the list of errors returned by this API.
func SubPayChUpdates(ch perun.ChAPI, notifier PayChUpdateNotifier) perun.APIError {
	return ch.SubChUpdates(func(notif perun.ChUpdateNotif) {
		var ProposedPayChInfo PayChInfo
		if notif.Type == perun.ChUpdateTypeClosed {
			ProposedPayChInfo = toPayChInfo(notif.CurrChInfo)
		} else {
			ProposedPayChInfo = toPayChInfo(notif.ProposedChInfo)
		}
		notifier(PayChUpdateNotif{
			UpdateID:          notif.UpdateID,
			ProposedPayChInfo: ProposedPayChInfo,
			Type:              notif.Type,
			Expiry:            notif.Expiry,
			Error:             notif.Error,
		})
	})
}

// UnsubPayChUpdates deletes the existing subscription for updates on this channel.
//
// See session.UnsubChUpdates for the list of errors returned by this API.
func UnsubPayChUpdates(ch perun.ChAPI) perun.APIError {
	return ch.UnsubChUpdates()
}

// RespondPayChUpdate sends a response for a channel update notification and
// interprets the updated channel info as payment channel info.
//
// See session.RespondChUpdate for the list of errors returned by this API.
func RespondPayChUpdate(pctx context.Context, ch perun.ChAPI, updateID string, accept bool) (
	PayChInfo, perun.APIError) {
	chInfo, err := ch.RespondChUpdate(pctx, updateID, accept)
	return toPayChInfo(chInfo), err
}

// ClosePayCh closes the channel and interprets the closing channel info as
// payment channel info.
//
// See session.CloseCh for the list of errors returned by this API.
func ClosePayCh(pctx context.Context, ch perun.ChAPI) (PayChInfo, perun.APIError) {
	chInfo, err := ch.Close(pctx)
	return toPayChInfo(chInfo), err
}

// toPaysChInfo converts ChInfo to PayChInfo.
func toPayChsInfo(chsInfo []perun.ChInfo) []PayChInfo {
	payChsInfo := make([]PayChInfo, len(chsInfo))
	for i := range chsInfo {
		payChsInfo[i] = toPayChInfo(chsInfo[i])
	}
	return payChsInfo
}

// toPayChInfo converts ChInfo to PayChInfo.
func toPayChInfo(chInfo perun.ChInfo) PayChInfo {
	return PayChInfo{
		ChID:    chInfo.ChID,
		BalInfo: chInfo.BalInfo,
		Version: chInfo.Version,
	}
}
