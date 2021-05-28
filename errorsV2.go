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

package perun

import (
	"fmt"
)

// APIError represents the error that will be returned by the API of perun node.
type apiErrorV2 struct {
	category ErrorCategory
	code     ErrorCode
	message  string
	addInfo  interface{}
}

// Category returns the error category for this API Error.
func (e apiErrorV2) Category() ErrorCategory {
	return e.category
}

// Code returns the error code for this API Error.
func (e apiErrorV2) Code() ErrorCode {
	return e.code
}

// Message returns the error message for this API Error.
func (e apiErrorV2) Message() string {
	return e.message
}

// AddInfo returns the additional info for this API Error.
func (e apiErrorV2) AddInfo() interface{} {
	return e.addInfo
}

// Error implement the error interface for API error.
func (e apiErrorV2) Error() string {
	return fmt.Sprintf("Category: %s, Code: %d, Message: %s, AddInfo: %+v",
		e.Category(), e.Code(), e.Message(), e.AddInfo())
}

// NewAPIErrV2 returns an APIErrV2 with given parameters.
//
// For most use cases, call the error code specific constructor functions.
// This function is intended for use in places only where an APIErrV2 is to be modified.
// Copy each field, modify and create a new one using this function.
func NewAPIErrV2(category ErrorCategory, code ErrorCode, message string, addInfo interface{}) APIErrorV2 {
	return apiErrorV2{
		category: category,
		code:     code,
		message:  message,
		addInfo:  addInfo,
	}
}

// NewAPIErrV2PeerRequestTimedOut returns an ErrV2PeerRequestTimedOut API Error
// with the given peer alias and response timeout.
func NewAPIErrV2PeerRequestTimedOut(peerAlias, timeout, message string) APIErrorV2 {
	return apiErrorV2{
		category: ParticipantError,
		code:     ErrV2PeerRequestTimedOut,
		message:  message,
		addInfo: ErrV2InfoPeerRequestTimedOut{
			PeerAlias: peerAlias,
			Timeout:   timeout,
		},
	}
}

// NewAPIErrV2PeerRejected returns an ErrV2PeerRejected API Error with the
// given peer alias and reason.
func NewAPIErrV2PeerRejected(peerAlias, reason, message string) APIErrorV2 {
	return apiErrorV2{
		category: ParticipantError,
		code:     ErrV2PeerRejected,
		message:  message,
		addInfo: ErrV2InfoPeerRejected{
			PeerAlias: peerAlias,
			Reason:    reason,
		},
	}
}

// NewAPIErrV2PeerNotFunded returns an ErrV2PeerNotFunded API Error with the
// given peer alias.
func NewAPIErrV2PeerNotFunded(peerAlias, message string) APIErrorV2 {
	return apiErrorV2{
		category: ParticipantError,
		code:     ErrV2PeerNotFunded,
		message:  message,
		addInfo: ErrV2InfoPeerNotFunded{
			PeerAlias: peerAlias,
		},
	}
}

// NewAPIErrV2UserResponseTimedOut returns an ErrUserResponseTimedOut API Error
// with the given expiry.
func NewAPIErrV2UserResponseTimedOut(expiry, receivedAt int64) APIErrorV2 {
	return apiErrorV2{
		category: ParticipantError,
		code:     ErrV2UserResponseTimedOut,
		message:  "user response timed out",
		addInfo: ErrV2InfoUserResponseTimedOut{
			Expiry:     expiry,
			ReceivedAt: receivedAt,
		},
	}
}

// NewAPIErrV2ResourceNotFound returns an ErrResourceNotFound API Error with
// the given resource type, ID and error message.
func NewAPIErrV2ResourceNotFound(resourceType, resourceID, message string) APIErrorV2 {
	return apiErrorV2{
		category: ClientError,
		code:     ErrV2ResourceNotFound,
		message:  message,
		addInfo: ErrV2InfoResourceNotFound{
			Type: resourceType,
			ID:   resourceID,
		},
	}
}

// NewAPIErrV2ResourceExists returns an ErrResourceExists API Error with
// the given resource type, ID and error message.
func NewAPIErrV2ResourceExists(resourceType, resourceID, message string) APIErrorV2 {
	return apiErrorV2{
		category: ClientError,
		code:     ErrV2ResourceExists,
		message:  message,
		addInfo: ErrV2InfoResourceExists{
			Type: resourceType,
			ID:   resourceID,
		},
	}
}

// NewAPIErrV2InvalidArgument returns an ErrInvalidArgument API Error with the given
// argument name, value, requirement for the argument and the error message.
func NewAPIErrV2InvalidArgument(name, value, requirement, message string) APIErrorV2 {
	return apiErrorV2{
		category: ClientError,
		code:     ErrV2InvalidArgument,
		message:  message,
		addInfo: ErrV2InfoInvalidArgument{
			Name:        name,
			Value:       value,
			Requirement: requirement,
		},
	}
}

// NewAPIErrV2FailedPreCondition returns an ErrV2FailedPreCondition API Error with the given
// error message.
//
// AddInfo is taken as interface because for this error, the callers can
// provide different type of additional info, as defined in the corresponding
// APIs.
func NewAPIErrV2FailedPreCondition(message string, addInfo interface{}) APIErrorV2 {
	return apiErrorV2{
		category: ClientError,
		code:     ErrV2FailedPreCondition,
		message:  message,
		addInfo:  addInfo,
	}
}

// NewAPIErrInfoFailedPreConditionUnclosedChs returns additional info for a
// specific case of ErrV2FailedPreCondition when Session.Close API is called
// without force option and there are one or more unclosed channels in it.
func NewAPIErrInfoFailedPreConditionUnclosedChs(chs []ChInfo) interface{} {
	return ErrV2InfoFailedPreCondUnclosedChs{
		ChInfos: chs,
	}
}

// NewAPIErrV2TxTimedOut returns an ErrV2TxTimedOut API Error with the given
// error message.
func NewAPIErrV2TxTimedOut(txType, txID, txTimeout, message string) APIErrorV2 {
	return apiErrorV2{
		category: ProtocolFatalError,
		code:     ErrV2TxTimedOut,
		message:  message,
		addInfo: ErrV2InfoTxTimedOut{
			TxType:    txType,
			TxID:      txID,
			TxTimeout: txTimeout,
		},
	}
}

// NewAPIErrV2ChainNotReachable returns an ErrV2ChainNotReachable API Error
// with the given error message.
func NewAPIErrV2ChainNotReachable(chainURL, message string) APIErrorV2 {
	return apiErrorV2{
		category: ProtocolFatalError,
		code:     ErrV2ChainNotReachable,
		message:  message,
		addInfo: ErrV2InfoChainNotReachable{
			ChainURL: chainURL,
		},
	}
}

// NewAPIErrV2UnknownInternal returns an ErrV2UnknownInternal API Error with the given
// error message.
func NewAPIErrV2UnknownInternal(err error) APIErrorV2 {
	return apiErrorV2{
		category: InternalError,
		code:     ErrV2UnknownInternal,
		message:  err.Error(),
	}
}

// APIErrV2AsMap returns a map containing entries for the method and each of
// the fields in the api error (except message). The map can be directly passed
// to the logger for logging the data in a structured format.
func APIErrV2AsMap(method string, err APIErrorV2) map[string]interface{} {
	return map[string]interface{}{
		"method":   method,
		"category": err.Category().String(),
		"code":     err.Code(),
		"add info": fmt.Sprintf("%+v", err.AddInfo()),
	}
}
