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
	"fmt"
	"sync"
	"time"

	"github.com/direct-state-transfer/dst-go/identity"
)

// ReadWriteLogging to configure logging during channel read/write, for demonstration purposes only
var ReadWriteLogging = false

// AdapterType represents adapter type for off chain communication protocol
type AdapterType string

// Enumeration of allowed adapter types for off chain communication protocol
const (
	Mock      AdapterType = AdapterType("mock")
	WebSocket AdapterType = AdapterType("websocket")
)

type genericChannelAdapter struct {
	connected bool //Status of the connection

	readHandlerPipe  handlerPipe //Set of channels to communicate with receive messagePipeHandlers
	writeHandlerPipe handlerPipe //Set of channels to communicate with send messagePipeHandlers

	access sync.Mutex //Access control when setting connection status
}

// ReadWriteCloser is the interface that groups Read, Write, Close and Connected method.
// Any channel adapter should implement these methods as it will be used by higher levels of code.
type ReadWriteCloser interface {
	Connected() bool
	Read() (chMsgPkt, error)
	Write(chMsgPkt) error
	Close() error
}

// Closer is the interface that wraps the Close method.
type Closer interface {
	Close() error
}

type handlerPipe struct {
	msgPacket    chan jsonMsgPacket
	handlerError chan error //When handler exits due to error, error to be posted on this channel
	quit         chan bool  //Use to signal the handler to quit from the main routine
}

type jsonMsgPacket struct {
	message chMsgPkt
	err     error
}

// Connected returns if the connection with the peer is active or not.
func (ch *genericChannelAdapter) Connected() (isConnected bool) {
	return ch.connected
}

// Read returns any new message that has been received by the read handler of this channel.
//
// If connection is not active, an error is returned.
func (ch *genericChannelAdapter) Read() (message chMsgPkt, err error) {

	ch.access.Lock()
	defer ch.access.Unlock()

	if !ch.connected {
		err = fmt.Errorf("Channel already closed")
		return chMsgPkt{}, err
	}

	select {
	case err = <-ch.readHandlerPipe.handlerError:
	case msgPacket := <-ch.readHandlerPipe.msgPacket:
		message = msgPacket.message
		err = msgPacket.err
	}

	if err == nil && ReadWriteLogging {
		fmt.Printf("\n\n<<<<<<<<<READ : %+v\n\n", message)
		logger.Debug("Incoming Message:", message)
	}

	return message, err
}

// Write sends the message to the write handler of this channel to be sent on the channel.
//
// If connection is not active, an error is returned.
func (ch *genericChannelAdapter) Write(message chMsgPkt) (err error) {

	ch.access.Lock()
	defer ch.access.Unlock()

	if !ch.connected {
		err = fmt.Errorf("Channel already closed")
		return err
	}

	//Send message if no handler error
	select {
	case err = <-ch.writeHandlerPipe.handlerError:
		return err
	default:
		zone, _ := time.LoadLocation("Local")
		message.Timestamp = time.Now().In(zone)
		ch.writeHandlerPipe.msgPacket <- jsonMsgPacket{message, nil}
	}

	//Wait for response from writeHandler
	select {
	case err = <-ch.writeHandlerPipe.handlerError:
	case response := <-ch.writeHandlerPipe.msgPacket:
		err = response.err
	}

	if err == nil && ReadWriteLogging {
		fmt.Printf("\n\n>>>>>>>>>WRITE : %+v\n\n", message)
		logger.Debug("Outgoing Message:", message)
	}

	return err
}

// Close closes the connection on this channel and also shuts the read and write handlers down.
func (ch *genericChannelAdapter) Close() (err error) {

	ch.access.Lock()
	defer ch.access.Unlock()

	if !ch.connected {
		err = fmt.Errorf("Channel already closed")
		return err
	}

	//Note on closing mechanism - write handler during it's closure,
	//will close the underlying websocket connection.
	//This will cause error in read handler, which will be blocking on it's read call
	//and hence it will close with error. The error itself is unimportant
	err = closeHandler(ch.writeHandlerPipe)

	ch.connected = false
	return nil
}

type handlerPipeMode string

const (
	handlerPipeModeRead  = handlerPipeMode("read")
	handlerPipeModeWrite = handlerPipeMode("write")
)

func newHandlerPipe(mode handlerPipeMode) handlerPipe {

	pipe := handlerPipe{
		handlerError: make(chan error, 1),
		quit:         make(chan bool), //unbuffered quit channel for synchronization
	}
	switch mode {
	case handlerPipeModeRead:
		pipe.msgPacket = make(chan jsonMsgPacket, 1) //Read msg pipe is buffered with time out
	case handlerPipeModeWrite:
		pipe.msgPacket = make(chan jsonMsgPacket) //Write msg pipe is unbuffered
	default:
		pipe = handlerPipe{}
	}

	return pipe
}
func closeHandler(pipe handlerPipe) (err error) {

	if len(pipe.handlerError) != 0 {
		//If any error related to before closing, available in handler error
		err = <-pipe.handlerError
		return fmt.Errorf("Channel close error : %s", err.Error())
	}

	pipe.quit <- true
	//For this to work, quit should be initialised as UNBUFFERED CHANNEL
	<-pipe.quit

	return err
}

func startListener(selfID identity.OffChainID, maxConn uint32, adapterType AdapterType) (idVerifiedConn chan *Instance,
	listener Shutdown, err error) {

	if adapterType != WebSocket {
		return nil, nil, fmt.Errorf("Unsupported adapter type - %s", string(adapterType))
	}

	idVerifiedConn = make(chan *Instance, maxConn)

	localAddr, err := selfID.ListenerLocalAddr()
	if err != nil {
		logger.Error("Error in listening on address:", localAddr)
		return nil, nil, err
	}

	//Only websocket adapter is supported currently
	listener, inConn, err := wsStartListener(localAddr, selfID.ListenerEndpoint, maxConn)
	if err != nil {
		logger.Debug("Error starting listen and serve,", err.Error())
		return nil, nil, err
	}

	go func() {
		for {
			newConn := <-inConn
			//Role of user in incoming connections is receiver
			newConn.SetRoleChannel(Receiver)

			peerID, err := newConn.IdentityRead()
			if err != nil {
				err2 := newConn.Close()
				if err2 != nil {
					logger.Error("Error reading peer id-", err, "connection dropped with error -", err2)
				} else {
					logger.Error("Error reading peer id-", err, "connection dropped with no error")
				}
				return
			}

			err = newConn.IdentityRespond(selfID)
			if err != nil {
				err2 := newConn.Close()
				if err2 != nil {
					logger.Error("Error sending self id-", err, "connection dropped with error -", err2)
				} else {
					logger.Error("Error sending self id-", err, "connection dropped with no error")
				}
				return
			}

			//Set self and peer id and return connection to id verified connections channel
			newConn.setSelfID(selfID)
			newConn.setPeerID(peerID)
			idVerifiedConn <- newConn
		}
	}()

	return idVerifiedConn, listener, nil
}

// NewChannel initialises a new channel connection with peer using the adapterType.
//
// selfID is the offchain identity of this user and will be exchanged with the peer when establishing offchain channel.
// peerID is the offchain identity of the peer with whom the offchain channel needs to be established.
//
// An identity exchange is performed after establishing a connection between the two node software instances and
// if it fails, an error is returned.
func NewChannel(selfID, peerID identity.OffChainID, adapterType AdapterType) (conn *Instance, err error) {

	switch adapterType {
	case WebSocket:
		conn, err = newWsChannel(peerID.ListenerIPAddr, peerID.ListenerEndpoint)
		if err != nil {
			logger.Error("Websockets connection dial error:", err)
			return nil, err
		}
		conn.SetRoleChannel(Sender)
	case Mock:
	default:
	}

	//Verify peer identity for all real adapter types
	if adapterType != Mock {

		var gotPeerID identity.OffChainID
		gotPeerID, err = conn.IdentityRequest(selfID)
		if err != nil {
			err = fmt.Errorf("Test connection failed")
			return nil, err
		}

		if !identity.Equal(peerID, gotPeerID) {
			errClose := conn.Close()
			if errClose != nil {
				err = fmt.Errorf("other id mismatch. error in closing conn - %s", errClose.Error())
			} else {
				err = fmt.Errorf("other id mismatch")
			}
			return nil, err
		}

	}

	conn.setSelfID(selfID)
	conn.setPeerID(peerID)

	return conn, nil
}
