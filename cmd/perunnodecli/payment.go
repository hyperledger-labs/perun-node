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

package main

import (
	"context"
	"errors"
	"io"

	"github.com/abiosoft/ishell"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

var (
	paymentCmdUsage = "Usage: payment [sub-command]"
	paymentCmd      = &ishell.Cmd{
		Name: "payment",
		Help: "Use this command to send/receive payments." + paymentCmdUsage,
		Func: paymentFn,
	}
	paymentSendCmdUsage = "Usage: payment send [channel alias] [amount]"
	paymentSendCmd      = &ishell.Cmd{
		Name: "send-to-peer",
		Help: "Send a payment to the peer." + paymentSendCmdUsage,
		Completer: func([]string) []string {
			return openChannelsList
		},
		Func: paymentSendFn,
	}
	paymentRequestCmdUsage = "Usage: payment request [channel alias] [amount]"
	paymentRequestCmd      = &ishell.Cmd{
		Name: "request-from-peer",
		Help: "Request a payment from the peer." + paymentRequestCmdUsage,
		Completer: func([]string) []string {
			return openChannelsList
		},
		Func: paymentRequestFn,
	}
	paymentSubCmdUsage = "Usage: payment sub [channel alias]"
	paymentSubCmd      = &ishell.Cmd{
		Name: "subscribe-to-notifications",
		Help: "Subscribe to payment notifications from the peer." + paymentSubCmdUsage,
		Completer: func([]string) []string {
			return openChannelsList
		},
		Func: paymentSubFn,
	}
	paymentUnsubCmdUsage = "Usage: payment unsub [channel alias]"
	paymentUnsubCmd      = &ishell.Cmd{
		Name: "unsubscribe-to-notifications",
		Help: "Unsubscribe from payment notifications from the peer." + paymentUnsubCmdUsage,
		Completer: func([]string) []string {
			return openChannelsList
		},
		Func: paymentUnsubFn,
	}
	paymentAcceptCmdUsage = "Usage: payment accept [channel alias]"
	paymentAcceptCmd      = &ishell.Cmd{
		Name: "accept-payment-update-from-peer",
		Help: "Accept latest payment update from peer on the channel." + paymentAcceptCmdUsage,
		Completer: func([]string) []string {
			return openChannelsList
		},
		Func: paymentAccept,
	}
	paymentRejectCmdUsage = "Usage: payment reject [channel alias]"
	paymentRejectCmd      = &ishell.Cmd{
		Name: "reject-payment-update-from-peer",
		Help: "Reject latest payment update from peer on the channel." + paymentRejectCmdUsage,
		Completer: func([]string) []string {
			return openChannelsList
		},
		Func: paymentReject,
	}
)

func init() {
	paymentCmd.AddCmd(paymentSendCmd)
	paymentCmd.AddCmd(paymentRequestCmd)
	paymentCmd.AddCmd(paymentSubCmd)
	paymentCmd.AddCmd(paymentUnsubCmd)
	paymentCmd.AddCmd(paymentAcceptCmd)
	paymentCmd.AddCmd(paymentRejectCmd)
}

func paymentFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	c.Println(c.Cmd.HelpText())
}

// nolint: dupl		// not a duplicate of paymentRequestFn
func paymentSendFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}

	// Usage: payment send/request [channel alias] [amount]
	countReqArgs := 2
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}
	chInfo, ok := openChannelsMap[c.Args[0]]
	if !ok {
		printUnknownChannelAliasError(c, c.Args[0])
		return
	}

	req := pb.SendPayChUpdateReq{
		SessionID: sessionID,
		ChID:      chInfo.id,
		Payee:     chInfo.peerAlias,
		Amount:    c.Args[1],
	}
	resp, err := client.SendPayChUpdate(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.SendPayChUpdateResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error sending payment to peer: %v.", apiErrorString(msgErr.Error)))
		return
	}

	msg := resp.Response.(*pb.SendPayChUpdateResp_MsgSuccess_)
	chAlias := openChannelsRevMap[msg.MsgSuccess.UpdatedPayChInfo.ChID]
	c.Printf("%s\n\n", greenf("Payment sent to peer on channel %s. Updated channel Info:\n%s.",
		chAlias, prettifyPayChInfo(msg.MsgSuccess.UpdatedPayChInfo)))
}

// nolint: dupl		// not a duplicate of paymentSendFn
func paymentRequestFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}

	// Usage: payment send/request [channel alias] [amount]
	countReqArgs := 2
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}
	chInfo, ok := openChannelsMap[c.Args[0]]
	if !ok {
		printUnknownChannelAliasError(c, c.Args[0])
		return
	}

	req := pb.SendPayChUpdateReq{
		SessionID: sessionID,
		ChID:      chInfo.id,
		Payee:     perun.OwnAlias,
		Amount:    c.Args[1],
	}
	resp, err := client.SendPayChUpdate(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.SendPayChUpdateResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error requesting payment from peer: %v.", apiErrorString(msgErr.Error)))
		return
	}

	msg := resp.Response.(*pb.SendPayChUpdateResp_MsgSuccess_)
	chAlias := openChannelsRevMap[msg.MsgSuccess.UpdatedPayChInfo.ChID]
	c.Printf("%s\n\n", greenf("Payment request accepted by peer on channel %s. Updated channel Info:\n%s.",
		chAlias, prettifyPayChInfo(msg.MsgSuccess.UpdatedPayChInfo)))
}

func paymentSubFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	// Usage: payment sub [channel alias]
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	paymentSub(c, c.Args[0])
}

func paymentSub(c *ishell.Context, chAlias string) {
	chInfo, ok := openChannelsMap[chAlias]
	if !ok {
		printUnknownChannelAliasError(c, chAlias)
		return
	}
	req := pb.SubpayChUpdatesReq{
		SessionID: sessionID,
		ChID:      chInfo.id,
	}
	sub, err := client.SubPayChUpdates(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	go paymentNotifHandler(sub, chAlias, chInfo.id)
	c.Printf("%s\n\n", greenf("Subscribed to payment notifications on channel %s (ID: %s)", chAlias, chInfo.id))
}

func paymentNotifHandler(sub pb.Payment_API_SubPayChUpdatesClient, chAlias, chID string) {
	for {
		notifMsg, err := sub.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				sh.Printf("%s\n\n",
					redf("Subscription to payment channel notifications closed on channel %s (ID: %s).", chAlias, chID))
			} else {
				sh.Printf("%s\n\n", redf("Error receiving channel update notification: %v", err))
			}
			return
		}
		msgErr, ok := notifMsg.Response.(*pb.SubPayChUpdatesResp_Error)
		if ok {
			sh.Printf("%s\n\n", redf("Error message received in update notification : %v", apiErrorString(msgErr.Error)))
			return
		}
		notif := notifMsg.Response.(*pb.SubPayChUpdatesResp_Notify_)

		nodeTime, err := getNodeTime()
		if err != nil {
			printCommandSendingError(sh, err)
			return
		}
		switch {
		case notif.Notify.Type == pb.SubPayChUpdatesResp_Notify_open:
			handleUpdateTypeOpenNotif(chAlias, notif, nodeTime)
		case notif.Notify.Type == pb.SubPayChUpdatesResp_Notify_final:
			handleUpdateTypeFinalNotif(chAlias, notif, nodeTime)
		case notif.Notify.Type == pb.SubPayChUpdatesResp_Notify_closed:
			handleUpdateTypeClosedNotif(chAlias, notif)
		}

	}
}

func handleUpdateTypeOpenNotif(chAlias string, notif *pb.SubPayChUpdatesResp_Notify_, nodeTime int64) {
	currPayChInfo := getChannelInfo(sh, chAlias)

	sh.Printf("%s", greenf("Payment notification received on channel %s. (ID:%s)\n",
		chAlias, notif.Notify.ProposedPayChInfo.ChID))
	sh.Printf("%s\n\n", greenf("Current:\t%s.\nProposed:\t%s.\nExpires in %ds.",
		prettifyBalanceInfo(currPayChInfo.BalInfo, currPayChInfo.Version),
		prettifyBalanceInfo(notif.Notify.ProposedPayChInfo.BalInfo, notif.Notify.ProposedPayChInfo.Version),
		notif.Notify.Expiry-nodeTime))

	if openChannelsMap[chAlias].latestPaymentNotifID != "" {
		sh.Printf("%s", greenf("Dropping older (now obselete) notification received on this channel.",
			openChannelsMap[chAlias].latestPaymentNotifID))
	}
	openChannelsMap[chAlias].latestPaymentNotifID = notif.Notify.UpdateID
}

func handleUpdateTypeFinalNotif(chAlias string, notif *pb.SubPayChUpdatesResp_Notify_, nodeTime int64) {
	currPayChInfo := getChannelInfo(sh, chAlias)

	sh.Printf("%s\n", greenf("Finalizing payment notification received on channel %s. (ID:%s)",
		chAlias, notif.Notify.ProposedPayChInfo.ChID))
	sh.Printf("%s\n", greenf("Channel will closed if this payment update is responded to."))
	sh.Printf("%s\n\n", greenf("Current:\t%s.\nProposed:\t%s.\nExpires in %ds.",
		prettifyBalanceInfo(currPayChInfo.BalInfo, currPayChInfo.Version),
		prettifyBalanceInfo(notif.Notify.ProposedPayChInfo.BalInfo, notif.Notify.ProposedPayChInfo.Version),
		notif.Notify.Expiry-nodeTime))

	if openChannelsMap[chAlias].latestPaymentNotifID != "" {
		sh.Printf("%s", greenf("Dropping older (now obselete) notification received on this channel.",
			openChannelsMap[chAlias].latestPaymentNotifID))
	}
	openChannelsMap[chAlias].latestPaymentNotifID = notif.Notify.UpdateID
}

func handleUpdateTypeClosedNotif(chAlias string, notif *pb.SubPayChUpdatesResp_Notify_) {
	sh.Printf("%s", greenf("Payment channel close notification received on channel %s (ID: %s)\n",
		chAlias, notif.Notify.ProposedPayChInfo.ChID))
	sh.Printf("%s\n\n", greenf("%s.",
		prettifyBalanceInfo(notif.Notify.ProposedPayChInfo.BalInfo, notif.Notify.ProposedPayChInfo.Version)))

	removeOpenChannelID(chAlias)
}

func paymentUnsubFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	// Usage: payment unsub [channel alias]
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	paymentUnsub(c, c.Args[0])
}

func paymentUnsub(c *ishell.Context, chAlias string) {
	chInfo, ok := openChannelsMap[c.Args[0]]
	if !ok {
		printUnknownChannelAliasError(c, c.Args[0])
		return
	}
	req := pb.UnsubPayChUpdatesReq{
		SessionID: sessionID,
		ChID:      chInfo.id,
	}
	resp, err := client.UnsubPayChUpdates(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.UnsubPayChUpdatesResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error unsubscribing from payment notifications: %v", apiErrorString(msgErr.Error)))
		return
	}

	c.Printf("%s\n\n", greenf("Unsubscribed from payment notifications for channel %s (ID: %s)", chAlias, chInfo.id))
}

func paymentAccept(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	// Usage: payment accept [channel alias]
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}
	chInfo, ok := openChannelsMap[c.Args[0]]
	if !ok {
		printUnknownChannelAliasError(c, c.Args[0])
		return
	}
	if chInfo.latestPaymentNotifID == "" {
		printNoPaymentNotifError(c, c.Args[0])
	}

	req := pb.RespondPayChUpdateReq{
		SessionID: sessionID,
		ChID:      chInfo.id,
		UpdateID:  chInfo.latestPaymentNotifID,
		Accept:    true,
	}
	resp, err := client.RespondPayChUpdate(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.RespondPayChUpdateResp_Error)
	if ok {
		sh.Printf("%s\n\n", redf("Error accepting payment channel update: %v", msgErr.Error.Error))
		return
	}
	msg := resp.Response.(*pb.RespondPayChUpdateResp_MsgSuccess_)
	chAlias := openChannelsRevMap[chInfo.id]
	sh.Printf("%s\n\n", greenf("Payment channel updated. Alias: %s. Updated Info:\n%s", chAlias,
		prettifyPayChInfo(msg.MsgSuccess.UpdatedPayChInfo)))

	chInfo.latestPaymentNotifID = ""
}

func paymentReject(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	// Usage: payment reject [channel alias]
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}
	chInfo, ok := openChannelsMap[c.Args[0]]
	if !ok {
		printUnknownChannelAliasError(c, c.Args[0])
		return
	}
	if chInfo.latestPaymentNotifID == "" {
		printNoPaymentNotifError(c, c.Args[0])
	}

	req := pb.RespondPayChUpdateReq{
		SessionID: sessionID,
		ChID:      chInfo.id,
		UpdateID:  chInfo.latestPaymentNotifID,
		Accept:    false,
	}
	resp, err := client.RespondPayChUpdate(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.RespondPayChUpdateResp_Error)
	if ok {
		sh.Printf("%s\n\n", redf("Error rejecting payment channel update: %v", msgErr.Error.Error))
		return
	}
	sh.Printf("%s\n\n", greenf("Payment channel update rejected successfully."))

	chInfo.latestPaymentNotifID = ""
}
