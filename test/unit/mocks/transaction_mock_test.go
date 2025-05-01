package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockTransactionManager(t *testing.T) {
	t.Run("Successful transaction", func(t *testing.T) {
		mockManager := new(MockTransactionManager)
		ctx := context.Background()

		mockManager.On("WithinTransaction", mock.Anything, mock.Anything).Return(nil)

		err := mockManager.WithinTransaction(ctx, func(ctx context.Context) error {
			return nil
		})

		assert.NoError(t, err)
		mockManager.AssertExpectations(t)
	})

	t.Run("Transaction error", func(t *testing.T) {
		mockManager := new(MockTransactionManager)
		ctx := context.Background()
		expectedErr := errors.New("transaction error")

		mockManager.On("WithinTransaction", mock.Anything, mock.Anything).Return(expectedErr)

		err := mockManager.WithinTransaction(ctx, func(ctx context.Context) error {
			return nil
		})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockManager.AssertExpectations(t)
	})

	t.Run("Function error", func(t *testing.T) {
		mockManager := new(MockTransactionManager)
		ctx := context.Background()
		expectedErr := errors.New("function error")

		mockManager.On("WithinTransaction", mock.Anything, mock.Anything).Return(expectedErr)

		err := mockManager.WithinTransaction(ctx, func(ctx context.Context) error {
			return expectedErr
		})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockManager.AssertExpectations(t)
	})
}
