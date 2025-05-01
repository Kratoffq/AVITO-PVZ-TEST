package pvz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avito/pvz/internal/domain/audit"
	"github.com/avito/pvz/internal/domain/pvz"
	"github.com/avito/pvz/internal/domain/transaction"
	"github.com/avito/pvz/internal/domain/user"
	"github.com/avito/pvz/internal/metrics"
	"github.com/google/uuid"
)

var (
	ErrInvalidCity      = errors.New("invalid city name")
	ErrAccessDenied     = errors.New("access denied")
	ErrPVZNotFound      = errors.New("pvz not found")
	ErrPVZAlreadyExists = errors.New("pvz already exists")
	ErrInvalidPVZData   = errors.New("неверные данные пвз")
	ErrDuplicatePVZ     = errors.New("пвз с таким городом уже существует")
	ErrUnauthorized     = errors.New("недостаточно прав для выполнения операции")
)

// Service определяет бизнес-логику для работы с ПВЗ
type Service struct {
	pvzRepo   pvz.Repository
	userRepo  user.Repository
	txManager transaction.Manager
	auditLog  audit.AuditLog
	userModel *user.User
}

// New создает новый экземпляр Service
func New(pvzRepo pvz.Repository, userRepo user.Repository, txManager transaction.Manager, auditLog audit.AuditLog, userModel *user.User) *Service {
	return &Service{
		pvzRepo:   pvzRepo,
		userRepo:  userRepo,
		txManager: txManager,
		auditLog:  auditLog,
		userModel: userModel,
	}
}

// Create создает новый ПВЗ
func (s *Service) Create(ctx context.Context, city string, userID uuid.UUID) (*pvz.PVZ, error) {
	// Валидация входных данных
	if err := validateCity(city); err != nil {
		return nil, ErrInvalidCity
	}

	if userID == uuid.Nil {
		return nil, ErrAccessDenied
	}

	// Проверяем права пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if !user.CanCreatePVZ() {
		return nil, ErrAccessDenied
	}

	newPVZ := &pvz.PVZ{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		City:      city,
	}

	// Выполняем операцию в транзакции для обеспечения атомарности
	err = s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Проверяем существование ПВЗ с таким же городом
		existingPVZ, err := s.pvzRepo.GetByCity(ctx, city)
		if err == nil && existingPVZ != nil {
			return ErrPVZAlreadyExists
		}
		if err != nil && !errors.Is(err, pvz.ErrNotFound) {
			return fmt.Errorf("failed to check existing pvz: %w", err)
		}

		// Создаем новый ПВЗ
		if err := s.pvzRepo.Create(ctx, newPVZ); err != nil {
			return fmt.Errorf("failed to create pvz: %w", err)
		}

		// Создаем запись в журнале аудита
		if err := s.auditLog.LogPVZCreation(ctx, newPVZ.ID, userID); err != nil {
			return fmt.Errorf("failed to log pvz creation: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newPVZ, nil
}

// GetByID получает ПВЗ по ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*pvz.PVZ, error) {
	start := time.Now()
	pvz, err := s.pvzRepo.GetByID(ctx, id)

	// Обновляем метрики
	metrics.TransactionDuration.WithLabelValues("get_pvz").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.TransactionErrors.WithLabelValues("get_pvz").Inc()
	}

	if err != nil {
		return nil, ErrPVZNotFound
	}
	return pvz, nil
}

// GetWithReceptions получает список ПВЗ с приемками за период
func (s *Service) GetWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*pvz.PVZWithReceptions, error) {
	if page <= 0 || limit <= 0 {
		return nil, errors.New("invalid pagination parameters")
	}
	return s.pvzRepo.GetWithReceptions(ctx, startDate, endDate, page, limit)
}

// GetAll возвращает список всех ПВЗ
func (s *Service) GetAll(ctx context.Context) ([]*pvz.PVZ, error) {
	start := time.Now()
	pvzs, err := s.pvzRepo.GetAll(ctx)

	// Обновляем метрики
	metrics.TransactionDuration.WithLabelValues("get_all_pvz").Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.TransactionErrors.WithLabelValues("get_all_pvz").Inc()
	}

	return pvzs, err
}

// Update обновляет данные ПВЗ
func (s *Service) Update(ctx context.Context, pvz *pvz.PVZ, moderatorID uuid.UUID) error {
	// Проверяем права модератора
	moderator, err := s.userRepo.GetByID(ctx, moderatorID)
	if err != nil {
		return err
	}
	if moderator.Role != user.RoleAdmin {
		return ErrAccessDenied
	}

	// Валидация города
	if err := validateCity(pvz.City); err != nil {
		return err
	}

	// Проверяем ID
	if pvz.ID == uuid.Nil {
		return ErrPVZNotFound
	}

	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Проверяем существование ПВЗ
		if _, err := s.pvzRepo.GetByID(ctx, pvz.ID); err != nil {
			return ErrPVZNotFound
		}

		return s.pvzRepo.Update(ctx, pvz)
	})
}

// Delete удаляет ПВЗ
func (s *Service) Delete(ctx context.Context, id uuid.UUID, moderatorID uuid.UUID) error {
	// Проверяем права модератора
	moderator, err := s.userRepo.GetByID(ctx, moderatorID)
	if err != nil {
		return err
	}
	if moderator.Role != user.RoleAdmin {
		return ErrAccessDenied
	}

	// Проверяем ID
	if id == uuid.Nil {
		return ErrPVZNotFound
	}

	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Проверяем существование ПВЗ
		if _, err := s.pvzRepo.GetByID(ctx, id); err != nil {
			return ErrPVZNotFound
		}

		return s.pvzRepo.Delete(ctx, id)
	})
}

// List возвращает список ПВЗ
func (s *Service) List(ctx context.Context, offset, limit int) ([]*pvz.PVZ, error) {
	return s.pvzRepo.List(ctx, offset, limit)
}

// validateCity проверяет корректность названия города
func validateCity(city string) error {
	if city == "" {
		return ErrInvalidCity
	}
	if len(city) < 2 {
		return ErrInvalidCity
	}
	if len(city) > 100 {
		return ErrInvalidCity
	}
	return nil
}
