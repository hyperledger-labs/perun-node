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

import { should } from "chai";
should();
import Web3 from "web3";
declare const web3: Web3;
import { AssetHolderETHContract } from "../../types/truffle-contracts";
import { AssetHolderSetup } from "./Setup";
import { genericAssetHolderTest } from "./AssetHolder";

const AssetHolderETH = artifacts.require<AssetHolderETHContract>("AssetHolderETH");

contract("AssetHolderETH", (accounts: any) => {
  // Pass `undefined` as AssetHolder and set it in the deploy step.
  // Needed because of how mocha binds variables.
  let setup: AssetHolderSetup = new AssetHolderSetup(undefined, accounts, deposit, balanceOf);

  it("should deploy the AssetHolderETH contract", async () => {
    setup.ah = await AssetHolderETH.new(setup.adj);
    const adjAddr = await setup.ah.adjudicator();
    adjAddr.should.equal(setup.adj);
  });

  function deposit(fid: string, amount: BN, from: string) {
    return setup.ah.deposit(fid, amount, { value: amount, from: from })
  }

  async function balanceOf(who: string): Promise<BN> {
    return web3.utils.toBN(await web3.eth.getBalance(who));
  }
  
  genericAssetHolderTest(setup);
})
