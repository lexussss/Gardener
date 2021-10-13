// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/gardener/gardener/pkg/operation/botanist/component/extensions/extension (interfaces: Interface)

// Package extension is a generated GoMock package.
package extension

import (
	context "context"
	reflect "reflect"

	v1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	extension "github.com/gardener/gardener/pkg/operation/botanist/component/extensions/extension"
	gomock "github.com/golang/mock/gomock"
)

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// DeleteStaleResources mocks base method.
func (m *MockInterface) DeleteStaleResources(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteStaleResources", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteStaleResources indicates an expected call of DeleteStaleResources.
func (mr *MockInterfaceMockRecorder) DeleteStaleResources(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteStaleResources", reflect.TypeOf((*MockInterface)(nil).DeleteStaleResources), arg0)
}

// Deploy mocks base method.
func (m *MockInterface) Deploy(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Deploy", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Deploy indicates an expected call of Deploy.
func (mr *MockInterfaceMockRecorder) Deploy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Deploy", reflect.TypeOf((*MockInterface)(nil).Deploy), arg0)
}

// Destroy mocks base method.
func (m *MockInterface) Destroy(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Destroy", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Destroy indicates an expected call of Destroy.
func (mr *MockInterfaceMockRecorder) Destroy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Destroy", reflect.TypeOf((*MockInterface)(nil).Destroy), arg0)
}

// Extensions mocks base method.
func (m *MockInterface) Extensions() map[string]extension.Extension {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Extensions")
	ret0, _ := ret[0].(map[string]extension.Extension)
	return ret0
}

// Extensions indicates an expected call of Extensions.
func (mr *MockInterfaceMockRecorder) Extensions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Extensions", reflect.TypeOf((*MockInterface)(nil).Extensions))
}

// Migrate mocks base method.
func (m *MockInterface) Migrate(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Migrate", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Migrate indicates an expected call of Migrate.
func (mr *MockInterfaceMockRecorder) Migrate(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Migrate", reflect.TypeOf((*MockInterface)(nil).Migrate), arg0)
}

// Restore mocks base method.
func (m *MockInterface) Restore(arg0 context.Context, arg1 *v1alpha1.ShootState) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Restore", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Restore indicates an expected call of Restore.
func (mr *MockInterfaceMockRecorder) Restore(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Restore", reflect.TypeOf((*MockInterface)(nil).Restore), arg0, arg1)
}

// Wait mocks base method.
func (m *MockInterface) Wait(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Wait", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Wait indicates an expected call of Wait.
func (mr *MockInterfaceMockRecorder) Wait(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wait", reflect.TypeOf((*MockInterface)(nil).Wait), arg0)
}

// WaitCleanup mocks base method.
func (m *MockInterface) WaitCleanup(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitCleanup", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitCleanup indicates an expected call of WaitCleanup.
func (mr *MockInterfaceMockRecorder) WaitCleanup(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitCleanup", reflect.TypeOf((*MockInterface)(nil).WaitCleanup), arg0)
}

// WaitCleanupStaleResources mocks base method.
func (m *MockInterface) WaitCleanupStaleResources(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitCleanupStaleResources", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitCleanupStaleResources indicates an expected call of WaitCleanupStaleResources.
func (mr *MockInterfaceMockRecorder) WaitCleanupStaleResources(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitCleanupStaleResources", reflect.TypeOf((*MockInterface)(nil).WaitCleanupStaleResources), arg0)
}

// WaitMigrate mocks base method.
func (m *MockInterface) WaitMigrate(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitMigrate", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitMigrate indicates an expected call of WaitMigrate.
func (mr *MockInterfaceMockRecorder) WaitMigrate(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitMigrate", reflect.TypeOf((*MockInterface)(nil).WaitMigrate), arg0)
}
