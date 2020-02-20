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
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/direct-state-transfer/dst-go/channel/primitives"
	"github.com/direct-state-transfer/dst-go/identity"
)

func setupMockChannel() (*Instance, *genericChannelAdapter) {

	adapter := &genericChannelAdapter{
		connected:        false,
		writeHandlerPipe: newHandlerPipe(handlerPipeModeWrite),
		readHandlerPipe:  newHandlerPipe(handlerPipeModeRead),
	}
	mockCh := &Instance{adapter: adapter}

	return mockCh, adapter
}

func flushMessagePipe(pipe handlerPipe) {

	//Other channels are unbuffered
	for len(pipe.handlerError) > 0 {
		<-pipe.handlerError
	}
}
func Test_genericChannelAdapter_Read(t *testing.T) {

	t.Run("Success_LoggingEnabled", func(t *testing.T) {

		ReadWriteLoggingOldValue := ReadWriteLogging
		ReadWriteLogging = true
		defer func() {
			ReadWriteLogging = ReadWriteLoggingOldValue
		}()

		mockCh, mockAdapter := setupMockChannel()
		mockAdapter.connected = true
		testMsgPacket := jsonMsgPacket{
			message: primitives.ChMsgPkt{Version: "0.1", MessageID: "test-msg"},
			err:     nil}

		flushMessagePipe(mockAdapter.readHandlerPipe)

		//mock to send test msg packet
		go func(msgPacketToSend jsonMsgPacket, msgPacketCh chan jsonMsgPacket) {

			msgPacketCh <- msgPacketToSend

		}(testMsgPacket, mockAdapter.readHandlerPipe.msgPacket)

		gotJSONMsg, err := mockCh.adapter.Read()
		if err != nil {
			t.Errorf("Read() Error %s, wantErr %s", err, "nil")
		}

		if !reflect.DeepEqual(gotJSONMsg, testMsgPacket.message) {
			t.Errorf("Read() got %s, wantr %s", gotJSONMsg, testMsgPacket.message)
		}

	})

	t.Run("HandlerError", func(t *testing.T) {

		mockCh, mockAdapter := setupMockChannel()
		mockAdapter.connected = true
		fakeHandlerError := fmt.Errorf("fake error for unit test")

		flushMessagePipe(mockAdapter.readHandlerPipe)
		mockAdapter.readHandlerPipe.handlerError <- fakeHandlerError

		_, err := mockCh.adapter.Read()
		if err != fakeHandlerError {
			t.Errorf("Handler Error is not properly received by Send Message")
		}
	})

	t.Run("message_error", func(t *testing.T) {

		mockCh, mockAdapter := setupMockChannel()
		mockAdapter.connected = true

		fakeParseError := fmt.Errorf("fake parse error for unit tests")
		testMsgPacket := jsonMsgPacket{
			err: fakeParseError}

		flushMessagePipe(mockAdapter.readHandlerPipe)

		//mock to send test msg packet
		go func(msgPacketToSend jsonMsgPacket, msgPacketCh chan jsonMsgPacket) {

			msgPacketCh <- msgPacketToSend

		}(testMsgPacket, mockAdapter.readHandlerPipe.msgPacket)

		_, err := mockCh.adapter.Read()
		if err != fakeParseError {
			t.Errorf("Read() Error %s, wantErr %s", err, fakeParseError)
		}

	})

	t.Run("fail_on_no_connection", func(t *testing.T) {
		mockCh, mockAdapter := setupMockChannel()
		mockAdapter.connected = false

		_, err := mockCh.adapter.Read()
		if err == nil {
			t.Errorf("readMessage(). not failing on no connection")
		}
	})
}
func Test_genericChannelAdapter_Write(t *testing.T) {

	t.Run("HandlerError", func(t *testing.T) {

		mockCh, mockAdapter := setupMockChannel()
		mockAdapter.connected = true
		fakeHandlerError := fmt.Errorf("fake error for unit test")

		flushMessagePipe(mockAdapter.writeHandlerPipe)
		mockAdapter.writeHandlerPipe.handlerError <- fakeHandlerError

		err := mockCh.adapter.Write(primitives.ChMsgPkt{})
		if err != fakeHandlerError {
			t.Errorf("Handler Error is not properly received by Send Message - %v", err)
		}
	})

	t.Run("Success_LoggingEnabled", func(t *testing.T) {

		ReadWriteLoggingOldValue := ReadWriteLogging
		ReadWriteLogging = true
		defer func() {
			ReadWriteLogging = ReadWriteLoggingOldValue
		}()

		mockCh, mockAdapter := setupMockChannel()
		mockAdapter.connected = true
		testMsg := primitives.ChMsgPkt{
			Version:   "0.1",
			MessageID: "test-msg",
		}
		var gotJSONMsg primitives.ChMsgPkt

		flushMessagePipe(mockAdapter.writeHandlerPipe)

		//mock to echo the message packet
		go func(gotJsonMsg *primitives.ChMsgPkt, msgPacketCh chan jsonMsgPacket) {

			gotMsgPacket := <-msgPacketCh
			*gotJsonMsg = gotMsgPacket.message
			msgPacketCh <- gotMsgPacket

		}(&gotJSONMsg, mockAdapter.writeHandlerPipe.msgPacket)

		err := mockCh.adapter.Write(testMsg)
		if err != nil {
			t.Errorf("Write() Error %s, wantErr %s", err, "nil")
		}

		//Reset timestamp to nil before comparison.
		gotJSONMsg.Timestamp = time.Time{}
		if !reflect.DeepEqual(gotJSONMsg, testMsg) {
			t.Errorf("Write(). should send jsonMsg %+v, but handlerGot %+v",
				testMsg, gotJSONMsg)
		}

	})

	t.Run("message_error", func(t *testing.T) {

		mockCh, mockAdapter := setupMockChannel()
		mockAdapter.connected = true
		testMsg := primitives.ChMsgPkt{
			Version:   "0.1",
			MessageID: "test-msg",
		}
		fakeParseError := fmt.Errorf("fake parse error for unit tests")

		flushMessagePipe(mockAdapter.writeHandlerPipe)

		//mock to echo the message packet
		go func(errToSend error, msgPacketCh chan jsonMsgPacket) {

			gotMsgPacket := <-msgPacketCh
			gotMsgPacket.err = errToSend
			msgPacketCh <- gotMsgPacket

		}(fakeParseError, mockAdapter.writeHandlerPipe.msgPacket)

		err := mockCh.adapter.Write(testMsg)
		if err != fakeParseError {
			t.Errorf("sendMessage() Error want %s, got %s", fakeParseError, err)
		}
	})

	t.Run("fail_on_no_connection", func(t *testing.T) {
		mockCh, mockAdapter := setupMockChannel()
		mockAdapter.connected = false

		err := mockCh.adapter.Write(primitives.ChMsgPkt{})
		if err == nil {
			t.Errorf("sendMessage(). not failing on no connection")
		}
		t.Log("got error :", err)
	})

}

func Test_genericChannelAdapter_Close(t *testing.T) {

	t.Run("CloseSuccess", func(t *testing.T) {
		mockCh, mockAdapter := setupMockChannel()
		mockAdapter.connected = true

		respondToQuit := func(quitChannel chan bool) {
			<-quitChannel
			quitChannel <- true
		}

		go respondToQuit(mockAdapter.writeHandlerPipe.quit)
		go respondToQuit(mockAdapter.readHandlerPipe.quit)

		err := mockCh.Close()

		if err != nil {
			t.Errorf("Close() error = %v, wantErr nil", err)
		}
	})

	t.Run("CloseClosedChannel", func(t *testing.T) {
		mockCh, mockAdapter := setupMockChannel()
		mockAdapter.connected = false

		err := mockCh.Close()
		if err == nil {
			t.Errorf("Close() error = nil, wantErr %+v", err)
		}

	})

	t.Run("Close_when_handler_error", func(t *testing.T) {
		mockCh, mockAdapter := setupMockChannel()
		mockAdapter.connected = true
		mockAdapter.writeHandlerPipe.handlerError <- fmt.Errorf("dummy-handler-error")

		err := mockCh.Close()

		if err != nil {
			t.Errorf("Close() error = %v, wantErr nil", err)
		}
	})
	//TODO : Test for receiver/sender errors during close.
	// Less important as it is of no consequence
}

func Test_newHandlerPipe(t *testing.T) {
	t.Run("success_handlerPipeModeRead", func(t *testing.T) {
		pipe := newHandlerPipe(handlerPipeModeRead)

		if cap(pipe.handlerError) != 1 {
			t.Errorf("newHandlerPipe mode Read handlerError Capacity is not %d", 1)
		}
		if cap(pipe.quit) != 0 {
			t.Errorf("newHandlerPipe mode Read quit Capacity is not %d", 0)
		}
		if cap(pipe.msgPacket) != 1 {
			t.Errorf("newHandlerPipe mode Read msgPacket Capacity is not %d", 1)
		}
	})

	t.Run("success_handlerPipeModeWrite", func(t *testing.T) {
		pipe := newHandlerPipe(handlerPipeModeWrite)

		if cap(pipe.handlerError) != 1 {
			t.Errorf("newHandlerPipe mode Write handlerError Capacity is not %d", 1)
		}
		if cap(pipe.quit) != 0 {
			t.Errorf("newHandlerPipe mode Write quit Capacity is not %d", 0)
		}
		if cap(pipe.msgPacket) != 0 {
			t.Errorf("newHandlerPipe mode Write msgPacket Capacity is not %d", 1)
		}
	})

	t.Run("success_handlerPipeModeDummy", func(t *testing.T) {
		pipe := newHandlerPipe(handlerPipeMode("dummy"))

		if cap(pipe.handlerError) != 0 {
			t.Errorf("newHandlerPipe mode Write handlerError Capacity is not %d", 1)
		}
		if cap(pipe.quit) != 0 {
			t.Errorf("newHandlerPipe mode Write quit Capacity is not %d", 0)
		}
		if cap(pipe.msgPacket) != 0 {
			t.Errorf("newHandlerPipe mode Write msgPacket Capacity is not %d", 1)
		}
	})
}

func Test_startListener(t *testing.T) {

	t.Run("valid_websocket_adapter", func(t *testing.T) {

		_ = exec.Command("fuser", "-k 9602/tcp").Run() //setup
		defer func() {
			_ = exec.Command("fuser", "-k 9602/tcp").Run() //teardown
		}()

		inConnChannel, listener, err := startListener(bobID, 10, WebSocket)
		_, _ = inConnChannel, listener
		if err != nil {
			t.Fatalf("wsStartListener() err = %v, want nil", err)
		}
		time.Sleep(200 * time.Millisecond) //Wait till the listener starts
		defer func() {
			_ = listener.Shutdown(context.Background())
		}()

		//Send in new connection
		_, err = newWsChannel(bobID.ListenerIPAddr, bobID.ListenerEndpoint)
		if err != nil {
			t.Fatalf("Test on startListener - newWsChannel() err = %v, want nil", err)
		}
	})

	t.Run("websocket_invalid_listener_address", func(t *testing.T) {

		invalidID := identity.OffChainID{
			ListenerIPAddr: "abc:jkl:xyz",
		}
		inConnChannel, listener, err := startListener(invalidID, 10, WebSocket)
		_, _ = inConnChannel, listener
		if err == nil {
			t.Fatalf("wsStartListener() err = nil, want non nil")
		}
	})

	t.Run("invalid_adapter", func(t *testing.T) {

		blankID := identity.OffChainID{}
		inConnChannel, listener, err := startListener(blankID, 10, AdapterType("invalid-adapter"))
		_, _ = inConnChannel, listener
		if err == nil {
			t.Fatalf("wsStartListener() err = nil, want non nil")
		}
	})

}
