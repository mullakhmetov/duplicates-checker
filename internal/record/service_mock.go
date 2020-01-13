package record

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockedService is a mocked Service implementation
type MockedService struct {
	mock.Mock
}

// AddRecord mocked
func (m *MockedService) AddRecord(ctx context.Context, record *Record) error {
	args := m.Called(ctx, record)
	return args.Error(1)
}

// BulkAddRecords mocked
func (m *MockedService) BulkAddRecords(ctx context.Context, records []*Record) error {
	args := m.Called(ctx, records)
	return args.Error(1)
}

// IsDuple mocked
func (m *MockedService) IsDuple(ctx context.Context, u1, u2 UserID) (bool, error) {
	args := m.Called(ctx, u1, u2)
	return args.Bool(0), args.Error(1)
}

// Clear mocked
func (m *MockedService) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(1)
}
