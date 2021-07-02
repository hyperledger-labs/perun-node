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

/// <reference types="truffle-typings" />
import Web3 from "web3";
declare const web3: Web3;
import { hash, asyncWeb3Send } from "./web3";

export function sleep(milliseconds: any) {
  return new Promise(resolve => setTimeout(resolve, milliseconds));
}

export async function advanceBlockTime(time: number): Promise<any> {
  await asyncWeb3Send('evm_increaseTime', [time]);
  return asyncWeb3Send('evm_mine', []);
}

export function fundingID(channelID: string, participant: string): string {
  return hash(web3.eth.abi.encodeParameters(
    ['bytes32', 'address'],
    [web3.utils.rightPad(channelID, 64, "0"),
      participant]));
}

// describe test suite followed by blockchain revert
export function describeWithBlockRevert(name: string, tests: any) {
  describe(name, () => {
    let snapshot_id: number;

    before("take snapshot before first test", async () => {
      snapshot_id = (await asyncWeb3Send('evm_snapshot', [])).result;
    });

    after("restore snapshot after last test", async () => {
      return asyncWeb3Send('evm_revert', [snapshot_id]);
    });

    tests();
  });
}

// it test followed by blockchain revert
export function itWithBlockRevert(name: string, test: any) {
  it(name, async () => {
    let snapshot_id = (await asyncWeb3Send('evm_snapshot', [])).result;
    await test();
    await asyncWeb3Send('evm_revert', [snapshot_id]);
  });
}
