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

package client

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	ppersistence "perun.network/go-perun/channel/persistence"
	pkeyvalue "perun.network/go-perun/channel/persistence/keyvalue"
	pclient "perun.network/go-perun/client"
	plog "perun.network/go-perun/log"
	pleveldb "perun.network/go-perun/pkg/sortedkv/leveldb"
	pwire "perun.network/go-perun/wire"
	pnet "perun.network/go-perun/wire/net"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
)

//go:generate mockery --name Closer --output ../internal/mocks

// Closer is used to call close on database.
type Closer interface {
	Close() error
}

// client is a wrapper type around the state channel client implementation from go-perun.
// It also manages the lifecycle of a message bus that is used for off-chain communication.
type client struct {
	pClient
	msgBus perun.WireBus

	// Registry that is used by the channel client for resolving off-chain address to comm address.
	msgBusRegistry perun.Registerer

	dbPath        string        // Database path, because restore will be called later.
	reconnTimeout time.Duration // Reconn Timeout, because restore will be called later.
	dbConn        Closer        // Database connection for closing it during client.Close.

	wg *sync.WaitGroup
}

// pClient represents the methods on client.Client that are used by client.
type pClient interface {
	ProposeChannel(context.Context, pclient.ChannelProposal) (perun.Channel, error)
	Handle(pclient.ProposalHandler, pclient.UpdateHandler)
	Channel(pchannel.ID) (perun.Channel, error)
	Close() error

	EnablePersistence(ppersistence.PersistRestorer)
	OnNewChannel(handler func(perun.Channel))
	Restore(context.Context) error

	Log() plog.Logger
}

// pclientWrapped is a wrapper around pclient.Client that returns a channel of interface type
// instead of struct type. This enables easier mocking of the returned value in tests.
type pclientWrapped struct {
	*pclient.Client
}

// ProposeChannel is a wrapper around the original function, that returns a channel of interface type instead of
// struct type.
func (c *pclientWrapped) ProposeChannel(ctx context.Context, proposal pclient.ChannelProposal) (perun.Channel, error) {
	return c.Client.ProposeChannel(ctx, proposal)
}

// Channel is a wrapper around the original function, that returns a channel of interface type instead of struct type.
func (c *pclientWrapped) Channel(id pchannel.ID) (perun.Channel, error) {
	return c.Client.Channel(id)
}

// OnNewChannel is a wrapper around the original function, that takes a handler that takes channel of interface type as
// argument instead of the handler in original function that takes channel of struct type as argument.
func (c *pclientWrapped) OnNewChannel(handler func(perun.Channel)) {
	c.Client.OnNewChannel(func(ch *pclient.Channel) {
		handler(ch)
	})
}

// NewEthereumPaymentClient initializes a two party, ethereum payment channel client for the given user.
// It establishes a connection to the blockchain and verifies the integrity of contracts at the given address.
// It uses the comm backend to initialize adapters for off-chain communication network.
func NewEthereumPaymentClient(cfg Config, user perun.User, comm perun.CommBackend) (*client, error) {
	funder, adjudicator, err := connectToChain(cfg.Chain, user.OnChain)
	if err != nil {
		return nil, err
	}
	offChainAcc, err := user.OffChain.Wallet.Unlock(user.OffChain.Addr)
	if err != nil {
		return nil, errors.WithMessage(err, "off-chain account")
	}
	dialer := comm.NewDialer()
	msgBus := pnet.NewBus(offChainAcc, dialer)

	pcClient, err := pclient.New(offChainAcc.Address(), msgBus, funder, adjudicator, user.OffChain.Wallet)
	if err != nil {
		return nil, errors.Wrap(err, "initializing state channel client")
	}

	c := &client{
		pClient:        &pclientWrapped{pcClient},
		msgBus:         msgBus,
		msgBusRegistry: dialer,
		dbPath:         cfg.DatabaseDir,
		reconnTimeout:  cfg.PeerReconnTimeout,
		wg:             &sync.WaitGroup{},
	}

	listener, err := comm.NewListener(user.CommAddr)
	if err != nil {
		return nil, err
	}
	c.runAsGoRoutine(func() { msgBus.Listen(listener) })

	return c, nil
}

// Register registers the comm address for the given off-chain address in the client.
func (c *client) Register(offChainAddr pwire.Address, commAddr string) {
	c.msgBusRegistry.Register(offChainAddr, commAddr)
}

// Handle registers the channel proposal handler and channel update handler for the client.
// It also starts the handle function as a go-routine.
func (c *client) Handle(ph pclient.ProposalHandler, ch pclient.UpdateHandler) {
	c.runAsGoRoutine(func() { c.pClient.Handle(ph, ch) })
}

// RestoreChs will restore the persisted channels. Register OnNewChannel Callback
// before calling this function.
func (c *client) RestoreChs(handler func(perun.Channel)) error {
	c.OnNewChannel(handler)
	db, err := pleveldb.LoadDatabase(c.dbPath)
	if err != nil {
		return errors.Wrap(err, "initializing persistence database in dir - "+c.dbPath)
	}
	c.dbConn = db

	pr := pkeyvalue.NewPersistRestorer(db)
	c.EnablePersistence(pr)
	ctx, cancel := context.WithTimeout(context.Background(), c.reconnTimeout)
	defer cancel()
	return c.Restore(ctx)
}

// Close closes the client and waits until the listener and handler go routines return. It then closes the
// database connection used for persistence.
//
// Close depends on the following mechanisms implemented in client.Close and bus.Close to signal the go-routines:
// 1. When client.Close is invoked, it cancels the Update and Proposal handlers via a context.
// 2. When bus.Close in invoked, it invokes EndpointRegistry.Close that shuts down the listener via onCloseCallback.
func (c *client) Close() error {
	if err := c.pClient.Close(); err != nil {
		return errors.Wrap(err, "closing channel client")
	}
	if busErr := c.msgBus.Close(); busErr != nil {
		return errors.Wrap(busErr, "closing message bus")
	}
	c.wg.Wait()
	return errors.Wrap(c.dbConn.Close(), "closing persistence database")
}

func connectToChain(cfg ChainConfig, cred perun.Credential) (pchannel.Funder,
	pchannel.Adjudicator, error) {
	walletBackend := ethereum.NewWalletBackend()
	assetAddr, err := walletBackend.ParseAddr(cfg.Asset)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "asset address")
	}
	adjudicatorAddr, err := walletBackend.ParseAddr(cfg.Adjudicator)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "adjudicator address")
	}

	chain, err := ethereum.NewChainBackend(cfg.URL, cfg.ConnTimeout, cfg.OnChainTxTimeout, cred)
	if err != nil {
		return nil, nil, err
	}
	err = chain.ValidateContracts(adjudicatorAddr, assetAddr)
	return chain.NewFunder(assetAddr, cred.Addr),
		chain.NewAdjudicator(adjudicatorAddr, cred.Addr), err
}

func (c *client) runAsGoRoutine(f func()) {
	c.wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		f()
	}(c.wg)
}
