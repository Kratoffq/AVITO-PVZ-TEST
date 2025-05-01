package mocks

import (
	"context"
	"testing"
	"time"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/avito/pvz/internal/domain/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMockPVZRepository(t *testing.T) {
	mockRepo := new(MockPVZRepository)
	ctx := context.Background()
	pvzID := uuid.New()
	testPVZ := &pvz.PVZ{
		ID:   pvzID,
		City: "Москва",
	}

	// Тест Create
	t.Run("Create", func(t *testing.T) {
		mockRepo.On("Create", ctx, testPVZ).Return(nil)
		err := mockRepo.Create(ctx, testPVZ)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetByID
	t.Run("GetByID", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, pvzID).Return(testPVZ, nil)
		result, err := mockRepo.GetByID(ctx, pvzID)
		assert.NoError(t, err)
		assert.Equal(t, testPVZ, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetByCity
	t.Run("GetByCity", func(t *testing.T) {
		mockRepo.On("GetByCity", ctx, "Москва").Return(testPVZ, nil)
		result, err := mockRepo.GetByCity(ctx, "Москва")
		assert.NoError(t, err)
		assert.Equal(t, testPVZ, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест Update
	t.Run("Update", func(t *testing.T) {
		mockRepo.On("Update", ctx, testPVZ).Return(nil)
		err := mockRepo.Update(ctx, testPVZ)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Тест Delete
	t.Run("Delete", func(t *testing.T) {
		mockRepo.On("Delete", ctx, pvzID).Return(nil)
		err := mockRepo.Delete(ctx, pvzID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Тест List
	t.Run("List", func(t *testing.T) {
		expectedPVZs := []*pvz.PVZ{testPVZ}
		mockRepo.On("List", ctx, 0, 10).Return(expectedPVZs, nil)
		result, err := mockRepo.List(ctx, 0, 10)
		assert.NoError(t, err)
		assert.Equal(t, expectedPVZs, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetAll
	t.Run("GetAll", func(t *testing.T) {
		expectedPVZs := []*pvz.PVZ{testPVZ}
		mockRepo.On("GetAll", ctx).Return(expectedPVZs, nil)
		result, err := mockRepo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedPVZs, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetWithReceptions
	t.Run("GetWithReceptions", func(t *testing.T) {
		startDate := time.Now()
		endDate := startDate.Add(24 * time.Hour)
		expectedPVZs := []*pvz.PVZWithReceptions{
			{
				PVZ: testPVZ,
				Receptions: []*pvz.ReceptionWithProducts{
					{
						Reception: &reception.Reception{
							ID:    uuid.New(),
							PVZID: pvzID,
						},
					},
				},
			},
		}
		mockRepo.On("GetWithReceptions", ctx, startDate, endDate, 1, 10).Return(expectedPVZs, nil)
		result, err := mockRepo.GetWithReceptions(ctx, startDate, endDate, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, expectedPVZs, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestMockReceptionRepository(t *testing.T) {
	mockRepo := new(MockReceptionRepository)
	ctx := context.Background()
	receptionID := uuid.New()
	pvzID := uuid.New()
	testReception := &reception.Reception{
		ID:     receptionID,
		PVZID:  pvzID,
		Status: reception.StatusInProgress,
	}

	// Тест Create
	t.Run("Create", func(t *testing.T) {
		mockRepo.On("Create", ctx, testReception).Return(nil)
		err := mockRepo.Create(ctx, testReception)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetByID
	t.Run("GetByID", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, receptionID).Return(testReception, nil)
		result, err := mockRepo.GetByID(ctx, receptionID)
		assert.NoError(t, err)
		assert.Equal(t, testReception, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест Update
	t.Run("Update", func(t *testing.T) {
		mockRepo.On("Update", ctx, testReception).Return(nil)
		err := mockRepo.Update(ctx, testReception)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Тест Delete
	t.Run("Delete", func(t *testing.T) {
		mockRepo.On("Delete", ctx, receptionID).Return(nil)
		err := mockRepo.Delete(ctx, receptionID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Тест List
	t.Run("List", func(t *testing.T) {
		expectedReceptions := []*reception.Reception{testReception}
		mockRepo.On("List", ctx, 0, 10).Return(expectedReceptions, nil)
		result, err := mockRepo.List(ctx, 0, 10)
		assert.NoError(t, err)
		assert.Equal(t, expectedReceptions, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetOpenByPVZID
	t.Run("GetOpenByPVZID", func(t *testing.T) {
		mockRepo.On("GetOpenByPVZID", ctx, pvzID).Return(testReception, nil)
		result, err := mockRepo.GetOpenByPVZID(ctx, pvzID)
		assert.NoError(t, err)
		assert.Equal(t, testReception, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetProducts
	t.Run("GetProducts", func(t *testing.T) {
		expectedProducts := []*product.Product{
			{
				ID:          uuid.New(),
				ReceptionID: receptionID,
				Type:        product.TypeElectronics,
			},
		}
		mockRepo.On("GetProducts", ctx, receptionID).Return(expectedProducts, nil)
		result, err := mockRepo.GetProducts(ctx, receptionID)
		assert.NoError(t, err)
		assert.Equal(t, expectedProducts, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetLastOpen
	t.Run("GetLastOpen", func(t *testing.T) {
		mockRepo.On("GetLastOpen", ctx, pvzID).Return(testReception, nil)
		result, err := mockRepo.GetLastOpen(ctx, pvzID)
		assert.NoError(t, err)
		assert.Equal(t, testReception, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestMockProductRepository(t *testing.T) {
	mockRepo := new(MockProductRepository)
	ctx := context.Background()
	productID := uuid.New()
	receptionID := uuid.New()
	testProduct := &product.Product{
		ID:          productID,
		ReceptionID: receptionID,
		Type:        product.TypeElectronics,
	}

	// Тест Create
	t.Run("Create", func(t *testing.T) {
		mockRepo.On("Create", ctx, testProduct).Return(nil)
		err := mockRepo.Create(ctx, testProduct)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetByID
	t.Run("GetByID", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, productID).Return(testProduct, nil)
		result, err := mockRepo.GetByID(ctx, productID)
		assert.NoError(t, err)
		assert.Equal(t, testProduct, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetByReceptionID
	t.Run("GetByReceptionID", func(t *testing.T) {
		expectedProducts := []*product.Product{testProduct}
		mockRepo.On("GetByReceptionID", ctx, receptionID).Return(expectedProducts, nil)
		result, err := mockRepo.GetByReceptionID(ctx, receptionID)
		assert.NoError(t, err)
		assert.Equal(t, expectedProducts, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест Delete
	t.Run("Delete", func(t *testing.T) {
		mockRepo.On("Delete", ctx, productID).Return(nil)
		err := mockRepo.Delete(ctx, productID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Тест CreateBatch
	t.Run("CreateBatch", func(t *testing.T) {
		products := []*product.Product{testProduct}
		mockRepo.On("CreateBatch", ctx, products).Return(nil)
		err := mockRepo.CreateBatch(ctx, products)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestMockUserRepository(t *testing.T) {
	mockRepo := new(MockUserRepository)
	ctx := context.Background()
	userID := uuid.New()
	testUser := &user.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: "password",
		Role:     user.RoleUser,
	}

	// Тест Create
	t.Run("Create", func(t *testing.T) {
		mockRepo.On("Create", ctx, testUser).Return(nil)
		err := mockRepo.Create(ctx, testUser)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetByID
	t.Run("GetByID", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, userID).Return(testUser, nil)
		result, err := mockRepo.GetByID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, testUser, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест GetByEmail
	t.Run("GetByEmail", func(t *testing.T) {
		mockRepo.On("GetByEmail", ctx, "test@example.com").Return(testUser, nil)
		result, err := mockRepo.GetByEmail(ctx, "test@example.com")
		assert.NoError(t, err)
		assert.Equal(t, testUser, result)
		mockRepo.AssertExpectations(t)
	})

	// Тест Update
	t.Run("Update", func(t *testing.T) {
		mockRepo.On("Update", ctx, testUser).Return(nil)
		err := mockRepo.Update(ctx, testUser)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Тест Delete
	t.Run("Delete", func(t *testing.T) {
		mockRepo.On("Delete", ctx, userID).Return(nil)
		err := mockRepo.Delete(ctx, userID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Тест List
	t.Run("List", func(t *testing.T) {
		expectedUsers := []*user.User{testUser}
		mockRepo.On("List", ctx, 0, 10).Return(expectedUsers, nil)
		result, err := mockRepo.List(ctx, 0, 10)
		assert.NoError(t, err)
		assert.Equal(t, expectedUsers, result)
		mockRepo.AssertExpectations(t)
	})
}
