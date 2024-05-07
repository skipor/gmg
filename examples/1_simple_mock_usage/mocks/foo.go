// Code generated by github.com/skipor/gmg - type-safe, fast and handy alternative GoMock generator. DO NOT EDIT.
// Source: github.com/skipor/gmg/examples/1_simple_mock_usage.Foo

package mocks_example

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// NewMockFoo creates a new GoMock for github.com/skipor/gmg/examples/1_simple_mock_usage.Foo.
func NewMockFoo(ctrl *gomock.Controller) *MockFoo {
	return &MockFoo{ctrl: ctrl}
}

// MockFoo is a GoMock of github.com/skipor/gmg/examples/1_simple_mock_usage.Foo.
type MockFoo struct{ ctrl *gomock.Controller }

// EXPECT returns GoMock recorder.
func (m_ *MockFoo) EXPECT() *MockFooMockRecorder {
	return (*MockFooMockRecorder)(m_)
}

// Bar implements mocked interface.
func (m_ *MockFoo) Bar(s string) error {
	m_.ctrl.T.Helper()
	res_ := m_.ctrl.Call(m_, "Bar", s)
	res0, _ := res_[0].(error)
	return res0
}

// MockFooMockRecorder is the mock recorder for MockFoo.
type MockFooMockRecorder MockFoo

// Bar(s string) error
func (r_ *MockFooMockRecorder) Bar(s interface{}) MockFooBarCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "Bar", reflect.TypeOf((*MockFoo)(nil).Bar), s)
	return MockFooBarCall{call}
}

// MockFooBarCall is type safe wrapper of *gomock.Call.
type MockFooBarCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooBarCall) DoAndReturn(f func(s string) error) MockFooBarCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooBarCall) Do(f func(s string)) MockFooBarCall {
	c_.Call.Do(f)
	return c_
}

// Return is type safe wrapper of *gomock.Call Return.
func (c_ MockFooBarCall) Return(res0 error) MockFooBarCall {
	c_.Call.Return(res0)
	return c_
}

func (r_ *MockFooMockRecorder) mock() *MockFoo {
	return (*MockFoo)(r_)
}
