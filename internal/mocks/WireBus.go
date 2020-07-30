// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	wallet "perun.network/go-perun/wallet"

	wire "perun.network/go-perun/wire"
)

// WireBus is an autogenerated mock type for the WireBus type
type WireBus struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *WireBus) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Publish provides a mock function with given fields: _a0, _a1
func (_m *WireBus) Publish(_a0 context.Context, _a1 *wire.Envelope) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *wire.Envelope) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubscribeClient provides a mock function with given fields: c, clientAddr
func (_m *WireBus) SubscribeClient(c wire.Consumer, clientAddr wallet.Address) error {
	ret := _m.Called(c, clientAddr)

	var r0 error
	if rf, ok := ret.Get(0).(func(wire.Consumer, wallet.Address) error); ok {
		r0 = rf(c, clientAddr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
