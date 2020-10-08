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

package perun

import (
	"context"
	"math/big"
	"time"

	pchannel "perun.network/go-perun/channel"
	ppersistence "perun.network/go-perun/channel/persistence"
	pclient "perun.network/go-perun/client"
	pLog "perun.network/go-perun/log"
	pwallet "perun.network/go-perun/wallet"
	pwire "perun.network/go-perun/wire"
	pnet "perun.network/go-perun/wire/net"
)

// Peer represents any participant in the off-chain network that the user wants to transact with.
type Peer struct {
	// Name assigned by user for referring to this peer in API requests to the node.
	// It is unique within a session on the node.
	Alias string `yaml:"alias"`

	// Permanent identity used for authenticating the peer in the off-chain network.
	OffChainAddr pwire.Address `yaml:"-"`
	// This field holds the string value of address for easy marshaling / unmarshaling.
	OffChainAddrString string `yaml:"offchain_address"`

	// Address for off-chain communication.
	CommAddr string `yaml:"comm_address"`
	// Type of off-chain communication protocol.
	CommType string `yaml:"comm_type"`
}

// OwnAlias is the alias for the entry of the user's own peer details.
// It will be used when translating addresses in incoming messages / proposals to aliases.
const OwnAlias = "self"

// ContactsReader represents a read only cached list of contacts.
type ContactsReader interface {
	ReadByAlias(alias string) (p Peer, contains bool)
	ReadByOffChainAddr(offChainAddr pwire.Address) (p Peer, contains bool)
}

// Contacts represents a cached list of contacts backed by a storage. Read, Write and Delete methods act on the
// cache. The state of cached list can be written to the storage by using the UpdateStorage method.
type Contacts interface {
	ContactsReader
	Write(alias string, p Peer) error
	Delete(alias string) error
	UpdateStorage() error
}

//go:generate mockery --name CommBackend --output ./internal/mocks

// CommBackend defines the set of methods required for initializing components required for off-chain communication.
// This can be protocols such as tcp, websockets, MQTT.
type CommBackend interface {
	// Returns a listener that can listen for incoming messages at the specified address.
	NewListener(address string) (pnet.Listener, error)

	// Returns a dialer that can dial for new outgoing connections.
	// If timeout is zero, program will use no timeout, but standard OS timeouts may still apply.
	NewDialer() Dialer
}

//go:generate mockery --name Dialer --output ./internal/mocks

// Dialer extends net.Dialer with Registerer interface.
type Dialer interface {
	pnet.Dialer
	Registerer
}

//go:generate mockery --name Registerer --output ./internal/mocks

// Registerer is used to register the commAddr corresponding to an offChainAddr to the wire.Bus in runtime.
type Registerer interface {
	Register(offChainAddr pwire.Address, commAddr string)
}

// Credential represents the parameters required to access the keys and make signatures for a given address.
type Credential struct {
	Addr     pwallet.Address
	Wallet   pwallet.Wallet
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
	PartAddrs []pwallet.Address
}

// Session provides a context for the user to interact with a node. It manages user data (such as IDs, contacts),
// and channel client.
//
// Once established, a user can establish and transact on state channels. All the channels within a session will use
// the same type and version of communication and state channel protocol. If a user desires to use multiple types or
// versions of any protocol, it should request a separate session for each combination of type and version of those.
type Session struct {
	ID   string // ID uniquely identifies a session instance.
	User User

	ChClient ChClient
}

//go:generate mockery --name ChClient --output ./internal/mocks

// ChClient allows the user to establish off-chain channels and transact on these channels.
//
// It allows the user to enable persistence, where all data pertaining to the lifecycle of a channel is
// persisted continuously. When it is enabled, the channel client can be stopped at any point of time and resumed later.
//
// However, the channel client is not responsible if any channel the user was participating in was closed
// with a wrong state when the channel client was not running.
// Hence it is highly recommended not to stop the channel client if there are open channels.
type ChClient interface {
	Registerer
	ProposeChannel(context.Context, pclient.ChannelProposal) (*pclient.Channel, error)
	Handle(pclient.ProposalHandler, pclient.UpdateHandler)
	Channel(pchannel.ID) (*pclient.Channel, error)
	Close() error

	EnablePersistence(ppersistence.PersistRestorer)
	OnNewChannel(handler func(*pclient.Channel))
	Restore(context.Context) error

	Log() pLog.Logger
}

//go:generate mockery --name WireBus --output ./internal/mocks

// WireBus is an extension of the wire.Bus interface in go-perun to include a "Close" method.
// pwire.Bus (in go-perun) is a central message bus over which all clients of a channel network
// communicate. It is used as the transport layer abstraction for the ChClient.
type WireBus interface {
	pwire.Bus
	Close() error
}

// ChainBackend wraps the methods required for instantiating and using components for
// making on-chain transactions and reading on-chain values on a specific blockchain platform.
// The timeout for on-chain transaction should be implemented by the corresponding backend. It is
// up to the implementation to make the value user configurable.
//
// It defines methods for deploying contracts; validating deployed contracts and instantiating a funder, adjudicator.
type ChainBackend interface {
	DeployAdjudicator() (adjAddr pwallet.Address, _ error)
	DeployAsset(adjAddr pwallet.Address) (assetAddr pwallet.Address, _ error)
	ValidateContracts(adjAddr, assetAddr pwallet.Address) error
	NewFunder(assetAddr pwallet.Address) pchannel.Funder
	NewAdjudicator(adjAddr, receiverAddr pwallet.Address) pchannel.Adjudicator
}

// WalletBackend wraps the methods for instantiating wallets and accounts that are specific to a blockchain platform.
type WalletBackend interface {
	ParseAddr(string) (pwallet.Address, error)
	NewWallet(keystore string, password string) (pwallet.Wallet, error)
	UnlockAccount(pwallet.Wallet, pwallet.Address) (pwallet.Account, error)
}

// Currency represents a parser that can convert between string representation of a currency and
// its equivalent value in base unit represented as a big integer.
type Currency interface {
	Parse(string) (*big.Int, error)
	Print(*big.Int) string
}

// NodeConfig represents the configurable parameters of a perun node.
type NodeConfig struct {
	LogLevel string // LogLevel represents the log level for the node and all derived loggers.
	LogFile  string // LogFile represents the file to write logs. Empty string represents stdout.

	ChainURL     string   // Address of the default blockchain node used by the perun node.
	Adjudicator  string   // Address of the default Adjudicator contract used by the perun node.
	Asset        string   // Address of the default Asset Holder contract used by the perun node.
	CommTypes    []string // Communication protocols supported by the node for off-chain communication.
	ContactTypes []string // Contacts Provider backends supported by the node.
	Currencies   []string // Currencies supported by the node.

	ChainConnTimeout time.Duration // Timeout for connecting to blockchain node.
	OnChainTxTimeout time.Duration // Timeout to wait for confirmation of on-chain tx.
	ResponseTimeout  time.Duration // Timeout to wait for a response from the peer / user.
}

// NodeAPI represents the APIs that can be accessed in the context of a perun node.
// Multiple sessions can be opened in a single node. Each instance will have a dedicated
// keystore and contacts provider.
type NodeAPI interface {
	Time() int64
	GetConfig() NodeConfig
	Help() []string
	OpenSession(configFile string) (string, error)

	// This function is used internally to get a SessionAPI instance.
	// Should not be exposed via user API.
	GetSession(string) (SessionAPI, error)
}

//go:generate mockery --name SessionAPI --output ./internal/mocks

// SessionAPI represents the APIs that can be accessed in the context of a perun node.
// First a session has to be instantiated using the NodeAPI. The session can then be used
// open channels and accept channel proposals.
type SessionAPI interface {
	ID() string
	AddContact(Peer) error
	GetContact(alias string) (Peer, error)
	OpenCh(context.Context, BalInfo, App, uint64) (ChInfo, error)
	HandleClose(string, error)
	GetChsInfo() []ChInfo
	HandleUpdate(pclient.ChannelUpdate, *pclient.UpdateResponder)
	HandleProposal(pclient.ChannelProposal, *pclient.ProposalResponder)
	SubChProposals(ChProposalNotifier) error
	UnsubChProposals() error
	RespondChProposal(context.Context, string, bool) (ChInfo, error)
	SubChCloses(ChCloseNotifier) error
	UnsubChCloses() error

	// This function is used internally to get a ChAPI instance.
	// Should not be exposed via user API.
	GetCh(string) (ChAPI, error)
}

type (
	// ChProposalNotifier is the notifier function that is used for sending channel proposal notifications.
	ChProposalNotifier func(ChProposalNotif)

	// ChProposalNotif represents the parameters sent in a channel proposal notifications.
	ChProposalNotif struct {
		ProposalID       string
		OpeningBalInfo   BalInfo
		App              App
		ChallengeDurSecs uint64
		Expiry           int64
	}

	// ChCloseNotifier is the notifier function that is used for sending channel close notifications.
	ChCloseNotifier func(ChCloseNotif)

	// ChCloseNotif represents the parameters sent in a channel close notifications.
	ChCloseNotif struct {
		ClosedChInfo ChInfo
		Error        string
	}
)

//go:generate mockery --name ChAPI --output ./internal/mocks

// ChAPI represents the APIs that can be accessed in the context of a perun channel.
// First a channel has to be initialized using the SessionAPI. The channel can then be used
// send and receive updates.
type ChAPI interface {
	// Methods for reading the channel information is doesn't change.
	// These APIs don't use mutex lock.
	ID() string
	Currency() string
	Parts() []string
	ChallengeDurSecs() uint64

	// Methods to trasact on, close the channel and read its state.
	// These APIs use a mutex lock.
	SendChUpdate(context.Context, StateUpdater) (ChInfo, error)
	SubChUpdates(ChUpdateNotifier) error
	UnsubChUpdates() error
	RespondChUpdate(context.Context, string, bool) (ChInfo, error)
	GetChInfo() ChInfo
	Close(context.Context) (ChInfo, error)
}

type (
	// ChUpdateNotifier is the notifier function that is used for sending channel update notifications.
	ChUpdateNotifier func(ChUpdateNotif)

	// ChUpdateNotif represents the parameters sent in a channel update notifications.
	ChUpdateNotif struct {
		UpdateID       string
		CurrChInfo     ChInfo
		ProposedChInfo ChInfo
		Expiry         int64
	}

	// App represents the app definition and the corresponding app data for a channel.
	App struct {
		Def  pchannel.App
		Data pchannel.Data
	}

	// ChInfo represents the info regarding a channel that will be sent to the user.
	ChInfo struct {
		ChID string
		// Represents the amount held by each participant in the channel.
		BalInfo BalInfo
		// App used in the channel.
		App App
		// Indicates if the state of the channel is marked as "final". Once a state of channel is marked final,
		// no further update can be made to the channel and the channel can be settled on the blockchain
		// instantaneously without a challenge duration. It is used when all participants of a channel decide to
		// close a channel collaboratively.
		IsFinal bool
		// Current Version Number for the channel. This will be zero when a channel is opened and will be incremented
		// during each update. When registering the state on-chain, if different participants register states with
		// different versions, channel will be settled according to the state with highest version number.
		Version string
	}

	// BalInfo represents the Balance information of the channel participants.
	// A valid BalInfo should meet the following conditions (will be validated before using the struct):
	//	1. Lengths of Parts list and Balance list are equal.
	//	2. All entries in Parts list are unique.
	//	3. Parts list has an entry "self", that represents the user of the session.
	//	4. No amount in Balance must be negative.
	BalInfo struct {
		Currency string   // Currency interpreter used to interpret the amounts in the balance.
		Parts    []string // List of aliases of channel participants.
		Bal      []string // Amounts held by each participant in this channel for the given currency.
	}

	// StateUpdater function is the function that will be used for applying state updates.
	StateUpdater func(*pchannel.State)
) // nolint:gofumpt // unknown error, maybe a false positive
