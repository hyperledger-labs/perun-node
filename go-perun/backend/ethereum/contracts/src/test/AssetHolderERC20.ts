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
const truffleAssert = require('truffle-assertions');
import { AssetHolderERC20Contract, PerunTokenContract, PerunTokenInstance } from "../../types/truffle-contracts";
import { ether } from "../lib/web3";
import { AssetHolderSetup } from "./Setup";
import { genericAssetHolderTest } from "./AssetHolder";

const AssetHolderERC20 = artifacts.require<AssetHolderERC20Contract>("AssetHolderERC20");
const PerunToken = artifacts.require<PerunTokenContract>("PerunToken");

contract("AssetHolderERC20", (accounts: any) => {
  let token: PerunTokenInstance;
  // Pass `undefined` as AssetHolder and set it in the deploy step.
  // Needed because of how mocha binds variables.
  let setup: AssetHolderSetup = new AssetHolderSetup(undefined, accounts, deposit, balanceOf);

  it("should deploy the PerunToken contract", async () => {
    token = await PerunToken.new(accounts, ether(100));
  });

  it("should deploy the AssetHolderERC20 contract", async () => {
    setup.ah = await AssetHolderERC20.new(setup.adj, token.address);
    let adjAddr = await setup.ah.adjudicator();
    adjAddr.should.equal(setup.adj);
  });

  async function deposit(fid: string, amount: BN, from: string) {
    truffleAssert.eventEmitted(
      await token.approve(setup.ah.address, amount, { from: from }),
      'Approval',
      (ev: any) => {
        return ev.owner == from && ev.spender == setup.ah.address && ev.value.eq(amount);
      }
    );
    // Do not set the value for token deposits.
    return setup.ah.deposit(fid, amount, { from: from })
  }

  async function balanceOf(who: string): Promise<BN> {
    return token.balanceOf.call(who);
  }

  genericAssetHolderTest(setup);
})
