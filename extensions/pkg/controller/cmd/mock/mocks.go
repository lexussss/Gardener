// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/gardener/gardener/extensions/pkg/controller/cmd (interfaces: Completer,Option,Flagger)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	pflag "github.com/spf13/pflag"
)

// MockCompleter is a mock of Completer interface.
type MockCompleter struct {
	ctrl     *gomock.Controller
	recorder *MockCompleterMockRecorder
}

// MockCompleterMockRecorder is the mock recorder for MockCompleter.
type MockCompleterMockRecorder struct {
	mock *MockCompleter
}

// NewMockCompleter creates a new mock instance.
func NewMockCompleter(ctrl *gomock.Controller) *MockCompleter {
	mock := &MockCompleter{ctrl: ctrl}
	mock.recorder = &MockCompleterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCompleter) EXPECT() *MockCompleterMockRecorder {
	return m.recorder
}

// Complete mocks base method.
func (m *MockCompleter) Complete() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Complete")
	ret0, _ := ret[0].(error)
	return ret0
}

// Complete indicates an expected call of Complete.
func (mr *MockCompleterMockRecorder) Complete() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Complete", reflect.TypeOf((*MockCompleter)(nil).Complete))
}

// MockOption is a mock of Option interface.
type MockOption struct {
	ctrl     *gomock.Controller
	recorder *MockOptionMockRecorder
}

// MockOptionMockRecorder is the mock recorder for MockOption.
type MockOptionMockRecorder struct {
	mock *MockOption
}

// NewMockOption creates a new mock instance.
func NewMockOption(ctrl *gomock.Controller) *MockOption {
	mock := &MockOption{ctrl: ctrl}
	mock.recorder = &MockOptionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOption) EXPECT() *MockOptionMockRecorder {
	return m.recorder
}

// AddFlags mocks base method.
func (m *MockOption) AddFlags(arg0 *pflag.FlagSet) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddFlags", arg0)
}

// AddFlags indicates an expected call of AddFlags.
func (mr *MockOptionMockRecorder) AddFlags(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddFlags", reflect.TypeOf((*MockOption)(nil).AddFlags), arg0)
}

// Complete mocks base method.
func (m *MockOption) Complete() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Complete")
	ret0, _ := ret[0].(error)
	return ret0
}

// Complete indicates an expected call of Complete.
func (mr *MockOptionMockRecorder) Complete() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Complete", reflect.TypeOf((*MockOption)(nil).Complete))
}

// MockFlagger is a mock of Flagger interface.
type MockFlagger struct {
	ctrl     *gomock.Controller
	recorder *MockFlaggerMockRecorder
}

// MockFlaggerMockRecorder is the mock recorder for MockFlagger.
type MockFlaggerMockRecorder struct {
	mock *MockFlagger
}

// NewMockFlagger creates a new mock instance.
func NewMockFlagger(ctrl *gomock.Controller) *MockFlagger {
	mock := &MockFlagger{ctrl: ctrl}
	mock.recorder = &MockFlaggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFlagger) EXPECT() *MockFlaggerMockRecorder {
	return m.recorder
}

// AddFlags mocks base method.
func (m *MockFlagger) AddFlags(arg0 *pflag.FlagSet) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddFlags", arg0)
}

// AddFlags indicates an expected call of AddFlags.
func (mr *MockFlaggerMockRecorder) AddFlags(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddFlags", reflect.TypeOf((*MockFlagger)(nil).AddFlags), arg0)
}
