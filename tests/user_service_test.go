package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/golf_messenger/internal/models"
	"github.com/yourusername/golf_messenger/internal/service"
	"github.com/yourusername/golf_messenger/pkg/storage"
)

func TestUserService_GetProfile_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)

	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	mockUserRepo.On("FindByID", userID).Return(user, nil)

	userService := service.NewUserService(mockUserRepo, nil)

	result, err := userService.GetProfile(userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "John", result.FirstName)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetProfile_NotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)

	userID := uuid.New()

	mockUserRepo.On("FindByID", userID).Return(nil, nil)

	userService := service.NewUserService(mockUserRepo, nil)

	result, err := userService.GetProfile(userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user not found", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfile_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)

	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	mockUserRepo.On("FindByID", userID).Return(user, nil)
	mockUserRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)

	userService := service.NewUserService(mockUserRepo, nil)

	handicap := 15.5
	result, err := userService.UpdateProfile(userID, "Jane", "Smith", &handicap, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Jane", result.FirstName)
	assert.Equal(t, "Smith", result.LastName)
	assert.Equal(t, &handicap, result.Handicap)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfile_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)

	userID := uuid.New()

	mockUserRepo.On("FindByID", userID).Return(nil, nil)

	userService := service.NewUserService(mockUserRepo, nil)

	result, err := userService.UpdateProfile(userID, "Jane", "Smith", nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user not found", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)

	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}
	user.SetPassword("oldpassword123")

	mockUserRepo.On("FindByID", userID).Return(user, nil)
	mockUserRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)

	userService := service.NewUserService(mockUserRepo, nil)

	err := userService.ChangePassword(userID, "oldpassword123", "newpassword123")

	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_InvalidOldPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)

	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}
	user.SetPassword("oldpassword123")

	mockUserRepo.On("FindByID", userID).Return(user, nil)

	userService := service.NewUserService(mockUserRepo, nil)

	err := userService.ChangePassword(userID, "wrongpassword", "newpassword123")

	assert.Error(t, err)
	assert.Equal(t, "invalid old password", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_SearchUsers_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)

	users := []*models.User{
		{
			ID:        uuid.New(),
			Email:     "john@example.com",
			FirstName: "John",
			LastName:  "Doe",
		},
		{
			ID:        uuid.New(),
			Email:     "jane@example.com",
			FirstName: "Jane",
			LastName:  "Doe",
		},
	}

	mockUserRepo.On("Search", "doe", 20, 0).Return(users, nil)

	userService := service.NewUserService(mockUserRepo, nil)

	result, err := userService.SearchUsers("doe", 20, 0)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_SearchUsers_EmptyQuery(t *testing.T) {
	mockUserRepo := new(MockUserRepository)

	userService := service.NewUserService(mockUserRepo, nil)

	result, err := userService.SearchUsers("  ", 20, 0)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}
