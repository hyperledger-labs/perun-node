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

import (
	"sync"

	"github.com/pkg/errors"
	pwire "perun.network/go-perun/wire"

	"github.com/hyperledger-labs/perun-node"
)

// idProviderCache represents a cached list of peer IDs indexed by both alias and off-chain address.
// The methods defined over it are safe for concurrent access.
type idProviderCache struct {
	mutex         sync.RWMutex
	walletBackend perun.WalletBackend
	peersByAlias  map[string]perun.Peer // Stores a list of peers indexed by Alias.
	aliasByAddr   map[string]string     // Stores a list of alias, indexed by off-chain address string.
}

// newIDProviderCache returns a ID Provider cache created from the given map. It indexes the Peers by both alias and
// off-chain address. The off-chain address strings are decoded using the passed backend.
func newIDProviderCache(peersByAlias map[string]perun.Peer, backend perun.WalletBackend) (*idProviderCache, error) {
	var err error
	aliasByAddr := make(map[string]string)
	for alias, peer := range peersByAlias {
		if peer.OffChainAddr, err = backend.ParseAddr(peer.OffChainAddrString); err != nil {
			return nil, err
		}
		peersByAlias[alias] = peer
		aliasByAddr[peer.OffChainAddrString] = peer.Alias
	}
	return &idProviderCache{
		peersByAlias:  peersByAlias,
		aliasByAddr:   aliasByAddr,
		walletBackend: backend,
	}, nil
}

// ReadByAlias returns the peer corresponding to given alias from the cache.
func (c *idProviderCache) ReadByAlias(alias string) (_ perun.Peer, isPresent bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.readByAlias(alias)
}

func (c *idProviderCache) readByAlias(alias string) (_ perun.Peer, isPresent bool) {
	var p perun.Peer
	p, isPresent = c.peersByAlias[alias]
	return p, isPresent
}

// ReadByOffChainAddr returns the peer corresponding to given off-chain address from the cache.
func (c *idProviderCache) ReadByOffChainAddr(offChainAddr pwire.Address) (_ perun.Peer, isPresent bool) {
	if offChainAddr == nil {
		return perun.Peer{}, false
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	var alias string
	alias, isPresent = c.aliasByAddr[offChainAddr.String()]
	if !isPresent {
		return perun.Peer{}, false
	}
	return c.readByAlias(alias)
}

// Write adds the peer to ID Provider cache. Returns an error if the alias is already used by same or different peer or,
// if the off-chain address string of the peer cannot be parsed using the wallet backend of this ID Provider.
func (c *idProviderCache) Write(alias string, p perun.Peer) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if oldPeer, ok := c.peersByAlias[alias]; ok {
		if PeerEqual(oldPeer, p) {
			return errors.New("peer already present in ID Provider")
		}
		return errors.New("alias already used by another peer in ID Provider")
	}

	var err error
	p.OffChainAddr, err = c.walletBackend.ParseAddr(p.OffChainAddrString)
	if err != nil {
		return err
	}
	c.peersByAlias[alias] = p
	c.aliasByAddr[p.OffChainAddrString] = alias
	return nil
}

// Delete deletes the peer from ID Provider cache.
// Returns an error if peer corresponding to given alias is not found.
func (c *idProviderCache) Delete(alias string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.peersByAlias[alias]; !ok {
		return errors.New("peer not found in ID Provider")
	}
	delete(c.peersByAlias, alias)
	return nil
}
