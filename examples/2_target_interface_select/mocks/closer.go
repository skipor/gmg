// Code generated by github.com/skipor/gmg - type-safe, fast and handy alternative GoMock generator. DO NOT EDIT.
// Source: io.Closer

package example_mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// NewMockCloser creates a new GoMock for io.Closer.
func NewMockCloser(ctrl *gomock.Controller) *MockCloser {
	return &MockCloser{ctrl: ctrl}
}

// MockCloser is a GoMock of io.Closer.
type MockCloser struct{ ctrl *gomock.Controller }

// EXPECT returns GoMock recorder.
func (m_ *MockCloser) EXPECT() *MockCloserMockRecorder {
	return (*MockCloserMockRecorder)(m_)
}

// Close implements mocked interface.
func (m_ *MockCloser) Close() error {
	m_.ctrl.T.Helper()
	res_ := m_.ctrl.Call(m_, "Close")
	res0, _ := res_[0].(error)
	return res0
}

// MockCloserMockRecorder is the mock recorder for MockCloser.
type MockCloserMockRecorder MockCloser

//   Close() error
func (r_ *MockCloserMockRecorder) Close() MockCloserCloseCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "Close", reflect.TypeOf((*MockCloser)(nil).Close))
	return MockCloserCloseCall{call}
}

// MockCloserCloseCall is type safe wrapper of *gomock.Call.
type MockCloserCloseCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockCloserCloseCall) DoAndReturn(f func() error) MockCloserCloseCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockCloserCloseCall) Do(f func()) MockCloserCloseCall {
	c_.Call.Do(f)
	return c_
}

// Return is type safe wrapper of *gomock.Call Return.
func (c_ MockCloserCloseCall) Return(res0 error) MockCloserCloseCall {
	c_.Call.Return(res0)
	return c_
}

func (r_ *MockCloserMockRecorder) mock() *MockCloser {
	return (*MockCloser)(r_)
}
