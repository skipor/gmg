// Code generated by github.com/skipor/gmg - type-safe, fast and handy alternative GoMock generator. DO NOT EDIT.
// Source: pkg.Foo

package mocks_pkg

import (
	context "context"
	reflect "reflect"
	testing "testing"

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

// AfterOtherPackagesNamesArgs implements mocked interface.
func (m_ *MockFoo) AfterOtherPackagesNamesArgs(context int) {
	m_.ctrl.T.Helper()
	m_.ctrl.Call(m_, "AfterOtherPackagesNamesArgs", context)
	return
}

// AfterOtherPackagesNamesResults implements mocked interface.
func (m_ *MockFoo) AfterOtherPackagesNamesResults() (context int) {
	m_.ctrl.T.Helper()
	res_ := m_.ctrl.Call(m_, "AfterOtherPackagesNamesResults")
	context, _ = res_[0].(int)
	return context
}

// BeforeOtherPackagesNamesArgs implements mocked interface.
func (m_ *MockFoo) BeforeOtherPackagesNamesArgs(testing int) {
	m_.ctrl.T.Helper()
	m_.ctrl.Call(m_, "BeforeOtherPackagesNamesArgs", testing)
	return
}

// BeforeOtherPackagesNamesResults implements mocked interface.
func (m_ *MockFoo) BeforeOtherPackagesNamesResults() (testing int) {
	m_.ctrl.T.Helper()
	res_ := m_.ctrl.Call(m_, "BeforeOtherPackagesNamesResults")
	testing, _ = res_[0].(int)
	return testing
}

// NamedArgsAndResults implements mocked interface.
func (m_ *MockFoo) NamedArgsAndResults(a int) (b int) {
	m_.ctrl.T.Helper()
	res_ := m_.ctrl.Call(m_, "NamedArgsAndResults", a)
	b, _ = res_[0].(int)
	return b
}

// NoArgsAndResults implements mocked interface.
func (m_ *MockFoo) NoArgsAndResults() {
	m_.ctrl.T.Helper()
	m_.ctrl.Call(m_, "NoArgsAndResults")
	return
}

// ReservedArgNames implements mocked interface.
func (m_ *MockFoo) ReservedArgNames(c int, r int, m int, res int, call int, reflect2 int, gomock2 int) {
	m_.ctrl.T.Helper()
	m_.ctrl.Call(m_, "ReservedArgNames", c, r, m, res, call, reflect2, gomock2)
	return
}

// ReservedResultNames implements mocked interface.
func (m_ *MockFoo) ReservedResultNames() (c int, r int, m int, res int, call int, reflect2 int, gomock2 int) {
	m_.ctrl.T.Helper()
	res_ := m_.ctrl.Call(m_, "ReservedResultNames")
	c, _ = res_[0].(int)
	r, _ = res_[1].(int)
	m, _ = res_[2].(int)
	res, _ = res_[3].(int)
	call, _ = res_[4].(int)
	reflect2, _ = res_[5].(int)
	gomock2, _ = res_[6].(int)
	return c, r, m, res, call, reflect2, gomock2
}

// UnderscoreArgsAndResults implements mocked interface.
func (m_ *MockFoo) UnderscoreArgsAndResults(arg int) (_ int) {
	m_.ctrl.T.Helper()
	res_ := m_.ctrl.Call(m_, "UnderscoreArgsAndResults", arg)
	res0, _ := res_[0].(int)
	return res0
}

// VariadicArgs implements mocked interface.
func (m_ *MockFoo) VariadicArgs(f string, as ...int) {
	m_.ctrl.T.Helper()
	m_.ctrl.Call(m_, "VariadicArgs", f, as)
	return
}

// WellKnownNamesArgs implements mocked interface.
func (m_ *MockFoo) WellKnownNamesArgs(arg context.Context, arg2 *testing.T, arg3 error) {
	m_.ctrl.T.Helper()
	m_.ctrl.Call(m_, "WellKnownNamesArgs", arg, arg2, arg3)
	return
}

// WellKnownNamesResults implements mocked interface.
func (m_ *MockFoo) WellKnownNamesResults() (context.Context, *testing.T, error) {
	m_.ctrl.T.Helper()
	res_ := m_.ctrl.Call(m_, "WellKnownNamesResults")
	res0, _ := res_[0].(context.Context)
	res1, _ := res_[1].(*testing.T)
	res2, _ := res_[2].(error)
	return res0, res1, res2
}

// MockFooMockRecorder is the mock recorder for MockFoo.
type MockFooMockRecorder MockFoo

//   AfterOtherPackagesNamesArgs(context int)
func (r_ *MockFooMockRecorder) AfterOtherPackagesNamesArgs(context2 interface{}) MockFooAfterOtherPackagesNamesArgsCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "AfterOtherPackagesNamesArgs", reflect.TypeOf((*MockFoo)(nil).AfterOtherPackagesNamesArgs), context2)
	return MockFooAfterOtherPackagesNamesArgsCall{call}
}

// MockFooAfterOtherPackagesNamesArgsCall is type safe wrapper of *gomock.Call.
type MockFooAfterOtherPackagesNamesArgsCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooAfterOtherPackagesNamesArgsCall) DoAndReturn(f func(context2 int)) MockFooAfterOtherPackagesNamesArgsCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooAfterOtherPackagesNamesArgsCall) Do(f func(context2 int)) MockFooAfterOtherPackagesNamesArgsCall {
	c_.Call.Do(f)
	return c_
}

//   AfterOtherPackagesNamesResults() (context int)
func (r_ *MockFooMockRecorder) AfterOtherPackagesNamesResults() MockFooAfterOtherPackagesNamesResultsCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "AfterOtherPackagesNamesResults", reflect.TypeOf((*MockFoo)(nil).AfterOtherPackagesNamesResults))
	return MockFooAfterOtherPackagesNamesResultsCall{call}
}

// MockFooAfterOtherPackagesNamesResultsCall is type safe wrapper of *gomock.Call.
type MockFooAfterOtherPackagesNamesResultsCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooAfterOtherPackagesNamesResultsCall) DoAndReturn(f func() (context2 int)) MockFooAfterOtherPackagesNamesResultsCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooAfterOtherPackagesNamesResultsCall) Do(f func()) MockFooAfterOtherPackagesNamesResultsCall {
	c_.Call.Do(f)
	return c_
}

// Return is type safe wrapper of *gomock.Call Return.
func (c_ MockFooAfterOtherPackagesNamesResultsCall) Return(context2 int) MockFooAfterOtherPackagesNamesResultsCall {
	c_.Call.Return(context2)
	return c_
}

//   BeforeOtherPackagesNamesArgs(testing int)
func (r_ *MockFooMockRecorder) BeforeOtherPackagesNamesArgs(testing2 interface{}) MockFooBeforeOtherPackagesNamesArgsCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "BeforeOtherPackagesNamesArgs", reflect.TypeOf((*MockFoo)(nil).BeforeOtherPackagesNamesArgs), testing2)
	return MockFooBeforeOtherPackagesNamesArgsCall{call}
}

// MockFooBeforeOtherPackagesNamesArgsCall is type safe wrapper of *gomock.Call.
type MockFooBeforeOtherPackagesNamesArgsCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooBeforeOtherPackagesNamesArgsCall) DoAndReturn(f func(testing2 int)) MockFooBeforeOtherPackagesNamesArgsCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooBeforeOtherPackagesNamesArgsCall) Do(f func(testing2 int)) MockFooBeforeOtherPackagesNamesArgsCall {
	c_.Call.Do(f)
	return c_
}

//   BeforeOtherPackagesNamesResults() (testing int)
func (r_ *MockFooMockRecorder) BeforeOtherPackagesNamesResults() MockFooBeforeOtherPackagesNamesResultsCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "BeforeOtherPackagesNamesResults", reflect.TypeOf((*MockFoo)(nil).BeforeOtherPackagesNamesResults))
	return MockFooBeforeOtherPackagesNamesResultsCall{call}
}

// MockFooBeforeOtherPackagesNamesResultsCall is type safe wrapper of *gomock.Call.
type MockFooBeforeOtherPackagesNamesResultsCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooBeforeOtherPackagesNamesResultsCall) DoAndReturn(f func() (testing2 int)) MockFooBeforeOtherPackagesNamesResultsCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooBeforeOtherPackagesNamesResultsCall) Do(f func()) MockFooBeforeOtherPackagesNamesResultsCall {
	c_.Call.Do(f)
	return c_
}

// Return is type safe wrapper of *gomock.Call Return.
func (c_ MockFooBeforeOtherPackagesNamesResultsCall) Return(testing2 int) MockFooBeforeOtherPackagesNamesResultsCall {
	c_.Call.Return(testing2)
	return c_
}

//   NamedArgsAndResults(a int) (b int)
func (r_ *MockFooMockRecorder) NamedArgsAndResults(a interface{}) MockFooNamedArgsAndResultsCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "NamedArgsAndResults", reflect.TypeOf((*MockFoo)(nil).NamedArgsAndResults), a)
	return MockFooNamedArgsAndResultsCall{call}
}

// MockFooNamedArgsAndResultsCall is type safe wrapper of *gomock.Call.
type MockFooNamedArgsAndResultsCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooNamedArgsAndResultsCall) DoAndReturn(f func(a int) (b int)) MockFooNamedArgsAndResultsCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooNamedArgsAndResultsCall) Do(f func(a int)) MockFooNamedArgsAndResultsCall {
	c_.Call.Do(f)
	return c_
}

// Return is type safe wrapper of *gomock.Call Return.
func (c_ MockFooNamedArgsAndResultsCall) Return(b int) MockFooNamedArgsAndResultsCall {
	c_.Call.Return(b)
	return c_
}

//   NoArgsAndResults()
func (r_ *MockFooMockRecorder) NoArgsAndResults() MockFooNoArgsAndResultsCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "NoArgsAndResults", reflect.TypeOf((*MockFoo)(nil).NoArgsAndResults))
	return MockFooNoArgsAndResultsCall{call}
}

// MockFooNoArgsAndResultsCall is type safe wrapper of *gomock.Call.
type MockFooNoArgsAndResultsCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooNoArgsAndResultsCall) DoAndReturn(f func()) MockFooNoArgsAndResultsCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooNoArgsAndResultsCall) Do(f func()) MockFooNoArgsAndResultsCall {
	c_.Call.Do(f)
	return c_
}

//   ReservedArgNames(c int, r int, m int, res int, call int, reflect int, gomock int)
func (r_ *MockFooMockRecorder) ReservedArgNames(c interface{}, r interface{}, m interface{}, res interface{}, call interface{}, reflect2 interface{}, gomock2 interface{}) MockFooReservedArgNamesCall {
	r_.ctrl.T.Helper()
	call2 := r_.ctrl.RecordCallWithMethodType(r_.mock(), "ReservedArgNames", reflect.TypeOf((*MockFoo)(nil).ReservedArgNames), c, r, m, res, call, reflect2, gomock2)
	return MockFooReservedArgNamesCall{call2}
}

// MockFooReservedArgNamesCall is type safe wrapper of *gomock.Call.
type MockFooReservedArgNamesCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooReservedArgNamesCall) DoAndReturn(f func(c int, r int, m int, res int, call int, reflect2 int, gomock2 int)) MockFooReservedArgNamesCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooReservedArgNamesCall) Do(f func(c int, r int, m int, res int, call int, reflect2 int, gomock2 int)) MockFooReservedArgNamesCall {
	c_.Call.Do(f)
	return c_
}

//   ReservedResultNames() (c int, r int, m int, res int, call int, reflect int, gomock int)
func (r_ *MockFooMockRecorder) ReservedResultNames() MockFooReservedResultNamesCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "ReservedResultNames", reflect.TypeOf((*MockFoo)(nil).ReservedResultNames))
	return MockFooReservedResultNamesCall{call}
}

// MockFooReservedResultNamesCall is type safe wrapper of *gomock.Call.
type MockFooReservedResultNamesCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooReservedResultNamesCall) DoAndReturn(f func() (c int, r int, m int, res int, call int, reflect2 int, gomock2 int)) MockFooReservedResultNamesCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooReservedResultNamesCall) Do(f func()) MockFooReservedResultNamesCall {
	c_.Call.Do(f)
	return c_
}

// Return is type safe wrapper of *gomock.Call Return.
func (c_ MockFooReservedResultNamesCall) Return(c int, r int, m int, res int, call int, reflect2 int, gomock2 int) MockFooReservedResultNamesCall {
	c_.Call.Return(c, r, m, res, call, reflect2, gomock2)
	return c_
}

//   UnderscoreArgsAndResults(_ int) (_ int)
func (r_ *MockFooMockRecorder) UnderscoreArgsAndResults(arg interface{}) MockFooUnderscoreArgsAndResultsCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "UnderscoreArgsAndResults", reflect.TypeOf((*MockFoo)(nil).UnderscoreArgsAndResults), arg)
	return MockFooUnderscoreArgsAndResultsCall{call}
}

// MockFooUnderscoreArgsAndResultsCall is type safe wrapper of *gomock.Call.
type MockFooUnderscoreArgsAndResultsCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooUnderscoreArgsAndResultsCall) DoAndReturn(f func(arg int) (_ int)) MockFooUnderscoreArgsAndResultsCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooUnderscoreArgsAndResultsCall) Do(f func(arg int)) MockFooUnderscoreArgsAndResultsCall {
	c_.Call.Do(f)
	return c_
}

// Return is type safe wrapper of *gomock.Call Return.
func (c_ MockFooUnderscoreArgsAndResultsCall) Return(res0 int) MockFooUnderscoreArgsAndResultsCall {
	c_.Call.Return(res0)
	return c_
}

//   VariadicArgs(f string, as ...int)
func (r_ *MockFooMockRecorder) VariadicArgs(f interface{}, as interface{}) MockFooVariadicArgsCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "VariadicArgs", reflect.TypeOf((*MockFoo)(nil).VariadicArgs), f, as)
	return MockFooVariadicArgsCall{call}
}

// MockFooVariadicArgsCall is type safe wrapper of *gomock.Call.
type MockFooVariadicArgsCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooVariadicArgsCall) DoAndReturn(f func(f string, as ...int)) MockFooVariadicArgsCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooVariadicArgsCall) Do(f func(f string, as ...int)) MockFooVariadicArgsCall {
	c_.Call.Do(f)
	return c_
}

//   WellKnownNamesArgs(context.Context, *testing.T, error)
func (r_ *MockFooMockRecorder) WellKnownNamesArgs(arg interface{}, arg2 interface{}, arg3 interface{}) MockFooWellKnownNamesArgsCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "WellKnownNamesArgs", reflect.TypeOf((*MockFoo)(nil).WellKnownNamesArgs), arg, arg2, arg3)
	return MockFooWellKnownNamesArgsCall{call}
}

// MockFooWellKnownNamesArgsCall is type safe wrapper of *gomock.Call.
type MockFooWellKnownNamesArgsCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooWellKnownNamesArgsCall) DoAndReturn(f func(arg context.Context, arg2 *testing.T, arg3 error)) MockFooWellKnownNamesArgsCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooWellKnownNamesArgsCall) Do(f func(arg context.Context, arg2 *testing.T, arg3 error)) MockFooWellKnownNamesArgsCall {
	c_.Call.Do(f)
	return c_
}

//   WellKnownNamesResults() (context.Context, *testing.T, error)
func (r_ *MockFooMockRecorder) WellKnownNamesResults() MockFooWellKnownNamesResultsCall {
	r_.ctrl.T.Helper()
	call := r_.ctrl.RecordCallWithMethodType(r_.mock(), "WellKnownNamesResults", reflect.TypeOf((*MockFoo)(nil).WellKnownNamesResults))
	return MockFooWellKnownNamesResultsCall{call}
}

// MockFooWellKnownNamesResultsCall is type safe wrapper of *gomock.Call.
type MockFooWellKnownNamesResultsCall struct{ *gomock.Call }

// DoAndReturn is type safe wrapper of *gomock.Call DoAndReturn.
func (c_ MockFooWellKnownNamesResultsCall) DoAndReturn(f func() (context.Context, *testing.T, error)) MockFooWellKnownNamesResultsCall {
	c_.Call.DoAndReturn(f)
	return c_
}

// Do is type safe wrapper of *gomock.Call Do.
func (c_ MockFooWellKnownNamesResultsCall) Do(f func()) MockFooWellKnownNamesResultsCall {
	c_.Call.Do(f)
	return c_
}

// Return is type safe wrapper of *gomock.Call Return.
func (c_ MockFooWellKnownNamesResultsCall) Return(res0 context.Context, res1 *testing.T, res2 error) MockFooWellKnownNamesResultsCall {
	c_.Call.Return(res0, res1, res2)
	return c_
}

func (r_ *MockFooMockRecorder) mock() *MockFoo {
	return (*MockFoo)(r_)
}
