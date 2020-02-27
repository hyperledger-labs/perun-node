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

package primitives

import (
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

// ChMsgPkt is packet definition for offchain communications.
type ChMsgPkt struct {
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

// JSONMsgIdentity is the packet definition for "identity request|response" message
type JSONMsgIdentity struct {
	ID identity.OffChainID `json:"id"`
}

// JSONMsgNewChannel is the packet definition for "new channel request|response" message
type JSONMsgNewChannel struct {
	//TODO : Check if contract lock in amount also need to be added
	ContractStoreVersion []byte        `json:"contract_store_version"`
	MsgProtocolVersion   string        `json:"msg_protocol_version"`
	Status               MessageStatus `json:"status"`
	Reason               string        `json:"reason"`
}

// JSONMsgSessionID is the packet definition for "session id request|response" message
type JSONMsgSessionID struct {
	Sid    SessionID     `json:"sid"`
	Status MessageStatus `json:"status"`
}

// JSONMsgContractAddr is the packet definition for "contract address request|response" message
type JSONMsgContractAddr struct {
	Addr         types.Address    `json:"addr"`
	ContractType contract.Handler `json:"contract_type"`
	Status       MessageStatus    `json:"status"`
}

// JSONMsgMSCBaseState is the packet definition for "MSC base state request|response" message
type JSONMsgMSCBaseState struct {
	SignedStateVal MSCBaseStateSigned `json:"signed_state_val"`
	Status         MessageStatus      `json:"status"`
}

// JSONMsgVPCState is the packet definition for "VPC state request|response" message
type JSONMsgVPCState struct {
	SignedStateVal VPCStateSigned `json:"signed_state_val"`
	Status         MessageStatus  `json:"status"`
}

// UnmarshalJSON implements json.Unmarshaller interface.
//
// The json message is first unmarshalled retaining the message as raw json.
// Then the message is unmarshalled to appropriate format depending upon the message id.
func (msgPkt *ChMsgPkt) UnmarshalJSON(data []byte) (err error) {

	//Unmarshal only the chMsgPkt, retaining the message as rawJSON
	var rawMsgPkt = chRawMsgPkt{}
	if err = json.Unmarshal(data, &rawMsgPkt); err != nil {
		return err
	}

	//Unmarshal the message to appropriate format depending upon message id
	switch rawMsgPkt.MessageID {
	case MsgIdentityRequest, MsgIdentityResponse:

		var msg JSONMsgIdentity
		if err = json.Unmarshal(rawMsgPkt.Message, &msg); err != nil {
			return err
		}
		msgPkt.Message = msg

	case MsgNewChannelRequest, MsgNewChannelResponse:

		var msg JSONMsgNewChannel
		if err = json.Unmarshal(rawMsgPkt.Message, &msg); err != nil {
			return err
		}
		msgPkt.Message = msg

	case MsgSessionIDRequest, MsgSessionIDResponse:
		var msg JSONMsgSessionID
		if err = json.Unmarshal(rawMsgPkt.Message, &msg); err != nil {
			return err
		}
		msgPkt.Message = msg

	case MsgContractAddrRequest, MsgContractAddrResponse:
		var msg JSONMsgContractAddr
		if err = json.Unmarshal(rawMsgPkt.Message, &msg); err != nil {
			return err
		}
		msgPkt.Message = msg

	case MsgMSCBaseStateRequest, MsgMSCBaseStateResponse:
		var msg JSONMsgMSCBaseState
		if err = json.Unmarshal(rawMsgPkt.Message, &msg); err != nil {
			return err
		}
		msgPkt.Message = msg

	case MsgVPCStateRequest, MsgVPCStateResponse:
		var msg JSONMsgVPCState
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

// ContainsStatus checks of the required value of staus is present in the list.
func ContainsStatus(list []MessageStatus, requiredValue MessageStatus) bool {
	for _, value := range list {
		if value == requiredValue {
			return true
		}
	}
	return false
}
