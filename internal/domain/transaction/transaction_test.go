package transaction

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockManager реализует интерфейс Manager для тестирования
type MockManager struct {
	OnWithinTransaction func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *MockManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.OnWithinTransaction != nil {
		return m.OnWithinTransaction(ctx, fn)
	}
	return nil
}

func TestTransactionCtxKey(t *testing.T) {
	// Проверяем, что ключ контекста можно использовать
	key := TransactionCtxKey{}
	ctx := context.WithValue(context.Background(), key, "test")

	// Проверяем, что значение можно получить из контекста
	value := ctx.Value(key)
	assert.Equal(t, "test", value)
}

func TestManager_Interface(t *testing.T) {
	// Проверяем, что MockManager реализует интерфейс Manager
	var _ Manager = &MockManager{}
}

func TestMockManager_WithinTransaction(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*MockManager, context.Context, func(ctx context.Context) error)
		wantErr bool
	}{
		{
			name: "успешное выполнение транзакции",
			setup: func() (*MockManager, context.Context, func(ctx context.Context) error) {
				manager := &MockManager{
					OnWithinTransaction: func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					},
				}
				return manager, context.Background(), func(ctx context.Context) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "ошибка в транзакции",
			setup: func() (*MockManager, context.Context, func(ctx context.Context) error) {
				expectedErr := assert.AnError
				manager := &MockManager{
					OnWithinTransaction: func(ctx context.Context, fn func(ctx context.Context) error) error {
						return expectedErr
					},
				}
				return manager, context.Background(), func(ctx context.Context) error {
					return nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, ctx, fn := tt.setup()
			err := manager.WithinTransaction(ctx, fn)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
