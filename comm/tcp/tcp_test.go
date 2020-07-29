// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/perun-node
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

package tcp_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/direct-state-transfer/perun-node"
	"github.com/direct-state-transfer/perun-node/comm/tcp"
)

func Test_CommBackend_Interface(t *testing.T) {
	assert.Implements(t, (*perun.CommBackend)(nil), new(tcp.Backend))
}

func Test_Backend(t *testing.T) {
	backend := tcp.NewTCPBackend(1 * time.Second)
	require.NotNil(t, backend)

	// Find a free port to start the listener.
	port, err := freeport.GetFreePort()
	require.NoError(t, err)
	listenerAddr := fmt.Sprintf("127.0.0.1:%d", port)

	listener, err := backend.NewListener(listenerAddr)
	t.Cleanup(func() {
		if err = listener.Close(); err != nil {
			t.Log("Error closing listener at address - " + listenerAddr)
		}
	})
	require.NoError(t, err)
	assert.NotNil(t, listener)

	dialer := backend.NewDialer()
	assert.NotNil(t, dialer)
}
