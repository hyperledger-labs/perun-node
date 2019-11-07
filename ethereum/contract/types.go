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

package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"time"
)

// MatchStatus represents the match result when contracts are compared.
type MatchStatus uint8

// Enumeration of allowed values for MatchStatus.
const (
	Match MatchStatus = iota
	NoMatch
	Missing
	Unknown
)

type fileType string

// Enumeration of different filetypes associated with a contract.
const (
	golangFile     fileType = ".go"
	solidityFile   fileType = ".sol"
	binRuntimeFile fileType = ".bin-runtime"
)

// Handler manages the files on disk corresponding to a contract, it's version and gas units required for deployment.
type Handler struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	GasUnits uint64 `json:"-"`

	HashSolFile        string `json:"hashSolFile"`
	HashGoFile         string `json:"-"`
	HashBinRuntimeFile string `json:"hashBinRuntimeFile"`
}

//Equal returns true if both the ContractHandler values are equal. HashGoFile and GasUnits fields are ignored though.
func (contractA Handler) Equal(contractB Handler) (match bool) {

	if (contractA.Name == contractB.Name) && (contractA.Version == contractB.Version) && (contractA.HashSolFile == contractB.HashSolFile) &&
		(contractA.HashBinRuntimeFile == contractB.HashBinRuntimeFile) {
		return true
	}
	return false
}

func (contractA Handler) getAbsFilepath(contractsDir string, fType fileType) (absFilePath string, err error) {

	//append file type to contract name in order to obtain filepath

	switch fType {
	case golangFile:
		absFilePath = contractA.Name + string(fType)
		absFilePath, err = filepath.Abs(absFilePath)
	case solidityFile, binRuntimeFile:
		absFilePath = filepath.Join(contractsDir, contractA.Name+string(fType))
		absFilePath, err = filepath.Abs(absFilePath)
	}

	return absFilePath, err
}

// StoreType represents the contract store with handlers for each contract, its timeout periods and storage directory on disk.
type StoreType struct {
	contractsDir string //contractsDir is the relative path of directory containing solidity contracts and runtime binaries

	libSignatures Handler
	msContract    Handler
	vpc           Handler

	timeoutMSContract          time.Duration //Timeperiod after which finalizeClose/finalizeRegister in MSContract can be called
	timeoutVPCValidity         time.Duration //Timeperiod before which other peer should respond to VPCClosing event from MSContract
	timeoutVPCExtendedValidity time.Duration //Timeperiod after which VPCFinalize in VPC can be called
}

// LibSignatures return the libsignatures contract handler from the store
func (cs StoreType) LibSignatures() Handler {
	return cs.libSignatures
}

// MSContract returns the ms contract handler from the store.
func (cs StoreType) MSContract() Handler {
	return cs.msContract
}

// VPC returns the vpc contract handler from the store.
func (cs StoreType) VPC() Handler {
	return cs.vpc
}

// TimeoutMSContract returns the ms contract timeout value.
func (cs StoreType) TimeoutMSContract() time.Duration {
	return cs.timeoutMSContract
}

// TimeoutVPCValidity returns the vpc validity timeout value.
func (cs StoreType) TimeoutVPCValidity() time.Duration {
	return cs.timeoutVPCValidity
}

// TimeoutVPCExtendedValidity returns the vpc extended timeout value.
func (cs StoreType) TimeoutVPCExtendedValidity() time.Duration {
	return cs.timeoutVPCExtendedValidity
}

// SHA256Sum calculates and returns SHA256 sum over the all hash of runtime binaries corresponding to each contract in contract store.
func (cs StoreType) SHA256Sum() []byte {

	//To generate SHA256Sum, concatenate hash of libsignatures binRuntime, vpc binRuntime and msc binRuntime,
	//Compute SHA256Sum over the resulting byte array.

	var concatenatedHash []byte
	hashByteSlice, _ := hex.DecodeString(cs.libSignatures.HashBinRuntimeFile)
	concatenatedHash = append(concatenatedHash, hashByteSlice...)
	hashByteSlice, _ = hex.DecodeString(cs.vpc.HashBinRuntimeFile)
	concatenatedHash = append(concatenatedHash, hashByteSlice...)
	hashByteSlice, _ = hex.DecodeString(cs.msContract.HashBinRuntimeFile)
	concatenatedHash = append(concatenatedHash, hashByteSlice...)

	sha256SumArray := sha256.Sum256(concatenatedHash)

	return sha256SumArray[:]
}
