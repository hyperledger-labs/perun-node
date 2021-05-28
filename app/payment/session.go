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

	pchannel "perun.network/go-perun/channel"

	"github.com/hyperledger-labs/perun-node"
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

// OpenSession opens a session and interprets the restored channels as payment channels.
func OpenSession(n perun.NodeAPI, configFile string) (string, []PayChInfo, perun.APIErrorV2) {
	sessionID, restoredChsInfo, err := n.OpenSession(configFile)
	return sessionID, toPayChsInfo(restoredChsInfo), err
}

// OpenPayCh opens a payment channel using the given sessionAPI instance with the specified parameters.
func OpenPayCh(pctx context.Context, s perun.SessionAPI, openingBalInfo perun.BalInfo, challengeDurSecs uint64) (
	PayChInfo, perun.APIErrorV2) {
	paymentApp := perun.App{
		Def:  pchannel.NoApp(),
		Data: pchannel.NoData(),
	}

	chInfo, err := s.OpenCh(pctx, openingBalInfo, paymentApp, challengeDurSecs)
	return toPayChInfo(chInfo), err
}

// GetPayChsInfo returns a list of payment channel info for all the channels in this session.
func GetPayChsInfo(s perun.SessionAPI) []PayChInfo {
	chsInfo := s.GetChsInfo()

	payChsInfo := make([]PayChInfo, len(chsInfo))
	for i := range chsInfo {
		payChsInfo[i] = toPayChInfo(chsInfo[i])
	}
	return payChsInfo
}

// SubPayChProposals sets up a subscription for payment channel proposals.
func SubPayChProposals(s perun.SessionAPI, notifier PayChProposalNotifier) perun.APIErrorV2 {
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
func UnsubPayChProposals(s perun.SessionAPI) perun.APIErrorV2 {
	return s.UnsubChProposals()
}

// RespondPayChProposal sends the response to a payment channel proposal notification.
func RespondPayChProposal(pctx context.Context, s perun.SessionAPI, proposalID string, accept bool) (PayChInfo,
	perun.APIErrorV2) {
	chInfo, apiErr := s.RespondChProposal(pctx, proposalID, accept)
	return toPayChInfo(chInfo), apiErr
}

// ErrV2InfoFailedPreCondUnclosedPayChs is the interpretation of
// ErrV2InfoFailedPreCondUnclosedChs for payment application.
type ErrV2InfoFailedPreCondUnclosedPayChs struct {
	PayChs []PayChInfo
}

// CloseSession closes the current session.
func CloseSession(s perun.SessionAPI, force bool) ([]PayChInfo, perun.APIErrorV2) {
	openChsInfo, err := s.Close(force)
	err = toPayChsCloseSessionErr(err)
	return toPayChsInfo(openChsInfo), err
}

func toPayChsCloseSessionErr(err perun.APIErrorV2) perun.APIErrorV2 {
	if err == nil {
		return err
	}
	addInfo, ok := err.AddInfo().(perun.ErrV2InfoFailedPreCondUnclosedChs)
	if !ok {
		return err
	}
	paymentAddInfo := ErrV2InfoFailedPreCondUnclosedPayChs{
		PayChs: toPayChsInfo(addInfo.ChInfos),
	}
	return perun.NewAPIErrV2(err.Category(), err.Code(), err.Message(), paymentAddInfo)
}
