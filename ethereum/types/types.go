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

package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

// Hash is byte array (size 32) representation  of Keccak256 hash over any arbitrary data.
type Hash struct {
	common.Hash
}

// Address is byte array (size 20) representation of ethereum onchain identity.
type Address struct {
	common.Address
}

// HexToAddress returns the address derived from byte slice represented by the hex string.
// If length of the string is larger than required size, then it will be cropped from left
func HexToAddress(addr string) Address {
	return Address{common.HexToAddress(addr)}
}

// HexToHash returns the Hash derived from byte slice represented by the hex string.
// If length of the string is larger than required size, then it will be cropped from left
func HexToHash(hash string) Hash {
	return Hash{common.HexToHash(hash)}
}

// Hex2Bytes returns the bytes slice corresponding to the hexadecimal string.
func Hex2Bytes(str string) []byte {
	return common.Hex2Bytes(str)
}

// WeiToEther converts the value from Wei to Ether
func WeiToEther(val *big.Int) *big.Int {
	return new(big.Int).Div(val, big.NewInt(params.Ether))
}

// WeiToGwei converts the value from Wei to Gwei
func WeiToGwei(val *big.Int) *big.Int {
	return new(big.Int).Div(val, big.NewInt(params.GWei))
}

// EtherToWei converts the value from Ether to Wei
func EtherToWei(val *big.Int) *big.Int {
	return new(big.Int).Mul(val, big.NewInt(params.Ether))
}

// GweiToWei converts the value from Gwei to Wei
func GweiToWei(val *big.Int) *big.Int {
	return new(big.Int).Mul(val, big.NewInt(params.GWei))
}
