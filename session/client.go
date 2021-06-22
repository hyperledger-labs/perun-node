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

package session

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

type (
	// ChClient allows the user to establish off-chain channels and transact on
	// these channels. The channel data are continuously persisted and hence it can
	// be shutdown and restarted without loosing any data.
	//
	// However, care should be taken when shutting down the client when it has open
	// channels.Because if any channel the user was participating in was closed
	// with a wrong state when the channel client was not running, dispute
	// resolution process will not be triggered.
	//
	// This interface is defined for isolating the client type from session.
	ChClient interface {
		perun.Registerer
		ProposeChannel(context.Context, pclient.ChannelProposal) (PChannel, error)
		Handle(pclient.ProposalHandler, pclient.UpdateHandler)
		Channel(pchannel.ID) (PChannel, error)
		Close() error

		EnablePersistence(ppersistence.PersistRestorer)
		OnNewChannel(handler func(PChannel))
		Restore(context.Context) error
		RestoreChs(func(PChannel)) error

		Log() plog.Logger
	}

	// Closer is used to call close on database.
	Closer interface {
		Close() error
	}

	// client is a wrapper type around the state channel client implementation from go-perun that
	// also manages the lifecycle of a message bus that is used for off-chain communication.
	//
	// It implements ChClient interface.
	client struct {
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
	// This interface is defined for abstracting the client.Client from the
	// client instance for the pupose of mocking in tests.
	pClient interface {
		ProposeChannel(context.Context, pclient.ChannelProposal) (PChannel, error)
		Handle(pclient.ProposalHandler, pclient.UpdateHandler)
		Channel(pchannel.ID) (PChannel, error)
		Close() error

		EnablePersistence(ppersistence.PersistRestorer)
		OnNewChannel(handler func(PChannel))
		Restore(context.Context) error

		Log() plog.Logger
	}

	// pclientWrapped is a wrapper around pclient.Client that returns a channel of interface type
	// instead of struct type. This enables easier mocking of the returned value in tests.
	pclientWrapped struct {
		*pclient.Client
	}
)

//go:generate mockery --name ChClient --output ../internal/mocks

//go:generate mockery --name Closer --output ../internal/mocks

// ProposeChannel is a wrapper around the original function, that returns a channel of interface type instead of
// struct type.
func (c *pclientWrapped) ProposeChannel(ctx context.Context, proposal pclient.ChannelProposal) (PChannel, error) {
	return c.Client.ProposeChannel(ctx, proposal)
}

// Channel is a wrapper around the original function, that returns a channel of interface type instead of struct type.
func (c *pclientWrapped) Channel(id pchannel.ID) (PChannel, error) {
	return c.Client.Channel(id)
}

// OnNewChannel is a wrapper around the original function, that takes a handler that takes channel of interface type as
// argument instead of the handler in original function that takes channel of struct type as argument.
func (c *pclientWrapped) OnNewChannel(handler func(PChannel)) {
	c.Client.OnNewChannel(func(ch *pclient.Channel) {
		handler(ch)
	})
}

// newEthereumPaymentClient initializes a two party, ethereum payment channel client for the given user.
// It establishes a connection to the blockchain and verifies the integrity of contracts at the given address.
// It uses the comm backend to initialize adapters for off-chain communication network.
func newEthereumPaymentClient(cfg clientConfig, user User, comm perun.CommBackend) (
	ChClient, perun.APIError) {
	funder, adjudicator, apiErr := connectToChain(cfg.Chain, user.OnChain)
	if apiErr != nil {
		return nil, apiErr
	}
	offChainAcc, err := user.OffChain.Wallet.Unlock(user.OffChain.Addr)
	if err != nil {
		return nil, perun.NewAPIErrUnknownInternal(errors.WithMessage(err, "off-chain account"))
	}
	dialer := comm.NewDialer()
	msgBus := pnet.NewBus(offChainAcc, dialer)

	pcClient, err := pclient.New(offChainAcc.Address(), msgBus, funder, adjudicator, user.OffChain.Wallet)
	if err != nil {
		return nil, perun.NewAPIErrUnknownInternal(errors.WithMessage(err, "off-chain account"))
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
		return nil, perun.NewAPIErrInvalidConfig(err, "commAddr", user.CommAddr)
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
func (c *client) RestoreChs(handler func(PChannel)) error {
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
	err = c.Restore(ctx)
	// Set the OnNewChannel call back to a dummy function, so it does not
	// process the channels that are created as a result of `ProposeChannel` or
	// `Accept` on a channel proposal.
	c.OnNewChannel(func(PChannel) {})
	return err
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

func connectToChain(cfg ChainConfig, cred perun.Credential) (pchannel.Funder, pchannel.Adjudicator, perun.APIError) {
	chain, err := ethereum.NewChainBackend(cfg.URL, cfg.ChainID, cfg.ConnTimeout, cfg.OnChainTxTimeout, cred)
	if err != nil {
		err = errors.WithMessage(err, "connecting to blockchain")
		return nil, nil, perun.NewAPIErrInvalidConfig(err, "chainURL", cfg.URL)
	}
	return chain.NewFunder(cfg.AssetETH, cred.Addr), chain.NewAdjudicator(cfg.Adjudicator, cred.Addr), nil
}

func (c *client) runAsGoRoutine(f func()) {
	c.wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		f()
	}(c.wg)
}
