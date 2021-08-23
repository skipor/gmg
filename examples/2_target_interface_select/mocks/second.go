// Code generated by github.com/skipor/gmg - type-safe, fast and handy alternative GoMock generator. DO NOT EDIT.
// Source: github.com/skipor/gmg/examples/2_target_interface_select.Second

package mocks_target_interface_selection

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// NewMockSecond creates a new GoMock for github.com/skipor/gmg/examples/2_target_interface_select.Second.
func NewMockSecond(ctrl *gomock.Controller) *MockSecond {
	return &MockSecond{ctrl: ctrl}
}

// MockSecond is a GoMock of github.com/skipor/gmg/examples/2_target_interface_select.Second.
type MockSecond struct{ ctrl *gomock.Controller }

// EXPECT returns GoMock recorder.
func (m_ *MockSecond) EXPECT() *MockSecondMockRecorder {
	return (*MockSecondMockRecorder)(m_)
}

// Two implements mocked interface.
func (m_ *MockSecond) Two() {
	m_.ctrl.T.Helper()
	m_.ctrl.Call(m_, "Two")
	return
}

// MockSecondMockRecorder is the mock recorder for MockSecond.
type MockSecondMockRecorder MockSecond

//   Two()
func (r_ *MockSecondMockRecorder) Two() MockSecondTwoCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "Two", reflect.TypeOf((*MockSecond)(nil).Two))
	return MockSecondTwoCall{call}
}

// MockSecondTwoCall is type safe wrapper of *gomock.Call.
type MockSecondTwoCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockSecondTwoCall) DoAndReturn(f func()) MockSecondTwoCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockSecondTwoCall) Do(f func()) MockSecondTwoCall {
	c_.Call.Do(f)
	return c_
}

func (r_ *MockSecondMockRecorder) mock() *MockSecond {
	return (*MockSecond)(r_)
}