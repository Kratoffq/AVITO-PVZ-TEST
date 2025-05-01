package repository

import (
	"context"
	"database/sql"
	"sort"

	"github.com/avito/pvz/internal/pvz"
)

// MockRepo представляет мок репозитория для тестирования
type MockRepo struct {
	pvzs        map[int64]*pvz.PVZ
	lastID      int64
	CreateError error
	UpdateError error
	DeleteError error
	ListError   error
}

// NewMockRepo создает новый экземпляр MockRepo
func NewMockRepo() *MockRepo {
	return &MockRepo{
		pvzs:   make(map[int64]*pvz.PVZ),
		lastID: 0,
	}
}

// Create реализует метод Repository.Create для мока
func (r *MockRepo) Create(ctx context.Context, p *pvz.PVZ) error {
	if r.CreateError != nil {
		return r.CreateError
	}
	r.lastID++
	p.ID = r.lastID
	r.pvzs[p.ID] = p
	return nil
}

// GetByID реализует метод Repository.GetByID для мока
func (r *MockRepo) GetByID(ctx context.Context, id int64) (*pvz.PVZ, error) {
	if p, ok := r.pvzs[id]; ok {
		return p, nil
	}
	return nil, sql.ErrNoRows
}

// GetLast реализует метод Repository.GetLast для мока
func (r *MockRepo) GetLast(ctx context.Context) (*pvz.PVZ, error) {
	if r.lastID == 0 {
		return nil, sql.ErrNoRows
	}
	return r.pvzs[r.lastID], nil
}

// List реализует метод Repository.List для мока
func (r *MockRepo) List(ctx context.Context, offset, limit int) ([]*pvz.PVZ, error) {
	if r.ListError != nil {
		return nil, r.ListError
	}

	// Создаем слайс для сортировки по ID
	ids := make([]int64, 0, len(r.pvzs))
	for id := range r.pvzs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	// Применяем пагинацию
	start := offset
	if start >= len(ids) {
		return []*pvz.PVZ{}, nil
	}

	end := start + limit
	if end > len(ids) {
		end = len(ids)
	}

	result := make([]*pvz.PVZ, 0, end-start)
	for _, id := range ids[start:end] {
		result = append(result, r.pvzs[id])
	}
	return result, nil
}

// Update реализует метод Repository.Update для мока
func (r *MockRepo) Update(ctx context.Context, p *pvz.PVZ) error {
	if r.UpdateError != nil {
		return r.UpdateError
	}
	if _, ok := r.pvzs[p.ID]; !ok {
		return sql.ErrNoRows
	}
	r.pvzs[p.ID] = p
	return nil
}

// Delete реализует метод Repository.Delete для мока
func (r *MockRepo) Delete(ctx context.Context, id int64) error {
	if r.DeleteError != nil {
		return r.DeleteError
	}
	if _, ok := r.pvzs[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.pvzs, id)
	return nil
}

// GetWithReceptions реализует метод Repository.GetWithReceptions для мока
func (r *MockRepo) GetWithReceptions(ctx context.Context, id int64) (*pvz.PVZ, []*pvz.Reception, error) {
	p, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return p, make([]*pvz.Reception, 0), nil
}

// MockTxManager реализует интерфейс TransactionManager для тестирования
type MockTxManager struct {
	BeginTxError    error
	CommitTxError   error
	RollbackTxError error
}

// NewMockTxManager создает новый экземпляр MockTxManager
func NewMockTxManager() *MockTxManager {
	return &MockTxManager{}
}

// BeginTx начинает новую транзакцию
func (m *MockTxManager) BeginTx(ctx context.Context) (context.Context, error) {
	if m.BeginTxError != nil {
		return nil, m.BeginTxError
	}
	return ctx, nil
}

// CommitTx фиксирует транзакцию
func (m *MockTxManager) CommitTx(ctx context.Context) error {
	return m.CommitTxError
}

// RollbackTx откатывает транзакцию
func (m *MockTxManager) RollbackTx(ctx context.Context) error {
	return m.RollbackTxError
}
