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

var Adjudicator = artifacts.require("Adjudicator");
var AssetHolderETH = artifacts.require("AssetHolderETH");
var AssetHolderERC20 = artifacts.require("AssetHolderERC20");
var PerunToken = artifacts.require("PerunToken");
var Channel = artifacts.require("Channel");

module.exports = async function(deployer, _network, accounts) {
  await deployer.deploy(Channel);
  await deployer.link(Channel, Adjudicator);
  await deployer.deploy(Adjudicator);

  await deployer.deploy(AssetHolderETH, Adjudicator.address);
  await deployer.deploy(PerunToken, accounts, 1<<10);
  await deployer.deploy(AssetHolderERC20, Adjudicator.address, PerunToken.address);
};
