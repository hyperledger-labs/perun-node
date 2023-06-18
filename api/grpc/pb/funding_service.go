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

package pb

import (
	pchannel "perun.network/go-perun/channel"
)

// ToFundingReq converts protobuf's FundingReq definition to perun's FundingReq
// definition.
func ToFundingReq(protoReq *FundReq) (req pchannel.FundingReq, err error) {
	if req.Params, err = ToParams(protoReq.Params); err != nil {
		return req, err
	}
	if req.State, err = ToState(protoReq.State); err != nil {
		return req, err
	}

	req.Idx = pchannel.Index(protoReq.Idx)
	req.Agreement = ToBalances(protoReq.Agreement.Balances)
	return req, nil
}
