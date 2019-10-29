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

package channel

import (
	"fmt"
	"sync"

	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/identity"
	"github.com/direct-state-transfer/dst-go/log"
)

var packageName = "channel"

// Role is the role of the user in an activity (such as opening, closing) on channel.
type Role string

// Enumeration of allowed values for role in activity on channel.
const (
	// Sender is the one who initializes the activity on the channel.
	// The activity can be Opening or Closing the channel.
	Sender Role = Role("Sender")

	// Receiver is the user other than the one who initializes the activity on the channel.
	// The activity can be Opening or Closing the channel.
	Receiver Role = Role("Receiver")
)

// ClosingMode represents the closing mode for the vpc state channel.
// It determines what the node software will do when a channel closing notification is received.
type ClosingMode string

// Enumeration of allowed values for Closing mode.
const (
	// In ClosingModeManual, the information will be passed on to the user via api interface.
	// This will occur irrespective of the closing state being the latest or not.
	ClosingModeManual ClosingMode = ClosingMode("manual")

	// In ClosingModeNormal, if the closing state is the latest state no action will be taken,
	// so that channel will be closed after timeout.
	// Else if it is an older state, then node software will refute with latest state.
	ClosingModeAutoNormal ClosingMode = ClosingMode("auto-normal")

	// In ClosingModeNormal, if the closing state is the latest state it will also call close,
	// so that the channel will be immediately closed without waiting until timeout.
	// If it is an older state, the node software will refute with latest state.
	ClosingModeAutoImmediate ClosingMode = ClosingMode("auto-immediate")
)

// Status of the channel.
type Status string

// Enumeration of allowed values for Status of the channel.
const (
	PreSetup       Status = Status("pre-setup")        //Channel pre-setup at node in progress
	Setup          Status = Status("setup")            //Channel setup at node in progress
	Init           Status = Status("init")             //Channel status Init defined in perun api description
	Open           Status = Status("open")             //Channel status Open defined in perun api description
	InConflict     Status = Status("in-conflict")      //Channel status In-Conflict defined in perun api description
	Settled        Status = Status("settled")          //Channel status Settled defined in perun api description
	WaitingToClose Status = Status("waiting-to-close") //Channel status Waiting-To-Close defined in perun api description
	VPCClosing     Status = Status("vpc-closing")      //Channel close invoked by one of the participants
	VPCClosed      Status = Status("vpc-closed")       //Channel close invoked by both the participants
	Closed         Status = Status("closed")           //Channel is closed. Funds redistributed and mscontract self destructed
)

// InitModule initializes this module with provided configuration.
// The logger is initialized.
func InitModule(cfg *Config) (err error) {

	logger, err = log.NewLogger(cfg.Logger.Level, cfg.Logger.Backend, packageName)
	if err != nil {
		logger.Error(err)
		return err
	}

	//Initialise connection
	logger.Debug("Initializing Channel module")

	return nil

}

// Instance represents an instance of offchain channel.
// It groups all the properties of the channel such as identity and role of each user,
// current and all previous values of channel state.
type Instance struct {
	adapter ReadWriteCloser

	closingMode ClosingMode //Configure Closing mode for channel. Takes only predefined constants

	selfID      identity.OffChainID //Identity of the self
	peerID      identity.OffChainID //Identity of the peer
	roleChannel Role                //Role in channel. Takes only predefined constants
	roleClosing Role                //Role in closing. Takes only predefined constants

	status        Status             //Status of the channel
	contractStore contract.StoreType //ContractStore used for this channel
	sessionID     SessionID          //Session Id agreed for this offchain transaction
	mscBaseState  MSCBaseStateSigned //MSContract Base state to use for state register
	vpcStatesList []VPCStateSigned   //List of all vpc state

	access sync.Mutex //Access control when setting connection status

}

// Connected returns if the channel connecton is currently active.
func (inst *Instance) Connected() bool {
	if inst.adapter == nil {
		return false
	}
	return inst.adapter.Connected()
}

// Close closes the channel.
func (inst *Instance) Close() (err error) {
	if inst.adapter == nil {
		return fmt.Errorf("adapter is nil")
	}
	return inst.adapter.Close()
}

// SetClosingMode sets the closing mode for the channel.
// Closing mode will determine what how the node software will act when a vpc closing event is received.
func (inst *Instance) SetClosingMode(closingMode ClosingMode) {
	if closingMode == ClosingModeManual || closingMode == ClosingModeAutoNormal || closingMode == ClosingModeAutoImmediate {
		inst.closingMode = closingMode
	}
}

// ClosingMode returns the current closing mode configuration of the channel.
func (inst *Instance) ClosingMode() ClosingMode {
	return inst.closingMode
}

// setSelfID sets the self id of the channel.
func (inst *Instance) setSelfID(selfID identity.OffChainID) {
	inst.selfID = selfID
}

// SelfID returns the id of this user as configured in the channel.
func (inst *Instance) SelfID() identity.OffChainID {
	return inst.selfID
}

// setPeerID sets the peer id of the channel.
func (inst *Instance) setPeerID(peerID identity.OffChainID) {
	inst.peerID = peerID
}

// PeerID returns the id of the peer in the channel.
func (inst *Instance) PeerID() identity.OffChainID {
	return inst.peerID
}

// SenderID returns the id of sender in the channel.
// Sender is the one who initialsed the channel connection.
func (inst *Instance) SenderID() identity.OffChainID {
	switch inst.roleChannel {
	case Sender:
		return inst.selfID
	case Receiver:
		return inst.peerID
	default:
		return identity.OffChainID{}
	}
}

// ReceiverID returns the id of receiver in the channel.
// Receiver is the one who received a new channel connection request and accepted it.
func (inst *Instance) ReceiverID() identity.OffChainID {
	switch inst.roleChannel {
	case Receiver:
		return inst.selfID
	case Sender:
		return inst.peerID
	default:
		return identity.OffChainID{}
	}
}

// SetRoleChannel sets the role of the self user in the channel.
func (inst *Instance) SetRoleChannel(role Role) {
	if role == Sender || role == Receiver {
		inst.roleChannel = role
	}
}

// RoleChannel returns the role of the self user in the channel.
func (inst *Instance) RoleChannel() Role {
	return inst.roleChannel
}

// SetRoleClosing sets the role of the self user in the channel closing procedure.
// If this user initializes the closing procedure, role is sender else it is receiver.
func (inst *Instance) SetRoleClosing(role Role) {
	if role == Sender || role == Receiver {
		inst.roleClosing = role
	}
}

// RoleClosing returns the role of the self user in the channel closing procedure.
// If this user initializes the closing procedure, role is sender else it is receiver.
func (inst *Instance) RoleClosing() Role {
	return inst.roleClosing
}

// SetStatus sets the current status of the channel and returns true if the status was successfully updated.
//
// Only specific status changes are allowed. For example, new status can be set to Setup only when the current status is PreSetup,
// if not, the status change will not occur and false is returned.
func (inst *Instance) SetStatus(status Status) bool {

	inst.access.Lock()
	defer inst.access.Unlock()

	switch status {
	case Setup:
		if inst.status != PreSetup {
			return false
		}
	case Open:
		if inst.status != Init {
			return false
		}
	case InConflict:
		if !((inst.status == Open) || (inst.status == WaitingToClose)) {
			return false
		}
	case Settled:
		if inst.status != InConflict {
			return false
		}
	case WaitingToClose:
		if inst.status != Open {
			return false
		}
	case VPCClosing:
		if inst.status != Settled {
			return false
		}
	case VPCClosed:
		if inst.status != VPCClosing {
			return false
		}
	case Closed:
		if !((inst.status == Init) || (inst.status == VPCClosing) || (inst.status == VPCClosed) || (inst.status == WaitingToClose)) {
			return false
		}
	default:
		return false
	}
	inst.status = status
	return true
}

// Status returns the current status of the channel.
func (inst *Instance) Status() Status {
	return inst.status
}

// SetSessionID validates and sets the session id in channel instance.
// If validation fails, the values is not set in channel instance and an error is returned.
func (inst *Instance) SetSessionID(sessionID SessionID) (err error) {
	isValid, err := sessionID.Validate()
	if !isValid {
		return fmt.Errorf("Session id invalid - %v", err.Error())
	}
	inst.sessionID = sessionID
	return nil
}

// SessionID returns the session id of the channel.
func (inst *Instance) SessionID() SessionID {
	return inst.sessionID
}

// SetContractStore sets contract store in the channel instance.
// ContractStore is set of contracts and its properties according that facilitates this offchain channel.
func (inst *Instance) SetContractStore(contractStore contract.StoreType) {
	inst.contractStore = contractStore
}

// ContractStore returns the contract store that is configured in the channel instance.
func (inst *Instance) ContractStore() contract.StoreType {
	return inst.contractStore
}

// SetMSCBaseState validates the integrity of newState and if successful, sets the msc base state of the channel.
func (inst *Instance) SetMSCBaseState(newState MSCBaseStateSigned) (err error) {

	//Validate integrity of the sender signature on the state
	isValidSender, err := newState.VerifySign(inst.SenderID(), Sender)
	if err != nil {
		return err
	}
	if !isValidSender {
		return fmt.Errorf("Sender signature on MSCBaseState invalid")
	}

	//Validate integrity of the receiver signature on the state
	isValidReceiver, err := newState.VerifySign(inst.ReceiverID(), Receiver)
	if err != nil {
		return err
	}
	if !isValidReceiver {
		return fmt.Errorf("Receiver signature on MSCBaseState invalid")
	}
	logger.Debug("New MSC base state set")
	inst.mscBaseState = newState
	return nil
}

// MscBaseState returns the msc base state of the channel.
func (inst *Instance) MscBaseState() MSCBaseStateSigned {
	return inst.mscBaseState
}

// SetCurrentVPCState validates the integrity of newState and if successful, sets the current vpc state of the channel.
func (inst *Instance) SetCurrentVPCState(newState VPCStateSigned) (err error) {

	//Validate integrity of the sender signature on the state
	isValidSender, err := newState.VerifySign(inst.SenderID(), Sender)
	if err != nil {
		return err
	}
	if !isValidSender {
		return fmt.Errorf("Sender signature on VPCState invalid")
	}

	//Validate integrity of the receiver signature on the state
	isValidReceiver, err := newState.VerifySign(inst.ReceiverID(), Receiver)
	if err != nil {
		return err
	}
	if !isValidReceiver {
		return fmt.Errorf("Receiver signature on VPCState invalid")
	}

	if lastVpcStateIndex := len(inst.vpcStatesList) - 1; lastVpcStateIndex != -1 {

		//when previous state exists, check if the current version number is greater than previous
		lastValidStateVersion := inst.vpcStatesList[lastVpcStateIndex].VPCState.Version
		if newState.VPCState.Version.Cmp(lastValidStateVersion) != 1 {
			return fmt.Errorf("Current Version number (%s) less than previous (%s)", newState.VPCState.Version.String(), lastValidStateVersion.String())
		}
	}
	logger.Debug("New MSC base state set")
	inst.vpcStatesList = append(inst.vpcStatesList, newState)
	return nil
}

// CurrentVpcState returns the current vpc state of the channel.
func (inst *Instance) CurrentVpcState() VPCStateSigned {
	return inst.vpcStatesList[len(inst.vpcStatesList)-1]
}

// NewSession initializes and returns a new channel session.
// Channel session has a listener running in the background with defined adapterType.
// All new incoming connections are processed by the session and if successful made available on idVerifiedConn channel.
// The higher layers of code can listen for new connections on this idVerifiedConn channel and use it for further communications.
func NewSession(selfID identity.OffChainID, adapterType AdapterType, maxConn uint32) (idVerifiedConn chan *Instance,
	listener Shutdown, err error) {

	//Start a new listener
	idVerifiedConn, listener, err = startListener(selfID, maxConn, adapterType)
	if err != nil {
		logger.Error("Error starting listener", err)
		return nil, nil, err
	}

	//Do a loopback test
	ch, err := NewChannel(selfID, selfID, WebSocket)
	if err != nil {
		logger.Error("Channel self check - Error in outgoing connection -", err)
		return nil, nil, err
	}
	err = ch.Close()
	if err != nil {
		logger.Error("Channel self check - Error in closing channel:", err)
		return nil, nil, err
	}
	<-idVerifiedConn //Remove the loopback test connection

	logger.Debug("Channel self check success")
	return idVerifiedConn, listener, nil
}
