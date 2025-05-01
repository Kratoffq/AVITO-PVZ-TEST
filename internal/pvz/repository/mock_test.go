package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/avito/pvz/internal/pvz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockRepo(t *testing.T) {
	t.Run("create with error", func(t *testing.T) {
		repo := NewMockRepo()
		repo.CreateError = errors.New("create error")
		err := repo.Create(context.Background(), &pvz.PVZ{})
		assert.Error(t, err)
		assert.Equal(t, repo.CreateError, err)
	})

	t.Run("update with error", func(t *testing.T) {
		repo := NewMockRepo()
		repo.UpdateError = errors.New("update error")
		err := repo.Update(context.Background(), &pvz.PVZ{})
		assert.Error(t, err)
		assert.Equal(t, repo.UpdateError, err)
	})

	t.Run("delete with error", func(t *testing.T) {
		repo := NewMockRepo()
		repo.DeleteError = errors.New("delete error")
		err := repo.Delete(context.Background(), 1)
		assert.Error(t, err)
		assert.Equal(t, repo.DeleteError, err)
	})
}

func TestMockTxManager(t *testing.T) {
	t.Run("begin tx with error", func(t *testing.T) {
		txManager := NewMockTxManager()
		txManager.BeginTxError = errors.New("begin error")
		ctx := context.Background()
		_, err := txManager.BeginTx(ctx)
		assert.Error(t, err)
		assert.Equal(t, txManager.BeginTxError, err)
	})

	t.Run("begin tx success", func(t *testing.T) {
		txManager := NewMockTxManager()
		ctx := context.Background()
		newCtx, err := txManager.BeginTx(ctx)
		require.NoError(t, err)
		assert.Equal(t, ctx, newCtx)
	})

	t.Run("commit tx with error", func(t *testing.T) {
		txManager := NewMockTxManager()
		txManager.CommitTxError = errors.New("commit error")
		err := txManager.CommitTx(context.Background())
		assert.Error(t, err)
		assert.Equal(t, txManager.CommitTxError, err)
	})

	t.Run("commit tx success", func(t *testing.T) {
		txManager := NewMockTxManager()
		err := txManager.CommitTx(context.Background())
		assert.NoError(t, err)
	})

	t.Run("rollback tx with error", func(t *testing.T) {
		txManager := NewMockTxManager()
		txManager.RollbackTxError = errors.New("rollback error")
		err := txManager.RollbackTx(context.Background())
		assert.Error(t, err)
		assert.Equal(t, txManager.RollbackTxError, err)
	})

	t.Run("rollback tx success", func(t *testing.T) {
		txManager := NewMockTxManager()
		err := txManager.RollbackTx(context.Background())
		assert.NoError(t, err)
	})
}
