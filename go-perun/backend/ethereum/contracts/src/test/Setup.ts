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

import { hash } from "../lib/web3";
import web3 from "web3";

// AssetHolderSetup is the setup for `genericAssetHolderTest`. 
export class AssetHolderSetup {
    channelID: string;
    unfundedChannelID: string;
    txSender: string;
    adj: string;
    recv: string[];
    parts: string[];
    A = 0; B = 1;
    accounts: string[];
    ah: any;    
    deposit: (fid: string, amount: BN, from: string) => Promise<Truffle.TransactionResponse>;
    balanceOf: (who: string) => Promise<BN>;
    /**
     * Index of the Adjudicator address in the `accounts` array.
     */

    constructor(ah: any, accounts: string[], deposit: (fid: string, amount: BN, from: string) => Promise<Truffle.TransactionResponse>, balanceOf: (who: string) => Promise<BN>) {
        this.channelID = hash(web3.utils.randomHex(32));
        this.unfundedChannelID = hash(web3.utils.randomHex(32));
        this.txSender = accounts[5];
        this.adj = accounts[9];
        this.parts = [accounts[1], accounts[2]];
        this.recv = [accounts[3], accounts[4]];
        this.accounts = accounts;
        this.ah = ah;
        this.deposit = deposit;
        this.balanceOf = balanceOf;
    }
}
