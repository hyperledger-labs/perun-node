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

	ppayment "perun.network/go-perun/apps/payment"
	pchannel "perun.network/go-perun/channel"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
)

type (
	// PayChProposalNotif represents the channel update notification data for payment app.
	PayChProposalNotif struct {
		ProposalID       string
		OpeningBalInfo   perun.BalInfo
		ChallengeDurSecs uint64
		Expiry           int64
	}

	// PayChProposalNotifier represents the channel update notification function for payment app.
	PayChProposalNotifier func(PayChProposalNotif)
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
func OpenPayCh(pctx context.Context, s perun.SessionAPI, openingBalInfo perun.BalInfo, challengeDurSecs uint64) (
	PayChInfo, error) {
	paymentApp := perun.App{
		Def:  ppayment.NewApp(),
		Data: pchannel.NoData(),
	}

	chInfo, err := s.OpenCh(pctx, openingBalInfo, paymentApp, challengeDurSecs)
	return ToPayChInfo(chInfo), err
}

// GetPayChsInfo returns a list of payment channel info for all the channels in this session.
func GetPayChsInfo(s perun.SessionAPI) []PayChInfo {
	chsInfo := s.GetChsInfo()

	payChsInfo := make([]PayChInfo, len(chsInfo))
	for i := range chsInfo {
		payChsInfo[i] = ToPayChInfo(chsInfo[i])
	}
	return payChsInfo
}

// SubPayChProposals sets up a subscription for payment channel proposals.
func SubPayChProposals(s perun.SessionAPI, notifier PayChProposalNotifier) error {
	return s.SubChProposals(func(notif perun.ChProposalNotif) {
		notifier(PayChProposalNotif{
			ProposalID:       notif.ProposalID,
			OpeningBalInfo:   notif.OpeningBalInfo,
			ChallengeDurSecs: notif.ChallengeDurSecs,
			Expiry:           notif.Expiry,
		})
	})
}

// UnsubPayChProposals deletes the existing subscription for payment channel proposals.
func UnsubPayChProposals(s perun.SessionAPI) error {
	return s.UnsubChProposals()
}

// RespondPayChProposal sends the response to a payment channel proposal notification.
func RespondPayChProposal(pctx context.Context, s perun.SessionAPI, proposalID string, accept bool) (PayChInfo, error) {
	chInfo, err := s.RespondChProposal(pctx, proposalID, accept)
	return ToPayChInfo(chInfo), err
}
