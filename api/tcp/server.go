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

package peruntcp

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/api/handlers"
	"github.com/hyperledger-labs/perun-node/app/payment"

	"google.golang.org/protobuf/proto"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	psync "polycry.pt/poly-go/sync"
)

type server struct {
	psync.Closer

	server net.Listener

	fundingHandler  *handlers.FundingHandler
	watchingHandler *handlers.WatchingHandler

	sessionID string // For timebeing use hard-coded session-id

	channels    map[string](chan *pb.StartWatchingLedgerChannelReq)
	channelsMtx psync.Mutex
}

// ServeFundingWatchingAPI starts a payment channel API server that listens for incoming grpc
// requests at the specified address and serves those requests using the node API instance.
func ServeFundingWatchingAPI(n perun.NodeAPI, port string) error {
	var err error
	sessionID, _, err := payment.OpenSession(n, "api/session.yaml")
	if err != nil {
		return err
	}

	fundingServer := &handlers.FundingHandler{
		N:          n,
		Subscribes: make(map[string]map[pchannel.ID]pchannel.AdjudicatorSubscription),
	}
	watchingServer := &handlers.WatchingHandler{
		N:          n,
		Subscribes: make(map[string]map[pchannel.ID]pchannel.AdjudicatorSubscription),
	}

	tcpServer, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("listener: %w", err)
	}

	s := &server{
		server:          tcpServer,
		fundingHandler:  fundingServer,
		watchingHandler: watchingServer,
		sessionID:       sessionID,

		channels: make(map[string](chan *pb.StartWatchingLedgerChannelReq), 10),
	}
	s.OnCloseAlways(func() { tcpServer.Close() }) //nolint: errcheck, gosec

	for {
		conn, err := s.server.Accept()
		if err != nil {
			return err
		}

		go s.handle(conn)
	}
}

func (s *server) handle(conn io.ReadWriteCloser) {
	defer conn.Close()                       //nolint: errcheck
	s.OnCloseAlways(func() { conn.Close() }) //nolint: errcheck, gosec

	var m psync.Mutex

	for {
		msg, err := recvMsg(conn)
		// log.Info("received message", msg, err)
		if err != nil {
			log.Errorf("%+v", msg)
			log.Errorf("here decoding message failed: %v", err)
			return
		}

		go func() {
			switch msg := msg.GetMsg().(type) {
			case *pb.APIMessage_FundReq:
				s.handleFundReq(msg, &m, conn)

			case *pb.APIMessage_RegisterReq:
				s.handleRegisterReq(msg, &m, conn)
			case *pb.APIMessage_WithdrawReq:
				s.handleWithdrawReq(msg, &m, conn)
			case *pb.APIMessage_StartWatchingLedgerChannelReq:
				s.handleStartWatchingLedgerChannelReq(msg)
			case *pb.APIMessage_StopWatchingReq:
				s.handleStopWatching(msg)

			}
		}()
	}
}

func (s *server) handleFundReq(msg *pb.APIMessage_FundReq, m *psync.Mutex, conn io.ReadWriteCloser) { //nolint: dupl
	log.Warnf("Server: Got Funding request")
	msg.FundReq.SessionID = s.sessionID
	fundResp, err := s.fundingHandler.Fund(context.Background(), msg.FundReq)
	if err != nil {
		log.Errorf("fund response error +%v", err)
	}
	err = sendMsg(m, conn, &pb.APIMessage{Msg: &pb.APIMessage_FundResp{
		FundResp: fundResp,
	}})
	if err != nil {
		log.Errorf("sending response error +%v", err)
	}
}

//nolint:dupl
func (s *server) handleRegisterReq(msg *pb.APIMessage_RegisterReq, m *psync.Mutex, conn io.ReadWriteCloser) {
	log.Warnf("Server: Got Registering request")
	msg.RegisterReq.SessionID = s.sessionID
	registerResp, err := s.fundingHandler.Register(context.Background(), msg.RegisterReq)
	if err != nil {
		log.Errorf("register response error +%v", err)
	}
	err = sendMsg(m, conn, &pb.APIMessage{Msg: &pb.APIMessage_RegisterResp{
		RegisterResp: registerResp,
	}})
	if err != nil {
		log.Errorf("sending response error +%v", err)
	}
}

//nolint:dupl
func (s *server) handleWithdrawReq(msg *pb.APIMessage_WithdrawReq, m *psync.Mutex, conn io.ReadWriteCloser) {
	log.Warnf("Server: Got Withdrawing request")
	msg.WithdrawReq.SessionID = s.sessionID
	withdrawResp, err := s.fundingHandler.Withdraw(context.Background(), msg.WithdrawReq)
	if err != nil {
		log.Errorf("withdraw response error +%v", err)
	}
	err = sendMsg(m, conn, &pb.APIMessage{Msg: &pb.APIMessage_WithdrawResp{
		WithdrawResp: withdrawResp,
	}})
	if err != nil {
		log.Errorf("sending response error +%v", err)
	}
}

func (s *server) handleStartWatchingLedgerChannelReq(msg *pb.APIMessage_StartWatchingLedgerChannelReq) {
	log.Warnf("Server: Got Watching request")
	msg.StartWatchingLedgerChannelReq.SessionID = s.sessionID

	s.channelsMtx.Lock()
	ch, ok := s.channels[string(msg.StartWatchingLedgerChannelReq.State.Id)]
	s.channelsMtx.Unlock()
	if ok {
		ch <- msg.StartWatchingLedgerChannelReq
		return
	}

	ch = make(chan *pb.StartWatchingLedgerChannelReq, 10)
	s.channelsMtx.Lock()
	s.channels[string(msg.StartWatchingLedgerChannelReq.State.Id)] = ch
	s.channelsMtx.Unlock()

	receiveState := func() (*pb.StartWatchingLedgerChannelReq, error) {
		update, ok := <-ch
		if !ok {
			return nil, errors.New("subscription closed")
		}
		return update, nil
	}

	sendAdjEvent := func(resp *pb.StartWatchingLedgerChannelResp) error {
		return nil
	}

	err := s.watchingHandler.StartWatchingLedgerChannel(
		msg.StartWatchingLedgerChannelReq,
		sendAdjEvent,
		receiveState)
	if err != nil {
		log.Errorf("start watching returned with error +%v", err)
	}
}

func (s *server) handleStopWatching(msg *pb.APIMessage_StopWatchingReq) {
	msg.StopWatchingReq.SessionID = s.sessionID
	_, err := s.watchingHandler.StopWatching(context.Background(), msg.StopWatchingReq)
	if err != nil {
		log.Errorf("start watching returned with error +%v", err)
	}
}

func recvMsg(conn io.Reader) (*pb.APIMessage, error) {
	var size uint16
	if err := binary.Read(conn, binary.BigEndian, &size); err != nil {
		return nil, fmt.Errorf("reading size of data from wire: %w", err)
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, fmt.Errorf("reading data from wire: %w", err)
	}
	var msg pb.APIMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("unmarshaling message: %w", err)
	}
	return &msg, nil
}

func sendMsg(m *psync.Mutex, conn io.Writer, msg *pb.APIMessage) error {
	m.Lock()
	defer m.Unlock()
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling message: %w", err)
	}
	if err = binary.Write(conn, binary.BigEndian, uint16(len(data))); err != nil {
		return fmt.Errorf("writing length to wire: %w", err)
	}
	if _, err = conn.Write(data); err != nil {
		return fmt.Errorf("writing data to wire: %w", err)
	}
	return nil
}
