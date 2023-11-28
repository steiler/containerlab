// Code generated by MockGen. DO NOT EDIT.
// Source: clab/dependency_manager/dependency_manager.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	dependency_manager "github.com/srl-labs/containerlab/clab/dependency_manager"
)

// MockDependencyManager is a mock of DependencyManager interface.
type MockDependencyManager struct {
	ctrl     *gomock.Controller
	recorder *MockDependencyManagerMockRecorder
}

// MockDependencyManagerMockRecorder is the mock recorder for MockDependencyManager.
type MockDependencyManagerMockRecorder struct {
	mock *MockDependencyManager
}

// NewMockDependencyManager creates a new mock instance.
func NewMockDependencyManager(ctrl *gomock.Controller) *MockDependencyManager {
	mock := &MockDependencyManager{ctrl: ctrl}
	mock.recorder = &MockDependencyManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDependencyManager) EXPECT() *MockDependencyManagerMockRecorder {
	return m.recorder
}

// AddDependency mocks base method.
func (m *MockDependencyManager) AddDependency(depender, dependee string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddDependency", depender, dependee)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddDependency indicates an expected call of AddDependency.
func (mr *MockDependencyManagerMockRecorder) AddDependency(depender, dependee interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddDependency", reflect.TypeOf((*MockDependencyManager)(nil).AddDependency), depender, dependee)
}

// AddNode mocks base method.
func (m *MockDependencyManager) AddNode(name string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddNode", name)
}

// AddNode indicates an expected call of AddNode.
func (mr *MockDependencyManagerMockRecorder) AddNode(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddNode", reflect.TypeOf((*MockDependencyManager)(nil).AddNode), name)
}

// CheckAcyclicity mocks base method.
func (m *MockDependencyManager) CheckAcyclicity() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckAcyclicity")
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckAcyclicity indicates an expected call of CheckAcyclicity.
func (mr *MockDependencyManagerMockRecorder) CheckAcyclicity() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckAcyclicity", reflect.TypeOf((*MockDependencyManager)(nil).CheckAcyclicity))
}

// SignalDone mocks base method.
func (m *MockDependencyManager) SignalDone(nodeName string, state dependency_manager.NodeState) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SignalDone", nodeName, state)
}

// SignalDone indicates an expected call of SignalDone.
func (mr *MockDependencyManagerMockRecorder) SignalDone(nodeName, state interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignalDone", reflect.TypeOf((*MockDependencyManager)(nil).SignalDone), nodeName, state)
}

// String mocks base method.
func (m *MockDependencyManager) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockDependencyManagerMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockDependencyManager)(nil).String))
}

// WaitForNodeDependencies mocks base method.
func (m *MockDependencyManager) WaitForNodeDependencies(nodeName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForNodeDependencies", nodeName)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitForNodeDependencies indicates an expected call of WaitForNodeDependencies.
func (mr *MockDependencyManagerMockRecorder) WaitForNodeDependencies(nodeName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForNodeDependencies", reflect.TypeOf((*MockDependencyManager)(nil).WaitForNodeDependencies), nodeName)
}
