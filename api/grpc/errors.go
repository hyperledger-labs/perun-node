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

// FromError is a helper function to convert APIError struct defined in perun-node
// to APIError struct defined in grpc package.
func FromError(err perun.APIError) *pb.MsgError { //nolint:funlen
	grpcErr := pb.MsgError{
		Category: pb.ErrorCategory(err.Category()),
		Code:     pb.ErrorCode(err.Code()),
		Message:  err.Message(),
	}
	switch info := err.AddInfo().(type) {
	case perun.ErrInfoPeerRequestTimedOut:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoPeerRequestTimedOut{
			ErrInfoPeerRequestTimedOut: &pb.ErrInfoPeerRequestTimedOut{
				Timeout: info.Timeout,
			},
		}
	case perun.ErrInfoPeerRejected:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoPeerRejected{
			ErrInfoPeerRejected: &pb.ErrInfoPeerRejected{
				PeerAlias: info.PeerAlias,
				Reason:    info.Reason,
			},
		}
	case perun.ErrInfoPeerNotFunded:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoPeerNotFunded{
			ErrInfoPeerNotFunded: &pb.ErrInfoPeerNotFunded{
				PeerAlias: info.PeerAlias,
			},
		}
	case perun.ErrInfoUserResponseTimedOut:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoUserResponseTimedOut{
			ErrInfoUserResponseTimedOut: &pb.ErrInfoUserResponseTimedOut{
				Expiry:     info.Expiry,
				ReceivedAt: info.ReceivedAt,
			},
		}
	case perun.ErrInfoResourceNotFound:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoResourceNotFound{
			ErrInfoResourceNotFound: &pb.ErrInfoResourceNotFound{
				Type: info.Type,
				Id:   info.ID,
			},
		}
	case perun.ErrInfoResourceExists:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoResourceExists{
			ErrInfoResourceExists: &pb.ErrInfoResourceExists{
				Type: info.Type,
				Id:   info.ID,
			},
		}
	case perun.ErrInfoInvalidArgument:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoInvalidArgument{
			ErrInfoInvalidArgument: &pb.ErrInfoInvalidArgument{
				Name:        info.Name,
				Value:       info.Value,
				Requirement: info.Requirement,
			},
		}
	case payment.ErrInfoFailedPreCondUnclosedPayChs:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoFailedPreCondUnclosedChs{
			ErrInfoFailedPreCondUnclosedChs: &pb.ErrInfoFailedPreCondUnclosedChs{
				Chs: FromPayChsInfo(info.PayChs),
			},
		}
	case perun.ErrInfoInvalidConfig:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoInvalidConfig{
			ErrInfoInvalidConfig: &pb.ErrInfoInvalidConfig{
				Name:  info.Name,
				Value: info.Value,
			},
		}
	case perun.ErrInfoInvalidContracts:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoInvalidContracts{
			ErrInfoInvalidContracts: &pb.ErrInfoInvalidContracts{
				ContractErrInfos: FromContractErrInfos(info.ContractErrInfos),
			},
		}
	case perun.ErrInfoTxTimedOut:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoTxTimedOut{
			ErrInfoTxTimedOut: &pb.ErrInfoTxTimedOut{
				TxType:    info.TxType,
				TxID:      info.TxID,
				TxTimeout: info.TxTimeout,
			},
		}
	case perun.ErrInfoChainNotReachable:
		grpcErr.AddInfo = &pb.MsgError_ErrInfoChainNotReachable{
			ErrInfoChainNotReachable: &pb.ErrInfoChainNotReachable{
				ChainURL: info.ChainURL,
			},
		}
	default:
		// It is Unknonwn Internal Error which has no additional info.
		grpcErr.AddInfo = nil
	}
	return &grpcErr
}

// FromContractErrInfos is a helper function to convert a slice of
// ContractErrInfo struct defined in perun-node to a slice of ContractErrInfo
// struct defined in grpc package.
func FromContractErrInfos(src []perun.ContractErrInfo) []*pb.ContractErrInfo {
	output := make([]*pb.ContractErrInfo, len(src))
	for i := range src {
		output[i].Name = src[i].Name
		output[i].Address = src[i].Address
		output[i].Error = src[i].Error
	}
	return output
}
