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

package client

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence/keyvalue"
	"perun.network/go-perun/client"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/sortedkv/leveldb"

	"github.com/direct-state-transfer/perun-node"
	"github.com/direct-state-transfer/perun-node/blockchain/ethereum"
)

// Client is a wrapper type around the state channel client implementation from go-perun.
// It also manages the lifecycle of listener and handler go-routines.
type Client struct {
	*client.Client

	wg *sync.WaitGroup
}

// NewEthereumPaymentClient initializes a two party, ethereum payment channel client for the given user.
// It establishes a connection to the blockchain and verifies the integrity of contracts at the given address.
// It uses the comm backend to initialize adapters for off-chain communication network.
func NewEthereumPaymentClient(cfg Config, user perun.User, comm perun.CommBackend) (*Client, error) {
	dialer, listener, err := initComm(comm, user.CommAddr)
	if err != nil {
		return nil, err
	}
	funder, adjudicator, err := connectToChain(cfg.Chain, user.OnChain)
	if err != nil {
		return nil, err
	}
	// Only off-chain account is unlocked. Accounts for the participant addresses
	// will be unlocked by the client when required.
	offChainAcc, err := user.OffChain.Wallet.Unlock(user.OffChain.Addr)
	if err != nil {
		return nil, errors.WithMessage(err, "off-chain account")
	}

	c := client.New(offChainAcc, dialer, funder, adjudicator, user.OffChain.Wallet)
	if err := loadPersister(c, cfg.DatabaseDir, cfg.PeerReconnTimeout); err != nil {
		return nil, err
	}
	client := &Client{
		Client: c,
		wg:     &sync.WaitGroup{},
	}

	client.runAsGoRoutine(func() { client.Handle(&proposalHandler{}, &updateHandler{}) })
	client.runAsGoRoutine(func() { client.Listen(listener) })

	return client, nil
}

// Close closes the client and waits until the listener and handler go routines return.
//
// Close depends on the following mechanisms implemented in client.Close:
// 1. A context is passed to the infinite loop of handle function. When client.Close is called,
//    this context is canceled, thereby causing the infinite loop in Handle function to stop.
// 2. A callback to close the listener (listener.Close()) is registered to the client.Closer.
//    When client.Close is called, this call back signals the listener to shutdown and hence
//    stop the infinite loop in Listen function.
func (c *Client) Close() error {
	err := c.Client.Close()
	c.wg.Wait()
	return errors.Wrap(err, "closing channel client")
}

func initComm(comm perun.CommBackend, listenAddr string) (peer.Dialer, peer.Listener, error) {
	dialer := comm.NewDialer()
	listener, err := comm.NewListener(listenAddr)
	return dialer, listener, err
}

func connectToChain(cfg ChainConfig, cred perun.Credential) (channel.Funder, channel.Adjudicator, error) {
	walletBackend := ethereum.NewWalletBackend()
	assetAddr, err := walletBackend.ParseAddr(cfg.Asset)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "asset address")
	}
	adjudicatorAddr, err := walletBackend.ParseAddr(cfg.Adjudicator)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "adjudicator address")
	}

	chain, err := ethereum.NewChainBackend(cfg.URL, cfg.ConnTimeout, cred)
	if err != nil {
		return nil, nil, err
	}
	err = chain.ValidateContracts(adjudicatorAddr, assetAddr)
	return chain.NewFunder(assetAddr), chain.NewAdjudicator(adjudicatorAddr, cred.Addr), err
}

func loadPersister(c *client.Client, dbPath string, reconnTimeout time.Duration) error {
	db, err := leveldb.LoadDatabase(dbPath)
	if err != nil {
		return errors.Wrap(err, "initializing persistence database in dir - "+dbPath)
	}
	pr, err := keyvalue.NewPersistRestorer(db)
	if err != nil {
		return errors.Wrap(err, "initializing persister restorer")
	}
	c.EnablePersistence(pr)
	ctx, cancel := context.WithTimeout(context.Background(), reconnTimeout)
	defer cancel()
	return c.Reconnect(ctx)
}

func (c *Client) runAsGoRoutine(f func()) {
	c.wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		f()
	}(c.wg)
}

type proposalHandler struct{}

// HandleProposal implements the client.ProposalHandler interface defined in go-perun.
// This method is called on every incoming channel proposal.
// TODO: (mano) Implement an accept all handler until user api components are implemented.
// TODO: (mano) Replace with proper implementation after user api components are implemented.
func (ph *proposalHandler) HandleProposal(_ *client.ChannelProposal, _ *client.ProposalResponder) {
	panic("proposalHandler.HandleProposal not implemented")
}

type updateHandler struct{}

// HandleUpdate implements the UpdateHandler interface.
// This method is called on every incoming state update for any channel managed by this client.
// TODO: (mano) Implement an accept all handler until user api components are implemented.
// TODO: (mano) Replace with proper implementation after user api components are implemented.
func (uh *updateHandler) HandleUpdate(_ client.ChannelUpdate, _ *client.UpdateResponder) {
	panic("updateHandler.HandleUpdate not implemented")
}
