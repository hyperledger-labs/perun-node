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

import "../vendor/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";
import "../vendor/openzeppelin-contracts/contracts/math/SafeMath.sol";
import "./AssetHolder.sol";

/**
 * @title The Perun AssetHolderERC20
 * @notice AssetHolderERC20 is a concrete implementation of the abstract
 * AssetHolder which holds a specific ERC20 token.
 * @dev Before calling `deposit`, the allowance for the AssetHolder must be set to
 * at least the amount that should be deposited.
 */
contract AssetHolderERC20 is AssetHolder {
	using SafeMath for uint256;

	IERC20 public immutable token;

	constructor(address _adjudicator, address _token) AssetHolder(_adjudicator) {
		token = IERC20(_token);
	}

	/**
	 * @notice Used to check the validity of a deposit of tokens into a channel.
	 * @dev The sender has to set the allowance for the assetHolder to
	 * at least `amount`.
	 */
	function depositCheck(bytes32, uint256) internal view override {
		require(msg.value == 0, "message value must be 0 for token deposit"); // solhint-disable-line reason-string
	}

	/**
	 * @notice Should not be called directly but only by the parent AssetHolder.
	 * @dev Transferes `amount` tokens from `msg.sender` to `fundingID`.	
 	 */
	function depositEnact(bytes32, uint256 amount) internal override {
		require(token.transferFrom(msg.sender, address(this), amount), "transferFrom failed");
	}
	
	/**
     * @notice Should not be called directly but only by the parent AssetHolder.
     * @dev Withdraws tokens for channel participant authorization.participant
	 * to authorization.receiver.
     * @param authorization Withdrawal Authorization to authorize token transer
     * from a channel participant to an on-chain receiver.
     */
    function withdrawEnact(WithdrawalAuth calldata authorization, bytes calldata) internal override {
		require(token.transfer(authorization.receiver, authorization.amount), "transfer failed");
	}
}
