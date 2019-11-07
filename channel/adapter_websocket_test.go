// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/dst-go
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

package channel

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/direct-state-transfer/dst-go/identity"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
)

func Test_wsStartListener(t *testing.T) {

	t.Run("invalid_address", func(t *testing.T) {

		sh, inConnChannel, err := wsStartListener("abc:jkl", "test-endpoint", 10)
		_, _ = sh, inConnChannel
		if err == nil {
			t.Errorf("wsStartListener() want non nil error, got nil")
		} else {
			t.Logf("wsStartListener() err = %v", err)
		}
	})

	t.Run("nil_address", func(t *testing.T) {

		sh, inConnChannel, err := wsStartListener("", "test-endpoint", 10)
		_, _ = sh, inConnChannel
		if err == nil {
			t.Errorf("wsStartListener() want non nil error, got nil")
		} else {
			t.Logf("wsStartListener() err = %v", err)
		}
	})

	t.Run("success_with_shutdown", func(t *testing.T) {

		sh, inConnChannel, err := wsStartListener("localhost:6170", "test-endpoint", 10)
		_, _ = sh, inConnChannel
		if err != nil {
			t.Errorf("wsStartListener() err = %v, want nil", err)
		}

		time.Sleep(200 * time.Millisecond) //Wait for the server to start

		err = sh.Shutdown(context.Background())
		if err != nil {
			t.Errorf("listner.Shutdown() = %v, want nil", err)
		}
	})

}

func Test_wsConnHandler(t *testing.T) {

	t.Run("valid_client", func(t *testing.T) {
		validClient := func(t *testing.T, addr, endpoint string, bobID identity.OffChainID) {

			u := url.URL{Scheme: "ws", Host: addr, Path: endpoint}

			c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				t.Fatalf("Error dialing to listner - %v", err)
			}
			defer func() {
				_ = c.Close()
			}()

			handshakeRequest := chMsgPkt{
				Version:   "0.1",
				MessageID: MsgIdentityRequest,
				Message:   jsonMsgIdentity{bobID},
			}
			wantHandshakeResponse := chMsgPkt{
				Version:   "0.1",
				MessageID: MsgIdentityResponse,
				Message:   jsonMsgIdentity{aliceID},
			}
			gotHandshakeResponse := chMsgPkt{}

			err = c.WriteJSON(handshakeRequest)
			if err != nil {
				t.Fatalf("Error writing response to listner : %v", err)
			}

			err = c.ReadJSON(&gotHandshakeResponse)
			if err != nil {
				t.Fatalf("Error reading response from listner : %v", err)
			}

			//Reset timestamp to nil before comparison.
			gotHandshakeResponse.Timestamp = time.Time{}
			if !reflect.DeepEqual(gotHandshakeResponse, wantHandshakeResponse) {
				t.Fatalf("Test signature mismatch.want %v, got %v", wantHandshakeResponse, gotHandshakeResponse)
			}
		}

		addr := aliceID.ListenerIPAddr
		endpoint := aliceID.ListenerEndpoint

		maxConn := uint32(100)

		cs, listener, err := startListener(aliceID, maxConn, WebSocket)
		if err != nil {
			t.Fatalf("startListener error - %v, want nil", err.Error())
		}
		defer func() {
			_ = listener.Shutdown(context.Background())
		}()

		gotConnections := 0
		for i := 1; i < 4; i++ {
			ticker := time.After(500 * time.Millisecond)

			validClient(t, addr, endpoint, bobID)
			select {
			case <-cs:
				gotConnections = gotConnections + 1
			case <-ticker:
				t.Errorf("want %d number of connections on listener, got %d", i, gotConnections)
			}
		}

	})

	t.Run("invalid_client_protocol", func(t *testing.T) {

		invalidClientProtocol := func(t *testing.T, addr, endpoint string) {

			u := url.URL{Scheme: "http", Host: addr, Path: endpoint}

			_, err := http.Post(u.String(), "", bytes.NewReader([]byte{}))
			if err != nil {
				t.Fatalf("Error dialing to listner - %v", err)
			}
		}

		addr := bobID.ListenerIPAddr
		endpoint := bobID.ListenerEndpoint

		maxConn := uint32(100)

		_, listener, err := startListener(bobID, maxConn, WebSocket)
		if err != nil {
			t.Fatalf("startListener error - %v, want nil", err.Error())
		}
		defer func() {
			_ = listener.Shutdown(context.Background())
		}()

		invalidClientProtocol(t, addr, endpoint)
		time.Sleep(200 * time.Millisecond) //Wait till goroutines exit

	})
}

func Test_newWsChannel(t *testing.T) {

	t.Run("Listener_not_running", func(t *testing.T) {
		_, err := newWsChannel(aliceID.ListenerIPAddr, aliceID.ListenerEndpoint+"xyz")
		if err == nil {
			t.Errorf("newWsChannel() err = nil, want non nil")
		}
	})

	t.Run("Listener_running", func(t *testing.T) {

		//Setup listener
		aliceListenerAddress, err := aliceID.ListenerLocalAddr()
		if err != nil {
			t.Fatalf("Setup for newWsChannel - ListenerLocalAddr() err = %v, want nil", err)
		}
		listener, err := mockListenerStub(aliceListenerAddress, aliceID.ListenerEndpoint, t)
		if err != nil {
			t.Fatalf("Setup for newWsChannel - mockListenerStub() err = %v, want nil", err)
		}
		defer func() {
			_ = listener.Shutdown(context.Background())
		}()

		_, err = newWsChannel(aliceID.ListenerIPAddr, aliceID.ListenerEndpoint)
		if err != nil {
			t.Errorf("newWsChannel() err = %v, want nil", err)
		}
	})
}

func mockListenerStub(addr string, endpoint string, t *testing.T) (sh Shutdown, err error) {

	listnerMux := http.NewServeMux()
	listnerMux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{}
		_, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade: %v", err)
			return
		}
	})

	///Starting listener and server separately enables the program to catch
	//errors when listening has failed to start
	srv := &http.Server{
		Addr:    addr,
		Handler: listnerMux,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return srv, err
	}

	go func(t *testing.T) {
		err := srv.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
		if err != nil && err != http.ErrServerClosed {
			t.Logf("http.ListenAndServe error - %s, want nil", err.Error())
		}
	}(t)

	return srv, nil
}

// var upgrader = websocket.Upgrader{
// 	HandshakeTimeout: 60000,
// 	ReadBufferSize:   1024,
// 	WriteBufferSize:  1024,
// }

// func TestClientRole(t *testing.T) {

// 	var (
// 		addr     = bobID.ListenerIPAddr
// 		endpoint = bobID.ListenerEndpoint
// 	)

// 	listener, err := setupListnerStub(addr, endpoint, bobID, t)
// 	if err != nil {
// 		t.Errorf("error in setting up listener stub : %v", err)
// 		t.FailNow()
// 	}
// 	defer listener.Shutdown(context.Background())

// 	conn, err := NewChannel(aliceID, bobID, WebSocket)
// 	if err != nil {
// 		t.Errorf("error in creating new channel. got : %v", err)
// 		t.FailNow()
// 	}

// 	// Check message echo
// 	sampleMessage := chMsgPkt{
// 		Version:   "1.0",
// 		MessageID: MsgIdentityRequest,
// 		Message:   jsonMsgIdentity{},
// 	}

// 	err = conn.Write(sampleMessage)
// 	if err != nil {
// 		t.Errorf("Error sending sample message. got : %v", err)
// 	}

// 	logger.Debug("read message")
// 	message, err := conn.Read()
// 	if err != nil {
// 		logger.Debug(
// 			fmt.Sprintf("Error reading sample message response. got : %v", err))

// 		t.Errorf("Error reading sample message response. got : %v", err)
// 	}
// 	fmt.Println("Read message - ", message)

// 	if reflect.DeepEqual(message, sampleMessage) {
// 		logger.Debug(
// 			fmt.Sprintf("Error in message formatting. want %v, got %v",
// 				sampleMessage, message))
// 		t.Errorf("Error in message formatting. want %v, got %v",
// 			sampleMessage, message)
// 	}

// 	time.Sleep(200 * time.Millisecond)

// 	err = conn.Close()
// 	if err != nil {
// 		t.Errorf("Close() want nil, gotErr %v", err)
// 	}

// 	time.Sleep(500 * time.Millisecond)

// }

func Test_wsReadHandler(t *testing.T) {
	t.Run("valid", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return()
		wsConn.On("SetReadDeadline", mock.Anything).Return(nil)
		wsConn.On("SetPongHandler", mock.Anything).Return()
		wsConn.On("ReadJSON", mock.Anything).Return(nil)
		wsConn.On("Close").Return(nil)

		closer.On("Close").Return(nil)

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		if !wsConn.AssertCalled(t, "SetReadLimit", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadLimit() was not called")
		}
		if !wsConn.AssertCalled(t, "SetReadDeadline", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "SetPongHandler", mock.Anything) {
			t.Errorf("wsReadHandler() - SetPongHandler() was not called")
		}
		if !wsConn.AssertCalled(t, "ReadJSON", mock.Anything) {
			t.Errorf("wsReadHandler() - ReadJSON() was not called")
		}
		//Empty read message pipe
		<-pipe.msgPacket

		pipe.quit <- false //Send close signal
		<-pipe.quit        //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsReadHandler() - Close() was not called")
		}

	})

	t.Run("websocket-connection-close-error", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return()
		wsConn.On("SetReadDeadline", mock.Anything).Return(nil)
		wsConn.On("SetPongHandler", mock.Anything).Return()
		wsConn.On("ReadJSON", mock.Anything).Return(nil)
		wsConn.On("Close").Return(fmt.Errorf("websocket-connection-close-erro"))

		closer.On("Close").Return(nil)

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		if !wsConn.AssertCalled(t, "SetReadLimit", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadLimit() was not called")
		}
		if !wsConn.AssertCalled(t, "SetReadDeadline", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "SetPongHandler", mock.Anything) {
			t.Errorf("wsReadHandler() - SetPongHandler() was not called")
		}
		if !wsConn.AssertCalled(t, "ReadJSON", mock.Anything) {
			t.Errorf("wsReadHandler() - ReadJSON() was not called")
		}
		//Empty read message pipe
		<-pipe.msgPacket

		pipe.quit <- false //Send close signal
		<-pipe.quit        //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsReadHandler() - Close() was not called")
		}

	})

	t.Run("SetReadDeadlineError", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return()
		wsConn.On("SetReadDeadline", mock.Anything).Return(fmt.Errorf("read-deadline-error"))
		wsConn.On("Close").Return(nil)

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		if !wsConn.AssertCalled(t, "SetReadLimit", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadLimit() was not called")
		}
		if !wsConn.AssertCalled(t, "SetReadDeadline", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadDeadline() was not called")
		}

		<-pipe.quit //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsReadHandler() - Close() was not called")
		}

	})

	t.Run("ReadJSONError", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return()
		wsConn.On("SetReadDeadline", mock.Anything).Return(nil)
		wsConn.On("SetPongHandler", mock.Anything).Return()
		wsConn.On("ReadJSON", mock.Anything).Return(fmt.Errorf("read-json-error"))
		wsConn.On("Close").Return(nil)

		closer.On("Close").Return(nil)

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		if !wsConn.AssertCalled(t, "SetReadLimit", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadLimit() was not called")
		}
		if !wsConn.AssertCalled(t, "SetReadDeadline", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "SetPongHandler", mock.Anything) {
			t.Errorf("wsReadHandler() - SetPongHandler() was not called")
		}
		if !wsConn.AssertCalled(t, "ReadJSON", mock.Anything) {
			t.Errorf("wsReadHandler() - ReadJSON() was not called")
		}
		//Empty read message pipe
		<-pipe.msgPacket

		pipe.quit <- false //Send close signal
		<-pipe.quit        //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsReadHandler() - Close() was not called")
		}

	})

	t.Run("ReadJSONUnexpectedCloseError", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return()
		wsConn.On("SetReadDeadline", mock.Anything).Return(nil)
		wsConn.On("SetPongHandler", mock.Anything).Return()
		wsConn.On("ReadJSON", mock.Anything).Return(&websocket.CloseError{
			Code: 0,
			Text: "",
		})
		wsConn.On("Close").Return(nil)

		closer.On("Close").Return(nil)

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		if !wsConn.AssertCalled(t, "SetReadLimit", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadLimit() was not called")
		}
		if !wsConn.AssertCalled(t, "SetReadDeadline", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "SetPongHandler", mock.Anything) {
			t.Errorf("wsReadHandler() - SetPongHandler() was not called")
		}
		if !wsConn.AssertCalled(t, "ReadJSON", mock.Anything) {
			t.Errorf("wsReadHandler() - ReadJSON() was not called")
		}

		<-pipe.quit //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsReadHandler() - Close() was not called")
		}

	})
	t.Run("ReadJSONUnexpectedCloseError_ChCloseError", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return()
		wsConn.On("SetReadDeadline", mock.Anything).Return(nil)
		wsConn.On("SetPongHandler", mock.Anything).Return()
		wsConn.On("ReadJSON", mock.Anything).Return(&websocket.CloseError{
			Code: 0,
			Text: "",
		})
		wsConn.On("Close").Return(nil)

		closer.On("Close").Return(fmt.Errorf("channel-close-error"))

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		if !wsConn.AssertCalled(t, "SetReadLimit", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadLimit() was not called")
		}
		if !wsConn.AssertCalled(t, "SetReadDeadline", mock.Anything) {
			t.Errorf("wsReadHandler() - SetReadDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "SetPongHandler", mock.Anything) {
			t.Errorf("wsReadHandler() - SetPongHandler() was not called")
		}
		if !wsConn.AssertCalled(t, "ReadJSON", mock.Anything) {
			t.Errorf("wsReadHandler() - ReadJSON() was not called")
		}

		<-pipe.quit //Wait for confirmation

		if !closer.AssertCalled(t, "Close") {
			t.Errorf("wsReadHandler() - Close() was not called")
		}
		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsReadHandler() - Close() was not called")
		}

	})
}

func Test_wsWriteHandler(t *testing.T) {
	t.Run("valid", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteJSON", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything).Return(nil)
		wsConn.On("Close").Return(nil)

		closer.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		pipe.msgPacket <- jsonMsgPacket{}
		<-pipe.msgPacket

		if !wsConn.AssertCalled(t, "SetWriteDeadline", mock.Anything) {
			t.Errorf("wsWriteHandler() - SetWriteDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "WriteJSON", mock.Anything) {
			t.Errorf("wsWriteHandler() - WriteJSON() was not called")
		}

		pipe.quit <- false //Send close signal
		<-pipe.quit        //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsWriteHandler() - Close() was not called")
		}

	})

	t.Run("CloseError", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteJSON", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything).Return(nil)
		wsConn.On("Close").Return(fmt.Errorf("closer-error"))

		closer.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		pipe.msgPacket <- jsonMsgPacket{}
		<-pipe.msgPacket

		if !wsConn.AssertCalled(t, "SetWriteDeadline", mock.Anything) {
			t.Errorf("wsWriteHandler() - SetWriteDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "WriteJSON", mock.Anything) {
			t.Errorf("wsWriteHandler() - WriteJSON() was not called")
		}

		pipe.quit <- false //Send close signal
		<-pipe.quit        //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsWriteHandler() - Close() was not called")
		}

	})

	t.Run("SetWriteDeadline_Error", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(fmt.Errorf("Set-write-deadline-error"))
		wsConn.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		pipe.msgPacket <- jsonMsgPacket{}
		<-pipe.msgPacket

		if !wsConn.AssertCalled(t, "SetWriteDeadline", mock.Anything) {
			t.Errorf("wsWriteHandler() - SetWriteDeadline() was not called")
		}

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsWriteHandler() - Close() was not called")
		}

	})

	t.Run("WriteJSONError", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteJSON", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything).Return(fmt.Errorf("write-json-error"))
		wsConn.On("Close").Return(nil)

		closer.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		pipe.msgPacket <- jsonMsgPacket{}
		<-pipe.msgPacket

		if !wsConn.AssertCalled(t, "SetWriteDeadline", mock.Anything) {
			t.Errorf("wsWriteHandler() - SetWriteDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "WriteJSON", mock.Anything) {
			t.Errorf("wsWriteHandler() - WriteJSON() was not called")
		}

		pipe.quit <- false //Send close signal
		<-pipe.quit        //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsWriteHandler() - Close() was not called")
		}

	})

	t.Run("WriteJSONUnexpectedCloseError", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteJSON", mock.Anything).Return(&websocket.CloseError{
			Code: 0,
			Text: "",
		})
		wsConn.On("WriteMessage", mock.Anything).Return(nil)

		wsConn.On("Close").Return(nil)

		closer.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		pipe.msgPacket <- jsonMsgPacket{}
		<-pipe.handlerError

		if !wsConn.AssertCalled(t, "SetWriteDeadline", mock.Anything) {
			t.Errorf("wsWriteHandler() - SetWriteDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "WriteJSON", mock.Anything) {
			t.Errorf("wsWriteHandler() - WriteJSON() was not called")
		}

		<-pipe.quit //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsWriteHandler() - Close() was not called")
		}

	})

	t.Run("WriteJSONUnexpectedCloseError_ChannelCloseError", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteJSON", mock.Anything).Return(&websocket.CloseError{
			Code: 0,
			Text: "",
		})
		wsConn.On("WriteMessage", mock.Anything).Return(nil)

		wsConn.On("Close").Return(nil)

		closer.On("Close").Return(fmt.Errorf("channel-close-error"))

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		pipe.msgPacket <- jsonMsgPacket{}
		<-pipe.handlerError

		if !wsConn.AssertCalled(t, "SetWriteDeadline", mock.Anything) {
			t.Errorf("wsWriteHandler() - SetWriteDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "WriteJSON", mock.Anything) {
			t.Errorf("wsWriteHandler() - WriteJSON() was not called")
		}

		<-pipe.quit //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsWriteHandler() - Close() was not called")
		}

	})
	t.Run("ticker", func(t *testing.T) {

		testWsConfig := wsConfigType{
			writeWait:      10 * time.Second,
			pongWait:       100 * time.Millisecond,
			pingPeriod:     ((100 * time.Millisecond) * 9) / 10,
			maxMessageSize: 1024,
		}
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteJSON", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything, mock.Anything).Return(nil)
		wsConn.On("Close").Return(nil)

		closer.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		//Wait until go routine is launched
		time.Sleep(200 * time.Millisecond)

		//Wait until ping period expires and a ping message is sent
		time.Sleep(testWsConfig.pingPeriod)

		if !wsConn.AssertCalled(t, "SetWriteDeadline", mock.Anything) {
			t.Errorf("wsWriteHandler() - SetWriteDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "WriteMessage", mock.Anything, mock.Anything) {
			t.Errorf("wsWriteHandler() - WriteMessage() was not called")
		}

		pipe.quit <- true
		<-pipe.quit //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsWriteHandler() - Close() was not called")
		}

	})
	t.Run("ticker_SetWriteDeadlineError", func(t *testing.T) {

		testWsConfig := wsConfigType{
			writeWait:      10 * time.Second,
			pongWait:       100 * time.Millisecond,
			pingPeriod:     ((100 * time.Millisecond) * 9) / 10,
			maxMessageSize: 1024,
		}
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(fmt.Errorf("set-write-deadline-error"))
		wsConn.On("WriteJSON", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything, mock.Anything).Return(nil)
		wsConn.On("Close").Return(nil)

		closer.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		//Wait until go routine is launched
		time.Sleep(200 * time.Millisecond)

		//Wait until ping period expires and a ping message is sent
		time.Sleep(testWsConfig.pingPeriod)

		if !wsConn.AssertCalled(t, "SetWriteDeadline", mock.Anything) {
			t.Errorf("wsWriteHandler() - SetWriteDeadline() was not called")
		}

		<-pipe.quit //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsWriteHandler() - Close() was not called")
		}

	})
	t.Run("WriteMessageError", func(t *testing.T) {

		testWsConfig := wsConfigType{
			writeWait:      10 * time.Second,
			pongWait:       100 * time.Millisecond,
			pingPeriod:     ((100 * time.Millisecond) * 9) / 10,
			maxMessageSize: 1024,
		}
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteJSON", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything, mock.Anything).Return(fmt.Errorf("write-message-error"))
		wsConn.On("Close").Return(nil)

		closer.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		//Wait until go routine is launched
		time.Sleep(200 * time.Millisecond)

		//Wait until ping period expires and a ping message is sent
		time.Sleep(testWsConfig.pingPeriod)

		if !wsConn.AssertCalled(t, "SetWriteDeadline", mock.Anything) {
			t.Errorf("wsWriteHandler() - SetWriteDeadline() was not called")
		}
		if !wsConn.AssertCalled(t, "WriteMessage", mock.Anything, mock.Anything) {
			t.Errorf("wsWriteHandler() - WriteMessage() was not called")
		}

		<-pipe.quit //Wait for confirmation

		if !wsConn.AssertCalled(t, "Close") {
			t.Errorf("wsWriteHandler() - Close() was not called")
		}

	})
}
