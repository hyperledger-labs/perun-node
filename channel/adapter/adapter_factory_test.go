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
	"fmt"
	"reflect"
	"testing"
)

func setupMockChannel() *GenericChannelAdapter {

	adapter := &GenericChannelAdapter{
		IsConnected:      false,
		WriteHandlerPipe: NewHandlerPipe(HandlerPipeModeWrite),
		ReadHandlerPipe:  NewHandlerPipe(HandlerPipeModeRead),
	}

	return adapter
}

func flushMessagePipe(pipe HandlerPipe) {

	//Other channels are unbuffered
	for len(pipe.HandlerError) > 0 {
		<-pipe.HandlerError
	}
}

func Test_genericChannelAdapter_Connected(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus bool
	}{
		{
			name:       "true",
			wantStatus: true,
		},
		{
			name:       "false",
			wantStatus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAdapter := setupMockChannel()
			mockAdapter.IsConnected = tt.wantStatus
			gotStatus := mockAdapter.Connected()
			if gotStatus != tt.wantStatus {
				t.Errorf("Connected() got %t, want %t", gotStatus, tt.wantStatus)
			}
		})
	}

}
func Test_genericChannelAdapter_Read(t *testing.T) {

	t.Run("Success", func(t *testing.T) {

		mockAdapter := setupMockChannel()
		mockAdapter.IsConnected = true
		testMsgPacket := JSONMsgPacket{
			Message: []byte("test-msg"),
			Err:     nil}

		flushMessagePipe(mockAdapter.ReadHandlerPipe)

		//mock to send test msg packet
		go func(msgPacketToSend JSONMsgPacket, msgPacketCh chan JSONMsgPacket) {

			msgPacketCh <- msgPacketToSend

		}(testMsgPacket, mockAdapter.ReadHandlerPipe.MsgPacket)

		gotJSONMsg, err := mockAdapter.Read()
		if err != nil {
			t.Errorf("Read() Error %s, wantErr %s", err, "nil")
		}

		if !reflect.DeepEqual(gotJSONMsg, testMsgPacket.Message) {
			t.Errorf("Read() got %s, want %s", gotJSONMsg, testMsgPacket.Message)
		}

	})

	t.Run("HandlerError", func(t *testing.T) {

		mockAdapter := setupMockChannel()
		mockAdapter.IsConnected = true
		fakeHandlerError := fmt.Errorf("fake error for unit test")

		flushMessagePipe(mockAdapter.ReadHandlerPipe)
		mockAdapter.ReadHandlerPipe.HandlerError <- fakeHandlerError

		_, err := mockAdapter.Read()
		if err != fakeHandlerError {
			t.Errorf("Handler Error is not properly received by Send Message")
		}
	})

	t.Run("message_error", func(t *testing.T) {

		mockAdapter := setupMockChannel()
		mockAdapter.IsConnected = true

		fakeParseError := fmt.Errorf("fake parse error for unit tests")
		testMsgPacket := JSONMsgPacket{
			Err: fakeParseError}

		flushMessagePipe(mockAdapter.ReadHandlerPipe)

		//mock to send test msg packet
		go func(msgPacketToSend JSONMsgPacket, msgPacketCh chan JSONMsgPacket) {

			msgPacketCh <- msgPacketToSend

		}(testMsgPacket, mockAdapter.ReadHandlerPipe.MsgPacket)

		_, err := mockAdapter.Read()
		if err != fakeParseError {
			t.Errorf("Read() Error %s, wantErr %s", err, fakeParseError)
		}

	})

	t.Run("fail_on_no_connection", func(t *testing.T) {
		mockAdapter := setupMockChannel()
		mockAdapter.IsConnected = false

		_, err := mockAdapter.Read()
		if err == nil {
			t.Errorf("readMessage(). not failing on no connection")
		}
	})
}
func Test_genericChannelAdapter_Write(t *testing.T) {

	t.Run("Success", func(t *testing.T) {

		mockAdapter := setupMockChannel()
		mockAdapter.IsConnected = true
		testMsg := []byte("test-msg")
		var gotJSONMsg []byte

		flushMessagePipe(mockAdapter.WriteHandlerPipe)

		//mock to echo the message packet
		go func(gotJsonMsg *[]byte, msgPacketCh chan JSONMsgPacket) {
			gotMsgPacket := <-msgPacketCh
			*gotJsonMsg = gotMsgPacket.Message
			msgPacketCh <- gotMsgPacket

		}(&gotJSONMsg, mockAdapter.WriteHandlerPipe.MsgPacket)

		err := mockAdapter.Write(testMsg)
		if err != nil {
			t.Errorf("Write() Error %s, wantErr %s", err, "nil")
		}

		//Reset timestamp to nil before comparison.
		if !reflect.DeepEqual(gotJSONMsg, testMsg) {
			t.Errorf("Write(). should send jsonMsg %+v, but handlerGot %+v",
				testMsg, gotJSONMsg)
		}

	})

	t.Run("HandlerError", func(t *testing.T) {

		mockAdapter := setupMockChannel()
		mockAdapter.IsConnected = true
		fakeHandlerError := fmt.Errorf("fake error for unit test")

		flushMessagePipe(mockAdapter.WriteHandlerPipe)
		mockAdapter.WriteHandlerPipe.HandlerError <- fakeHandlerError

		err := mockAdapter.Write([]byte{})
		if err != fakeHandlerError {
			t.Errorf("Handler Error is not properly received by Send Message - %v", err)
		}
	})

	t.Run("message_error", func(t *testing.T) {

		mockAdapter := setupMockChannel()
		mockAdapter.IsConnected = true
		testMsg := []byte("test-msg")
		fakeParseError := fmt.Errorf("fake parse error for unit tests")

		flushMessagePipe(mockAdapter.WriteHandlerPipe)

		//mock to echo the message packet
		go func(errToSend error, msgPacketCh chan JSONMsgPacket) {
			gotMsgPacket := <-msgPacketCh
			gotMsgPacket.Err = errToSend
			msgPacketCh <- gotMsgPacket

		}(fakeParseError, mockAdapter.WriteHandlerPipe.MsgPacket)

		err := mockAdapter.Write(testMsg)
		if err != fakeParseError {
			t.Errorf("sendMessage() Error want %s, got %s", fakeParseError, err)
		}
	})

	t.Run("fail_on_no_connection", func(t *testing.T) {
		mockAdapter := setupMockChannel()
		mockAdapter.IsConnected = false

		err := mockAdapter.Write([]byte{})
		if err == nil {
			t.Errorf("sendMessage(). not failing on no connection")
		}
		t.Log("got error :", err)
	})

}

func Test_genericChannelAdapter_Close(t *testing.T) {

	t.Run("CloseSuccess", func(t *testing.T) {
		mockAdapter := setupMockChannel()
		mockAdapter.IsConnected = true

		respondToQuit := func(quitChannel chan bool) {
			<-quitChannel
			quitChannel <- true
		}

		go respondToQuit(mockAdapter.WriteHandlerPipe.Quit)
		go respondToQuit(mockAdapter.ReadHandlerPipe.Quit)

		err := mockAdapter.Close()

		if err != nil {
			t.Errorf("Close() error = %v, wantErr nil", err)
		}
	})

	t.Run("CloseClosedChannel", func(t *testing.T) {
		mockAdapter := setupMockChannel()
		mockAdapter.IsConnected = false

		err := mockAdapter.Close()
		if err == nil {
			t.Errorf("Close() error = nil, wantErr %+v", err)
		}

	})

	t.Run("Close_when_handler_error", func(t *testing.T) {
		mockAdapter := setupMockChannel()
		mockAdapter.IsConnected = true
		mockAdapter.WriteHandlerPipe.HandlerError <- fmt.Errorf("dummy-handler-error")

		err := mockAdapter.Close()

		if err != nil {
			t.Errorf("Close() error = %v, wantErr nil", err)
		}
	})
	//TODO : Test for receiver/sender errors during close.
	// Less important as it is of no consequence
}

func Test_newHandlerPipe(t *testing.T) {
	t.Run("success_handlerPipeModeRead", func(t *testing.T) {
		pipe := NewHandlerPipe(HandlerPipeModeRead)

		if cap(pipe.HandlerError) != 1 {
			t.Errorf("newHandlerPipe mode Read handlerError Capacity is not %d", 1)
		}
		if cap(pipe.Quit) != 0 {
			t.Errorf("newHandlerPipe mode Read quit Capacity is not %d", 0)
		}
		if cap(pipe.MsgPacket) != 1 {
			t.Errorf("newHandlerPipe mode Read msgPacket Capacity is not %d", 1)
		}
	})

	t.Run("success_handlerPipeModeWrite", func(t *testing.T) {
		pipe := NewHandlerPipe(HandlerPipeModeWrite)

		if cap(pipe.HandlerError) != 1 {
			t.Errorf("newHandlerPipe mode Write handlerError Capacity is not %d", 1)
		}
		if cap(pipe.Quit) != 0 {
			t.Errorf("newHandlerPipe mode Write quit Capacity is not %d", 0)
		}
		if cap(pipe.MsgPacket) != 0 {
			t.Errorf("newHandlerPipe mode Write msgPacket Capacity is not %d", 1)
		}
	})

	t.Run("success_handlerPipeModeDummy", func(t *testing.T) {
		pipe := NewHandlerPipe(HandlerPipeMode("dummy"))

		if cap(pipe.HandlerError) != 0 {
			t.Errorf("newHandlerPipe mode Write handlerError Capacity is not %d", 1)
		}
		if cap(pipe.Quit) != 0 {
			t.Errorf("newHandlerPipe mode Write quit Capacity is not %d", 0)
		}
		if cap(pipe.MsgPacket) != 0 {
			t.Errorf("newHandlerPipe mode Write msgPacket Capacity is not %d", 1)
		}
	})
}
