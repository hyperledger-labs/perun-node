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

/// <reference types="truffle-typings" />

import { promisify } from "util";
import Web3 from "web3";
declare const web3: Web3;

export async function sign(data: string, account: string) {
  let sig = await web3.eth.sign(web3.utils.soliditySha3(data) as string, account);
  // fix wrong v value (add 27)
  let v = sig.slice(130, 132);
  return sig.slice(0, 130) + (parseInt(v, 16) + 27).toString(16);
}

export function ether(x: number): BN { return web3.utils.toWei(web3.utils.toBN(x), "ether"); }

export function wei2eth(x: BN): BN { return web3.utils.toBN(web3.utils.fromWei(x, "ether")); }

export function hash(...val: any[]): string {
  return web3.utils.soliditySha3(...val) as string
}

export async function asyncWeb3Send(method: string, params: any[], id?: number): Promise<any> {
  let req: any = { jsonrpc: '2.0', method: method, params: params };
  if (id != undefined) req.id = id;

  return promisify((callback) => {
    (web3.currentProvider as any).send(req, callback)
  })();
}

export async function currentTimestamp(): Promise<number> {
  let blocknumber = await web3.eth.getBlockNumber();
  let block = await web3.eth.getBlock(blocknumber);
  return block.timestamp as number;
}
