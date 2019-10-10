// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// LibSignaturesABI is the input ABI used to generate the binding from.
const LibSignaturesABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"message\",\"type\":\"bytes32\"},{\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"verify\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// LibSignaturesBin is the compiled bytecode used for deploying new contracts.
const LibSignaturesBin = `0x608060405234801561001057600080fd5b506101ff806100206000396000f3006080604052600436106100405763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a86b5508114610045575b600080fd5b34801561005157600080fd5b50604080516020600460443581810135601f81018490048402850184019095528484526100bb94823573ffffffffffffffffffffffffffffffffffffffff169460248035953695946064949201919081908401838280828437509497506100cf9650505050505050565b604080519115158252519081900360200190f35b600080600080845160411415156100e957600093506101c9565b50505060208201516040830151606084015160001a601b60ff8216101561010e57601b015b8060ff16601b1415801561012657508060ff16601c14155b1561013457600093506101c9565b60408051600080825260208083018085528a905260ff8516838501526060830187905260808301869052925173ffffffffffffffffffffffffffffffffffffffff8b169360019360a0808201949293601f198101939281900390910191865af11580156101a5573d6000803e3d6000fd5b5050506020604051035173ffffffffffffffffffffffffffffffffffffffff161493505b50505093925050505600a165627a7a72305820e6be47295531dbab9d03e72c0105a751b261bd915c5eee2abb07c427707e57b30029`

// DeployLibSignatures deploys a new Ethereum contract, binding an instance of LibSignatures to it.
func DeployLibSignatures(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *LibSignatures, error) {
	parsed, err := abi.JSON(strings.NewReader(LibSignaturesABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(LibSignaturesBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &LibSignatures{LibSignaturesCaller: LibSignaturesCaller{contract: contract}, LibSignaturesTransactor: LibSignaturesTransactor{contract: contract}, LibSignaturesFilterer: LibSignaturesFilterer{contract: contract}}, nil
}

// LibSignatures is an auto generated Go binding around an Ethereum contract.
type LibSignatures struct {
	LibSignaturesCaller     // Read-only binding to the contract
	LibSignaturesTransactor // Write-only binding to the contract
	LibSignaturesFilterer   // Log filterer for contract events
}

// LibSignaturesCaller is an auto generated read-only Go binding around an Ethereum contract.
type LibSignaturesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LibSignaturesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LibSignaturesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LibSignaturesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LibSignaturesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LibSignaturesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LibSignaturesSession struct {
	Contract     *LibSignatures    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// LibSignaturesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LibSignaturesCallerSession struct {
	Contract *LibSignaturesCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// LibSignaturesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LibSignaturesTransactorSession struct {
	Contract     *LibSignaturesTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// LibSignaturesRaw is an auto generated low-level Go binding around an Ethereum contract.
type LibSignaturesRaw struct {
	Contract *LibSignatures // Generic contract binding to access the raw methods on
}

// LibSignaturesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LibSignaturesCallerRaw struct {
	Contract *LibSignaturesCaller // Generic read-only contract binding to access the raw methods on
}

// LibSignaturesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LibSignaturesTransactorRaw struct {
	Contract *LibSignaturesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLibSignatures creates a new instance of LibSignatures, bound to a specific deployed contract.
func NewLibSignatures(address common.Address, backend bind.ContractBackend) (*LibSignatures, error) {
	contract, err := bindLibSignatures(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &LibSignatures{LibSignaturesCaller: LibSignaturesCaller{contract: contract}, LibSignaturesTransactor: LibSignaturesTransactor{contract: contract}, LibSignaturesFilterer: LibSignaturesFilterer{contract: contract}}, nil
}

// NewLibSignaturesCaller creates a new read-only instance of LibSignatures, bound to a specific deployed contract.
func NewLibSignaturesCaller(address common.Address, caller bind.ContractCaller) (*LibSignaturesCaller, error) {
	contract, err := bindLibSignatures(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LibSignaturesCaller{contract: contract}, nil
}

// NewLibSignaturesTransactor creates a new write-only instance of LibSignatures, bound to a specific deployed contract.
func NewLibSignaturesTransactor(address common.Address, transactor bind.ContractTransactor) (*LibSignaturesTransactor, error) {
	contract, err := bindLibSignatures(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LibSignaturesTransactor{contract: contract}, nil
}

// NewLibSignaturesFilterer creates a new log filterer instance of LibSignatures, bound to a specific deployed contract.
func NewLibSignaturesFilterer(address common.Address, filterer bind.ContractFilterer) (*LibSignaturesFilterer, error) {
	contract, err := bindLibSignatures(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LibSignaturesFilterer{contract: contract}, nil
}

// bindLibSignatures binds a generic wrapper to an already deployed contract.
func bindLibSignatures(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(LibSignaturesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LibSignatures *LibSignaturesRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _LibSignatures.Contract.LibSignaturesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LibSignatures *LibSignaturesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LibSignatures.Contract.LibSignaturesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LibSignatures *LibSignaturesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LibSignatures.Contract.LibSignaturesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LibSignatures *LibSignaturesCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _LibSignatures.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LibSignatures *LibSignaturesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LibSignatures.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LibSignatures *LibSignaturesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LibSignatures.Contract.contract.Transact(opts, method, params...)
}

// Verify is a free data retrieval call binding the contract method 0x1a86b550.
//
// Solidity: function verify(addr address, message bytes32, signature bytes) constant returns(bool)
func (_LibSignatures *LibSignaturesCaller) Verify(opts *bind.CallOpts, addr common.Address, message [32]byte, signature []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _LibSignatures.contract.Call(opts, out, "verify", addr, message, signature)
	return *ret0, err
}

// Verify is a free data retrieval call binding the contract method 0x1a86b550.
//
// Solidity: function verify(addr address, message bytes32, signature bytes) constant returns(bool)
func (_LibSignatures *LibSignaturesSession) Verify(addr common.Address, message [32]byte, signature []byte) (bool, error) {
	return _LibSignatures.Contract.Verify(&_LibSignatures.CallOpts, addr, message, signature)
}

// Verify is a free data retrieval call binding the contract method 0x1a86b550.
//
// Solidity: function verify(addr address, message bytes32, signature bytes) constant returns(bool)
func (_LibSignatures *LibSignaturesCallerSession) Verify(addr common.Address, message [32]byte, signature []byte) (bool, error) {
	return _LibSignatures.Contract.Verify(&_LibSignatures.CallOpts, addr, message, signature)
}
