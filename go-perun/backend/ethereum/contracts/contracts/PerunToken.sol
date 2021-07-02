// Copyright 2020 - See NOTICE file for copyright holders.
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
pragma experimental ABIEncoderV2;

import "../vendor/openzeppelin-contracts/contracts/utils/Address.sol";
import "../vendor/openzeppelin-contracts/contracts/token/ERC20/ERC20.sol";

contract PerunToken is ERC20 {
    using SafeMath for uint256;

    /**
     * @dev Creates a new PerunToken contract instance with `accounts` being
     * funded with `initBalance` tokens.
     */
    constructor (address[] memory accounts, uint256 initBalance) ERC20("PerunToken", "PRN") {
        for (uint256 i = 0; i < accounts.length; i++) {
            _mint(accounts[i], initBalance);
        }
    }
}
