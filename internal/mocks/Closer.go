// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Closer is an autogenerated mock type for the Closer type
type Closer struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Closer) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewCloser interface {
	mock.TestingT
	Cleanup(func())
}

// NewCloser creates a new instance of Closer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewCloser(t mockConstructorTestingTNewCloser) *Closer {
	mock := &Closer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
