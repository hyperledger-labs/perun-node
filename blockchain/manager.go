// Copyright (c) 2019 - for information on the respective copyright owner
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

package blockchain

import (
	"fmt"
	"math/big"

	"github.com/direct-state-transfer/dst-go/ethereum/adapter"
	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
	"github.com/direct-state-transfer/dst-go/log"
	"golang.org/x/net/context"
)

var packageName = "blockchain"

// EventsChan provides an interface to each event subscription and channel to receive the events defined in offchain protocol.
type EventsChan struct {
	MSCInitializingChan     chan *contract.MSContractEventInitializing
	MSCInitializedChan      chan *contract.MSContractEventInitialized
	MSCStateRegisteringChan chan *contract.MSContractEventStateRegistering
	MSCStateRegisteredChan  chan *contract.MSContractEventStateRegistered
	MSCClosingChan          chan *contract.MSContractEventClosing
	MSCClosedChan           chan *contract.MSContractEventClosed
	VPCVPCClosingChan       chan *contract.VPCEventVpcClosing
	VPCVPCClosedChan        chan *contract.VPCEventVpcClosed

	MSCInitalizingSub      adapter.EventSubscription
	MSCInitalizedSub       adapter.EventSubscription
	MSCStateRegisteringSub adapter.EventSubscription
	MSCStateRegisterdSub   adapter.EventSubscription
	MSCClosingSub          adapter.EventSubscription
	MSCClosedSub           adapter.EventSubscription
	VPCVPCClosingSub       adapter.EventSubscription
	VPCVPCClosedSub        adapter.EventSubscription
}

func init() {
	if logger == nil {
		logger, _ = log.NewLogger(log.InfoLevel, log.StdoutBackend, packageName)
	}
}

// Instance represents a set of interfaces to interact with the blockchain.
// It consists of a blockchain connection, addresses and instances of used contracts and the id of the user from which blockchain calls are to be made.
type Instance struct {
	Conn adapter.ContractBackend

	OwnerID identity.OffChainID //Offchain identity of the user for who will sign transactions (offchain and onchain)

	libSignaturesAddr types.Address //Address at which lib signatures contract is deployed
	msContractAddr    types.Address //Address at which ms contract is deployed
	vpcAddr           types.Address //Address at which vpc is deployed

	MSContractInst *contract.MSContract
	VPCInst        *contract.VPC

	EventsChan EventsChan
}

// NewInstance initialises and returns a new blockchain instance.
func NewInstance(conn adapter.ContractBackend, ownerID identity.OffChainID) Instance {
	return Instance{
		Conn:    conn,
		OwnerID: ownerID,
	}
}

// LibSignatures returns the address of libSignatures contract used in the instance.
func (inst *Instance) LibSignatures() types.Address {
	return inst.libSignaturesAddr
}

// SetLibSignatures sets the address of libSignatures contract in the instance.
func (inst *Instance) SetLibSignatures(libsignaturesAddr types.Address) (err error) {

	//Validat the integrity of code at specified address
	match, err := adapter.VerifyCodeAt(libsignaturesAddr, contract.Store.LibSignatures().HashBinRuntimeFile, false, inst.Conn)
	if err != nil {
		return fmt.Errorf("error validating contract at given address - %v", err)
	}
	if match != contract.Match {
		return fmt.Errorf("contract at address does not match expected version. Got status - %v", match)
	}

	inst.libSignaturesAddr = libsignaturesAddr
	return nil
}

// MSContractAddr returns the address of mscontract used in the instance.
func (inst *Instance) MSContractAddr() types.Address {
	return inst.msContractAddr
}

// SetMSContractAddr sets the address of mscontract in the instance.
func (inst *Instance) SetMSContractAddr(msContractAddr types.Address) (err error) {

	//Validate the integrity of code at specified address
	match, err := adapter.VerifyCodeAt(msContractAddr, contract.Store.MSContract().HashBinRuntimeFile, false, inst.Conn)
	if err != nil {
		return fmt.Errorf("error validating contract at given address - %v", err)
	}
	if match != contract.Match {
		return fmt.Errorf("contract at address does not match expected version. Got status - %v", match)
	}

	inst.msContractAddr = msContractAddr
	return nil
}

// VPCAddr returns the address of vpc contract used in the instance.
func (inst *Instance) VPCAddr() types.Address {
	return inst.vpcAddr
}

// SetVPCAddr sets the address of vpc contract in the instance.
func (inst *Instance) SetVPCAddr(vpcAddr types.Address) (err error) {

	//Validate the integrity of code at specified address
	match, err := adapter.VerifyCodeAt(vpcAddr, contract.Store.VPC().HashBinRuntimeFile, false, inst.Conn)
	if err != nil {
		return fmt.Errorf("error validating contract at given address - %v", err)
	}
	if match != contract.Match {
		return fmt.Errorf("contract at address does not match expected version. Got status - %v", match)
	}

	inst.vpcAddr = vpcAddr
	return nil
}

// DeployVPC deploys a new vpc contract from the ownerID and sets the address in instance.
// It also initiases and set an vpc instance to access that contract.
func (inst *Instance) DeployVPC() (err error) {
	logger.Info("Deploying and setting up VPC Contract")
	var params []interface{}
	params = append(params, inst.LibSignatures())

	var vpcAddr types.Address
	var vpcInst *contract.VPC

	conn := inst.Conn
	vpcAddr, _, _, err = adapter.DeployContract(contract.Store.VPC(), conn, params, inst.OwnerID)
	if err != nil {
		return fmt.Errorf("deploy vpc error - %v", err)
	}
	conn.Commit()

	//Create a contract instance to make function calls or read/write variables
	vpcInst, err = contract.NewVPC(vpcAddr.Address, conn)
	if err != nil {
		return fmt.Errorf("instantiate vpc error -%v", err)
	}

	err = inst.SetVPCAddr(vpcAddr)
	if err != nil {
		return err
	}
	inst.VPCInst = vpcInst

	return nil
}

// DeployMSContract deploys a new mscontract from the ownerID and sets the address in instance.
// It also initiases and set an mscontract instance to access that contract.
func (inst *Instance) DeployMSContract(senderAddr, receiverAddr types.Address) (err error) {
	logger.Info("Deploying and setting up MS Contract")
	var params []interface{}
	params = append(params, inst.LibSignatures(), senderAddr, receiverAddr)

	var msContractAddr types.Address
	var msContractInst *contract.MSContract

	conn := inst.Conn
	msContractAddr, _, _, err = adapter.DeployContract(contract.Store.MSContract(), conn, params, inst.OwnerID)
	if err != nil {
		return fmt.Errorf("deploy MSContract error - %v", err)
	}
	conn.Commit()

	//Create a contract instance to make function calls or read/write variables
	msContractInst, err = contract.NewMSContract(msContractAddr.Address, conn)
	if err != nil {
		return fmt.Errorf("instantiate MSContract error - %v", err)
	}

	err = inst.SetMSContractAddr(msContractAddr)
	if err != nil {
		return err
	}
	inst.MSContractInst = msContractInst

	//Initialize background routine to listen for events from the deployed contract and put it to appropriate go channels
	eventsChan, err := inst.InitializeEventsChan()
	if err != nil {
		return fmt.Errorf("initializing events channel error - %v", err)
	}
	logger.Debug("Initialized filterers for reading contract events")
	inst.EventsChan = eventsChan

	return nil
}

// Confirm makes Confirm call on the deployed instance of MSContract.
// This call moves amountToBlock (in Wei) from Instance owner's account to contract account.
// It is the maximum value of amount that can blocked in an offchain channel on behalf of the this user.
func (inst *Instance) Confirm(amountToBlock *big.Int) (err error) {

	//Make transaction opts
	conn := inst.Conn
	transactOpts, err := adapter.MakeTransactOpts(conn, inst.OwnerID, amountToBlock, uint64(20e5))
	if err != nil {
		return fmt.Errorf("confirm() - txOpts - %v", err)
	}

	//Call function
	tx, err := inst.MSContractInst.MSContractTransactor.Confirm(transactOpts.TransactOpts)
	if err != nil {
		return fmt.Errorf("confirm() - function call - %v", err)
	}
	inst.Conn.Commit()

	hash := types.Hash{Hash: tx.Hash()}
	_, err = adapter.WaitTillTxMined(inst.Conn, hash)
	if err != nil {
		return fmt.Errorf("confirm() - tx not mined error - %v", err)
	}

	//Check transaction receipt for success / failure of execution
	txReceipt, err := inst.Conn.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return fmt.Errorf("confirm() - txReceipt - %v", err)
	}
	if txReceipt.Status == 0 {
		return fmt.Errorf("confirm() - txExecution status = 0 (fail)")
	}
	logger.Info("Amount confirmed and locked in contract:", amountToBlock)
	return nil
}

// StateRegister makes a StateRegister call on the deployed instance of MSContract.
// This call will register this user's confirmation for initial state (MSCBaseState) of the offchain channel.
//
// Sid should be the unique session id of the channel and Version its initial version.
// BlockedSender and BlockedReceiver (both in Wei) should be the initial balances of the respective users in offchain channel.
// SignSender and SignReceiver should be the valid signatures of the respective users over the MSCBaseState.
//
// For more information on SessionID and MSCBaseState, see offchain_primitve.go.
func (inst *Instance) StateRegister(Sid, Version *big.Int, BlockedSender *big.Int, BlockedReceiver *big.Int,
	SignSender, SignReceiver []byte) (err error) {

	//Make transaction opts
	conn := inst.Conn
	transactOpts, err := adapter.MakeTransactOpts(conn, inst.OwnerID, types.EtherToWei(big.NewInt(0)), uint64(30e4))
	if err != nil {
		return fmt.Errorf("stateRegister() - txOpts - %v", err)
	}

	//Call function
	tx, err := inst.MSContractInst.MSContractTransactor.StateRegister(
		transactOpts.TransactOpts,
		inst.VPCAddr().Address, Sid, BlockedSender, BlockedReceiver, Version, SignSender, SignReceiver)
	if err != nil {
		return fmt.Errorf("stateRegister() - function call - %v", err)
	}
	inst.Conn.Commit()

	hash := types.Hash{Hash: tx.Hash()}
	_, err = adapter.WaitTillTxMined(conn, hash)
	if err != nil {
		return fmt.Errorf("confirm() - tx not mined error - %v", err)
	}

	//Check transaction receipt for success / failure of execution
	txReceipt, err := conn.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return fmt.Errorf("stateRegister() - txReceipt - %v", err)
	}
	if txReceipt.Status == 0 {
		return fmt.Errorf("stateRegister() - txExecution status = 0 (fail)")
	}
	return nil
}

// VPCClose makes a VPCClose call on the deployed instance of VPC.
// This call register this user's request to finalise the state of the offchain channel.
//
// Sid should be the unique session id of the channel and Version its latest version.
// AddrSender and AddrReceiver should be the onchain address of the respective users in offchain channel.
// BlockedSender and BlockedReceiver (both in Wei) should be the latest balances of the respective users in offchain channel.
// SignSender and SignReceiver should be the valid signatures of the respective users over the latest VPCState.
//
// For more information on SessionID and VPCState, see offchain_primitve.go.
func (inst *Instance) VPCClose(Sid, Version *big.Int, AddrSender, AddrReceiver types.Address,
	BlockedSender *big.Int, BlockedReceiver *big.Int, SignSender, SignReceiver []byte) (err error) {

	//Make transaction opts
	conn := inst.Conn
	transactOpts, err := adapter.MakeTransactOpts(conn, inst.OwnerID, types.EtherToWei(big.NewInt(0)), uint64(40e4))
	if err != nil {
		return fmt.Errorf("vpcClose() - txOpts - %v", err)
	}

	//Call function
	tx, err := inst.VPCInst.Close(transactOpts.TransactOpts, AddrSender.Address, AddrReceiver.Address, Sid, Version,
		BlockedSender, BlockedReceiver, SignSender, SignReceiver)
	if err != nil {
		return fmt.Errorf("vpcClose() - function call - %v", err)
	}
	inst.Conn.Commit()

	hash := types.Hash{Hash: tx.Hash()}
	_, err = adapter.WaitTillTxMined(conn, hash)
	if err != nil {
		return fmt.Errorf("confirm() - tx not mined error - %v", err)
	}

	//Check transaction receipt for success / failure of execution
	txReceipt, err := conn.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return fmt.Errorf("vpcClose() - txReceipt - %v", err)
	}
	if txReceipt.Status == 0 {
		return fmt.Errorf("vpcClose() - txExecution status = 0 (fail)")
	}

	return nil
}

// Execute makes an Execute call on the deployed instance of MSContract.
// This call will distribute the funds back to users' accounts based on final states of the offchain channel.
//
// AddrSender and AddrReceiver should be the onchain address of the respective users in offchain channel.
func (inst *Instance) Execute(AddrSender, AddrReceiver types.Address) (err error) {

	//Make transaction opts
	conn := inst.Conn
	transactOpts, err := adapter.MakeTransactOpts(conn, inst.OwnerID, types.EtherToWei(big.NewInt(0)), uint64(40e4))
	if err != nil {
		return fmt.Errorf("execute() - txOpts - %v", err)
	}

	//Call function
	tx, err := inst.MSContractInst.MSContractTransactor.Execute(transactOpts.TransactOpts, AddrSender.Address, AddrReceiver.Address)
	if err != nil {
		return fmt.Errorf("execute() - function call - %v", err)
	}
	inst.Conn.Commit()

	hash := types.Hash{Hash: tx.Hash()}
	_, err = adapter.WaitTillTxMined(conn, hash)
	if err != nil {
		return fmt.Errorf("confirm() - tx not mined error - %v", err)
	}

	//Check transaction receipt for success / failure of execution
	txReceipt, err := conn.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return fmt.Errorf("execute() - txReceipt - %v", err)
	}
	if txReceipt.Status == 0 {
		return fmt.Errorf("execute() - txExecution status = 0 (fail)")
	}

	return nil
}

// States makes a (read only) States call on the deployed instance of vpc.
// This call will return the current state of vpc channel in blockchain.
//
// states will be a struct represting the last updated value of offchain state in the blockchain.
func (inst *Instance) States() (states interface{}, err error) {

	callOpts := adapter.MakeCallOpts(context.Background(), false, inst.OwnerID.OnChainID)

	//Call function
	states, err = inst.VPCInst.S(callOpts)
	if err != nil {
		return states, fmt.Errorf("states() - function call - %v", err)
	}

	return states, nil
}

// InitializeEventsChan initialises subscriptions for all the event defined in offchain protocol.
//
// List of events initialised currently
// MscEventInitialized, MscEventStateRegistering, MscEventStateRegistered, vpcEventClosing, vpcEventClosed, MscEventClosed
//
// Only for MSEventInitialing it also starts a filter query to check if the event has occurred in past.
// This special consideration is required because this occurs immediately after MSContract deploy is successful,
// and the event can be missed if there is a delay in the sharing of the deployed contract address between the node software instances.
func (inst *Instance) InitializeEventsChan() (eventsChan EventsChan, err error) {

	//Watch for MscEventInitializing
	eventsChan.MSCInitializingChan = make(chan *contract.MSContractEventInitializing, 2)
	eventsChan.MSCInitalizingSub, err = inst.MSContractInst.WatchEventInitializing(nil, eventsChan.MSCInitializingChan)
	if err != nil {
		err = fmt.Errorf("watch event initialzing error - %v", err)
		return EventsChan{}, err
	}

	//Filter for MsInitializing event and drop it into the channel, if it has occurred in the past
	go func(bcInst *Instance, ch chan *contract.MSContractEventInitializing) {

		var startBlock = uint64(0)
		var endBlock = (*uint64)(nil)
		filterOpts := adapter.MakeFilterOpts(context.TODO(), &startBlock, endBlock)
		itr, err := bcInst.MSContractInst.FilterEventInitializing(filterOpts)
		if err != nil {
			logger.Info("filter event initialzing error - " + err.Error())
			return
		}
		nextPresent := itr.Next()
		if nextPresent && itr.Event != nil {
			eventsChan.MSCInitializingChan <- itr.Event
		}
	}(inst, eventsChan.MSCInitializingChan)

	//Watch for MscEventInitialized
	eventsChan.MSCInitializedChan = make(chan *contract.MSContractEventInitialized, 2)
	eventsChan.MSCInitalizedSub, err = inst.MSContractInst.WatchEventInitialized(nil, eventsChan.MSCInitializedChan)
	if err != nil {
		err = fmt.Errorf("watch event initialzed error - %v", err)
		return EventsChan{}, err
	}

	//Watch for MscEventStateRegistering
	eventsChan.MSCStateRegisteringChan = make(chan *contract.MSContractEventStateRegistering, 10)
	eventsChan.MSCStateRegisteringSub, err = inst.MSContractInst.WatchEventStateRegistering(nil, eventsChan.MSCStateRegisteringChan)
	if err != nil {
		err = fmt.Errorf("watch event state registering error - %v", err)
		return EventsChan{}, err
	}

	//Watch for MscEventStateRegistered
	eventsChan.MSCStateRegisteredChan = make(chan *contract.MSContractEventStateRegistered, 10)
	eventsChan.MSCStateRegisterdSub, err = inst.MSContractInst.WatchEventStateRegistered(nil, eventsChan.MSCStateRegisteredChan)
	if err != nil {
		err = fmt.Errorf("watch event state registered error - %v", err)
		return EventsChan{}, err
	}

	//Watch for vpcEventClosing
	var _id [][32]byte
	eventsChan.VPCVPCClosingChan = make(chan *contract.VPCEventVpcClosing, 10)
	eventsChan.VPCVPCClosingSub, err = inst.VPCInst.WatchEventVpcClosing(nil, eventsChan.VPCVPCClosingChan, _id)
	if err != nil {
		err = fmt.Errorf("watch event vpc closing error - %v", err)
		return EventsChan{}, err
	}
	//Watch for vpcEventClosed
	eventsChan.VPCVPCClosedChan = make(chan *contract.VPCEventVpcClosed, 10)
	eventsChan.VPCVPCClosedSub, err = inst.VPCInst.WatchEventVpcClosed(nil, eventsChan.VPCVPCClosedChan, _id)
	if err != nil {
		err = fmt.Errorf("watch event vpc closed error - %v", err)
		return EventsChan{}, err
	}

	//Watch for MscEventClosed
	eventsChan.MSCClosedChan = make(chan *contract.MSContractEventClosed, 10)
	eventsChan.MSCClosedSub, err = inst.MSContractInst.WatchEventClosed(nil, eventsChan.MSCClosedChan)
	if err != nil {
		err = fmt.Errorf("watch event Msc close error - %v", err)
		return EventsChan{}, err
	}

	return eventsChan, nil
}

// InitModule initialises blockchain module. It initialises logger and also checks if libSignatures Contract at LibSignAddr is valid.
// If libSignAddr is empty, the validity check is skipped.
func InitModule(cfg *Config) (conn *adapter.RealBackend, libSignAddr types.Address, err error) {

	logger, err = log.NewLogger(cfg.Logger.Level, cfg.Logger.Backend, packageName)
	if err != nil {
		return nil, libSignAddr, err
	}

	//Initialise connection
	logger.Debug("Initialising Blockchain module")
	logger.Info("Connecting to blockchain node at ", cfg.gethURL, "...")

	//Establish a connection with blockchain node
	conn, err = adapter.NewRealBackend(cfg.gethURL)
	if err != nil {
		logger.Error(err.Error())
		return nil, libSignAddr, err
	}

	//Retrieve and store network id
	var networkID *big.Int
	networkID, err = conn.NetworkID(context.Background())
	if err != nil {
		logger.Error(err.Error())
		return nil, libSignAddr, err
	}
	cfg.networkID = networkID
	logger.Info("Connected to ethereum network. id :", networkID)

	//If libsignatures address is provided, validate it's integrity
	if cfg.libSignaturesAddr != types.HexToAddress("") {

		isMatch, err := adapter.VerifyCodeAt(cfg.libSignaturesAddr, contract.Store.LibSignatures().HashBinRuntimeFile, false, conn)

		if err != nil {
			logger.Error("Specified library contract address cannot be used. Error in verification -", err)
		}

		switch isMatch {
		case contract.Match:
			libSignAddr = cfg.libSignaturesAddr
			logger.Debug("Specified library contract address was used -", cfg.libSignaturesAddr.String())
		case contract.NoMatch:
			logger.Info("Specified library contract address cannot be used. Verification failed")
		default:
			logger.Debug("Unknown status on verifying contract address", isMatch)
		}

	} else {
		logger.Info("Library contract address not specified")
	}

	logger.Debug("Blockchain module successfully initialised")

	return conn, libSignAddr, nil
}

// SetupLibSignatures checks if deployed libSignatures contract is valid, if not it deploys a new instance and returns the address.
// Error is returned if the credentials of sessionOwner is not set or invalid.
func SetupLibSignatures(nodeLibSignaturesAddr types.Address, conn adapter.ContractBackend, sessionOwner identity.OffChainID) (
	libSignAddr types.Address, err error) {

	if nodeLibSignaturesAddr != types.HexToAddress("") {

		var isMatch contract.MatchStatus

		isMatch, err = adapter.VerifyCodeAt(nodeLibSignaturesAddr, contract.Store.LibSignatures().HashBinRuntimeFile, false, conn)

		switch {
		case err != nil:
			logger.Error("Specified library contract address cannot be used. Error in verification -", err)
			goto DeployLibSignatures

		case isMatch == contract.NoMatch:
			logger.Info("Specified library contract address cannot be used. Verification failed")
			goto DeployLibSignatures

		case isMatch == contract.Match:
			return nodeLibSignaturesAddr, nil
		}
	}

DeployLibSignatures:
	logger.Info("Deploying new lib signatures contract")
	libSignAddr, _, _, err = adapter.DeployContract(contract.Store.LibSignatures(), conn, nil, sessionOwner)

	return libSignAddr, err
}
