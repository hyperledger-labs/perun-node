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
import "./Sig.sol";

/**
 * @title The Perun AssetHolder
 * @notice AssetHolder is an abstract contract that holds the funds for a
 * Perun state channel.
 */
abstract contract AssetHolder {
    using SafeMath for uint256;

    /**
     * @dev WithdrawalAuth authorizes an on-chain public key to withdraw from an ephemeral key.
     */
    struct WithdrawalAuth {
        bytes32 channelID;
        address participant; // The account used to sign the authorization which is debited.
        address payable receiver; // The receiver of the authorization.
        uint256 amount; // The amount that can be withdrawn.
    }

    event OutcomeSet(bytes32 indexed channelID);
    event Deposited(bytes32 indexed fundingID, uint256 amount);
    event Withdrawn(bytes32 indexed fundingID, uint256 amount, address receiver);

    /**
     * @notice This mapping stores the balances of participants to their fundingID.
     * @dev Mapping H(channelID||participant) => money
     */
    mapping(bytes32 => uint256) public holdings;

    /**
     * @notice This mapping stores whether a channel was already settled.
     * @dev Mapping channelID => settled
     */
    mapping(bytes32 => bool) public settled;

    /**
     * @notice Address of the adjudicator contract that can call setOutcome.
     * @dev Set by the constructor.
     */
    address public adjudicator;

    /**
     * @notice The onlyAdjudicator modifier specifies functions that can only be called from the adjudicator contract.
     */
    modifier onlyAdjudicator {
        require(msg.sender == adjudicator, "can only be called by the adjudicator"); // solhint-disable-line reason-string
        _;
    }

    /**
     * @notice Sets the adjudicator contract that is able to call setOutcome on this contract.
     * @param _adjudicator Address of the adjudicator contract.
     */
    constructor(address _adjudicator) {
        adjudicator = _adjudicator;
    }

    /**
     * @notice Sets the final outcome of a channel. Can only be called by the adjudicator.
     * @dev This method should not be overwritten by the implementing contract.
     * @param channelID ID of the channel that should be disbursed.
     * @param parts Array of participants of the channel.
     * @param newBals New Balances after execution of the channel.
     */
    function setOutcome(
        bytes32 channelID,
        address[] calldata parts,
        uint256[] calldata newBals)
    external onlyAdjudicator {
        require(parts.length == newBals.length, "participants length should equal balances"); // solhint-disable-line reason-string
        require(settled[channelID] == false, "trying to set already settled channel"); // solhint-disable-line reason-string

        // The channelID itself might already be funded
        uint256 sumHeld = holdings[channelID];
        holdings[channelID] = 0;
        uint256 sumOutcome = 0;

        bytes32[] memory fundingIDs = new bytes32[](parts.length);
        for (uint256 i = 0; i < parts.length; i++) {
            bytes32 id = calcFundingID(channelID, parts[i]);
            // Save calculated ids to save gas.
            fundingIDs[i] = id;
            // Compute old balances.
            sumHeld = sumHeld.add(holdings[id]);
            // Compute new balances.
            sumOutcome = sumOutcome.add(newBals[i]);
        }

        // We allow overfunding channels, who overfunds looses their funds.
        if (sumHeld >= sumOutcome) {
            for (uint256 i = 0; i < parts.length; i++) {
                holdings[fundingIDs[i]] = newBals[i];
            }
        }
        settled[channelID] = true;
        emit OutcomeSet(channelID);
    }

    /**
     * @notice Function that is used to fund a channel.
     * @dev Generic function which uses the virtual functions `depositCheck` and
     * `depositEnact` to execute the user specific code.
     * Requires that:
     *  - `depositCheck` does not revert
     *  - `depositEnact` does not revert
     * Increases the holdings for the participant.
     * Emits a `Deposited` event upon success.
     * @param fundingID Unique identifier for a participant in a channel.
     * Calculated as the hash of the channel id and the participant address.
     * @param amount Amount of money that should be deposited.
     */
    function deposit(bytes32 fundingID, uint256 amount) external payable {
        depositCheck(fundingID, amount);
        holdings[fundingID] = holdings[fundingID].add(amount);
        depositEnact(fundingID, amount);       
        emit Deposited(fundingID, amount);
    }

    /**
     * @notice Sends money from authorization.participant to authorization.receiver.
     * @dev Generic function which uses the virtual functions `withdrawCheck` and
     * `withdrawEnact` to execute the user specific code.
     * Requires that:
     *  - Channel is settled
     *  - Signature is valid
     *  - Enough holdings are available
     *  - `withdrawCheck` does not revert
     *  - `withdrawEnact` does not revert
     * Decreases the holdings for the participant.
     * Emits a `Withdrawn` event upon success.
     * @param authorization WithdrawalAuth that specifies which account receives
     * what amounf of asset from which channel participant.
     * @param signature Signature on the withdrawal authorization.
     */
    function withdraw(WithdrawalAuth calldata authorization, bytes calldata signature) external {
        require(settled[authorization.channelID], "channel not settled");
        require(Sig.verify(abi.encode(authorization), signature, authorization.participant), "signature verification failed");
        bytes32 id = calcFundingID(authorization.channelID, authorization.participant);
        require(holdings[id] >= authorization.amount, "insufficient ETH for withdrawal");
        withdrawCheck(authorization, signature);
        holdings[id] = holdings[id].sub(authorization.amount);
        withdrawEnact(authorization, signature);
        emit Withdrawn(id, authorization.amount, authorization.receiver);
    }

    /**
     * @notice Checks a deposit for validity and reverts otherwise.
     * @dev Should be overridden by all contracts that inherit it since it is
     * called by `deposit` before `depositEnact`.
     * This function is empty by default and the overrider does not need to
     * call it via `super`.
     */
    function depositCheck(bytes32 fundingID, uint256 amount) internal view virtual
    {} // solhint-disable no-empty-blocks

    /**
     * @notice Enacts a deposit or reverts otherwise.
     * @dev Should be overridden by all contracts that inherit it since it is
     * called by `deposit` after `depositCheck`.
     * This function is empty by default and the overrider does not need to
     * call it via `super`.
     */
    function depositEnact(bytes32 fundingID, uint256 amount) internal virtual
    {} // solhint-disable no-empty-blocks

    /**
     * @notice Checks a withdrawal for validity and reverts otherwise.
     * @dev Should be overridden by all contracts that inherit it since it is
     * called by `withdraw` before `withdrawEnact`.
     * This function is empty by default and the overrider does not need to
     * call it via `super`.
     */
    function withdrawCheck(WithdrawalAuth calldata authorization, bytes calldata signature) internal view virtual
    {} // solhint-disable no-empty-blocks

    /**
     * @notice Enacts a withdrawal or reverts otherwise.
     * @dev Should be overridden by all contracts that inherit it since it is
     * called by `withdraw` after `withdrawCheck`.
     * This function is empty by default and the overrider does not need to
     * call it via `super`.
     */
    function withdrawEnact(WithdrawalAuth calldata authorization, bytes calldata signature) internal virtual
    {} // solhint-disable no-empty-blocks

    /**
     * @notice Internal helper function that calculates the fundingID.
     * @param channelID ID of the channel.
     * @param participant Address of a participant in the channel.
     * @return The funding ID, an identifier used for indexing.
     */
    function calcFundingID(bytes32 channelID, address participant) internal pure returns (bytes32) {
        return keccak256(abi.encode(channelID, participant));
    }
}
