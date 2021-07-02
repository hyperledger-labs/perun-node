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
pragma experimental ABIEncoderV2;

import "../vendor/openzeppelin-contracts/contracts/math/SafeMath.sol";
import "./AssetHolder.sol";

/**
 * @title The Perun AssetHolderETH
 * @notice AssetHolderETH is a concrete implementation of the abstract
 * AssetHolder which holds ETH.
 */
contract AssetHolderETH is AssetHolder {
    using SafeMath for uint256;

    /**
     * @notice Sets the adjudicator contract by calling the constructor of the
     * base asset holder contract.
     * @param _adjudicator Address of the adjudicator contract.
     */
    constructor(address _adjudicator) AssetHolder(_adjudicator) 
    {} // solhint-disable-line no-empty-blocks

    /**
     * @notice Should not be called directly but only by the parent AssetHolder.
     * @dev Checks that `msg.value` is equal to `amount`.
     */
    function depositCheck(bytes32, uint256 amount) internal override view {
        require(msg.value == amount, "wrong amount of ETH for deposit");
    }

    /**
     * @notice Should not be called directly but only by the parent AssetHolder.
     * @dev Withdraws ethereum for channel participant authorization.participant
     * to authorization.receiver.
     * @param authorization Withdrawal Authorization to authorize token transer
     * from a channel participant to an on-chain receiver.
     */
    function withdrawEnact(WithdrawalAuth calldata authorization, bytes calldata) internal override {
        authorization.receiver.transfer(authorization.amount);
    }
}
