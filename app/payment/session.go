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

// OpenSession opens a session and interprets the restored channels info as
// payment channels info.
//
// See node.OpenSession for the list of errors returned by this API.
func OpenSession(n perun.NodeAPI, configFile string) (string, []PayChInfo, perun.APIError) {
	sessionID, restoredChsInfo, err := n.OpenSession(configFile)
	return sessionID, toPayChsInfo(restoredChsInfo), err
}

// OpenPayCh opens a channel with payment app with the specified parameters. It
// interprets the returned channel info as payment channel info.
//
// See session.OpenCh for the list of errors returned by this API.
func OpenPayCh(pctx context.Context, s perun.SessionAPI, openingBalInfo perun.BalInfo, challengeDurSecs uint64) (
	PayChInfo, perun.APIError) {
	paymentApp := perun.App{
		Def:  pchannel.NoApp(),
		Data: pchannel.NoData(),
	}

	chInfo, err := s.OpenCh(pctx, openingBalInfo, paymentApp, challengeDurSecs)
	return toPayChInfo(chInfo), err
}

// GetPayChsInfo fetches the list of all channels info in the session and
// interprets them as payment channel info.
func GetPayChsInfo(s perun.SessionAPI) []PayChInfo {
	chsInfo := s.GetChsInfo()

	payChsInfo := make([]PayChInfo, len(chsInfo))
	for i := range chsInfo {
		payChsInfo[i] = toPayChInfo(chsInfo[i])
	}
	return payChsInfo
}

// SubPayChProposals sets up a subscription for incoming channel proposals and
// interprets the notifications as payment channel notifiations.
//
// See session.SubChProposals for the list of errors returned by this API.
func SubPayChProposals(s perun.SessionAPI, notifier PayChProposalNotifier) perun.APIError {
	return s.SubChProposals(func(notif perun.ChProposalNotif) {
		notifier(PayChProposalNotif{
			ProposalID:       notif.ProposalID,
			OpeningBalInfo:   notif.OpeningBalInfo,
			ChallengeDurSecs: notif.ChallengeDurSecs,
			Expiry:           notif.Expiry,
		})
	})
}

// UnsubPayChProposals deletes the existing subscription for channel proposals.
//
// See session.UnsubChProposals for the list of errors returned by this API.
func UnsubPayChProposals(s perun.SessionAPI) perun.APIError {
	return s.UnsubChProposals()
}

// RespondPayChProposal sends the response to a payment channel proposal
// notification and interprets the opening channel info as payment channel info.
//
// See session.RespondChProposal for the list of errors returned by this API.
func RespondPayChProposal(pctx context.Context, s perun.SessionAPI, proposalID string, accept bool) (PayChInfo,
	perun.APIError) {
	chInfo, apiErr := s.RespondChProposal(pctx, proposalID, accept)
	return toPayChInfo(chInfo), apiErr
}

// ErrInfoFailedPreCondUnclosedPayChs is the interpretation of
// ErrInfoFailedPreCondUnclosedChs for payment application.
type ErrInfoFailedPreCondUnclosedPayChs struct {
	PayChs []PayChInfo
}

// CloseSession closes the current session.
//
// See session.CloseSession for the list of errors returned by this API.
func CloseSession(s perun.SessionAPI, force bool) ([]PayChInfo, perun.APIError) {
	openChsInfo, err := s.Close(force)
	err = toPayChsCloseSessionErr(err)
	return toPayChsInfo(openChsInfo), err
}

func toPayChsCloseSessionErr(err perun.APIError) perun.APIError {
	if err == nil {
		return err
	}
	addInfo, ok := err.AddInfo().(perun.ErrInfoFailedPreCondUnclosedChs)
	if !ok {
		return err
	}
	paymentAddInfo := ErrInfoFailedPreCondUnclosedPayChs{
		PayChs: toPayChsInfo(addInfo.ChInfos),
	}
	return perun.NewAPIErr(err.Category(), err.Code(), err, paymentAddInfo)
}
