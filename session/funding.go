// Copyright (c) 2023 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/hyperledger-labs/perun-node
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

package session

import (
	"context"

	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
)

// Fund provides a wrapper to call the fund method on session funder.
// On-chain wallet can be direcly used without additional locks,
// because the methods on wallet are concurrency safe by themselves.
func (s *Session) Fund(ctx context.Context, req pchannel.FundingReq) error {
	err := s.funder.Fund(ctx, req)
	if err == nil {
		s.user.OnChain.Wallet.IncrementUsage(s.user.OnChain.Addr)
	}
	return err
}

// RegisterAssetERC20 is a stub that always returns false. Because, the remote
// funder does not support use of assets other than the default ERC20 asset.
//
// TODO: Make actual implementation.
func (s *Session) RegisterAssetERC20(asset pchannel.Asset, token, acc pwallet.Address) bool {
	s.WithField("method", "RegisterAssetERC20").
		Infof("\nReceived request with params %+v, %+v, %+v", asset, token, acc)
	s.WithField("method", "RegisterAssetERC20").Infof("Unimplemented method. Hence returning false")
	return false
}

// IsAssetRegistered provides a wrapper to call the IsAssetRegistered method
// on session funder.
func (s *Session) IsAssetRegistered(asset pchannel.Asset) bool {
	s.WithField("method", "IsAssetRegistered").Infof("\nReceived request with params %+v", asset)
	isAssetRegistered := s.funder.IsAssetRegistered(asset)
	s.WithField("method", "IsAssetRegistered").Infof("Response: %v", isAssetRegistered)
	return isAssetRegistered
}
