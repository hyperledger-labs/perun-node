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

package identity

import (
	"bytes"
	"fmt"

	"github.com/direct-state-transfer/dst-go/ethereum/keystore"
)

// SignHashWithPasswordEth signs the hash as per ethereum specifications if credentials set in the id are correct.
// Signature will in [R | S | V ] format with last byte V = 27/28.
func SignHashWithPasswordEth(idWithCredentials OffChainID, hash []byte) (signature []byte, err error) {

	hash = RehashWithEthereumPrefix(hash)
	signature, err = SignHashWithPassword(idWithCredentials, hash)
	if err == nil {
		signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	}
	logger.Debug("Message is signed")
	return signature, err
}

// SignHashWithPassword signs the hash as per ecdsa specifications if credentials set in the id are correct.
// Signature will in [R | S | V ] format with last byte V = 0/1.
func SignHashWithPassword(idWithCredentials OffChainID, hash []byte) (signature []byte, err error) {

	ks, password, isSetCredentials := idWithCredentials.GetCredentials()

	if !isSetCredentials {
		return nil, fmt.Errorf("credentials not set in identity")
	}
	defer idWithCredentials.ClearCredentials()

	onChainID := idWithCredentials.OnChainID

	userKeystoreAccount, err := ks.Find(keystore.MakeAccount(onChainID))
	if err != nil {
		return nil, err
	}

	signOnMsg, err := ks.SignHashWithPassphrase(userKeystoreAccount, password, hash)
	if err != nil {
		return nil, err
	}

	return signOnMsg, nil
}

// VerifySignatureEth checks if the given ethereum address created the ethereum signature over hash.
// The signature should be of size of 65 byte and in in [R | S | V] format with V = 27/28.
// The hash should have been created as per ethereum specifications.
func VerifySignatureEth(hash, sign, ethAddr []byte) (isSuccess bool, err error) {

	if len(sign) != 65 {
		return false, fmt.Errorf("invalid Ethereum signature (length is not 65 bytes)")
	}

	if sign[64] != 27 && sign[64] != 28 {
		return false, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sign[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	hash = RehashWithEthereumPrefix(hash)
	logger.Debug("Signature Verified")
	return VerifySignature(hash, sign, ethAddr)
}

// VerifySignature checks if the given ethereum address created the ecdsa signature over hash.
// The signature should be of size of 65 byte (in [R | S | V] format with V = 0/1).
func VerifySignature(hash, sign, ethAddr []byte) (isSuccess bool, err error) {

	//Retrieve public key from signature. Derive ethereum address from the public key
	//and check if the address matches that of the expected signer
	pubKey, err := keystore.SigToPub(hash, sign)
	if err != nil {
		return false, err
	}

	signerAddr := keystore.PubkeyToAddress(*pubKey)
	isSignerCorrect := bytes.Equal(signerAddr.Bytes(), ethAddr)
	if !isSignerCorrect {
		return isSignerCorrect, nil
	}

	//Retrieve uncompressed public key and validate integrity of the signature
	uncompressedPubKey := keystore.FromECDSAPub(pubKey)
	isHashCorrect := keystore.VerifySignature(uncompressedPubKey, hash, sign[:len(sign)-1])

	return isHashCorrect, nil
}

// RehashWithEthereumPrefix will prepend the original hash with ethereum specific message and rehash it.
// This is required to create ethereum compatible signatures.
func RehashWithEthereumPrefix(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	logger.Debug("Rehashing of signature done")
	return keystore.Keccak256([]byte(msg))
}
