// Code generated by github.com/skipor/gmg - type-safe, fast and handy alternative GoMock generator. DO NOT EDIT.
// Source: github.com/skipor/gmg/examples/4_all-file.A2

package mocks_example

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// NewMockA2 creates a new GoMock for github.com/skipor/gmg/examples/4_all-file.A2.
func NewMockA2(ctrl *gomock.Controller) *MockA2 {
	return &MockA2{ctrl: ctrl}
}

// MockA2 is a GoMock of github.com/skipor/gmg/examples/4_all-file.A2.
type MockA2 struct{ ctrl *gomock.Controller }

// EXPECT returns GoMock recorder.
func (m_ *MockA2) EXPECT() *MockA2MockRecorder {
	return (*MockA2MockRecorder)(m_)
}

// A2 implements mocked interface.
func (m_ *MockA2) A2() {
	m_.ctrl.T.Helper()
	m_.ctrl.Call(m_, "A2")
	return
}

// MockA2MockRecorder is the mock recorder for MockA2.
type MockA2MockRecorder MockA2

// A2()
func (r_ *MockA2MockRecorder) A2() MockA2A2Call {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "A2", reflect.TypeOf((*MockA2)(nil).A2))
	return MockA2A2Call{call}
}

// MockA2A2Call is type safe wrapper of *gomock.Call.
type MockA2A2Call struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockA2A2Call) DoAndReturn(f func()) MockA2A2Call {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockA2A2Call) Do(f func()) MockA2A2Call {
	c_.Call.Do(f)
	return c_
}

func (r_ *MockA2MockRecorder) mock() *MockA2 {
	return (*MockA2)(r_)
}
