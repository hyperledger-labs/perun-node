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
type apiError struct {
	category ErrorCategory
	code     ErrorCode
	message  string
	addInfo  interface{}
}

// Category returns the error category for this API Error.
func (e apiError) Category() ErrorCategory {
	return e.category
}

// Code returns the error code for this API Error.
func (e apiError) Code() ErrorCode {
	return e.code
}

// Message returns the error message for this API Error.
func (e apiError) Message() string {
	return e.message
}

// AddInfo returns the additional info for this API Error.
func (e apiError) AddInfo() interface{} {
	return e.addInfo
}

// Error implement the error interface for API error.
func (e apiError) Error() string {
	return fmt.Sprintf("Category: %s, Code: %d, Message: %s, AddInfo: %+v",
		e.Category(), e.Code(), e.Message(), e.AddInfo())
}

// NewAPIErr returns an APIErr with given parameters.
//
// For most use cases, call the error code specific constructor functions.
// This function is intended for use in places only where an APIErr is to be modified.
// Copy each field, modify and create a new one using this function.
func NewAPIErr(category ErrorCategory, code ErrorCode, message string, addInfo interface{}) APIError {
	return apiError{
		category: category,
		code:     code,
		message:  message,
		addInfo:  addInfo,
	}
}

// NewAPIErrPeerRequestTimedOut returns an ErrPeerRequestTimedOut API Error
// with the given peer alias and response timeout.
func NewAPIErrPeerRequestTimedOut(peerAlias, timeout, message string) APIError {
	return apiError{
		category: ParticipantError,
		code:     ErrPeerRequestTimedOut,
		message:  message,
		addInfo: ErrInfoPeerRequestTimedOut{
			PeerAlias: peerAlias,
			Timeout:   timeout,
		},
	}
}

// NewAPIErrPeerRejected returns an ErrPeerRejected API Error with the
// given peer alias and reason.
func NewAPIErrPeerRejected(peerAlias, reason, message string) APIError {
	return apiError{
		category: ParticipantError,
		code:     ErrPeerRejected,
		message:  message,
		addInfo: ErrInfoPeerRejected{
			PeerAlias: peerAlias,
			Reason:    reason,
		},
	}
}

// NewAPIErrPeerNotFunded returns an ErrPeerNotFunded API Error with the
// given peer alias.
func NewAPIErrPeerNotFunded(peerAlias, message string) APIError {
	return apiError{
		category: ParticipantError,
		code:     ErrPeerNotFunded,
		message:  message,
		addInfo: ErrInfoPeerNotFunded{
			PeerAlias: peerAlias,
		},
	}
}

// NewAPIErrUserResponseTimedOut returns an ErrUserResponseTimedOut API Error
// with the given expiry.
func NewAPIErrUserResponseTimedOut(expiry, receivedAt int64) APIError {
	return apiError{
		category: ParticipantError,
		code:     ErrUserResponseTimedOut,
		message:  "user response timed out",
		addInfo: ErrInfoUserResponseTimedOut{
			Expiry:     expiry,
			ReceivedAt: receivedAt,
		},
	}
}

// NewAPIErrResourceNotFound returns an ErrResourceNotFound API Error with
// the given resource type, ID and error message.
func NewAPIErrResourceNotFound(resourceType, resourceID, message string) APIError {
	return apiError{
		category: ClientError,
		code:     ErrResourceNotFound,
		message:  message,
		addInfo: ErrInfoResourceNotFound{
			Type: resourceType,
			ID:   resourceID,
		},
	}
}

// NewAPIErrResourceExists returns an ErrResourceExists API Error with
// the given resource type, ID and error message.
func NewAPIErrResourceExists(resourceType, resourceID, message string) APIError {
	return apiError{
		category: ClientError,
		code:     ErrResourceExists,
		message:  message,
		addInfo: ErrInfoResourceExists{
			Type: resourceType,
			ID:   resourceID,
		},
	}
}

// NewAPIErrInvalidArgument returns an ErrInvalidArgument API Error with the given
// argument name, value, requirement for the argument and the error message.
func NewAPIErrInvalidArgument(name, value, requirement, message string) APIError {
	return apiError{
		category: ClientError,
		code:     ErrInvalidArgument,
		message:  message,
		addInfo: ErrInfoInvalidArgument{
			Name:        name,
			Value:       value,
			Requirement: requirement,
		},
	}
}

// NewAPIErrFailedPreCondition returns an ErrFailedPreCondition API Error with the given
// error message.
//
// AddInfo is taken as interface because for this error, the callers can
// provide different type of additional info, as defined in the corresponding
// APIs.
func NewAPIErrFailedPreCondition(message string, addInfo interface{}) APIError {
	return apiError{
		category: ClientError,
		code:     ErrFailedPreCondition,
		message:  message,
		addInfo:  addInfo,
	}
}

// NewAPIErrInvalidConfig returns an ErrInvalidConfig, API Error with the
// given error message.
func NewAPIErrInvalidConfig(name, value, message string) APIError {
	return apiError{
		category: ClientError,
		code:     ErrInvalidConfig,
		message:  message,
		addInfo: ErrInfoInvalidConfig{
			Name:  name,
			Value: value,
		},
	}
}

// NewAPIErrInfoFailedPreConditionUnclosedChs returns additional info for a
// specific case of ErrFailedPreCondition when Session.Close API is called
// without force option and there are one or more unclosed channels in it.
func NewAPIErrInfoFailedPreConditionUnclosedChs(chs []ChInfo) interface{} {
	return ErrInfoFailedPreCondUnclosedChs{
		ChInfos: chs,
	}
}

// NewAPIErrInvalidContracts returns an ErrInvalidContracts API Error with the
// given error message.
func NewAPIErrInvalidContracts(contractsInfo []ContractErrInfo) APIError {
	return apiError{
		category: ClientError,
		code:     ErrInvalidContracts,
		message:  "error in contract validation",
		addInfo: ErrInfoInvalidContracts{
			ContractErrInfos: contractsInfo,
		},
	}
}

// NewAPIErrTxTimedOut returns an ErrTxTimedOut API Error with the given
// error message.
func NewAPIErrTxTimedOut(txType, txID, txTimeout, message string) APIError {
	return apiError{
		category: ProtocolFatalError,
		code:     ErrTxTimedOut,
		message:  message,
		addInfo: ErrInfoTxTimedOut{
			TxType:    txType,
			TxID:      txID,
			TxTimeout: txTimeout,
		},
	}
}

// NewAPIErrChainNotReachable returns an ErrChainNotReachable API Error
// with the given error message.
func NewAPIErrChainNotReachable(chainURL, message string) APIError {
	return apiError{
		category: ProtocolFatalError,
		code:     ErrChainNotReachable,
		message:  message,
		addInfo: ErrInfoChainNotReachable{
			ChainURL: chainURL,
		},
	}
}

// NewAPIErrUnknownInternal returns an ErrUnknownInternal API Error with the given
// error message.
func NewAPIErrUnknownInternal(err error) APIError {
	return apiError{
		category: InternalError,
		code:     ErrUnknownInternal,
		message:  err.Error(),
	}
}

// APIErrAsMap returns a map containing entries for the method and each of
// the fields in the api error (except message). The map can be directly passed
// to the logger for logging the data in a structured format.
func APIErrAsMap(method string, err APIError) map[string]interface{} {
	return map[string]interface{}{
		"method":   method,
		"category": err.Category().String(),
		"code":     err.Code(),
		"add info": fmt.Sprintf("%+v", err.AddInfo()),
	}
}
