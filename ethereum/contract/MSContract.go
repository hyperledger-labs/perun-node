// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// ILibSignaturesABI is the input ABI used to generate the binding from.
const ILibSignaturesABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"message\",\"type\":\"bytes32\"},{\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"verify\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ILibSignaturesBin is the compiled bytecode used for deploying new contracts.
const ILibSignaturesBin = `0x`

// DeployILibSignatures deploys a new Ethereum contract, binding an instance of ILibSignatures to it.
func DeployILibSignatures(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ILibSignatures, error) {
	parsed, err := abi.JSON(strings.NewReader(ILibSignaturesABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ILibSignaturesBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ILibSignatures{ILibSignaturesCaller: ILibSignaturesCaller{contract: contract}, ILibSignaturesTransactor: ILibSignaturesTransactor{contract: contract}, ILibSignaturesFilterer: ILibSignaturesFilterer{contract: contract}}, nil
}

// ILibSignatures is an auto generated Go binding around an Ethereum contract.
type ILibSignatures struct {
	ILibSignaturesCaller     // Read-only binding to the contract
	ILibSignaturesTransactor // Write-only binding to the contract
	ILibSignaturesFilterer   // Log filterer for contract events
}

// ILibSignaturesCaller is an auto generated read-only Go binding around an Ethereum contract.
type ILibSignaturesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ILibSignaturesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ILibSignaturesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ILibSignaturesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ILibSignaturesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ILibSignaturesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ILibSignaturesSession struct {
	Contract     *ILibSignatures   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ILibSignaturesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ILibSignaturesCallerSession struct {
	Contract *ILibSignaturesCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// ILibSignaturesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ILibSignaturesTransactorSession struct {
	Contract     *ILibSignaturesTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// ILibSignaturesRaw is an auto generated low-level Go binding around an Ethereum contract.
type ILibSignaturesRaw struct {
	Contract *ILibSignatures // Generic contract binding to access the raw methods on
}

// ILibSignaturesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ILibSignaturesCallerRaw struct {
	Contract *ILibSignaturesCaller // Generic read-only contract binding to access the raw methods on
}

// ILibSignaturesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ILibSignaturesTransactorRaw struct {
	Contract *ILibSignaturesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewILibSignatures creates a new instance of ILibSignatures, bound to a specific deployed contract.
func NewILibSignatures(address common.Address, backend bind.ContractBackend) (*ILibSignatures, error) {
	contract, err := bindILibSignatures(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ILibSignatures{ILibSignaturesCaller: ILibSignaturesCaller{contract: contract}, ILibSignaturesTransactor: ILibSignaturesTransactor{contract: contract}, ILibSignaturesFilterer: ILibSignaturesFilterer{contract: contract}}, nil
}

// NewILibSignaturesCaller creates a new read-only instance of ILibSignatures, bound to a specific deployed contract.
func NewILibSignaturesCaller(address common.Address, caller bind.ContractCaller) (*ILibSignaturesCaller, error) {
	contract, err := bindILibSignatures(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ILibSignaturesCaller{contract: contract}, nil
}

// NewILibSignaturesTransactor creates a new write-only instance of ILibSignatures, bound to a specific deployed contract.
func NewILibSignaturesTransactor(address common.Address, transactor bind.ContractTransactor) (*ILibSignaturesTransactor, error) {
	contract, err := bindILibSignatures(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ILibSignaturesTransactor{contract: contract}, nil
}

// NewILibSignaturesFilterer creates a new log filterer instance of ILibSignatures, bound to a specific deployed contract.
func NewILibSignaturesFilterer(address common.Address, filterer bind.ContractFilterer) (*ILibSignaturesFilterer, error) {
	contract, err := bindILibSignatures(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ILibSignaturesFilterer{contract: contract}, nil
}

// bindILibSignatures binds a generic wrapper to an already deployed contract.
func bindILibSignatures(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ILibSignaturesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ILibSignatures *ILibSignaturesRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ILibSignatures.Contract.ILibSignaturesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ILibSignatures *ILibSignaturesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ILibSignatures.Contract.ILibSignaturesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ILibSignatures *ILibSignaturesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ILibSignatures.Contract.ILibSignaturesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ILibSignatures *ILibSignaturesCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ILibSignatures.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ILibSignatures *ILibSignaturesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ILibSignatures.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ILibSignatures *ILibSignaturesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ILibSignatures.Contract.contract.Transact(opts, method, params...)
}

// Verify is a free data retrieval call binding the contract method 0x1a86b550.
//
// Solidity: function verify(addr address, message bytes32, signature bytes) constant returns(bool)
func (_ILibSignatures *ILibSignaturesCaller) Verify(opts *bind.CallOpts, addr common.Address, message [32]byte, signature []byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ILibSignatures.contract.Call(opts, out, "verify", addr, message, signature)
	return *ret0, err
}

// Verify is a free data retrieval call binding the contract method 0x1a86b550.
//
// Solidity: function verify(addr address, message bytes32, signature bytes) constant returns(bool)
func (_ILibSignatures *ILibSignaturesSession) Verify(addr common.Address, message [32]byte, signature []byte) (bool, error) {
	return _ILibSignatures.Contract.Verify(&_ILibSignatures.CallOpts, addr, message, signature)
}

// Verify is a free data retrieval call binding the contract method 0x1a86b550.
//
// Solidity: function verify(addr address, message bytes32, signature bytes) constant returns(bool)
func (_ILibSignatures *ILibSignaturesCallerSession) Verify(addr common.Address, message [32]byte, signature []byte) (bool, error) {
	return _ILibSignatures.Contract.Verify(&_ILibSignatures.CallOpts, addr, message, signature)
}

// MSContractABI is the input ABI used to generate the binding from.
const MSContractABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"status\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"close\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_vpc\",\"type\":\"address\"},{\"name\":\"_sid\",\"type\":\"uint256\"},{\"name\":\"_blockedA\",\"type\":\"uint256\"},{\"name\":\"_blockedB\",\"type\":\"uint256\"},{\"name\":\"_version\",\"type\":\"uint256\"},{\"name\":\"sigA\",\"type\":\"bytes\"},{\"name\":\"sigB\",\"type\":\"bytes\"}],\"name\":\"stateRegister\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"refund\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"finalizeClose\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"confirm\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"timeout\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"bob\",\"outputs\":[{\"name\":\"id\",\"type\":\"address\"},{\"name\":\"cash\",\"type\":\"uint256\"},{\"name\":\"waitForInput\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"finalizeRegister\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"c\",\"outputs\":[{\"name\":\"active\",\"type\":\"bool\"},{\"name\":\"vpc\",\"type\":\"address\"},{\"name\":\"sid\",\"type\":\"uint256\"},{\"name\":\"blockedA\",\"type\":\"uint256\"},{\"name\":\"blockedB\",\"type\":\"uint256\"},{\"name\":\"version\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_alice\",\"type\":\"address\"},{\"name\":\"_bob\",\"type\":\"address\"}],\"name\":\"execute\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"alice\",\"outputs\":[{\"name\":\"id\",\"type\":\"address\"},{\"name\":\"cash\",\"type\":\"uint256\"},{\"name\":\"waitForInput\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_libSig\",\"type\":\"address\"},{\"name\":\"_addressAlice\",\"type\":\"address\"},{\"name\":\"_addressBob\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"addressAlice\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"addressBob\",\"type\":\"address\"}],\"name\":\"EventInitializing\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"cashAlice\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"cashBob\",\"type\":\"uint256\"}],\"name\":\"EventInitialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EventRefunded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EventStateRegistering\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"blockedAlice\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"blockedBob\",\"type\":\"uint256\"}],\"name\":\"EventStateRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EventClosing\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EventClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EventNotClosed\",\"type\":\"event\"}]"

// MSContractBin is the compiled bytecode used for deploying new contracts.
const MSContractBin = `0x608060405234801561001057600080fd5b5060405160608061130083398101604090815281516020830151919092015160008054600160a060020a03808516600160a060020a03199283161783556003805482861693169290921790915561177042016006556002805460ff19908116600190811790925560058054821683179055600c80549388166101000261010060a860020a03199094169390931780845516908302179055506007805460ff1916905560408051600160a060020a0380851682528316602082015281517f6cb11985dbf588b27bf25de9727dc78faff0c068234f103b54525ec5e4ddebc5929181900390910190a15050506111f7806101096000396000f3006080604052600436106100b95763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663200d2ed281146100be57806343d726d6146100f757806346cc2f8a1461010e578063590e1ae3146101c157806367516ea1146101d65780637022b58e146101eb57806370dea79a146101f3578063c09cec771461021a578063c399584614610259578063c3da42b81461026e578063d80aea15146102c0578063fb47e3a2146102e7575b600080fd5b3480156100ca57600080fd5b506100d36102fc565b604051808260058111156100e357fe5b60ff16815260200191505060405180910390f35b34801561010357600080fd5b5061010c610305565b005b34801561011a57600080fd5b50604080516020600460a43581810135601f810184900484028501840190955284845261010c948235600160a060020a03169460248035956044359560643595608435953695929460c494920191819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506104ef9650505050505050565b3480156101cd57600080fd5b5061010c610ace565b3480156101e257600080fd5b5061010c610bfc565b61010c610d2e565b3480156101ff57600080fd5b50610208610e59565b60408051918252519081900360200190f35b34801561022657600080fd5b5061022f610e5f565b60408051600160a060020a0390941684526020840192909252151582820152519081900360600190f35b34801561026557600080fd5b5061010c610e7a565b34801561027a57600080fd5b50610283610f4a565b604080519615158752600160a060020a039095166020870152858501939093526060850191909152608084015260a0830152519081900360c00190f35b3480156102cc57600080fd5b5061010c600160a060020a0360043581169060243516610f73565b3480156102f357600080fd5b5061022f6111b0565b600c5460ff1681565b600054600160a060020a03163314806103285750600354600160a060020a031633145b151561033357600080fd5b6001600c5460ff16600581111561034657fe5b14156103a657600c805460ff1990811660041790915561465042016006556002805482166001908117909155600580549092161790556040517f1cf3c51ed35ad77fe26da47a4b40e561be39d952d0a8e2d3c5d6f688e8cd4a7f90600090a15b6004600c5460ff1660058111156103b957fe5b146103c3576104ed565b60025460ff1680156103df5750600054600160a060020a031633145b156103ef576002805460ff191690555b60055460ff16801561040b5750600354600160a060020a031633145b1561041b576005805460ff191690555b60025460ff16158015610431575060055460ff16155b156104ed5760008054600154604051600160a060020a039092169281156108fc029290818181858888f193505050501561046b5760006001555b600354600454604051600160a060020a039092169181156108fc0291906000818181858888f19350505050156104a15760006004555b6001541580156104b15750600454155b156104ed576040517f84b734c38d9368aeb5efcfdc2b8dd5da3027b2f694d640e017ba23beafe6c06c90600090a1600054600160a060020a0316ff5b565b600080548190600160a060020a03163314806105155750600354600160a060020a031633145b151561052057600080fd5b600154871180159061053457506004548611155b151561053f57600080fd5b604080516c01000000000000000000000000600160a060020a038c1602602080830191909152603482018b9052605482018a90526074820189905260948083018990528351808403909101815260b490920192839052815191929182918401908083835b602083106105c25780518252601f1990920191602091820191016105a3565b51815160209384036101000a6000190180199092169116179052604080519290940182900382207f19457468657265756d205369676e6564204d6573736167653a0a33320000000083830152603c80840182905285518085039091018152605c9093019485905282519098509195509293508392850191508083835b6020831061065d5780518252601f19909201916020918201910161063e565b6001836020036101000a03801982511681845116808217855250505050505090500191505060405180910390209050600c60019054906101000a9004600160a060020a0316600160a060020a0316631a86b5506000800160009054906101000a9004600160a060020a031683876040518463ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018084600160a060020a0316600160a060020a03168152602001836000191660001916815260200180602001828103825283818151815260200191508051906020019080838360005b83811015610758578181015183820152602001610740565b50505050905090810190601f1680156107855780820380516001836020036101000a031916815260200191505b50945050505050602060405180830381600087803b1580156107a657600080fd5b505af11580156107ba573d6000803e3d6000fd5b505050506040513d60208110156107d057600080fd5b505180156108eb5750600c546003546040517f1a86b550000000000000000000000000000000000000000000000000000000008152600160a060020a03918216600482018181526024830186905260606044840190815288516064850152885161010090960490941694631a86b55094929387938a9392909160840190602085019080838360005b83811015610870578181015183820152602001610858565b50505050905090810190601f16801561089d5780820380516001836020036101000a031916815260200191505b50945050505050602060405180830381600087803b1580156108be57600080fd5b505af11580156108d2573d6000803e3d6000fd5b505050506040513d60208110156108e857600080fd5b50515b15156108f657600080fd5b6001600c5460ff16600581111561090957fe5b148061092557506004600c5460ff16600581111561092357fe5b145b1561098557600c8054600260ff19918216811790925581548116600190811790925560058054909116909117905542611770016006556040517f82ce8e9a702f5f21ca7b1a4bbc6fd40407c770d7ae87544871fe505a728c7e7190600090a15b6002600c5460ff16600581111561099857fe5b146109a257610ac3565b600054600160a060020a03163314156109c0576002805460ff191690555b600354600160a060020a03163314156109de576005805460ff191690555b600b54851115610a335760078054600160ff199091161774ffffffffffffffffffffffffffffffffffffffff001916610100600160a060020a038c160217905560088890556009879055600a869055600b8590555b60025460ff16158015610a49575060055460ff16155b15610ac357600c805460ff199081166003179091556002805482169055600580549091169055600954600180548290039055600a5460048054829003905560408051928352602083019190915280517ff025b0293fd70914bfa444833429b6c197b8d1aefa338e5d4e6fb3ccfd1aa2519281900390910190a15b505050505050505050565b600054600160a060020a0316331480610af15750600354600160a060020a031633145b1515610afc57600080fd5b6000600c5460ff166005811115610b0f57fe5b148015610b1d575060065442115b1515610b2857600080fd5b60025460ff168015610b3c57506001546000105b15610b765760008054600154604051600160a060020a039092169281156108fc029290818181858888f193505050501515610b7657600080fd5b60055460ff168015610b8a57506004546000105b15610bc557600354600454604051600160a060020a039092169181156108fc0291906000818181858888f193505050501515610bc557600080fd5b6040517fefbc23b4c571c79ee832603f52f6fafc9321a042b7c57a9bd82f809741db384990600090a1600054600160a060020a0316ff5b600054600160a060020a0316331480610c1f5750600354600160a060020a031633145b1515610c2a57600080fd5b6004600c5460ff166005811115610c3d57fe5b14610c70576040517fcd08a02eb4f3c86c4b3352cbe9447051bfc1af76dd9388d388d23c5624a4524b90600090a16104ed565b6006544211156104ed5760008054600154604051600160a060020a039092169281156108fc029290818181858888f193505050501561046b576000600155600354600454604051600160a060020a039092169181156108fc0291906000818181858888f19350505050156104a15760006004556001541580156104b1575060045415156104ed576040517f84b734c38d9368aeb5efcfdc2b8dd5da3027b2f694d640e017ba23beafe6c06c90600090a1600054600160a060020a0316ff5b600054600160a060020a0316331480610d515750600354600160a060020a031633145b1515610d5c57600080fd5b6000600c5460ff166005811115610d6f57fe5b148015610d7d575060065442105b1515610d8857600080fd5b60025460ff168015610da45750600054600160a060020a031633145b15610db857346001556002805460ff191690555b60055460ff168015610dd45750600354600160a060020a031633145b15610de857346004556005805460ff191690555b60025460ff16158015610dfe575060055460ff16155b156104ed57600c805460ff1916600190811790915560006006555460045460408051928352602083019190915280517f66923fdac57bb7f27112472d419ef406c021433b6cf64d88b2f78c5280fe0b629281900390910190a1565b60065481565b600354600454600554600160a060020a039092169160ff1683565b600054600160a060020a0316331480610e9d5750600354600160a060020a031633145b1515610ea857600080fd5b6002600c5460ff166005811115610ebb57fe5b148015610ec9575060065442115b1515610ed457600080fd5b600c805460ff199081166003179091556002805482169055600580549091169055600954600180548290039055600a5460048054829003905560408051928352602083019190915280517ff025b0293fd70914bfa444833429b6c197b8d1aefa338e5d4e6fb3ccfd1aa2519281900390910190a1565b600754600854600954600a54600b5460ff8516946101009004600160a060020a03169392919086565b60008054819081908190600160a060020a0316331480610f9d5750600354600160a060020a031633145b1515610fa857600080fd5b6003600c5460ff166005811115610fbb57fe5b14610fc557600080fd5b50600754600854604080517f5e9d2afe000000000000000000000000000000000000000000000000000000008152600160a060020a038981166004830152888116602483015260448201939093529051610100909304909116918291635e9d2afe9160648083019260609291908290030181600087803b15801561104857600080fd5b505af115801561105c573d6000803e3d6000fd5b505050506040513d606081101561107257600080fd5b50805160208201516040909201519095509093509150831515611094576111a8565b600a546009540182840114156110c75760018054840190556009805484900390556004805483019055600a805483900390555b60008054600154604051600160a060020a039092169281156108fc029290818181858888f19350505050156110fc5760006001555b600354600454604051600160a060020a039092169181156108fc0291906000818181858888f19350505050156111325760006004555b6001541580156111425750600454155b1561117e576040517f84b734c38d9368aeb5efcfdc2b8dd5da3027b2f694d640e017ba23beafe6c06c90600090a1600054600160a060020a0316ff5b6040517fcd08a02eb4f3c86c4b3352cbe9447051bfc1af76dd9388d388d23c5624a4524b90600090a15b505050505050565b600054600154600254600160a060020a039092169160ff16835600a165627a7a7230582023d3f4b8d230fd9e18103894f0985040f77bafe2e338f4b1f96af780fccdcb8a0029`

// DeployMSContract deploys a new Ethereum contract, binding an instance of MSContract to it.
func DeployMSContract(auth *bind.TransactOpts, backend bind.ContractBackend, _libSig common.Address, _addressAlice common.Address, _addressBob common.Address) (common.Address, *types.Transaction, *MSContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MSContractABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(MSContractBin), backend, _libSig, _addressAlice, _addressBob)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MSContract{MSContractCaller: MSContractCaller{contract: contract}, MSContractTransactor: MSContractTransactor{contract: contract}, MSContractFilterer: MSContractFilterer{contract: contract}}, nil
}

// MSContract is an auto generated Go binding around an Ethereum contract.
type MSContract struct {
	MSContractCaller     // Read-only binding to the contract
	MSContractTransactor // Write-only binding to the contract
	MSContractFilterer   // Log filterer for contract events
}

// MSContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type MSContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MSContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MSContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MSContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MSContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MSContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MSContractSession struct {
	Contract     *MSContract       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MSContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MSContractCallerSession struct {
	Contract *MSContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// MSContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MSContractTransactorSession struct {
	Contract     *MSContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// MSContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type MSContractRaw struct {
	Contract *MSContract // Generic contract binding to access the raw methods on
}

// MSContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MSContractCallerRaw struct {
	Contract *MSContractCaller // Generic read-only contract binding to access the raw methods on
}

// MSContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MSContractTransactorRaw struct {
	Contract *MSContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMSContract creates a new instance of MSContract, bound to a specific deployed contract.
func NewMSContract(address common.Address, backend bind.ContractBackend) (*MSContract, error) {
	contract, err := bindMSContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MSContract{MSContractCaller: MSContractCaller{contract: contract}, MSContractTransactor: MSContractTransactor{contract: contract}, MSContractFilterer: MSContractFilterer{contract: contract}}, nil
}

// NewMSContractCaller creates a new read-only instance of MSContract, bound to a specific deployed contract.
func NewMSContractCaller(address common.Address, caller bind.ContractCaller) (*MSContractCaller, error) {
	contract, err := bindMSContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MSContractCaller{contract: contract}, nil
}

// NewMSContractTransactor creates a new write-only instance of MSContract, bound to a specific deployed contract.
func NewMSContractTransactor(address common.Address, transactor bind.ContractTransactor) (*MSContractTransactor, error) {
	contract, err := bindMSContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MSContractTransactor{contract: contract}, nil
}

// NewMSContractFilterer creates a new log filterer instance of MSContract, bound to a specific deployed contract.
func NewMSContractFilterer(address common.Address, filterer bind.ContractFilterer) (*MSContractFilterer, error) {
	contract, err := bindMSContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MSContractFilterer{contract: contract}, nil
}

// bindMSContract binds a generic wrapper to an already deployed contract.
func bindMSContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MSContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MSContract *MSContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MSContract.Contract.MSContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MSContract *MSContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MSContract.Contract.MSContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MSContract *MSContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MSContract.Contract.MSContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MSContract *MSContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MSContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MSContract *MSContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MSContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MSContract *MSContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MSContract.Contract.contract.Transact(opts, method, params...)
}

// Alice is a free data retrieval call binding the contract method 0xfb47e3a2.
//
// Solidity: function alice() constant returns(id address, cash uint256, waitForInput bool)
func (_MSContract *MSContractCaller) Alice(opts *bind.CallOpts) (struct {
	Id           common.Address
	Cash         *big.Int
	WaitForInput bool
}, error) {
	ret := new(struct {
		Id           common.Address
		Cash         *big.Int
		WaitForInput bool
	})
	out := ret
	err := _MSContract.contract.Call(opts, out, "alice")
	return *ret, err
}

// Alice is a free data retrieval call binding the contract method 0xfb47e3a2.
//
// Solidity: function alice() constant returns(id address, cash uint256, waitForInput bool)
func (_MSContract *MSContractSession) Alice() (struct {
	Id           common.Address
	Cash         *big.Int
	WaitForInput bool
}, error) {
	return _MSContract.Contract.Alice(&_MSContract.CallOpts)
}

// Alice is a free data retrieval call binding the contract method 0xfb47e3a2.
//
// Solidity: function alice() constant returns(id address, cash uint256, waitForInput bool)
func (_MSContract *MSContractCallerSession) Alice() (struct {
	Id           common.Address
	Cash         *big.Int
	WaitForInput bool
}, error) {
	return _MSContract.Contract.Alice(&_MSContract.CallOpts)
}

// Bob is a free data retrieval call binding the contract method 0xc09cec77.
//
// Solidity: function bob() constant returns(id address, cash uint256, waitForInput bool)
func (_MSContract *MSContractCaller) Bob(opts *bind.CallOpts) (struct {
	Id           common.Address
	Cash         *big.Int
	WaitForInput bool
}, error) {
	ret := new(struct {
		Id           common.Address
		Cash         *big.Int
		WaitForInput bool
	})
	out := ret
	err := _MSContract.contract.Call(opts, out, "bob")
	return *ret, err
}

// Bob is a free data retrieval call binding the contract method 0xc09cec77.
//
// Solidity: function bob() constant returns(id address, cash uint256, waitForInput bool)
func (_MSContract *MSContractSession) Bob() (struct {
	Id           common.Address
	Cash         *big.Int
	WaitForInput bool
}, error) {
	return _MSContract.Contract.Bob(&_MSContract.CallOpts)
}

// Bob is a free data retrieval call binding the contract method 0xc09cec77.
//
// Solidity: function bob() constant returns(id address, cash uint256, waitForInput bool)
func (_MSContract *MSContractCallerSession) Bob() (struct {
	Id           common.Address
	Cash         *big.Int
	WaitForInput bool
}, error) {
	return _MSContract.Contract.Bob(&_MSContract.CallOpts)
}

// C is a free data retrieval call binding the contract method 0xc3da42b8.
//
// Solidity: function c() constant returns(active bool, vpc address, sid uint256, blockedA uint256, blockedB uint256, version uint256)
func (_MSContract *MSContractCaller) C(opts *bind.CallOpts) (struct {
	Active   bool
	Vpc      common.Address
	Sid      *big.Int
	BlockedA *big.Int
	BlockedB *big.Int
	Version  *big.Int
}, error) {
	ret := new(struct {
		Active   bool
		Vpc      common.Address
		Sid      *big.Int
		BlockedA *big.Int
		BlockedB *big.Int
		Version  *big.Int
	})
	out := ret
	err := _MSContract.contract.Call(opts, out, "c")
	return *ret, err
}

// C is a free data retrieval call binding the contract method 0xc3da42b8.
//
// Solidity: function c() constant returns(active bool, vpc address, sid uint256, blockedA uint256, blockedB uint256, version uint256)
func (_MSContract *MSContractSession) C() (struct {
	Active   bool
	Vpc      common.Address
	Sid      *big.Int
	BlockedA *big.Int
	BlockedB *big.Int
	Version  *big.Int
}, error) {
	return _MSContract.Contract.C(&_MSContract.CallOpts)
}

// C is a free data retrieval call binding the contract method 0xc3da42b8.
//
// Solidity: function c() constant returns(active bool, vpc address, sid uint256, blockedA uint256, blockedB uint256, version uint256)
func (_MSContract *MSContractCallerSession) C() (struct {
	Active   bool
	Vpc      common.Address
	Sid      *big.Int
	BlockedA *big.Int
	BlockedB *big.Int
	Version  *big.Int
}, error) {
	return _MSContract.Contract.C(&_MSContract.CallOpts)
}

// Status is a free data retrieval call binding the contract method 0x200d2ed2.
//
// Solidity: function status() constant returns(uint8)
func (_MSContract *MSContractCaller) Status(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _MSContract.contract.Call(opts, out, "status")
	return *ret0, err
}

// Status is a free data retrieval call binding the contract method 0x200d2ed2.
//
// Solidity: function status() constant returns(uint8)
func (_MSContract *MSContractSession) Status() (uint8, error) {
	return _MSContract.Contract.Status(&_MSContract.CallOpts)
}

// Status is a free data retrieval call binding the contract method 0x200d2ed2.
//
// Solidity: function status() constant returns(uint8)
func (_MSContract *MSContractCallerSession) Status() (uint8, error) {
	return _MSContract.Contract.Status(&_MSContract.CallOpts)
}

// Timeout is a free data retrieval call binding the contract method 0x70dea79a.
//
// Solidity: function timeout() constant returns(uint256)
func (_MSContract *MSContractCaller) Timeout(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MSContract.contract.Call(opts, out, "timeout")
	return *ret0, err
}

// Timeout is a free data retrieval call binding the contract method 0x70dea79a.
//
// Solidity: function timeout() constant returns(uint256)
func (_MSContract *MSContractSession) Timeout() (*big.Int, error) {
	return _MSContract.Contract.Timeout(&_MSContract.CallOpts)
}

// Timeout is a free data retrieval call binding the contract method 0x70dea79a.
//
// Solidity: function timeout() constant returns(uint256)
func (_MSContract *MSContractCallerSession) Timeout() (*big.Int, error) {
	return _MSContract.Contract.Timeout(&_MSContract.CallOpts)
}

// Close is a paid mutator transaction binding the contract method 0x43d726d6.
//
// Solidity: function close() returns()
func (_MSContract *MSContractTransactor) Close(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MSContract.contract.Transact(opts, "close")
}

// Close is a paid mutator transaction binding the contract method 0x43d726d6.
//
// Solidity: function close() returns()
func (_MSContract *MSContractSession) Close() (*types.Transaction, error) {
	return _MSContract.Contract.Close(&_MSContract.TransactOpts)
}

// Close is a paid mutator transaction binding the contract method 0x43d726d6.
//
// Solidity: function close() returns()
func (_MSContract *MSContractTransactorSession) Close() (*types.Transaction, error) {
	return _MSContract.Contract.Close(&_MSContract.TransactOpts)
}

// Confirm is a paid mutator transaction binding the contract method 0x7022b58e.
//
// Solidity: function confirm() returns()
func (_MSContract *MSContractTransactor) Confirm(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MSContract.contract.Transact(opts, "confirm")
}

// Confirm is a paid mutator transaction binding the contract method 0x7022b58e.
//
// Solidity: function confirm() returns()
func (_MSContract *MSContractSession) Confirm() (*types.Transaction, error) {
	return _MSContract.Contract.Confirm(&_MSContract.TransactOpts)
}

// Confirm is a paid mutator transaction binding the contract method 0x7022b58e.
//
// Solidity: function confirm() returns()
func (_MSContract *MSContractTransactorSession) Confirm() (*types.Transaction, error) {
	return _MSContract.Contract.Confirm(&_MSContract.TransactOpts)
}

// Execute is a paid mutator transaction binding the contract method 0xd80aea15.
//
// Solidity: function execute(_alice address, _bob address) returns()
func (_MSContract *MSContractTransactor) Execute(opts *bind.TransactOpts, _alice common.Address, _bob common.Address) (*types.Transaction, error) {
	return _MSContract.contract.Transact(opts, "execute", _alice, _bob)
}

// Execute is a paid mutator transaction binding the contract method 0xd80aea15.
//
// Solidity: function execute(_alice address, _bob address) returns()
func (_MSContract *MSContractSession) Execute(_alice common.Address, _bob common.Address) (*types.Transaction, error) {
	return _MSContract.Contract.Execute(&_MSContract.TransactOpts, _alice, _bob)
}

// Execute is a paid mutator transaction binding the contract method 0xd80aea15.
//
// Solidity: function execute(_alice address, _bob address) returns()
func (_MSContract *MSContractTransactorSession) Execute(_alice common.Address, _bob common.Address) (*types.Transaction, error) {
	return _MSContract.Contract.Execute(&_MSContract.TransactOpts, _alice, _bob)
}

// FinalizeClose is a paid mutator transaction binding the contract method 0x67516ea1.
//
// Solidity: function finalizeClose() returns()
func (_MSContract *MSContractTransactor) FinalizeClose(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MSContract.contract.Transact(opts, "finalizeClose")
}

// FinalizeClose is a paid mutator transaction binding the contract method 0x67516ea1.
//
// Solidity: function finalizeClose() returns()
func (_MSContract *MSContractSession) FinalizeClose() (*types.Transaction, error) {
	return _MSContract.Contract.FinalizeClose(&_MSContract.TransactOpts)
}

// FinalizeClose is a paid mutator transaction binding the contract method 0x67516ea1.
//
// Solidity: function finalizeClose() returns()
func (_MSContract *MSContractTransactorSession) FinalizeClose() (*types.Transaction, error) {
	return _MSContract.Contract.FinalizeClose(&_MSContract.TransactOpts)
}

// FinalizeRegister is a paid mutator transaction binding the contract method 0xc3995846.
//
// Solidity: function finalizeRegister() returns()
func (_MSContract *MSContractTransactor) FinalizeRegister(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MSContract.contract.Transact(opts, "finalizeRegister")
}

// FinalizeRegister is a paid mutator transaction binding the contract method 0xc3995846.
//
// Solidity: function finalizeRegister() returns()
func (_MSContract *MSContractSession) FinalizeRegister() (*types.Transaction, error) {
	return _MSContract.Contract.FinalizeRegister(&_MSContract.TransactOpts)
}

// FinalizeRegister is a paid mutator transaction binding the contract method 0xc3995846.
//
// Solidity: function finalizeRegister() returns()
func (_MSContract *MSContractTransactorSession) FinalizeRegister() (*types.Transaction, error) {
	return _MSContract.Contract.FinalizeRegister(&_MSContract.TransactOpts)
}

// Refund is a paid mutator transaction binding the contract method 0x590e1ae3.
//
// Solidity: function refund() returns()
func (_MSContract *MSContractTransactor) Refund(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MSContract.contract.Transact(opts, "refund")
}

// Refund is a paid mutator transaction binding the contract method 0x590e1ae3.
//
// Solidity: function refund() returns()
func (_MSContract *MSContractSession) Refund() (*types.Transaction, error) {
	return _MSContract.Contract.Refund(&_MSContract.TransactOpts)
}

// Refund is a paid mutator transaction binding the contract method 0x590e1ae3.
//
// Solidity: function refund() returns()
func (_MSContract *MSContractTransactorSession) Refund() (*types.Transaction, error) {
	return _MSContract.Contract.Refund(&_MSContract.TransactOpts)
}

// StateRegister is a paid mutator transaction binding the contract method 0x46cc2f8a.
//
// Solidity: function stateRegister(_vpc address, _sid uint256, _blockedA uint256, _blockedB uint256, _version uint256, sigA bytes, sigB bytes) returns()
func (_MSContract *MSContractTransactor) StateRegister(opts *bind.TransactOpts, _vpc common.Address, _sid *big.Int, _blockedA *big.Int, _blockedB *big.Int, _version *big.Int, sigA []byte, sigB []byte) (*types.Transaction, error) {
	return _MSContract.contract.Transact(opts, "stateRegister", _vpc, _sid, _blockedA, _blockedB, _version, sigA, sigB)
}

// StateRegister is a paid mutator transaction binding the contract method 0x46cc2f8a.
//
// Solidity: function stateRegister(_vpc address, _sid uint256, _blockedA uint256, _blockedB uint256, _version uint256, sigA bytes, sigB bytes) returns()
func (_MSContract *MSContractSession) StateRegister(_vpc common.Address, _sid *big.Int, _blockedA *big.Int, _blockedB *big.Int, _version *big.Int, sigA []byte, sigB []byte) (*types.Transaction, error) {
	return _MSContract.Contract.StateRegister(&_MSContract.TransactOpts, _vpc, _sid, _blockedA, _blockedB, _version, sigA, sigB)
}

// StateRegister is a paid mutator transaction binding the contract method 0x46cc2f8a.
//
// Solidity: function stateRegister(_vpc address, _sid uint256, _blockedA uint256, _blockedB uint256, _version uint256, sigA bytes, sigB bytes) returns()
func (_MSContract *MSContractTransactorSession) StateRegister(_vpc common.Address, _sid *big.Int, _blockedA *big.Int, _blockedB *big.Int, _version *big.Int, sigA []byte, sigB []byte) (*types.Transaction, error) {
	return _MSContract.Contract.StateRegister(&_MSContract.TransactOpts, _vpc, _sid, _blockedA, _blockedB, _version, sigA, sigB)
}

// MSContractEventClosedIterator is returned from FilterEventClosed and is used to iterate over the raw logs and unpacked data for EventClosed events raised by the MSContract contract.
type MSContractEventClosedIterator struct {
	Event *MSContractEventClosed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MSContractEventClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MSContractEventClosed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MSContractEventClosed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MSContractEventClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MSContractEventClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MSContractEventClosed represents a EventClosed event raised by the MSContract contract.
type MSContractEventClosed struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEventClosed is a free log retrieval operation binding the contract event 0x84b734c38d9368aeb5efcfdc2b8dd5da3027b2f694d640e017ba23beafe6c06c.
//
// Solidity: e EventClosed()
func (_MSContract *MSContractFilterer) FilterEventClosed(opts *bind.FilterOpts) (*MSContractEventClosedIterator, error) {

	logs, sub, err := _MSContract.contract.FilterLogs(opts, "EventClosed")
	if err != nil {
		return nil, err
	}
	return &MSContractEventClosedIterator{contract: _MSContract.contract, event: "EventClosed", logs: logs, sub: sub}, nil
}

// WatchEventClosed is a free log subscription operation binding the contract event 0x84b734c38d9368aeb5efcfdc2b8dd5da3027b2f694d640e017ba23beafe6c06c.
//
// Solidity: e EventClosed()
func (_MSContract *MSContractFilterer) WatchEventClosed(opts *bind.WatchOpts, sink chan<- *MSContractEventClosed) (event.Subscription, error) {

	logs, sub, err := _MSContract.contract.WatchLogs(opts, "EventClosed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MSContractEventClosed)
				if err := _MSContract.contract.UnpackLog(event, "EventClosed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// MSContractEventClosingIterator is returned from FilterEventClosing and is used to iterate over the raw logs and unpacked data for EventClosing events raised by the MSContract contract.
type MSContractEventClosingIterator struct {
	Event *MSContractEventClosing // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MSContractEventClosingIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MSContractEventClosing)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MSContractEventClosing)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MSContractEventClosingIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MSContractEventClosingIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MSContractEventClosing represents a EventClosing event raised by the MSContract contract.
type MSContractEventClosing struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEventClosing is a free log retrieval operation binding the contract event 0x1cf3c51ed35ad77fe26da47a4b40e561be39d952d0a8e2d3c5d6f688e8cd4a7f.
//
// Solidity: e EventClosing()
func (_MSContract *MSContractFilterer) FilterEventClosing(opts *bind.FilterOpts) (*MSContractEventClosingIterator, error) {

	logs, sub, err := _MSContract.contract.FilterLogs(opts, "EventClosing")
	if err != nil {
		return nil, err
	}
	return &MSContractEventClosingIterator{contract: _MSContract.contract, event: "EventClosing", logs: logs, sub: sub}, nil
}

// WatchEventClosing is a free log subscription operation binding the contract event 0x1cf3c51ed35ad77fe26da47a4b40e561be39d952d0a8e2d3c5d6f688e8cd4a7f.
//
// Solidity: e EventClosing()
func (_MSContract *MSContractFilterer) WatchEventClosing(opts *bind.WatchOpts, sink chan<- *MSContractEventClosing) (event.Subscription, error) {

	logs, sub, err := _MSContract.contract.WatchLogs(opts, "EventClosing")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MSContractEventClosing)
				if err := _MSContract.contract.UnpackLog(event, "EventClosing", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// MSContractEventInitializedIterator is returned from FilterEventInitialized and is used to iterate over the raw logs and unpacked data for EventInitialized events raised by the MSContract contract.
type MSContractEventInitializedIterator struct {
	Event *MSContractEventInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MSContractEventInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MSContractEventInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MSContractEventInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MSContractEventInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MSContractEventInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MSContractEventInitialized represents a EventInitialized event raised by the MSContract contract.
type MSContractEventInitialized struct {
	CashAlice *big.Int
	CashBob   *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEventInitialized is a free log retrieval operation binding the contract event 0x66923fdac57bb7f27112472d419ef406c021433b6cf64d88b2f78c5280fe0b62.
//
// Solidity: e EventInitialized(cashAlice uint256, cashBob uint256)
func (_MSContract *MSContractFilterer) FilterEventInitialized(opts *bind.FilterOpts) (*MSContractEventInitializedIterator, error) {

	logs, sub, err := _MSContract.contract.FilterLogs(opts, "EventInitialized")
	if err != nil {
		return nil, err
	}
	return &MSContractEventInitializedIterator{contract: _MSContract.contract, event: "EventInitialized", logs: logs, sub: sub}, nil
}

// WatchEventInitialized is a free log subscription operation binding the contract event 0x66923fdac57bb7f27112472d419ef406c021433b6cf64d88b2f78c5280fe0b62.
//
// Solidity: e EventInitialized(cashAlice uint256, cashBob uint256)
func (_MSContract *MSContractFilterer) WatchEventInitialized(opts *bind.WatchOpts, sink chan<- *MSContractEventInitialized) (event.Subscription, error) {

	logs, sub, err := _MSContract.contract.WatchLogs(opts, "EventInitialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MSContractEventInitialized)
				if err := _MSContract.contract.UnpackLog(event, "EventInitialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// MSContractEventInitializingIterator is returned from FilterEventInitializing and is used to iterate over the raw logs and unpacked data for EventInitializing events raised by the MSContract contract.
type MSContractEventInitializingIterator struct {
	Event *MSContractEventInitializing // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MSContractEventInitializingIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MSContractEventInitializing)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MSContractEventInitializing)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MSContractEventInitializingIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MSContractEventInitializingIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MSContractEventInitializing represents a EventInitializing event raised by the MSContract contract.
type MSContractEventInitializing struct {
	AddressAlice common.Address
	AddressBob   common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterEventInitializing is a free log retrieval operation binding the contract event 0x6cb11985dbf588b27bf25de9727dc78faff0c068234f103b54525ec5e4ddebc5.
//
// Solidity: e EventInitializing(addressAlice address, addressBob address)
func (_MSContract *MSContractFilterer) FilterEventInitializing(opts *bind.FilterOpts) (*MSContractEventInitializingIterator, error) {

	logs, sub, err := _MSContract.contract.FilterLogs(opts, "EventInitializing")
	if err != nil {
		return nil, err
	}
	return &MSContractEventInitializingIterator{contract: _MSContract.contract, event: "EventInitializing", logs: logs, sub: sub}, nil
}

// WatchEventInitializing is a free log subscription operation binding the contract event 0x6cb11985dbf588b27bf25de9727dc78faff0c068234f103b54525ec5e4ddebc5.
//
// Solidity: e EventInitializing(addressAlice address, addressBob address)
func (_MSContract *MSContractFilterer) WatchEventInitializing(opts *bind.WatchOpts, sink chan<- *MSContractEventInitializing) (event.Subscription, error) {

	logs, sub, err := _MSContract.contract.WatchLogs(opts, "EventInitializing")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MSContractEventInitializing)
				if err := _MSContract.contract.UnpackLog(event, "EventInitializing", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// MSContractEventNotClosedIterator is returned from FilterEventNotClosed and is used to iterate over the raw logs and unpacked data for EventNotClosed events raised by the MSContract contract.
type MSContractEventNotClosedIterator struct {
	Event *MSContractEventNotClosed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MSContractEventNotClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MSContractEventNotClosed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MSContractEventNotClosed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MSContractEventNotClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MSContractEventNotClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MSContractEventNotClosed represents a EventNotClosed event raised by the MSContract contract.
type MSContractEventNotClosed struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEventNotClosed is a free log retrieval operation binding the contract event 0xcd08a02eb4f3c86c4b3352cbe9447051bfc1af76dd9388d388d23c5624a4524b.
//
// Solidity: e EventNotClosed()
func (_MSContract *MSContractFilterer) FilterEventNotClosed(opts *bind.FilterOpts) (*MSContractEventNotClosedIterator, error) {

	logs, sub, err := _MSContract.contract.FilterLogs(opts, "EventNotClosed")
	if err != nil {
		return nil, err
	}
	return &MSContractEventNotClosedIterator{contract: _MSContract.contract, event: "EventNotClosed", logs: logs, sub: sub}, nil
}

// WatchEventNotClosed is a free log subscription operation binding the contract event 0xcd08a02eb4f3c86c4b3352cbe9447051bfc1af76dd9388d388d23c5624a4524b.
//
// Solidity: e EventNotClosed()
func (_MSContract *MSContractFilterer) WatchEventNotClosed(opts *bind.WatchOpts, sink chan<- *MSContractEventNotClosed) (event.Subscription, error) {

	logs, sub, err := _MSContract.contract.WatchLogs(opts, "EventNotClosed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MSContractEventNotClosed)
				if err := _MSContract.contract.UnpackLog(event, "EventNotClosed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// MSContractEventRefundedIterator is returned from FilterEventRefunded and is used to iterate over the raw logs and unpacked data for EventRefunded events raised by the MSContract contract.
type MSContractEventRefundedIterator struct {
	Event *MSContractEventRefunded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MSContractEventRefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MSContractEventRefunded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MSContractEventRefunded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MSContractEventRefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MSContractEventRefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MSContractEventRefunded represents a EventRefunded event raised by the MSContract contract.
type MSContractEventRefunded struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEventRefunded is a free log retrieval operation binding the contract event 0xefbc23b4c571c79ee832603f52f6fafc9321a042b7c57a9bd82f809741db3849.
//
// Solidity: e EventRefunded()
func (_MSContract *MSContractFilterer) FilterEventRefunded(opts *bind.FilterOpts) (*MSContractEventRefundedIterator, error) {

	logs, sub, err := _MSContract.contract.FilterLogs(opts, "EventRefunded")
	if err != nil {
		return nil, err
	}
	return &MSContractEventRefundedIterator{contract: _MSContract.contract, event: "EventRefunded", logs: logs, sub: sub}, nil
}

// WatchEventRefunded is a free log subscription operation binding the contract event 0xefbc23b4c571c79ee832603f52f6fafc9321a042b7c57a9bd82f809741db3849.
//
// Solidity: e EventRefunded()
func (_MSContract *MSContractFilterer) WatchEventRefunded(opts *bind.WatchOpts, sink chan<- *MSContractEventRefunded) (event.Subscription, error) {

	logs, sub, err := _MSContract.contract.WatchLogs(opts, "EventRefunded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MSContractEventRefunded)
				if err := _MSContract.contract.UnpackLog(event, "EventRefunded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// MSContractEventStateRegisteredIterator is returned from FilterEventStateRegistered and is used to iterate over the raw logs and unpacked data for EventStateRegistered events raised by the MSContract contract.
type MSContractEventStateRegisteredIterator struct {
	Event *MSContractEventStateRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MSContractEventStateRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MSContractEventStateRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MSContractEventStateRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MSContractEventStateRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MSContractEventStateRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MSContractEventStateRegistered represents a EventStateRegistered event raised by the MSContract contract.
type MSContractEventStateRegistered struct {
	BlockedAlice *big.Int
	BlockedBob   *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterEventStateRegistered is a free log retrieval operation binding the contract event 0xf025b0293fd70914bfa444833429b6c197b8d1aefa338e5d4e6fb3ccfd1aa251.
//
// Solidity: e EventStateRegistered(blockedAlice uint256, blockedBob uint256)
func (_MSContract *MSContractFilterer) FilterEventStateRegistered(opts *bind.FilterOpts) (*MSContractEventStateRegisteredIterator, error) {

	logs, sub, err := _MSContract.contract.FilterLogs(opts, "EventStateRegistered")
	if err != nil {
		return nil, err
	}
	return &MSContractEventStateRegisteredIterator{contract: _MSContract.contract, event: "EventStateRegistered", logs: logs, sub: sub}, nil
}

// WatchEventStateRegistered is a free log subscription operation binding the contract event 0xf025b0293fd70914bfa444833429b6c197b8d1aefa338e5d4e6fb3ccfd1aa251.
//
// Solidity: e EventStateRegistered(blockedAlice uint256, blockedBob uint256)
func (_MSContract *MSContractFilterer) WatchEventStateRegistered(opts *bind.WatchOpts, sink chan<- *MSContractEventStateRegistered) (event.Subscription, error) {

	logs, sub, err := _MSContract.contract.WatchLogs(opts, "EventStateRegistered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MSContractEventStateRegistered)
				if err := _MSContract.contract.UnpackLog(event, "EventStateRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// MSContractEventStateRegisteringIterator is returned from FilterEventStateRegistering and is used to iterate over the raw logs and unpacked data for EventStateRegistering events raised by the MSContract contract.
type MSContractEventStateRegisteringIterator struct {
	Event *MSContractEventStateRegistering // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MSContractEventStateRegisteringIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MSContractEventStateRegistering)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MSContractEventStateRegistering)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MSContractEventStateRegisteringIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MSContractEventStateRegisteringIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MSContractEventStateRegistering represents a EventStateRegistering event raised by the MSContract contract.
type MSContractEventStateRegistering struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEventStateRegistering is a free log retrieval operation binding the contract event 0x82ce8e9a702f5f21ca7b1a4bbc6fd40407c770d7ae87544871fe505a728c7e71.
//
// Solidity: e EventStateRegistering()
func (_MSContract *MSContractFilterer) FilterEventStateRegistering(opts *bind.FilterOpts) (*MSContractEventStateRegisteringIterator, error) {

	logs, sub, err := _MSContract.contract.FilterLogs(opts, "EventStateRegistering")
	if err != nil {
		return nil, err
	}
	return &MSContractEventStateRegisteringIterator{contract: _MSContract.contract, event: "EventStateRegistering", logs: logs, sub: sub}, nil
}

// WatchEventStateRegistering is a free log subscription operation binding the contract event 0x82ce8e9a702f5f21ca7b1a4bbc6fd40407c770d7ae87544871fe505a728c7e71.
//
// Solidity: e EventStateRegistering()
func (_MSContract *MSContractFilterer) WatchEventStateRegistering(opts *bind.WatchOpts, sink chan<- *MSContractEventStateRegistering) (event.Subscription, error) {

	logs, sub, err := _MSContract.contract.WatchLogs(opts, "EventStateRegistering")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MSContractEventStateRegistering)
				if err := _MSContract.contract.UnpackLog(event, "EventStateRegistering", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// VPCABI is the input ABI used to generate the binding from.
const VPCABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"libSig\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"alice\",\"type\":\"address\"},{\"name\":\"bob\",\"type\":\"address\"},{\"name\":\"sid\",\"type\":\"uint256\"}],\"name\":\"finalize\",\"outputs\":[{\"name\":\"v\",\"type\":\"bool\"},{\"name\":\"a\",\"type\":\"uint256\"},{\"name\":\"b\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"alice\",\"type\":\"address\"},{\"name\":\"bob\",\"type\":\"address\"},{\"name\":\"sid\",\"type\":\"uint256\"},{\"name\":\"version\",\"type\":\"uint256\"},{\"name\":\"aliceCash\",\"type\":\"uint256\"},{\"name\":\"bobCash\",\"type\":\"uint256\"},{\"name\":\"signA\",\"type\":\"bytes\"},{\"name\":\"signB\",\"type\":\"bytes\"}],\"name\":\"close\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"s\",\"outputs\":[{\"name\":\"AliceCash\",\"type\":\"uint256\"},{\"name\":\"BobCash\",\"type\":\"uint256\"},{\"name\":\"seqNo\",\"type\":\"uint256\"},{\"name\":\"validity\",\"type\":\"uint256\"},{\"name\":\"extendedValidity\",\"type\":\"uint256\"},{\"name\":\"open\",\"type\":\"bool\"},{\"name\":\"waitingForAlice\",\"type\":\"bool\"},{\"name\":\"waitingForBob\",\"type\":\"bool\"},{\"name\":\"init\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"id\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"states\",\"outputs\":[{\"name\":\"AliceCash\",\"type\":\"uint256\"},{\"name\":\"BobCash\",\"type\":\"uint256\"},{\"name\":\"seqNo\",\"type\":\"uint256\"},{\"name\":\"validity\",\"type\":\"uint256\"},{\"name\":\"extendedValidity\",\"type\":\"uint256\"},{\"name\":\"open\",\"type\":\"bool\"},{\"name\":\"waitingForAlice\",\"type\":\"bool\"},{\"name\":\"waitingForBob\",\"type\":\"bool\"},{\"name\":\"init\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_libSig\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_id\",\"type\":\"bytes32\"}],\"name\":\"EventVpcClosing\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_id\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"cashAlice\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"cashBob\",\"type\":\"uint256\"}],\"name\":\"EventVpcClosed\",\"type\":\"event\"}]"

// VPCBin is the compiled bytecode used for deploying new contracts.
const VPCBin = `0x608060405234801561001057600080fd5b50604051602080610e28833981016040525160088054600160a060020a031916600160a060020a03909216919091179055610dd8806100506000396000f30060806040526004361061005e5763ffffffff60e060020a6000350416631cd6bbbd81146100635780635e9d2afe146100945780635eb3d246146100de57806386b714e2146101a0578063af640d0f14610204578063fbdc1ef11461022b575b600080fd5b34801561006f57600080fd5b50610078610243565b60408051600160a060020a039092168252519081900360200190f35b3480156100a057600080fd5b506100be600160a060020a0360043581169060243516604435610252565b604080519315158452602084019290925282820152519081900360600190f35b3480156100ea57600080fd5b50604080516020601f60c43560048181013592830184900484028501840190955281845261019e94600160a060020a038135811695602480359092169560443595606435956084359560a435953695919460e49492939091019190819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506104329650505050505050565b005b3480156101ac57600080fd5b506101b5610d1e565b60408051998a5260208a01989098528888019690965260608801949094526080870192909252151560a0860152151560c0850152151560e0840152151561010083015251908190036101200190f35b34801561021057600080fd5b50610219610d52565b60408051918252519081900360200190f35b34801561023757600080fd5b506101b5600435610d58565b600854600160a060020a031681565b60008060008585856040516020018084600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140183600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140182815260200193505050506040516020818303038152906040526040518082805190602001908083835b602083106102f85780518252601f1990920191602091820191016102d9565b51815160209384036101000a60001901801990921691161790526040805192909401829003909120600781905560009081529081905291909120600501546301000000900460ff1615925061041f915050576007546000908152602081905260409020600401544211156103d05760078054600090815260208181526040808320600501805460ff19169055925480835291839020805460019190910154845191825291810191909152825191927f26aa54eb5022b6fca25647b788775f49744c0c0df0ab3f674b858856d4dbf00492918290030190a25b60075460009081526020819052604090206005015460ff16156103fb57506000915081905080610429565b50506007546000908152602081905260409020805460019182015491925090610429565b5060009150819050805b93509350939050565b600080808033600160a060020a038d161480610456575033600160a060020a038c16145b151561046157600080fd5b8b8b8b6040516020018084600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140183600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140182815260200193505050506040516020818303038152906040526040518082805190602001908083835b602083106105025780518252601f1990920191602091820191016104e3565b518151600019602094850361010090810a919091019182169119929092161790915260408051959093018590038520600781905560008181528084528490208054600190815581015460029081558101546003908155810154600490815581015460059081550180546006805460ff928316151560ff199091161780825583548690048316151590950261ff00199095169490941780855582546201000090819004831615150262ff00001990911617808555915463010000009081900490911615150263ff0000001990911617909155848201528382018f9052606084018e905260808085018e90528251808603909101815260a0909401918290528351939550909350839290850191508083835b602083106106315780518252601f199092019160209182019101610612565b51815160209384036101000a6000190180199092169116179052604080519290940182900382207f19457468657265756d205369676e6564204d6573736167653a0a33320000000083830152603c80840182905285518085039091018152605c909301948590528251909a509195509293508392850191508083835b602083106106cc5780518252601f1990920191602091820191016106ad565b6001836020036101000a03801982511681845116808217855250505050505090500191505060405180910390209250600860009054906101000a9004600160a060020a0316600160a060020a0316631a86b5508d85896040518463ffffffff1660e060020a0281526004018084600160a060020a0316600160a060020a03168152602001836000191660001916815260200180602001828103825283818151815260200191508051906020019080838360005b8381101561079757818101518382015260200161077f565b50505050905090810190601f1680156107c45780820380516001836020036101000a031916815260200191505b50945050505050602060405180830381600087803b1580156107e557600080fd5b505af11580156107f9573d6000803e3d6000fd5b505050506040513d602081101561080f57600080fd5b5051801561092f5750600860009054906101000a9004600160a060020a0316600160a060020a0316631a86b5508c85886040518463ffffffff1660e060020a0281526004018084600160a060020a0316600160a060020a03168152602001836000191660001916815260200180602001828103825283818151815260200191508051906020019080838360005b838110156108b457818101518382015260200161089c565b50505050905090810190601f1680156108e15780820380516001836020036101000a031916815260200191505b50945050505050602060405180830381600087803b15801561090257600080fd5b505af1158015610916573d6000803e3d6000fd5b505050506040513d602081101561092c57600080fd5b50515b151561093a57600080fd5b6006546301000000900460ff161515610a1e575050604080516101208101825287815260208101879052808201899052426102588101606083018190526104b090910160808301819052600160a0840181905260c0840181905260e084018190526101009384018190528a815560028a905560038c9055600483905560058290556006805463010000006201000060ff199290921690931761ff00191690951762ff000019169490941763ff00000019161790925560075492519092907fd2e6dd92165017c1e9454967126e093291ed52383300a83cc477adca7963128f90600090a25b60065460ff161580610a31575060055442115b15610a3b57610d10565b60045442118015610a66575033600160a060020a038d161480610a66575033600160a060020a038c16145b15610a7057610d10565b33600160a060020a038d161415610a8d576006805461ff00191690555b33600160a060020a038c161415610aab576006805462ff0000191690555b600354891115610be757610120604051908101604052808981526020018881526020018a815260200160016003015481526020016001600401548152602001600115158152602001600160050160019054906101000a900460ff1615158152602001600160050160029054906101000a900460ff1615158152602001600115158152506001600082015181600001556020820151816001015560408201518160020155606082015181600301556080820151816004015560a08201518160050160006101000a81548160ff02191690831515021790555060c08201518160050160016101000a81548160ff02191690831515021790555060e08201518160050160026101000a81548160ff0219169083151502179055506101008201518160050160036101000a81548160ff0219169083151502179055509050505b600654610100900460ff16158015610c08575060065462010000900460ff16155b15610c5c576006805460ff1916905560075460015460025460408051928352602083019190915280517f26aa54eb5022b6fca25647b788775f49744c0c0df0ab3f674b858856d4dbf0049281900390910190a25b600754600090815260208190526040902060018054825560028054918301919091556003805491830191909155600480549183019190915560058054918301919091556006805491909201805460ff191660ff928316151517808255835461ff0019909116610100918290048416151590910217808255835462ff00001990911662010000918290048416151590910217808255925463ff0000001990931663010000009384900490921615159092021790555b505050505050505050505050565b60015460025460035460045460055460065460ff808216916101008104821691620100008204811691630100000090041689565b60075481565b600060208190529081526040902080546001820154600283015460038401546004850154600590950154939492939192909160ff8082169161010081048216916201000082048116916301000000900416895600a165627a7a72305820961efcd27e83ce67d61fb863cf045166b11ee0a6c8dc361d868669471577f74c0029`

// DeployVPC deploys a new Ethereum contract, binding an instance of VPC to it.
func DeployVPC(auth *bind.TransactOpts, backend bind.ContractBackend, _libSig common.Address) (common.Address, *types.Transaction, *VPC, error) {
	parsed, err := abi.JSON(strings.NewReader(VPCABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(VPCBin), backend, _libSig)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &VPC{VPCCaller: VPCCaller{contract: contract}, VPCTransactor: VPCTransactor{contract: contract}, VPCFilterer: VPCFilterer{contract: contract}}, nil
}

// VPC is an auto generated Go binding around an Ethereum contract.
type VPC struct {
	VPCCaller     // Read-only binding to the contract
	VPCTransactor // Write-only binding to the contract
	VPCFilterer   // Log filterer for contract events
}

// VPCCaller is an auto generated read-only Go binding around an Ethereum contract.
type VPCCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VPCTransactor is an auto generated write-only Go binding around an Ethereum contract.
type VPCTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VPCFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type VPCFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VPCSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type VPCSession struct {
	Contract     *VPC              // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VPCCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type VPCCallerSession struct {
	Contract *VPCCaller    // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// VPCTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type VPCTransactorSession struct {
	Contract     *VPCTransactor    // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VPCRaw is an auto generated low-level Go binding around an Ethereum contract.
type VPCRaw struct {
	Contract *VPC // Generic contract binding to access the raw methods on
}

// VPCCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type VPCCallerRaw struct {
	Contract *VPCCaller // Generic read-only contract binding to access the raw methods on
}

// VPCTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type VPCTransactorRaw struct {
	Contract *VPCTransactor // Generic write-only contract binding to access the raw methods on
}

// NewVPC creates a new instance of VPC, bound to a specific deployed contract.
func NewVPC(address common.Address, backend bind.ContractBackend) (*VPC, error) {
	contract, err := bindVPC(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &VPC{VPCCaller: VPCCaller{contract: contract}, VPCTransactor: VPCTransactor{contract: contract}, VPCFilterer: VPCFilterer{contract: contract}}, nil
}

// NewVPCCaller creates a new read-only instance of VPC, bound to a specific deployed contract.
func NewVPCCaller(address common.Address, caller bind.ContractCaller) (*VPCCaller, error) {
	contract, err := bindVPC(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &VPCCaller{contract: contract}, nil
}

// NewVPCTransactor creates a new write-only instance of VPC, bound to a specific deployed contract.
func NewVPCTransactor(address common.Address, transactor bind.ContractTransactor) (*VPCTransactor, error) {
	contract, err := bindVPC(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &VPCTransactor{contract: contract}, nil
}

// NewVPCFilterer creates a new log filterer instance of VPC, bound to a specific deployed contract.
func NewVPCFilterer(address common.Address, filterer bind.ContractFilterer) (*VPCFilterer, error) {
	contract, err := bindVPC(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &VPCFilterer{contract: contract}, nil
}

// bindVPC binds a generic wrapper to an already deployed contract.
func bindVPC(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(VPCABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VPC *VPCRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _VPC.Contract.VPCCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VPC *VPCRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VPC.Contract.VPCTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VPC *VPCRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VPC.Contract.VPCTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VPC *VPCCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _VPC.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VPC *VPCTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VPC.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VPC *VPCTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VPC.Contract.contract.Transact(opts, method, params...)
}

// Id is a free data retrieval call binding the contract method 0xaf640d0f.
//
// Solidity: function id() constant returns(bytes32)
func (_VPC *VPCCaller) Id(opts *bind.CallOpts) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _VPC.contract.Call(opts, out, "id")
	return *ret0, err
}

// Id is a free data retrieval call binding the contract method 0xaf640d0f.
//
// Solidity: function id() constant returns(bytes32)
func (_VPC *VPCSession) Id() ([32]byte, error) {
	return _VPC.Contract.Id(&_VPC.CallOpts)
}

// Id is a free data retrieval call binding the contract method 0xaf640d0f.
//
// Solidity: function id() constant returns(bytes32)
func (_VPC *VPCCallerSession) Id() ([32]byte, error) {
	return _VPC.Contract.Id(&_VPC.CallOpts)
}

// LibSig is a free data retrieval call binding the contract method 0x1cd6bbbd.
//
// Solidity: function libSig() constant returns(address)
func (_VPC *VPCCaller) LibSig(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _VPC.contract.Call(opts, out, "libSig")
	return *ret0, err
}

// LibSig is a free data retrieval call binding the contract method 0x1cd6bbbd.
//
// Solidity: function libSig() constant returns(address)
func (_VPC *VPCSession) LibSig() (common.Address, error) {
	return _VPC.Contract.LibSig(&_VPC.CallOpts)
}

// LibSig is a free data retrieval call binding the contract method 0x1cd6bbbd.
//
// Solidity: function libSig() constant returns(address)
func (_VPC *VPCCallerSession) LibSig() (common.Address, error) {
	return _VPC.Contract.LibSig(&_VPC.CallOpts)
}

// S is a free data retrieval call binding the contract method 0x86b714e2.
//
// Solidity: function s() constant returns(AliceCash uint256, BobCash uint256, seqNo uint256, validity uint256, extendedValidity uint256, open bool, waitingForAlice bool, waitingForBob bool, init bool)
func (_VPC *VPCCaller) S(opts *bind.CallOpts) (struct {
	AliceCash        *big.Int
	BobCash          *big.Int
	SeqNo            *big.Int
	Validity         *big.Int
	ExtendedValidity *big.Int
	Open             bool
	WaitingForAlice  bool
	WaitingForBob    bool
	Init             bool
}, error) {
	ret := new(struct {
		AliceCash        *big.Int
		BobCash          *big.Int
		SeqNo            *big.Int
		Validity         *big.Int
		ExtendedValidity *big.Int
		Open             bool
		WaitingForAlice  bool
		WaitingForBob    bool
		Init             bool
	})
	out := ret
	err := _VPC.contract.Call(opts, out, "s")
	return *ret, err
}

// S is a free data retrieval call binding the contract method 0x86b714e2.
//
// Solidity: function s() constant returns(AliceCash uint256, BobCash uint256, seqNo uint256, validity uint256, extendedValidity uint256, open bool, waitingForAlice bool, waitingForBob bool, init bool)
func (_VPC *VPCSession) S() (struct {
	AliceCash        *big.Int
	BobCash          *big.Int
	SeqNo            *big.Int
	Validity         *big.Int
	ExtendedValidity *big.Int
	Open             bool
	WaitingForAlice  bool
	WaitingForBob    bool
	Init             bool
}, error) {
	return _VPC.Contract.S(&_VPC.CallOpts)
}

// S is a free data retrieval call binding the contract method 0x86b714e2.
//
// Solidity: function s() constant returns(AliceCash uint256, BobCash uint256, seqNo uint256, validity uint256, extendedValidity uint256, open bool, waitingForAlice bool, waitingForBob bool, init bool)
func (_VPC *VPCCallerSession) S() (struct {
	AliceCash        *big.Int
	BobCash          *big.Int
	SeqNo            *big.Int
	Validity         *big.Int
	ExtendedValidity *big.Int
	Open             bool
	WaitingForAlice  bool
	WaitingForBob    bool
	Init             bool
}, error) {
	return _VPC.Contract.S(&_VPC.CallOpts)
}

// States is a free data retrieval call binding the contract method 0xfbdc1ef1.
//
// Solidity: function states( bytes32) constant returns(AliceCash uint256, BobCash uint256, seqNo uint256, validity uint256, extendedValidity uint256, open bool, waitingForAlice bool, waitingForBob bool, init bool)
func (_VPC *VPCCaller) States(opts *bind.CallOpts, arg0 [32]byte) (struct {
	AliceCash        *big.Int
	BobCash          *big.Int
	SeqNo            *big.Int
	Validity         *big.Int
	ExtendedValidity *big.Int
	Open             bool
	WaitingForAlice  bool
	WaitingForBob    bool
	Init             bool
}, error) {
	ret := new(struct {
		AliceCash        *big.Int
		BobCash          *big.Int
		SeqNo            *big.Int
		Validity         *big.Int
		ExtendedValidity *big.Int
		Open             bool
		WaitingForAlice  bool
		WaitingForBob    bool
		Init             bool
	})
	out := ret
	err := _VPC.contract.Call(opts, out, "states", arg0)
	return *ret, err
}

// States is a free data retrieval call binding the contract method 0xfbdc1ef1.
//
// Solidity: function states( bytes32) constant returns(AliceCash uint256, BobCash uint256, seqNo uint256, validity uint256, extendedValidity uint256, open bool, waitingForAlice bool, waitingForBob bool, init bool)
func (_VPC *VPCSession) States(arg0 [32]byte) (struct {
	AliceCash        *big.Int
	BobCash          *big.Int
	SeqNo            *big.Int
	Validity         *big.Int
	ExtendedValidity *big.Int
	Open             bool
	WaitingForAlice  bool
	WaitingForBob    bool
	Init             bool
}, error) {
	return _VPC.Contract.States(&_VPC.CallOpts, arg0)
}

// States is a free data retrieval call binding the contract method 0xfbdc1ef1.
//
// Solidity: function states( bytes32) constant returns(AliceCash uint256, BobCash uint256, seqNo uint256, validity uint256, extendedValidity uint256, open bool, waitingForAlice bool, waitingForBob bool, init bool)
func (_VPC *VPCCallerSession) States(arg0 [32]byte) (struct {
	AliceCash        *big.Int
	BobCash          *big.Int
	SeqNo            *big.Int
	Validity         *big.Int
	ExtendedValidity *big.Int
	Open             bool
	WaitingForAlice  bool
	WaitingForBob    bool
	Init             bool
}, error) {
	return _VPC.Contract.States(&_VPC.CallOpts, arg0)
}

// Close is a paid mutator transaction binding the contract method 0x5eb3d246.
//
// Solidity: function close(alice address, bob address, sid uint256, version uint256, aliceCash uint256, bobCash uint256, signA bytes, signB bytes) returns()
func (_VPC *VPCTransactor) Close(opts *bind.TransactOpts, alice common.Address, bob common.Address, sid *big.Int, version *big.Int, aliceCash *big.Int, bobCash *big.Int, signA []byte, signB []byte) (*types.Transaction, error) {
	return _VPC.contract.Transact(opts, "close", alice, bob, sid, version, aliceCash, bobCash, signA, signB)
}

// Close is a paid mutator transaction binding the contract method 0x5eb3d246.
//
// Solidity: function close(alice address, bob address, sid uint256, version uint256, aliceCash uint256, bobCash uint256, signA bytes, signB bytes) returns()
func (_VPC *VPCSession) Close(alice common.Address, bob common.Address, sid *big.Int, version *big.Int, aliceCash *big.Int, bobCash *big.Int, signA []byte, signB []byte) (*types.Transaction, error) {
	return _VPC.Contract.Close(&_VPC.TransactOpts, alice, bob, sid, version, aliceCash, bobCash, signA, signB)
}

// Close is a paid mutator transaction binding the contract method 0x5eb3d246.
//
// Solidity: function close(alice address, bob address, sid uint256, version uint256, aliceCash uint256, bobCash uint256, signA bytes, signB bytes) returns()
func (_VPC *VPCTransactorSession) Close(alice common.Address, bob common.Address, sid *big.Int, version *big.Int, aliceCash *big.Int, bobCash *big.Int, signA []byte, signB []byte) (*types.Transaction, error) {
	return _VPC.Contract.Close(&_VPC.TransactOpts, alice, bob, sid, version, aliceCash, bobCash, signA, signB)
}

// Finalize is a paid mutator transaction binding the contract method 0x5e9d2afe.
//
// Solidity: function finalize(alice address, bob address, sid uint256) returns(v bool, a uint256, b uint256)
func (_VPC *VPCTransactor) Finalize(opts *bind.TransactOpts, alice common.Address, bob common.Address, sid *big.Int) (*types.Transaction, error) {
	return _VPC.contract.Transact(opts, "finalize", alice, bob, sid)
}

// Finalize is a paid mutator transaction binding the contract method 0x5e9d2afe.
//
// Solidity: function finalize(alice address, bob address, sid uint256) returns(v bool, a uint256, b uint256)
func (_VPC *VPCSession) Finalize(alice common.Address, bob common.Address, sid *big.Int) (*types.Transaction, error) {
	return _VPC.Contract.Finalize(&_VPC.TransactOpts, alice, bob, sid)
}

// Finalize is a paid mutator transaction binding the contract method 0x5e9d2afe.
//
// Solidity: function finalize(alice address, bob address, sid uint256) returns(v bool, a uint256, b uint256)
func (_VPC *VPCTransactorSession) Finalize(alice common.Address, bob common.Address, sid *big.Int) (*types.Transaction, error) {
	return _VPC.Contract.Finalize(&_VPC.TransactOpts, alice, bob, sid)
}

// VPCEventVpcClosedIterator is returned from FilterEventVpcClosed and is used to iterate over the raw logs and unpacked data for EventVpcClosed events raised by the VPC contract.
type VPCEventVpcClosedIterator struct {
	Event *VPCEventVpcClosed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VPCEventVpcClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VPCEventVpcClosed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VPCEventVpcClosed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VPCEventVpcClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VPCEventVpcClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VPCEventVpcClosed represents a EventVpcClosed event raised by the VPC contract.
type VPCEventVpcClosed struct {
	Id        [32]byte
	CashAlice *big.Int
	CashBob   *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEventVpcClosed is a free log retrieval operation binding the contract event 0x26aa54eb5022b6fca25647b788775f49744c0c0df0ab3f674b858856d4dbf004.
//
// Solidity: e EventVpcClosed(_id indexed bytes32, cashAlice uint256, cashBob uint256)
func (_VPC *VPCFilterer) FilterEventVpcClosed(opts *bind.FilterOpts, _id [][32]byte) (*VPCEventVpcClosedIterator, error) {

	var _idRule []interface{}
	for _, _idItem := range _id {
		_idRule = append(_idRule, _idItem)
	}

	logs, sub, err := _VPC.contract.FilterLogs(opts, "EventVpcClosed", _idRule)
	if err != nil {
		return nil, err
	}
	return &VPCEventVpcClosedIterator{contract: _VPC.contract, event: "EventVpcClosed", logs: logs, sub: sub}, nil
}

// WatchEventVpcClosed is a free log subscription operation binding the contract event 0x26aa54eb5022b6fca25647b788775f49744c0c0df0ab3f674b858856d4dbf004.
//
// Solidity: e EventVpcClosed(_id indexed bytes32, cashAlice uint256, cashBob uint256)
func (_VPC *VPCFilterer) WatchEventVpcClosed(opts *bind.WatchOpts, sink chan<- *VPCEventVpcClosed, _id [][32]byte) (event.Subscription, error) {

	var _idRule []interface{}
	for _, _idItem := range _id {
		_idRule = append(_idRule, _idItem)
	}

	logs, sub, err := _VPC.contract.WatchLogs(opts, "EventVpcClosed", _idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VPCEventVpcClosed)
				if err := _VPC.contract.UnpackLog(event, "EventVpcClosed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// VPCEventVpcClosingIterator is returned from FilterEventVpcClosing and is used to iterate over the raw logs and unpacked data for EventVpcClosing events raised by the VPC contract.
type VPCEventVpcClosingIterator struct {
	Event *VPCEventVpcClosing // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VPCEventVpcClosingIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VPCEventVpcClosing)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VPCEventVpcClosing)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VPCEventVpcClosingIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VPCEventVpcClosingIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VPCEventVpcClosing represents a EventVpcClosing event raised by the VPC contract.
type VPCEventVpcClosing struct {
	Id  [32]byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEventVpcClosing is a free log retrieval operation binding the contract event 0xd2e6dd92165017c1e9454967126e093291ed52383300a83cc477adca7963128f.
//
// Solidity: e EventVpcClosing(_id indexed bytes32)
func (_VPC *VPCFilterer) FilterEventVpcClosing(opts *bind.FilterOpts, _id [][32]byte) (*VPCEventVpcClosingIterator, error) {

	var _idRule []interface{}
	for _, _idItem := range _id {
		_idRule = append(_idRule, _idItem)
	}

	logs, sub, err := _VPC.contract.FilterLogs(opts, "EventVpcClosing", _idRule)
	if err != nil {
		return nil, err
	}
	return &VPCEventVpcClosingIterator{contract: _VPC.contract, event: "EventVpcClosing", logs: logs, sub: sub}, nil
}

// WatchEventVpcClosing is a free log subscription operation binding the contract event 0xd2e6dd92165017c1e9454967126e093291ed52383300a83cc477adca7963128f.
//
// Solidity: e EventVpcClosing(_id indexed bytes32)
func (_VPC *VPCFilterer) WatchEventVpcClosing(opts *bind.WatchOpts, sink chan<- *VPCEventVpcClosing, _id [][32]byte) (event.Subscription, error) {

	var _idRule []interface{}
	for _, _idItem := range _id {
		_idRule = append(_idRule, _idItem)
	}

	logs, sub, err := _VPC.contract.WatchLogs(opts, "EventVpcClosing", _idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VPCEventVpcClosing)
				if err := _VPC.contract.UnpackLog(event, "EventVpcClosing", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
