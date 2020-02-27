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

	"github.com/direct-state-transfer/dst-go/channel/primitives"
	"github.com/direct-state-transfer/dst-go/ethereum/contract"
	"github.com/direct-state-transfer/dst-go/identity"
	"github.com/direct-state-transfer/dst-go/log"
)

var packageName = "channel"

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
	roleChannel primitives.Role     //Role in channel. Takes only predefined constants
	roleClosing primitives.Role     //Role in closing. Takes only predefined constants

	status        Status                        //Status of the channel
	contractStore contract.StoreType            //ContractStore used for this channel
	sessionID     primitives.SessionID          //Session Id agreed for this offchain transaction
	mscBaseState  primitives.MSCBaseStateSigned //MSContract Base state to use for state register
	vpcStatesList []primitives.VPCStateSigned   //List of all vpc state

	access sync.Mutex //Access control when setting connection status

}

// Connected returns if the channel connection is currently active.
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
// Sender is the one who initialized the channel connection.
func (inst *Instance) SenderID() identity.OffChainID {
	switch inst.roleChannel {
	case primitives.Sender:
		return inst.selfID
	case primitives.Receiver:
		return inst.peerID
	default:
		return identity.OffChainID{}
	}
}

// ReceiverID returns the id of receiver in the channel.
// Receiver is the one who received a new channel connection request and accepted it.
func (inst *Instance) ReceiverID() identity.OffChainID {
	switch inst.roleChannel {
	case primitives.Receiver:
		return inst.selfID
	case primitives.Sender:
		return inst.peerID
	default:
		return identity.OffChainID{}
	}
}

// SetRoleChannel sets the role of the self user in the channel.
func (inst *Instance) SetRoleChannel(role primitives.Role) {
	if role == primitives.Sender || role == primitives.Receiver {
		inst.roleChannel = role
	}
}

// RoleChannel returns the role of the self user in the channel.
func (inst *Instance) RoleChannel() primitives.Role {
	return inst.roleChannel
}

// SetRoleClosing sets the role of the self user in the channel closing procedure.
// If this user initializes the closing procedure, role is sender else it is receiver.
func (inst *Instance) SetRoleClosing(role primitives.Role) {
	if role == primitives.Sender || role == primitives.Receiver {
		inst.roleClosing = role
	}
}

// RoleClosing returns the role of the self user in the channel closing procedure.
// If this user initializes the closing procedure, role is sender else it is receiver.
func (inst *Instance) RoleClosing() primitives.Role {
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
func (inst *Instance) SetSessionID(sessionID primitives.SessionID) (err error) {
	isValid, err := sessionID.Validate()
	if !isValid {
		return fmt.Errorf("Session id invalid - %v", err.Error())
	}
	inst.sessionID = sessionID
	return nil
}

// SessionID returns the session id of the channel.
func (inst *Instance) SessionID() primitives.SessionID {
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
func (inst *Instance) SetMSCBaseState(newState primitives.MSCBaseStateSigned) (err error) {

	//Validate integrity of the sender signature on the state
	isValidSender, err := newState.VerifySign(inst.SenderID(), primitives.Sender)
	if err != nil {
		return err
	}
	if !isValidSender {
		return fmt.Errorf("Sender signature on MSCBaseState invalid")
	}

	//Validate integrity of the receiver signature on the state
	isValidReceiver, err := newState.VerifySign(inst.ReceiverID(), primitives.Receiver)
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
func (inst *Instance) MscBaseState() primitives.MSCBaseStateSigned {
	return inst.mscBaseState
}

// ValidateIncomingState validates the integrity of incoming state and if unsuccessful, returns the reason.
// Only version number and peer signature are validated.
func (inst *Instance) ValidateIncomingState(newState primitives.VPCStateSigned) (isValid bool, reason string) {

	var peerRole primitives.Role

	if inst.RoleChannel() == primitives.Sender {
		peerRole = primitives.Receiver
	} else {
		peerRole = primitives.Sender
	}

	//Validate integrity of the peer signature on the state
	isValidPeer, err := newState.VerifySign(inst.PeerID(), peerRole)

	if err != nil {
		return false, err.Error()
	}
	if !isValidPeer {
		return false, "Invalid peer signature"
	}

	//when previous state exists, check if the current version number is greater than previous
	lastVpcStateIndex := len(inst.vpcStatesList) - 1
	if lastVpcStateIndex != -1 {
		lastValidStateVersion := inst.vpcStatesList[lastVpcStateIndex].VPCState.Version
		if newState.VPCState.Version.Cmp(lastValidStateVersion) != 1 {
			return false, fmt.Sprintf("Current Version number (%s) less than previous (%s)", newState.VPCState.Version.String(), lastValidStateVersion.String())
		}
	}

	return true, ""
}

// ValidateFullState validates the integrity of newState and if unsuccessful, returns the reason.
// Version number, self and peer signatures are validated.
func (inst *Instance) ValidateFullState(newState primitives.VPCStateSigned) (isValid bool, reason string) {

	//Validate integrity of the sender signature on the state
	isValidSender, err := newState.VerifySign(inst.SenderID(), primitives.Sender)
	if err != nil {
		return false, "Invalid sender signature - " + err.Error()
	}
	if !isValidSender {
		return false, "Invalid sender signature"
	}

	//Validate integrity of the receiver signature on the state
	isValidReceiver, err := newState.VerifySign(inst.ReceiverID(), primitives.Receiver)
	if err != nil {
		return false, "Invalid receiver signature - " + err.Error()
	}
	if !isValidReceiver {
		return false, "Invalid receiver signature"
	}

	//when previous state exists, check if the current version number is greater than previous
	lastVpcStateIndex := len(inst.vpcStatesList) - 1
	if lastVpcStateIndex != -1 {
		lastValidStateVersion := inst.vpcStatesList[lastVpcStateIndex].VPCState.Version
		if newState.VPCState.Version.Cmp(lastValidStateVersion) != 1 {
			return false, fmt.Sprintf("Current Version number (%s) less than previous (%s)", newState.VPCState.Version.String(), lastValidStateVersion.String())
		}
	}

	return true, ""
}

// SetCurrentVPCState adds newState to vpc state list of the channel.
// Validation of the state concerning the application logic should be done before adding signatures.
func (inst *Instance) SetCurrentVPCState(newState primitives.VPCStateSigned) (err error) {

	isValid, reason := inst.ValidateFullState(newState)
	if !isValid {
		return fmt.Errorf("New state is invalid - %s", reason)
	}
	inst.vpcStatesList = append(inst.vpcStatesList, newState)
	logger.Debug("New MSC base state set")
	return nil
}

// CurrentVpcState returns the current vpc state of the channel.
func (inst *Instance) CurrentVpcState() primitives.VPCStateSigned {
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
