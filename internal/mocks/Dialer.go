// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	net "perun.network/go-perun/wire/net"

	wire "perun.network/go-perun/wire"
)

// Dialer is an autogenerated mock type for the Dialer type
type Dialer struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Dialer) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Dial provides a mock function with given fields: ctx, addr, ser
func (_m *Dialer) Dial(ctx context.Context, addr wire.Address, ser wire.EnvelopeSerializer) (net.Conn, error) {
	ret := _m.Called(ctx, addr, ser)

	var r0 net.Conn
	if rf, ok := ret.Get(0).(func(context.Context, wire.Address, wire.EnvelopeSerializer) net.Conn); ok {
		r0 = rf(ctx, addr, ser)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(net.Conn)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, wire.Address, wire.EnvelopeSerializer) error); ok {
		r1 = rf(ctx, addr, ser)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Register provides a mock function with given fields: offChainAddr, commAddr
func (_m *Dialer) Register(offChainAddr wire.Address, commAddr string) {
	_m.Called(offChainAddr, commAddr)
}

type mockConstructorTestingTNewDialer interface {
	mock.TestingT
	Cleanup(func())
}

// NewDialer creates a new instance of Dialer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDialer(t mockConstructorTestingTNewDialer) *Dialer {
	mock := &Dialer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
