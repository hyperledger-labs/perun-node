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

package channel

import (
	"bytes"
	"fmt"

	"github.com/direct-state-transfer/dst-go/channel/primitives"
	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
)

// IdentityRequest sends an identity request and waits for identity response from the peer node.
// If response is successfully received, it returns the peer id in the response message.
func (ch *Instance) IdentityRequest(selfID identity.OffChainID) (peerID identity.OffChainID, err error) {

	idRequestMsg := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgIdentityRequest,
		Message:   primitives.JSONMsgIdentity{ID: selfID},
	}

	err = ch.adapter.Write(idRequestMsg)
	if err != nil {
		return peerID, err
	}

	response, err := ch.adapter.Read()
	if err != nil {
		return peerID, err
	}

	if response.MessageID != primitives.MsgIdentityResponse {
		errMsg := ("Invalid response received for id request")
		return peerID, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(primitives.JSONMsgIdentity)
	if !ok {
		errMsg := ("Message packet type error")
		return peerID, fmt.Errorf(errMsg)
	}

	peerID = msg.ID
	return peerID, nil
}

// IdentityRead reads the identity request sent by the peer node and returns the peer id in the message.
func (ch *Instance) IdentityRead() (peerID identity.OffChainID, err error) {

	msg, err := ch.adapter.Read()
	if err != nil {
		errMsg := "Error waiting for id request - connection dropped -" + err.Error()
		return peerID, fmt.Errorf(errMsg)
	}

	if msg.MessageID != primitives.MsgIdentityRequest {
		errMsg := "First message is not id request"
		return peerID, fmt.Errorf(errMsg)
	}

	idRequestMsg, ok := msg.Message.(primitives.JSONMsgIdentity)
	if !ok {
		errMsg := ("Message packet type error")
		return peerID, fmt.Errorf(errMsg)
	}

	peerID = idRequestMsg.ID
	return peerID, nil
}

// IdentityRespond sends an identity response to the peer node with self id in the message.
func (ch *Instance) IdentityRespond(selfID identity.OffChainID) (err error) {

	selfIDMsg := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgIdentityResponse,
		Message:   primitives.JSONMsgIdentity{ID: selfID},
	}
	err = ch.adapter.Write(selfIDMsg)
	if err != nil {
		errMsg := "Error responding to id request" + err.Error()
		return fmt.Errorf(errMsg)
	}
	return nil
}

// NewChannelRequest sends an new channel request and waits for new channel response from the peer node.
// If response is successfully received, it returns the acceptance status and reason in the response message.
func (ch *Instance) NewChannelRequest(msgProtocolVersion string, contractStoreVersion []byte) (accept primitives.MessageStatus, reason string, err error) {

	idRequestMsg := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgNewChannelRequest,
		Message: primitives.JSONMsgNewChannel{
			MsgProtocolVersion:   msgProtocolVersion,
			ContractStoreVersion: contractStoreVersion,
			Status:               primitives.MessageStatusRequire,
		},
	}
	logger.Debug("Requesting new channel with the other node")
	err = ch.adapter.Write(idRequestMsg)
	if err != nil {
		return primitives.MessageStatusUnknown, "", err
	}

	response, err := ch.adapter.Read()
	if err != nil {
		return primitives.MessageStatusUnknown, "", err
	}

	if response.MessageID != primitives.MsgNewChannelResponse {
		errMsg := ("Invalid response received for id request")
		return primitives.MessageStatusUnknown, "", fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(primitives.JSONMsgNewChannel)
	if !ok {
		errMsg := ("Message packet type error")
		return primitives.MessageStatusUnknown, "", fmt.Errorf(errMsg)
	}

	if !bytes.Equal(contractStoreVersion, msg.ContractStoreVersion) {
		errMsg := ("Contract store version modified by peer")
		return primitives.MessageStatusUnknown, "", fmt.Errorf(errMsg)
	}

	if msgProtocolVersion != msg.MsgProtocolVersion {
		errMsg := ("Message protocol version modified by peer")
		return primitives.MessageStatusUnknown, "", fmt.Errorf(errMsg)
	}

	accept = msg.Status
	reason = msg.Reason

	return accept, reason, nil
}

// NewChannelRead reads the new channel request sent by the peer node and returns the message protocol version and contract store version in the message.
func (ch *Instance) NewChannelRead() (msgProtocolVersion string, contractStoreVersion []byte, err error) {
	logger.Debug("Reading new channel request from other node")
	response, err := ch.adapter.Read()
	if err != nil {
		return "", contractStoreVersion, err
	}

	if response.MessageID != primitives.MsgNewChannelRequest {
		errMsg := ("Invalid response received for id request")
		return "", contractStoreVersion, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(primitives.JSONMsgNewChannel)
	if !ok {
		errMsg := ("Message packet type error")
		return "", contractStoreVersion, fmt.Errorf(errMsg)
	}

	if !primitives.ContainsStatus(primitives.RequestStatusList, msg.Status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", msg.Status, primitives.RequestStatusList)
		return "", contractStoreVersion, fmt.Errorf(errMsg)
	}

	return msg.MsgProtocolVersion, msg.ContractStoreVersion, nil
}

// NewChannelRespond sends an new channel response to the peer node with acceptance status in the message.
func (ch *Instance) NewChannelRespond(msgProtocolVersion string, contractStoreVersion []byte, accept primitives.MessageStatus, reason string) (err error) {

	responsePkt := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgNewChannelResponse,
		Message: primitives.JSONMsgNewChannel{
			MsgProtocolVersion:   msgProtocolVersion,
			ContractStoreVersion: contractStoreVersion,
			Status:               accept,
			Reason:               reason,
		},
	}
	logger.Debug("Sending response to new channel request")
	err = ch.adapter.Write(responsePkt)
	return err
}

// SessionIDRequest sends an sessiod id request with partial session id and waits for sessiod id response from the peer node.
// If response is successfully received, it returns the complete sid and acceptance status in the response message.
func (ch *Instance) SessionIDRequest(sid primitives.SessionID) (gotSid primitives.SessionID, status primitives.MessageStatus, err error) {

	idRequestMsg := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgSessionIDRequest,
		Message: primitives.JSONMsgSessionID{
			Sid:    sid,
			Status: primitives.MessageStatusRequire,
		},
	}
	logger.Debug("Requesting session ID")
	err = ch.adapter.Write(idRequestMsg)
	if err != nil {
		return gotSid, "", err
	}

	response, err := ch.adapter.Read()
	if err != nil {
		return gotSid, "", err
	}

	if response.MessageID != primitives.MsgSessionIDResponse {
		errMsg := ("Invalid response received for id request")
		return gotSid, "", fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(primitives.JSONMsgSessionID)
	if !ok {
		errMsg := ("Message packet type error")
		return gotSid, "", fmt.Errorf(errMsg)
	}

	if !sid.EqualSender(msg.Sid) {
		errMsg := ("Sid sender components modified by peer")
		return gotSid, "", fmt.Errorf(errMsg)
	}

	return msg.Sid, msg.Status, nil
}

// SessionIDRead reads the session id request sent by the peer node and returns the session id in the message.
func (ch *Instance) SessionIDRead() (sid primitives.SessionID, err error) {
	logger.Debug("Reading session ID request")
	response, err := ch.adapter.Read()
	if err != nil {
		return sid, err
	}

	if response.MessageID != primitives.MsgSessionIDRequest {
		errMsg := ("Invalid response received instead of session id request")
		return sid, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(primitives.JSONMsgSessionID)
	if !ok {
		errMsg := ("Message packet type error")
		return sid, fmt.Errorf(errMsg)
	}

	if !primitives.ContainsStatus(primitives.RequestStatusList, msg.Status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", msg.Status, primitives.RequestStatusList)
		return sid, fmt.Errorf(errMsg)
	}

	return msg.Sid, nil

}

// SessionIDRespond sends an session id response to the peer node with complete session id (optional) and acceptance status in the message.
func (ch *Instance) SessionIDRespond(sid primitives.SessionID, status primitives.MessageStatus) (err error) {

	if !primitives.ContainsStatus(primitives.ResponseStatusList, status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", status, primitives.ResponseStatusList)
		return fmt.Errorf(errMsg)
	}

	idRequestMsg := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgSessionIDResponse,
		Message: primitives.JSONMsgSessionID{
			Sid:    sid,
			Status: status,
		},
	}
	logger.Debug("Responding to session ID request")
	err = ch.adapter.Write(idRequestMsg)
	return err
}

// ContractAddrRequest sends a contract id request with details of deployed contract and waits for contract address response from the peer node.
// If response is successfully received, it returns the acceptance status in the response message.
func (ch *Instance) ContractAddrRequest(addr types.Address, id contract.Handler) (status primitives.MessageStatus, err error) {

	idRequestMsg := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgContractAddrRequest,
		Message: primitives.JSONMsgContractAddr{
			Addr:         addr,
			ContractType: id,
			Status:       primitives.MessageStatusRequire,
		},
	}
	logger.Debug("Requesting Contract Address")
	err = ch.adapter.Write(idRequestMsg)
	if err != nil {
		return "", err
	}

	response, err := ch.adapter.Read()
	if err != nil {
		return "", err
	}

	if response.MessageID != primitives.MsgContractAddrResponse {
		errMsg := ("Invalid response received for id request")
		return "", fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(primitives.JSONMsgContractAddr)
	if !ok {
		errMsg := ("Message packet type error")
		return "", fmt.Errorf(errMsg)
	}

	if !id.Equal(msg.ContractType) {
		errMsg := ("Contract handler modified by peer")
		return "", fmt.Errorf(errMsg)
	}

	return msg.Status, nil
}

// ContractAddrRead reads the contract address request sent by the peer node and returns the contract address and handler in the message.
func (ch *Instance) ContractAddrRead() (addr types.Address, id contract.Handler, err error) {
	logger.Debug("Reading Contract Address request")
	response, err := ch.adapter.Read()
	if err != nil {
		return addr, id, err
	}

	if response.MessageID != primitives.MsgContractAddrRequest {
		errMsg := ("Invalid response received for id request")
		return addr, id, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(primitives.JSONMsgContractAddr)
	if !ok {
		errMsg := ("Message packet type error")
		return addr, id, fmt.Errorf(errMsg)
	}

	if !primitives.ContainsStatus(primitives.RequestStatusList, msg.Status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", msg.Status, primitives.RequestStatusList)
		return addr, id, fmt.Errorf(errMsg)
	}

	return msg.Addr, msg.ContractType, nil

}

// ContractAddrRespond sends an contract address response to the peer node with acceptance status in the message.
func (ch *Instance) ContractAddrRespond(addr types.Address, id contract.Handler, status primitives.MessageStatus) (err error) {

	if !primitives.ContainsStatus(primitives.ResponseStatusList, status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", status, primitives.ResponseStatusList)
		return fmt.Errorf(errMsg)
	}

	idRequestMsg := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgContractAddrResponse,
		Message: primitives.JSONMsgContractAddr{
			Addr:         addr,
			ContractType: id,
			Status:       status,
		},
	}
	logger.Debug("Responding to Contract Address request")
	err = ch.adapter.Write(idRequestMsg)
	return err
}

// NewMSCBaseStateRequest sends a new msc base request with partial signature and waits for msc base state response from the peer node.
// If response is successfully received, it returns the fully signed msc base state and acceptance status in the response message.
func (ch *Instance) NewMSCBaseStateRequest(newSignedState primitives.MSCBaseStateSigned) (responseState primitives.MSCBaseStateSigned, status primitives.MessageStatus, err error) {

	requestMsg := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgMSCBaseStateRequest,
		Message: primitives.JSONMsgMSCBaseState{
			SignedStateVal: newSignedState,
			Status:         primitives.MessageStatusRequire,
		},
	}
	logger.Debug("Requesting new MSC base state")
	err = ch.adapter.Write(requestMsg)
	if err != nil {
		return responseState, "", err
	}

	response, err := ch.adapter.Read()
	if err != nil {
		return responseState, "", err
	}

	if response.MessageID != primitives.MsgMSCBaseStateResponse {
		errMsg := ("Invalid response received for msc base state request")
		return responseState, "", fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(primitives.JSONMsgMSCBaseState)
	if !ok {
		errMsg := ("Message packet type error")
		return responseState, "", fmt.Errorf(errMsg)
	}

	if !newSignedState.MSContractBaseState.Equal(msg.SignedStateVal.MSContractBaseState) {
		errMsg := ("MSContract base state modified by peer")
		return responseState, "", fmt.Errorf(errMsg)
	}

	return msg.SignedStateVal, msg.Status, nil
}

// NewMSCBaseStateRead reads the new msc base state request sent by the peer node and returns the msc base state in the message.
func (ch *Instance) NewMSCBaseStateRead() (state primitives.MSCBaseStateSigned, err error) {
	logger.Debug("Reading new MSC base state request")
	response, err := ch.adapter.Read()
	if err != nil {
		return state, err
	}

	if response.MessageID != primitives.MsgMSCBaseStateRequest {
		errMsg := ("Invalid response received for msc base state request")
		return state, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(primitives.JSONMsgMSCBaseState)
	if !ok {
		errMsg := ("Message packet type error")
		return state, fmt.Errorf(errMsg)
	}

	if !primitives.ContainsStatus(primitives.RequestStatusList, msg.Status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", msg.Status, primitives.RequestStatusList)
		return state, fmt.Errorf(errMsg)
	}

	return msg.SignedStateVal, nil
}

// NewMSCBaseStateRespond sends an msc base state response to the peer node with fully signed state (optional) and acceptance status in the message.
func (ch *Instance) NewMSCBaseStateRespond(state primitives.MSCBaseStateSigned, status primitives.MessageStatus) (err error) {

	if !primitives.ContainsStatus(primitives.ResponseStatusList, status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", status, primitives.ResponseStatusList)
		return fmt.Errorf(errMsg)
	}

	response := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgMSCBaseStateResponse,
		Message: primitives.JSONMsgMSCBaseState{
			SignedStateVal: state,
			Status:         status,
		},
	}
	logger.Debug("Responding to new MSC base state request")
	err = ch.adapter.Write(response)
	return err
}

// NewVPCStateRequest sends a new vpc request with partial signature and waits for vpc state response from the peer node.
// If response is successfully received, it returns the fully signed vpc state and acceptance status in the response message.
func (ch *Instance) NewVPCStateRequest(newStateSigned primitives.VPCStateSigned) (responseState primitives.VPCStateSigned, status primitives.MessageStatus, err error) {

	requestMsg := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgVPCStateRequest,
		Message: primitives.JSONMsgVPCState{
			SignedStateVal: newStateSigned,
			Status:         primitives.MessageStatusRequire,
		},
	}
	logger.Debug("Requesting new VPC state")
	err = ch.adapter.Write(requestMsg)
	if err != nil {
		return responseState, "", err
	}

	response, err := ch.adapter.Read()
	if err != nil {
		return responseState, "", err
	}

	if response.MessageID != primitives.MsgVPCStateResponse {
		errMsg := ("Invalid response received for vpc state request")
		return responseState, "", fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(primitives.JSONMsgVPCState)
	if !ok {
		errMsg := ("Message packet type error")
		return responseState, "", fmt.Errorf(errMsg)
	}

	if !newStateSigned.VPCState.Equal(msg.SignedStateVal.VPCState) {
		errMsg := ("VPC state modified by peer")
		return responseState, "", fmt.Errorf(errMsg)
	}

	return msg.SignedStateVal, msg.Status, nil
}

// NewVPCStateRead reads the new vpc state request sent by the peer node and returns the vpc state in the message.
func (ch *Instance) NewVPCStateRead() (state primitives.VPCStateSigned, err error) {
	logger.Debug("Reading new VPC state request")
	response, err := ch.adapter.Read()
	if err != nil {
		return state, err
	}

	if response.MessageID != primitives.MsgVPCStateRequest {
		errMsg := ("Invalid response received for vpc state request")
		return state, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(primitives.JSONMsgVPCState)
	if !ok {
		errMsg := ("Message packet type error")
		return state, fmt.Errorf(errMsg)
	}

	if !primitives.ContainsStatus(primitives.RequestStatusList, msg.Status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", msg.Status, primitives.RequestStatusList)
		return state, fmt.Errorf(errMsg)
	}

	return msg.SignedStateVal, nil
}

// NewVPCStateRespond sends an vpc state response to the peer node with fully signed state (optional) and acceptance status in the message.
func (ch *Instance) NewVPCStateRespond(state primitives.VPCStateSigned, status primitives.MessageStatus) (err error) {

	if !primitives.ContainsStatus(primitives.ResponseStatusList, status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", status, primitives.ResponseStatusList)
		return fmt.Errorf(errMsg)
	}

	response := primitives.ChMsgPkt{
		Version:   primitives.Version,
		MessageID: primitives.MsgVPCStateResponse,
		Message: primitives.JSONMsgVPCState{
			SignedStateVal: state,
			Status:         status,
		},
	}
	logger.Debug("Responding to new VPC state request")
	err = ch.adapter.Write(response)
	return err
}
