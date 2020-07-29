// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/dst-go
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

package perun

import (
	"context"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/client"
	perunLog "perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/wallet"
)

// Peer represents any participant in the off-chain network that the user wants to transact with.
type Peer struct {
	// Name assigned by user for referring to this peer in api requests to the node.
	// It is unique within a session on the node.
	Alias string

	OffChainAddr peer.Address // Permanent identity used for authenticating the peer in the off-chain network.

	CommAddr string // Address for off-chain communication.
	CommType string // Type of off-chain communication protocol.
}

//go:generate mockery -name CommBackend -output ./internal/mocks

// CommBackend defines the set of methods required for initializing components required for off-chain communication.
// This can be protocols such as tcp, websockets, MQTT.
type CommBackend interface {
	// Returns a listener that can listen for incoming messages at the specified address.
	NewListener(address string) (peer.Listener, error)

	// Returns a dialer that can dial for new outgoing connections.
	// If timeout is zero, program will use no timeout, but standard OS timeouts may still apply.
	NewDialer() peer.Dialer
}

// Credential represents the parameters required to access the keys and make signatures for a given address.
type Credential struct {
	Addr     wallet.Address
	Wallet   wallet.Wallet
	Keystore string
	Password string
}

// User represents a participant in the off-chain network that uses a session on this node for sending transactions.
type User struct {
	Peer

	OnChain  Credential // Account for funding the channel and the on-chain transactions.
	OffChain Credential // Account (corresponding to off-chain address) used for signing authentication messages.

	// List of participant addresses for this user in each open channel.
	// OffChain credential is used for managing all these accounts.
	PartAddrs []wallet.Address
}

// Session provides a context for the user to interact with a node. It manages user data (such as IDs, contacts),
// and channel client.
//
// Once established, a user can establish and transact on state channels. All the channels within a session will use
// the same type and version of communication and state channel protocol. If a user desires to use multiple types or
// versions of any protocol, it should request a seprate session for each combination of type and version of those.
type Session struct {
	ID   string // ID uniquely identifies a session instance.
	User User

	ChannelClient ChannelClient
}

// ChannelClient allows the user to establish off-chain channels and transact on these channels.
//
// It allows the user to enable persistence, where all data pertaining to the lifecycle of a channel is
// persisted continuously. When it is enabled, the channel client can be stopped at any point of time and resumed later.
//
// However, the channel client is not responsible if any channel the user was participating in was closed
// with a wrong state when the channel client was not running.
// Hence it is highly recommended not to stop the channel client if there are open channels.
type ChannelClient interface {
	Listen(listener peer.Listener)

	ProposeChannel(ctx context.Context, req *client.ChannelProposal) (*client.Channel, error)
	Handle(ph client.ProposalHandler, uh client.UpdateHandler)
	Channel(id channel.ID) (*client.Channel, error)
	Close() error

	EnablePersistence(pr persistence.PersistRestorer)
	OnNewChannel(handler func(*client.Channel))
	Reconnect(ctx context.Context) error

	Log() perunLog.Logger
}

// ChainBackend wraps the methods required for instantiating and using components for
// making on-chain transactions and reading on-chain values on a specific blockchain platform.
// The timeout for on-chain transaction should be implemented by the corresponding backend. It is
// upto the implementation to make the value user configurable.
//
// It defines methods for deploying contracts; validating deployed contracts and instantiating a funder, adjudicator.
type ChainBackend interface {
	DeployAdjudicator() (wallet.Address, error)
	DeployAsset(adjAddr wallet.Address) (wallet.Address, error)
	ValidateContracts(adjAddr, assetAddr wallet.Address) error
	NewFunder(assetAddr wallet.Address) channel.Funder
	NewAdjudicator(adjAddr, receiverAddr wallet.Address) channel.Adjudicator
}

// WalletBackend wraps the methods for instantiating wallets and accounts that are specific to a blockchain platform.
type WalletBackend interface {
	ParseAddr(string) (wallet.Address, error)
	NewWallet(keystore string, password string) (wallet.Wallet, error)
	UnlockAccount(wallet.Wallet, wallet.Address) (wallet.Account, error)
} // nolint:gofumpt // unknown error, maybe a false positive
