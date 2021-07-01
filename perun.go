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
	pwallet "perun.network/go-perun/wallet"
	pwire "perun.network/go-perun/wire"
	pnet "perun.network/go-perun/wire/net"
)

// PeerID represents any participant in the off-chain network that the user wants to transact with.
type PeerID struct {
	// Name assigned by user for referring to this PeerID in API requests to the node.
	// It is unique within a session on the node.
	Alias string `yaml:"alias"`

	// Permanent identity used for authenticating the PeerID in the off-chain network.
	OffChainAddr pwire.Address `yaml:"-"`
	// This field holds the string value of address for easy marshaling / unmarshaling.
	OffChainAddrString string `yaml:"offchain_address"`

	// Address for off-chain communication.
	CommAddr string `yaml:"comm_address"`
	// Type of off-chain communication protocol.
	CommType string `yaml:"comm_type"`
}

// OwnAlias is the alias for the entry of the user's own PeerID details.
// It will be used when translating addresses in incoming messages / proposals to aliases.
const OwnAlias = "self"

// IDReader represents the functions to read peer IDs from a cache connected to a peer ID provider.
type IDReader interface {
	ReadByAlias(alias string) (p PeerID, contains bool)
	ReadByOffChainAddr(offChainAddr pwire.Address) (p PeerID, contains bool)
}

// IDProvider represents the functions to read, write peer IDs from and to the local cache connected to a
// peer ID provider. It also includes a function to sync the changes in the cache with the ID provider backend.
type IDProvider interface {
	IDReader
	Write(alias string, p PeerID) error
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

//go:generate mockery --name WireBus --output ./internal/mocks

// WireBus is an extension of the wire.Bus interface in go-perun to include a "Close" method.
// pwire.Bus (in go-perun) is a central message bus over which all clients of a channel network
// communicate. It is used as the transport layer abstraction for the ChClient.
type WireBus interface {
	pwire.Bus
	Close() error
}

// Credential represents the parameters required to access the keys and make signatures for a given address.
type Credential struct {
	Addr     pwallet.Address
	Wallet   pwallet.Wallet
	Keystore string
	Password string
}

// ChainBackend wraps the methods required for deploy contracts, validating
// them and instantiating funde, adjudicator instances.
//
// The timeout for on-chain transaction should be implemented by the
// corresponding backend. It is up to the implementation to make the value user
// configurable.
type ChainBackend interface {
	ROChainBackend

	DeployAdjudicator(txSender pwallet.Address) (pwallet.Address, error)
	DeployAssetETH(adjudicator, txSender pwallet.Address) (pwallet.Address, error)
	DeployAssetERC20(adjudicator, tokenERC20, txSender pwallet.Address) (pwallet.Address, error)
	DeployPerunToken(initAccs []pwallet.Address, initBal *big.Int, txSender pwallet.Address) (pwallet.Address, error)

	NewFunder(assetETH, txSender pwallet.Address) pchannel.Funder
	NewAdjudicator(adjudicator, txSender pwallet.Address) pchannel.Adjudicator
}

// ROChainBackend wraps the methods required for validating contracts.
//
// The timeout for on-chain transaction should be implemented by the
// corresponding backend. It is up to the implementation to make the value user
// configurable.
type ROChainBackend interface {
	ValidateAdjudicator(adjudicator pwallet.Address) error
	ValidateAssetETH(adjudicator, assetETH pwallet.Address) error
	ValidateAssetERC20(adjudicator, tokenERC20, assetERC20 pwallet.Address) (symbol string, maxDecimals uint8, _ error)
	ERC20Info(token pwallet.Address) (symbol string, decimal uint8, _ error)
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
	Symbol() string
}

// CurrencyRegistry provides an interface to register and retrieve currency
// parsers.
type CurrencyRegistry interface {
	ROCurrencyRegistry
	Register(symbol string, maxDecimals uint8) (Currency, error)
}

// ROCurrencyRegistry provides an interface to retrieve currency parsers.
type ROCurrencyRegistry interface {
	IsRegistered(symbol string) bool
	Currency(symbol string) Currency
	Symbols() []string
}

// ContractRegistry provides an interface to register and retrieve adjudicator
// and asset contracts.
type ContractRegistry interface {
	ROContractRegistry
	RegisterAssetERC20(token, asset pwallet.Address) (symbol string, maxDecimals uint8, _ error)
}

//go:generate mockery --name ROContractRegistry --output ./internal/mocks

// ROContractRegistry provides an interface to retrieve contracts.
type ROContractRegistry interface {
	Adjudicator() pwallet.Address
	AssetETH() pwallet.Address
	Asset(symbol string) (asset pwallet.Address, found bool)
	Symbol(asset pwallet.Address) (symbol string, found bool)
	Assets() map[string]string
}

// NodeConfig represents the configurable parameters of a perun node.
type NodeConfig struct {
	// User configurable values.
	LogLevel         string            // LogLevel represents the log level for the node and all derived loggers.
	LogFile          string            // LogFile represents the file to write logs. Empty string represents stdout.
	ChainURL         string            // URL of the blockchain node.
	ChainID          int               // See session.chainconfig.
	Adjudicator      string            // Address of the Adjudicator contract.
	AssetETH         string            // Address of the ETH Asset holder contract.
	AssetERC20s      map[string]string // Address of ERC20 token contracts and corresponding asset contracts.
	ChainConnTimeout time.Duration     // Timeout for connecting to blockchain node.
	OnChainTxTimeout time.Duration     // Timeout to wait for confirmation of on-chain tx.
	ResponseTimeout  time.Duration     // Timeout to wait for a response from the peer / user.

	// Hard coded values. See cmd/perunnode/run.go.
	CommTypes            []string // Communication protocols supported by the node for off-chain communication.
	IDProviderTypes      []string // ID Provider types supported by the node.
	CurrencyInterpreters []string // Currencies Interpreters supported by the node.

}

// APIError represents the newer version of error returned by node, session
// and channel APIs.
//
// Along with the error message, this error type assigns to each error
// an error category that describes how the error should be handled,
// an error code that identifies specific types of error and
// additional info that contains data related to the error as key value pairs.
type APIError interface {
	Category() ErrorCategory
	Code() ErrorCode
	Message() string
	AddInfo() interface{}
	Error() string
}

// ErrorCategory represents the category of the error, which describes how the
// error should be handled by the client.
type ErrorCategory int

const (
	// ParticipantError is caused by one of the channel participants not acting
	// as per the perun protocol.
	//
	// To resolve this, the client should negotiate with the peer outside of
	// this system to act in accordance with the perun protocol.
	ParticipantError ErrorCategory = iota

	// ClientError is caused by the errors in the request from the client. It
	// could be errors in arguments or errors in configuration provided by the
	// client to access the external systems or errors in the state of external
	// systems not managed by the node.
	//
	// To resolve this, the client should provide valid arguments, provide
	// correct configuration to access the external systems or fix the external
	// systems; and then retry.
	ClientError

	// ProtocolFatalError is caused when the protocol aborts due to unexpected
	// failure in external system during execution. It could also result in loss
	// of funds.
	//
	// To resolve this, user should manually inspect the error message and
	// handle it.
	ProtocolFatalError
	// InternalError is caused due to unintended behavior in the node software.
	//
	// To resolve this, user should manually inspect the error message and
	// handle it.
	InternalError
)

// String implements the stringer interface for ErrorCategory.
func (c ErrorCategory) String() string {
	return [...]string{
		"Participant",
		"Client",
		"Protocol Fatal",
		"Internal",
	}[c]
}

// ErrorCode is a numeric code assigned to identify the specific type of error.
// The keys in the additional field is fixed for each error code.
type ErrorCode int

// Error code definitions.
const (
	ErrPeerRequestTimedOut  ErrorCode = 101
	ErrPeerRejected         ErrorCode = 102
	ErrPeerNotFunded        ErrorCode = 103
	ErrUserResponseTimedOut ErrorCode = 104
	ErrResourceNotFound     ErrorCode = 201
	ErrResourceExists       ErrorCode = 202
	ErrInvalidArgument      ErrorCode = 203
	ErrFailedPreCondition   ErrorCode = 204
	ErrInvalidConfig        ErrorCode = 205
	ErrInvalidContracts     ErrorCode = 206
	ErrTxTimedOut           ErrorCode = 301
	ErrChainNotReachable    ErrorCode = 302
	ErrUnknownInternal      ErrorCode = 401
)

type (
	// ErrInfoPeerRequestTimedOut represents the fields in the additional
	// info for ErrPeerRequestTimedOut.
	ErrInfoPeerRequestTimedOut struct {
		PeerAlias string
		Timeout   string
	}

	// ErrInfoPeerRejected represents the fields in the additional info for
	// ErrRejectedByPeer.
	ErrInfoPeerRejected struct {
		PeerAlias string
		Reason    string
	}

	// ErrInfoPeerNotFunded represents the fields in the additional info for
	// ErrPeerNotFunded.
	ErrInfoPeerNotFunded struct {
		PeerAlias string
	}

	// ErrInfoUserResponseTimedOut represents the fields in the additional info for
	// ErrUserResponseTimedOut.
	ErrInfoUserResponseTimedOut struct {
		Expiry     int64
		ReceivedAt int64
	}

	// ErrInfoResourceNotFound represents the fields in the additional info for
	// ErrResourceNotFound.
	ErrInfoResourceNotFound struct {
		Type string
		ID   string
	}

	// ErrInfoResourceExists represents the fields in the additional info for
	// ErrResourceExists.
	ErrInfoResourceExists struct {
		Type string
		ID   string
	}

	// ErrInfoInvalidArgument represents the fields in the additional info for
	// ErrInvalidArgument.
	ErrInfoInvalidArgument struct {
		Name        string
		Value       string
		Requirement string
	}

	// ErrInfoFailedPreCondUnclosedChs represents the fields in the
	// additional info for the ErrFailedPreCondition when session closed is
	// called without force option and the session has unclosed channels.
	//
	// This additional info should not be used in any other context.
	ErrInfoFailedPreCondUnclosedChs struct {
		ChInfos []ChInfo
	}

	// ErrInfoInvalidConfig represents the fields in the additional info for
	// ErrInfoInvalidConfig.
	ErrInfoInvalidConfig struct {
		Name  string
		Value string
	}

	// ContractErrInfo is used to pass the contract information (name, address,
	// error message) encountered when validating a contract.
	ContractErrInfo struct {
		Name    string
		Address string
		Error   string
	}

	// ErrInfoInvalidContracts represents the fields in the additional info for
	// ErrInvalidContracts.
	ErrInfoInvalidContracts struct {
		ContractErrInfos []ContractErrInfo
	}

	// ErrInfoTxTimedOut represents the fields in the additional info
	// for ErrTxTimedOut.
	ErrInfoTxTimedOut struct {
		TxType    string
		TxID      string
		TxTimeout string
	}

	// ErrInfoChainNotReachable represents the fields in the additional info
	// for ErrChainNotReachable.
	ErrInfoChainNotReachable struct {
		ChainURL string
	}
)

// NodeAPI represents the APIs that can be accessed in the context of a perun node.
// Multiple sessions can be opened in a single node. Each instance will have a dedicated
// keystore and ID provider.
type NodeAPI interface {
	Time() int64
	GetConfig() NodeConfig
	Help() []string
	OpenSession(configFile string) (string, []ChInfo, APIError)

	RegisterCurrency(tokenAddr, assetAddr string) (symbol string, _ APIError)

	// This function is used internally to get a SessionAPI instance.
	// Should not be exposed via user API.
	GetSession(string) (SessionAPI, APIError)
}

//go:generate mockery --name SessionAPI --output ./internal/mocks

// SessionAPI represents the APIs that can be accessed in the context of a perun node.
// First a session has to be instantiated using the NodeAPI. The session can then be used
// open channels and accept channel proposals.
type SessionAPI interface {
	ID() string
	AddPeerID(PeerID) APIError
	GetPeerID(alias string) (PeerID, APIError)
	OpenCh(context.Context, BalInfo, App, uint64) (ChInfo, APIError)
	GetChsInfo() []ChInfo
	SubChProposals(ChProposalNotifier) APIError
	UnsubChProposals() APIError
	RespondChProposal(context.Context, string, bool) (ChInfo, APIError)
	Close(force bool) ([]ChInfo, APIError)

	DeployAssetERC20(tokenERC20 string) (asset string, _ APIError)

	// This function is used internally to get a ChAPI instance.
	// Should not be exposed via user API.
	GetCh(string) (ChAPI, APIError)
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
)

//go:generate mockery --name ChAPI --output ./internal/mocks

// ChAPI represents the APIs that can be accessed in the context of a perun channel.
// First a channel has to be initialized using the SessionAPI. The channel can then be used
// send and receive updates.
type ChAPI interface {
	// Methods for reading the channel information is doesn't change.
	// These APIs don't use mutex lock.
	ID() string
	Currencies() []Currency
	Parts() []string
	ChallengeDurSecs() uint64

	// Methods to transact on, close the channel and read its state.
	// These APIs use a mutex lock.
	SendChUpdate(context.Context, StateUpdater) (ChInfo, APIError)
	SubChUpdates(ChUpdateNotifier) APIError
	UnsubChUpdates() APIError
	RespondChUpdate(context.Context, string, bool) (ChInfo, APIError)
	GetChInfo() ChInfo
	Close(context.Context) (ChInfo, APIError)
}

// Enumeration of values for ChUpdateType:
// Open: If accepted, channel will be updated and it will remain in open for off-chain tx.
// Final: If accepted, channel will be updated and closed (settled on-chain and amount withdrawn).
// Closed: Channel has been closed (settled on-chain and amount withdrawn).
const (
	ChUpdateTypeOpen ChUpdateType = iota
	ChUpdateTypeFinal
	ChUpdateTypeClosed
)

type (
	// ChUpdateType is the type of channel update. It can have three values: "open", "final" and "closed".
	ChUpdateType uint8

	// ChUpdateNotifier is the notifier function that is used for sending channel update notifications.
	ChUpdateNotifier func(ChUpdateNotif)

	// ChUpdateNotif represents the parameters sent in a channel update notification.
	// The update can be of two types
	// 1. Regular update proposed by the peer to progress the off-chain state of the channel.
	// 2. Closing update when a channel is closed, balance is settled on the blockchain and
	// the amount corresponding to this user is withdrawn.
	//
	// The two types of updates can be differentiated using the status field,
	// which is "open" or "final" for a regular update and "closed" for a closing update.
	//
	ChUpdateNotif struct {
		// UpdateID denotes the unique ID for this update. It is derived from the channel ID and version number.
		UpdateID       string
		CurrChInfo     ChInfo
		ProposedChInfo ChInfo

		Type ChUpdateType

		// It is with reference to the system clock on the computer running the perun-node.
		// Time (in unix timestamp) before which response to this notification should be sent.
		//
		// It is 0, when no response is expected.
		Expiry int64

		// Error represents any error encountered while processing incoming updates or
		// while a channel is closed by the watcher.
		// When this is non empty, expiry will also be zero and no response is expected
		Error APIError
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
		// Current Version Number for the channel. This will be zero when a channel is opened and will be incremented
		// during each update. When registering the state on-chain, if different participants register states with
		// different versions, channel will be settled according to the state with highest version number.
		Version string
	}

	// BalInfo represents the Balance information of the channel participants.
	// Bal[0] represents the balance of the channel for asset Currency[0] for
	// the all the channel participants as mentioned in Parts; Bal[1] specifies
	// for asset Currency[1] and so on.
	//
	// A valid BalInfo should meet the following conditions (will be validated
	// before using the struct):
	//  1. Length of Currencies and outer length of Bal are equal.
	//	1. Lengths of Parts and inner length of Bal are equal.
	//	2. All entries in Parts list are unique unique.
	//	3. Parts list has an entry "self", that represents the user of the session.
	//	4. No amount in Balance must be negative.
	//
	BalInfo struct {
		Currencies []string   // List of currencies for the specifying amounts in the balance.
		Parts      []string   // List of aliases of channel participants.
		Bals       [][]string // Amounts held by each participant in this channel for the each currency.
	}

	// StateUpdater function is the function that will be used for applying state updates.
	StateUpdater func(*pchannel.State) error
)
