package storage

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockStorageIface is a mock of StorageIface interface.
type MockStorageIface struct {
	mock.Mock
}

// Init provides a mock function with given fields.
func (_m *MockStorageIface) Init(_ context.Context) error {
	_mockArgs := _m.Called()
	return _mockArgs.Error(0)
}

// Name provides a mock function with given fields.
func (_m *MockStorageIface) Name() string {
	_mockArgs := _m.Called()
	return _mockArgs.String(0)
}

// Upload provides a mock function with given fields.
func (_m *MockStorageIface) Upload(_ context.Context, localPath string) (string, error) {
	_mockArgs := _m.Called(localPath)
	return _mockArgs.String(0), _mockArgs.Error(1)
}

// List provides a mock function with given fields.
func (_m *MockStorageIface) List(_ context.Context) ([]string, error) {
	_mockArgs := _m.Called()
	if _mockArgs.Get(0) == nil {
		return nil, _mockArgs.Error(1)
	}
	return _mockArgs.Get(0).([]string), _mockArgs.Error(1) //nolint:errcheck // reason: type assertion on mock, error not possible/needed
}

// Delete provides a mock function with given fields.
func (_m *MockStorageIface) Delete(_ context.Context, key string) error {
	_mockArgs := _m.Called(key)
	return _mockArgs.Error(0)
}

// TrimPrefix provides a mock function with given fields.
func (_m *MockStorageIface) TrimPrefix(keys []string) []string {
	_mockArgs := _m.Called(keys)
	return _mockArgs.Get(0).([]string) //nolint:errcheck // reason: type assertion on mock, error not possible/needed
}

// NewMockStorageIface creates a new instance of MockStorageIface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockStorageIface(t mock.TestingT) *MockStorageIface {
	mock := &MockStorageIface{}
	mock.Test(t)
	return mock
}
