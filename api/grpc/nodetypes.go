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

package grpc

import (
	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/app/payment"
)

// FromPayments is a helper function to convert slice of Payment struct
// defined in perun-node package to slice of Payment struct defined in grpc
// package.
func FromPayments(payments []payment.Payment) []*pb.Payment {
	grpcPayments := make([]*pb.Payment, len(payments))
	for i := range payments {
		grpcPayments[i] = FromPayment(payments[i])
	}
	return grpcPayments
}

// FromPayment is a helper function to convert Payment struct defined in
// perun-node package to Payment struct defined in gprc package.
func FromPayment(src payment.Payment) *pb.Payment {
	return &pb.Payment{
		Currency: src.Currency,
		Payee:    src.Payee,
		Amount:   src.Amount,
	}
}

// fromGrpcPayment is a helper function to convert slice of Payment struct defined in
// grpc package to slice of Payment struct defined in perun-node.
func ToPayments(payments []*pb.Payment) []payment.Payment {
	grpcPayments := make([]payment.Payment, len(payments))
	for i := range payments {
		grpcPayments[i] = ToPayment(payments[i])
	}
	return grpcPayments
}

// ToPayment is a helper function to convert Payment struct defined in
// grpc package to Payment struct defined in perun-node.
func ToPayment(src *pb.Payment) payment.Payment {
	return payment.Payment{
		Currency: src.Currency,
		Payee:    src.Payee,
		Amount:   src.Amount,
	}
}

// toGrpcPayChInfo is a helper function to convert slice of PayChInfo struct defined in perun-node
// to a slice of PayChInfo struct defined in grpc package.
func FromPayChsInfo(payChsInfo []payment.PayChInfo) []*pb.PayChInfo {
	grpcPayChsInfo := make([]*pb.PayChInfo, len(payChsInfo))
	for i := range payChsInfo {
		grpcPayChsInfo[i] = FromPayChInfo(payChsInfo[i])
	}
	return grpcPayChsInfo
}

// FromPayChInfo is a helper function to convert PayChInfo struct defined in perun-node
// to PayChInfo struct defined in grpc package.
func FromPayChInfo(src payment.PayChInfo) *pb.PayChInfo {
	return &pb.PayChInfo{
		ChID:    src.ChID,
		BalInfo: FromBalInfo(src.BalInfo),
		Version: src.Version,
	}
}

// ToBalInfo is a helper function to convert BalInfo struct defined in grpc package
// to BalInfo struct defined in perun-node.
func ToBalInfo(src *pb.BalInfo) perun.BalInfo {
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
func FromBalInfo(src perun.BalInfo) *pb.BalInfo {
	bals := make([]*pb.BalInfoBal, len(src.Bals))
	for i := range src.Bals {
		bals[i] = &pb.BalInfoBal{}
		bals[i].Bal = src.Bals[i]
	}
	return &pb.BalInfo{
		Currencies: src.Currencies,
		Parts:      src.Parts,
		Bals:       bals,
	}
}
