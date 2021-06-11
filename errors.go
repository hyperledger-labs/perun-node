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
	"io"

	"github.com/pkg/errors"
)

// APIError represents the error that will be returned by the API of perun node.
//
// It implements Cause() and Unwrap() methods that implements the underlying
// error, which can further be unwrapped, inspected.
//
// It also implements a customer Formatter, so that the stack trace of
// underlying error is printed when using "%+v" verb.
type apiError struct {
	category ErrorCategory
	code     ErrorCode
	err      error
	addInfo  interface{}
}

// Category returns the error category for this API Error.
func (e apiError) Category() ErrorCategory { return e.category }

// Code returns the error code for this API Error.
func (e apiError) Code() ErrorCode { return e.code }

// Message returns the error message for this API Error.
func (e apiError) Message() string { return e.err.Error() }

// AddInfo returns the additional info for this API Error.
func (e apiError) AddInfo() interface{} {
	return e.addInfo
}

// Error implement the error interface for API error.
func (e apiError) Error() string {
	return fmt.Sprintf("%s %d:%v", e.Category(), e.Code(), e.Message())
}

func (e apiError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s %d:%+v", e.Category(), e.Code(), e.err)
			return
		}
		fallthrough
	case 's':
		//nolint: errcheck,gosec	// Error of ioString need not be checked.
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

func (e apiError) Cause() error { return e.err }

func (e apiError) Unwrap() error { return e.err }

// NewAPIErr returns an APIErr with given parameters.
//
// For most use cases, call the error code specific constructor functions.
// This function is intended for use in places only where an APIErr is to be modified.
// For example, in app packages where the data in additional field is to be modified.
// Copy each field, modify and create a new one using this function.
func NewAPIErr(category ErrorCategory, code ErrorCode, err error, addInfo interface{}) APIError {
	return apiError{
		category: category,
		code:     code,
		err:      err,
		addInfo:  addInfo,
	}
}

// NewAPIErrPeerRequestTimedOut returns an ErrPeerRequestTimedOut API Error
// with the given peer alias and response timeout.
//
// It does not validate the time string and it is in a proper format.  Eg: 10s.
func NewAPIErrPeerRequestTimedOut(err error, peerAlias, timeout string) APIError {
	message := fmt.Sprintf("timed out waiting for response from %s for %s", peerAlias, timeout)
	return NewAPIErr(
		ParticipantError,
		ErrPeerRequestTimedOut,
		errors.WithMessage(err, message),
		ErrInfoPeerRequestTimedOut{
			PeerAlias: peerAlias,
			Timeout:   timeout,
		},
	)
}

// NewAPIErrPeerRejected returns an ErrPeerRejected API Error with the
// given peer alias and reason.
func NewAPIErrPeerRejected(err error, peerAlias, reason string) APIError {
	message := fmt.Sprintf("peer %s rejected with reason %s", peerAlias, reason)
	return NewAPIErr(
		ParticipantError,
		ErrPeerRejected,
		errors.WithMessage(err, message),
		ErrInfoPeerRejected{
			PeerAlias: peerAlias,
			Reason:    reason,
		},
	)
}

// NewAPIErrPeerNotFunded returns an ErrPeerNotFunded API Error with the
// given peer alias.
func NewAPIErrPeerNotFunded(err error, peerAlias string) APIError {
	message := fmt.Sprintf("peer %s did not fund within expected time", peerAlias)
	return NewAPIErr(
		ParticipantError,
		ErrPeerNotFunded,
		errors.WithMessage(err, message),
		ErrInfoPeerNotFunded{
			PeerAlias: peerAlias,
		},
	)
}

// NewAPIErrUserResponseTimedOut returns an ErrUserResponseTimedOut API Error
// with the given expiry.
func NewAPIErrUserResponseTimedOut(expiry, receivedAt int64) APIError {
	message := fmt.Sprintf("response received (at %v) after timeout expired (at %v)", receivedAt, expiry)
	return NewAPIErr(
		ParticipantError,
		ErrUserResponseTimedOut,
		errors.New(message),
		ErrInfoUserResponseTimedOut{
			Expiry:     expiry,
			ReceivedAt: receivedAt,
		},
	)
}

// ResourceType is used to enumerate valid resource types in ResourceNotFound
// and ResourceExists errors.
//
// The enumeration of valid constants should be defined in the package using
// the error constructors.
type ResourceType string

// NewAPIErrResourceNotFound returns an ErrResourceNotFound API Error with
// the given resource type and ID.
func NewAPIErrResourceNotFound(resourceType ResourceType, resourceID string) APIError {
	message := fmt.Sprintf("cannot find %s with ID: %s", resourceType, resourceID)
	return NewAPIErr(
		ClientError,
		ErrResourceNotFound,
		errors.New(message),
		ErrInfoResourceNotFound{
			Type: string(resourceType),
			ID:   resourceID,
		},
	)
}

// NewAPIErrResourceExists returns an ErrResourceExists API Error with
// the given resource type and ID.
func NewAPIErrResourceExists(resourceType ResourceType, resourceID string) APIError {
	message := fmt.Sprintf("%s with ID: %s already exists", resourceType, resourceID)
	return NewAPIErr(
		ClientError,
		ErrResourceExists,
		errors.New(message),
		ErrInfoResourceExists{
			Type: string(resourceType),
			ID:   resourceID,
		},
	)
}

// ArgumentName type is used enumerate valid argument names for use
// InvalidArgument error.
//
// The enumeration of valid constants should be defined in the package using
// the error constructors.
type ArgumentName string

// NewAPIErrInvalidArgument returns an ErrInvalidArgument API Error with the given
// argument name and value.
func NewAPIErrInvalidArgument(err error, name ArgumentName, value string) APIError {
	message := fmt.Sprintf("invalid value for %s: %s", name, value)
	return NewAPIErr(
		ClientError,
		ErrInvalidArgument,
		errors.WithMessage(err, message),
		ErrInfoInvalidArgument{
			Name:  string(name),
			Value: value,
		},
	)
}

// NewAPIErrFailedPreCondition returns an ErrFailedPreCondition API Error with
// the given error message.
//
// By default, the additional info for this type of err is nil. In case where
// additional info is to be included (to be documented in the API),
// constructors for APIErrFailedPreCondition specific to that case should be used.
func NewAPIErrFailedPreCondition(err error) APIError {
	return newAPIErrFailedPreCondition(err, nil)
}

// NewAPIErrFailedPreConditionUnclosedChs returns an ErrFailedPreCondition
// API Error for the special case where channel close was closed with no-force
// option and, one or more open channels exists.
//
func NewAPIErrFailedPreConditionUnclosedChs(err error, chs []ChInfo) APIError {
	return newAPIErrFailedPreCondition(err,
		ErrInfoFailedPreCondUnclosedChs{
			ChInfos: chs,
		})
}

// NewAPIErrFailedPreCondition returns an ErrFailedPreCondition API Error with the given
// error message.
//
// AddInfo is taken as interface because for this error, the callers can
// provide different type of additional info, as defined in the corresponding
// APIs.
func newAPIErrFailedPreCondition(err error, addInfo interface{}) APIError {
	message := "failed pre-condition"
	return NewAPIErr(
		ClientError,
		ErrFailedPreCondition,
		errors.WithMessage(err, message),
		addInfo,
	)
}

// NewAPIErrInvalidConfig returns an ErrInvalidConfig, API Error with the given
// config name and value.
func NewAPIErrInvalidConfig(err error, name, value string) APIError {
	message := fmt.Sprintf("invalid value for %s: %s", name, value)
	return NewAPIErr(
		ClientError,
		ErrInvalidConfig,
		errors.WithMessage(err, message),
		ErrInfoInvalidConfig{
			Name:  name,
			Value: value,
		},
	)
}

// NewAPIErrInvalidContracts returns an ErrInvalidContracts API Error with the
// given contract error infos.
//
// For this error, stack traces of errors for each contract cannot be
// retreived, as there are more than one.
func NewAPIErrInvalidContracts(contractErrInfos []ContractErrInfo) APIError {
	contractsErrInfosMsg := ""
	for _, c := range contractErrInfos {
		contractsErrInfosMsg += fmt.Sprintf("(%s at address %s: %v) ", c.Name, c.Address, c.Error)
	}
	message := "invalid contracts: " + contractsErrInfosMsg
	return NewAPIErr(
		ClientError,
		ErrInvalidContracts,
		errors.New(message),
		ErrInfoInvalidContracts{
			ContractErrInfos: contractErrInfos,
		},
	)
}

// NewAPIErrTxTimedOut returns an ErrTxTimedOut API Error with the given
// error message.
func NewAPIErrTxTimedOut(err error, txType, txID, txTimeout string) APIError {
	message := fmt.Sprintf("timed out waiting for %s tx (ID:%s) to be mined in %s: %v", txType, txID, txTimeout, err)
	return NewAPIErr(
		ProtocolFatalError,
		ErrTxTimedOut,
		errors.WithMessage(err, message),
		ErrInfoTxTimedOut{
			TxType:    txType,
			TxID:      txID,
			TxTimeout: txTimeout,
		},
	)
}

// NewAPIErrChainNotReachable returns an ErrChainNotReachable API Error
// with the given error message.
func NewAPIErrChainNotReachable(err error, chainURL string) APIError {
	message := fmt.Sprintf("chain not reachable at URL %s", chainURL)
	return NewAPIErr(
		ProtocolFatalError,
		ErrChainNotReachable,
		errors.WithMessage(err, message),
		ErrInfoChainNotReachable{
			ChainURL: chainURL,
		},
	)
}

// NewAPIErrUnknownInternal returns an ErrUnknownInternal API Error with the given
// error message.
func NewAPIErrUnknownInternal(err error) APIError {
	message := "unknown internal error"
	return NewAPIErr(
		InternalError,
		ErrUnknownInternal,
		errors.WithMessage(err, message),
		nil,
	)
}

// APIErrAsMap returns a map containing entries for the method and each of
// the fields in the api error (except message). The map can be directly passed
// to the logger for logging the data in a structured format.
func APIErrAsMap(method string, err APIError) map[string]interface{} {
	return map[string]interface{}{
		"method":   method,
		"category": err.Category().String(),
		"code":     err.Code(),
	}
}
