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

// FromError is a helper function to convert APIError struct defined in perun-node
// to APIError struct defined in grpc package.
func FromError(err perun.APIError) *MsgError { //nolint:funlen
	grpcErr := MsgError{
		Category: ErrorCategory(err.Category()),
		Code:     ErrorCode(err.Code()),
		Message:  err.Message(),
	}
	switch info := err.AddInfo().(type) {
	case perun.ErrInfoPeerRequestTimedOut:
		grpcErr.AddInfo = &MsgError_ErrInfoPeerRequestTimedOut{
			ErrInfoPeerRequestTimedOut: &ErrInfoPeerRequestTimedOut{
				Timeout: info.Timeout,
			},
		}
	case perun.ErrInfoPeerRejected:
		grpcErr.AddInfo = &MsgError_ErrInfoPeerRejected{
			ErrInfoPeerRejected: &ErrInfoPeerRejected{
				PeerAlias: info.PeerAlias,
				Reason:    info.Reason,
			},
		}
	case perun.ErrInfoPeerNotFunded:
		grpcErr.AddInfo = &MsgError_ErrInfoPeerNotFunded{
			ErrInfoPeerNotFunded: &ErrInfoPeerNotFunded{
				PeerAlias: info.PeerAlias,
			},
		}
	case perun.ErrInfoUserResponseTimedOut:
		grpcErr.AddInfo = &MsgError_ErrInfoUserResponseTimedOut{
			ErrInfoUserResponseTimedOut: &ErrInfoUserResponseTimedOut{
				Expiry:     info.Expiry,
				ReceivedAt: info.ReceivedAt,
			},
		}
	case perun.ErrInfoResourceNotFound:
		grpcErr.AddInfo = &MsgError_ErrInfoResourceNotFound{
			ErrInfoResourceNotFound: &ErrInfoResourceNotFound{
				Type: info.Type,
				Id:   info.ID,
			},
		}
	case perun.ErrInfoResourceExists:
		grpcErr.AddInfo = &MsgError_ErrInfoResourceExists{
			ErrInfoResourceExists: &ErrInfoResourceExists{
				Type: info.Type,
				Id:   info.ID,
			},
		}
	case perun.ErrInfoInvalidArgument:
		grpcErr.AddInfo = &MsgError_ErrInfoInvalidArgument{
			ErrInfoInvalidArgument: &ErrInfoInvalidArgument{
				Name:        info.Name,
				Value:       info.Value,
				Requirement: info.Requirement,
			},
		}
	case payment.ErrInfoFailedPreCondUnclosedPayChs:
		grpcErr.AddInfo = &MsgError_ErrInfoFailedPreCondUnclosedChs{
			ErrInfoFailedPreCondUnclosedChs: &ErrInfoFailedPreCondUnclosedChs{
				Chs: FromPayChsInfo(info.PayChs),
			},
		}
	case perun.ErrInfoInvalidConfig:
		grpcErr.AddInfo = &MsgError_ErrInfoInvalidConfig{
			ErrInfoInvalidConfig: &ErrInfoInvalidConfig{
				Name:  info.Name,
				Value: info.Value,
			},
		}
	case perun.ErrInfoInvalidContracts:
		grpcErr.AddInfo = &MsgError_ErrInfoInvalidContracts{
			ErrInfoInvalidContracts: &ErrInfoInvalidContracts{
				ContractErrInfos: FromContractErrInfos(info.ContractErrInfos),
			},
		}
	case perun.ErrInfoTxTimedOut:
		grpcErr.AddInfo = &MsgError_ErrInfoTxTimedOut{
			ErrInfoTxTimedOut: &ErrInfoTxTimedOut{
				TxType:    info.TxType,
				TxID:      info.TxID,
				TxTimeout: info.TxTimeout,
			},
		}
	case perun.ErrInfoChainNotReachable:
		grpcErr.AddInfo = &MsgError_ErrInfoChainNotReachable{
			ErrInfoChainNotReachable: &ErrInfoChainNotReachable{
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
func FromContractErrInfos(src []perun.ContractErrInfo) []*ContractErrInfo {
	output := make([]*ContractErrInfo, len(src))
	for i := range src {
		output[i].Name = src[i].Name
		output[i].Address = src[i].Address
		output[i].Error = src[i].Error
	}
	return output
}
