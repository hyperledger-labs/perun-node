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
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
)

func Test_wsStartListener(t *testing.T) {

	t.Run("invalid_address", func(t *testing.T) {

		inConnChannel := make(chan ReadWriteCloser, 10)

		sh, err := wsStartListener("abc:jkl", "test-endpoint", inConnChannel)
		_ = sh
		if err == nil {
			t.Errorf("wsStartListener() want non nil error, got nil")
		} else {
			t.Logf("wsStartListener() err = %v", err)
		}
	})

	t.Run("nil_address", func(t *testing.T) {

		inConnChannel := make(chan ReadWriteCloser, 10)

		sh, err := wsStartListener("", "test-endpoint", inConnChannel)
		_ = sh
		if err == nil {
			t.Errorf("wsStartListener() want non nil error, got nil")
		} else {
			t.Logf("wsStartListener() err = %v", err)
		}
	})

	t.Run("success_with_shutdown", func(t *testing.T) {

		inConnChannel := make(chan ReadWriteCloser, 10)

		sh, err := wsStartListener("localhost:6170", "test-endpoint", inConnChannel)
		_ = sh
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
		validClient := func(t *testing.T, addr, endpoint string) {

			u := url.URL{Scheme: "ws", Host: addr, Path: endpoint}

			c, response, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				t.Fatalf("Error dialing to listner - %v", err)
			}
			_ = response.Body.Close()

			defer func() {
				_ = c.Close()
			}()
		}

		addr := aliceID.ListenerIPAddr
		endpoint := aliceID.ListenerEndpoint

		maxConn := uint32(10)
		inConnChan := make(chan ReadWriteCloser, maxConn)
		//Start Listener for Alice
		listener, err := wsStartListener(addr, endpoint, inConnChan)
		if err != nil {
			t.Fatalf("startListener error - %v, want nil", err.Error())
		}
		defer func() {
			_ = listener.Shutdown(context.Background())
		}()

		//Send connection requests to Alice
		gotConnections := 0
		for i := 1; i < 4; i++ {
			ticker := time.After(500 * time.Millisecond)

			validClient(t, addr, endpoint)
			select {
			case <-inConnChan:
				gotConnections = gotConnections + 1
			case <-ticker:
				t.Errorf("want %d number of connections on listener, got %d", i, gotConnections)
			}
		}

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
		_, err = upgrader.Upgrade(w, r, nil)
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
// 	sampleMessage := primitives.ChMsgPkt{
// 		Version:   "1.0",
// 		MessageID: primitives.MsgIdentityRequest,
// 		Message:   primitives.JSONMsgIdentity{},
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

		wsConn.On("SetReadLimit", mock.Anything).Return().Once()
		wsConn.On("SetReadDeadline", mock.Anything).Return(nil).Once()
		wsConn.On("SetPongHandler", mock.Anything).Return().Once()
		wsConn.On("ReadMessage").Return(websocket.BinaryMessage, []byte("test-message"), nil)
		wsConn.On("Close").Return(nil).Once()

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Empty read message pipe
		<-pipe.msgPacket

		pipe.quit <- false //Send close signal
		<-pipe.quit        //Wait for confirmation

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)
	})

	t.Run("Error_TextMessage", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return().Once()
		wsConn.On("SetReadDeadline", mock.Anything).Return(nil).Once()
		wsConn.On("SetPongHandler", mock.Anything).Return().Once()
		wsConn.On("ReadMessage").Return(websocket.TextMessage, []byte("test-message"), nil)
		wsConn.On("Close").Return(nil).Once()

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Empty read message pipe
		message := <-pipe.msgPacket
		if !bytes.Equal(message.message, []byte{}) {
			t.Errorf("ReadMessage() want message.message = nil, got not nil")
		}
		if message.err == nil {
			t.Errorf("ReadMessage() want message.err = not nil, got nil")
		}

		pipe.quit <- false //Send close signal
		<-pipe.quit        //Wait for confirmation

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)
	})

	t.Run("Error_ReadMessageError", func(t *testing.T) {

		dummyCloseError := fmt.Errorf("dummy close error")

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return().Once()
		wsConn.On("SetReadDeadline", mock.Anything).Return(nil).Once()
		wsConn.On("SetPongHandler", mock.Anything).Return().Once()
		wsConn.On("ReadMessage").Return(websocket.BinaryMessage, []byte("test-message"), dummyCloseError)
		wsConn.On("Close").Return(nil).Once()

		closer.On("Close").Return(nil).Once()

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Empty read message pipe
		err := <-pipe.handlerError
		if err != dummyCloseError {
			t.Errorf("ReadMessage() want err = nil, got not nil")
		}

		<-pipe.quit //Wait for closed signal

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)
	})

	t.Run("Error_ReadDeadlineError", func(t *testing.T) {

		dummyCloseError := fmt.Errorf("dummy close error")

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return().Once()
		wsConn.On("SetReadDeadline", mock.Anything).Return(dummyCloseError).Once()
		wsConn.On("Close").Return(nil).Once()

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		<-pipe.quit //Wait for closed signal

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)
	})

	t.Run("Error_ConnectionCloseError", func(t *testing.T) {

		dummyCloseError := fmt.Errorf("dummy close error")

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return().Once()
		wsConn.On("SetReadDeadline", mock.Anything).Return(nil).Once()
		wsConn.On("SetPongHandler", mock.Anything).Return().Once()
		wsConn.On("ReadMessage").Return(websocket.BinaryMessage, []byte("test-message"), dummyCloseError)
		wsConn.On("Close").Return(dummyCloseError).Once()

		closer.On("Close").Return(nil).Once()

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Empty read message pipe
		err := <-pipe.handlerError
		if err != dummyCloseError {
			t.Errorf("ReadMessage() want err = nil, got not nil")
		}

		<-pipe.quit //Wait for closed signal

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)
	})

	t.Run("Error_ChannelCloseError", func(t *testing.T) {

		dummyCloseError := fmt.Errorf("dummy close error")

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return().Once()
		wsConn.On("SetReadDeadline", mock.Anything).Return(nil).Once()
		wsConn.On("SetPongHandler", mock.Anything).Return().Once()
		wsConn.On("ReadMessage").Return(websocket.BinaryMessage, []byte("test-message"), dummyCloseError)
		wsConn.On("Close").Return(nil).Once()

		closer.On("Close").Return(dummyCloseError).Once()

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Empty read message pipe
		err := <-pipe.handlerError
		if err != dummyCloseError {
			t.Errorf("ReadMessage() want err = nil, got not nil")
		}

		<-pipe.quit //Wait for closed signal

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)
	})

	t.Run("Error_PeerDisconnectedError", func(t *testing.T) {

		dummyCloseError := &websocket.CloseError{
			Code: websocket.CloseProtocolError,
			Text: "dummy error",
		}

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeRead)
		closer := &MockCloser{}

		wsConn.On("SetReadLimit", mock.Anything).Return().Once()
		wsConn.On("SetReadDeadline", mock.Anything).Return(nil).Once()
		wsConn.On("SetPongHandler", mock.Anything).Return().Once()
		wsConn.On("ReadMessage").Return(websocket.BinaryMessage, []byte("test-message"), dummyCloseError)
		wsConn.On("Close").Return(nil).Once()

		closer.On("Close").Return(dummyCloseError).Once()

		go wsReadHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Empty read message pipe
		err := <-pipe.handlerError
		if err != dummyCloseError {
			t.Errorf("ReadMessage() want err = nil, got not nil")
		}

		<-pipe.quit //Wait for closed signal

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)
	})
}

func Test_wsWriteHandler(t *testing.T) {
	t.Run("valid", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything, mock.Anything).Return(nil)
		wsConn.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		testPacketToSend := jsonMsgPacket{message: []byte("test-message"), err: nil}
		pipe.msgPacket <- testPacketToSend
		receivedPacket := <-pipe.msgPacket

		if !(bytes.Equal(testPacketToSend.message, receivedPacket.message)) {
			t.Errorf("WriteMessage() want message.message = nil, got not nil")
		}
		if receivedPacket.err != nil {
			t.Errorf("WriteMessage() want message.err = nil, got not nil")
		}

		pipe.quit <- false //Send close signal
		<-pipe.quit        //Wait for confirmation

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)

	})

	t.Run("Error_WriteMessage", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		dummyError := fmt.Errorf("dummy error")

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything, mock.Anything).Return(dummyError)
		wsConn.On("Close").Return(nil)
		closer.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		testPacketToSend := jsonMsgPacket{message: []byte("test-message"), err: nil}
		pipe.msgPacket <- testPacketToSend

		//Empty read message pipe
		err := <-pipe.handlerError
		if err != dummyError {
			t.Errorf("WriteMessage() want err = nil, got not nil")
		}

		<-pipe.quit //Wait for closed signal

		time.Sleep(10 * time.Millisecond)
		//Delay required for testify to recognize the function call on closer mock in go routine.
		//10ms is an empirical value

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)

	})

	t.Run("Error_WriteMessage_CloseError", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		dummyError := &websocket.CloseError{
			Code: websocket.CloseProtocolError,
			Text: "dummy error",
		}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything, mock.Anything).Return(dummyError)
		wsConn.On("Close").Return(nil)
		closer.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		testPacketToSend := jsonMsgPacket{message: []byte("test-message"), err: nil}
		pipe.msgPacket <- testPacketToSend

		//Empty read message pipe
		err := <-pipe.handlerError
		if err != dummyError {
			t.Errorf("WriteMessage() want err = nil, got not nil")
		}

		<-pipe.quit //Wait for closed signal

		time.Sleep(10 * time.Millisecond)
		//Delay required for testify to recognize the function call on closer mock in go routine.
		//10ms is an empirical value

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)

	})
	t.Run("Error_SetWriteDeadline", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		dummyError := fmt.Errorf("dummy error")

		wsConn.On("SetWriteDeadline", mock.Anything).Return(dummyError)
		wsConn.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		testPacketToSend := jsonMsgPacket{message: []byte("test-message"), err: nil}
		pipe.msgPacket <- testPacketToSend
		receivedPacket := <-pipe.msgPacket

		if !(bytes.Equal(testPacketToSend.message, receivedPacket.message)) {
			t.Errorf("WriteMessage() want message.message = nil, got not nil")
		}
		if receivedPacket.err == nil {
			t.Errorf("WriteMessage() want message.err = not nil, got nil")
		}

		<-pipe.quit //Wait for handler closed signal

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)

	})

	t.Run("Error_Closer_Close", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		dummyError := fmt.Errorf("dummy error")

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything, mock.Anything).Return(dummyError)
		wsConn.On("Close").Return(nil)
		closer.On("Close").Return(dummyError)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		testPacketToSend := jsonMsgPacket{message: []byte("test-message"), err: nil}
		pipe.msgPacket <- testPacketToSend

		//Empty read message pipe
		err := <-pipe.handlerError
		if err != dummyError {
			t.Errorf("WriteMessage() want err = nil, got not nil")
		}

		<-pipe.quit //Wait for closed signal

		time.Sleep(10 * time.Millisecond)
		//Delay required for testify to recognize the function call on closer mock in go routine.
		//10ms is an empirical value

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)

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
		wsConn.On("WriteMessage", mock.Anything, mock.Anything).Return(nil)
		wsConn.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		//Wait until go routine is launched
		time.Sleep(200 * time.Millisecond)

		//Wait until ping period expires and a ping message is sent
		time.Sleep(2 * testWsConfig.pingPeriod)

		pipe.quit <- true
		<-pipe.quit //Wait for confirmation

		wsConn.AssertExpectations(t)

	})

	t.Run("Error_ticker_SetWriteDeadline", func(t *testing.T) {

		testWsConfig := wsConfigType{
			writeWait:      10 * time.Second,
			pongWait:       100 * time.Millisecond,
			pingPeriod:     ((100 * time.Millisecond) * 9) / 10,
			maxMessageSize: 1024,
		}

		dummyCloseError := fmt.Errorf("dummy close error")

		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(dummyCloseError)
		wsConn.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		//Wait until go routine is launched
		time.Sleep(200 * time.Millisecond)

		//Wait until ping period expires and a ping message is sent
		time.Sleep(2 * testWsConfig.pingPeriod)

		err := pipe.handlerError
		if err == nil {
			t.Errorf("WriteMessage() pipe.handlerError = nil, want not nil")
		}

		<-pipe.quit //Wait for confirmation

		wsConn.AssertExpectations(t)

	})

	t.Run("Error_ticker_WriteMessageError", func(t *testing.T) {

		testWsConfig := wsConfigType{
			writeWait:      10 * time.Second,
			pongWait:       100 * time.Millisecond,
			pingPeriod:     ((100 * time.Millisecond) * 9) / 10,
			maxMessageSize: 1024,
		}

		dummyCloseError := fmt.Errorf("dummy close error")

		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything, mock.Anything).Return(dummyCloseError)
		wsConn.On("Close").Return(nil)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		//Wait until go routine is launched
		time.Sleep(200 * time.Millisecond)

		//Wait until ping period expires and a ping message is sent
		time.Sleep(2 * testWsConfig.pingPeriod)

		err := pipe.handlerError
		if err == nil {
			t.Errorf("WriteMessage() pipe.handlerError = nil, want not nil")
		}

		<-pipe.quit //Wait for confirmation

		wsConn.AssertExpectations(t)

	})

	t.Run("Error_Connection_Close", func(t *testing.T) {

		testWsConfig := wsConfig
		wsConn := &mockWsConnInterface{}
		pipe := newHandlerPipe(handlerPipeModeWrite)
		closer := &MockCloser{}

		dummyCloseError := fmt.Errorf("dummy close error")

		wsConn.On("SetWriteDeadline", mock.Anything).Return(nil)
		wsConn.On("WriteMessage", mock.Anything, mock.Anything).Return(nil)
		wsConn.On("Close").Return(dummyCloseError)

		go wsWriteHandler(testWsConfig, wsConn, pipe, closer)

		time.Sleep(200 * time.Millisecond)

		//Send message and receive response
		testPacketToSend := jsonMsgPacket{message: []byte("test-message"), err: nil}
		pipe.msgPacket <- testPacketToSend
		receivedPacket := <-pipe.msgPacket

		if !(bytes.Equal(testPacketToSend.message, receivedPacket.message)) {
			t.Errorf("WriteMessage() want message.message = nil, got not nil")
		}
		if receivedPacket.err != nil {
			t.Errorf("WriteMessage() want message.err = nil, got not nil")
		}

		pipe.quit <- false //Send close signal
		<-pipe.quit        //Wait for confirmation

		wsConn.AssertExpectations(t)
		closer.AssertExpectations(t)

	})
}
