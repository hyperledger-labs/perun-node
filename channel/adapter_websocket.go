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
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type wsConnInterface interface {
	SetWriteDeadline(time.Time) error
	WriteMessage(int, []byte) error

	SetReadLimit(int64)
	SetPongHandler(func(string) error)
	SetReadDeadline(time.Time) error
	ReadMessage() (int, []byte, error)

	Close() error
}

type wsConfigType struct {
	writeWait      time.Duration
	pongWait       time.Duration
	pingPeriod     time.Duration
	maxMessageSize int64
}

var wsConfig = wsConfigType{
	writeWait:      10 * time.Second,
	pongWait:       60 * time.Second,
	pingPeriod:     ((60 * time.Second) * 9) / 10, //ping period = (pongWait * 9)/10
	maxMessageSize: 1024,
}

type wsChannel struct {
	*genericChannelAdapter
	wsConn *websocket.Conn
}

//Shutdown enforces the specific adapter to provide a mechanism to shutdown listener
type Shutdown interface {
	Shutdown(context.Context) error
}

func wsStartListener(addr, endpoint string, inConn chan ReadWriteCloser) (
	sh Shutdown, err error) {

	listnerMux := http.NewServeMux()
	listnerMux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		wsConnHandler(inConn, w, r)
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: listnerMux,
	}

	if addr == "" {
		addr = ":http"
	}

	///Starting listener and server separately enables the program to catch
	//errors when listening has failed to start
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return srv, err
	}

	go func() {
		err := srv.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
		if err != nil {
			//ErrServerClosed is returned when the server is shutdown by user intentionally
			if err == http.ErrServerClosed {
				logger.Info("Listener at ", addr, " shutdown successfully")
			} else {
				logger.Error("Listener at ", addr, " shutdown with error -", err.Error())
			}
		}
	}()

	return srv, nil
}

func wsConnHandler(inConn chan ReadWriteCloser, w http.ResponseWriter, r *http.Request) {

	var upgrader = websocket.Upgrader{}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		//Errors returned by upgrader.Upgrade are due to issues in the
		//incoming request. Hence log and ignore the connection
		logger.Error("Error in incoming request format :", err.Error())
		return
	}

	ch := &wsChannel{
		genericChannelAdapter: &genericChannelAdapter{
			connected:        true,
			writeHandlerPipe: newHandlerPipe(handlerPipeModeWrite),
			readHandlerPipe:  newHandlerPipe(handlerPipeModeRead),
		},
		wsConn: conn,
	}

	//start read and write handler go routines
	go wsWriteHandler(wsConfig, ch.wsConn, ch.writeHandlerPipe, ch)
	go wsReadHandler(wsConfig, ch.wsConn, ch.readHandlerPipe, ch)

	inConn <- ch
}

func newWsChannel(addr, endpoint string) (_ ReadWriteCloser, err error) {

	peerURL := url.URL{Scheme: "ws", Host: addr, Path: endpoint}

	conn, response, err := websocket.DefaultDialer.Dial(peerURL.String(), nil)
	if err != nil {
		return nil, err
	}
	_ = response.Body.Close()

	wsCh := &wsChannel{
		genericChannelAdapter: &genericChannelAdapter{
			connected:        true,
			writeHandlerPipe: newHandlerPipe(handlerPipeModeWrite),
			readHandlerPipe:  newHandlerPipe(handlerPipeModeRead),
		},
		wsConn: conn,
	}

	//start read and write handler go routines
	go wsWriteHandler(wsConfig, wsCh.wsConn, wsCh.writeHandlerPipe, wsCh)
	go wsReadHandler(wsConfig, wsCh.wsConn, wsCh.readHandlerPipe, wsCh)

	return wsCh, err
}

func wsReadHandler(wsConfig wsConfigType, wsConn wsConnInterface, pipe handlerPipe, ch Closer) {
	defer func() {
		err := wsConn.Close()
		if err != nil {
			logger.Error("Error closing connection -", err)
		}
		logger.Debug("Exiting messageReceiver")
		pipe.quit <- true
	}()

	//Set initial configuration for reading on the websocket connection
	wsConn.SetReadLimit(wsConfig.maxMessageSize)
	err := wsConn.SetReadDeadline(time.Now().Add(wsConfig.pongWait))
	if err != nil {
		logger.Error("Error setting read deadline -", err)
		return
	}
	wsConn.SetPongHandler(func(string) error {
		return wsConn.SetReadDeadline(time.Now().Add(wsConfig.pongWait))
	})

	var (
		message     []byte
		messageType int
	)

	//Timeperiod to do repeat reads
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-pipe.quit:
			ticker.Stop()
			return
		case <-ticker.C:

			messageType, message, err = wsConn.ReadMessage()

			//Error handling from websocket read
			//1. ReadMessage internally calls NextReader and io.Readall
			//   a. Errors from Nextreader are permanent. So exit when it returns error
			//	 b. Errors from Readall is received only due to buffer overflow die to insufficient memory.
			//      All other errors within Readall will result in panic.

			if err != nil {
				ticker.Stop()

				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Error("Connection closed with unexpected error -", err)
				} else {
					logger.Info("Connection closed by peer -", err)
				}

				//If receiver has obtained lock, signal handler error it so that it exists
				//And Lock will be available for Close()
				pipe.handlerError <- err
				go func() {
					err = ch.Close()
					if err != nil {
						logger.Error("Error closing channel-", err)
					}
				}()
				return
			}

			_ = messageType
			if err == nil && messageType != websocket.BinaryMessage {
				message = []byte{}
				err = fmt.Errorf("Only BinaryMessage type is supported by websockets adapter")
			}

			msgPacket := jsonMsgPacket{message, err}
			pipe.msgPacket <- msgPacket
		}
	}
}

func wsWriteHandler(wsConfig wsConfigType, wsConn wsConnInterface, pipe handlerPipe, ch Closer) {

	ticker := time.NewTicker(wsConfig.pingPeriod)

	defer func() {
		ticker.Stop()
		err := wsConn.Close()
		if err != nil {
			logger.Info("error already closed by peer -", err)
		}
		logger.Debug("Exiting messageSender")
		pipe.quit <- true
	}()
	for {
		select {
		case msgPacket := <-pipe.msgPacket:
			err := wsConn.SetWriteDeadline(time.Now().Add(wsConfig.writeWait))
			if err != nil {
				logger.Error("Error setting write deadline -", err)
				msgPacket.err = err
				pipe.msgPacket <- msgPacket
				return
			}

			err = wsConn.WriteMessage(websocket.BinaryMessage, msgPacket.message)

			if err != nil {

				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Error("Connection closed with unexpected error -", err)
				} else {
					logger.Info("Connection closed by peer -", err)
				}

				//If receiver has obtained lock, signal handler error it so that it exists
				//And Lock will be available for Close()
				pipe.handlerError <- err
				go func() {
					err = ch.Close()
					if err != nil {
						logger.Error("Error closing channel-", err)
					}
				}()
				return
			}

			msgPacket.err = err
			pipe.msgPacket <- msgPacket

		case <-ticker.C:
			//Ping period has passed, send ping message
			err := wsConn.SetWriteDeadline(time.Now().Add(wsConfig.writeWait))
			if err != nil {
				logger.Error("Error setting write deadline -", err)
				pipe.handlerError <- err
				return
			}
			err = wsConn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				pipe.handlerError <- err
				return
			}
		case <-pipe.quit:
			return

		}
	}
}

// tcpKeepAliveListener is defined to override Accept method of default listener to enable keepAlive
type tcpKeepAliveListener struct {
	*net.TCPListener
}

// Accept sets keepAlive option and timeout on incoming connections so dead TCP connections go away eventually
func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	err = tc.SetKeepAlive(true)
	if err != nil {
		return
	}
	err = tc.SetKeepAlivePeriod(3 * time.Minute)
	if err != nil {
		return
	}
	return tc, nil
}
