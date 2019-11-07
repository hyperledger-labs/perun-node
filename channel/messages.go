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
	"encoding/json"
	"fmt"
	"time"

	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
)

// Version defines the version of offchain messaging protocol.
const Version = "0.1"

// MessageID is the unique id for the channel message format.
type MessageID string

type chRawMsgPkt struct {
	Version   string          `json:"version"`
	MessageID MessageID       `json:"message_id"`
	Message   json.RawMessage `json:"message"`
	Timestamp time.Time       `json:"timestamp"`
}

type chMsgPkt struct {
	Version   string      `json:"version"`
	MessageID MessageID   `json:"message_id"`
	Message   interface{} `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
}

// MessageStatus represents the status in a request-response message.
type MessageStatus string

// Enumeration of allowed values for message status
var (
	MessageStatusRequire = MessageStatus("require")
	MessageStatusAccept  = MessageStatus("accept")
	MessageStatusDecline = MessageStatus("decline")
	MessageStatusUnknown = MessageStatus("unknown")

	RequestStatusList  = []MessageStatus{MessageStatusRequire}
	ResponseStatusList = []MessageStatus{MessageStatusAccept, MessageStatusDecline}
)

// Enumeration of allowed values for message id.
const (
	// MsgIdentityRequest is the id for "identity request" message.
	MsgIdentityRequest MessageID = "MsgIdentityRequest"

	// MsgIdentityResponse is the id for "identity response" message.
	MsgIdentityResponse MessageID = "MsgIdentityResponse"

	// MsgNewChannelRequest is the id for "new channel request" message.
	MsgNewChannelRequest MessageID = "MsgNewChannelRequest"

	// MsgNewChannelResponse is the id for "new channel response" message.
	MsgNewChannelResponse MessageID = "MsgNewChannelResponse"

	// MsgSessionIDRequest is the id for "session id request" message.
	MsgSessionIDRequest MessageID = "MsgSessionIdRequest"

	// MsgSessionIDResponse is the id for "session id response" message.
	MsgSessionIDResponse MessageID = "MsgSessionIdResponse"

	// MsgContractAddrRequest is the id for "contract address request" message.
	MsgContractAddrRequest MessageID = "MsgContractAddrRequest"

	// MsgContractAddrResponse is the id for "contract address response" message.
	MsgContractAddrResponse MessageID = "MsgContractAddrResponse"

	// MsgMSCBaseStateRequest is the id for "MSC base state request" message.
	MsgMSCBaseStateRequest MessageID = "MsgMSCBaseStateRequest"

	// MsgMSCBaseStateResponse is the id for "MSC base state response" message.
	MsgMSCBaseStateResponse MessageID = "MsgMSCBaseStateResponse"

	// MsgVPCStateRequest is the id for "VPC state request" message.
	MsgVPCStateRequest MessageID = "MsgVPCStateRequest"

	// MsgVPCStateResponse is the id for "VPC state response" message.
	MsgVPCStateResponse MessageID = "MsgVPCStateResponse"
)

type jsonMsgIdentity struct {
	ID identity.OffChainID `json:"id"`
}

type jsonMsgNewChannel struct {
	//TODO : Check if contract lock in amount also need to be added
	ContractStoreVersion []byte        `json:"contract_store_version"`
	MsgProtocolVersion   string        `json:"msg_protocol_version"`
	Status               MessageStatus `json:"status"`
	Reason               string        `json:"reason"`
}

type jsonMsgSessionID struct {
	Sid    SessionID     `json:"sid"`
	Status MessageStatus `json:"status"`
}

type jsonMsgContractAddr struct {
	Addr         types.Address    `json:"addr"`
	ContractType contract.Handler `json:"contract_type"`
	Status       MessageStatus    `json:"status"`
}

type jsonMsgMSCBaseState struct {
	SignedStateVal MSCBaseStateSigned `json:"signed_state_val"`
	Status         MessageStatus      `json:"status"`
}

type jsonMsgVPCState struct {
	SignedStateVal VPCStateSigned `json:"signed_state_val"`
	Status         MessageStatus  `json:"status"`
}

// UnmarshalJSON implements json.Unmarshaller interface.
//
// The json message is first unmarshalled retaining the message as raw json.
// Then the message is unmarshalled to appropriate format depending upon the message id.
func (msgPkt *chMsgPkt) UnmarshalJSON(data []byte) (err error) {

	//Unmarshal only the chMsgPkt, retaining the message as rawJSON
	var rawMsgPkt = chRawMsgPkt{}
	if err = json.Unmarshal(data, &rawMsgPkt); err != nil {
		return err
	}

	//Unmarshal the message to appropriate format depending upon message id
	switch rawMsgPkt.MessageID {
	case MsgIdentityRequest, MsgIdentityResponse:

		var msg jsonMsgIdentity
		if err = json.Unmarshal(rawMsgPkt.Message, &msg); err != nil {
			return err
		}
		msgPkt.Message = msg

	case MsgNewChannelRequest, MsgNewChannelResponse:

		var msg jsonMsgNewChannel
		if err = json.Unmarshal(rawMsgPkt.Message, &msg); err != nil {
			return err
		}
		msgPkt.Message = msg

	case MsgSessionIDRequest, MsgSessionIDResponse:
		var msg jsonMsgSessionID
		if err = json.Unmarshal(rawMsgPkt.Message, &msg); err != nil {
			return err
		}
		msgPkt.Message = msg

	case MsgContractAddrRequest, MsgContractAddrResponse:
		var msg jsonMsgContractAddr
		if err = json.Unmarshal(rawMsgPkt.Message, &msg); err != nil {
			return err
		}
		msgPkt.Message = msg

	case MsgMSCBaseStateRequest, MsgMSCBaseStateResponse:
		var msg jsonMsgMSCBaseState
		if err = json.Unmarshal(rawMsgPkt.Message, &msg); err != nil {
			return err
		}
		msgPkt.Message = msg

	case MsgVPCStateRequest, MsgVPCStateResponse:
		var msg jsonMsgVPCState
		if err = json.Unmarshal(rawMsgPkt.Message, &msg); err != nil {
			return err
		}
		msgPkt.Message = msg

	default:
		err = fmt.Errorf("Unsupported message id - " + string(rawMsgPkt.MessageID))
		return err
	}

	//Assign other properties from rawMsgPkt to parsed msgPkt
	msgPkt.Version = rawMsgPkt.Version
	msgPkt.MessageID = rawMsgPkt.MessageID
	msgPkt.Timestamp = rawMsgPkt.Timestamp

	return nil
}

// IdentityRequest sends an identity request and waits for identity response from the peer node.
// If response is successfully received, it returns the peer id in the response message.
func (ch *Instance) IdentityRequest(selfID identity.OffChainID) (peerID identity.OffChainID, err error) {

	idRequestMsg := chMsgPkt{
		Version:   Version,
		MessageID: MsgIdentityRequest,
		Message:   jsonMsgIdentity{selfID},
	}

	err = ch.adapter.Write(idRequestMsg)
	if err != nil {
		return peerID, err
	}

	response, err := ch.adapter.Read()
	if err != nil {
		return peerID, err
	}

	if response.MessageID != MsgIdentityResponse {
		errMsg := ("Invalid response received for id request")
		return peerID, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(jsonMsgIdentity)
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

	if msg.MessageID != MsgIdentityRequest {
		errMsg := "First message is not id request"
		return peerID, fmt.Errorf(errMsg)
	}

	idRequestMsg, ok := msg.Message.(jsonMsgIdentity)
	if !ok {
		errMsg := ("Message packet type error")
		return peerID, fmt.Errorf(errMsg)
	}

	peerID = idRequestMsg.ID
	return peerID, nil
}

// IdentityRespond sends an identity response to the peer node with self id in the message.
func (ch *Instance) IdentityRespond(selfID identity.OffChainID) (err error) {

	selfIDMsg := chMsgPkt{
		Version:   Version,
		MessageID: MsgIdentityResponse,
		Message:   jsonMsgIdentity{selfID},
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
func (ch *Instance) NewChannelRequest(msgProtocolVersion string, contractStoreVersion []byte) (accept MessageStatus, reason string, err error) {

	idRequestMsg := chMsgPkt{
		Version:   Version,
		MessageID: MsgNewChannelRequest,
		Message: jsonMsgNewChannel{
			MsgProtocolVersion:   msgProtocolVersion,
			ContractStoreVersion: contractStoreVersion,
			Status:               MessageStatusRequire,
		},
	}
	logger.Debug("Requesting new channel with the other node")
	err = ch.adapter.Write(idRequestMsg)
	if err != nil {
		return MessageStatusUnknown, "", err
	}

	response, err := ch.adapter.Read()
	if err != nil {
		return MessageStatusUnknown, "", err
	}

	if response.MessageID != MsgNewChannelResponse {
		errMsg := ("Invalid response received for id request")
		return MessageStatusUnknown, "", fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(jsonMsgNewChannel)
	if !ok {
		errMsg := ("Message packet type error")
		return MessageStatusUnknown, "", fmt.Errorf(errMsg)
	}

	if !bytes.Equal(contractStoreVersion, msg.ContractStoreVersion) {
		errMsg := ("Contract store version modified by peer")
		return MessageStatusUnknown, "", fmt.Errorf(errMsg)
	}

	if msgProtocolVersion != msg.MsgProtocolVersion {
		errMsg := ("Message protocol version modified by peer")
		return MessageStatusUnknown, "", fmt.Errorf(errMsg)
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

	if response.MessageID != MsgNewChannelRequest {
		errMsg := ("Invalid response received for id request")
		return "", contractStoreVersion, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(jsonMsgNewChannel)
	if !ok {
		errMsg := ("Message packet type error")
		return "", contractStoreVersion, fmt.Errorf(errMsg)
	}

	if !containsStatus(RequestStatusList, msg.Status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", msg.Status, RequestStatusList)
		return "", contractStoreVersion, fmt.Errorf(errMsg)
	}

	return msg.MsgProtocolVersion, msg.ContractStoreVersion, nil
}

// NewChannelRespond sends an new channel response to the peer node with acceptance status in the message.
func (ch *Instance) NewChannelRespond(msgProtocolVersion string, contractStoreVersion []byte, accept MessageStatus, reason string) (err error) {

	responsePkt := chMsgPkt{
		Version:   Version,
		MessageID: MsgNewChannelResponse,
		Message: jsonMsgNewChannel{
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
func (ch *Instance) SessionIDRequest(sid SessionID) (gotSid SessionID, status MessageStatus, err error) {

	idRequestMsg := chMsgPkt{
		Version:   Version,
		MessageID: MsgSessionIDRequest,
		Message: jsonMsgSessionID{
			Sid:    sid,
			Status: MessageStatusRequire,
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

	if response.MessageID != MsgSessionIDResponse {
		errMsg := ("Invalid response received for id request")
		return gotSid, "", fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(jsonMsgSessionID)
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
func (ch *Instance) SessionIDRead() (sid SessionID, err error) {
	logger.Debug("Reading session ID request")
	response, err := ch.adapter.Read()
	if err != nil {
		return sid, err
	}

	if response.MessageID != MsgSessionIDRequest {
		errMsg := ("Invalid response received instead of session id request")
		return sid, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(jsonMsgSessionID)
	if !ok {
		errMsg := ("Message packet type error")
		return sid, fmt.Errorf(errMsg)
	}

	if !containsStatus(RequestStatusList, msg.Status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", msg.Status, RequestStatusList)
		return sid, fmt.Errorf(errMsg)
	}

	return msg.Sid, nil

}

// SessionIDRespond sends an session id response to the peer node with complete session id (optional) and acceptance status in the message.
func (ch *Instance) SessionIDRespond(sid SessionID, status MessageStatus) (err error) {

	if !containsStatus(ResponseStatusList, status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", status, ResponseStatusList)
		return fmt.Errorf(errMsg)
	}

	idRequestMsg := chMsgPkt{
		Version:   Version,
		MessageID: MsgSessionIDResponse,
		Message: jsonMsgSessionID{
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
func (ch *Instance) ContractAddrRequest(addr types.Address, id contract.Handler) (status MessageStatus, err error) {

	idRequestMsg := chMsgPkt{
		Version:   Version,
		MessageID: MsgContractAddrRequest,
		Message: jsonMsgContractAddr{
			Addr:         addr,
			ContractType: id,
			Status:       MessageStatusRequire,
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

	if response.MessageID != MsgContractAddrResponse {
		errMsg := ("Invalid response received for id request")
		return "", fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(jsonMsgContractAddr)
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

	if response.MessageID != MsgContractAddrRequest {
		errMsg := ("Invalid response received for id request")
		return addr, id, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(jsonMsgContractAddr)
	if !ok {
		errMsg := ("Message packet type error")
		return addr, id, fmt.Errorf(errMsg)
	}

	if !containsStatus(RequestStatusList, msg.Status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", msg.Status, RequestStatusList)
		return addr, id, fmt.Errorf(errMsg)
	}

	return msg.Addr, msg.ContractType, nil

}

// ContractAddrRespond sends an contract address response to the peer node with acceptance status in the message.
func (ch *Instance) ContractAddrRespond(addr types.Address, id contract.Handler, status MessageStatus) (err error) {

	if !containsStatus(ResponseStatusList, status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", status, ResponseStatusList)
		return fmt.Errorf(errMsg)
	}

	idRequestMsg := chMsgPkt{
		Version:   Version,
		MessageID: MsgContractAddrResponse,
		Message: jsonMsgContractAddr{
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
func (ch *Instance) NewMSCBaseStateRequest(newSignedState MSCBaseStateSigned) (responseState MSCBaseStateSigned, status MessageStatus, err error) {

	requestMsg := chMsgPkt{
		Version:   Version,
		MessageID: MsgMSCBaseStateRequest,
		Message: jsonMsgMSCBaseState{
			SignedStateVal: newSignedState,
			Status:         MessageStatusRequire,
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

	if response.MessageID != MsgMSCBaseStateResponse {
		errMsg := ("Invalid response received for msc base state request")
		return responseState, "", fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(jsonMsgMSCBaseState)
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
func (ch *Instance) NewMSCBaseStateRead() (state MSCBaseStateSigned, err error) {
	logger.Debug("Reading new MSC base state request")
	response, err := ch.adapter.Read()
	if err != nil {
		return state, err
	}

	if response.MessageID != MsgMSCBaseStateRequest {
		errMsg := ("Invalid response received for msc base state request")
		return state, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(jsonMsgMSCBaseState)
	if !ok {
		errMsg := ("Message packet type error")
		return state, fmt.Errorf(errMsg)
	}

	if !containsStatus(RequestStatusList, msg.Status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", msg.Status, RequestStatusList)
		return state, fmt.Errorf(errMsg)
	}

	return msg.SignedStateVal, nil
}

// NewMSCBaseStateRespond sends an msc base state response to the peer node with fully signed state (optional) and acceptance status in the message.
func (ch *Instance) NewMSCBaseStateRespond(state MSCBaseStateSigned, status MessageStatus) (err error) {

	if !containsStatus(ResponseStatusList, status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", status, ResponseStatusList)
		return fmt.Errorf(errMsg)
	}

	response := chMsgPkt{
		Version:   Version,
		MessageID: MsgMSCBaseStateResponse,
		Message: jsonMsgMSCBaseState{
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
func (ch *Instance) NewVPCStateRequest(newStateSigned VPCStateSigned) (responseState VPCStateSigned, status MessageStatus, err error) {

	requestMsg := chMsgPkt{
		Version:   Version,
		MessageID: MsgVPCStateRequest,
		Message: jsonMsgVPCState{
			SignedStateVal: newStateSigned,
			Status:         MessageStatusRequire,
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

	if response.MessageID != MsgVPCStateResponse {
		errMsg := ("Invalid response received for vpc state request")
		return responseState, "", fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(jsonMsgVPCState)
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
func (ch *Instance) NewVPCStateRead() (state VPCStateSigned, err error) {
	logger.Debug("Reading new VPC state request")
	response, err := ch.adapter.Read()
	if err != nil {
		return state, err
	}

	if response.MessageID != MsgVPCStateRequest {
		errMsg := ("Invalid response received for vpc state request")
		return state, fmt.Errorf(errMsg)
	}

	msg, ok := response.Message.(jsonMsgVPCState)
	if !ok {
		errMsg := ("Message packet type error")
		return state, fmt.Errorf(errMsg)
	}

	if !containsStatus(RequestStatusList, msg.Status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", msg.Status, RequestStatusList)
		return state, fmt.Errorf(errMsg)
	}

	return msg.SignedStateVal, nil
}

// NewVPCStateRespond sends an vpc state response to the peer node with fully signed state (optional) and acceptance status in the message.
func (ch *Instance) NewVPCStateRespond(state VPCStateSigned, status MessageStatus) (err error) {

	if !containsStatus(ResponseStatusList, status) {
		errMsg := fmt.Sprintf("Invalid status received - %v. Use %v ", status, ResponseStatusList)
		return fmt.Errorf(errMsg)
	}

	response := chMsgPkt{
		Version:   Version,
		MessageID: MsgVPCStateResponse,
		Message: jsonMsgVPCState{
			SignedStateVal: state,
			Status:         status,
		},
	}
	logger.Debug("Responding to new VPC state request")
	err = ch.adapter.Write(response)
	return err
}

// containsStatus checks of the required value of staus is present in the list.
func containsStatus(list []MessageStatus, requiredValue MessageStatus) bool {
	for _, value := range list {
		if value == requiredValue {
			return true
		}
	}
	return false
}
