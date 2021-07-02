<h1 align="center"><br>
  <a href="https://perun.network/"><img src=".assets/logo.png" alt="Perun" width="196"></a>
<br></h1>

<h4 align="center">Perun State Channels Framework - Ethereum Backend Smart Contracts</h4>

This repository contains the Ethereum smart contracts for [go-perun](https://github.com/perun-network/go-perun)'s Ethereum backend.

## Security Disclaimer
The smart contracts presented in this directory are under active development and are not ready for production use.
The authors take no responsibility for any loss of digital assets or other damage caused by their use.

## Contracts
Perun's Generalized State Channels Framework uses a set of interconnected smart contracts to define the on-chain logic for channel deposits, disputes, settlements and withdrawals.
For more detailed information, check out the [wiki](https://github.com/perun-network/contracts-eth/wiki).

### Asset Holder
Asset holders are singleton contracts that hold the assets for ledger channels.
They are deployed once per asset (ETH, ERC-20, ...) and are shared between all channels that reference the same Adjudicator contract for channel disputing and closing.

Deposits are directly transferred to the Asset Holders.
The outcome of closed channels are set by the Adjudicator on the channel's asset holders.
After the outcome has been set, channel participants can withdraw their assets from the asset holders, sending a Withdrawal Authorization that has to be signed by the respective channel participant.

### Adjudicator
The Adjudicator contract is called to dispute or close a channel.
It interprets channel states and sets finalized channel outcomes on the asset holders.

**Collaborative Close**&emsp;
All channel participants can agree on a final state off-chain.
In this case they can settle a channel without waiting for any timeouts by calling `concludeFinal` on the Adjudicator.
The Adjudicator will set the outcome on the individual asset holders, ready for withdrawal.

**Dispute**&emsp;
In case of a channel dispute, any party can `register` their final state on the Adjudicator contract.
After state registration, the other channel participants have the chance to `refute` the submitted state with a higher-version state during the challenge period.
After the challenge period is over, the channel outcome can either be finalized on the asset holders by calling `conclude` or the app's state can be progressed on-chain by calling `progress`.

### App Contracts
State Channel apps define a single method, `validTransition`, which defines the app-specific state transition rules.
When a channel state is progressed on-chain on the Adjudicator by calling `progress`, the Adjudicator reads the address of the channel app from the channel parameters and, after performing generic state progression checks, calls the `validTransition` method on the app.
It is assumed to revert if any app-specific check fails.

## Testing
The repository must be cloned recursively including [submodules](https://git-scm.com/book/en/v2/Git-Tools-Submodules).
[Yarn](https://yarnpkg.com), [Typescript](https://www.typescriptlang.org), and [Truffle](https://truffleframework.com/) are expected to be installed globally.
To run the tests, run
```sh
$ yarn
$ yarn build
$ yarn test
```
This has been tested with Truffle version `5.1.46`.

## Copyright
Copyright 2020 - See [NOTICE](NOTICE) file for copyright holders.
Use of the source code is governed by the Apache 2.0 license that can be found in the [LICENSE file](LICENSE).

Contact us at [info@perun.network](mailto:info@perun.network).
