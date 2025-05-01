package user

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/avito/pvz/internal/domain/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository реализует мок для user.Repository
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

// MockTransactionManager реализует мок для transaction.Manager
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func TestService_Register(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		role        user.Role
		setupMocks  func(*MockUserRepository, *MockTransactionManager)
		expectedErr error
	}{
		{
			name:     "успешная регистрация",
			email:    "test@example.com",
			password: "StrongPass123!",
			role:     user.RoleAdmin,
			setupMocks: func(userRepo *MockUserRepository, txManager *MockTransactionManager) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, user.ErrNotFound)
				userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
					return u.Email == "test@example.com" && u.Role == user.RoleAdmin && len(u.Password) > 0
				})).Run(func(args mock.Arguments) {
					u := args.Get(1).(*user.User)
					u.ID = uuid.New()
				}).Return(nil)

				txManager.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:     "пользователь уже существует",
			email:    "existing@example.com",
			password: "StrongPass123!",
			role:     user.RoleAdmin,
			setupMocks: func(userRepo *MockUserRepository, txManager *MockTransactionManager) {
				existingUser := &user.User{
					ID:    uuid.New(),
					Email: "existing@example.com",
				}
				userRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
				txManager.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrUserAlreadyExists)
			},
			expectedErr: ErrUserAlreadyExists,
		},
		{
			name:     "неверный email",
			email:    "invalid-email",
			password: "StrongPass123!",
			role:     user.RoleAdmin,
			setupMocks: func(userRepo *MockUserRepository, txManager *MockTransactionManager) {
				// Моки не нужны, так как валидация происходит до их вызова
			},
			expectedErr: ErrInvalidEmail,
		},
		{
			name:     "слабый пароль",
			email:    "test@example.com",
			password: "weak",
			role:     user.RoleAdmin,
			setupMocks: func(userRepo *MockUserRepository, txManager *MockTransactionManager) {
				// Моки не нужны, так как валидация происходит до их вызова
			},
			expectedErr: ErrInvalidPassword,
		},
		{
			name:     "неверная роль",
			email:    "test@example.com",
			password: "StrongPass123!",
			role:     "invalid_role",
			setupMocks: func(userRepo *MockUserRepository, txManager *MockTransactionManager) {
				// Моки не нужны, так как валидация происходит до их вызова
			},
			expectedErr: ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			txManager := new(MockTransactionManager)
			tt.setupMocks(userRepo, txManager)

			service := New(userRepo, txManager)
			result, err := service.Register(context.Background(), tt.email, tt.password, tt.role)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.email, result.Email)
				assert.Equal(t, tt.role, result.Role)
				assert.NotEmpty(t, result.Password)
				assert.NotEqual(t, uuid.Nil, result.ID)
			}

			userRepo.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestService_Login(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		setupMocks  func(*MockUserRepository, *MockTransactionManager)
		expectedErr error
	}{
		{
			name:     "успешный вход",
			email:    "test@example.com",
			password: "StrongPass123!",
			setupMocks: func(userRepo *MockUserRepository, txManager *MockTransactionManager) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
				user := &user.User{
					Email:    "test@example.com",
					Password: string(hashedPassword),
				}
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			expectedErr: nil,
		},
		{
			name:     "пользователь не найден",
			email:    "nonexistent@example.com",
			password: "StrongPass123!",
			setupMocks: func(userRepo *MockUserRepository, txManager *MockTransactionManager) {
				userRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, user.ErrNotFound)
			},
			expectedErr: ErrUserNotFound,
		},
		{
			name:     "неверный пароль",
			email:    "test@example.com",
			password: "WrongPass123!",
			setupMocks: func(userRepo *MockUserRepository, txManager *MockTransactionManager) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
				user := &user.User{
					Email:    "test@example.com",
					Password: string(hashedPassword),
				}
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			expectedErr: ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			txManager := new(MockTransactionManager)
			tt.setupMocks(userRepo, txManager)

			service := New(userRepo, txManager)
			result, err := service.Login(context.Background(), tt.email, tt.password)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.email, result.Email)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		id            uuid.UUID
		setupMocks    func(*MockUserRepository)
		expectedError error
	}{
		{
			name: "успешное получение пользователя",
			id:   uuid.New(),
			setupMocks: func(repo *MockUserRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&user.User{}, nil)
			},
			expectedError: nil,
		},
		{
			name: "пользователь не найден",
			id:   uuid.New(),
			setupMocks: func(repo *MockUserRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, user.ErrUserNotFound)
			},
			expectedError: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockUserRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(repo)

			service := New(repo, tx)
			_, err := service.GetByID(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name          string
		user          *user.User
		setupMocks    func(*MockUserRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name: "успешное обновление",
			user: &user.User{
				ID:    uuid.New(),
				Email: "test@example.com",
				Role:  user.RoleAdmin,
			},
			setupMocks: func(repo *MockUserRepository, tx *MockTransactionManager) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&user.User{}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "пользователь не найден",
			user: &user.User{
				ID:    uuid.New(),
				Email: "test@example.com",
				Role:  user.RoleAdmin,
			},
			setupMocks: func(repo *MockUserRepository, tx *MockTransactionManager) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, user.ErrUserNotFound)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrUserNotFound)
			},
			expectedError: ErrUserNotFound,
		},
		{
			name: "некорректный email",
			user: &user.User{
				ID:    uuid.New(),
				Email: "",
				Role:  user.RoleAdmin,
			},
			setupMocks: func(repo *MockUserRepository, tx *MockTransactionManager) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&user.User{}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrInvalidEmail)
			},
			expectedError: ErrInvalidEmail,
		},
		{
			name: "ошибка при обновлении пользователя",
			user: &user.User{
				ID:    uuid.New(),
				Email: "test@example.com",
				Role:  user.RoleAdmin,
			},
			setupMocks: func(repo *MockUserRepository, tx *MockTransactionManager) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&user.User{}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(errors.New("database error"))
				repo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "некорректный email при обновлении",
			user: &user.User{
				ID:    uuid.New(),
				Email: "",
				Role:  user.RoleAdmin,
			},
			setupMocks: func(repo *MockUserRepository, tx *MockTransactionManager) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&user.User{}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrInvalidEmail)
			},
			expectedError: ErrInvalidEmail,
		},
		{
			name: "некорректная роль при обновлении",
			user: &user.User{
				ID:    uuid.New(),
				Email: "test@example.com",
				Role:  "invalid_role",
			},
			setupMocks: func(repo *MockUserRepository, tx *MockTransactionManager) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&user.User{}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrInvalidRole)
			},
			expectedError: ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockUserRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(repo, tx)

			service := New(repo, tx)
			err := service.Update(context.Background(), tt.user)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		id            uuid.UUID
		setupMocks    func(*MockUserRepository, *MockTransactionManager)
		expectedError error
	}{
		{
			name: "успешное удаление",
			id:   uuid.New(),
			setupMocks: func(repo *MockUserRepository, tx *MockTransactionManager) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&user.User{}, nil)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(nil)
				repo.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "пользователь не найден",
			id:   uuid.New(),
			setupMocks: func(repo *MockUserRepository, tx *MockTransactionManager) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, user.ErrUserNotFound)
				tx.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					fn(context.Background())
				}).Return(ErrUserNotFound)
			},
			expectedError: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockUserRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(repo, tx)

			service := New(repo, tx)
			err := service.Delete(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name          string
		offset        int
		limit         int
		setupMocks    func(*MockUserRepository)
		expectedError error
	}{
		{
			name:   "успешное получение списка",
			offset: 0,
			limit:  10,
			setupMocks: func(repo *MockUserRepository) {
				repo.On("List", mock.Anything, 0, 10).Return([]*user.User{}, nil)
			},
			expectedError: nil,
		},
		{
			name:   "ошибка при получении списка",
			offset: 0,
			limit:  10,
			setupMocks: func(repo *MockUserRepository) {
				repo.On("List", mock.Anything, 0, 10).Return([]*user.User(nil), errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockUserRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(repo)

			service := New(repo, tx)
			_, err := service.List(context.Background(), tt.offset, tt.limit)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_LoginUser(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func(*MockUserRepository)
		expectedError error
	}{
		{
			name:     "успешный вход",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(repo *MockUserRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				repo.On("GetByEmail", mock.Anything, "test@example.com").Return(&user.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Password: string(hashedPassword),
					Role:     user.RoleAdmin,
				}, nil)
			},
			expectedError: nil,
		},
		{
			name:     "пользователь не найден",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMocks: func(repo *MockUserRepository) {
				repo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, user.ErrUserNotFound)
			},
			expectedError: user.ErrUserNotFound,
		},
		{
			name:     "неверный пароль",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMocks: func(repo *MockUserRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				repo.On("GetByEmail", mock.Anything, "test@example.com").Return(&user.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Password: string(hashedPassword),
					Role:     user.RoleAdmin,
				}, nil)
			},
			expectedError: errors.New("invalid password"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockUserRepository)
			tx := new(MockTransactionManager)
			tt.setupMocks(repo)

			service := New(repo, tx)
			_, err := service.LoginUser(context.Background(), tt.email, tt.password)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_DummyLogin(t *testing.T) {
	tests := []struct {
		name          string
		role          user.Role
		expectedError error
	}{
		{
			name:          "успешный вход с ролью админа",
			role:          user.RoleAdmin,
			expectedError: nil,
		},
		{
			name:          "успешный вход с ролью сотрудника",
			role:          user.RoleEmployee,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockUserRepository)
			tx := new(MockTransactionManager)

			service := New(repo, tx)
			token, err := service.DummyLogin(context.Background(), tt.role)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{
			name:    "валидный email",
			email:   "test@example.com",
			wantErr: nil,
		},
		{
			name:    "пустой email",
			email:   "",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "email без @",
			email:   "testexample.com",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "email без домена",
			email:   "test@",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "email без локальной части",
			email:   "@example.com",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "email с недопустимыми символами",
			email:   "test!@example.com",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "слишком длинный email",
			email:   strings.Repeat("a", 256) + "@example.com",
			wantErr: ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.email)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "валидный пароль",
			password: "Test1234!",
			wantErr:  nil,
		},
		{
			name:     "пустой пароль",
			password: "",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "слишком короткий пароль",
			password: "Test1!",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "слишком длинный пароль",
			password: strings.Repeat("A", 73) + "a1!",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "пароль без заглавных букв",
			password: "test1234!",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "пароль без строчных букв",
			password: "TEST1234!",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "пароль без цифр",
			password: "TestTest!",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "пароль без специальных символов",
			password: "Test1234",
			wantErr:  ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
