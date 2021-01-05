// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/darkowlzz/operator-toolkit/controller/composite/v1 (interfaces: Controller)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	reflect "reflect"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// MockController is a mock of Controller interface
type MockController struct {
	ctrl     *gomock.Controller
	recorder *MockControllerMockRecorder
}

// MockControllerMockRecorder is the mock recorder for MockController
type MockControllerMockRecorder struct {
	mock *MockController
}

// NewMockController creates a new mock instance
func NewMockController(ctrl *gomock.Controller) *MockController {
	mock := &MockController{ctrl: ctrl}
	mock.recorder = &MockControllerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockController) EXPECT() *MockControllerMockRecorder {
	return m.recorder
}

// Cleanup mocks base method
func (m *MockController) Cleanup(arg0 context.Context, arg1 client.Object) (reconcile.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Cleanup", arg0, arg1)
	ret0, _ := ret[0].(reconcile.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Cleanup indicates an expected call of Cleanup
func (mr *MockControllerMockRecorder) Cleanup(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Cleanup", reflect.TypeOf((*MockController)(nil).Cleanup), arg0, arg1)
}

// Default mocks base method
func (m *MockController) Default(arg0 context.Context, arg1 client.Object) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Default", arg0, arg1)
}

// Default indicates an expected call of Default
func (mr *MockControllerMockRecorder) Default(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Default", reflect.TypeOf((*MockController)(nil).Default), arg0, arg1)
}

// Initialize mocks base method
func (m *MockController) Initialize(arg0 context.Context, arg1 client.Object, arg2 v1.Condition) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Initialize", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Initialize indicates an expected call of Initialize
func (mr *MockControllerMockRecorder) Initialize(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Initialize", reflect.TypeOf((*MockController)(nil).Initialize), arg0, arg1, arg2)
}

// Operate mocks base method
func (m *MockController) Operate(arg0 context.Context, arg1 client.Object) (reconcile.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Operate", arg0, arg1)
	ret0, _ := ret[0].(reconcile.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Operate indicates an expected call of Operate
func (mr *MockControllerMockRecorder) Operate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Operate", reflect.TypeOf((*MockController)(nil).Operate), arg0, arg1)
}

// UpdateStatus mocks base method
func (m *MockController) UpdateStatus(arg0 context.Context, arg1 client.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateStatus", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateStatus indicates an expected call of UpdateStatus
func (mr *MockControllerMockRecorder) UpdateStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockController)(nil).UpdateStatus), arg0, arg1)
}

// Validate mocks base method
func (m *MockController) Validate(arg0 context.Context, arg1 client.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validate", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Validate indicates an expected call of Validate
func (mr *MockControllerMockRecorder) Validate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockController)(nil).Validate), arg0, arg1)
}
