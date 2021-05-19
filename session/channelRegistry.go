// Copyright (c) 2021 - for information on the respective copyright owner
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

package session

type (
	// chRegistery stores a list of channels indexed by the channel ID.
	// It preserves the order in which the channels are added.
	chRegistry struct {
		chIdxs map[string]int
		chs    []*Channel
	}
)

// newChRegistry initializes and returns a new channel registry.
//
// The registry is initialized to hold the passed number of channels. If more
// channels are added, the size will automatically be increased.
func newChRegistry(initialChRegistrySize uint16) *chRegistry {
	return &chRegistry{
		chIdxs: make(map[string]int),
		chs:    make([]*Channel, 0, initialChRegistrySize),
	}
}

// put adds the channel to the channel registry.
//
// If a channel with the same channel ID already exists, the new channel
// replaces it. This however, should not occur, as the channel ID is expected to be unique.
func (r *chRegistry) put(ch *Channel) {
	r.chs = append(r.chs, ch)
	r.chIdxs[ch.id] = len(r.chs) - 1
}

// get returns the channel corresponding to the passed channel ID.
// If not channel is found, it returns nil.
func (r *chRegistry) get(chID string) *Channel {
	chIdx, ok := r.chIdxs[chID]
	if !ok {
		return nil
	}
	return r.chs[chIdx]
}

// count returns the count of channels in the registry.
func (r *chRegistry) count() int { return len(r.chs) }

// forEach runs the passed function on each of the channels in the registry.
func (r *chRegistry) forEach(f func(i int, ch *Channel)) {
	for i, ch := range r.chs {
		f(i, ch)
	}
}
