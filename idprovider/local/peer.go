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

package local

import "github.com/hyperledger-labs/perun-node"

// PeerEqual returns true if all fields in the Peer ID except OffChainAddr are equal.
func PeerEqual(p1, p2 perun.PeerID) bool {
	return p1.Alias == p2.Alias && p1.OffChainAddrString == p2.OffChainAddrString &&
		p1.CommType == p2.CommType && p1.CommAddr == p2.CommAddr
}
