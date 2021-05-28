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

package main

import "sync"

// registry stores the mapping of channels to rows in the table.
// channels can by looked up by both row number as well as channel ID.
type registry struct {
	sync.Mutex
	proposals []*channel
	byChID    map[string]int
}

func newRegistry(size int) *registry {
	return &registry{
		proposals: make([]*channel, size),
		byChID:    make(map[string]int),
	}
}

func (r *registry) putAtIndex(index int, p *channel) {
	r.Lock()
	if index < len(r.proposals) {
		r.proposals[index] = p
	}
	r.Unlock()
}

func (r *registry) getByIndex(index int) (p *channel) {
	r.Lock()
	defer r.Unlock()
	if index >= len(r.proposals) {
		return nil
	}
	return r.proposals[index]
}

func (r *registry) setChID(chID string, sNo int) {
	r.Lock()
	r.byChID[chID] = sNo
	r.Unlock()
}

func (r *registry) getByChID(chID string) (p *channel) {
	r.Lock()
	defer r.Unlock()
	index, ok := r.byChID[chID]
	if !ok {
		return nil
	}
	return r.proposals[index]
}
