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

package perun_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/peruntest"
)

var errTest = fmt.Errorf("const error for test")

type customError1 struct{ actualErr error }

func (c customError1) Error() string { return "custom error 1: " + c.actualErr.Error() }
func (c customError1) Unwrap() error { return c.actualErr }

type customError2 struct{ actualErr error }

func (c customError2) Error() string { return "custom error 2: " + c.actualErr.Error() }
func (c customError2) Unwrap() error { return c.actualErr }

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func Test_NewAPIErr(t *testing.T) {
	// Construct an error that contain an underlying custom error and error constant.
	// So that, they can be inspected using errors.(Is|As)
	err := errors.WithStack(customError1{
		actualErr: customError2{errTest},
	})

	// Construct an APIError.
	category := perun.InternalError
	code := perun.ErrUnknownInternal
	apiErr := perun.NewAPIErr(category, code, err, nil)

	// Test if code, category and message are assigned properly.
	peruntest.AssertAPIError(t, apiErr, category, code, err.Error())

	// Test if message is formated properly when using error and fmt.Formatter interfaces.
	wantErrMsg := "Internal 401:custom error 1: custom error 2: const error for test"
	stackTrace, ok := err.(stackTracer)
	require.True(t, ok)

	assert.Equal(t, wantErrMsg, apiErr.Error())
	assert.Equal(t, wantErrMsg, fmt.Sprintf("%v", apiErr))
	assert.Equal(t, fmt.Sprintf("%q", wantErrMsg), fmt.Sprintf("%q", apiErr))
	assert.Contains(t, fmt.Sprintf("%+v", apiErr), fmt.Sprintf("%+v", stackTrace))

	// Explicitly test Unwrap and Cause functions.
	eCause := errors.Cause(apiErr)
	require.EqualError(t, eCause, err.Error())
	eUnwrap := errors.Unwrap(apiErr)
	require.EqualError(t, eUnwrap, err.Error())

	// Test if underlying errors can be retrieved.
	e1 := customError1{}
	ok = errors.As(err, &e1)
	require.True(t, ok)

	e2 := customError2{}
	ok = errors.As(err, &e2)
	require.True(t, ok)

	ok = errors.Is(e2, errTest)
	require.True(t, ok)
}

func Test_NewAPIErrPeerRequestTimedOut(t *testing.T) {
	tests := []struct {
		name    string
		timeout string
	}{
		{"valid_timeout", "1m0s"},
		{"invalid_timeout", "invalid-duration-string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New("sample error")
			peerAlias := "some-alias"
			timeout := tt.timeout
			wantMsg := fmt.Sprintf("timed out waiting for response from %s for %s: %v", peerAlias, timeout, err)

			apiErr := perun.NewAPIErrPeerRequestTimedOut(err, peerAlias, timeout)
			peruntest.AssertAPIError(t, apiErr, perun.ParticipantError, perun.ErrPeerRequestTimedOut, wantMsg)
			peruntest.AssertErrInfoPeerRequestTimedOut(t, apiErr.AddInfo(), peerAlias, timeout)
		})
	}
}

func Test_NewAPIErrPeerRejected(t *testing.T) {
	err := errors.New("sample error")
	peerAlias := "some-alias"
	reason := "random reason"
	wantMsg := fmt.Sprintf("peer %s rejected with reason %s: %v", peerAlias, reason, err)

	apiErr := perun.NewAPIErrPeerRejected(err, peerAlias, reason)
	peruntest.AssertAPIError(t, apiErr, perun.ParticipantError, perun.ErrPeerRejected, wantMsg)
	peruntest.AssertErrInfoPeerRejected(t, apiErr.AddInfo(), peerAlias, reason)
}

func Test_NewAPIErrPeerNotFunded(t *testing.T) {
	err := errors.New("sample error")
	peerAlias := "some-alias"
	wantMsg := fmt.Sprintf("peer %s did not fund within expected time: %v", peerAlias, err)

	apiErr := perun.NewAPIErrPeerNotFunded(err, peerAlias)
	peruntest.AssertAPIError(t, apiErr, perun.ParticipantError, perun.ErrPeerNotFunded, wantMsg)
	peruntest.AssertErrInfoPeerNotFunded(t, apiErr.AddInfo(), peerAlias)
}

func Test_NewAPIErrUserResponseTimedOut(t *testing.T) {
	expiry := time.Now().Unix()
	receivedAt := time.Now().Add(1 * time.Second).Unix()
	wantMsg := fmt.Sprintf("response received (at %v) after timeout expired (at %v)", receivedAt, expiry)

	apiErr := perun.NewAPIErrUserResponseTimedOut(expiry, receivedAt)
	peruntest.AssertAPIError(t, apiErr, perun.ParticipantError, perun.ErrUserResponseTimedOut, wantMsg)
	// not using helper function, because it will only test
	// if receivedAt is later than expiry and if expiry is in the past.
	// here test only if proper values are assigned.
	addInfo, ok := apiErr.AddInfo().(perun.ErrInfoUserResponseTimedOut)
	require.True(t, ok)
	assert.Equal(t, expiry, addInfo.Expiry)
	assert.Equal(t, receivedAt, addInfo.ReceivedAt)
}

func Test_NewErrResourceNotFound(t *testing.T) {
	resourceType := perun.ResourceType("any-type")
	resourceID := "any-id"
	wantMsg := fmt.Sprintf("cannot find %s with ID: %s", resourceType, resourceID)

	apiErr := perun.NewAPIErrResourceNotFound(resourceType, resourceID)
	peruntest.AssertAPIError(t, apiErr, perun.ClientError, perun.ErrResourceNotFound, wantMsg)
	peruntest.AssertErrInfoResourceNotFound(t, apiErr.AddInfo(), resourceType, resourceID)
}

func Test_NewErrResourceExists(t *testing.T) {
	resourceType := perun.ResourceType("any-type")
	resourceID := "any-id"
	wantMsg := fmt.Sprintf("%s with ID: %s already exists", resourceType, resourceID)

	apiErr := perun.NewAPIErrResourceExists(resourceType, resourceID)
	peruntest.AssertAPIError(t, apiErr, perun.ClientError, perun.ErrResourceExists, wantMsg)
	peruntest.AssertErrInfoResourceExists(t, apiErr.AddInfo(), resourceType, resourceID)
}

func Test_NewErrInvalidArgument(t *testing.T) {
	name := perun.ArgumentName("any-name")
	value := "any-value"
	err := errors.New("any-error")
	wantMsg := fmt.Sprintf("invalid value for %s: %s: %v", name, value, err)

	apiErr := perun.NewAPIErrInvalidArgument(err, name, value)
	peruntest.AssertAPIError(t, apiErr, perun.ClientError, perun.ErrInvalidArgument, wantMsg)
	peruntest.AssertErrInfoInvalidArgument(t, apiErr.AddInfo(), name, value)
}

func Test_NewErrFailedPreCondition(t *testing.T) {
	err := errors.New("any-error")
	wantMsg := fmt.Sprintf("failed pre-condition: %v", err)

	apiErr := perun.NewAPIErrFailedPreCondition(err)
	peruntest.AssertAPIError(t, apiErr, perun.ClientError, perun.ErrFailedPreCondition, wantMsg)
}

func Test_NewErrFailedPreConditionUnclosedChs(t *testing.T) {
	err := errors.New("session cannot be closed with channels in unexpected phase")
	wantMsg := fmt.Sprintf("failed pre-condition: %v", err)
	unclosedChs := []perun.ChInfo{
		{
			ChID:    "one",
			Version: "1",
		},
	}

	apiErr := perun.NewAPIErrFailedPreConditionUnclosedChs(err, unclosedChs)
	peruntest.AssertAPIError(t, apiErr, perun.ClientError, perun.ErrFailedPreCondition, wantMsg)
	peruntest.AssertErrInfoFailedPreCondUnclosedChs(t, apiErr.AddInfo(), unclosedChs)
}

func Test_NewErrInvalidConfig(t *testing.T) {
	name := "any-name"
	value := "any-value"
	err := errors.New("any-error")
	wantMsg := fmt.Sprintf("invalid value for %s: %s: %v", name, value, err)

	apiErr := perun.NewAPIErrInvalidConfig(err, name, value)
	peruntest.AssertAPIError(t, apiErr, perun.ClientError, perun.ErrInvalidConfig, wantMsg)
	peruntest.AssertErrInfoInvalidConfig(t, apiErr.AddInfo(), name, value)
}

func Test_NewErrInvalidContracts(t *testing.T) {
	contractErrInfos := []perun.ContractErrInfo{
		{"contracts-1", "addr 1", errors.New("1").Error()},
		{"contracts-2", "addr 2", errors.New("2").Error()},
	}
	contractsErrInfosMsg := ""
	for _, c := range contractErrInfos {
		contractsErrInfosMsg += fmt.Sprintf("(%s at address %s: %v) ", c.Name, c.Address, c.Error)
	}
	wantMsg := "invalid contracts: " + contractsErrInfosMsg

	apiErr := perun.NewAPIErrInvalidContracts(contractErrInfos)
	peruntest.AssertAPIError(t, apiErr, perun.ClientError, perun.ErrInvalidContracts, wantMsg)
	peruntest.AssertErrInfoInvalidContracts(t, apiErr.AddInfo(), contractErrInfos)
}

func Test_NewErrTxTimedOut(t *testing.T) {
	txType := "any-type"
	txID := "any-id"
	txTimeout := "10m1s"
	err := errors.New("any-error")
	wantMsg := fmt.Sprintf("timed out waiting for %s tx (ID:%s) to be mined in %s: %v", txType, txID, txTimeout, err)

	apiErr := perun.NewAPIErrTxTimedOut(err, txType, txID, txTimeout)
	peruntest.AssertAPIError(t, apiErr, perun.ProtocolFatalError, perun.ErrTxTimedOut, wantMsg)
	peruntest.AssertErrInfoTxTimedOut(t, apiErr.AddInfo(), txType, txID, txTimeout)
}

func Test_NewErrChainNotReachable(t *testing.T) {
	chainURL := "any-chain-url"
	err := errors.New("any-error")
	wantMsg := fmt.Sprintf("chain not reachable at URL %s: %v", chainURL, err)

	apiErr := perun.NewAPIErrChainNotReachable(err, chainURL)
	peruntest.AssertAPIError(t, apiErr, perun.ProtocolFatalError, perun.ErrChainNotReachable, wantMsg)
	peruntest.AssertErrInfoChainNotReachable(t, apiErr.AddInfo(), chainURL)
}

func Test_NewErrUnknownInternal(t *testing.T) {
	err := errors.New("any-error")
	wantMsg := fmt.Sprintf("unknown internal error: %v", err)

	apiErr := perun.NewAPIErrUnknownInternal(err)
	peruntest.AssertAPIError(t, apiErr, perun.InternalError, perun.ErrUnknownInternal, wantMsg)
}

func Test_APIErrAsMap(t *testing.T) {
	method := "test method"
	apiErr := perun.NewAPIErrUnknownInternal(errors.New("any-error"))
	errAsMap := perun.APIErrAsMap(method, apiErr)
	assert.Equal(t, errAsMap, map[string]interface{}{
		"method":   method,
		"category": perun.InternalError.String(),
		"code":     perun.ErrUnknownInternal,
	})
}
