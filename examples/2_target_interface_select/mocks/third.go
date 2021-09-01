// Code generated by github.com/skipor/gmg - type-safe, fast and handy alternative GoMock generator. DO NOT EDIT.
// Source: github.com/skipor/gmg/examples/2_target_interface_select.Third

package example_mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// NewMockThird creates a new GoMock for github.com/skipor/gmg/examples/2_target_interface_select.Third.
func NewMockThird(ctrl *gomock.Controller) *MockThird {
	return &MockThird{ctrl: ctrl}
}

// MockThird is a GoMock of github.com/skipor/gmg/examples/2_target_interface_select.Third.
type MockThird struct{ ctrl *gomock.Controller }

// EXPECT returns GoMock recorder.
func (m_ *MockThird) EXPECT() *MockThirdMockRecorder {
	return (*MockThirdMockRecorder)(m_)
}

// Three implements mocked interface.
func (m_ *MockThird) Three() {
	m_.ctrl.T.Helper()
	m_.ctrl.Call(m_, "Three")
	return
}

// MockThirdMockRecorder is the mock recorder for MockThird.
type MockThirdMockRecorder MockThird

//   Three()
func (r_ *MockThirdMockRecorder) Three() MockThirdThreeCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "Three", reflect.TypeOf((*MockThird)(nil).Three))
	return MockThirdThreeCall{call}
}

// MockThirdThreeCall is type safe wrapper of *gomock.Call.
type MockThirdThreeCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockThirdThreeCall) DoAndReturn(f func()) MockThirdThreeCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockThirdThreeCall) Do(f func()) MockThirdThreeCall {
	c_.Call.Do(f)
	return c_
}

func (r_ *MockThirdMockRecorder) mock() *MockThird {
	return (*MockThird)(r_)
}
