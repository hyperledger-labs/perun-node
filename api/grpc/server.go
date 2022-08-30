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
	"fmt"
	"math"
	"math/big"
	"net"

	"github.com/pkg/errors"
	grpclib "google.golang.org/grpc"
	pethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/channel"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	pwallet "perun.network/go-perun/wallet"
	psync "polycry.pt/poly-go/sync"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/app/payment"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/currency"
)

var withdrawn = false

// payChAPIServer represents a grpc server that can serve payment channel API.
type payChAPIServer struct {
	// Embeddeing this types is required by generated gRPC stubs.
	pb.UnimplementedPayment_APIServer

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

	subscribes map[string]map[pchannel.ID]pchannel.AdjudicatorSubscription

	// chWatcher map[string]map[string]chan bool

	APIServerEnabled bool

	currRegistry perun.CurrencyRegistry
	partsMap     map[string]string
}

// ListenAndServePayChAPI starts a payment channel API server that listens for incoming grpc
// requests at the specified address and serves those requests using the node API instance.
func ListenAndServePayChAPI(n perun.NodeAPI, grpcPort string, apiServerOnly bool) error {
	apiServer := &payChAPIServer{
		n:                n,
		chProposalsNotif: make(map[string]chan bool),
		chUpdatesNotif:   make(map[string]map[string]chan bool),
		subscribes:       make(map[string]map[pchannel.ID]pchannel.AdjudicatorSubscription),
		currRegistry:     currency.NewRegistry(),
		partsMap:         make(map[string]string),
		APIServerEnabled: apiServerOnly,
	}

	if apiServerOnly {
		D = InitDashboard()
	}

	walletBackend := ethereum.NewWalletBackend()
	carAddr, err := walletBackend.ParseAddr("0x7b7E212652b9C3755C4E1f1718a142dDE3817523")
	if err != nil {
		return errors.WithMessage(err, "parsing car address")
	}
	chargerAddr, err := walletBackend.ParseAddr("0xa617fa2cc5eC8d72d4A60b9F424677e74E6bef68")
	if err != nil {
		return errors.WithMessage(err, "parsing charger ETH address")
	}
	apiServer.partsMap[carAddr.String()] = "car"
	apiServer.partsMap[chargerAddr.String()] = "charger"

	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return errors.Wrap(err, "starting listener")
	}
	grpcServer := grpclib.NewServer()
	pb.RegisterPayment_APIServer(grpcServer, apiServer)

	apiServer.currRegistry.Register(currency.ETHSymbol, currency.ETHMaxDecimals)

	return grpcServer.Serve(listener)
}

// GetConfig wraps node.GetConfig.
func (a *payChAPIServer) GetConfig(context.Context, *pb.GetConfigReq) (*pb.GetConfigResp, error) {
	cfg := a.n.GetConfig()
	return &pb.GetConfigResp{
		ChainAddress:    cfg.ChainURL,
		Adjudicator:     cfg.Adjudicator,
		AssetETH:        cfg.AssetETH,
		CommTypes:       cfg.CommTypes,
		IdProviderTypes: cfg.IDProviderTypes,
	}, nil
}

// Time wraps node.Time.
func (a *payChAPIServer) Time(context.Context, *pb.TimeReq) (*pb.TimeResp, error) {
	return &pb.TimeResp{
		Time: a.n.Time(),
	}, nil
}

// RegisterCurrency wraps node.RegisterCurrency.
func (a *payChAPIServer) RegisterCurrency(ctx context.Context, req *pb.RegisterCurrencyReq) (
	*pb.RegisterCurrencyResp, error) {
	errResponse := func(err perun.APIError) *pb.RegisterCurrencyResp {
		return &pb.RegisterCurrencyResp{
			Response: &pb.RegisterCurrencyResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	symbol, err := a.n.RegisterCurrency(req.TokenAddr, req.AssetAddr)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.RegisterCurrencyResp{
		Response: &pb.RegisterCurrencyResp_MsgSuccess_{
			MsgSuccess: &pb.RegisterCurrencyResp_MsgSuccess{
				Symbol: symbol,
			},
		},
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
	a.subscribes[sessionID] = make(map[pchannel.ID]pchannel.AdjudicatorSubscription)
	// a.chWatcher[sessionID] = make(map[string]chan bool)
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
					BalInfo: ToGrpcBalInfo(payChInfo.BalInfo),
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
				OpeningBalInfo:   ToGrpcBalInfo(notif.OpeningBalInfo),
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

// CloseSession wraps payment.CloseSession.
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

// DeployAssetERC20 wraps session.DeployAssetERC20.
func (a *payChAPIServer) DeployAssetERC20(ctx context.Context, req *pb.DeployAssetERC20Req) (
	*pb.DeployAssetERC20Resp, error) {
	errResponse := func(err perun.APIError) *pb.DeployAssetERC20Resp {
		return &pb.DeployAssetERC20Resp{
			Response: &pb.DeployAssetERC20Resp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	assetAddr, err := sess.DeployAssetERC20(req.TokenAddr)
	if err != nil {
		return errResponse(err), nil
	}

	return &pb.DeployAssetERC20Resp{
		Response: &pb.DeployAssetERC20Resp_MsgSuccess_{
			MsgSuccess: &pb.DeployAssetERC20Resp_MsgSuccess{
				AssetAddr: assetAddr,
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
	updatedPayChInfo, err := payment.SendPayChUpdate(ctx, ch, fromGrpcPayments(req.Payments))
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

// Fund wraps session.Fund.
func (a *payChAPIServer) Fund(ctx context.Context, req *pb.FundReq) (*pb.FundResp, error) {
	errResponse := func(err perun.APIError) *pb.FundResp {
		return &pb.FundResp{
			Error: toGrpcError(err),
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	req2, err2 := fromGrpcFundingReq(req)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}

	if a.APIServerEnabled {
		balInfo := a.GetBalInfo(req2.State, req2.Params)
		D.FundingRequest(balInfo.Parts, balInfo.Bals[0])
	}

	err2 = sess.Fund(ctx, req2)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}

	if a.APIServerEnabled {
		withdrawn = false
		D.FundingSuccessful()
		D.PrintBlank()
	}
	return &pb.FundResp{
		Error: nil,
	}, nil
}

func (a *payChAPIServer) GetBalInfo(state *pchannel.State, params *pchannel.Params) perun.BalInfo {
	currencies := []perun.Currency{a.currRegistry.Currency(currency.ETHSymbol)}
	parts := make([]string, len(params.Parts))
	for i := range params.Parts {
		parts[i] = a.partsMap[params.Parts[i].String()]
	}
	return makeBalInfoFromState(parts, currencies, state)
}

// RegisterAssetERC20 is a stub that always returns false. Because, the remote
// funder does not support use of assets other than the default ERC20 asset.
//
// TODO: Make actual implementation.
func (a *payChAPIServer) RegisterAssetERC20(ctx context.Context, req *pb.RegisterAssetERC20Req) (
	*pb.RegisterAssetERC20Resp, error) {
	return &pb.RegisterAssetERC20Resp{
		MsgSuccess: false,
	}, nil
}

// IsAssetRegistered wraps session.IsAssetRegistered.
func (a *payChAPIServer) IsAssetRegistered(ctx context.Context, req *pb.IsAssetRegisteredReq) (*pb.IsAssetRegisteredResp, error) {
	errResponse := func(err perun.APIError) *pb.IsAssetRegisteredResp {
		return &pb.IsAssetRegisteredResp{
			Response: &pb.IsAssetRegisteredResp_Error{
				Error: toGrpcError(err),
			},
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	asset := pchannel.NewAsset()
	err2 := asset.UnmarshalBinary(req.Asset)
	if err2 != nil {
		err = perun.NewAPIErrInvalidArgument(err2, "asset", fmt.Sprintf("%x", req.Asset))
		return errResponse(err), nil
	}

	isRegistered := sess.IsAssetRegistered(asset)

	return &pb.IsAssetRegisteredResp{
		Response: &pb.IsAssetRegisteredResp_MsgSuccess_{
			MsgSuccess: &pb.IsAssetRegisteredResp_MsgSuccess{
				IsRegistered: isRegistered,
			},
		},
	}, nil
}

// Register wraps session.Register.
func (a *payChAPIServer) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	errResponse := func(err perun.APIError) *pb.RegisterResp {
		return &pb.RegisterResp{
			Error: toGrpcError(err),
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	adjReq, err2 := fromGrpcAdjReq(req.AdjReq)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}
	signedStates := make([]pchannel.SignedState, len(req.SignedStates))
	for i := range signedStates {
		signedStates[i], err2 = fromGrpcSignedState(req.SignedStates[i])
		if err2 != nil {
			return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
		}
	}

	err2 = sess.Register(ctx, adjReq, signedStates)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}

	return &pb.RegisterResp{
		Error: nil,
	}, nil
}

// Withdraw wraps session.Withdraw.
func (a *payChAPIServer) Withdraw(ctx context.Context, req *pb.WithdrawReq) (*pb.WithdrawResp, error) {
	errResponse := func(err perun.APIError) *pb.WithdrawResp {
		return &pb.WithdrawResp{
			Error: toGrpcError(err),
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	adjReq, err2 := fromGrpcAdjReq(req.AdjReq)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}
	if a.APIServerEnabled && !withdrawn {
		balInfo := a.GetBalInfo(adjReq.Tx.State, adjReq.Params)
		D.WithdrawRequest(balInfo.Parts, balInfo.Bals[0])
	}
	stateMap := channel.StateMap(make(map[pchannel.ID]*pchannel.State))

	for i := range req.StateMap {
		var id pchannel.ID
		copy(id[:], req.StateMap[i].Id)
		stateMap[id], err2 = fromGrpcState(req.StateMap[i].State)
		if err2 != nil {
			return errResponse(err), nil
		}
	}

	err2 = sess.Withdraw(ctx, adjReq, stateMap)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}

	if a.APIServerEnabled && !withdrawn {
		D.WithdrawSuccessful()
		D.PrintBlank()
		withdrawn = true
	}
	return &pb.WithdrawResp{
		Error: nil,
	}, nil
}

// Progress wraps session.Progress.
func (a *payChAPIServer) Progress(ctx context.Context, req *pb.ProgressReq) (*pb.ProgressResp, error) {
	errResponse := func(err perun.APIError) *pb.ProgressResp {
		return &pb.ProgressResp{
			Error: toGrpcError(err),
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	var progReq perun.ProgressReq
	var err2 error
	progReq.AdjudicatorReq, err2 = fromGrpcAdjReq(req.AdjReq)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}
	progReq.NewState, err2 = fromGrpcState(req.NewState)
	if err2 != nil {
		return errResponse(err), nil
	}
	copy(progReq.Sig, req.Sig)

	err2 = sess.Progress(ctx, progReq)
	if err2 != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(err2)), nil
	}

	return &pb.ProgressResp{
		Error: nil,
	}, nil
}

// Subscribe wraps session.Subscribe.

func (a *payChAPIServer) Subscribe(req *pb.SubscribeReq, stream pb.Payment_API_SubscribeServer) error {
	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errors.WithMessage(err, "retrieving session")
	}

	var chID pchannel.ID
	copy(chID[:], req.ChID)

	adjSub, err := sess.Subscribe(context.Background(), chID)
	if err != nil {
		return errors.WithMessage(err, "setting up subscription")
	}
	a.Lock()
	a.subscribes[req.SessionID][chID] = adjSub
	a.Unlock()

	// This stream is anyways closed when StopWatching is called for.
	// Hence, that will act as the exit condition for the loop.
	go func() {
		// will return nil, when the sub is closed.
		// so, we need a mechanism to call close on the server side.
		// so, add a call Unsubscribe, which simply calls close.
		for {
			adjEvent := adjSub.Next()
			if adjEvent == nil {
				err := errors.WithMessage(adjSub.Err(), "sub closed with error")
				notif := &pb.SubscribeResp_Error{
					Error: toGrpcError(perun.NewAPIErrUnknownInternal(err)),
				}
				stream.Send(&pb.SubscribeResp{Response: notif})
				return
			}
			notif, err := toSubscribeResponse(adjEvent)
			if err != nil {
				return
			}
			err = stream.Send(notif)
			if err != nil {
				return
			}
		}
	}()

	return nil
}

func (a *payChAPIServer) Unsubscribe(ctx context.Context, req *pb.UnsubscribeReq) (*pb.UnsubscribeResp, error) {
	errResponse := func(err perun.APIError) *pb.UnsubscribeResp {
		return &pb.UnsubscribeResp{
			Error: toGrpcError(err),
		}
	}

	var chID pchannel.ID
	copy(chID[:], req.ChID)

	a.Lock()
	adjSub := a.subscribes[req.SessionID][chID]
	delete(a.subscribes[req.SessionID], chID)
	a.Unlock()

	if err := adjSub.Close(); err != nil {
		return errResponse(perun.NewAPIErrUnknownInternal(errors.WithMessage(err, "retrieving session"))), nil

	}

	return &pb.UnsubscribeResp{
		Error: nil,
	}, nil
}

// StartWatchingLedgerChannel wraps session.StartWatchingLedgerChannel.
func (a *payChAPIServer) StartWatchingLedgerChannel(srv pb.Payment_API_StartWatchingLedgerChannelServer) error {
	req, err := srv.Recv()
	if err != nil {
		return errors.WithMessage(err, "reading request data")
	}
	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errors.WithMessage(err, "retrieving session")
	}

	signedState, err := toSignedState(req)
	if err != nil {
		return errors.WithMessage(err, "parsing signed state")
	}

	if a.APIServerEnabled {
		// balInfo := a.GetBalInfo(signedState.State, signedState.Params)
		D.WatchingRequest()
		// D.WatchingRequest(balInfo.Parts, balInfo.Bals[0])
	}

	statesPub, adjSub, err := sess.StartWatchingLedgerChannel(context.TODO(), *signedState)
	if err != nil {
		return errors.WithMessage(err, "start watching")
	}
	if a.APIServerEnabled {
		D.WatchingSuccessful()
	}

	// This stream is anyways closed when StopWatching is called for.
	// Hence, that will act as the exit condition for the loop.
	go func() {
		adjEventStream := adjSub.EventStream()
		for {
			select {
			case adjEvent, isOpen := <-adjEventStream:
				if !isOpen {
					return
				}
				if a.APIServerEnabled {
					switch e := adjEvent.(type) {
					case *pchannel.RegisteredEvent:
						balInfo := a.GetBalInfo(e.State, signedState.Params)
						D.ChannelRegistered(balInfo.Parts, balInfo.Bals[0])
					case *pchannel.ConcludedEvent:
						D.PrintBlank()
						D.ChannelConcluded()
						D.PrintBlank()
					}
				}
				protoResponse, err := toProtoResponse(adjEvent)
				if err != nil {
					return
				}

				// Handle error while sending notification.
				err = srv.Send(protoResponse)
				if err != nil {
					return
				}
			}
		}
	}()

	// signal := make(chan bool)
	// a.Lock()
	// a.chWatcher[req.SessionID][chID] = signal
	// a.Unlock()

	// It should be the responsibility of the streamer to close things.
	// Hence, the client should be closing this stream, which will cause srv.Recv to return an error.
	// The error will act as the exit condition for this for{} loop.
StatesPubLoop:
	for {
		req, err = srv.Recv()
		if err != nil {
			err = errors.WithMessage(err, "reading published states pub data")
			break StatesPubLoop
		}
		tx, err := toTransaction(req)
		if err != nil {
			err = errors.WithMessage(err, "parsing published states pub data")
			break StatesPubLoop
		}

		if a.APIServerEnabled {
			balInfo := a.GetBalInfo(tx.State, signedState.Params)
			if !tx.State.IsFinal {
				D.ChannelUpdated(balInfo.Parts, balInfo.Bals[0])
			} else {
				D.ChannelFinalized(balInfo.Parts, balInfo.Bals[0])
			}
		}
		err = statesPub.Publish(context.TODO(), *tx)
		if err != nil {
			err = errors.WithMessage(err, "locally relaying published states pub data")
			break StatesPubLoop
		}
	}

	// TODO: Ensure adjEventSteam go-routine is killed. It should not be leaking.
	// It is to allow for that, label breaks are used instead of return statements in the above for-select.
	return nil
}

// StopWatching wraps session.StopWatching.
func (a *payChAPIServer) StopWatching(ctx context.Context, req *pb.StopWatchingReq) (*pb.StopWatchingResp, error) {
	errResponse := func(err perun.APIError) *pb.StopWatchingResp {
		return &pb.StopWatchingResp{
			Error: toGrpcError(err),
		}
	}

	sess, err := a.n.GetSession(req.SessionID)
	if err != nil {
		return errResponse(err), nil
	}
	var chID pchannel.ID
	copy(chID[:], req.ChID)
	err2 := sess.StopWatching(ctx, chID)
	if err2 != nil {
		return errResponse(err), nil
	}

	return &pb.StopWatchingResp{Error: nil}, nil
}

func toSubscribeResponse(adjEvent pchannel.AdjudicatorEvent) (*pb.SubscribeResp, error) {
	protoResponse := &pb.SubscribeResp{}
	switch e := adjEvent.(type) {
	case *pchannel.RegisteredEvent:
		registeredEvent, err := toGrpcRegisteredEvent(e)
		protoResponse.Response = &pb.SubscribeResp_RegisteredEvent{&registeredEvent}
		return protoResponse, err
	case *pchannel.ProgressedEvent:
		progressedEvent, err := toGrpcProgressedEvent(e)
		protoResponse.Response = &pb.SubscribeResp_ProgressedEvent{&progressedEvent}
		return protoResponse, err
	case *pchannel.ConcludedEvent:
		concludedEvent, err := toGrpcConcludedEvent(e)
		protoResponse.Response = &pb.SubscribeResp_ConcludedEvent{&concludedEvent}
		return protoResponse, err
	default:
		apiErr := perun.NewAPIErrUnknownInternal(errors.New("unknown even type"))
		protoResponse.Response = &pb.SubscribeResp_Error{
			Error: toGrpcError(apiErr),
		}
		return protoResponse, nil
	}
}

func toProtoResponse(adjEvent pchannel.AdjudicatorEvent) (*pb.StartWatchingLedgerChannelResp, error) {
	protoResponse := &pb.StartWatchingLedgerChannelResp{}
	switch e := adjEvent.(type) {
	case *pchannel.RegisteredEvent:
		registeredEvent, err := toGrpcRegisteredEvent(e)
		protoResponse.Response = &pb.StartWatchingLedgerChannelResp_RegisteredEvent{&registeredEvent}
		return protoResponse, err
	case *pchannel.ProgressedEvent:
		progressedEvent, err := toGrpcProgressedEvent(e)
		protoResponse.Response = &pb.StartWatchingLedgerChannelResp_ProgressedEvent{&progressedEvent}
		return protoResponse, err
	case *pchannel.ConcludedEvent:
		concludedEvent, err := toGrpcConcludedEvent(e)
		protoResponse.Response = &pb.StartWatchingLedgerChannelResp_ConcludedEvent{&concludedEvent}
		return protoResponse, err
	default:
		apiErr := perun.NewAPIErrUnknownInternal(errors.New("unknown even type"))
		protoResponse.Response = &pb.StartWatchingLedgerChannelResp_Error{
			Error: toGrpcError(apiErr),
		}
		return protoResponse, nil
	}
}

func toGrpcRegisteredEvent(event *pchannel.RegisteredEvent) (grpcEvent pb.RegisteredEvent, err error) {
	grpcEvent.AdjudicatorEventBase = toGrpcAdjudicatorEventBase(&event.AdjudicatorEventBase)
	grpcEvent.Sigs = make([][]byte, len(event.Sigs))
	for i := range event.Sigs {
		grpcEvent.Sigs[i] = []byte(event.Sigs[i])
	}
	grpcEvent.State, err = toGrpcState(event.State)
	return grpcEvent, errors.WithMessage(err, "parsing state")
}

func toGrpcProgressedEvent(event *pchannel.ProgressedEvent) (grpcEvent pb.ProgressedEvent, err error) {
	grpcEvent.AdjudicatorEventBase = toGrpcAdjudicatorEventBase(&event.AdjudicatorEventBase)
	grpcEvent.Idx = uint32(event.Idx)
	grpcEvent.State, err = toGrpcState(event.State)
	return grpcEvent, errors.WithMessage(err, "parsing state")
}

func toGrpcConcludedEvent(event *pchannel.ConcludedEvent) (grpcEvent pb.ConcludedEvent, err error) {
	grpcEvent.AdjudicatorEventBase = toGrpcAdjudicatorEventBase(&event.AdjudicatorEventBase)
	return grpcEvent, errors.WithMessage(err, "parsing adjudicator event base")
}

func toGrpcAdjudicatorEventBase(event *pchannel.AdjudicatorEventBase) (protoEvent *pb.AdjudicatorEventBase) {
	// Does a type switch on the underlying timeout type, because timeout cannot be passed as such
	// TODO: Make timeout wire friendly.
	protoEvent = &pb.AdjudicatorEventBase{}
	protoEvent.ChID = event.IDV[:]
	protoEvent.Version = event.VersionV
	protoEvent.Timeout = &pb.AdjudicatorEventBase_Timeout{}
	switch t := event.TimeoutV.(type) {
	case *pchannel.ElapsedTimeout:
		protoEvent.Timeout.Sec = -1
		protoEvent.Timeout.Type = pb.AdjudicatorEventBase_elapsed
	case *pchannel.TimeTimeout:
		protoEvent.Timeout.Sec = t.Unix()
		protoEvent.Timeout.Type = pb.AdjudicatorEventBase_time
	case *pethchannel.BlockTimeout:
		// TODO: Validate if number is less than int64max before type casting.
		protoEvent.Timeout.Sec = int64(t.Time)
		protoEvent.Timeout.Type = pb.AdjudicatorEventBase_ethBlock
	}
	return protoEvent
}

func toGrpcState(state *channel.State) (protoState *pb.State, err error) {
	protoState = &pb.State{}

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

func fromAppAndData(app channel.App, data channel.Data) (protoApp, protoData []byte, err error) {
	if channel.IsNoApp(app) {
		return []byte{}, []byte{}, nil
	}
	protoApp, err = app.Def().MarshalBinary()
	if err != nil {
		return []byte{}, []byte{}, err
	}
	protoData, err = data.MarshalBinary()
	return protoApp, protoData, err
}

func fromAllocation(alloc channel.Allocation) (protoAlloc *pb.Allocation, err error) {
	protoAlloc = &pb.Allocation{}
	protoAlloc.Assets = make([][]byte, len(alloc.Assets))
	for i := range alloc.Assets {
		protoAlloc.Assets[i], err = alloc.Assets[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th asset", i)
		}
	}
	locked := make([]*pb.SubAlloc, len(alloc.Locked))
	for i := range alloc.Locked {
		locked[i], err = fromSubAlloc(alloc.Locked[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th sub alloc", i)
		}
	}
	protoAlloc.Balances, err = fromBalances(alloc.Balances)
	return protoAlloc, err
}

func fromSubAlloc(subAlloc channel.SubAlloc) (protoSubAlloc *pb.SubAlloc, err error) {
	protoSubAlloc = &pb.SubAlloc{}
	protoSubAlloc.Id = make([]byte, len(subAlloc.ID))
	copy(protoSubAlloc.Id, subAlloc.ID[:])
	protoSubAlloc.IndexMap = &pb.IndexMap{IndexMap: fromIndexMap(subAlloc.IndexMap)}
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

func fromBalances(balances channel.Balances) (protoBalances *pb.Balances, err error) {
	protoBalances = &pb.Balances{
		Balances: make([]*pb.Balance, len(balances)),
	}
	for i := range balances {
		protoBalances.Balances[i], err = fromBalance(balances[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th balance", i)
		}
	}
	return protoBalances, nil
}

func fromBalance(balance []channel.Bal) (protoBalance *pb.Balance, err error) {
	protoBalance = &pb.Balance{
		Balance: make([][]byte, len(balance)),
	}
	for i := range balance {
		if balance[i] == nil {
			return nil, fmt.Errorf("%d'th amount is nil", i) //nolint:goerr113  // We do not want to define this as constant error.
		}
		if balance[i].Sign() == -1 {
			return nil, fmt.Errorf("%d'th amount is negative", i) //nolint:goerr113  // We do not want to define this as constant error.
		}
		protoBalance.Balance[i] = balance[i].Bytes()
	}
	return protoBalance, nil
}

func fromGrpcAdjReq(protoReq *pb.AdjudicatorReq) (req perun.AdjudicatorReq, err error) {
	if req.Params, err = fromGrpcParams(protoReq.Params); err != nil {
		return req, err
	}
	if req.Tx, err = fromGrpcTransaction(protoReq.Tx); err != nil {
		return req, err
	}
	req.Acc = pwallet.NewAddress()
	err = req.Acc.UnmarshalBinary(protoReq.Acc)
	if err != nil {
		return req, err
	}
	req.Idx = pchannel.Index(protoReq.Idx)
	req.Secondary = protoReq.Secondary
	return req, nil
}

func fromGrpcFundingReq(protoReq *pb.FundReq) (req pchannel.FundingReq, err error) {
	if req.Params, err = fromGrpcParams(protoReq.Params); err != nil {
		return req, err
	}
	if req.State, err = fromGrpcState(protoReq.State); err != nil {
		return req, err
	}

	req.Idx = pchannel.Index(protoReq.Idx)
	req.Agreement = fromGrpcBalances(protoReq.Agreement.Balances)
	return req, nil
}

func fromGrpcParams(protoParams *pb.Params) (*channel.Params, error) {
	app, err := toApp(protoParams.App)
	if err != nil {
		return nil, err
	}
	parts, err := toWalletAddrs(protoParams.Parts)
	if err != nil {
		return nil, errors.WithMessage(err, "parts")
	}
	params := channel.NewParamsUnsafe(
		protoParams.ChallengeDuration,
		parts,
		app,
		(new(big.Int)).SetBytes(protoParams.Nonce),
		protoParams.LedgerChannel,
		protoParams.VirtualChannel)

	return params, nil
}

func toApp(protoApp []byte) (app channel.App, err error) {
	if len(protoApp) == 0 {
		app = channel.NoApp()
		return app, nil
	}
	appDef := pwallet.NewAddress()
	err = appDef.UnmarshalBinary(protoApp)
	if err != nil {
		return app, err
	}
	app, err = channel.Resolve(appDef)
	return app, err
}

func toWalletAddrs(protoAddrs [][]byte) ([]pwallet.Address, error) {
	addrs := make([]wallet.Address, len(protoAddrs))
	for i := range protoAddrs {
		addrs[i] = pwallet.NewAddress()
		err := addrs[i].UnmarshalBinary(protoAddrs[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return addrs, nil
}

func fromGrpcSignedState(protoSignedState *pb.SignedState) (signedState channel.SignedState, err error) {
	signedState.Params, err = fromGrpcParams(protoSignedState.Params)
	if err != nil {
		return signedState, err
	}
	signedState.State, err = fromGrpcState(protoSignedState.State)
	if err != nil {
		return signedState, err
	}
	sigs := make([]pwallet.Sig, len(protoSignedState.Sigs))
	for i := range sigs {
		sigs[i] = pwallet.Sig(protoSignedState.Sigs[i])
	}
	return signedState, nil
}

func fromGrpcTransaction(protoSignedState *pb.Transaction) (transaction pchannel.Transaction, err error) {
	transaction.State, err = fromGrpcState(protoSignedState.State)
	if err != nil {
		return transaction, err
	}
	transaction.Sigs = make([]pwallet.Sig, len(protoSignedState.Sigs))
	for i := range transaction.Sigs {
		transaction.Sigs[i] = pwallet.Sig(protoSignedState.Sigs[i])
	}
	return transaction, nil

}

func toSignedState(protoSignedState *pb.StartWatchingLedgerChannelReq) (signedState *channel.SignedState, err error) {
	signedState = &channel.SignedState{}
	signedState.Params, err = fromGrpcParams(protoSignedState.Params)
	if err != nil {
		return nil, err
	}
	signedState.State, err = fromGrpcState(protoSignedState.State)
	if err != nil {
		return nil, err
	}
	sigs := make([]pwallet.Sig, len(protoSignedState.Sigs))
	for i := range sigs {
		sigs[i] = pwallet.Sig(protoSignedState.Sigs[i])
	}
	return signedState, nil

}

func toTransaction(protoSignedState *pb.StartWatchingLedgerChannelReq) (transaction *channel.Transaction, err error) {
	transaction = &channel.Transaction{}
	transaction.State, err = fromGrpcState(protoSignedState.State)
	if err != nil {
		return nil, err
	}
	transaction.Sigs = make([]pwallet.Sig, len(protoSignedState.Sigs))
	for i := range transaction.Sigs {
		transaction.Sigs[i] = pwallet.Sig(protoSignedState.Sigs[i])
	}
	return transaction, nil

}

func fromGrpcState(protoState *pb.State) (state *channel.State, err error) {
	state = &channel.State{}
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

func toAppAndData(protoApp, protoData []byte) (app channel.App, data channel.Data, err error) {
	if len(protoApp) == 0 {
		app = channel.NoApp()
		data = channel.NoData()
		return app, data, nil
	}
	appDef := wallet.NewAddress()
	err = appDef.UnmarshalBinary(protoApp)
	if err != nil {
		return nil, nil, err
	}
	app, err = channel.Resolve(appDef)
	if err != nil {
		return
	}
	data = app.NewData()
	return app, data, data.UnmarshalBinary(protoData)
}

func toAllocation(protoAlloc *pb.Allocation) (alloc *channel.Allocation, err error) {
	alloc = &channel.Allocation{}
	alloc.Assets = make([]channel.Asset, len(protoAlloc.Assets))
	for i := range protoAlloc.Assets {
		alloc.Assets[i] = channel.NewAsset()
		err = alloc.Assets[i].UnmarshalBinary(protoAlloc.Assets[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th asset", i)
		}
	}
	alloc.Locked = make([]channel.SubAlloc, len(protoAlloc.Locked))
	for i := range protoAlloc.Locked {
		alloc.Locked[i], err = fromGrpcSubAlloc(protoAlloc.Locked[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th sub alloc", i)
		}
	}
	alloc.Balances = fromGrpcBalances(protoAlloc.Balances.Balances)
	return alloc, nil
}

func fromGrpcBalances(protoBalances []*pb.Balance) [][]*big.Int {
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

func fromGrpcSubAlloc(protoSubAlloc *pb.SubAlloc) (subAlloc channel.SubAlloc, err error) {
	subAlloc = channel.SubAlloc{}
	subAlloc.Bals = fromGrpcBalance(protoSubAlloc.Bals)
	if len(protoSubAlloc.Id) != len(subAlloc.ID) {
		return subAlloc, errors.New("sub alloc id has incorrect length")
	}
	copy(subAlloc.ID[:], protoSubAlloc.Id)
	subAlloc.IndexMap, err = fromGrpcIndexMap(protoSubAlloc.IndexMap.IndexMap)
	return subAlloc, err
}

func fromGrpcBalance(protoBalance *pb.Balance) (balance []channel.Bal) {
	balance = make([]channel.Bal, len(protoBalance.Balance))
	for j := range protoBalance.Balance {
		balance[j] = new(big.Int).SetBytes(protoBalance.Balance[j])
	}
	return balance
}

func fromGrpcIndexMap(protoIndexMap []uint32) (indexMap []channel.Index, err error) {
	indexMap = make([]channel.Index, len(protoIndexMap))
	for i := range protoIndexMap {
		if protoIndexMap[i] > math.MaxUint16 {
			return nil, fmt.Errorf("%d'th index is invalid", i) //nolint:goerr113  // We do not want to define this as constant error.
		}
		indexMap[i] = channel.Index(uint16(protoIndexMap[i]))
	}
	return indexMap, nil
}

// ToGrpcPayments is a helper function to convert slice of Payment struct
// defined in perun-node package to slice of Payment struct defined in grpc
// package.
func ToGrpcPayments(payments []payment.Payment) []*pb.Payment {
	grpcPayments := make([]*pb.Payment, len(payments))
	for i := range payments {
		grpcPayments[i] = ToGrpcPayment(payments[i])
	}
	return grpcPayments
}

// ToGrpcPayment is a helper function to convert Payment struct defined in
// perun-node package to Payment struct defined in gprc package.
func ToGrpcPayment(src payment.Payment) *pb.Payment {
	return &pb.Payment{
		Currency: src.Currency,
		Payee:    src.Payee,
		Amount:   src.Amount,
	}
}

// fromGrpcPayment is a helper function to convert slice of Payment struct defined in
// grpc package to slice of Payment struct defined in perun-node.
func fromGrpcPayments(payments []*pb.Payment) []payment.Payment {
	grpcPayments := make([]payment.Payment, len(payments))
	for i := range payments {
		grpcPayments[i] = fromGrpcPayment(payments[i])
	}
	return grpcPayments
}

// fromGrpcPayment is a helper function to convert Payment struct defined in
// grpc package to Payment struct defined in perun-node.
func fromGrpcPayment(src *pb.Payment) payment.Payment {
	return payment.Payment{
		Currency: src.Currency,
		Payee:    src.Payee,
		Amount:   src.Amount,
	}
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
		BalInfo: ToGrpcBalInfo(src.BalInfo),
		Version: src.Version,
	}
}

// fromGrpcBalInfo is a helper function to convert BalInfo struct defined in grpc package
// to BalInfo struct defined in perun-node.
func fromGrpcBalInfo(src *pb.BalInfo) perun.BalInfo {
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

// ToGrpcBalInfo is a helper function to convert BalInfo struct defined in perun-node
// to BalInfo struct defined in grpc package.
func ToGrpcBalInfo(src perun.BalInfo) *pb.BalInfo {
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

// makeBalInfoFromState retrieves balance information from the channel state.
func makeBalInfoFromState(parts []string, currencies []perun.Currency, state *pchannel.State) perun.BalInfo {
	return makeBalInfoFromRawBal(parts, currencies, state.Balances)
}

// makeBalInfoFromRawBal retrieves balance information from the raw balance.
func makeBalInfoFromRawBal(parts []string, currencies []perun.Currency, rawBal [][]*big.Int) perun.BalInfo {
	balInfo := perun.BalInfo{
		Currencies: make([]string, len(currencies)),
		Parts:      parts,
		Bals:       make([][]string, len(rawBal)),
	}
	for i := range rawBal {
		balInfo.Currencies[i] = currencies[i].Symbol()
		balInfo.Bals[i] = make([]string, len(rawBal[i]))
		for j := range rawBal[i] {
			balInfo.Bals[i][j] = currencies[i].Print(rawBal[i][j])
		}
	}

	return balInfo
}
