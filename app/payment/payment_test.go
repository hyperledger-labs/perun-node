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

// +build integration

package payment_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/app/payment"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/log"
	"github.com/hyperledger-labs/perun-node/session"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

const aliceAlias, bobAlias = "alice", "bob"

func Test_Integ_PaymentAPI(t *testing.T) {
	handleError := func(err error, msg string) {
		if err != nil {
			fmt.Printf("Error %s:%+v\n", msg, err)
			t.FailNow()
		}
	}
	wg := &sync.WaitGroup{}

	ctx := context.Background()

	// Pre-Setup: Init loggers, deploy contracts and generate session config.
	// *********************************************************************
	fmt.Println("=== Pre-Setup: Init loggers, deploy contracts and generate session config ===")

	err := log.InitLogger("debug", "perun.log")
	handleError(err, "initializing logger")
	fmt.Printf("initialized logger\n")

	currencies := currency.NewRegistry()
	_, err = currencies.Register(currency.ETHSymbol, currency.ETHMaxDecimals)
	require.NoError(t, err)

	chainID := ethereumtest.ChainID
	chainURL := ethereumtest.ChainURL
	onChainTxTimeout := ethereumtest.OnChainTxTimeout
	adjudicator, assetETH, err := ethereumtest.SetupContracts(chainURL, chainID, onChainTxTimeout)
	handleError(err, "deploying contracts")
	fmt.Printf("contracts deployed at: adjudicator:%s, asset ETH:%s\n", adjudicator, assetETH)

	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))

	aliceCfg, err := sessiontest.NewConfig(prng)
	handleError(err, "generating config for alice")

	bobCfg, err := sessiontest.NewConfig(prng)
	handleError(err, "generating config for bob")
	fmt.Printf("configuration for session generated\n")

	// Setup: Initialize sessions and cross register peer IDs.
	// *******************************************************
	fmt.Println("\n=== Setup: Initialize sessions and cross register peer IDs ===")

	aliceSess, err := session.New(aliceCfg, currencies)
	handleError(err, "alice: initializing session")
	fmt.Printf("alice: initialized session, session ID: %s\n", aliceSess.ID())

	bobSess, err := session.New(bobCfg, currencies)
	handleError(err, "initializing session for bob")
	fmt.Printf("bob: initialized session, session ID: %s\n\n", bobSess.ID())

	alicePeerID, err := aliceSess.GetPeerID(perun.OwnAlias)
	handleError(err, "alice: getting own peer ID")
	alicePeerID.Alias = aliceAlias
	fmt.Printf("alice: got own peer ID\n")

	bobPeerID, err := bobSess.GetPeerID(perun.OwnAlias)
	handleError(err, "bob: getting own peer ID")
	bobPeerID.Alias = bobAlias
	fmt.Printf("bob: got own peer ID\n")

	err = aliceSess.AddPeerID(bobPeerID)
	handleError(err, "alice: adding peer ID of bob to ID provider")
	fmt.Printf("alice: added peer ID of bob to ID provider\n")

	err = bobSess.AddPeerID(alicePeerID)
	handleError(err, "bob: adding peer ID of alice to ID provider")
	fmt.Printf("bob: added peer ID of alice to ID provider\n\n")

	// Open phase: Alice opens a channel with bob.
	// *******************************************
	fmt.Println("\n=== Open phase: Alice opens a channel with bob ===")

	var aliceChInfo payment.PayChInfo
	wg.Add(1)
	go func() {
		defer wg.Done()
		openingBalInfo := perun.BalInfo{
			Currency: currency.ETHSymbol,
			Parts:    []string{perun.OwnAlias, bobAlias},
			Bal:      []string{"1", "2"},
		}
		var challengeDurSecs uint64 = 10
		aliceChInfo, err = payment.OpenPayCh(ctx, aliceSess, openingBalInfo, challengeDurSecs)
		handleError(err, "sending open channel request")
		fmt.Printf("alice: channel opened. ID: %s\n", aliceChInfo.ChID)
	}()

	incomingChProposalNotifsBob := make(chan payment.PayChProposalNotif, 2)
	proposalNotifier := func(notif payment.PayChProposalNotif) {
		fmt.Printf("bob: received channel proposal notification\n")
		incomingChProposalNotifsBob <- notif
	}
	err = payment.SubPayChProposals(bobSess, proposalNotifier)
	handleError(err, "subscribing to channel proposals")
	fmt.Printf("bob: subscribed to channel proposal notifications\n")

	notif := <-incomingChProposalNotifsBob
	bobChInfo, err := payment.RespondPayChProposal(ctx, bobSess, notif.ProposalID, true)
	handleError(err, "bob: responding to channel proposal")
	fmt.Printf("bob: accepted channel proposal\n")
	fmt.Printf("opening Balance:%+v\n\n", bobChInfo.BalInfo)
	wg.Wait()

	// Transact phase: Alice sends payment to bob
	// ******************************************
	fmt.Println("\n=== Transact phase: Alice sends payment to bob ===")

	aliceCh, err := aliceSess.GetCh(aliceChInfo.ChID)
	handleError(err, "getting alice channel instance")
	bobCh, err := bobSess.GetCh(bobChInfo.ChID)
	handleError(err, "getting bob channel instance")

	incomingChUpdateNotifsBob := make(chan payment.PayChUpdateNotif, 2)
	updateNotifierBob := func(notif payment.PayChUpdateNotif) {
		incomingChUpdateNotifsBob <- notif
	}
	err = payment.SubPayChUpdates(bobCh, updateNotifierBob)
	handleError(err, "subscribing to channel updates")
	fmt.Printf("bob: subscribed to channel update notifications\n")

	bobUpdatedChInfos := make(chan payment.PayChInfo, 2)
	bobAcceptIncomingUpdate := func() {
		defer wg.Done()
		notif := <-incomingChUpdateNotifsBob
		fmt.Printf("bob: received channel update notification\n")
		updatedPayChInfo, err := payment.RespondPayChUpdate(ctx, bobCh, notif.UpdateID, true)
		handleError(err, "responding to channel updates")
		fmt.Printf("bob: accepted payment from alice\n")
		bobUpdatedChInfos <- updatedPayChInfo
	}
	wg.Add(1)
	go bobAcceptIncomingUpdate()

	aliceUpdateChInfo, err := payment.SendPayChUpdate(ctx, aliceCh, bobAlias, "0.1")
	handleError(err, "sending payment")
	fmt.Printf("alice: sent payment to bob, updated version: %s\n", aliceUpdateChInfo.Version)

	bobUpdatedChInfo := <-bobUpdatedChInfos
	fmt.Printf("updated balance:%+v\n\n", bobUpdatedChInfo.BalInfo)
	wg.Wait()

	// Register and Close phase: Alice closes the channel
	// **************************************************
	fmt.Println("\n=== Register and Close phase: Alice closes the channel ===")

	// Since a subscription to channel updates already exists for bob,
	// we make a new one only for alice.
	fmt.Printf("bob: subscription to channel update notifications already exists\n")
	wg.Add(1)
	go bobAcceptIncomingUpdate() // Accept the finalizing notification.

	incomingChUpdateNotifsAlice := make(chan payment.PayChUpdateNotif, 2)
	updateNotifierAlice := func(notif payment.PayChUpdateNotif) {
		incomingChUpdateNotifsAlice <- notif
	}
	err = payment.SubPayChUpdates(aliceCh, updateNotifierAlice)
	handleError(err, "subscribing to channel updates")
	fmt.Printf("alice: subscribed to channel update notifications\n")

	_, err = payment.ClosePayCh(ctx, aliceCh)
	handleError(err, "closing the channel")

	updateNotif := <-incomingChUpdateNotifsAlice
	if updateNotif.Type != perun.ChUpdateTypeClosed {
		handleError(err, "alice: expecting channel close notification")
	}
	fmt.Printf("alice: channel close notification received\n")

	updateNotif = <-incomingChUpdateNotifsBob
	if updateNotif.Type != perun.ChUpdateTypeClosed {
		handleError(err, "bob: expecting channel close notification")
	}
	fmt.Printf("bob: channel close notification received\n")
	fmt.Printf("closing balance:%+v\n\n", updateNotif.ProposedPayChInfo.BalInfo)
	wg.Wait()
}
