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
	"fmt"
	"math"
	"math/big"

	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
)

// ToParams converts protobuf's Params definition to perun's Params definition.
func ToParams(protoParams *Params) (*pchannel.Params, error) {
	app, err := toApp(protoParams.App)
	if err != nil {
		return nil, err
	}
	parts, err := toWalletAddrs(protoParams.Parts)
	if err != nil {
		return nil, errors.WithMessage(err, "parts")
	}
	params := pchannel.NewParamsUnsafe(
		protoParams.ChallengeDuration,
		parts,
		app,
		(new(big.Int)).SetBytes(protoParams.Nonce),
		protoParams.LedgerChannel,
		protoParams.VirtualChannel)

	return params, nil
}

func toApp(protoApp []byte) (app pchannel.App, err error) {
	if len(protoApp) == 0 {
		app = pchannel.NoApp()
		return app, nil
	}
	appDef := pwallet.NewAddress()
	err = appDef.UnmarshalBinary(protoApp)
	if err != nil {
		return app, err
	}
	app, err = pchannel.Resolve(appDef)
	return app, err
}

func toWalletAddrs(protoAddrs [][]byte) ([]pwallet.Address, error) {
	addrs := make([]pwallet.Address, len(protoAddrs))
	for i := range protoAddrs {
		addrs[i] = pwallet.NewAddress()
		err := addrs[i].UnmarshalBinary(protoAddrs[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return addrs, nil
}

// ToState converts protobuf's State definition to perun's State definition.
func ToState(protoState *State) (state *pchannel.State, err error) {
	state = &pchannel.State{}
	copy(state.ID[:], protoState.Id)
	state.Version = protoState.Version
	state.IsFinal = protoState.IsFinal
	allocation, err := toAllocation(protoState.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "allocation")
	}
	state.Allocation = *allocation
	state.App, state.Data, err = toAppAndData(protoState.App, protoState.Data)
	return state, err
}

func toAllocation(protoAlloc *Allocation) (alloc *pchannel.Allocation, err error) {
	alloc = &pchannel.Allocation{}
	alloc.Assets = make([]pchannel.Asset, len(protoAlloc.Assets))
	for i := range protoAlloc.Assets {
		alloc.Assets[i] = pchannel.NewAsset()
		err = alloc.Assets[i].UnmarshalBinary(protoAlloc.Assets[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th asset", i)
		}
	}
	alloc.Locked = make([]pchannel.SubAlloc, len(protoAlloc.Locked))
	for i := range protoAlloc.Locked {
		alloc.Locked[i], err = toSubAlloc(protoAlloc.Locked[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th sub alloc", i)
		}
	}
	alloc.Balances = ToBalances(protoAlloc.Balances.Balances)
	return alloc, nil
}

func toAppAndData(protoApp, protoData []byte) (app pchannel.App, data pchannel.Data, err error) {
	if len(protoApp) == 0 {
		app = pchannel.NoApp()
		data = pchannel.NoData()
		return app, data, nil
	}
	appDef := pwallet.NewAddress()
	err = appDef.UnmarshalBinary(protoApp)
	if err != nil {
		return nil, nil, err
	}
	app, err = pchannel.Resolve(appDef)
	if err != nil {
		return
	}
	data = app.NewData()
	return app, data, data.UnmarshalBinary(protoData)
}

func toSubAlloc(protoSubAlloc *SubAlloc) (subAlloc pchannel.SubAlloc, err error) {
	subAlloc = pchannel.SubAlloc{}
	subAlloc.Bals = toBalance(protoSubAlloc.Bals)
	if len(protoSubAlloc.Id) != len(subAlloc.ID) {
		return subAlloc, errors.New("sub alloc id has incorrect length")
	}
	copy(subAlloc.ID[:], protoSubAlloc.Id)
	subAlloc.IndexMap, err = toIndexMap(protoSubAlloc.IndexMap.IndexMap)
	return subAlloc, err
}

// ToBalances converts protobuf's Balances definition to perun's Balances definition.
func ToBalances(protoBalances []*Balance) [][]*big.Int {
	balances := make([][]*big.Int, len(protoBalances))
	for i, protoBalance := range protoBalances {
		balances[i] = make([]*big.Int, len(protoBalance.Balance))
		for j := range protoBalance.Balance {
			balances[i][j] = (&big.Int{}).SetBytes(protoBalance.Balance[j])
			balances[i][j].SetBytes(protoBalance.Balance[j])
		}
	}
	return balances
}

func toBalance(protoBalance *Balance) (balance []pchannel.Bal) {
	balance = make([]pchannel.Bal, len(protoBalance.Balance))
	for j := range protoBalance.Balance {
		balance[j] = new(big.Int).SetBytes(protoBalance.Balance[j])
	}
	return balance
}

func toIndexMap(protoIndexMap []uint32) (indexMap []pchannel.Index, err error) {
	indexMap = make([]pchannel.Index, len(protoIndexMap))
	for i := range protoIndexMap {
		if protoIndexMap[i] > math.MaxUint16 {
			//nolint:goerr113  // We do not want to define this as constant error.
			return nil, fmt.Errorf("%d'th index is invalid", i)
		}
		indexMap[i] = pchannel.Index(uint16(protoIndexMap[i]))
	}
	return indexMap, nil
}

// FromParams converts perun's Params definition to protobuf's Params
// definition.
func FromParams(params *pchannel.Params) (protoParams *Params, err error) {
	protoParams = &Params{}

	protoParams.Nonce = params.Nonce.Bytes()
	protoParams.ChallengeDuration = params.ChallengeDuration
	protoParams.LedgerChannel = params.LedgerChannel
	protoParams.VirtualChannel = params.VirtualChannel
	protoParams.Parts, err = fromWalletAddrs(params.Parts)
	if err != nil {
		return nil, errors.WithMessage(err, "parts")
	}
	protoParams.App, err = fromApp(params.App)
	return protoParams, err
}

func fromWalletAddrs(addrs []pwallet.Address) (protoAddrs [][]byte, err error) {
	protoAddrs = make([][]byte, len(addrs))
	for i := range addrs {
		protoAddrs[i], err = addrs[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return protoAddrs, nil
}

func fromApp(app pchannel.App) (protoApp []byte, err error) {
	if pchannel.IsNoApp(app) {
		return []byte{}, nil
	}
	protoApp, err = app.Def().MarshalBinary()
	return protoApp, err
}

// FromState converts perun's State definition to protobuf's State
// definition.
func FromState(state *pchannel.State) (protoState *State, err error) {
	protoState = &State{}

	protoState.Id = make([]byte, len(state.ID))
	copy(protoState.Id, state.ID[:])
	protoState.Version = state.Version
	protoState.IsFinal = state.IsFinal
	protoState.Allocation, err = fromAllocation(state.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "allocation")
	}
	protoState.App, protoState.Data, err = fromAppAndData(state.App, state.Data)
	return protoState, err
}

func fromAllocation(alloc pchannel.Allocation) (protoAlloc *Allocation, err error) {
	protoAlloc = &Allocation{}
	protoAlloc.Assets = make([][]byte, len(alloc.Assets))
	for i := range alloc.Assets {
		protoAlloc.Assets[i], err = alloc.Assets[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th asset", i)
		}
	}
	locked := make([]*SubAlloc, len(alloc.Locked))
	for i := range alloc.Locked {
		locked[i], err = fromSubAlloc(alloc.Locked[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th sub alloc", i)
		}
	}
	protoAlloc.Balances, err = FromBalances(alloc.Balances)
	return protoAlloc, err
}

// FromBalances converts perun's Balances definition to protobuf's Balances
// definition.
func FromBalances(balances pchannel.Balances) (protoBalances *Balances, err error) {
	protoBalances = &Balances{
		Balances: make([]*Balance, len(balances)),
	}
	for i := range balances {
		protoBalances.Balances[i], err = fromBalance(balances[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th balance", i)
		}
	}
	return protoBalances, nil
}

func fromBalance(balance []pchannel.Bal) (protoBalance *Balance, err error) {
	protoBalance = &Balance{
		Balance: make([][]byte, len(balance)),
	}
	for i := range balance {
		if balance[i] == nil {
			return nil, fmt.Errorf("%d'th amount is nil", i) //nolint:goerr113  // constant error is not needed.
		}
		if balance[i].Sign() == -1 {
			return nil, fmt.Errorf("%d'th amount is negative", i) //nolint:goerr113  // constant error is not needed.
		}
		protoBalance.Balance[i] = balance[i].Bytes()
	}
	return protoBalance, nil
}

func fromAppAndData(app pchannel.App, data pchannel.Data) (protoApp, protoData []byte, err error) {
	if pchannel.IsNoApp(app) {
		return []byte{}, []byte{}, nil
	}
	protoApp, err = app.Def().MarshalBinary()
	if err != nil {
		return []byte{}, []byte{}, err
	}
	protoData, err = data.MarshalBinary()
	return protoApp, protoData, err
}

func fromSubAlloc(subAlloc pchannel.SubAlloc) (protoSubAlloc *SubAlloc, err error) {
	protoSubAlloc = &SubAlloc{}
	protoSubAlloc.Id = make([]byte, len(subAlloc.ID))
	copy(protoSubAlloc.Id, subAlloc.ID[:])
	protoSubAlloc.IndexMap = &IndexMap{IndexMap: fromIndexMap(subAlloc.IndexMap)}
	protoSubAlloc.Bals, err = fromBalance(subAlloc.Bals)
	return protoSubAlloc, err
}

func fromIndexMap(indexMap []pchannel.Index) (protoIndexMap []uint32) {
	protoIndexMap = make([]uint32, len(indexMap))
	for i := range indexMap {
		protoIndexMap[i] = uint32(indexMap[i])
	}
	return protoIndexMap
}
