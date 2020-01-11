package record

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockedService is a mocked Service implementation
type MockedService struct {
	mock.Mock
}

// Create mocked
func (m *MockedService) Create(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(1)
}

// IsDouble mocked
func (m *MockedService) IsDouble(ctx context.Context, u1, u2 UserID) (bool, error) {
	args := m.Called(ctx, u1, u2)
	return args.Bool(0), args.Error(1)
}

// Clear mocked
func (m *MockedService) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(1)
}

// Close mocked
func (m *MockedService) Close() error {
	args := m.Called()
	return args.Error(1)
}
