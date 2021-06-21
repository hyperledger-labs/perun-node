// Copyright (c) 2021 - for information on the respective copyright owner
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

package blockchain

import (
	"fmt"

	"github.com/pkg/errors"
)

// ContractName identifies specific type of contract.
type ContractName string

// Enumeration of contract names for perun node.
const (
	Adjudicator = "adjudicator"
	AssetETH    = "ETH"
)

// InvalidContractError indicates an error in validating the contract on the
// blockchain.
type InvalidContractError struct {
	Name    string
	Address string
	err     error
}

// Error implements error interface.
func (e InvalidContractError) Error() string {
	return fmt.Sprintf("invalid %s contract at address %s: %v", e.Name, e.Address, e.err)
}

// Unwrap returns the original error.
func (e InvalidContractError) Unwrap() error {
	return e.err
}

// NewInvalidContractError constructs and returns an InvalidContractError.
func NewInvalidContractError(name string, address string, err error) error {
	return errors.WithStack(InvalidContractError{
		Name:    name,
		Address: address,
		err:     err,
	})
}
