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
	"go.uber.org/zap"
)

type MockTTRRepository struct {
	mock.Mock
}

func (m *MockTTRRepository) Create(ttr *models.TTR) error {
	args := m.Called(ttr)
	return args.Error(0)
}

func (m *MockTTRRepository) FindByID(id uuid.UUID) (*models.TTR, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TTR), args.Error(1)
}

func (m *MockTTRRepository) FindAll(limit int, offset int, status string) ([]*models.TTR, error) {
	args := m.Called(limit, offset, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TTR), args.Error(1)
}

func (m *MockTTRRepository) Update(ttr *models.TTR) error {
	args := m.Called(ttr)
	return args.Error(0)
}

func (m *MockTTRRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTTRRepository) FindUpcomingByUserID(userID uuid.UUID) ([]*models.TTR, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TTR), args.Error(1)
}

func (m *MockTTRRepository) FindPastByUserID(userID uuid.UUID) ([]*models.TTR, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TTR), args.Error(1)
}

func (m *MockTTRRepository) AddCoCaptain(ttrID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ttrID, userID)
	return args.Error(0)
}

func (m *MockTTRRepository) RemoveCoCaptain(ttrID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ttrID, userID)
	return args.Error(0)
}

func (m *MockTTRRepository) IsCoCaptain(ttrID uuid.UUID, userID uuid.UUID) (bool, error) {
	args := m.Called(ttrID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTTRRepository) AddPlayer(ttrID uuid.UUID, userID uuid.UUID, status string) error {
	args := m.Called(ttrID, userID, status)
	return args.Error(0)
}

func (m *MockTTRRepository) RemovePlayer(ttrID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ttrID, userID)
	return args.Error(0)
}

func (m *MockTTRRepository) GetPlayers(ttrID uuid.UUID) ([]*models.TTRPlayer, error) {
	args := m.Called(ttrID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TTRPlayer), args.Error(1)
}

func (m *MockTTRRepository) IsPlayer(ttrID uuid.UUID, userID uuid.UUID) (bool, error) {
	args := m.Called(ttrID, userID)
	return args.Bool(0), args.Error(1)
}

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

func TestCreateTTR(t *testing.T) {
	mockTTRRepo := new(MockTTRRepository)
	mockUserRepo := new(MockUserRepository)
	logger, _ := zap.NewDevelopment()
	ttrService := service.NewTTRService(mockTTRRepo, mockUserRepo, logger)

	userID := uuid.New()
	courseName := "Pebble Beach"
	courseLocation := "California"
	teeDate := time.Now().Add(24 * time.Hour)
	teeTime := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)
	maxPlayers := 4
	notes := "Casual round"

	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	mockUserRepo.On("FindByID", userID).Return(user, nil)
	mockTTRRepo.On("Create", mock.AnythingOfType("*models.TTR")).Return(nil)
	mockTTRRepo.On("AddPlayer", mock.AnythingOfType("uuid.UUID"), userID, models.TTRPlayerStatusConfirmed).Return(nil)
	mockTTRRepo.On("FindByID", mock.AnythingOfType("uuid.UUID")).Return(&models.TTR{
		ID:              uuid.New(),
		CourseName:      courseName,
		CourseLocation:  &courseLocation,
		TeeDate:         teeDate,
		TeeTime:         teeTime,
		MaxPlayers:      maxPlayers,
		CreatedByUserID: userID,
		CaptainUserID:   userID,
		Status:          models.TTRStatusOpen,
		Notes:           &notes,
	}, nil)

	ttr, err := ttrService.CreateTTR(userID, courseName, &courseLocation, teeDate, teeTime, maxPlayers, &notes)

	assert.NoError(t, err)
	assert.NotNil(t, ttr)
	assert.Equal(t, userID, ttr.CaptainUserID)
	assert.Equal(t, userID, ttr.CreatedByUserID)
	assert.Equal(t, models.TTRStatusOpen, ttr.Status)
	mockTTRRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUpdateTTR_Authorization(t *testing.T) {
	mockTTRRepo := new(MockTTRRepository)
	mockUserRepo := new(MockUserRepository)
	logger, _ := zap.NewDevelopment()
	ttrService := service.NewTTRService(mockTTRRepo, mockUserRepo, logger)

	captainID := uuid.New()
	nonCaptainID := uuid.New()
	ttrID := uuid.New()

	ttr := &models.TTR{
		ID:            ttrID,
		CaptainUserID: captainID,
	}

	mockTTRRepo.On("FindByID", ttrID).Return(ttr, nil)
	mockTTRRepo.On("IsCoCaptain", ttrID, nonCaptainID).Return(false, nil)

	newCourseName := "Augusta National"
	_, err := ttrService.UpdateTTR(ttrID, nonCaptainID, &newCourseName, nil, nil, nil, nil, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, "unauthorized: only captain or co-captain can update TTR", err.Error())
	mockTTRRepo.AssertExpectations(t)
}

func TestAddCoCaptain_Authorization(t *testing.T) {
	mockTTRRepo := new(MockTTRRepository)
	mockUserRepo := new(MockUserRepository)
	logger, _ := zap.NewDevelopment()
	ttrService := service.NewTTRService(mockTTRRepo, mockUserRepo, logger)

	captainID := uuid.New()
	nonCaptainID := uuid.New()
	coCaptainID := uuid.New()
	ttrID := uuid.New()

	ttr := &models.TTR{
		ID:            ttrID,
		CaptainUserID: captainID,
	}

	mockTTRRepo.On("FindByID", ttrID).Return(ttr, nil)

	err := ttrService.AddCoCaptain(ttrID, nonCaptainID, coCaptainID)

	assert.Error(t, err)
	assert.Equal(t, "unauthorized: only captain can add co-captains", err.Error())
	mockTTRRepo.AssertExpectations(t)
}

func TestJoinTTR_WhenFull(t *testing.T) {
	mockTTRRepo := new(MockTTRRepository)
	mockUserRepo := new(MockUserRepository)
	logger, _ := zap.NewDevelopment()
	ttrService := service.NewTTRService(mockTTRRepo, mockUserRepo, logger)

	userID := uuid.New()
	ttrID := uuid.New()

	ttr := &models.TTR{
		ID:         ttrID,
		MaxPlayers: 4,
	}

	players := []*models.TTRPlayer{
		{UserID: uuid.New()},
		{UserID: uuid.New()},
		{UserID: uuid.New()},
		{UserID: uuid.New()},
	}

	mockTTRRepo.On("FindByID", ttrID).Return(ttr, nil)
	mockTTRRepo.On("GetPlayers", ttrID).Return(players, nil)

	err := ttrService.JoinTTR(ttrID, userID)

	assert.Error(t, err)
	assert.Equal(t, "TTR is full", err.Error())
	mockTTRRepo.AssertExpectations(t)
}

func TestUpdatePlayerStatus_Authorization(t *testing.T) {
	mockTTRRepo := new(MockTTRRepository)
	mockUserRepo := new(MockUserRepository)
	logger, _ := zap.NewDevelopment()
	ttrService := service.NewTTRService(mockTTRRepo, mockUserRepo, logger)

	captainID := uuid.New()
	nonManagerID := uuid.New()
	playerID := uuid.New()
	ttrID := uuid.New()

	ttr := &models.TTR{
		ID:            ttrID,
		CaptainUserID: captainID,
	}

	mockTTRRepo.On("FindByID", ttrID).Return(ttr, nil)
	mockTTRRepo.On("IsCoCaptain", ttrID, nonManagerID).Return(false, nil)

	err := ttrService.UpdatePlayerStatus(ttrID, nonManagerID, playerID, models.TTRPlayerStatusMaybe)

	assert.Error(t, err)
	assert.Equal(t, "unauthorized: only captain or co-captain can update player status", err.Error())
	mockTTRRepo.AssertExpectations(t)
}
