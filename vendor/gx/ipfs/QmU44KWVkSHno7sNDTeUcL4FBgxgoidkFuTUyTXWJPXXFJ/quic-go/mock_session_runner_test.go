// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/lucas-clemente/quic-go (interfaces: SessionRunner)

// Package quic is a generated GoMock package.
package quic

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	protocol "gx/ipfs/QmU44KWVkSHno7sNDTeUcL4FBgxgoidkFuTUyTXWJPXXFJ/quic-go/internal/protocol"
)

// MockSessionRunner is a mock of SessionRunner interface
type MockSessionRunner struct {
	ctrl     *gomock.Controller
	recorder *MockSessionRunnerMockRecorder
}

// MockSessionRunnerMockRecorder is the mock recorder for MockSessionRunner
type MockSessionRunnerMockRecorder struct {
	mock *MockSessionRunner
}

// NewMockSessionRunner creates a new mock instance
func NewMockSessionRunner(ctrl *gomock.Controller) *MockSessionRunner {
	mock := &MockSessionRunner{ctrl: ctrl}
	mock.recorder = &MockSessionRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSessionRunner) EXPECT() *MockSessionRunnerMockRecorder {
	return m.recorder
}

// onHandshakeComplete mocks base method
func (m *MockSessionRunner) onHandshakeComplete(arg0 Session) {
	m.ctrl.Call(m, "onHandshakeComplete", arg0)
}

// onHandshakeComplete indicates an expected call of onHandshakeComplete
func (mr *MockSessionRunnerMockRecorder) onHandshakeComplete(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "onHandshakeComplete", reflect.TypeOf((*MockSessionRunner)(nil).onHandshakeComplete), arg0)
}

// removeConnectionID mocks base method
func (m *MockSessionRunner) removeConnectionID(arg0 protocol.ConnectionID) {
	m.ctrl.Call(m, "removeConnectionID", arg0)
}

// removeConnectionID indicates an expected call of removeConnectionID
func (mr *MockSessionRunnerMockRecorder) removeConnectionID(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "removeConnectionID", reflect.TypeOf((*MockSessionRunner)(nil).removeConnectionID), arg0)
}
