// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	tableland "github.com/textileio/go-tableland/internal/tableland"
)

// SQLRunner is an autogenerated mock type for the SQLRunner type
type SQLRunner struct {
	mock.Mock
}

type SQLRunner_Expecter struct {
	mock *mock.Mock
}

func (_m *SQLRunner) EXPECT() *SQLRunner_Expecter {
	return &SQLRunner_Expecter{mock: &_m.Mock}
}

// RunReadQuery provides a mock function with given fields: ctx, stmt
func (_m *SQLRunner) RunReadQuery(ctx context.Context, stmt string) (*tableland.UserRows, error) {
	ret := _m.Called(ctx, stmt)

	var r0 *tableland.UserRows
	if rf, ok := ret.Get(0).(func(context.Context, string) *tableland.UserRows); ok {
		r0 = rf(ctx, stmt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tableland.UserRows)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, stmt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SQLRunner_RunReadQuery_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RunReadQuery'
type SQLRunner_RunReadQuery_Call struct {
	*mock.Call
}

// RunReadQuery is a helper method to define mock.On call
//   - ctx context.Context
//   - stmt string
func (_e *SQLRunner_Expecter) RunReadQuery(ctx interface{}, stmt interface{}) *SQLRunner_RunReadQuery_Call {
	return &SQLRunner_RunReadQuery_Call{Call: _e.mock.On("RunReadQuery", ctx, stmt)}
}

func (_c *SQLRunner_RunReadQuery_Call) Run(run func(ctx context.Context, stmt string)) *SQLRunner_RunReadQuery_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *SQLRunner_RunReadQuery_Call) Return(_a0 *tableland.UserRows, _a1 error) *SQLRunner_RunReadQuery_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewSQLRunner interface {
	mock.TestingT
	Cleanup(func())
}

// NewSQLRunner creates a new instance of SQLRunner. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSQLRunner(t mockConstructorTestingTNewSQLRunner) *SQLRunner {
	mock := &SQLRunner{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
