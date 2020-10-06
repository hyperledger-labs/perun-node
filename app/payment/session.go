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
	"fmt"
	"math/big"

	ppayment "perun.network/go-perun/apps/payment"
	pchannel "perun.network/go-perun/channel"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/currency"
)

type (
	// PayChProposalNotif represents the channel update notification data for payment app.
	PayChProposalNotif struct {
		ProposalID       string
		Currency         string
		OpeningBalInfo   perun.BalInfo
		ChallengeDurSecs uint64
		Expiry           int64
	}

	// PayChProposalNotifier represents the channel update notification function for payment app.
	PayChProposalNotifier func(PayChProposalNotif)

	// PayChCloseNotif represents the channel close notification data for payment app.
	PayChCloseNotif struct {
		ClosedPayChInfo PayChInfo
		Error           string
	}

	// PayChCloseNotifier represents the channel close notification function for payment app.
	PayChCloseNotifier func(PayChCloseNotif)
)

// init() initializes the payment app in go-perun.
func init() {
	wb := ethereum.NewWalletBackend()
	emptyAddr, err := wb.ParseAddr("0x0")
	if err != nil {
		panic("Error parsing zero address for app payment def: " + err.Error())
	}
	ppayment.SetAppDef(emptyAddr) // dummy app def.
}

// OpenPayCh opens a payment channel using the given sessionAPI instance with the specified parameters.
func OpenPayCh(pctx context.Context,
	s perun.SessionAPI,
	peerAlias string,
	openingBalInfo perun.BalInfo,
	challengeDurSecs uint64) (PayChInfo, error) {
	paymentApp := perun.App{
		Def:  ppayment.NewApp(),
		Data: pchannel.NoData(),
	}

	chInfo, err := s.OpenCh(pctx, peerAlias, openingBalInfo, paymentApp, challengeDurSecs)
	if err != nil {
		return PayChInfo{}, err
	}
	return PayChInfo{
		ChID:    chInfo.ChID,
		BalInfo: balInfoFromState(chInfo.Currency, chInfo.State, chInfo.Parts),
		Version: fmt.Sprintf("%d", chInfo.State.Version),
	}, nil
}

// GetPayChsInfo returns a list of payment channel info for all the channels in this session.
func GetPayChsInfo(s perun.SessionAPI) []PayChInfo {
	chsInfo := s.GetChsInfo()

	openPayChsInfo := make([]PayChInfo, len(chsInfo))
	for i := range chsInfo {
		openPayChsInfo[i] = PayChInfo{
			ChID:    chsInfo[i].ChID,
			BalInfo: balInfoFromState(chsInfo[i].Currency, chsInfo[i].State, chsInfo[i].Parts),
			Version: fmt.Sprintf("%d", chsInfo[i].State.Version),
		}
	}
	return openPayChsInfo
}

// SubPayChProposals sets up a subscription for payment channel proposals.
func SubPayChProposals(s perun.SessionAPI, notifier PayChProposalNotifier) error {
	return s.SubChProposals(func(notif perun.ChProposalNotif) {
		balsBigInt := notif.ChProposal.Proposal().InitBals.Balances[0]
		notifier(PayChProposalNotif{
			ProposalID:       notif.ProposalID,
			Currency:         notif.Currency,
			OpeningBalInfo:   balInfoFromRawBal("ETH", balsBigInt, notif.Parts),
			ChallengeDurSecs: notif.ChProposal.Proposal().ChallengeDuration,
			Expiry:           notif.Expiry,
		})
	})
}

// UnsubPayChProposals deletes the existing subscription for payment channel proposals.
func UnsubPayChProposals(s perun.SessionAPI) error {
	return s.UnsubChProposals()
}

// RespondPayChProposal sends the response to a payment channel proposal notification.
func RespondPayChProposal(pctx context.Context, s perun.SessionAPI, proposalID string, accept bool) error {
	return s.RespondChProposal(pctx, proposalID, accept)
}

// SubPayChCloses sets up a subscription for payment channel closes.
func SubPayChCloses(s perun.SessionAPI, notifier PayChCloseNotifier) error {
	return s.SubChCloses(func(notif perun.ChCloseNotif) {
		notifier(PayChCloseNotif{
			ClosedPayChInfo: PayChInfo{
				ChID:    notif.ChID,
				BalInfo: balInfoFromState(notif.Currency, notif.ChState, notif.Parts),
				Version: fmt.Sprintf("%d", notif.ChState.Version),
			},
		})
	})
}

// UnsubPayChCloses deletes the existing subscription for payment channel closes.
func UnsubPayChCloses(s perun.SessionAPI) error {
	return s.UnsubChCloses()
}

func balInfoFromState(currency string, state *pchannel.State, parts []string) perun.BalInfo {
	return balInfoFromRawBal(currency, state.Balances[0], parts)
}

func balInfoFromRawBal(chCurrency string, bigInt []*big.Int, parts []string) perun.BalInfo {
	balInfo := perun.BalInfo{
		Currency: chCurrency,
		Bals:     make(map[string]string, len(parts)),
	}

	parser := currency.NewParser(chCurrency)
	for i := range parts {
		balInfo.Bals[parts[i]] = parser.Print(bigInt[i])
		balInfo.Bals[parts[i]] = parser.Print(bigInt[i])
	}
	return balInfo
}
