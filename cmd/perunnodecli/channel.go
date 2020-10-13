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
	"fmt"
	"io"
	"strings"

	"github.com/abiosoft/ishell"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/currency"
)

type (
	// openChannelInfo stores the minimal information required by this app for processing
	// payment commands on an open channel.
	openChannelInfo struct {
		id                   string
		peerAlias            string
		latestPaymentNotifID string
	}
)

var (
	channelCmdUsage = "Usage: channel [sub-command]"
	channelCmd      = &ishell.Cmd{
		Name: "channel",
		Help: "Use this command to open/close payment channels." + channelCmdUsage,
		Func: channelFn,
	}

	channelSendCmdUsage = "Usage: channel send-opening-request [peer alias] [own amount] [peer amount]"
	channelSendCmd      = &ishell.Cmd{
		Name: "send-opening-request",
		Help: "Send a request to open a channel with the peer." + channelSendCmdUsage,
		Completer: func([]string) []string {
			return knownAliasesList
		},
		Func: channelSendFn,
	}

	channelSubCmdUsage = "Usage: channel subscribe-opening-request"
	channelSubCmd      = &ishell.Cmd{
		Name: "subscribe-to-opening-requests",
		Help: "Subscribe to channel opening request notications from the peer." + channelSubCmdUsage,
		Func: channelSubFn,
	}

	channelUnsubCmdUsage = "Usage: channel unsubcribe-opening-request"
	channelUnsubCmd      = &ishell.Cmd{
		Name: "unsubcribe-from-opening-requests",
		Help: "Unsubscribe from channel opening request notications from the peer." + channelUnsubCmdUsage,
		Func: channelUnsubFn,
	}

	channelAcceptCmdUsage = "Usage: channel accept-opening-request [channel notification alias]"
	channelAcceptCmd      = &ishell.Cmd{
		Name: "accept-opening-request",
		Help: "Accept channel opening request from the peer." + channelAcceptCmdUsage,
		Completer: func([]string) []string {
			return channelNotifList
		},
		Func: channelAcceptFn,
	}

	channelRejectCmdUsage = "Usage: channel reject-opening-request [channel notification alias]"
	channelRejectCmd      = &ishell.Cmd{
		Name: "reject-opening-request",
		Help: "Reject channel opening request from the peer." + channelRejectCmdUsage,
		Completer: func([]string) []string {
			return channelNotifList
		},
		Func: channelRejectFn,
	}

	channelCloseNSettleCmdUsage = "Usage: channel settle-on-chain [channel alias]"
	channelCloseNSettleCmd      = &ishell.Cmd{
		Name: "close-n-settle-on-chain",
		Help: "Close the channel and settle the balance on the blockchain." + channelCloseNSettleCmdUsage,
		Completer: func([]string) []string {
			return openChannelsList
		},
		Func: chCloseNSettleFn,
	}
	channelListPendingCmdUsage = "Usage: channel list-notifications-pending-response"
	channelListPendingCmd      = &ishell.Cmd{
		Name: "list-notifications-pending-response",
		Help: "List channel opening request notifications pending a response." + channelListPendingCmdUsage,
		Func: channelPendingFn,
	}
	channelListOpenCmdUsage = "Usage: channel list-open-channels"
	channelListOpenCmd      = &ishell.Cmd{
		Name: "list-open-channels",
		Help: "Print the current channel info for all open channels." + channelListOpenCmdUsage,
		Func: channelListOpenFn,
	}
	channelInfoCmdUsage = "Usage: channel info"
	channelInfoCmd      = &ishell.Cmd{
		Name: "info",
		Help: "Print the current channel info." + channelInfoCmdUsage,
		Completer: func([]string) []string {
			return openChannelsList
		},
		Func: channelInfoFn,
	}

	openChannelsCounter = 0 // counter to track the number of channel opened to assign alias numbers.

	openChannelsList   []string                    // List of open channel alias for autocompletion.
	openChannelsMap    map[string]*openChannelInfo // Map of open channel alias to open channel info.
	openChannelsRevMap map[string]string           // Map of open channel id to open channel alias.

	channelNotifCounter = 0 // counter to track the number of proposal opened to assign alias numbers.

	// List of channel notification ids for autocompletion.
	channelNotifList []string
	// Map of channel notification id to the notification payload.
	channelNotifMap map[string]*pb.SubPayChProposalsResp_Notify
)

func init() {
	channelCmd.AddCmd(channelSendCmd)
	channelCmd.AddCmd(channelSubCmd)
	channelCmd.AddCmd(channelUnsubCmd)
	channelCmd.AddCmd(channelAcceptCmd)
	channelCmd.AddCmd(channelRejectCmd)
	channelCmd.AddCmd(channelListPendingCmd)
	channelCmd.AddCmd(channelListOpenCmd)
	channelCmd.AddCmd(channelInfoCmd)
	channelCmd.AddCmd(channelCloseNSettleCmd)

	openChannelsMap = make(map[string]*openChannelInfo)
	openChannelsRevMap = make(map[string]string)

	channelNotifMap = make(map[string]*pb.SubPayChProposalsResp_Notify)
}

// Creates an alias for the channel id, adds the channel to the open channels map and returns the alias.
// It also adds the entry to open channels list for autocompletetion.
func addOpenChannelID(id, peer string) (alias string) {
	openChannelsCounter = openChannelsCounter + 1
	alias = fmt.Sprintf("ch_%d_%s", openChannelsCounter, peer)
	openChannelsMap[alias] = &openChannelInfo{id, peer, ""}
	openChannelsRevMap[id] = alias

	openChannelsList = append(openChannelsList, alias)
	return alias
}

// Removes the entry corresponding to the alias from the openChannelMap and openChannelList.
func removeOpenChannelID(alias string) {
	delete(openChannelsMap, alias)
	aliasIdx := -1
	for idx := range openChannelsList {
		if openChannelsList[idx] == alias {
			aliasIdx = idx
			break
		}
	}
	if aliasIdx != -1 {
		openChannelsList[aliasIdx] = openChannelsList[len(openChannelsList)-1]
		openChannelsList[len(openChannelsList)-1] = ""
		openChannelsList = openChannelsList[:len(openChannelsList)-1]
	}
}

// Creates an alias for the channel opening request notification proposal, adds it to the channel notifications
// map and returns the alias.
// It also adds the entry to channel notifications list for autocompletetion.
func addChannelNotif(notif *pb.SubPayChProposalsResp_Notify) (alias string) {
	channelNotifCounter = channelNotifCounter + 1
	alias = fmt.Sprintf("request_%d_%s", channelNotifCounter, findPeerAlias(notif.OpeningBalInfo.Parts))
	channelNotifMap[alias] = notif

	channelNotifList = append(channelNotifList, alias)
	return alias
}

// Removes the entry corresponding to the alias from the channelNotifMap and channelNotifList.
func removeChannelNotif(alias string) {
	delete(channelNotifMap, alias)
	aliasIdx := -1
	for idx := range channelNotifList {
		if channelNotifList[idx] == alias {
			aliasIdx = idx
			break
		}
	}
	if aliasIdx != -1 {
		channelNotifList[aliasIdx] = channelNotifList[len(channelNotifList)-1]
		channelNotifList[len(channelNotifList)-1] = ""
		channelNotifList = channelNotifList[:len(channelNotifList)-1]
	}
}

func channelFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	c.Println(c.Cmd.HelpText())
}

func channelSendFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}

	// Usage: channel send [peer alias] [own amount] [peer amount]",
	countReqArgs := 3
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	req := pb.OpenPayChReq{
		SessionID: sessionID,
		OpeningBalInfo: &pb.BalInfo{
			Currency: currency.ETH,
			Parts:    []string{perun.OwnAlias, c.Args[0]},
			Bal:      []string{c.Args[1], c.Args[2]},
		},
		ChallengeDurSecs: challengeDurSecs,
	}
	resp, err := client.OpenPayCh(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.OpenPayChResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error opening channel : %v", msgErr.Error.Error))
		return
	}
	msg := resp.Response.(*pb.OpenPayChResp_MsgSuccess_)
	chAlias := addOpenChannelID(msg.MsgSuccess.OpenedPayChInfo.ChID,
		findPeerAlias(msg.MsgSuccess.OpenedPayChInfo.BalInfo.Parts))
	c.Printf("%s\n\n", greenf("Channel opened. Alias: %s.\n%s.", chAlias,
		prettifyPayChInfo(msg.MsgSuccess.OpenedPayChInfo)))

	paymentSub(c, chAlias)
}

func channelSubFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	// Usage: channel sub
	countReqArgs := 0
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}
	channelSub(c)
}

func channelSub(c *ishell.Context) {
	req := pb.SubPayChProposalsReq{
		SessionID: sessionID,
	}
	sub, err := client.SubPayChProposals(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	go channelNotifHandler(sub)
	c.Printf("%s\n\n", greenf("Subscribed to channel opening request notifications."))
}

func channelNotifHandler(sub pb.Payment_API_SubPayChProposalsClient) {
	for {
		notifMsg, err := sub.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				sh.Printf("%s\n\n", redf("Channel opening request subscription closed."))
			} else {
				sh.Printf("%s\n\n", redf("Error receiving channel opening request notification: %v.", err))
			}
			return
		}
		msgErr, ok := notifMsg.Response.(*pb.SubPayChProposalsResp_Error)
		if ok {
			sh.Printf("%s\n\n", redf("Error received in channel opening request notification: %v.", msgErr.Error.Error))
			return
		}
		notif := notifMsg.Response.(*pb.SubPayChProposalsResp_Notify_)
		channelNotifAlias := addChannelNotif(notif.Notify)
		nodeTime, err := getNodeTime()
		if err != nil {
			printCommandSendingError(sh, err)
			return
		}
		sh.Printf("%s\n\n",
			greenf("Channel opening request notification received. Notification Alias: %s.\n%s.\nExpires in %ds.",
				channelNotifAlias, prettifyChannelOpeningRequest(notif.Notify), notif.Notify.Expiry-nodeTime))
	}
}

func channelUnsubFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	// Usage: channel unsub
	countReqArgs := 0
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	channelUnsub(c)
}

func channelUnsub(c *ishell.Context) {
	req := pb.UnsubPayChProposalsReq{
		SessionID: sessionID,
	}
	resp, err := client.UnsubPayChProposals(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.UnsubPayChProposalsResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error unsubscribing from channel proposals : %v.", msgErr.Error.Error))
		return
	}
}

func channelAcceptFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}

	// Usage: channel accept [channel notification alias]",
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	channelNotif, ok := channelNotifMap[c.Args[0]]
	if !ok {
		printUnknownChNotifError(c, c.Args[0])
		return
	}
	defer removeChannelNotif(c.Args[0])

	req := pb.RespondPayChProposalReq{
		SessionID:  sessionID,
		ProposalID: channelNotif.ProposalID,
		Accept:     true,
	}
	resp, err := client.RespondPayChProposal(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.RespondPayChProposalResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error responding accept to channel opening request : %v.", msgErr.Error.Error))
		return
	}
	msg := resp.Response.(*pb.RespondPayChProposalResp_MsgSuccess_)
	chAlias := addOpenChannelID(msg.MsgSuccess.OpenedPayChInfo.ChID,
		findPeerAlias(msg.MsgSuccess.OpenedPayChInfo.BalInfo.Parts))

	c.Printf("%s\n\n", greenf("Channel opened. Alias: %s.\n%s.", chAlias,
		prettifyPayChInfo(msg.MsgSuccess.OpenedPayChInfo)))

	paymentSub(c, chAlias)
}

// FindPeerAlias finds the alias of the peer in the given list of channel participants.
// It expects that, the list has two entries and the entry other self is returned as peer.
func findPeerAlias(parts []string) string {
	if len(parts) != 2 {
		return ""
	}
	for idx := range parts {
		if parts[idx] != perun.OwnAlias {
			return parts[idx]
		}
	}
	return ""
}

func channelRejectFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}

	// Usage: channel reject [channel notification alias]",
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}

	channelNotif, ok := channelNotifMap[c.Args[0]]
	if !ok {
		printUnknownChNotifError(c, c.Args[0])
		return
	}
	defer removeChannelNotif(c.Args[0])

	req := pb.RespondPayChProposalReq{
		SessionID:  sessionID,
		ProposalID: channelNotif.ProposalID,
		Accept:     false,
	}
	resp, err := client.RespondPayChProposal(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.RespondPayChProposalResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error responding accept to channel opening request : %v.", msgErr.Error.Error))
		return
	}
	c.Printf("%s\n\n", greenf("Channel proposal rejected successfully."))
}

func chCloseNSettleFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	// Usage: channel settle-on-chain [channel alias]",
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
	req := pb.ClosePayChReq{
		SessionID: sessionID,
		ChID:      chInfo.id,
	}
	resp, err := client.ClosePayCh(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return
	}
	msgErr, ok := resp.Response.(*pb.ClosePayChResp_Error)
	if ok {
		sh.Printf("%s\n\n", redf("Error closing channel update: %v", msgErr.Error.Error))
		return
	}
	msg := resp.Response.(*pb.ClosePayChResp_MsgSuccess_)
	sh.Printf("%s\n\n", greenf("Channel closed. Alias: %s. Updated Info:\n%s", c.Args[0],
		prettifyPayChInfo(msg.MsgSuccess.ClosedPayChInfo)))

	removeOpenChannelID(c.Args[0])
}

func channelPendingFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}
	// Usage: channel list-notifications-pending-response
	countReqArgs := 0
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}
	nodeTime, err := getNodeTime()
	if err != nil {
		printCommandSendingError(sh, err)
		return
	}
	for notifAlias, notif := range channelNotifMap {
		sh.Printf("%s\n\n", greenf("Notification Alias: %s.\n%s.\nExpires in %ds.",
			notifAlias, prettifyChannelOpeningRequest(notif), notif.Expiry-nodeTime))
	}
}

func channelListOpenFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}

	// Usage: channel list-open-channels
	countReqArgs := 0
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}
	for i := range openChannelsList {
		payChInfo := getChannelInfo(c, openChannelsList[i])
		c.Printf("%s\n", greenf("Channel %s Info:\n%s.", openChannelsList[i], prettifyPayChInfo(payChInfo)))
	}
	c.Printf("\n")
}

func channelInfoFn(c *ishell.Context) {
	if client == nil {
		printNodeNotConnectedError(c)
		return
	}

	// Usage: channel info [channel alias]
	countReqArgs := 1
	if len(c.Args) != countReqArgs {
		printArgCountError(c, countReqArgs)
		return
	}
	payChInfo := getChannelInfo(c, c.Args[0])
	c.Printf("%s\n\n", greenf("Channel %s Info:\n%s.", c.Args[0], prettifyPayChInfo(payChInfo)))
}

func getChannelInfo(c ishell.Actions, chAlias string) *pb.PayChInfo {
	chInfo, ok := openChannelsMap[chAlias]
	if !ok {
		printUnknownChannelAliasError(c, chAlias)
		return nil
	}

	req := pb.GetPayChInfoReq{
		SessionID: sessionID,
		ChID:      chInfo.id,
	}
	resp, err := client.GetPayChInfo(context.Background(), &req)
	if err != nil {
		printCommandSendingError(c, err)
		return nil
	}
	msgErr, ok := resp.Response.(*pb.GetPayChInfoResp_Error)
	if ok {
		c.Printf("%s\n\n", redf("Error getting channel info: %v.", msgErr.Error.Error))
		return nil
	}

	msg := resp.Response.(*pb.GetPayChInfoResp_MsgSuccess_)
	return msg.MsgSuccess.PayChInfo
}

func prettifyChannelOpeningRequest(notif *pb.SubPayChProposalsResp_Notify) string {
	return fmt.Sprintf("Currency: %s, Balance: %v",
		notif.OpeningBalInfo.Currency, toBalanceMap(notif.OpeningBalInfo.Parts, notif.OpeningBalInfo.Bal))
}

func prettifyBalanceInfo(balInfo *pb.BalInfo, v string) string {
	return fmt.Sprintf("Currency: %s, Balance: %v, Version: %s",
		balInfo.Currency, toBalanceMap(balInfo.Parts, balInfo.Bal), v)
}

func prettifyPayChInfo(chInfo *pb.PayChInfo) string {
	return fmt.Sprintf("ID: %s, Currency: %s, Version: %s, Balance %v",
		chInfo.ChID, chInfo.BalInfo.Currency, chInfo.Version, toBalanceMap(chInfo.BalInfo.Parts, chInfo.BalInfo.Bal))
}

func toBalanceMap(parts, bal []string) string {
	m := make(map[string]string)
	for i := range parts {
		m[parts[i]] = bal[i]
	}
	return strings.Replace(fmt.Sprintf("%v", m), "map", "", -1)
}
