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

import "./Channel.sol";

/**
 * @title The App interface
 * @author The Perun Authors
 * @dev Every App that should be played in a state channel needs to implement this interface.
 */
interface App {
    /**
     * @notice ValidTransition checks if there was a valid transition between two states.
     * @dev ValidTransition should revert on an invalid transition.
     * Only App specific checks should be performed.
     * The adjudicator already checks the following:
     * - state corresponds to the params
     * - correct dimensions of the allocation
     * - preservation of balances
     * - params.participants[actorIdx] signed the to state
     * @param params The parameters of the channel.
     * @param from The current state.
     * @param to The potenrial next state.
     * @param actorIdx Index of the actor who signed this transition.
     */
    function validTransition(
        Channel.Params calldata params,
        Channel.State calldata from,
        Channel.State calldata to,
        uint256 actorIdx
    ) external pure;
}
