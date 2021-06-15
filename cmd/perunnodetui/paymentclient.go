// Copyright (c) 2021 - for information on the respective copyright owner
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

package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
	grpclib "google.golang.org/grpc"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/currency"
)

var errNotConnectedToNode = fmt.Errorf("not connected to perun node")

func connectToNode(perunNodeURL, configFileURL string) (string, pb.Payment_APIClient, error) {
	conn, err := grpclib.Dial(perunNodeURL, grpclib.WithInsecure())
	if err != nil {
		return "", nil, errors.Wrap(err, "connecting to perun node")
	}
	apiClient := pb.NewPayment_APIClient(conn)

	req := pb.OpenSessionReq{
		ConfigFile: configFileURL,
	}
	resp, err := apiClient.OpenSession(context.Background(), &req)
	if err != nil {
		return "", nil, errors.Wrap(err, "opening session")
	}
	msgErr, ok := resp.Response.(*pb.OpenSessionResp_Error)
	if ok {
		return "", nil, errors.New(printAPIError(msgErr.Error))
	}

	// If not error, it must be success response.So it's safe to type assert.
	sessID := resp.Response.(*pb.OpenSessionResp_MsgSuccess_).MsgSuccess.SessionID
	return sessID, apiClient, nil
}

func newOutgoingChannelProposalFn(peer, ours, theirs string) proposerFn {
	return func() (bool, string, balInfo, error) {
		if client == nil {
			return false, "", balInfo{}, errors.WithStack(errNotConnectedToNode)
		}

		req := pb.OpenPayChReq{
			SessionID: sessionID,
			OpeningBalInfo: &pb.BalInfo{
				Currency: currency.ETH,
				Parts:    []string{perun.OwnAlias, peer},
				Bal:      []string{ours, theirs},
			},
			ChallengeDurSecs: challengeDurSecs,
		}
		resp, err := client.OpenPayCh(context.Background(), &req)
		if err != nil {
			return false, "", balInfo{}, errors.Wrap(err, "sending open request")
		}
		msgErr, ok := resp.Response.(*pb.OpenPayChResp_Error)
		if ok {
			if msgErr.Error.Code == pb.ErrorCode_ErrPeerRejected {
				return false, "", balInfo{}, nil
			}
			return false, "", balInfo{}, fmt.Errorf("open request failed: %v", printAPIError(msgErr.Error))
		}

		payChInfo := resp.Response.(*pb.OpenPayChResp_MsgSuccess_).MsgSuccess.OpenedPayChInfo
		return true, payChInfo.ChID, grpcPayChInfotoBalInfo(payChInfo), nil
	}
}

func subProposals() error {
	if client == nil {
		return errors.WithStack(errNotConnectedToNode)
	}
	req := pb.SubPayChProposalsReq{
		SessionID: sessionID,
	}
	sub, err := client.SubPayChProposals(context.Background(), &req)
	if err != nil {
		return errors.Wrap(err, "subscribing to proposals")
	}
	go incomingChannelsHandler(sub)
	return nil
}

func incomingChannelsHandler(sub pb.Payment_API_SubPayChProposalsClient) {
	for {
		notifMsg, err := sub.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				logInfo("Subscription to incoming proposal closed")
			} else {
				logErrorf("Subscription to incoming proposal closed with error %v", err)
			}
			return
		}

		msgErr, ok := notifMsg.Response.(*pb.SubPayChProposalsResp_Error)
		if ok {
			logErrorf("In incoming proposal notification %v", printAPIError(msgErr.Error))
			return
		}

		notif := notifMsg.Response.(*pb.SubPayChProposalsResp_Notify_).Notify
		updateID := notif.ProposalID
		peer := notif.OpeningBalInfo.Parts[findPeerIndex(notif.OpeningBalInfo.Parts)]
		proposed := grpcBalInfotoBalInfo(notif.OpeningBalInfo)
		timeout := time.Unix(notif.Expiry, 0)

		p, err := newIncomingChannel(updateID, peer, proposed.ours, proposed.theirs, timeout.Format("15:04:05"))
		if err != nil {
			logError(errors.WithMessage(err, "adding incoming channel to table"))
		} else {
			R.putAtIndex(p.sNo, p)
			go p.notifyIncomingChannel(time.Until(timeout))
		}
	}
}

func respondToProposal(proposalID string, accept bool) (string, balInfo, error) {
	if client == nil {
		return "", balInfo{}, errors.WithStack(errNotConnectedToNode)
	}

	req := pb.RespondPayChProposalReq{
		SessionID:  sessionID,
		ProposalID: proposalID,
		Accept:     accept,
	}
	resp, err := client.RespondPayChProposal(context.Background(), &req)
	if err != nil {
		return "", balInfo{}, errors.Wrap(err, "responding to proposal")
	}
	msgErr, ok := resp.Response.(*pb.RespondPayChProposalResp_Error)
	if ok {
		return "", balInfo{}, errors.New("Error responding to proposal: " + printAPIError(msgErr.Error))
	}

	if !accept {
		return "", balInfo{}, nil // No channel info is available when a proposal is rejected.
	}
	payChInfo := resp.Response.(*pb.RespondPayChProposalResp_MsgSuccess_).MsgSuccess.OpenedPayChInfo
	return payChInfo.ChID, grpcPayChInfotoBalInfo(payChInfo), nil
}

func subUpdates(chID string) error {
	if client == nil {
		return errors.WithStack(errNotConnectedToNode)
	}
	req := pb.SubpayChUpdatesReq{
		SessionID: sessionID,
		ChID:      chID,
	}
	sub, err := client.SubPayChUpdates(context.Background(), &req)
	if err != nil {
		return errors.Wrap(err, "subscribing to updates")
	}
	go incomingUpdateHandler(sub)
	return nil
}

func incomingUpdateHandler(sub pb.Payment_API_SubPayChUpdatesClient) {
	for {
		notifMsg, err := sub.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				logInfo("Subscription to incoming updates closed")
			} else {
				logErrorf("Subscription to incoming updates closed with error %v", err)
			}
			return
		}
		msgErr, ok := notifMsg.Response.(*pb.SubPayChUpdatesResp_Error)
		if ok {
			logErrorf("In incoming update notification %v", printAPIError(msgErr.Error))
			return
		}

		notif := notifMsg.Response.(*pb.SubPayChUpdatesResp_Notify_).Notify
		p := R.getByChID(notif.ProposedPayChInfo.ChID)
		if p == nil {
			logError("update received for unknown channel id")
			return
		}

		if notif.Type == pb.SubPayChUpdatesResp_Notify_closed {
			errMsg := ""
			if notif.Error != nil {
				errMsg = printAPIError(notif.Error)
			}
			p.notifyClosingUpdate(grpcPayChInfotoBalInfo(notif.ProposedPayChInfo), errMsg)
			return
		}

		proposed := grpcPayChInfotoBalInfo(notif.ProposedPayChInfo)
		go p.notifyNonClosingUpdate(notif.UpdateID, proposed, notif.Expiry, notif.Type == pb.SubPayChUpdatesResp_Notify_final)
	}
}

func respondToUpdate(chID, updateID string, accept bool) (balInfo, error) {
	if client == nil {
		return balInfo{}, errors.WithStack(errNotConnectedToNode)
	}

	req := pb.RespondPayChUpdateReq{
		SessionID: sessionID,
		ChID:      chID,
		UpdateID:  updateID,
		Accept:    accept,
	}
	resp, err := client.RespondPayChUpdate(context.Background(), &req)
	if err != nil {
		return balInfo{}, errors.Wrap(err, "responding to update")
	}
	msgErr, ok := resp.Response.(*pb.RespondPayChUpdateResp_Error)
	if ok {
		return balInfo{}, errors.New("responding to update: " + printAPIError(msgErr.Error))
	}

	// If it is not error, it must be success response.
	payChInfo := resp.Response.(*pb.RespondPayChUpdateResp_MsgSuccess_).MsgSuccess.UpdatedPayChInfo
	return grpcPayChInfotoBalInfo(payChInfo), nil
}

func sendUpdate(chID, peer, amount string) (bool, balInfo, error) {
	if client == nil {
		return false, balInfo{}, errors.WithStack(errNotConnectedToNode)
	}

	req := pb.SendPayChUpdateReq{
		SessionID: sessionID,
		ChID:      chID,
		Payee:     peer,
		Amount:    amount,
	}
	resp, err := client.SendPayChUpdate(context.Background(), &req)
	if err != nil {
		return false, balInfo{}, errors.Wrap(err, "sending channel update")
	}
	msgErr, ok := resp.Response.(*pb.SendPayChUpdateResp_Error)
	if ok {
		if msgErr.Error.Code == pb.ErrorCode_ErrPeerRejected ||
			msgErr.Error.Code == pb.ErrorCode_ErrPeerRequestTimedOut {
			return false, balInfo{}, nil
		}
		return false, balInfo{}, fmt.Errorf("send update request failed: %v", printAPIError(msgErr.Error))
	}

	payChInfo := resp.Response.(*pb.SendPayChUpdateResp_MsgSuccess_).MsgSuccess.UpdatedPayChInfo
	return true, grpcPayChInfotoBalInfo(payChInfo), nil
}

func closeCh(chID string) (balInfo, error) {
	if client == nil {
		return balInfo{}, errors.WithStack(errNotConnectedToNode)
	}

	req := pb.ClosePayChReq{
		SessionID: sessionID,
		ChID:      chID,
	}
	resp, err := client.ClosePayCh(context.Background(), &req)
	if err != nil {
		return balInfo{}, errors.Wrap(err, "sending channel update")
	}
	msgErr, ok := resp.Response.(*pb.ClosePayChResp_Error)
	if ok {
		return balInfo{}, errors.New("close request failed: " + printAPIError(msgErr.Error))
	}

	msg := resp.Response.(*pb.ClosePayChResp_MsgSuccess_)
	return grpcPayChInfotoBalInfo(msg.MsgSuccess.ClosedPayChInfo), nil
}

// printAPIError formats the error message returned by the API.
func printAPIError(e *pb.MsgError) string {
	return e.Message
	// return fmt.Sprintf("category: %s, code: %d, message: %s, additional info: %+v",
	// e.Category, e.Code, e.Message, e.AddInfo)
}
