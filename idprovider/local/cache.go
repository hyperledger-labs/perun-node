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
	"github.com/hyperledger-labs/perun-node/idprovider"
)

// idProviderCache represents a cached list of peer IDs indexed by both alias and off-chain address.
// The methods defined over it are safe for concurrent access.
type idProviderCache struct {
	mutex          sync.RWMutex
	walletBackend  perun.WalletBackend
	peerIDsByAlias map[string]perun.PeerID // Stores a list of peer IDs indexed by Alias.
	aliasByAddr    map[string]string       // Stores a list of alias, indexed by off-chain address string.
}

// newIDProviderCache returns a ID Provider cache created from the given map. It indexes the Peer IDs by both alias and
// off-chain address. The off-chain address strings are decoded using the passed backend.
func newIDProviderCache(peerIDsByAlias map[string]perun.PeerID, backend perun.WalletBackend) (*idProviderCache, error) {
	var err error
	aliasByAddr := make(map[string]string)
	for alias, peer := range peerIDsByAlias {
		if peer.OffChainAddr, err = backend.ParseAddr(peer.OffChainAddrString); err != nil {
			return nil, err
		}
		peerIDsByAlias[alias] = peer
		aliasByAddr[peer.OffChainAddrString] = peer.Alias
	}
	return &idProviderCache{
		peerIDsByAlias: peerIDsByAlias,
		aliasByAddr:    aliasByAddr,
		walletBackend:  backend,
	}, nil
}

// ReadByAlias returns the peer ID corresponding to given alias from the cache.
func (c *idProviderCache) ReadByAlias(alias string) (_ perun.PeerID, isPresent bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.readByAlias(alias)
}

func (c *idProviderCache) readByAlias(alias string) (_ perun.PeerID, isPresent bool) {
	var p perun.PeerID
	p, isPresent = c.peerIDsByAlias[alias]
	return p, isPresent
}

// ReadByOffChainAddr returns the peer ID corresponding to given off-chain address from the cache.
func (c *idProviderCache) ReadByOffChainAddr(offChainAddr pwire.Address) (_ perun.PeerID, isPresent bool) {
	if offChainAddr == nil {
		return perun.PeerID{}, false
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	var alias string
	alias, isPresent = c.aliasByAddr[offChainAddr.String()]
	if !isPresent {
		return perun.PeerID{}, false
	}
	return c.readByAlias(alias)
}

// Write adds the peer ID to ID Provider cache. Returns an error if the alias is already used by same or different
// peer ID or if the off-chain address of the peer ID cannot be parsed using the wallet backend of this ID Provider.
func (c *idProviderCache) Write(alias string, p perun.PeerID) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if oldPeerID, ok := c.peerIDsByAlias[alias]; ok {
		if PeerIDEqual(oldPeerID, p) {
			return idprovider.ErrPeerIDAlreadyRegistered
		}
		return idprovider.ErrPeerAliasAlreadyUsed
	}

	var err error
	p.OffChainAddr, err = c.walletBackend.ParseAddr(p.OffChainAddrString)
	if err != nil {
		return errors.Wrap(idprovider.ErrParsingOffChainAddress, err.Error())
	}
	c.peerIDsByAlias[alias] = p
	c.aliasByAddr[p.OffChainAddrString] = alias
	return nil
}

// Delete deletes the peer from ID Provider cache.
// Returns an error if peer corresponding to given alias is not found.
func (c *idProviderCache) Delete(alias string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.peerIDsByAlias[alias]; !ok {
		return errors.New("peer not found in ID Provider")
	}
	delete(c.peerIDsByAlias, alias)
	return nil
}
