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
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"

	solsha3 "github.com/miguelmota/go-solidity-sha3"

	"github.com/direct-state-transfer/dst-go/ethereum/keystore"
	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/direct-state-transfer/dst-go/identity"
)

// Size of nonce (in bytes) for session id generation.
var sessionIDNonceSize = int(32)

// SessionID represents the unique identification of offchain channel.
type SessionID struct {

	//SidComplete is *big.Int is to create compatible solidiySHA3 hash
	SidComplete     *big.Int `json:"sid_complete"`
	SidSenderPart   []byte   `json:"sid_sender_part"`
	SidReceiverPart []byte   `json:"sid_receiver_part"`

	AddrSender    types.Address `json:"addr_sender"`
	AddrReceiver  types.Address `json:"addr_receiver"`
	NonceSender   []byte        `json:"nonce_sender"`
	NonceReceiver []byte        `json:"nonce_receiver"`

	//SessionId is locked for further changes
	Locked bool `json:"locked"`
}

// NewSessionID returns a initialized session id.
func NewSessionID(AddrSender, AddrReceiver types.Address) SessionID {
	return SessionID{
		AddrSender:   AddrSender,
		AddrReceiver: AddrReceiver,
		Locked:       false,
	}
}

// Equal returns true if two session ids are equal.
func (sid *SessionID) Equal(b SessionID) bool {
	if sid.EqualSender(b) && sid.EqualReceiver(b) {

		if sid.SidComplete != nil && b.SidComplete != nil {
			if sid.SidComplete.Cmp(b.SidComplete) == 0 {
				return true
			}
		} else if sid.SidComplete == nil && b.SidComplete == nil {
			return true
		}
	}
	return false
}

// EqualSender returns true true if the sender parts of two session ids are equal.
func (sid *SessionID) EqualSender(b SessionID) bool {
	if bytes.Equal(sid.AddrSender.Bytes(), b.AddrSender.Bytes()) &&
		bytes.Equal(sid.NonceSender, b.NonceSender) &&
		bytes.Equal(sid.SidSenderPart, b.SidSenderPart) {
		return true
	}
	return false
}

// EqualReceiver returns true true if the receiver parts of two session ids are equal.
func (sid *SessionID) EqualReceiver(b SessionID) bool {
	if bytes.Equal(sid.AddrReceiver.Bytes(), b.AddrReceiver.Bytes()) &&
		bytes.Equal(sid.NonceReceiver, b.NonceReceiver) &&
		bytes.Equal(sid.SidReceiverPart, b.SidReceiverPart) {
		return true
	}
	return false
}

// GenerateSenderPart generates sid sender part = hash(nonce + addrSender).
func (sid *SessionID) GenerateSenderPart(AddrSender types.Address) (err error) {

	if sid.Locked {
		return fmt.Errorf("SessionId is locked for further changes")
	}

	nonce, err := GenerateRandomNumber(sessionIDNonceSize)
	if err != nil {
		return err
	}
	sid.AddrSender = AddrSender
	sid.NonceSender = nonce

	sid.SidSenderPart = keystore.Keccak256(append(nonce, sid.AddrSender.Bytes()...))
	return nil
}

// GenerateReceiverPart generates sid receiver part = hash(nonce + addrReceiver).
func (sid *SessionID) GenerateReceiverPart(AddrReceiver types.Address) (err error) {

	if sid.Locked {
		return fmt.Errorf("SessionId is locked for further changes")
	}

	nonce, err := GenerateRandomNumber(sessionIDNonceSize)
	if err != nil {
		return err
	}
	sid.AddrReceiver = AddrReceiver
	sid.NonceReceiver = nonce

	sid.SidReceiverPart = keystore.Keccak256(append(nonce, sid.AddrReceiver.Bytes()...))
	return nil
}

// GenerateCompleteSid generates sid = hash(sender part + receiver part).
func (sid *SessionID) GenerateCompleteSid() (err error) {

	if sid.Locked {
		return fmt.Errorf("SessionId is locked for further changes")
	}

	if sid.SidSenderPart == nil {
		return fmt.Errorf("SidSenderPart is nil")
	}
	if sid.SidReceiverPart == nil {
		return fmt.Errorf("SidReceiverPart is nil")
	}

	sidCompleteBytes := keystore.Keccak256(append(sid.SidSenderPart, sid.SidReceiverPart...))

	sid.SidComplete = big.NewInt(0).SetBytes(sidCompleteBytes)

	sid.Locked = true
	logger.Info("Complete Session ID generated successfully")
	return nil
}

// Validate checks presence of lock and validity of senderPart, receiverPart and complete sid.
func (sid *SessionID) Validate() (valid bool, err error) {

	if !sid.Locked {
		return false, fmt.Errorf("SessionId is not locked")
	}

	if sid.SidSenderPart == nil {
		return false, fmt.Errorf("SidSenderPart is nil")
	}
	if sid.SidReceiverPart == nil {
		return false, fmt.Errorf("SidReceiverPart is nil")
	}

	//validate senderPart
	sidSenderPart := keystore.Keccak256(append(sid.NonceSender, sid.AddrSender.Bytes()...))
	if 0 != bytes.Compare(sidSenderPart, sid.SidSenderPart) {
		return false, fmt.Errorf("SidSenderPart invalid")
	}

	//validate receiverPart
	sidReceiverPart := keystore.Keccak256(append(sid.NonceReceiver, sid.AddrReceiver.Bytes()...))
	if 0 != bytes.Compare(sidReceiverPart, sid.SidReceiverPart) {
		return false, fmt.Errorf("SidReceiverPart invalid")
	}

	//validate sid
	sidCompleteBytes := keystore.Keccak256(append(sidSenderPart, sidReceiverPart...))
	sidComplete := big.NewInt(0).SetBytes(sidCompleteBytes)

	if sid.SidComplete == nil || 0 != sid.SidComplete.Cmp(sidComplete) {
		return false, fmt.Errorf("SidComplete invalid")
	}

	return true, nil
}

// SoliditySHA3 provides hash of sid as will be returned in VPCClosing and VPCClosed events.
func (sid *SessionID) SoliditySHA3() []byte {
	return solsha3.SoliditySHA3(
		solsha3.Int256(sid.SidComplete),
		solsha3.Address(sid.AddrSender.Address),
		solsha3.Address(sid.AddrReceiver.Address),
	)
}

// MSCBaseState is the structure of base state in mscontract.
// Data types of state members should match with those defined in solidity code.
type MSCBaseState struct {
	VpcAddress      types.Address `json:"vpc_address"`
	Sid             *big.Int      `json:"sid"`
	BlockedSender   *big.Int      `json:"blocked_sender"`
	BlockedReceiver *big.Int      `json:"blocked_receiver"`
	Version         *big.Int      `json:"version"`
}

// String implements fmt.Stringer interface.
// It prints msc base state struct with vpc address in hexadecimal form.
func (state MSCBaseState) String() string {
	return fmt.Sprintf("{VpcAddress:0x%x Sid:0x%x BlockedSender:%s BlockedReceiver:%s Version:%s}",
		state.VpcAddress.Bytes(), state.Sid.Bytes(), state.BlockedSender, state.BlockedReceiver, state.Version)
}

// Equal returns true if two MSContractBaseStates are equal.
func (state MSCBaseState) Equal(b MSCBaseState) (result bool) {
	if bytes.Equal(state.VpcAddress.Bytes(), b.VpcAddress.Bytes()) &&
		state.BlockedReceiver.Cmp(b.BlockedReceiver) == 0 &&
		state.BlockedSender.Cmp(b.BlockedSender) == 0 &&
		state.Sid.Cmp(b.Sid) == 0 &&
		state.Version.Cmp(b.Version) == 0 {
		return true
	}
	return false
}

// SoliditySHA3 generates solidity compatible sha3 hash over MSContractBaseState.
func (state MSCBaseState) SoliditySHA3() []byte {
	return solsha3.SoliditySHA3(
		solsha3.Address(state.VpcAddress.Address),
		solsha3.Int256(state.Sid),
		solsha3.Int256(state.BlockedSender),
		solsha3.Int256(state.BlockedReceiver),
		solsha3.Int256(state.Version),
	)
}

// MSCBaseStateSigned is MSContractBaseState with signatures.
type MSCBaseStateSigned struct {
	MSContractBaseState MSCBaseState `json:"ms_contract_state"`
	SignSender          []byte       `json:"sign_sender"`
	SignReceiver        []byte       `json:"sign_receiver"`
}

// String implements fmt.Stringer interface.
// It prints msc base state signed struct with signatures in hexadecimal form.
func (state MSCBaseStateSigned) String() string {
	return fmt.Sprintf("{MSCBaseState:%+v SignSender:0x%x SignReceiver:0x%x}",
		state.MSContractBaseState, state.SignSender, state.SignReceiver)
}

// Equal returns true if the two MSContractBaseStateSigned are equal.
func (state *MSCBaseStateSigned) Equal(b MSCBaseStateSigned) (result bool) {
	if state.MSContractBaseState.Equal(b.MSContractBaseState) &&
		bytes.Equal(state.SignSender, b.SignSender) &&
		bytes.Equal(state.SignReceiver, b.SignReceiver) {
		return true
	}
	return false
}

// AddSign adds signature over MSContractBaseState for the defined role using idWithCreds as signer.
func (state *MSCBaseStateSigned) AddSign(idWithCreds identity.OffChainID, role Role) (err error) {

	if role != Sender && role != Receiver {
		return fmt.Errorf("Invalid role")
	}

	hash := state.MSContractBaseState.SoliditySHA3()
	sign, err := identity.SignHashWithPasswordEth(idWithCreds, hash)
	if err != nil {
		return err
	}

	switch role {
	case Sender:
		state.SignSender = sign
	case Receiver:
		state.SignReceiver = sign
	}

	return nil
}

// VerifySign verifies the signature of user corresponding to defined role over MSContractBaseState.
func (state *MSCBaseStateSigned) VerifySign(id identity.OffChainID, role Role) (isValid bool, err error) {

	var signToValidate []byte

	switch role {
	case Sender:
		signToValidate = state.SignSender
	case Receiver:
		signToValidate = state.SignReceiver
	default:
		return false, fmt.Errorf("Invalid role")
	}

	hash := state.MSContractBaseState.SoliditySHA3()
	isValid, err = identity.VerifySignatureEth(hash, signToValidate, id.OnChainID.Bytes())
	if err != nil {
		return false, err
	}
	return isValid, nil
}

// VPCStateID is the structure of id vpc state.
type VPCStateID struct {
	AddSender    types.Address
	AddrReceiver types.Address
	SID          *big.Int
}

// SoliditySHA3 generates solidity compatible sha3 hash over VPCStateId.
func (stateID *VPCStateID) SoliditySHA3() []byte {
	return solsha3.SoliditySHA3(
		solsha3.Address(stateID.AddSender.Address),
		solsha3.Address(stateID.AddrReceiver.Address),
		solsha3.Int256(stateID.SID),
	)
}

// VPCState is the structure of state in vpc and hence in off-chain channel.
// Data types of state members should match with those defined in solidity code.
type VPCState struct {
	ID              []byte   `json:"id"`
	Version         *big.Int `json:"version"`
	BlockedSender   *big.Int `json:"blocked_alice"`
	BlockedReceiver *big.Int `json:"blocked_bob"`
}

// String implements fmt.Stringer.
// It prints vpcState struct with id in hexadecimal form.
func (state VPCState) String() string {
	return fmt.Sprintf("{Id:0x%x Version:%s BlockedSender:%s BlockedReceiver:%s}",
		state.ID, state.Version.String(), state.BlockedSender, state.BlockedReceiver)
}

// SoliditySHA3 generates solidity compatible sha3 hash over VPCState.
func (state VPCState) SoliditySHA3() []byte {
	return solsha3.SoliditySHA3(
		solsha3.Bytes32(state.ID),
		solsha3.Int256(state.Version),
		solsha3.Int256(state.BlockedSender),
		solsha3.Int256(state.BlockedReceiver),
	)
}

// Equal returns true if the two VPCStates are equal.
func (state VPCState) Equal(b VPCState) (result bool) {
	if bytes.Equal(state.ID, b.ID) &&
		state.Version.Cmp(b.Version) == 0 &&
		state.BlockedReceiver.Cmp(b.BlockedReceiver) == 0 &&
		state.BlockedSender.Cmp(b.BlockedSender) == 0 {
		return true
	}
	return false
}

// VPCStateSigned is VPCState with signatures.
type VPCStateSigned struct {
	VPCState     VPCState `json:"vpc_state"`
	SignSender   []byte   `json:"sign_sender"`
	SignReceiver []byte   `json:"sign_receiver"`
}

// String implements fmt.Stringer interface. It prints vpc state signed struct with signatures in hexadecimal form.
func (state VPCStateSigned) String() string {
	return fmt.Sprintf("{VPCState:%+v SignSender:0x%x SignReceiver:0x%x}",
		state.VPCState, state.SignSender, state.SignReceiver)
}

// Equal returns true if the two VPCStateSigned are equal.
func (state *VPCStateSigned) Equal(b VPCStateSigned) (result bool) {
	if state.VPCState.Equal(b.VPCState) &&
		bytes.Equal(state.SignSender, b.SignSender) &&
		bytes.Equal(state.SignReceiver, b.SignReceiver) {
		return true
	}
	return false
}

// AddSign adds signature over VPCState for the defined role using idWithCreds as signer.
func (state *VPCStateSigned) AddSign(idWithCreds identity.OffChainID, role Role) (err error) {

	if role != Sender && role != Receiver {
		return fmt.Errorf("Invalid role")
	}

	hash := state.VPCState.SoliditySHA3()
	sign, err := identity.SignHashWithPasswordEth(idWithCreds, hash)
	if err != nil {
		return err
	}

	switch role {
	case Sender:
		state.SignSender = sign
	case Receiver:
		state.SignReceiver = sign
	}

	return nil
}

// VerifySign verifies the signature of user corresponding to defined role over VPCState.
func (state *VPCStateSigned) VerifySign(id identity.OffChainID, role Role) (isValid bool, err error) {

	var signToValidate []byte

	switch role {
	case Sender:
		signToValidate = state.SignSender
	case Receiver:
		signToValidate = state.SignReceiver
	default:
		return false, fmt.Errorf("Invalid role")
	}

	hash := state.VPCState.SoliditySHA3()
	isValid, err = identity.VerifySignatureEth(hash, signToValidate, id.OnChainID.Bytes())
	if err != nil {
		return false, err
	}
	return isValid, nil
}

// GenerateRandomNumber generates a random byte array of given size.
func GenerateRandomNumber(sizeInBytes int) ([]byte, error) {
	randomnBytes := make([]byte, sizeInBytes)
	_, err := rand.Read(randomnBytes)
	if err != nil {
		return nil, err
	}

	return randomnBytes, nil
}
