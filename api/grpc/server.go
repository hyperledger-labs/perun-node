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

	psync "perun.network/go-perun/pkg/sync"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
)

// PayChServer represents a grpc server that implements payment channel API.
type PayChServer struct {
	n perun.NodeAPI

	// The mutex should be used when accessing the map data structures.
	psync.Mutex

	// These maps are used to hold an signal channel for each active subscription.
	// When a subscription is registered, subscribe function will add an entry to the
	// map corresponding to the subscription type.
	// The unsubscribe call should retrieve the channel from the map and close it, which
	// will signal the subscription routine to end.
	//
	// chProposalsNotif & chCloseNotifs work on per session basis and hence this is a map
	// of session id to signaling channel.
	// chUpdatesNotif work on a per channel basis and hence this is a map of session id to
	// channel id to signaling channel.

	chProposalsNotif map[string]chan bool
	chUpdatesNotif   map[string]map[string]chan bool
	chClosesNotif    map[string]chan bool
}

// NewPayChServer returns a new grpc server that can server the payment channel API.
func NewPayChServer(n perun.NodeAPI) *PayChServer {
	return &PayChServer{
		n:                n,
		chProposalsNotif: make(map[string]chan bool),
		chUpdatesNotif:   make(map[string]map[string]chan bool),
		chClosesNotif:    make(map[string]chan bool),
	}
}

// GetConfig wraps node.GetConfig.
func (a *PayChServer) GetConfig(context.Context, *pb.GetConfigReq) (*pb.GetConfigResp, error) {
	cfg := a.n.GetConfig()
	return &pb.GetConfigResp{
		ChainAddress:       cfg.ChainURL,
		AdjudicatorAddress: cfg.Adjudicator,
		AssetAddress:       cfg.Asset,
		CommTypes:          cfg.CommTypes,
		ContactTypes:       cfg.ContactTypes,
	}, nil
}

// Time wraps node.Time.
func (a *PayChServer) Time(context.Context, *pb.TimeReq) (*pb.TimeResp, error) {
	return &pb.TimeResp{
		Time: a.n.Time(),
	}, nil
}

// Help wraps node.Help.
func (a *PayChServer) Help(context.Context, *pb.HelpReq) (*pb.HelpResp, error) {
	return &pb.HelpResp{
		Apis: a.n.Help(),
	}, nil
}

// OpenSession wraps node.OpenSession.
func (a *PayChServer) OpenSession(ctx context.Context, req *pb.OpenSessionReq) (*pb.OpenSessionResp, error) {
	errResponse := func(err error) *pb.OpenSessionResp {
		return &pb.OpenSessionResp{
			Response: &pb.OpenSessionResp_Error{
				Error: &pb.MsgError{
					Error: err.Error(),
				},
			},
		}
	}

	sessionID, err := a.n.OpenSession(req.ConfigFile)
	if err != nil {
		return errResponse(err), nil
	}

	a.Lock()
	a.chUpdatesNotif[sessionID] = make(map[string]chan bool)
	a.Unlock()

	return &pb.OpenSessionResp{
		Response: &pb.OpenSessionResp_MsgSuccess_{
			MsgSuccess: &pb.OpenSessionResp_MsgSuccess{
				SessionID: sessionID,
			},
		},
	}, nil
}

// AddContact wraps session.AddContact.
func (a *PayChServer) AddContact(context.Context, *pb.AddContactReq) (*pb.AddContactResp, error) {
	return nil, nil
}

// GetContact wraps session.GetContact.
func (a *PayChServer) GetContact(context.Context, *pb.GetContactReq) (*pb.GetContactResp, error) {
	return nil, nil
}

// OpenPayCh wraps session.OpenPayCh.
func (a *PayChServer) OpenPayCh(context.Context, *pb.OpenPayChReq) (*pb.OpenPayChResp, error) {
	return nil, nil
}

// GetPayChs wraps session.GetPayChs.
func (a *PayChServer) GetPayChs(context.Context, *pb.GetPayChsReq) (*pb.GetPayChsResp, error) {
	return nil, nil
}

// SubPayChProposals wraps session.SubPayChProposals.
func (a *PayChServer) SubPayChProposals(*pb.SubPayChProposalsReq, pb.Payment_API_SubPayChProposalsServer) error {
	return nil
}

// UnsubPayChProposals wraps session.UnsubPayChProposals.
func (a *PayChServer) UnsubPayChProposals(context.Context, *pb.UnsubPayChProposalsReq) (
	*pb.UnsubPayChProposalsResp, error) {
	return nil, nil
}

// RespondPayChProposal wraps session.RespondPayChProposal.
func (a *PayChServer) RespondPayChProposal(context.Context, *pb.RespondPayChProposalReq) (
	*pb.RespondPayChProposalResp, error) {
	return nil, nil
}

// SubPayChCloses wraps session.SubPayChCloses.
func (a *PayChServer) SubPayChCloses(*pb.SubPayChClosesReq, pb.Payment_API_SubPayChClosesServer) error {
	return nil
}

// UnsubPayChClose wraps session.UnsubPayChClose.
func (a *PayChServer) UnsubPayChClose(context.Context, *pb.UnsubPayChClosesReq) (*pb.UnsubPayChClosesResp, error) {
	return nil, nil
}

// CloseSession wraps session.CloseSession.
func (a *PayChServer) CloseSession(context.Context, *pb.CloseSessionReq) (*pb.CloseSessionResp, error) {
	return nil, nil
}

// SendPayChUpdate wraps channel.SendPayChUpdate.
func (a *PayChServer) SendPayChUpdate(context.Context, *pb.SendPayChUpdateReq) (*pb.SendPayChUpdateResp, error) {
	return nil, nil
}

// SubPayChUpdates wraps channel.SubPayChUpdates.
func (a *PayChServer) SubPayChUpdates(*pb.SubpayChUpdatesReq, pb.Payment_API_SubPayChUpdatesServer) error {
	return nil
}

// UnsubPayChUpdates wraps channel.UnsubPayChUpdates.
func (a *PayChServer) UnsubPayChUpdates(context.Context, *pb.UnsubPayChUpdatesReq) (*pb.UnsubPayChUpdatesResp, error) {
	return nil, nil
}

// RespondPayChUpdate wraps channel.RespondPayChUpdate.
func (a *PayChServer) RespondPayChUpdate(context.Context, *pb.RespondPayChUpdateReq) (
	*pb.RespondPayChUpdateResp, error) {
	return nil, nil
}

// GetPayChBalance wraps channel.GetPayChBalance.
func (a *PayChServer) GetPayChBalance(context.Context, *pb.GetPayChBalanceReq) (*pb.GetPayChBalanceResp, error) {
	return nil, nil
}

// ClosePayCh wraps channel.ClosePayCh.
func (a *PayChServer) ClosePayCh(context.Context, *pb.ClosePayChReq) (*pb.ClosePayChResp, error) {
	return nil, nil
}
