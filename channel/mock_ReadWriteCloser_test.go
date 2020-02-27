// Code generated by mockery v1.0.0. DO NOT EDIT.

package channel

import (
	"github.com/direct-state-transfer/dst-go/channel/primitives"
	mock "github.com/stretchr/testify/mock"
)

// MockReadWriteCloser is an autogenerated mock type for the ReadWriteCloser type
type MockReadWriteCloser struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *MockReadWriteCloser) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Connected provides a mock function with given fields:
func (_m *MockReadWriteCloser) Connected() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Read provides a mock function with given fields:
func (_m *MockReadWriteCloser) Read() (primitives.ChMsgPkt, error) {
	ret := _m.Called()

	var r0 primitives.ChMsgPkt
	if rf, ok := ret.Get(0).(func() primitives.ChMsgPkt); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(primitives.ChMsgPkt)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Write provides a mock function with given fields: _a0
func (_m *MockReadWriteCloser) Write(_a0 primitives.ChMsgPkt) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(primitives.ChMsgPkt) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
