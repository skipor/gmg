// Code generated by github.com/skipor/gmg - type-safe, fast and handy alternative GoMock generator. DO NOT EDIT.
// Source: github.com/skipor/gmg/examples/3_all.First

package mocks_example

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// NewMockFirst creates a new GoMock for github.com/skipor/gmg/examples/3_all.First.
func NewMockFirst(ctrl *gomock.Controller) *MockFirst {
	return &MockFirst{ctrl: ctrl}
}

// MockFirst is a GoMock of github.com/skipor/gmg/examples/3_all.First.
type MockFirst struct{ ctrl *gomock.Controller }

// EXPECT returns GoMock recorder.
func (m_ *MockFirst) EXPECT() *MockFirstMockRecorder {
	return (*MockFirstMockRecorder)(m_)
}

// Bar implements mocked interface.
func (m_ *MockFirst) Bar() {
	m_.ctrl.T.Helper()
	m_.ctrl.Call(m_, "Bar")
	return
}

// MockFirstMockRecorder is the mock recorder for MockFirst.
type MockFirstMockRecorder MockFirst

// Bar()
func (r_ *MockFirstMockRecorder) Bar() MockFirstBarCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "Bar", reflect.TypeOf((*MockFirst)(nil).Bar))
	return MockFirstBarCall{call}
}

// MockFirstBarCall is type safe wrapper of *gomock.Call.
type MockFirstBarCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFirstBarCall) DoAndReturn(f func()) MockFirstBarCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFirstBarCall) Do(f func()) MockFirstBarCall {
	c_.Call.Do(f)
	return c_
}

func (r_ *MockFirstMockRecorder) mock() *MockFirst {
	return (*MockFirst)(r_)
}
