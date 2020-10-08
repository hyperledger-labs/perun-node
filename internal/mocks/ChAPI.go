// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	context "context"

	perun "github.com/hyperledger-labs/perun-node"
	mock "github.com/stretchr/testify/mock"
)

// ChAPI is an autogenerated mock type for the ChAPI type
type ChAPI struct {
	mock.Mock
}

// ChallengeDurSecs provides a mock function with given fields:
func (_m *ChAPI) ChallengeDurSecs() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// Close provides a mock function with given fields: _a0
func (_m *ChAPI) Close(_a0 context.Context) (perun.ChInfo, error) {
	ret := _m.Called(_a0)

	var r0 perun.ChInfo
	if rf, ok := ret.Get(0).(func(context.Context) perun.ChInfo); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(perun.ChInfo)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Currency provides a mock function with given fields:
func (_m *ChAPI) Currency() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetChInfo provides a mock function with given fields:
func (_m *ChAPI) GetChInfo() perun.ChInfo {
	ret := _m.Called()

	var r0 perun.ChInfo
	if rf, ok := ret.Get(0).(func() perun.ChInfo); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(perun.ChInfo)
	}

	return r0
}

// ID provides a mock function with given fields:
func (_m *ChAPI) ID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Parts provides a mock function with given fields:
func (_m *ChAPI) Parts() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// RespondChUpdate provides a mock function with given fields: _a0, _a1, _a2
func (_m *ChAPI) RespondChUpdate(_a0 context.Context, _a1 string, _a2 bool) (perun.ChInfo, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 perun.ChInfo
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) perun.ChInfo); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Get(0).(perun.ChInfo)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, bool) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendChUpdate provides a mock function with given fields: _a0, _a1
func (_m *ChAPI) SendChUpdate(_a0 context.Context, _a1 perun.StateUpdater) (perun.ChInfo, error) {
	ret := _m.Called(_a0, _a1)

	var r0 perun.ChInfo
	if rf, ok := ret.Get(0).(func(context.Context, perun.StateUpdater) perun.ChInfo); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(perun.ChInfo)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, perun.StateUpdater) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SubChUpdates provides a mock function with given fields: _a0
func (_m *ChAPI) SubChUpdates(_a0 perun.ChUpdateNotifier) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(perun.ChUpdateNotifier) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UnsubChUpdates provides a mock function with given fields:
func (_m *ChAPI) UnsubChUpdates() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
