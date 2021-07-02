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

// his file contains the types:
// DisputePhase, Channel, Params, State, Allocation, SubAlloc, Transaction and
// Authorization.

import Web3 from "web3";
declare const web3: Web3;
import { sign, hash } from "../lib/web3";

export enum DisputePhase { DISPUTE, FORCEEXEC, CONCLUDED }

export class Channel {
  params: Params
  state: State

  constructor(params: Params, state: State) {
    this.params = params
    this.state = state
  }
}

export class Params {
  challengeDuration: number;
  nonce: string;
  app: string;
  participants: string[];

  constructor(_app: string, _challengeDuration: number, _nonce: string, _parts: string[]) {
    this.app = _app;
    this.challengeDuration = _challengeDuration;
    this.nonce = _nonce;
    this.participants = _parts;
  }

  serialize() {
    return {
      app: this.app,
      challengeDuration: this.challengeDuration,
      nonce: this.nonce,
      participants: this.participants
    };
  }

  encode() {
    const paramsType = {
      "components": [
        {
          "internalType": "uint256",
          "name": "challengeDuration",
          "type": "uint256"
        },
        {
          "internalType": "uint256",
          "name": "nonce",
          "type": "uint256"
        },
        {
          "internalType": "address",
          "name": "app",
          "type": "address"
        },
        {
          "internalType": "address[]",
          "name": "participants",
          "type": "address[]"
        }
      ],
      "internalType": "struct Channel.Params",
      "name": "params",
      "type": "tuple"
    }
    return web3.eth.abi.encodeParameter(paramsType, this)
  }

  channelID() {
    return hash(this.encode());
  }
}

export class State {
  channelID: string;
  version: string;
  outcome: Allocation;
  appData: string;
  isFinal: boolean;

  constructor(_channelID: string, _version: string, _outcome: Allocation, _appData: string, _isFinal: boolean) {
    this.channelID = _channelID;
    this.version = _version;
    this.outcome = _outcome;
    this.appData = _appData;
    this.isFinal = _isFinal;
  }

  serialize() {
    return {
      channelID: this.channelID,
      version: this.version,
      outcome: this.outcome.serialize(),
      appData: this.appData,
      isFinal: this.isFinal
    }
  }

  encode() {
    const stateType = {
      "components": [
        {
          "internalType": "bytes32",
          "name": "channelID",
          "type": "bytes32"
        },
        {
          "internalType": "uint64",
          "name": "version",
          "type": "uint64"
        },
        {
          "components": [
            {
              "internalType": "address[]",
              "name": "assets",
              "type": "address[]"
            },
            {
              "internalType": "uint256[][]",
              "name": "balances",
              "type": "uint256[][]"
            },
            {
              "components": [
                {
                  "internalType": "bytes32",
                  "name": "ID",
                  "type": "bytes32"
                },
                {
                  "internalType": "uint256[]",
                  "name": "balances",
                  "type": "uint256[]"
                }
              ],
              "internalType": "struct Channel.SubAlloc[]",
              "name": "locked",
              "type": "tuple[]"
            }
          ],
          "internalType": "struct Channel.Allocation",
          "name": "outcome",
          "type": "tuple"
        },
        {
          "internalType": "bytes",
          "name": "appData",
          "type": "bytes"
        },
        {
          "internalType": "bool",
          "name": "isFinal",
          "type": "bool"
        }
      ],
      "internalType": "struct Channel.State",
      "name": "state",
      "type": "tuple"
    };

    return web3.eth.abi.encodeParameter(stateType, this);
  }

  incrementVersion() {
    this.version = (Number(this.version) + 1).toString()
  }

  async sign(signers: string[]): Promise<string[]> {
    return Promise.all(signers.map(signer => sign(this.encode(), signer)))
  }
}

export class Allocation {
    assets: string[];
    balances: string[][];
    locked: SubAlloc[];
  
    constructor(_assets: string[], _balances: string[][], _locked: SubAlloc[]) {
      this.assets = _assets;
      this.balances = _balances;
      this.locked = _locked;
    }
  
    serialize() {
      let _locked: any[] = this.locked.map(e => e.serialize());
      return { assets: this.assets, balances: this.balances, locked: _locked };
    }
  }

export class SubAlloc {
    ID: string;
    balances: string[];

    constructor(id: string, _balances: string[]) {
        this.ID = id;
        this.balances = _balances;
    }

    serialize() {
        return { ID: this.ID, balances: this.balances };
    }
}

export class Transaction extends Channel {
    sigs: string[];
  
    constructor(parts: string[], balances: BN[], challengeDuration: number, nonce: string, asset: string, app: string) {
      const params = new Params(app, challengeDuration, nonce, [parts[0], parts[1]]);
      const outcome = new Allocation([asset], [[balances[0].toString(), balances[1].toString()]], []);
      const state = new State(params.channelID(), "0", outcome, "0x00", false);
      super(params, state);
      this.sigs = [];
    }
  
    async sign(parts: string[]) {
      let stateEncoded = this.state.encode();
      this.sigs = await Promise.all(parts.map(participant => sign(stateEncoded, participant)));
    }
  }
  

export class Authorization {
    channelID: string;
    participant: string;
    receiver: string;
    amount: string;
  
    constructor(_channelID: string, _participant: string, _receiver: string, _amount: string) {
      this.channelID = _channelID;
      this.participant = _participant;
      this.receiver = _receiver;
      this.amount = _amount;
    }
  
    serialize() {
      return {
        channelID: this.channelID,
        participant: this.participant,
        receiver: this.receiver,
        amount: this.amount
      };
    }
  
    encode() {
      return web3.eth.abi.encodeParameters(
        ['bytes32', 'address', 'address', 'uint256'],
        [
          web3.utils.rightPad(this.channelID, 64, "0"),
          this.participant,
          this.receiver,
          web3.utils.padLeft(this.amount.toString(), 64, "0")
        ]
      );
    }
  }
