// Copyright 2019 - See NOTICE file for copyright holders.
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

// SPDX-License-Identifier: Apache-2.0

pragma solidity ^0.7.0;

import "../vendor/openzeppelin-contracts/contracts/cryptography/ECDSA.sol";

// Sig is a library to verify signatures.
library Sig {
    // Verify verifies whether a piece of data was signed correctly.
    function verify(bytes memory data, bytes memory signature, address signer) internal pure returns (bool) {
        bytes32 prefixedHash = ECDSA.toEthSignedMessageHash(keccak256(data));
        address recoveredAddr = ECDSA.recover(prefixedHash, signature);
        return recoveredAddr == signer;
    }
}
