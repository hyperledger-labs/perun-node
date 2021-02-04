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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger-labs/perun-node"
)

// nolint: dupl	// not duplicate of test Test_NewErrResourceExists.
func Test_NewErrResourceNotFound(t *testing.T) {
	resourceType := "any-type"
	resourceID := "any-id"
	message := "any-message"

	err := perun.NewAPIErrV2ResourceNotFound(resourceType, resourceID, message)
	require.NotNil(t, err)

	assert.Equal(t, perun.ClientError, err.Category())
	assert.Equal(t, perun.ErrV2ResourceNotFound, err.Code())
	assert.Equal(t, message, err.Message())
	addInfo, ok := err.AddInfo().(perun.ErrV2InfoResourceNotFound)
	require.True(t, ok)
	assert.Equal(t, addInfo.Type, resourceType)
	assert.Equal(t, addInfo.ID, resourceID)
	t.Log(err.Error())
}

// nolint: dupl	// not duplicate of test Test_NewErrResourceNotFound.
func Test_NewErrResourceExists(t *testing.T) {
	resourceType := "any-type"
	resourceID := "any-id"
	message := "any-message"

	err := perun.NewAPIErrV2ResourceExists(resourceType, resourceID, message)
	require.NotNil(t, err)

	assert.Equal(t, perun.ClientError, err.Category())
	assert.Equal(t, perun.ErrV2ResourceExists, err.Code())
	assert.Equal(t, message, err.Message())
	addInfo, ok := err.AddInfo().(perun.ErrV2InfoResourceExists)
	require.True(t, ok)
	assert.Equal(t, addInfo.Type, resourceType)
	assert.Equal(t, addInfo.ID, resourceID)
	t.Log(err.Error())
}

func Test_NewErrInvalidArgument(t *testing.T) {
	resourceType := "any-type"
	resourceID := "any-id"
	requirement := "any-requirement"
	message := "any-message"

	err := perun.NewAPIErrV2InvalidArgument(resourceType, resourceID, requirement, message)
	require.NotNil(t, err)

	assert.Equal(t, perun.ClientError, err.Category())
	assert.Equal(t, perun.ErrV2InvalidArgument, err.Code())
	assert.Equal(t, message, err.Message())
	addInfo, ok := err.AddInfo().(perun.ErrV2InfoInvalidArgument)
	require.True(t, ok)
	assert.Equal(t, addInfo.Name, resourceType)
	assert.Equal(t, addInfo.Value, resourceID)
	assert.Equal(t, addInfo.Requirement, requirement)
	t.Log(err.Error())
}
