// Code generated by github.com/skipor/gmg - type-safe, fast and handy alternative GoMock generator. DO NOT EDIT.
// Source: github.com/skipor/gmg/examples/3_generate_all_interfaces/1_primary.Second

package mocks_primary

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// NewMockSecond creates a new GoMock for github.com/skipor/gmg/examples/3_generate_all_interfaces/1_primary.Second.
func NewMockSecond(ctrl *gomock.Controller) *MockSecond {
	return &MockSecond{ctrl: ctrl}
}

// MockSecond is a GoMock of github.com/skipor/gmg/examples/3_generate_all_interfaces/1_primary.Second.
type MockSecond struct{ ctrl *gomock.Controller }

// EXPECT returns GoMock recorder.
func (m_ *MockSecond) EXPECT() *MockSecondMockRecorder {
	return (*MockSecondMockRecorder)(m_)
}

// Foo implements mocked interface.
func (m_ *MockSecond) Foo() {
	m_.ctrl.T.Helper()
	m_.ctrl.Call(m_, "Foo")
	return
}

// MockSecondMockRecorder is the mock recorder for MockSecond.
type MockSecondMockRecorder MockSecond

//   Foo()
func (r_ *MockSecondMockRecorder) Foo() MockSecondFooCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "Foo", reflect.TypeOf((*MockSecond)(nil).Foo))
	return MockSecondFooCall{call}
}

// MockSecondFooCall is type safe wrapper of *gomock.Call.
type MockSecondFooCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockSecondFooCall) DoAndReturn(f func()) MockSecondFooCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockSecondFooCall) Do(f func()) MockSecondFooCall {
	c_.Call.Do(f)
	return c_
}

func (r_ *MockSecondMockRecorder) mock() *MockSecond {
	return (*MockSecond)(r_)
}
