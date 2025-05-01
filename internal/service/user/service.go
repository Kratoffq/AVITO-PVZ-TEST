package user

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/avito/pvz/internal/domain/transaction"
	"github.com/avito/pvz/internal/domain/user"
	"github.com/avito/pvz/pkg/auth"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidRole       = errors.New("invalid role")
)

// ServiceInterface определяет интерфейс для работы с пользователями
type ServiceInterface interface {
	Register(ctx context.Context, email, password string, role user.Role) (*user.User, error)
	Login(ctx context.Context, email, password string) (*user.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	Update(ctx context.Context, user *user.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*user.User, error)
	LoginUser(ctx context.Context, email, password string) (string, error)
}

// Service реализует ServiceInterface
type Service struct {
	userRepo  user.Repository
	txManager transaction.Manager
}

// New создает новый экземпляр Service
func New(userRepo user.Repository, txManager transaction.Manager) *Service {
	return &Service{
		userRepo:  userRepo,
		txManager: txManager,
	}
}

// Register регистрирует нового пользователя
func (s *Service) Register(ctx context.Context, email, password string, role user.Role) (*user.User, error) {
	// Валидация входных данных
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	if err := validateRole(role); err != nil {
		return nil, err
	}

	var result *user.User
	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Проверяем существование пользователя
		if _, err := s.userRepo.GetByEmail(ctx, email); err != user.ErrNotFound {
			if err == nil {
				return ErrUserAlreadyExists
			}
			return err
		}

		// Хешируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		// Создаем пользователя
		newUser := user.New(email, string(hashedPassword), role)
		if err := s.userRepo.Create(ctx, newUser); err != nil {
			return err
		}

		result = newUser
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Login выполняет авторизацию пользователя
func (s *Service) Login(ctx context.Context, email, password string) (*user.User, error) {
	// Получаем пользователя
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return u, nil
}

// GetByID получает пользователя по ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// Update обновляет данные пользователя
func (s *Service) Update(ctx context.Context, u *user.User) error {
	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Проверяем существование пользователя
		if _, err := s.userRepo.GetByID(ctx, u.ID); err != nil {
			return ErrUserNotFound
		}

		// Валидация данных
		if err := validateEmail(u.Email); err != nil {
			return err
		}
		if err := validateRole(u.Role); err != nil {
			return err
		}

		return s.userRepo.Update(ctx, u)
	})
}

// Delete удаляет пользователя
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Проверяем существование пользователя
		if _, err := s.userRepo.GetByID(ctx, id); err != nil {
			return ErrUserNotFound
		}

		return s.userRepo.Delete(ctx, id)
	})
}

// List возвращает список пользователей
func (s *Service) List(ctx context.Context, offset, limit int) ([]*user.User, error) {
	return s.userRepo.List(ctx, offset, limit)
}

// validateEmail проверяет корректность email
func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	if len(email) < 3 || len(email) > 255 {
		return ErrInvalidEmail
	}

	// Проверяем формат email
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return ErrInvalidEmail
	}

	// Проверяем допустимые символы
	for _, char := range email {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) && char != '@' && char != '.' && char != '_' && char != '-' {
			return ErrInvalidEmail
		}
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ErrInvalidEmail
	}

	return nil
}

// validatePassword проверяет корректность пароля
func validatePassword(password string) error {
	if password == "" {
		return ErrInvalidPassword
	}
	if len(password) < 8 || len(password) > 72 {
		return ErrInvalidPassword
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return ErrInvalidPassword
	}

	return nil
}

// validateRole проверяет корректность роли
func validateRole(role user.Role) error {
	switch role {
	case user.RoleAdmin, user.RoleEmployee:
		return nil
	default:
		return ErrInvalidRole
	}
}

// LoginUser выполняет вход пользователя
func (s *Service) LoginUser(ctx context.Context, email, password string) (string, error) {
	// Получаем пользователя по email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	// Проверяем пароль
	if !auth.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid password")
	}

	// Генерируем JWT токен
	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}

// DummyLogin выполняет вход с фиктивными данными (для тестирования)
func (s *Service) DummyLogin(ctx context.Context, role user.Role) (string, error) {
	// Генерируем фиктивный ID
	id := uuid.New()

	// Генерируем JWT токен
	token, err := auth.GenerateToken(id, role)
	if err != nil {
		return "", err
	}

	return token, nil
}
