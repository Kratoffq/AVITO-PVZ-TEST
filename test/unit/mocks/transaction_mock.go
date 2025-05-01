package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockTransactionManager реализует интерфейс transaction.Manager
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}
