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

package keystore

import (
	"crypto/ecdsa"

	"github.com/direct-state-transfer/dst-go/ethereum/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// CryptoWrapperInterface defines signatures for cryptography related function that will be used to handle ethereum keys.
type CryptoWrapperInterface interface {
	SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error)
	PubkeyToAddress(p ecdsa.PublicKey) common.Address
	FromECDSAPub(pub *ecdsa.PublicKey) []byte
	VerifySignature(pubkey, hash, signature []byte) bool
	Keccak256(data ...[]byte) []byte
}

// EthereumCryptoWrapper wraps the ethereum specific crypto related functions.
type EthereumCryptoWrapper struct {
}

// CryptoWrapperInstance defines the crypto wrapper instance that can be used by other modules.
var CryptoWrapperInstance CryptoWrapperInterface = EthereumCryptoWrapper{}

// SigToPub wraps the SigToPub function from ethereum/crypto package.
func (wrapper EthereumCryptoWrapper) SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	return crypto.SigToPub(hash, sig)
}

// PubkeyToAddress wraps the PubkeyToAddress function from ethereum/crypto package.
func (wrapper EthereumCryptoWrapper) PubkeyToAddress(p ecdsa.PublicKey) common.Address {
	return crypto.PubkeyToAddress(p)
}

// FromECDSAPub wraps the FromECDSAPub function from ethereum/crypto package.
func (wrapper EthereumCryptoWrapper) FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	return crypto.FromECDSAPub(pub)
}

// VerifySignature wraps the VerifySignature function from ethereum/crypto package.
func (wrapper EthereumCryptoWrapper) VerifySignature(pubkey, hash, signature []byte) bool {
	return crypto.VerifySignature(pubkey, hash, signature)
}

// Keccak256 wraps the Keccak256 function from ethereum/crypto package.
func (wrapper EthereumCryptoWrapper) Keccak256(data ...[]byte) []byte {
	return crypto.Keccak256(data...)
}

// SigToPub  retrieves and returns the public key that created the given signature.
func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	return CryptoWrapperInstance.SigToPub(hash, sig)
}

// PubkeyToAddress derives and returns the ethereum specific onchain identity from the ecdsa the public key.
func PubkeyToAddress(p ecdsa.PublicKey) types.Address {
	ethAddr := CryptoWrapperInstance.PubkeyToAddress(p)
	return types.Address{Address: ethAddr}
}

// FromECDSAPub returns the byte representation of the ecdsa public key.
func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	return CryptoWrapperInstance.FromECDSAPub(pub)
}

// VerifySignature checks if the given ethereum address created the ecdsa signature over hash.
// The signature should be of size of 65 byte (in [R | S | V] format with V = 0/1).
func VerifySignature(pubkey, hash, signature []byte) bool {
	return CryptoWrapperInstance.VerifySignature(pubkey, hash, signature)
}

// Keccak256 computes and returns the Keccak256 hash over the data.
func Keccak256(data ...[]byte) []byte {
	return CryptoWrapperInstance.Keccak256(data...)
}
