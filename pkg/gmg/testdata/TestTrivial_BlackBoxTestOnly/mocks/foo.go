// Code generated by github.com/skipor/gmg - type-safe, fast and handy alternative GoMock generator. DO NOT EDIT.
// Source: pkg.Foo

package mocks_pkg

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// NewMockFoo creates a new GoMock for pkg.Foo.
func NewMockFoo(ctrl *gomock.Controller) *MockFoo {
	return &MockFoo{ctrl: ctrl}
}

// MockFoo is a GoMock of pkg.Foo.
type MockFoo struct{ ctrl *gomock.Controller }

// EXPECT returns GoMock recorder.
func (m_ *MockFoo) EXPECT() *MockFooMockRecorder {
	return (*MockFooMockRecorder)(m_)
}

// Bar implements mocked interface.
func (m_ *MockFoo) Bar() string {
	m_.ctrl.T.Helper()
	res_ := m_.ctrl.Call(m_, "Bar")
	res0, _ := res_[0].(string)
	return res0
}

// MockFooMockRecorder is the mock recorder for MockFoo.
type MockFooMockRecorder MockFoo

//   Bar() string
func (r_ *MockFooMockRecorder) Bar() MockFooBarCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "Bar", reflect.TypeOf((*MockFoo)(nil).Bar))
	return MockFooBarCall{call}
}

// MockFooBarCall is type safe wrapper of *gomock.Call.
type MockFooBarCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooBarCall) DoAndReturn(f func() string) MockFooBarCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooBarCall) Do(f func()) MockFooBarCall {
	c_.Call.Do(f)
	return c_
}

// Return is type safe wrapper of *gomock.Call Return.
func (c_ MockFooBarCall) Return(res0 string) MockFooBarCall {
	c_.Call.Return(res0)
	return c_
}

func (r_ *MockFooMockRecorder) mock() *MockFoo {
	return (*MockFoo)(r_)
}
