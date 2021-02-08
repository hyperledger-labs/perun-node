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

// APIError represents the error that will returned by the API of perun node.
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
func NewAPIErrV2FailedPreCondition(message string) APIErrorV2 {
	return apiErrorV2{
		category: ClientError,
		code:     ErrV2FailedPreCondition,
		message:  message,
	}
}

// NewAPIErrV2UnknownInternal returns an ErrV2UnknownInternal API Error with the given
// error message.
func NewAPIErrV2UnknownInternal(err error) APIErrorV2 {
	return apiErrorV2{
		category: ClientError,
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
