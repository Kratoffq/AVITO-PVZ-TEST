package mocks

import (
	"context"
	"time"

	"github.com/avito/pvz/internal/domain/product"
	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/reception"
	"github.com/avito/pvz/internal/domain/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockPVZRepository реализует интерфейс pvz.Repository
type MockPVZRepository struct {
	mock.Mock
}

func (m *MockPVZRepository) Create(ctx context.Context, p *pvz.PVZ) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPVZRepository) GetByID(ctx context.Context, id uuid.UUID) (*pvz.PVZ, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetByCity(ctx context.Context, city string) (*pvz.PVZ, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) Update(ctx context.Context, p *pvz.PVZ) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPVZRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPVZRepository) List(ctx context.Context, offset, limit int) ([]*pvz.PVZ, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetAll(ctx context.Context) ([]*pvz.PVZ, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*pvz.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*pvz.PVZWithReceptions, error) {
	args := m.Called(ctx, startDate, endDate, page, limit)
	return args.Get(0).([]*pvz.PVZWithReceptions), args.Error(1)
}

// MockReceptionRepository реализует интерфейс reception.Repository
type MockReceptionRepository struct {
	mock.Mock
}

func (m *MockReceptionRepository) Create(ctx context.Context, r *reception.Reception) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockReceptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *MockReceptionRepository) Update(ctx context.Context, r *reception.Reception) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockReceptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReceptionRepository) List(ctx context.Context, offset, limit int) ([]*reception.Reception, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*reception.Reception), args.Error(1)
}

func (m *MockReceptionRepository) GetOpenByPVZID(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

func (m *MockReceptionRepository) GetProducts(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	args := m.Called(ctx, receptionID)
	return args.Get(0).([]*product.Product), args.Error(1)
}

func (m *MockReceptionRepository) GetLastOpen(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception.Reception), args.Error(1)
}

// MockProductRepository реализует интерфейс product.Repository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, p *product.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *MockProductRepository) GetByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]*product.Product, error) {
	args := m.Called(ctx, receptionID)
	return args.Get(0).([]*product.Product), args.Error(1)
}

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) CreateBatch(ctx context.Context, products []*product.Product) error {
	args := m.Called(ctx, products)
	return args.Error(0)
}

// MockUserRepository реализует интерфейс user.Repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*user.User, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*user.User), args.Error(1)
}
