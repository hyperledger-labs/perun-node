// Copyright (c) 2020 - for information on the respective copyright owner
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
	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/app/payment"
)

// FromPayments is a helper function to convert slice of Payment struct
// defined in perun-node package to slice of Payment struct defined in grpc
// package.
func FromPayments(payments []payment.Payment) []*Payment {
	grpcPayments := make([]*Payment, len(payments))
	for i := range payments {
		grpcPayments[i] = FromPayment(payments[i])
	}
	return grpcPayments
}

// FromPayment is a helper function to convert Payment struct defined in
// perun-node package to Payment struct defined in gprc package.
func FromPayment(src payment.Payment) *Payment {
	return &Payment{
		Currency: src.Currency,
		Payee:    src.Payee,
		Amount:   src.Amount,
	}
}

// ToPayments is a helper function to convert slice of Payment struct defined in
// grpc package to slice of Payment struct defined in perun-node.
func ToPayments(payments []*Payment) []payment.Payment {
	grpcPayments := make([]payment.Payment, len(payments))
	for i := range payments {
		grpcPayments[i] = ToPayment(payments[i])
	}
	return grpcPayments
}

// ToPayment is a helper function to convert Payment struct defined in
// grpc package to Payment struct defined in perun-node.
func ToPayment(src *Payment) payment.Payment {
	return payment.Payment{
		Currency: src.Currency,
		Payee:    src.Payee,
		Amount:   src.Amount,
	}
}

// FromPayChsInfo is a helper function to convert slice of PayChInfo struct
// defined in perun-node to a slice of PayChInfo struct defined in grpc
// package.
func FromPayChsInfo(payChsInfo []payment.PayChInfo) []*PayChInfo {
	grpcPayChsInfo := make([]*PayChInfo, len(payChsInfo))
	for i := range payChsInfo {
		grpcPayChsInfo[i] = FromPayChInfo(payChsInfo[i])
	}
	return grpcPayChsInfo
}

// FromPayChInfo is a helper function to convert PayChInfo struct defined in perun-node
// to PayChInfo struct defined in grpc package.
func FromPayChInfo(src payment.PayChInfo) *PayChInfo {
	return &PayChInfo{
		ChID:    src.ChID,
		BalInfo: FromBalInfo(src.BalInfo),
		Version: src.Version,
	}
}

// ToBalInfo is a helper function to convert BalInfo struct defined in grpc package
// to BalInfo struct defined in perun-node.
func ToBalInfo(src *BalInfo) perun.BalInfo {
	bals := make([][]string, len(src.Bals))
	for i := range src.Bals {
		bals[i] = src.Bals[i].Bal
	}
	return perun.BalInfo{
		Currencies: src.Currencies,
		Parts:      src.Parts,
		Bals:       bals,
	}
}

// FromBalInfo is a helper function to convert BalInfo struct defined in perun-node
// to BalInfo struct defined in grpc package.
func FromBalInfo(src perun.BalInfo) *BalInfo {
	bals := make([]*BalInfoBal, len(src.Bals))
	for i := range src.Bals {
		bals[i] = &BalInfoBal{}
		bals[i].Bal = src.Bals[i]
	}
	return &BalInfo{
		Currencies: src.Currencies,
		Parts:      src.Parts,
		Bals:       bals,
	}
}

// ToPayChsInfo is a helper function to convert a slice of PayChInfo struct
// defined in grpc package to a slice of PayChInfo struct defined in perun-node package.
func ToPayChsInfo(grpcPayChsInfo []*PayChInfo) []payment.PayChInfo {
	payChsInfo := make([]payment.PayChInfo, len(grpcPayChsInfo))
	for i := range grpcPayChsInfo {
		payChsInfo[i] = ToPayChInfo(grpcPayChsInfo[i])
	}
	return payChsInfo
}

// ToPayChInfo is a helper function to convert PayChInfo struct defined in grpc
// package to PayChInfo struct defined in perun-node package.
func ToPayChInfo(src *PayChInfo) payment.PayChInfo {
	return payment.PayChInfo{
		ChID:    src.ChID,
		BalInfo: ToBalInfo(src.BalInfo),
		Version: src.Version,
	}
}
