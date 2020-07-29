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

package tcp

import (
	"time"

	"github.com/pkg/errors"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/peer/net"
)

// Backend is an off-chain communication backend that implements `CommBackend` for
// for tcp protocol. It stores configuration required for initializing the adapters.
type Backend struct {
	// timeout to be used when dialing for new outgoing connections.
	dialerTimeout time.Duration
}

// NewListener returns a listener that can listen for incomig connections at
// the specified address using tcp protocol.
func (b Backend) NewListener(addr string) (peer.Listener, error) {
	listener, err := net.NewListener("tcp", addr)
	return listener, errors.Wrap(err, "initializing listener")
}

// NewDialer returns a dialer that can dial outgoing connections using on
// tcp protocol.
//
// It uses the dial timeout configured during backend initialization.
// If the duration was set to zero, this program will not use any timeout.
// However default timeouts based on the operating system will still apply.
func (b Backend) NewDialer() peer.Dialer {
	return net.NewDialer("tcp", b.dialerTimeout)
}

// NewTCPBackend returns a backend that can initialize off-chain communication
// adapters for tcp protocol.
//
// The provided dialerTimeout will be used when dialing for new outgoing connections.
// If the duration was set to zero, this program will not use any timeout.
// However default timeouts based on the operating system will still apply.
func NewTCPBackend(dialerTimeout time.Duration) Backend {
	return Backend{dialerTimeout: dialerTimeout}
}
