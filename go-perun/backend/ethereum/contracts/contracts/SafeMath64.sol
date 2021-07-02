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

library SafeMath64 {
    /**
     * @dev Function `add` returns the sum of `x` and `y` if less than or equal
     * to the maximum of type uint64. Otherwise, the function reverts.
     */
    function add(uint64 x, uint64 y) internal pure returns (uint64 z) {
        require((z = x + y) >= x, "overflow");
    }
}
