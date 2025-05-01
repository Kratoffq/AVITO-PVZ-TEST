package service

import (
	"context"

	"github.com/avito/pvz/internal/pvz"
	"github.com/avito/pvz/internal/pvz/repository"
)

// PVZService представляет сервис для работы с PVZ
type PVZService struct {
	repo      repository.Repository
	txManager repository.TransactionManager
}

// NewPVZService создает новый экземпляр PVZService
func NewPVZService(repo repository.Repository, txManager repository.TransactionManager) *PVZService {
	return &PVZService{
		repo:      repo,
		txManager: txManager,
	}
}

// Create создает новый PVZ
func (s *PVZService) Create(ctx context.Context, pvz *pvz.PVZ) error {
	ctx, err := s.txManager.BeginTx(ctx)
	if err != nil {
		return err
	}

	err = s.repo.Create(ctx, pvz)
	if err != nil {
		if rbErr := s.txManager.RollbackTx(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return s.txManager.CommitTx(ctx)
}

// GetByID возвращает PVZ по ID
func (s *PVZService) GetByID(ctx context.Context, id int64) (*pvz.PVZ, error) {
	return s.repo.GetByID(ctx, id)
}

// GetLast возвращает последний добавленный PVZ
func (s *PVZService) GetLast(ctx context.Context) (*pvz.PVZ, error) {
	return s.repo.GetLast(ctx)
}

// Update обновляет существующий PVZ
func (s *PVZService) Update(ctx context.Context, pvz *pvz.PVZ) error {
	ctx, err := s.txManager.BeginTx(ctx)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, pvz)
	if err != nil {
		if rbErr := s.txManager.RollbackTx(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return s.txManager.CommitTx(ctx)
}

// Delete удаляет PVZ по ID
func (s *PVZService) Delete(ctx context.Context, id int64) error {
	ctx, err := s.txManager.BeginTx(ctx)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		if rbErr := s.txManager.RollbackTx(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return s.txManager.CommitTx(ctx)
}

// List возвращает список всех PVZ с поддержкой пагинации
func (s *PVZService) List(ctx context.Context, offset, limit int) ([]*pvz.PVZ, error) {
	return s.repo.List(ctx, offset, limit)
}

// GetWithReceptions возвращает PVZ вместе со связанными приемами товаров
func (s *PVZService) GetWithReceptions(ctx context.Context, id int64) (*pvz.PVZ, []*pvz.Reception, error) {
	return s.repo.GetWithReceptions(ctx, id)
}

// GetPvzList возвращает список PVZ с поддержкой пагинации и фильтрации
func (s *PVZService) GetPvzList(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*pvz.PVZ, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10 // значение по умолчанию
	}
	if limit > 100 {
		limit = 100 // максимальное значение
	}

	// Начинаем транзакцию для обеспечения атомарности
	ctx, err := s.txManager.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	// Получаем список PVZ
	pvzList, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		if rbErr := s.txManager.RollbackTx(ctx); rbErr != nil {
			return nil, rbErr
		}
		return nil, err
	}

	// Применяем фильтры, если они есть
	if len(filters) > 0 {
		filteredList := make([]*pvz.PVZ, 0)
		for _, p := range pvzList {
			if matchesFilters(p, filters) {
				filteredList = append(filteredList, p)
			}
		}
		pvzList = filteredList
	}

	// Завершаем транзакцию
	if err := s.txManager.CommitTx(ctx); err != nil {
		return nil, err
	}

	return pvzList, nil
}

// matchesFilters проверяет, соответствует ли PVZ заданным фильтрам
func matchesFilters(p *pvz.PVZ, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "name":
			if name, ok := value.(string); ok && p.Name != name {
				return false
			}
		case "address":
			if address, ok := value.(string); ok && p.Address != address {
				return false
			}
		case "status":
			if status, ok := value.(pvz.Status); ok && p.Status != status {
				return false
			}
		}
	}
	return true
}
