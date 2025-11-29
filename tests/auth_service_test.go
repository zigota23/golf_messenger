package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/golf_messenger/internal/models"
	"github.com/yourusername/golf_messenger/internal/service"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Search(query string, limit int, offset int) ([]*models.User, error) {
	args := m.Called(query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(token *models.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) FindByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	args := m.Called(tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) RevokeByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpired() error {
	args := m.Called()
	return args.Error(0)
}

func TestAuthService_Register_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)

	mockUserRepo.On("FindByEmail", "test@example.com").Return(nil, nil)
	mockUserRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
	mockRefreshTokenRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	authService := service.NewAuthService(
		mockUserRepo,
		mockRefreshTokenRepo,
		"test-secret",
		15*time.Minute,
		7*24*time.Hour,
	)

	user, tokenPair, err := authService.Register("test@example.com", "password123", "John", "Doe")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotNil(t, tokenPair)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)

	mockUserRepo.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)

	existingUser := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	mockUserRepo.On("FindByEmail", "test@example.com").Return(existingUser, nil)

	authService := service.NewAuthService(
		mockUserRepo,
		mockRefreshTokenRepo,
		"test-secret",
		15*time.Minute,
		7*24*time.Hour,
	)

	user, tokenPair, err := authService.Register("test@example.com", "password123", "John", "Doe")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Nil(t, tokenPair)
	assert.Equal(t, "user with this email already exists", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)

	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}
	user.SetPassword("password123")

	mockUserRepo.On("FindByEmail", "test@example.com").Return(user, nil)
	mockRefreshTokenRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	authService := service.NewAuthService(
		mockUserRepo,
		mockRefreshTokenRepo,
		"test-secret",
		15*time.Minute,
		7*24*time.Hour,
	)

	loggedInUser, tokenPair, err := authService.Login("test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, loggedInUser)
	assert.NotNil(t, tokenPair)
	assert.Equal(t, user.ID, loggedInUser.ID)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)

	mockUserRepo.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)

	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}
	user.SetPassword("password123")

	mockUserRepo.On("FindByEmail", "test@example.com").Return(user, nil)

	authService := service.NewAuthService(
		mockUserRepo,
		mockRefreshTokenRepo,
		"test-secret",
		15*time.Minute,
		7*24*time.Hour,
	)

	loggedInUser, tokenPair, err := authService.Login("test@example.com", "wrongpassword")

	assert.Error(t, err)
	assert.Nil(t, loggedInUser)
	assert.Nil(t, tokenPair)
	assert.Equal(t, "invalid email or password", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)

	mockUserRepo.On("FindByEmail", "test@example.com").Return(nil, nil)

	authService := service.NewAuthService(
		mockUserRepo,
		mockRefreshTokenRepo,
		"test-secret",
		15*time.Minute,
		7*24*time.Hour,
	)

	loggedInUser, tokenPair, err := authService.Login("test@example.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, loggedInUser)
	assert.Nil(t, tokenPair)
	assert.Equal(t, "invalid email or password", err.Error())

	mockUserRepo.AssertExpectations(t)
}
