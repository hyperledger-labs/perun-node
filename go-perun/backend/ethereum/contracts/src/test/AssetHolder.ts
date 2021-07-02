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

import { assert, should } from "chai";
should();
const truffleAssert = require('truffle-assertions');
import { sign, ether, wei2eth } from "../lib/web3";
import { fundingID, describeWithBlockRevert } from "../lib/test";
import { AssetHolderSetup } from "./Setup";
import { Authorization } from "./Channel";

// All accounts in `setup` must have `ether(100)` worth of funds.
export function genericAssetHolderTest(setup: AssetHolderSetup) {
  const finalBalance = [ether(20), ether(10)];

  async function assertHoldings(fid: string, amount: BN) {
    const c = await setup.ah.holdings.call(fid);
    assert(amount.eq(c), `Wrong holdings. Wanted: ${wei2eth(amount)}, got: ${wei2eth(c)}`);
  }

  async function testDeposit(idx: number, amount: BN, cid: string) {
    const fid = fundingID(cid, setup.parts[idx]);
    const oldBal = await setup.ah.holdings.call(fid);
    truffleAssert.eventEmitted(
      await setup.deposit(fid, amount, setup.recv[idx]),
      'Deposited',
      (ev: any) => {
        return ev.fundingID == fid && ev.amount.eq(amount);
      }
    );
    await assertHoldings(fid, oldBal.add(amount));
  }

  async function testWithdraw(idx: number, amount: BN, cid: string) {
    const fid = fundingID(cid, setup.parts[idx]);
    let balanceBefore = await setup.balanceOf(setup.recv[idx]);
    let authorization = new Authorization(cid, setup.parts[idx], setup.recv[idx], amount.toString());
    let signature = await sign(authorization.encode(), setup.parts[idx]);
    truffleAssert.eventEmitted(
      await setup.ah.withdraw(authorization, signature, { from: setup.txSender }),
      'Withdrawn',
      (ev: any) => {
        return ev.fundingID == fid
          && amount.eq(ev.amount)
          && ev.receiver == setup.recv[idx];
      }
    );
    let balanceAfter = await setup.balanceOf(setup.recv[idx]);
    assert(amount.add(balanceBefore).eq(balanceAfter), "wrong receiver balance");
  }

  describe("Funding...", () => {
    it("A deposits eth", async () => {
      await testDeposit(setup.A, ether(9), setup.channelID);
    });

    it("B deposits eth", async () => {
      await testDeposit(setup.B, ether(20), setup.channelID);
    });

    it("wrong msg.value", async () => {
      let id = fundingID(setup.channelID, setup.parts[setup.A]);
      // AsserHolderETH   should revert for msg.value != amount and
      // AssetHolderToken should revert for msg.value != 0.
      await truffleAssert.reverts(
        setup.ah.deposit(id, ether(2), { value: ether(1), from: setup.parts[setup.A] })
      );
      await assertHoldings(id, ether(9));
    });

    it("A deposits eth", async () => {
      await testDeposit(setup.A, ether(1), setup.channelID);
    });
  })

  describe("Invalid withdraw", () => {
    it("unsettled channel should fail", async () => {
      assert(finalBalance.length == setup.parts.length);
      assert(await setup.ah.settled.call(setup.channelID) == false);
     
      let authorization = new Authorization(setup.channelID, setup.parts[setup.A], setup.recv[setup.A], finalBalance[setup.A].toString());
      let signature = await sign(authorization.encode(), setup.parts[setup.A]);
      return truffleAssert.reverts(
        setup.ah.withdraw(authorization, signature, { from: setup.txSender })
      );
    });
  })

  describe("Setting outcome", () => {
    it("wrong parts length", async () => {
      const wrongParts = [setup.parts[setup.A]]
      await truffleAssert.reverts(
        setup.ah.setOutcome(setup.channelID, wrongParts, finalBalance, { from: setup.adj }),
      );
    });

    it("wrong balances length", async () => {
      const wrongBals = [ether(1)]
      await truffleAssert.reverts(
        setup.ah.setOutcome(setup.channelID, setup.parts, wrongBals, { from: setup.adj }),
      );
    });
    
    it("wrong sender", async () => {
      await truffleAssert.reverts(
        setup.ah.setOutcome(setup.channelID, setup.parts, finalBalance, { from: setup.txSender }),
      );
    });

    it("correct sender", async () => {
      truffleAssert.eventEmitted(
        await setup.ah.setOutcome(setup.channelID, setup.parts, finalBalance, { from: setup.adj }),
        'OutcomeSet',
        (ev: any) => { return ev.channelID == setup.channelID }
      );
      assert(await setup.ah.settled.call(setup.channelID) == true);
      for (var i = 0; i < setup.parts.length; i++) {
        let id = fundingID(setup.channelID, setup.parts[i]);
        await assertHoldings(id, finalBalance[i]);
      }
    });

    it("correct sender (twice)", async () => {
      await truffleAssert.reverts(
        setup.ah.setOutcome(setup.channelID, setup.parts, finalBalance, { from: setup.adj })
      );
    });
  })

  describeWithBlockRevert("Invalid withdrawals", () => {
    it("withdraw with invalid signature", async () => {
      let authorization = new Authorization(setup.channelID, setup.parts[setup.A], setup.parts[setup.B], finalBalance[setup.A].toString());
      let signature = await sign(authorization.encode(), setup.parts[setup.B]);
      await truffleAssert.reverts(
        setup.ah.withdraw(authorization, signature, { from: setup.txSender })
      );
    });

    it("invalid balance", async () => {
      let authorization = new Authorization(setup.channelID, setup.parts[setup.A], setup.parts[setup.B], ether(30).toString());
      let signature = await sign(authorization.encode(), setup.parts[setup.A]);
      await truffleAssert.reverts(
        setup.ah.withdraw(authorization, signature, { from: setup.txSender })
      );
    });
  })

  describe("Withdraw", () => {
    it("A withdraws with valid allowance", async () => {
      await testWithdraw(setup.A, finalBalance[setup.A], setup.channelID);
    })
    it("B withdraws with valid allowance", async () => {
      await testWithdraw(setup.B, finalBalance[setup.B], setup.channelID);
    })

    it("A fails to overdraw with valid allowance", async () => {
      let authorization = new Authorization(setup.channelID, setup.parts[setup.A], setup.recv[setup.A], finalBalance[setup.A].toString());
      let signature = await sign(authorization.encode(), setup.parts[setup.A]);
      return truffleAssert.reverts(
        setup.ah.withdraw(authorization, signature, { from: setup.txSender })
      );
    });
  })

  describe("Test underfunded channel", () => {
    let channelID: string

    it("initialize", () => {
      channelID = setup.unfundedChannelID;
    })

    it("A deposits eth", async () => {
      testDeposit(setup.A, ether(1), channelID);
    });

    it("set outcome of the asset holder with deposit refusal", async () => {
      assert(await setup.ah.settled.call(channelID) == false);
      truffleAssert.eventEmitted(
        await setup.ah.setOutcome(channelID, setup.parts, finalBalance, { from: setup.adj }),
        'OutcomeSet',
        (ev: any) => { return ev.channelID == channelID; }
      );
      assert(await setup.ah.settled(channelID), "channel not settled");
      let id = fundingID(channelID, setup.parts[setup.A]);
      assertHoldings(id, ether(1));
    });

    it("A fails to withdraw 2 eth after B's deposit refusal", async () => {
      let authorization = new Authorization(channelID, setup.parts[setup.A], setup.recv[setup.A], ether(2).toString());
      let signature = await sign(authorization.encode(), setup.parts[setup.A]);
      await truffleAssert.reverts(
        setup.ah.withdraw(authorization, signature, { from: setup.txSender })
      );
    });

    it("A withdraws 1 ETH", async () => {
      await testWithdraw(setup.A, ether(1), channelID);
    })
  });
}
