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
	"context"
	"net"

	"github.com/pkg/errors"
	grpclib "google.golang.org/grpc"
	psync "perun.network/go-perun/pkg/sync"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/app/payment"
)

// payChAPIServer represents a grpc server that can serve payment channel API.
type payChAPIServer struct {
	n perun.NodeAPI

	// The mutex should be used when accessing the map data structures.
	psync.Mutex

	// These maps are used to hold an signal channel for each active subscription.
	// When a subscription is registered, subscribe function will add an entry to the
	// map corresponding to the subscription type.
	// The unsubscribe call should retrieve the channel from the map and close it, which
	// will signal the subscription routine to end.
	//
	// chProposalsNotif works on per session basis and hence this is a map
	// of session id to signaling channel.
	// chUpdatesNotif works on a per channel basis and hence this is a map of session id to
	// channel id to signaling channel.

	chProposalsNotif map[string]chan bool
	chUpdatesNotif   map[string]map[string]chan bool
}

// ListenAndServePayChAPI starts a payment channel API server that listens for incoming grpc
// requests at the specified address and serves those requests using the node API instance.
func ListenAndServePayChAPI(n perun.NodeAPI, grpcPort string) error {
	apiServer := &payChAPIServer{
		n:                n,
		chProposalsNotif: make(map[string]chan bool),
		chUpdatesNotif:   make(map[string]map[string]chan bool),
	}

	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return errors.Wrap(err, "starting listener")
	}
	grpcServer := grpclib.NewServer()
	pb.RegisterPayment_APIServer(grpcServer, apiServer)

	return grpcServer.Serve(listener)
}

// GetConfig wraps node.GetConfig.
func (a *payChAPIServer) GetConfig(context.Context, *pb.GetConfigReq) (*pb.GetConfigResp, error) {
	cfg := a.n.GetConfig()
	return &pb.GetConfigResp{
		ChainAddress:       cfg.ChainURL,
		AdjudicatorAddress: cfg.Adjudicator,
		AssetAddress:       cfg.Asset,
		CommTypes:          cfg.CommTypes,
		IdProviderTypes:    cfg.IDProviderTypes,
	}, nil
}

// Time wraps node.Time.
func (a *payChAPIServer) Time(context.Context, *pb.TimeReq) (*pb.TimeResp, error) {
	return &pb.TimeResp{
		Time: a.n.Time(),
	}, nil
}

// Help wraps node.Help.
func (a *payChAPIServer) Help(context.Context, *pb.HelpReq) (*pb.HelpResp, error) {
	return &pb.HelpResp{
		Apis: a.n.Help(),
	}, nil
}

// OpenSession wraps node.OpenSession.
func (a *payChAPIServer) OpenSession(ctx context.Context, req *pb.OpenSessionReq) (*pb.OpenSessionResp, error) {
	errResponse := func(err perun.APIError) *pb.OpenSessionResp {
		return &pb.OpenSessionResp{
			Response: &pb.OpenSessionResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sessionID, restoredChs, err := payment.OpenSession(a.n, req.ConfigFile)
	if err != nil {
		return errResponse(err), nil
	}

	a.Lock()
	a.chUpdatesNotif[sessionID] = make(map[string]chan bool)
	a.Unlock()

	return &pb.OpenSessionResp{
		Response: &pb.OpenSessionResp_MsgSuccess_{
			MsgSuccess: &pb.OpenSessionResp_MsgSuccess{
				SessionID:   sessionID,
				RestoredChs: toGrpcPayChsInfo(restoredChs),
			},
		},
	}, nil
}

// AddPeerID wraps session.AddPeerID.
func (a *payChAPIServer) AddPeerID(ctx context.Context, req *pb.AddPeerIDReq) (*pb.AddPeerIDResp, error) {
	errResponse := func(err perun.APIError) *pb.AddPeerIDResp {
		return &pb.AddPeerIDResp{
			Response: &pb.AddPeerIDResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	err = sess.AddPeerID(perun.PeerID{
		Alias:              req.PeerID.Alias,
		OffChainAddrString: req.PeerID.OffChainAddress,
		CommAddr:           req.PeerID.CommAddress,
		CommType:           req.PeerID.CommType,
	})
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.AddPeerIDResp{
		Response: &pb.AddPeerIDResp_MsgSuccess_{
			MsgSuccess: &pb.AddPeerIDResp_MsgSuccess{
				Success: true,
			},
		},
	}, nil
}

// GetPeerID wraps session.GetPeerID.
func (a *payChAPIServer) GetPeerID(ctx context.Context, req *pb.GetPeerIDReq) (*pb.GetPeerIDResp, error) {
	errResponse := func(err perun.APIError) *pb.GetPeerIDResp {
		return &pb.GetPeerIDResp{
			Response: &pb.GetPeerIDResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	peerID, err := sess.GetPeerID(req.Alias)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.GetPeerIDResp{
		Response: &pb.GetPeerIDResp_MsgSuccess_{
			MsgSuccess: &pb.GetPeerIDResp_MsgSuccess{
				PeerID: &pb.PeerID{
					Alias:           peerID.Alias,
					OffChainAddress: peerID.OffChainAddrString,
					CommAddress:     peerID.CommAddr,
					CommType:        peerID.CommType,
				},
			},
		},
	}, nil
}

// OpenPayCh wraps payment.OpenPayCh.
func (a *payChAPIServer) OpenPayCh(ctx context.Context, req *pb.OpenPayChReq) (*pb.OpenPayChResp, error) {
	errResponse := func(err perun.APIError) *pb.OpenPayChResp {
		return &pb.OpenPayChResp{
			Response: &pb.OpenPayChResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	openingBalInfo := fromGrpcBalInfo(req.OpeningBalInfo)
	payChInfo, err := payment.OpenPayCh(ctx, sess, openingBalInfo, req.ChallengeDurSecs)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.OpenPayChResp{
		Response: &pb.OpenPayChResp_MsgSuccess_{
			MsgSuccess: &pb.OpenPayChResp_MsgSuccess{
				OpenedPayChInfo: &pb.PayChInfo{
					ChID:    payChInfo.ChID,
					BalInfo: toGrpcBalInfo(payChInfo.BalInfo),
					Version: payChInfo.Version,
				},
			},
		},
	}, nil
}

// GetPayChsInfo wraps payment.GetPayChs.
func (a *payChAPIServer) GetPayChsInfo(ctx context.Context, req *pb.GetPayChsInfoReq) (*pb.GetPayChsInfoResp, error) {
	errResponse := func(err perun.APIError) *pb.GetPayChsInfoResp {
		return &pb.GetPayChsInfoResp{
			Response: &pb.GetPayChsInfoResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	openPayChsInfo := payment.GetPayChsInfo(sess)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.GetPayChsInfoResp{
		Response: &pb.GetPayChsInfoResp_MsgSuccess_{
			MsgSuccess: &pb.GetPayChsInfoResp_MsgSuccess{
				OpenPayChsInfo: toGrpcPayChsInfo(openPayChsInfo),
			},
		},
	}, nil
}

// SubPayChProposals wraps payment.SubPayChProposals.
func (a *payChAPIServer) SubPayChProposals(req *pb.SubPayChProposalsReq,
	srv pb.Payment_API_SubPayChProposalsServer) error {
	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		// TODO: (mano) Return a error response and not a protocol error
		return errors.WithMessage(err, "cannot register subscription")
	}

	notifier := func(notif payment.PayChProposalNotif) {
		err := srv.Send(&pb.SubPayChProposalsResp{Response: &pb.SubPayChProposalsResp_Notify_{
			Notify: &pb.SubPayChProposalsResp_Notify{
				ProposalID:       notif.ProposalID,
				OpeningBalInfo:   toGrpcBalInfo(notif.OpeningBalInfo),
				ChallengeDurSecs: notif.ChallengeDurSecs,
				Expiry:           notif.Expiry,
			},
		}})
		_ = err
		// if err != nil {
		// TODO: (mano) Handle error while sending.
		// }
	}
	err = payment.SubPayChProposals(sess, notifier)
	if err != nil {
		// TODO: (mano) Return a error response and not a protocol error
		return errors.WithMessage(err, "cannot register subscription")
	}

	signal := make(chan bool)
	a.Lock()
	a.chProposalsNotif[req.SessionID] = signal
	a.Unlock()

	<-signal
	return nil
}

// UnsubPayChProposals wraps payment.UnsubPayChProposals.
func (a *payChAPIServer) UnsubPayChProposals(ctx context.Context, req *pb.UnsubPayChProposalsReq) (
	*pb.UnsubPayChProposalsResp, error) {
	errResponse := func(err perun.APIError) *pb.UnsubPayChProposalsResp {
		return &pb.UnsubPayChProposalsResp{
			Response: &pb.UnsubPayChProposalsResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	err = payment.UnsubPayChProposals(sess)
	if err != nil {
		return errResponse(err), nil
	}

	a.closeGrpcPayChProposalSub(req.SessionID)

	return &pb.UnsubPayChProposalsResp{
		Response: &pb.UnsubPayChProposalsResp_MsgSuccess_{
			MsgSuccess: &pb.UnsubPayChProposalsResp_MsgSuccess{
				Success: true,
			},
		},
	}, nil
}

func (a *payChAPIServer) closeGrpcPayChProposalSub(sessionID string) {
	a.Lock()
	signal := a.chProposalsNotif[sessionID]
	delete(a.chProposalsNotif, sessionID)
	a.Unlock()
	close(signal)
}

// RespondPayChProposal wraps payment.RespondPayChProposal.
func (a *payChAPIServer) RespondPayChProposal(ctx context.Context, req *pb.RespondPayChProposalReq) (
	*pb.RespondPayChProposalResp, error) {
	errResponse := func(err perun.APIError) *pb.RespondPayChProposalResp {
		return &pb.RespondPayChProposalResp{
			Response: &pb.RespondPayChProposalResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	openedPayChInfo, err := payment.RespondPayChProposal(ctx, sess, req.ProposalID, req.Accept)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.RespondPayChProposalResp{
		Response: &pb.RespondPayChProposalResp_MsgSuccess_{
			MsgSuccess: &pb.RespondPayChProposalResp_MsgSuccess{
				OpenedPayChInfo: toGrpcPayChInfo(openedPayChInfo),
			},
		},
	}, nil
}

// CloseSession wraps payment.CloseSession. For now, this is a stub.
func (a *payChAPIServer) CloseSession(ctx context.Context, req *pb.CloseSessionReq) (*pb.CloseSessionResp, error) {
	errResponse := func(err perun.APIError) *pb.CloseSessionResp {
		return &pb.CloseSessionResp{
			Response: &pb.CloseSessionResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	openPayChsInfo, err := payment.CloseSession(sess, req.Force)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.CloseSessionResp{
		Response: &pb.CloseSessionResp_MsgSuccess_{
			MsgSuccess: &pb.CloseSessionResp_MsgSuccess{
				OpenPayChsInfo: toGrpcPayChsInfo(openPayChsInfo),
			},
		},
	}, nil
}

// SendPayChUpdate wraps payment.SendPayChUpdate.
func (a *payChAPIServer) SendPayChUpdate(ctx context.Context, req *pb.SendPayChUpdateReq) (
	*pb.SendPayChUpdateResp, error) {
	errResponse := func(err perun.APIError) *pb.SendPayChUpdateResp {
		return &pb.SendPayChUpdateResp{
			Response: &pb.SendPayChUpdateResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errResponse(err), nil
	}
	updatedPayChInfo, err := payment.SendPayChUpdate(ctx, ch, req.Payee, req.Amount)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.SendPayChUpdateResp{
		Response: &pb.SendPayChUpdateResp_MsgSuccess_{
			MsgSuccess: &pb.SendPayChUpdateResp_MsgSuccess{
				UpdatedPayChInfo: toGrpcPayChInfo(updatedPayChInfo),
			},
		},
	}, nil
}

// SubPayChUpdates wraps payment.SubPayChUpdates.
func (a *payChAPIServer) SubPayChUpdates(req *pb.SubpayChUpdatesReq, srv pb.Payment_API_SubPayChUpdatesServer) error {
	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		// TODO: (mano) Return a error response and not a protocol error.
		return errors.WithMessage(err, "cannot register subscription")
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errors.WithMessage(err, "cannot register subscription")
	}

	notifier := func(notif payment.PayChUpdateNotif) {
		var notifErr *pb.MsgError
		if notif.Error != nil {
			notifErr = toGrpcError(notif.Error)
		}

		err := srv.Send(&pb.SubPayChUpdatesResp{Response: &pb.SubPayChUpdatesResp_Notify_{
			Notify: &pb.SubPayChUpdatesResp_Notify{
				UpdateID:          notif.UpdateID,
				ProposedPayChInfo: toGrpcPayChInfo(notif.ProposedPayChInfo),
				Type:              ToGrpcChUpdateType[notif.Type],
				Expiry:            notif.Expiry,
				Error:             notifErr,
			},
		}})
		_ = err
		// if err != nil {
		// 	// TODO: (mano) Error handling when sending notification.
		// }

		// Close grpc subscription function (SubPayChUpdates) that will be running in the background.
		if perun.ChUpdateTypeClosed == notif.Type {
			a.closeGrpcPayChUpdateSub(req.SessionID, req.ChID)
		}
	}
	err = payment.SubPayChUpdates(ch, notifier)
	if err != nil {
		// TODO: (mano) Error handling when sending notification.
		return errors.WithMessage(err, "cannot register subscription")
	}

	signal := make(chan bool)
	a.Lock()
	a.chUpdatesNotif[req.SessionID][req.ChID] = signal
	a.Unlock()

	<-signal
	return nil
}

// ToGrpcChUpdateType is a helper var that maps enums from ChUpdateType type defined in perun-node
// to ChUpdateType type defined in grpc package.
var ToGrpcChUpdateType = map[perun.ChUpdateType]pb.SubPayChUpdatesResp_Notify_ChUpdateType{
	perun.ChUpdateTypeOpen:   pb.SubPayChUpdatesResp_Notify_open,
	perun.ChUpdateTypeFinal:  pb.SubPayChUpdatesResp_Notify_final,
	perun.ChUpdateTypeClosed: pb.SubPayChUpdatesResp_Notify_closed,
}

// UnsubPayChUpdates wraps payment.UnsubPayChUpdates.
func (a *payChAPIServer) UnsubPayChUpdates(ctx context.Context, req *pb.UnsubPayChUpdatesReq) (
	*pb.UnsubPayChUpdatesResp, error) {
	errResponse := func(err perun.APIError) *pb.UnsubPayChUpdatesResp {
		return &pb.UnsubPayChUpdatesResp{
			Response: &pb.UnsubPayChUpdatesResp_Error{
				Error: toGrpcError(err),
			},
		}
	}
	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errResponse(err), nil
	}
	err = payment.UnsubPayChUpdates(ch)
	if err != nil {
		return errResponse(err), nil
	}
	a.closeGrpcPayChUpdateSub(req.SessionID, req.ChID)

	return &pb.UnsubPayChUpdatesResp{
		Response: &pb.UnsubPayChUpdatesResp_MsgSuccess_{
			MsgSuccess: &pb.UnsubPayChUpdatesResp_MsgSuccess{
				Success: true,
			},
		},
	}, nil
}

func (a *payChAPIServer) closeGrpcPayChUpdateSub(sessionID, chID string) {
	a.Lock()
	signal := a.chUpdatesNotif[sessionID][chID]
	delete(a.chUpdatesNotif[sessionID], chID)
	a.Unlock()
	close(signal)
}

// RespondPayChUpdate wraps payment.RespondPayChUpdate.
func (a *payChAPIServer) RespondPayChUpdate(ctx context.Context, req *pb.RespondPayChUpdateReq) (
	*pb.RespondPayChUpdateResp, error) {
	errResponse := func(err perun.APIError) *pb.RespondPayChUpdateResp {
		return &pb.RespondPayChUpdateResp{
			Response: &pb.RespondPayChUpdateResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errResponse(err), nil
	}
	updatedPayChInfo, err := payment.RespondPayChUpdate(ctx, ch, req.UpdateID, req.Accept)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.RespondPayChUpdateResp{
		Response: &pb.RespondPayChUpdateResp_MsgSuccess_{
			MsgSuccess: &pb.RespondPayChUpdateResp_MsgSuccess{
				UpdatedPayChInfo: toGrpcPayChInfo(updatedPayChInfo),
			},
		},
	}, nil
}

// GetPayChInfo wraps payment.GetBalInfo.
func (a *payChAPIServer) GetPayChInfo(ctx context.Context, req *pb.GetPayChInfoReq) (
	*pb.GetPayChInfoResp, error) {
	errResponse := func(err perun.APIError) *pb.GetPayChInfoResp {
		return &pb.GetPayChInfoResp{
			Response: &pb.GetPayChInfoResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errResponse(err), nil
	}
	payChInfo := payment.GetPayChInfo(ch)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.GetPayChInfoResp{
		Response: &pb.GetPayChInfoResp_MsgSuccess_{
			MsgSuccess: &pb.GetPayChInfoResp_MsgSuccess{
				PayChInfo: toGrpcPayChInfo(payChInfo),
			},
		},
	}, nil
}

// ClosePayCh wraps payment.ClosePayCh.
func (a *payChAPIServer) ClosePayCh(ctx context.Context, req *pb.ClosePayChReq) (*pb.ClosePayChResp, error) {
	errResponse := func(err perun.APIError) *pb.ClosePayChResp {
		return &pb.ClosePayChResp{
			Response: &pb.ClosePayChResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	ch, err := sess.GetCh(req.ChID)
	if err != nil {
		return errResponse(err), nil
	}
	closedPayChInfo, err := payment.ClosePayCh(ctx, ch)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.ClosePayChResp{
		Response: &pb.ClosePayChResp_MsgSuccess_{
			MsgSuccess: &pb.ClosePayChResp_MsgSuccess{
				ClosedPayChInfo: toGrpcPayChInfo(closedPayChInfo),
			},
		},
	}, nil
}

// toGrpcPayChInfo is a helper function to convert slice of PayChInfo struct defined in perun-node
// to a slice of PayChInfo struct defined in grpc package.
func toGrpcPayChsInfo(payChsInfo []payment.PayChInfo) []*pb.PayChInfo {
	grpcPayChsInfo := make([]*pb.PayChInfo, len(payChsInfo))
	for i := range payChsInfo {
		grpcPayChsInfo[i] = toGrpcPayChInfo(payChsInfo[i])
	}
	return grpcPayChsInfo
}

// toGrpcPayChInfo is a helper function to convert PayChInfo struct defined in perun-node
// to PayChInfo struct defined in grpc package.
func toGrpcPayChInfo(src payment.PayChInfo) *pb.PayChInfo {
	return &pb.PayChInfo{
		ChID:    src.ChID,
		BalInfo: toGrpcBalInfo(src.BalInfo),
		Version: src.Version,
	}
}

// fromGrpcBalInfo is a helper function to convert BalInfo struct defined in grpc package
// to BalInfo struct defined in perun-node.
func fromGrpcBalInfo(src *pb.BalInfo) perun.BalInfo {
	return perun.BalInfo{
		Currency: src.Currency,
		Parts:    src.Parts,
		Bal:      src.Bal,
	}
}

// toGrpcBalInfo is a helper function to convert BalInfo struct defined in perun-node
// to BalInfo struct defined in grpc package.
func toGrpcBalInfo(src perun.BalInfo) *pb.BalInfo {
	return &pb.BalInfo{
		Currency: src.Currency,
		Parts:    src.Parts,
		Bal:      src.Bal,
	}
}

// toGrpcError is a helper function to convert APIError struct defined in perun-node
// to APIError struct defined in grpc package.
func toGrpcError(err perun.APIError) *pb.MsgError { //nolint: funlen
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
				Chs: toGrpcPayChsInfo(info.PayChs),
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
				ContractErrInfos: toGrpcContractErrInfos(info.ContractErrInfos),
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

// toGrpcContractErrInfos is a helper function to convert a slice of
// ContractErrInfo struct defined in perun-node to a slice of ContractErrInfo
// struct defined in grpc package.
func toGrpcContractErrInfos(src []perun.ContractErrInfo) []*pb.ContractErrInfo {
	output := make([]*pb.ContractErrInfo, len(src))
	for i := range src {
		output[i].Name = src[i].Name
		output[i].Address = src[i].Address
		output[i].Error = src[i].Error
	}
	return output
}
