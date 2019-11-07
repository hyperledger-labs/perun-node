// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/dst-go
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
	"math/big"
	"sync"
	"time"

	"github.com/direct-state-transfer/dst-go/blockchain"
	"github.com/direct-state-transfer/dst-go/channel"
	"github.com/direct-state-transfer/dst-go/ethereum/adapter"
	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/fatih/color"
)

func realBlockchain(alice, bob bool, wg *sync.WaitGroup, dispute bool) {
	defer wg.Done()

	fmt.Printf("\nStart of Walkthrough with real blockchain connection")
	fmt.Printf("\nGreen is from alice (sender),\nYellow is from Bob(receiver) \n\n")

	wg2 := &sync.WaitGroup{}

	if bob {
		wg2.Add(1)
		err := realBlockchainBob(bobColor, wg2)
		if err != nil {
			fmt.Printf("\nError initializing bob walkthrough - %s\n", err)
			return
		}
	}

	if alice {
		wg2.Add(1)
		realBlockchainAlice(aliceColor, wg2, dispute)
	}

	wg2.Wait()

	fmt.Printf("\nEnd of Walkthrough with real blockchain connection\n\n")
}

func realBlockchainAlice(printer *color.Color, wg *sync.WaitGroup, dispute bool) {
	defer wg.Done()

	conn, err := adapter.NewRealBackend(ethereumNodeURL)
	if err != nil {
		_, _ = printer.Printf("\nError connection to blockchain - %s\n", err)
		return
	}
	_, _ = printer.Printf("\nConnected to ethereum blockchain node at - %s", ethereumNodeURL)
	bcInst := blockchain.NewInstance(conn, aliceID)

	var params []interface{}

	newConnToBob, err := channel.NewChannel(aliceID, bobID, channel.WebSocket)
	if err != nil {
		_, _ = printer.Printf("\nNew channel error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\n\nFound user at %s\n\n", newConnToBob.PeerID())

	status, reason, err := newConnToBob.NewChannelRequest(channel.Version, contract.Store.SHA256Sum())
	if err != nil {
		_, _ = printer.Printf("\nNew channel request error - %v\n", err)
		return
	}
	if status != channel.MessageStatusAccept {
		_, _ = printer.Printf("\nNew channel request not accepted by peer, got status - %s, reason -%s\n", status, reason)
		return
	}
	_, _ = printer.Printf("\n\nNew outgoing channel established with %s\n\n", newConnToBob.PeerID())

	sid := channel.NewSessionID(aliceID.OnChainID, bobID.OnChainID)
	err = sid.GenerateSenderPart(aliceID.OnChainID)
	if err != nil {
		_, _ = printer.Printf("\nGenerate sid sender part error - %v\n", err)
		return
	}
	sid, status, err = newConnToBob.SessionIDRequest(sid)
	if err != nil {
		_, _ = printer.Printf("\nSession id request error - %v\n", err)
		return
	}
	if status != channel.MessageStatusAccept {
		_, _ = printer.Printf("\nSession id declined by peer\n")
		return
	}

	err = newConnToBob.SetSessionID(sid)
	if err != nil {
		_, _ = printer.Printf("\nSession id invalid - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nSession Id - 0x%x\n", sid.SidComplete.Bytes())

	//Deploy and share libsignatures contract
	params = nil

	_, _ = printer.Printf("\nDeploy Libsignature\n")
	aliceID.SetCredentials(testKeystore, alicePassword)
	libSigAddr, _, _, err := adapter.DeployContract(contract.Store.LibSignatures(), conn, params, aliceID)
	if err != nil {
		_, _ = printer.Printf("\nDeploy Libsignature error - %v\n", err)
		return
	}
	err = bcInst.SetLibSignatures(libSigAddr)
	if err != nil {
		_, _ = printer.Printf("\nSet libsignatures contract address error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nLibsignature deployed at %x \n", libSigAddr)

	status, err = newConnToBob.ContractAddrRequest(bcInst.LibSignatures(), contract.Store.LibSignatures())
	if err != nil {
		_, _ = printer.Printf("\nContract Address Request (LibSignature) error - %v\n", err)
		return
	}
	if status != channel.MessageStatusAccept {
		_, _ = printer.Printf("\nContract Address (LibSignature) not accepted by peer - %s\n", status)
		return
	}

	//Deploy and share vpc contract
	_, _ = printer.Printf("\nDeploy VPC\n")
	_ = bcInst.OwnerID.SetCredentials(testKeystore, alicePassword)
	err = bcInst.DeployVPC()
	if err != nil {
		_, _ = printer.Printf("\nDeploy VPC error %v\n", err)
		return
	}
	_, _ = printer.Printf("\nVpc deployed at %x \n", bcInst.VPCAddr())

	status, err = newConnToBob.ContractAddrRequest(bcInst.VPCAddr(), contract.Store.VPC())
	if err != nil {
		_, _ = printer.Printf("\nContract Address Request (VPC) error - %v\n", err)
		return
	}
	if status != channel.MessageStatusAccept {
		_, _ = printer.Printf("\nContract Address (vpc) not accepted by peer - %s\n", status)
		return
	}

	//Deploy and share mscontract
	_, _ = printer.Printf("\nDeploy MSContract\n")
	aliceID.SetCredentials(testKeystore, alicePassword)
	err = bcInst.DeployMSContract(aliceID.OnChainID, bobID.OnChainID)
	if err != nil {
		_, _ = printer.Printf("\nDeploy MSContract error %v\n", err)
		return
	}
	_, _ = printer.Printf("\nMscontract deployed at %x \n", bcInst.MSContractAddr())

	status, err = newConnToBob.ContractAddrRequest(bcInst.MSContractAddr(), contract.Store.MSContract())
	if err != nil {
		_, _ = printer.Printf("\nContract Address Request (mscontract) error - %v\n", err)
		return
	}
	if status != channel.MessageStatusAccept {
		_, _ = printer.Printf("\nContract Address (ms contract) not accepted by peer - %s\n", status)
		return
	}

	eventsChan := bcInst.EventsChan

	//Filtering events is not working in geth v 1.8.20. Works in v1.8.18
	//Hence timeout in 5 seconds if no event is coming. This is applicable only to msEventInitializing
	ticker := time.After(5 * time.Second)
	select {
	case mscEventInitializing := <-eventsChan.MSCInitializingChan:
		_, _ = printer.Printf("\nMscEventInitializing : Address sender - %s, Address receiver - %s\n",
			mscEventInitializing.AddressAlice.String(), mscEventInitializing.AddressBob.String())
	case err := <-eventsChan.MSCInitalizingSub.Err():
		_, _ = printer.Printf("\nMscEventInitializing : Error - %s\n", err)
	case <-ticker:
		_, _ = printer.Printf("\nMscEventInitializing : Timedout %s\n", err)
	}

	blockedSender := types.EtherToWei(big.NewInt(10))
	_ = bcInst.OwnerID.SetCredentials(testKeystore, alicePassword)
	_, _ = printer.Printf("\nConfirm call\n")
	_ = bcInst.Confirm(blockedSender)
	_, _ = printer.Printf("\nConfirm call successful\n")

	mscEventInitialized := <-eventsChan.MSCInitializedChan
	_, _ = printer.Printf("\nMscEventInitialized : Cash sender - %s, Cash receiver - %s\n",
		mscEventInitialized.CashAlice.String(), mscEventInitialized.CashBob.String())

	//MSC Base State - by alice
	mscBaseStateSignedPartial := channel.MSCBaseStateSigned{
		MSContractBaseState: channel.MSCBaseState{
			VpcAddress:      bcInst.VPCAddr(),
			Sid:             sid.SidComplete,
			BlockedSender:   types.EtherToWei(big.NewInt(10)),
			BlockedReceiver: types.EtherToWei(big.NewInt(10)),
			Version:         big.NewInt(1),
		}}
	aliceID.SetCredentials(testKeystore, alicePassword)
	if err = mscBaseStateSignedPartial.AddSign(aliceID, channel.Sender); err != nil {
		_, _ = printer.Printf("\nSigning msc base state error - %v\n", err)
		return
	}
	mscBaseStateSigned, _, err := newConnToBob.NewMSCBaseStateRequest(mscBaseStateSignedPartial)
	if err != nil {
		_, _ = printer.Printf("\nMsc base state request error %v\n", err)
		return
	}

	err = newConnToBob.SetMSCBaseState(mscBaseStateSigned)
	if err != nil {
		_, _ = printer.Printf("\nMsc base state error %v\n", err)
		return
	}
	_, _ = printer.Printf("\nMsc base state - %+v\n", mscBaseStateSigned)

	_ = bcInst.OwnerID.SetCredentials(testKeystore, alicePassword)
	_, _ = printer.Printf("\nState register call\n")
	_ = bcInst.StateRegister(
		mscBaseStateSigned.MSContractBaseState.Sid,
		mscBaseStateSigned.MSContractBaseState.Version,
		mscBaseStateSigned.MSContractBaseState.BlockedSender,
		mscBaseStateSigned.MSContractBaseState.BlockedReceiver,
		mscBaseStateSigned.SignSender,
		mscBaseStateSigned.SignReceiver)
	_, _ = printer.Printf("\nState register call successful\n")

	<-eventsChan.MSCStateRegisteringChan
	_, _ = printer.Printf("\nMscEventStateRegistering\n")

	mscEventStateRegistered := <-eventsChan.MSCStateRegisteredChan
	_, _ = printer.Printf("\nMscEventStateRegistered - Cash sender - %s, Cash receiver -%s\n",
		mscEventStateRegistered.BlockedAlice.String(), mscEventStateRegistered.BlockedBob.String())

	vpcStateID := channel.VPCStateID{
		AddSender:    aliceEthereumAddr,
		AddrReceiver: bobEthereumAddr,
		SID:          sid.SidComplete,
	}

	//VPC State version 1 - by alice
	vpcStateSignedPartial01 := channel.VPCStateSigned{
		VPCState: channel.VPCState{
			ID:              vpcStateID.SoliditySHA3(),
			Version:         big.NewInt(1),
			BlockedSender:   types.EtherToWei(big.NewInt(9)),
			BlockedReceiver: types.EtherToWei(big.NewInt(11)),
		},
	}
	aliceID.SetCredentials(testKeystore, alicePassword)
	if err = vpcStateSignedPartial01.AddSign(aliceID, channel.Sender); err != nil {
		_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
		return
	}
	vpcStateSigned01, _, err := newConnToBob.NewVPCStateRequest(vpcStateSignedPartial01)
	if err != nil {
		_, _ = printer.Printf("\nVpc state request error - %v\n", err)
		return
	}

	err = newConnToBob.SetCurrentVPCState(vpcStateSigned01)
	if err != nil {
		_, _ = printer.Printf("\nVpc state error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned01)

	//VPC State version 2 - by bob
	//Read, Sign and send back the vpc state with accept
	vpcStateSigned02, err := newConnToBob.NewVPCStateRead()
	if err != nil {
		_, _ = printer.Printf("\nVpc state read error - %v\n", err)
		return
	}
	aliceID.SetCredentials(testKeystore, alicePassword)
	if err = vpcStateSigned02.AddSign(aliceID, channel.Sender); err != nil {
		_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
		return
	}
	err = newConnToBob.NewVPCStateRespond(vpcStateSigned02, channel.MessageStatusAccept)
	if err != nil {
		_, _ = printer.Printf("\nVpc state respond error - %v\n", err)
		return
	}

	err = newConnToBob.SetCurrentVPCState(vpcStateSigned02)
	if err != nil {
		_, _ = printer.Printf("\nVpc state error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned02)

	//VPC State version 3 - by bob
	//Read, Sign and send back the vpc state with accept
	vpcStateSigned03, err := newConnToBob.NewVPCStateRead()
	if err != nil {
		_, _ = printer.Printf("\nVpc state read error - %v\n", err)
		return
	}
	aliceID.SetCredentials(testKeystore, alicePassword)
	if err = vpcStateSigned03.AddSign(aliceID, channel.Sender); err != nil {
		_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
		return
	}
	err = newConnToBob.NewVPCStateRespond(vpcStateSigned03, channel.MessageStatusAccept)
	if err != nil {
		_, _ = printer.Printf("\nvpc state respond error= %v\n", err)
		return
	}

	err = newConnToBob.SetCurrentVPCState(vpcStateSigned03)
	if err != nil {
		_, _ = printer.Printf("\nVpc state error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned03)

	//VPC State version 4 - by alice
	vpcStateSignedPartial04 := channel.VPCStateSigned{
		VPCState: channel.VPCState{
			ID:              vpcStateID.SoliditySHA3(),
			Version:         big.NewInt(4),
			BlockedSender:   types.EtherToWei(big.NewInt(6)),
			BlockedReceiver: types.EtherToWei(big.NewInt(14)),
		},
	}
	aliceID.SetCredentials(testKeystore, alicePassword)
	if err = vpcStateSignedPartial04.AddSign(aliceID, channel.Sender); err != nil {
		_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
		return
	}
	vpcStateSigned04, _, err := newConnToBob.NewVPCStateRequest(vpcStateSignedPartial04)
	if err != nil {
		_, _ = printer.Printf("\nvpc state request error %v\n", err)
		return
	}

	err = newConnToBob.SetCurrentVPCState(vpcStateSigned04)
	if err != nil {
		_, _ = printer.Printf("\nVpc state error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned04)

	//VPC State version 5 - by bob
	//Read, Sign and send back the vpc state with accept
	vpcStateSigned05, err := newConnToBob.NewVPCStateRead()
	if err != nil {
		_, _ = printer.Printf("\nVpc state read error - %v\n", err)
		return
	}
	aliceID.SetCredentials(testKeystore, alicePassword)
	if err = vpcStateSigned05.AddSign(aliceID, channel.Sender); err != nil {
		_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
		return
	}
	err = newConnToBob.NewVPCStateRespond(vpcStateSigned05, channel.MessageStatusAccept)
	if err != nil {
		_, _ = printer.Printf("\nvpc state respond error= %v\n", err)
		return
	}

	err = newConnToBob.SetCurrentVPCState(vpcStateSigned05)
	if err != nil {
		_, _ = printer.Printf("\nVpc state error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned05)

	//VPC State version 6 - by alice
	vpcStateSignedPartial06 := channel.VPCStateSigned{
		VPCState: channel.VPCState{
			ID:              vpcStateID.SoliditySHA3(),
			Version:         big.NewInt(6),
			BlockedSender:   types.EtherToWei(big.NewInt(4)),
			BlockedReceiver: types.EtherToWei(big.NewInt(16)),
		},
	}
	aliceID.SetCredentials(testKeystore, alicePassword)
	if err = vpcStateSignedPartial06.AddSign(aliceID, channel.Sender); err != nil {
		_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
		return
	}
	vpcStateSigned06, _, err := newConnToBob.NewVPCStateRequest(vpcStateSignedPartial06)
	if err != nil {
		_, _ = printer.Printf("\nvpc state request error %v\n", err)
		return
	}

	err = newConnToBob.SetCurrentVPCState(vpcStateSigned06)
	if err != nil {
		_, _ = printer.Printf("\nVpc state error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned06)

	//VPC State version 7 - by bob
	//Read, Sign and send back the vpc state with accept
	vpcStateSigned07, err := newConnToBob.NewVPCStateRead()
	if err != nil {
		_, _ = printer.Printf("\nVpc state read error - %v\n", err)
		return
	}
	aliceID.SetCredentials(testKeystore, alicePassword)
	if err = vpcStateSigned07.AddSign(aliceID, channel.Sender); err != nil {
		_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
		return
	}
	err = newConnToBob.NewVPCStateRespond(vpcStateSigned07, channel.MessageStatusAccept)
	if err != nil {
		_, _ = printer.Printf("\nvpc state respond error= %v\n", err)
		return
	}
	err = newConnToBob.SetCurrentVPCState(vpcStateSigned07)
	if err != nil {
		_, _ = printer.Printf("\nVpc state error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned07)

	//VPC State version 8 - by alice
	vpcStateSignedPartial08 := channel.VPCStateSigned{
		VPCState: channel.VPCState{
			ID:              vpcStateID.SoliditySHA3(),
			Version:         big.NewInt(8),
			BlockedSender:   types.EtherToWei(big.NewInt(2)),
			BlockedReceiver: types.EtherToWei(big.NewInt(18)),
		},
	}
	aliceID.SetCredentials(testKeystore, alicePassword)
	if err = vpcStateSignedPartial08.AddSign(aliceID, channel.Sender); err != nil {
		_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
		return
	}
	vpcStateSigned08, _, err := newConnToBob.NewVPCStateRequest(vpcStateSignedPartial08)
	if err != nil {
		_, _ = printer.Printf("\nvpc state request error %v\n", err)
		return
	}

	err = newConnToBob.SetCurrentVPCState(vpcStateSigned08)
	if err != nil {
		_, _ = printer.Printf("\nVpc state error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned08)

	//VPC State version 9 - by alice
	vpcStateSignedPartial09 := channel.VPCStateSigned{
		VPCState: channel.VPCState{
			ID:              vpcStateID.SoliditySHA3(),
			Version:         big.NewInt(9),
			BlockedSender:   types.EtherToWei(big.NewInt(1)),
			BlockedReceiver: types.EtherToWei(big.NewInt(19)),
		},
	}
	aliceID.SetCredentials(testKeystore, alicePassword)
	if err = vpcStateSignedPartial09.AddSign(aliceID, channel.Sender); err != nil {
		_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
		return
	}
	vpcStateSigned09, _, err := newConnToBob.NewVPCStateRequest(vpcStateSignedPartial09)
	if err != nil {
		_, _ = printer.Printf("\nvpc state request error %v\n", err)
		return
	}

	err = newConnToBob.SetCurrentVPCState(vpcStateSigned09)
	if err != nil {
		_, _ = printer.Printf("\nVpc state error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned09)

	//VPC State version 10 - by bob
	//Read, Sign and send back the vpc state with accept
	vpcStateSigned10, err := newConnToBob.NewVPCStateRead()
	if err != nil {
		_, _ = printer.Printf("\nVpc state read error - %v\n", err)
		return
	}
	aliceID.SetCredentials(testKeystore, alicePassword)
	if err = vpcStateSigned10.AddSign(aliceID, channel.Sender); err != nil {
		_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
		return
	}
	err = newConnToBob.NewVPCStateRespond(vpcStateSigned10, channel.MessageStatusAccept)
	if err != nil {
		_, _ = printer.Printf("\nvpc state respond error= %v\n", err)
		return
	}

	err = newConnToBob.SetCurrentVPCState(vpcStateSigned10)
	if err != nil {
		_, _ = printer.Printf("\nVpc state error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned10)

	var vpcClosingState channel.VPCStateSigned
	if dispute {
		//Call close with an older state in order to create dispute
		vpcClosingState = vpcStateSigned06
	} else {
		vpcClosingState = vpcStateSigned10
	}

	_ = bcInst.OwnerID.SetCredentials(testKeystore, alicePassword)
	_, _ = printer.Printf("\nVPCClose call with state %+v\n", vpcClosingState.VPCState)
	_ = bcInst.VPCClose(
		vpcStateID.SID, vpcClosingState.VPCState.Version,
		aliceEthereumAddr, bobEthereumAddr,
		vpcClosingState.VPCState.BlockedSender,
		vpcClosingState.VPCState.BlockedReceiver,
		vpcClosingState.SignSender, vpcClosingState.SignReceiver)
	_, _ = printer.Printf("\nVPCClose call successful\n")

	vpcClosing := <-eventsChan.VPCVPCClosingChan
	_, _ = printer.Printf("\nVpcEventVpcClosing : id - %x\n", vpcClosing.Id)
	states, err := bcInst.States()
	if err != nil {
		_, _ = printer.Printf("Retrieve vpc states error - %v", err)
	}
	_, _ = printer.Printf("closing state - %+v\n", states)

	vpcClosed := <-eventsChan.VPCVPCClosedChan
	_, _ = printer.Printf("\nVpcEventVpcClosed : id - %x, Cash sender - %s, Cash receiver -%s\n",
		vpcClosed.Id, vpcClosed.CashAlice.String(), vpcClosed.CashBob.String())
	states, err = bcInst.States()
	if err != nil {
		_, _ = printer.Printf("Retrieve vpc states error - %v", err)
	}
	_, _ = printer.Printf("closed with state - %+v\n", states)

	_ = bcInst.OwnerID.SetCredentials(testKeystore, alicePassword)
	_, _ = printer.Printf("\nExecute call\n")
	_ = bcInst.Execute(aliceID.OnChainID, bobID.OnChainID)
	_, _ = printer.Printf("\nExecute call successful\n")

	<-eventsChan.MSCClosedChan
	_, _ = printer.Printf("\nMscEventClosed\n")

	err = newConnToBob.Close()
	if err != nil {
		_, _ = printer.Printf("\nClose channel error - %v\n", err)
		return
	}
	_, _ = printer.Printf("\nClose channel successful\n")
}

func realBlockchainBob(printer *color.Color, wg *sync.WaitGroup) (err error) {

	//Initialize bob's bcInstance
	conn2, err := adapter.NewRealBackend(ethereumNodeURL)
	if err != nil {
		_, _ = printer.Printf("\nError connecting to blockchain -%s\n", err)
		return
	}
	_, _ = printer.Printf("\nConnected to ethereum blockchain node at - %s", ethereumNodeURL)
	bcInst2 := blockchain.NewInstance(conn2, bobID)

	//Initialize a new channel listener for bob
	maxConn := uint32(100)
	incomingConnChan, listener, err := channel.NewSession(bobID, channel.WebSocket, maxConn)
	if err != nil {
		_, _ = printer.Printf("\nNew channel session error - %v\n", err)
		return
	}

	_, _ = printer.Printf("Node initialised. Open for incoming channel requests\n")

	go func(bcInst2 *blockchain.Instance,
		incomingConnChan chan *channel.Instance, wg *sync.WaitGroup) {
		defer wg.Done()

		newConnFromAlice := <-incomingConnChan
		_, _ = printer.Printf("\n\nNew channel request from Ethereum Address %s\n\n",
			newConnFromAlice.PeerID().OnChainID.String())

		msgProtocolVersion, contractStoreVersion, err := newConnFromAlice.NewChannelRead()
		if err != nil {
			_, _ = printer.Printf("\nNew channel read error - %v\n", err)
			return
		}

		err = newConnFromAlice.NewChannelRespond(msgProtocolVersion, contractStoreVersion, channel.MessageStatusAccept, "")
		if err != nil {
			_, _ = printer.Printf("\nNew channel respond error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\n\nNew incoming connection accepted from Ethereum Address %s\n\n",
			newConnFromAlice.PeerID().OnChainID.String())

		newSid, err := newConnFromAlice.SessionIDRead()
		if err != nil {
			_, _ = printer.Printf("\nSessionIdRead error - %s\n", err)
		}
		err = newSid.GenerateReceiverPart(bobID.OnChainID)
		if err != nil {
			_, _ = printer.Printf("\nSessionId generate receiver part error - %s\n", err)
		}
		err = newSid.GenerateCompleteSid()
		if err != nil {
			_, _ = printer.Printf("\nSessionId generate complete sid error - %s\n", err)
		}

		err = newConnFromAlice.SessionIDRespond(newSid, channel.MessageStatusAccept)
		if err != nil {
			_, _ = printer.Printf("\nSessionIdRespond error - %s\n", err)
		}
		err = newConnFromAlice.SetSessionID(newSid)
		if err != nil {
			_, _ = printer.Printf("\nSession id invalid - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nSession Id - 0x%x\n", newSid.SidComplete.Bytes())

		//Read and accept lib signatures address
		libSignAddr, libSignHandlerType, err := newConnFromAlice.ContractAddrRead()
		if err != nil {
			_, _ = printer.Printf("\nContract Address Read (LibSignature) error - %v\n", err)
			return
		}
		err = bcInst2.SetLibSignatures(libSignAddr)
		if err != nil {
			_, _ = printer.Printf("\n Set LibSignatures address error %v\n", err)

			err = newConnFromAlice.ContractAddrRespond(libSignAddr, libSignHandlerType, channel.MessageStatusDecline)
			if err != nil {
				_, _ = printer.Printf("\nContract Addresss Respond (LibSignature) error - %v\n", err)
				return
			}
			_, _ = printer.Printf("\nDeclined libsignatures address from peer - %s\n", libSignAddr.Hex())
			return
		}

		err = newConnFromAlice.ContractAddrRespond(libSignAddr, libSignHandlerType, channel.MessageStatusAccept)
		if err != nil {
			_, _ = printer.Printf("\nContract Addresss Respond (LibSignature) error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nAccepted libsignatures address from peer - %s\n", libSignAddr.Hex())

		//Read and accept vpc address
		vpcAddr, vpcHandlerType, err := newConnFromAlice.ContractAddrRead()
		if err != nil {
			_, _ = printer.Printf("\nContract Addresss Read (vpc) error - %v\n", err)
			return
		}
		err = bcInst2.SetVPCAddr(vpcAddr)
		if err != nil {
			_, _ = printer.Printf("\n Set vpc address error %v\n", err)

			err = newConnFromAlice.ContractAddrRespond(vpcAddr, vpcHandlerType, channel.MessageStatusDecline)
			if err != nil {
				_, _ = printer.Printf("\nContract Addresss Respond (vpc) error - %v\n", err)
				return
			}
			_, _ = printer.Printf("\nDeclined vpc address from peer - %s\n", vpcAddr.Hex())
			return
		}
		err = newConnFromAlice.ContractAddrRespond(vpcAddr, vpcHandlerType, channel.MessageStatusAccept)
		if err != nil {
			_, _ = printer.Printf("\nContract Addresss Respond (vpc) error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nAccepted vpc address from peer - %s\n", vpcAddr.Hex())

		//Read and accept mscontract address
		mscontractAddr, mscontractHandlerType, err := newConnFromAlice.ContractAddrRead()
		if err != nil {
			_, _ = printer.Printf("\nContract Addresss Read (ms contract) error - %v\n", err)
			return
		}
		err = bcInst2.SetMSContractAddr(mscontractAddr)
		if err != nil {
			_, _ = printer.Printf("\n Set mscontract address error %v\n", err)

			err = newConnFromAlice.ContractAddrRespond(mscontractAddr, mscontractHandlerType, channel.MessageStatusDecline)
			if err != nil {
				_, _ = printer.Printf("\nContract Addresss Respond (ms contract) error - %v\n", err)
				return
			}
			_, _ = printer.Printf("\nDeclined ms contract address from peer - %s\n", mscontractAddr.Hex())
			return
		}
		err = newConnFromAlice.ContractAddrRespond(mscontractAddr, mscontractHandlerType, channel.MessageStatusAccept)
		if err != nil {
			_, _ = printer.Printf("\nContract Addresss Respond (ms contract) error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nAccepted ms contract address from peer - %s\n", mscontractAddr.Hex())

		_ = bcInst2.SetLibSignatures(libSignAddr)
		_ = bcInst2.SetMSContractAddr(mscontractAddr)
		_ = bcInst2.SetVPCAddr(vpcAddr)
		vpcInst, err := contract.NewVPC(bcInst2.VPCAddr().Address, bcInst2.Conn)
		if err != nil {
			_, _ = printer.Printf("Instantiate vpc error - %s", err.Error())
			return
		}
		bcInst2.VPCInst = vpcInst
		mscontractInst, err := contract.NewMSContract(bcInst2.MSContractAddr().Address, bcInst2.Conn)
		if err != nil {
			_, _ = printer.Printf("Instantiate ms contract error - %s", err.Error())
			return
		}
		bcInst2.MSContractInst = mscontractInst

		eventsChan, err := bcInst2.InitializeEventsChan()
		if err != nil {
			_, _ = printer.Println("Error initializing events channel")
			return
		}

		//Filtering events is not working in geth v 1.8.20. Works in v1.8.18
		//Hence timeout in 5 seconds if no event is coming. This is applicable only to msEventInitializing
		ticker := time.After(5 * time.Second)
		select {
		case mscEventInitializing := <-eventsChan.MSCInitializingChan:
			_, _ = printer.Printf("\nMscEventInitializing : Address sender - %s, Address receiver - %s\n",
				mscEventInitializing.AddressAlice.String(), mscEventInitializing.AddressBob.String())
		case err := <-eventsChan.MSCInitalizingSub.Err():
			_, _ = printer.Printf("\nMscEventInitializing : Error - %s\n", err)
		case <-ticker:
			_, _ = printer.Printf("\nMscEventInitializing : Timedout %s\n", err)
		}

		blockedReceiver := types.EtherToWei(big.NewInt(10))
		_ = bcInst2.OwnerID.SetCredentials(testKeystore, bobPassword)
		_, _ = printer.Printf("\nConfirm call\n")
		_ = bcInst2.Confirm(blockedReceiver)
		_, _ = printer.Printf("\nConfirm call successful\n")

		mscEventInitialized := <-eventsChan.MSCInitializedChan
		_, _ = printer.Printf("\nMscEventInitialized : Cash sender - %s, Cash receiver - %s\n",
			mscEventInitialized.CashAlice.String(), mscEventInitialized.CashBob.String())

		mscBaseStateSigned, err := newConnFromAlice.NewMSCBaseStateRead()
		if err != nil {
			_, _ = printer.Printf("\nMsc base state read error - %v\n", err)
			return
		}

		bobID.SetCredentials(testKeystore, bobPassword)
		if err = mscBaseStateSigned.AddSign(bobID, channel.Receiver); err != nil {
			_, _ = printer.Printf("\nSigning msc base state error - %v\n", err)
			return
		}

		err = newConnFromAlice.NewMSCBaseStateRespond(mscBaseStateSigned, channel.MessageStatusAccept)
		if err != nil {
			_, _ = printer.Printf("\nMsc base state respond error - %v\n", err)
			return
		}

		err = newConnFromAlice.SetMSCBaseState(mscBaseStateSigned)
		if err != nil {
			_, _ = printer.Printf("\nMsc base state error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nMsc base state - %+v\n", mscBaseStateSigned)

		<-eventsChan.MSCStateRegisteringChan
		_, _ = printer.Printf("\nMscEventStateRegistering\n")

		_ = bcInst2.OwnerID.SetCredentials(testKeystore, bobPassword)
		_, _ = printer.Printf("\nState register call\n")
		_ = bcInst2.StateRegister(
			mscBaseStateSigned.MSContractBaseState.Sid,
			mscBaseStateSigned.MSContractBaseState.Version,
			mscBaseStateSigned.MSContractBaseState.BlockedSender,
			mscBaseStateSigned.MSContractBaseState.BlockedReceiver,
			mscBaseStateSigned.SignSender,
			mscBaseStateSigned.SignReceiver)
		_, _ = printer.Printf("\nState register call successful\n")

		mscEventStateRegistered := <-eventsChan.MSCStateRegisteredChan
		_, _ = printer.Printf("\nMscEventStateRegistered - Cash sender - %s, Cash receiver - %s\n",
			mscEventStateRegistered.BlockedAlice.String(), mscEventStateRegistered.BlockedBob.String())

		vpcStateID := channel.VPCStateID{
			AddSender:    aliceEthereumAddr,
			AddrReceiver: bobEthereumAddr,
			SID:          newSid.SidComplete,
		}

		//VPC State version 1 - by alice
		//Read, Sign and send back the vpc state with accept
		vpcStateSigned01, err := newConnFromAlice.NewVPCStateRead()
		if err != nil {
			_, _ = printer.Printf("\nVpc state read error - %v\n", err)
			return
		}
		bobID.SetCredentials(testKeystore, bobPassword)
		if err = vpcStateSigned01.AddSign(bobID, channel.Receiver); err != nil {
			_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
			return
		}
		err = newConnFromAlice.NewVPCStateRespond(vpcStateSigned01, channel.MessageStatusAccept)
		if err != nil {
			_, _ = printer.Printf("\nvpc state respond error= %v\n", err)
			return
		}

		err = newConnFromAlice.SetCurrentVPCState(vpcStateSigned01)
		if err != nil {
			_, _ = printer.Printf("\nVpc state error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned01)

		//VPC State version 2 - by bob
		vpcStateSignedPartial02 := channel.VPCStateSigned{
			VPCState: channel.VPCState{
				ID:              vpcStateID.SoliditySHA3(),
				Version:         big.NewInt(2),
				BlockedSender:   types.EtherToWei(big.NewInt(8)),
				BlockedReceiver: types.EtherToWei(big.NewInt(12)),
			},
		}
		bobID.SetCredentials(testKeystore, bobPassword)
		if err = vpcStateSignedPartial02.AddSign(bobID, channel.Receiver); err != nil {
			_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
			return
		}
		vpcStateSigned02, status, err := newConnFromAlice.NewVPCStateRequest(vpcStateSignedPartial02)
		if err != nil {
			_, _ = printer.Printf("\nvpc state request error %v\n", err)
			return
		}
		_ = status

		err = newConnFromAlice.SetCurrentVPCState(vpcStateSigned02)
		if err != nil {
			_, _ = printer.Printf("\nVpc state error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned02)

		//VPC State version 3 - by bob
		vpcStateSignedPartial03 := channel.VPCStateSigned{
			VPCState: channel.VPCState{
				ID:              vpcStateID.SoliditySHA3(),
				Version:         big.NewInt(3),
				BlockedSender:   types.EtherToWei(big.NewInt(7)),
				BlockedReceiver: types.EtherToWei(big.NewInt(13)),
			},
		}
		bobID.SetCredentials(testKeystore, bobPassword)
		if err = vpcStateSignedPartial03.AddSign(bobID, channel.Receiver); err != nil {
			_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
			return
		}
		vpcStateSigned03, _, err := newConnFromAlice.NewVPCStateRequest(vpcStateSignedPartial03)
		if err != nil {
			_, _ = printer.Printf("\nvpc state request error %v\n", err)
			return
		}

		err = newConnFromAlice.SetCurrentVPCState(vpcStateSigned03)
		if err != nil {
			_, _ = printer.Printf("\nVpc state error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned03)

		//VPC State version 4 - by alice
		//Read, Sign and send back the vpc state with accept
		vpcStateSigned04, err := newConnFromAlice.NewVPCStateRead()
		if err != nil {
			_, _ = printer.Printf("\nVpc state read error - %v\n", err)
			return
		}
		bobID.SetCredentials(testKeystore, bobPassword)
		if err = vpcStateSigned04.AddSign(bobID, channel.Receiver); err != nil {
			_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
			return
		}
		err = newConnFromAlice.NewVPCStateRespond(vpcStateSigned04, channel.MessageStatusAccept)
		if err != nil {
			_, _ = printer.Printf("\nvpc state respond error= %v\n", err)
			return
		}

		err = newConnFromAlice.SetCurrentVPCState(vpcStateSigned04)
		if err != nil {
			_, _ = printer.Printf("\nVpc state error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned04)

		//VPC State version 5 - by bob
		vpcStateSignedPartial05 := channel.VPCStateSigned{
			VPCState: channel.VPCState{
				ID:              vpcStateID.SoliditySHA3(),
				Version:         big.NewInt(5),
				BlockedSender:   types.EtherToWei(big.NewInt(5)),
				BlockedReceiver: types.EtherToWei(big.NewInt(15)),
			},
		}
		bobID.SetCredentials(testKeystore, bobPassword)
		if err = vpcStateSignedPartial05.AddSign(bobID, channel.Receiver); err != nil {
			_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
			return
		}
		vpcStateSigned05, _, err := newConnFromAlice.NewVPCStateRequest(vpcStateSignedPartial05)
		if err != nil {
			_, _ = printer.Printf("\nvpc state request error %v\n", err)
			return
		}

		err = newConnFromAlice.SetCurrentVPCState(vpcStateSigned05)
		if err != nil {
			_, _ = printer.Printf("\nVpc state error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned05)

		//VPC State version 6 - by alice
		//Read, Sign and send back the vpc state with accept
		vpcStateSigned06, err := newConnFromAlice.NewVPCStateRead()
		if err != nil {
			_, _ = printer.Printf("\nVpc state read error - %v\n", err)
			return
		}
		bobID.SetCredentials(testKeystore, bobPassword)
		if err = vpcStateSigned06.AddSign(bobID, channel.Receiver); err != nil {
			_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
			return
		}
		err = newConnFromAlice.NewVPCStateRespond(vpcStateSigned06, channel.MessageStatusAccept)
		if err != nil {
			_, _ = printer.Printf("\nvpc state respond error= %v\n", err)
			return
		}

		err = newConnFromAlice.SetCurrentVPCState(vpcStateSigned06)
		if err != nil {
			_, _ = printer.Printf("\nVpc state error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned06)

		//VPC State version 7 - by bob
		vpcStateSignedPartial07 := channel.VPCStateSigned{
			VPCState: channel.VPCState{
				ID:              vpcStateID.SoliditySHA3(),
				Version:         big.NewInt(7),
				BlockedSender:   types.EtherToWei(big.NewInt(3)),
				BlockedReceiver: types.EtherToWei(big.NewInt(17)),
			},
		}
		bobID.SetCredentials(testKeystore, bobPassword)
		if err = vpcStateSignedPartial07.AddSign(bobID, channel.Receiver); err != nil {
			_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
			return
		}
		vpcStateSigned07, _, err := newConnFromAlice.NewVPCStateRequest(vpcStateSignedPartial07)
		if err != nil {
			_, _ = printer.Printf("\nvpc state request error %v\n", err)
			return
		}

		err = newConnFromAlice.SetCurrentVPCState(vpcStateSigned07)
		if err != nil {
			_, _ = printer.Printf("\nVpc state error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned07)

		//VPC State version 8 - by alice
		//Read, Sign and send back the vpc state with accept
		vpcStateSigned08, err := newConnFromAlice.NewVPCStateRead()
		if err != nil {
			_, _ = printer.Printf("\nVpc state read error - %v\n", err)
			return
		}
		bobID.SetCredentials(testKeystore, bobPassword)
		if err = vpcStateSigned08.AddSign(bobID, channel.Receiver); err != nil {
			_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
			return
		}
		err = newConnFromAlice.NewVPCStateRespond(vpcStateSigned08, channel.MessageStatusAccept)
		if err != nil {
			_, _ = printer.Printf("\nvpc state respond error= %v\n", err)
			return
		}

		err = newConnFromAlice.SetCurrentVPCState(vpcStateSigned08)
		if err != nil {
			_, _ = printer.Printf("\nVpc state error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned08)

		//VPC State version 9 - by alice
		//Read, Sign and send back the vpc state with accept
		vpcStateSigned09, err := newConnFromAlice.NewVPCStateRead()
		if err != nil {
			_, _ = printer.Printf("\nVpc state read error - %v\n", err)
			return
		}
		bobID.SetCredentials(testKeystore, bobPassword)
		if err = vpcStateSigned09.AddSign(bobID, channel.Receiver); err != nil {
			_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
			return
		}
		err = newConnFromAlice.NewVPCStateRespond(vpcStateSigned09, channel.MessageStatusAccept)
		if err != nil {
			_, _ = printer.Printf("\nvpc state respond error= %v\n", err)
			return
		}

		err = newConnFromAlice.SetCurrentVPCState(vpcStateSigned09)
		if err != nil {
			_, _ = printer.Printf("\nVpc state error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned09)

		//VPC State version 10 - by bob
		vpcStateSignedPartial10 := channel.VPCStateSigned{
			VPCState: channel.VPCState{
				ID:              vpcStateID.SoliditySHA3(),
				Version:         big.NewInt(10),
				BlockedSender:   types.EtherToWei(big.NewInt(0)),
				BlockedReceiver: types.EtherToWei(big.NewInt(20)),
			},
		}
		bobID.SetCredentials(testKeystore, bobPassword)
		if err = vpcStateSignedPartial10.AddSign(bobID, channel.Receiver); err != nil {
			_, _ = printer.Printf("\nSigning vpc state error - %v\n", err)
			return
		}
		vpcStateSigned10, _, err := newConnFromAlice.NewVPCStateRequest(vpcStateSignedPartial10)
		if err != nil {
			_, _ = printer.Printf("\nVpc state request error %v\n", err)
			return
		}

		err = newConnFromAlice.SetCurrentVPCState(vpcStateSigned10)
		if err != nil {
			_, _ = printer.Printf("\nVpc state error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nVpc state - %+v\n", vpcStateSigned10)

		vpcClosing := <-eventsChan.VPCVPCClosingChan
		_, _ = printer.Printf("\nVpcEventVpcClosing : id - %x\n", vpcClosing.Id)
		states, err := bcInst2.States()
		if err != nil {
			_, _ = printer.Printf("Retrieve vpc states error - %v", err)
		}
		_, _ = printer.Printf("closing state - %+v\n", states)

		vpcClosingState := vpcStateSigned10

		_ = bcInst2.OwnerID.SetCredentials(testKeystore, bobPassword)
		_, _ = printer.Printf("\nVPCClose call with state %+v\n", vpcClosingState.VPCState)
		_ = bcInst2.VPCClose(
			vpcStateID.SID, vpcClosingState.VPCState.Version,
			aliceEthereumAddr, bobEthereumAddr,
			vpcClosingState.VPCState.BlockedSender,
			vpcClosingState.VPCState.BlockedReceiver,
			vpcClosingState.SignSender, vpcClosingState.SignReceiver)
		_, _ = printer.Printf("\nVPCClose call successful\n")

		vpcClosed := <-eventsChan.VPCVPCClosedChan
		_, _ = printer.Printf("\nVpcEventVpcClosed : id - %x, Cash sender - %s, Cash receiver -%s\n",
			vpcClosed.Id, vpcClosed.CashAlice.String(), vpcClosed.CashBob.String())
		states, err = bcInst2.States()
		if err != nil {
			_, _ = printer.Printf("Retrieve vpc states error - %v", err)
		}
		_, _ = printer.Printf("closed with state - %+v\n", states)

		<-eventsChan.MSCClosedChan
		_, _ = printer.Printf("\nMscEventClosed\n")

		err = newConnFromAlice.Close()
		if err != nil {
			_, _ = printer.Printf("\nClose channel error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nClose channel successful\n")

		// waitTimeForListenerShutdown := 10 * time.Second
		// ctx, _ := context.WithTimeout(context.Background(), waitTimeForListenerShutdown)
		// if err != nil {
		// 	_, _ = printer.Printf("\nMaking context error - %v\n", err)
		// 	return
		// }
		ctx := context.Background()
		err = listener.Shutdown(ctx)
		if err != nil {
			_, _ = printer.Printf("\nListener shutdown error - %v\n", err)
			return
		}
		_, _ = printer.Printf("\nChannel listener closed\n")

	}(&bcInst2, incomingConnChan, wg)

	return nil
}
