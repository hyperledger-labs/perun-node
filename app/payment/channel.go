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

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/currency"
)

type (
	// PayChInfo represents the interpretation of channelInfo for payment app.
	PayChInfo struct {
		ChID    string
		BalInfo perun.BalInfo
		Version string
	}
	// PayChUpdateNotifier represents the channel update notification function for payment app.
	PayChUpdateNotifier func(PayChUpdateNotif)

	// PayChUpdateNotif represents the channel update notification data for payment app.
	PayChUpdateNotif struct {
		UpdateID          string
		ProposedPayChInfo PayChInfo
		IsFinal           bool
		Expiry            int64
	}
)

// SendPayChUpdate send the given amount to the payee. Payee should be one of the channel participants.
// Use "self" to request payments.
func SendPayChUpdate(pctx context.Context, ch perun.ChAPI, payee, amount string) (PayChInfo, error) {
	parsedAmount, err := parseAmount(ch.Currency(), amount)
	if err != nil {
		return PayChInfo{}, err
	}
	payerIdx, payeeIdx, err := getPayerPayeeIdx(ch.Parts(), payee)
	if err != nil {
		return PayChInfo{}, err
	}
	chInfo, err := ch.SendChUpdate(pctx, newUpdate(payerIdx, payeeIdx, parsedAmount))
	return ToPayChInfo(chInfo), err
}

func parseAmount(chCurrency string, amount string) (*big.Int, error) {
	parsedAmount, err := currency.NewParser(chCurrency).Parse(amount)
	if err != nil {
		return nil, perun.ErrInvalidAmount
	}
	return parsedAmount, nil
}

func getPayerPayeeIdx(parts []string, payee string) (payerIdx, payeeIdx int, _ error) {
	for i := range parts {
		if parts[i] == payee {
			payeeIdx = i
			payerIdx = payeeIdx ^ 1
			return payerIdx, payeeIdx, nil
		}
	}
	return 0, 0, perun.ErrInvalidPayee
}

func newUpdate(payerIdx, payeeIdx int, parsedAmount *big.Int) perun.StateUpdater {
	return func(state *pchannel.State) {
		bal := state.Allocation.Balances[0]
		bal[payerIdx].Sub(bal[payerIdx], parsedAmount)
		bal[payeeIdx].Add(bal[payeeIdx], parsedAmount)
	}
}

// GetPayChInfo returns the balance information for this channel.
func GetPayChInfo(ch perun.ChAPI) PayChInfo {
	return ToPayChInfo(ch.GetChInfo())
}

// SubPayChUpdates sets up a subscription for updates on this channel.
func SubPayChUpdates(ch perun.ChAPI, notifier PayChUpdateNotifier) error {
	return ch.SubChUpdates(func(notif perun.ChUpdateNotif) {
		notifier(PayChUpdateNotif{
			UpdateID:          notif.UpdateID,
			ProposedPayChInfo: ToPayChInfo(notif.ProposedChInfo),
			IsFinal:           notif.ProposedChInfo.IsFinal,
			Expiry:            notif.Expiry,
		})
	})
}

// UnsubPayChUpdates deletes the existing subscription for updates on this channel.
func UnsubPayChUpdates(ch perun.ChAPI) error {
	return ch.UnsubChUpdates()
}

// RespondPayChUpdate sends a response for a channel update notification.
func RespondPayChUpdate(pctx context.Context, ch perun.ChAPI, updateID string, accept bool) (PayChInfo, error) {
	chInfo, err := ch.RespondChUpdate(pctx, updateID, accept)
	return ToPayChInfo(chInfo), err
}

// ClosePayCh closes the payment channel.
func ClosePayCh(pctx context.Context, ch perun.ChAPI) (PayChInfo, error) {
	chInfo, err := ch.Close(pctx)
	return ToPayChInfo(chInfo), err
}

// ToPayChInfo converts ChInfo to PayChInfo.
func ToPayChInfo(chInfo perun.ChInfo) PayChInfo {
	return PayChInfo{
		ChID:    chInfo.ChID,
		BalInfo: chInfo.BalInfo,
		Version: chInfo.Version,
	}
} // nolint:gofumpt // unknown error, maybe a false positive
