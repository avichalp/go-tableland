// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	common "github.com/ethereum/go-ethereum/common"

	mock "github.com/stretchr/testify/mock"

	sqlstore "github.com/textileio/go-tableland/pkg/sqlstore"

	tableland "github.com/textileio/go-tableland/internal/tableland"

	tables "github.com/textileio/go-tableland/pkg/tables"
)

// Gateway is an autogenerated mock type for the Gateway type
type Gateway struct {
	mock.Mock
}

type Gateway_Expecter struct {
	mock *mock.Mock
}

func (_m *Gateway) EXPECT() *Gateway_Expecter {
	return &Gateway_Expecter{mock: &_m.Mock}
}

// GetReceiptByTransactionHash provides a mock function with given fields: _a0, _a1
func (_m *Gateway) GetReceiptByTransactionHash(_a0 context.Context, _a1 common.Hash) (sqlstore.Receipt, bool, error) {
	ret := _m.Called(_a0, _a1)

	var r0 sqlstore.Receipt
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) sqlstore.Receipt); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(sqlstore.Receipt)
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(context.Context, common.Hash) bool); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Get(1).(bool)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, common.Hash) error); ok {
		r2 = rf(_a0, _a1)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Gateway_GetReceiptByTransactionHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetReceiptByTransactionHash'
type Gateway_GetReceiptByTransactionHash_Call struct {
	*mock.Call
}

// GetReceiptByTransactionHash is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 common.Hash
func (_e *Gateway_Expecter) GetReceiptByTransactionHash(_a0 interface{}, _a1 interface{}) *Gateway_GetReceiptByTransactionHash_Call {
	return &Gateway_GetReceiptByTransactionHash_Call{Call: _e.mock.On("GetReceiptByTransactionHash", _a0, _a1)}
}

func (_c *Gateway_GetReceiptByTransactionHash_Call) Run(run func(_a0 context.Context, _a1 common.Hash)) *Gateway_GetReceiptByTransactionHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Hash))
	})
	return _c
}

func (_c *Gateway_GetReceiptByTransactionHash_Call) Return(_a0 sqlstore.Receipt, _a1 bool, _a2 error) *Gateway_GetReceiptByTransactionHash_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

// GetTableMetadata provides a mock function with given fields: _a0, _a1
func (_m *Gateway) GetTableMetadata(_a0 context.Context, _a1 tables.TableID) (sqlstore.TableMetadata, error) {
	ret := _m.Called(_a0, _a1)

	var r0 sqlstore.TableMetadata
	if rf, ok := ret.Get(0).(func(context.Context, tables.TableID) sqlstore.TableMetadata); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(sqlstore.TableMetadata)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, tables.TableID) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Gateway_GetTableMetadata_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTableMetadata'
type Gateway_GetTableMetadata_Call struct {
	*mock.Call
}

// GetTableMetadata is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 tables.TableID
func (_e *Gateway_Expecter) GetTableMetadata(_a0 interface{}, _a1 interface{}) *Gateway_GetTableMetadata_Call {
	return &Gateway_GetTableMetadata_Call{Call: _e.mock.On("GetTableMetadata", _a0, _a1)}
}

func (_c *Gateway_GetTableMetadata_Call) Run(run func(_a0 context.Context, _a1 tables.TableID)) *Gateway_GetTableMetadata_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(tables.TableID))
	})
	return _c
}

func (_c *Gateway_GetTableMetadata_Call) Return(_a0 sqlstore.TableMetadata, _a1 error) *Gateway_GetTableMetadata_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// RunReadQuery provides a mock function with given fields: ctx, stmt
func (_m *Gateway) RunReadQuery(ctx context.Context, stmt string) (*tableland.TableData, error) {
	ret := _m.Called(ctx, stmt)

	var r0 *tableland.TableData
	if rf, ok := ret.Get(0).(func(context.Context, string) *tableland.TableData); ok {
		r0 = rf(ctx, stmt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tableland.TableData)
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

// Gateway_RunReadQuery_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RunReadQuery'
type Gateway_RunReadQuery_Call struct {
	*mock.Call
}

// RunReadQuery is a helper method to define mock.On call
//   - ctx context.Context
//   - stmt string
func (_e *Gateway_Expecter) RunReadQuery(ctx interface{}, stmt interface{}) *Gateway_RunReadQuery_Call {
	return &Gateway_RunReadQuery_Call{Call: _e.mock.On("RunReadQuery", ctx, stmt)}
}

func (_c *Gateway_RunReadQuery_Call) Run(run func(ctx context.Context, stmt string)) *Gateway_RunReadQuery_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Gateway_RunReadQuery_Call) Return(_a0 *tableland.TableData, _a1 error) *Gateway_RunReadQuery_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewGateway interface {
	mock.TestingT
	Cleanup(func())
}

// NewGateway creates a new instance of Gateway. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewGateway(t mockConstructorTestingTNewGateway) *Gateway {
	mock := &Gateway{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
