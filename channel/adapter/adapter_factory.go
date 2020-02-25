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

package adapter

import (
	"context"
	"fmt"
	"sync"
)

// CommunicationProtocol represents adapter type for off chain communication protocol
type CommunicationProtocol string

// Enumeration of allowed adapter types for off chain communication protocol
const (
	Mock      CommunicationProtocol = CommunicationProtocol("mock")
	WebSocket CommunicationProtocol = CommunicationProtocol("websocket")
)

// GenericChannelAdapter implements the basic functionalities of a channel adapter.
// Specifically it implements the ReadWriteCloser interface.
//
// The read, write and close functions expects handlers functions to be running concurrently as go-routines.
// These handler functions are to be implemented by each protocol specific adapters and do the actual communication.
// Data transfer between handler functions and read, write, close functions is through handler pipes, which are
// technically go channels.
type GenericChannelAdapter struct {
	IsConnected bool //Status of the connection

	ReadHandlerPipe  HandlerPipe //Set of channels to communicate with receive messagePipeHandlers
	WriteHandlerPipe HandlerPipe //Set of channels to communicate with send messagePipeHandlers

	Access sync.Mutex //Access control when setting connection status
}

// ReadWriteCloser is the interface that groups Read, Write, Close and Connected method.
// Any channel adapter should implement these methods as it will be used by higher levels of code.
type ReadWriteCloser interface {
	Connected() bool
	Read() ([]byte, error)
	Write([]byte) error
	Close() error
}

//Shutdown enforces the specific adapter to provide a mechanism to shutdown listener
type Shutdown interface {
	Shutdown(context.Context) error
}

// Closer is the interface that wraps the Close method.
type Closer interface {
	Close() error
}

// HandlerPipe defines channels for transferring message packets and error information between the main program
// and the handler functions, running concurrently as go-routines.
type HandlerPipe struct {
	MsgPacket    chan JSONMsgPacket
	HandlerError chan error //When handler exits due to error, error to be posted on this channel
	Quit         chan bool  //Use to signal the handler to quit from the main routine
}

// JSONMsgPacket defines a structure for exchanging message packets and error information between the
// main program and the handler functions, via the HandlerPipe.
type JSONMsgPacket struct {
	Message []byte
	Err     error
}

// Connected returns if the connection with the peer is active or not.
func (ch *GenericChannelAdapter) Connected() (isConnected bool) {
	return ch.IsConnected
}

// Read returns any new message that has been received by the read handler of this channel.
//
// If connection is not active, an error is returned.
func (ch *GenericChannelAdapter) Read() (message []byte, err error) {

	ch.Access.Lock()
	defer ch.Access.Unlock()

	if !ch.IsConnected {
		err = fmt.Errorf("Channel already closed")
		return []byte{}, err
	}

	select {
	case err = <-ch.ReadHandlerPipe.HandlerError:
	case msgPacket := <-ch.ReadHandlerPipe.MsgPacket:
		message = msgPacket.Message
		err = msgPacket.Err
	}

	return message, err
}

// Write sends the message to the write handler of this channel to be sent on the channel.
//
// If connection is not active, an error is returned.
func (ch *GenericChannelAdapter) Write(message []byte) (err error) {

	ch.Access.Lock()
	defer ch.Access.Unlock()

	if !ch.IsConnected {
		err = fmt.Errorf("Channel already closed")
		return err
	}

	//Send message if no handler error
	select {
	case err = <-ch.WriteHandlerPipe.HandlerError:
		return err
	default:
		ch.WriteHandlerPipe.MsgPacket <- JSONMsgPacket{message, nil}
	}

	//Wait for response from writeHandler
	select {
	case err = <-ch.WriteHandlerPipe.HandlerError:
	case response := <-ch.WriteHandlerPipe.MsgPacket:
		err = response.Err
	}

	return err
}

// Close closes the connection on this channel and also shuts the read and write handlers down.
func (ch *GenericChannelAdapter) Close() (err error) {

	ch.Access.Lock()
	defer ch.Access.Unlock()

	if !ch.IsConnected {
		err = fmt.Errorf("Channel already closed")
		return err
	}

	//Note on closing mechanism - write handler during it's closure,
	//will close the underlying websocket connection.
	//This will cause error in read handler, which will be blocking on it's read call
	//and hence it will close with error. The error itself is unimportant
	err = CloseHandler(ch.WriteHandlerPipe)

	ch.IsConnected = false
	return nil
}

// HandlerPipeMode is a type to enumerate different modes of a HandlerPipe.
// The mode impacts whether the message packet channel is buffered (non-blocking) or unbuffered (blocking).
type HandlerPipeMode string

const (
	// HandlerPipeModeRead implies the message packet channel is buffered with length of 1 (non-blocking).
	// This allows the read handler to add the received message to channel and read the next message,
	// Even if it is not immediately consumed by the main program.
	HandlerPipeModeRead = HandlerPipeMode("read")

	// HandlerPipeModeWrite implies the message packet channel is unbuffered (blocking).
	// So that, the main program can is blocked until the message to write is received by the write handler.
	// It can then, read the error information, corresponding to the write by the protocol specific handler.
	HandlerPipeModeWrite = HandlerPipeMode("write")
)

// NewHandlerPipe initializes and returns a HandlerPipe as per the given mode.
func NewHandlerPipe(mode HandlerPipeMode) HandlerPipe {

	pipe := HandlerPipe{
		HandlerError: make(chan error, 1),
		Quit:         make(chan bool), //unbuffered quit channel for synchronization
	}
	switch mode {
	case HandlerPipeModeRead:
		pipe.MsgPacket = make(chan JSONMsgPacket, 1) //Read msg pipe is buffered with time out
	case HandlerPipeModeWrite:
		pipe.MsgPacket = make(chan JSONMsgPacket) //Write msg pipe is unbuffered
	default:
		pipe = HandlerPipe{}
	}

	return pipe
}

// CloseHandler shutsdown a protocol specific read/write handler running concurrently as a go-routine.
// It does so by signaling via the quit channel of the HandlerPipe.
func CloseHandler(pipe HandlerPipe) (err error) {

	if len(pipe.HandlerError) != 0 {
		//If any error related to before closing, available in handler error
		err = <-pipe.HandlerError
		return fmt.Errorf("Channel close error : %s", err.Error())
	}

	pipe.Quit <- true
	//For this to work, quit should be initialized as UNBUFFERED CHANNEL
	<-pipe.Quit

	return err
}
