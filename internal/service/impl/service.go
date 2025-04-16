package impl

import (
	"context"
	"errors"
	"time"

	"github.com/avito/pvz/internal/config"
	"github.com/avito/pvz/internal/models"
	"github.com/avito/pvz/internal/repository"
	"github.com/avito/pvz/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ServiceImpl struct {
	repo   repository.Repository
	config *config.Config
}

func NewService(repo repository.Repository, config *config.Config) service.Service {
	return &ServiceImpl{
		repo:   repo,
		config: config,
	}
}

func (s *ServiceImpl) RegisterUser(ctx context.Context, email, password string, role models.UserRole) (*models.User, error) {
	// Проверяем, существует ли пользователь с таким email
	existingUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создаем нового пользователя
	user := &models.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hashedPassword),
		Role:      role,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *ServiceImpl) LoginUser(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	return s.generateToken(user.ID, user.Role)
}

func (s *ServiceImpl) DummyLogin(ctx context.Context, role models.UserRole) (string, error) {
	// Создаем временного пользователя для тестирования
	user := &models.User{
		ID:        uuid.New(),
		Role:      role,
		CreatedAt: time.Now(),
	}

	return s.generateToken(user.ID, role)
}

func (s *ServiceImpl) CreatePVZ(ctx context.Context, city string) (*models.PVZ, error) {
	// Проверяем, что город находится в списке разрешенных
	allowedCities := map[string]bool{
		"Москва":          true,
		"Санкт-Петербург": true,
		"Казань":          true,
	}

	if !allowedCities[city] {
		return nil, errors.New("city not allowed")
	}

	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             city,
	}

	if err := s.repo.CreatePVZ(ctx, pvz); err != nil {
		return nil, err
	}

	return pvz, nil
}

func (s *ServiceImpl) GetPVZsWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]*repository.PVZWithReceptions, error) {
	return s.repo.GetPVZsWithReceptions(ctx, startDate, endDate, page, limit)
}

func (s *ServiceImpl) CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	// Проверяем существование ПВЗ
	pvz, err := s.repo.GetPVZByID(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if pvz == nil {
		return nil, errors.New("pvz not found")
	}

	// Проверяем, есть ли открытая приемка
	lastReception, err := s.repo.GetLastOpenReception(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if lastReception != nil {
		return nil, errors.New("there is an open reception")
	}

	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.StatusInProgress,
	}

	if err := s.repo.CreateReception(ctx, reception); err != nil {
		return nil, err
	}

	return reception, nil
}

func (s *ServiceImpl) CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	// Получаем последнюю открытую приемку
	reception, err := s.repo.GetLastOpenReception(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if reception == nil {
		return nil, errors.New("no open reception found")
	}

	// Проверяем, что приемка не закрыта
	if reception.Status == models.StatusClose {
		return nil, errors.New("reception is already closed")
	}

	if err := s.repo.CloseReception(ctx, reception.ID); err != nil {
		return nil, err
	}

	reception.Status = models.StatusClose
	return reception, nil
}

func (s *ServiceImpl) AddProducts(ctx context.Context, pvzID uuid.UUID, productTypes []models.ProductType) ([]*models.Product, error) {
	// Получаем последнюю открытую приемку
	reception, err := s.repo.GetLastOpenReception(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if reception == nil {
		return nil, errors.New("no open reception found")
	}

	products := make([]*models.Product, len(productTypes))
	for i, productType := range productTypes {
		products[i] = &models.Product{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        productType,
			ReceptionID: reception.ID,
		}
	}

	if err := s.repo.CreateProducts(ctx, products); err != nil {
		return nil, err
	}

	return products, nil
}

func (s *ServiceImpl) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	// Получаем последнюю открытую приемку
	reception, err := s.repo.GetLastOpenReception(ctx, pvzID)
	if err != nil {
		return err
	}
	if reception == nil {
		return errors.New("no open reception found")
	}

	// Проверяем, что приемка не закрыта
	if reception.Status == models.StatusClose {
		return errors.New("cannot delete products from closed reception")
	}

	// Получаем последний добавленный товар
	product, err := s.repo.GetLastProduct(ctx, reception.ID)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("no products found")
	}

	return s.repo.DeleteProduct(ctx, product.ID)
}

func (s *ServiceImpl) AddProduct(ctx context.Context, pvzID uuid.UUID, productType models.ProductType) (*models.Product, error) {
	// Получаем последнюю открытую приемку
	reception, err := s.repo.GetLastOpenReception(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if reception == nil {
		return nil, errors.New("no open reception found")
	}

	product := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        productType,
		ReceptionID: reception.ID,
	}

	if err := s.repo.CreateProduct(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ServiceImpl) generateToken(userID uuid.UUID, role models.UserRole) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(s.config.JWTConfig.Expiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTConfig.Secret))
}
