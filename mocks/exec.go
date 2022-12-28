// Code generated by MockGen. DO NOT EDIT.
// Source: clab/exec/exec.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	exec "github.com/srl-labs/containerlab/clab/exec"
)

// MockExecResultHolderSetter is a mock of ExecResultHolderSetter interface.
type MockExecResultHolderSetter struct {
	ctrl     *gomock.Controller
	recorder *MockExecResultHolderSetterMockRecorder
}

// MockExecResultHolderSetterMockRecorder is the mock recorder for MockExecResultHolderSetter.
type MockExecResultHolderSetterMockRecorder struct {
	mock *MockExecResultHolderSetter
}

// NewMockExecResultHolderSetter creates a new mock instance.
func NewMockExecResultHolderSetter(ctrl *gomock.Controller) *MockExecResultHolderSetter {
	mock := &MockExecResultHolderSetter{ctrl: ctrl}
	mock.recorder = &MockExecResultHolderSetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExecResultHolderSetter) EXPECT() *MockExecResultHolderSetterMockRecorder {
	return m.recorder
}

// GetExecResultHolder mocks base method.
func (m *MockExecResultHolderSetter) GetExecResultHolder() exec.ExecResultHolder {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExecResultHolder")
	ret0, _ := ret[0].(exec.ExecResultHolder)
	return ret0
}

// GetExecResultHolder indicates an expected call of GetExecResultHolder.
func (mr *MockExecResultHolderSetterMockRecorder) GetExecResultHolder() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExecResultHolder", reflect.TypeOf((*MockExecResultHolderSetter)(nil).GetExecResultHolder))
}

// SetReturnCode mocks base method.
func (m *MockExecResultHolderSetter) SetReturnCode(arg0 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetReturnCode", arg0)
}

// SetReturnCode indicates an expected call of SetReturnCode.
func (mr *MockExecResultHolderSetterMockRecorder) SetReturnCode(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetReturnCode", reflect.TypeOf((*MockExecResultHolderSetter)(nil).SetReturnCode), arg0)
}

// SetStdErr mocks base method.
func (m *MockExecResultHolderSetter) SetStdErr(arg0 []byte) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetStdErr", arg0)
}

// SetStdErr indicates an expected call of SetStdErr.
func (mr *MockExecResultHolderSetterMockRecorder) SetStdErr(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStdErr", reflect.TypeOf((*MockExecResultHolderSetter)(nil).SetStdErr), arg0)
}

// SetStdOut mocks base method.
func (m *MockExecResultHolderSetter) SetStdOut(arg0 []byte) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetStdOut", arg0)
}

// SetStdOut indicates an expected call of SetStdOut.
func (mr *MockExecResultHolderSetterMockRecorder) SetStdOut(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStdOut", reflect.TypeOf((*MockExecResultHolderSetter)(nil).SetStdOut), arg0)
}

// MockExecResultHolder is a mock of ExecResultHolder interface.
type MockExecResultHolder struct {
	ctrl     *gomock.Controller
	recorder *MockExecResultHolderMockRecorder
}

// MockExecResultHolderMockRecorder is the mock recorder for MockExecResultHolder.
type MockExecResultHolderMockRecorder struct {
	mock *MockExecResultHolder
}

// NewMockExecResultHolder creates a new mock instance.
func NewMockExecResultHolder(ctrl *gomock.Controller) *MockExecResultHolder {
	mock := &MockExecResultHolder{ctrl: ctrl}
	mock.recorder = &MockExecResultHolderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExecResultHolder) EXPECT() *MockExecResultHolderMockRecorder {
	return m.recorder
}

// Dump mocks base method.
func (m *MockExecResultHolder) Dump(format exec.ExecFormat) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dump", format)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Dump indicates an expected call of Dump.
func (mr *MockExecResultHolderMockRecorder) Dump(format interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dump", reflect.TypeOf((*MockExecResultHolder)(nil).Dump), format)
}

// GetCmdString mocks base method.
func (m *MockExecResultHolder) GetCmdString() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCmdString")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetCmdString indicates an expected call of GetCmdString.
func (mr *MockExecResultHolderMockRecorder) GetCmdString() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCmdString", reflect.TypeOf((*MockExecResultHolder)(nil).GetCmdString))
}

// GetReturnCode mocks base method.
func (m *MockExecResultHolder) GetReturnCode() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetReturnCode")
	ret0, _ := ret[0].(int)
	return ret0
}

// GetReturnCode indicates an expected call of GetReturnCode.
func (mr *MockExecResultHolderMockRecorder) GetReturnCode() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetReturnCode", reflect.TypeOf((*MockExecResultHolder)(nil).GetReturnCode))
}

// GetStdErrByteSlice mocks base method.
func (m *MockExecResultHolder) GetStdErrByteSlice() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStdErrByteSlice")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// GetStdErrByteSlice indicates an expected call of GetStdErrByteSlice.
func (mr *MockExecResultHolderMockRecorder) GetStdErrByteSlice() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStdErrByteSlice", reflect.TypeOf((*MockExecResultHolder)(nil).GetStdErrByteSlice))
}

// GetStdErrString mocks base method.
func (m *MockExecResultHolder) GetStdErrString() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStdErrString")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetStdErrString indicates an expected call of GetStdErrString.
func (mr *MockExecResultHolderMockRecorder) GetStdErrString() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStdErrString", reflect.TypeOf((*MockExecResultHolder)(nil).GetStdErrString))
}

// GetStdOutByteSlice mocks base method.
func (m *MockExecResultHolder) GetStdOutByteSlice() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStdOutByteSlice")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// GetStdOutByteSlice indicates an expected call of GetStdOutByteSlice.
func (mr *MockExecResultHolderMockRecorder) GetStdOutByteSlice() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStdOutByteSlice", reflect.TypeOf((*MockExecResultHolder)(nil).GetStdOutByteSlice))
}

// GetStdOutString mocks base method.
func (m *MockExecResultHolder) GetStdOutString() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStdOutString")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetStdOutString indicates an expected call of GetStdOutString.
func (mr *MockExecResultHolderMockRecorder) GetStdOutString() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStdOutString", reflect.TypeOf((*MockExecResultHolder)(nil).GetStdOutString))
}

// String mocks base method.
func (m *MockExecResultHolder) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockExecResultHolderMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockExecResultHolder)(nil).String))
}
