package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/golf_messenger/internal/models"
	"github.com/yourusername/golf_messenger/internal/service"
	"go.uber.org/zap"
)

type MockInvitationRepository struct {
	mock.Mock
}

func (m *MockInvitationRepository) Create(invitation *models.Invitation) error {
	args := m.Called(invitation)
	return args.Error(0)
}

func (m *MockInvitationRepository) FindByID(id uuid.UUID) (*models.Invitation, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invitation), args.Error(1)
}

func (m *MockInvitationRepository) FindReceivedByUserID(userID uuid.UUID) ([]*models.Invitation, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Invitation), args.Error(1)
}

func (m *MockInvitationRepository) FindSentByUserID(userID uuid.UUID) ([]*models.Invitation, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Invitation), args.Error(1)
}

func (m *MockInvitationRepository) Update(invitation *models.Invitation) error {
	args := m.Called(invitation)
	return args.Error(0)
}

func (m *MockInvitationRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockInvitationRepository) FindByTTRAndInvitee(ttrID uuid.UUID, inviteeUserID uuid.UUID) (*models.Invitation, error) {
	args := m.Called(ttrID, inviteeUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invitation), args.Error(1)
}

func TestCreateInvitation_Authorization(t *testing.T) {
	mockInvitationRepo := new(MockInvitationRepository)
	mockTTRRepo := new(MockTTRRepository)
	mockUserRepo := new(MockUserRepository)
	logger, _ := zap.NewDevelopment()
	notificationService := service.NewNotificationService(logger)
	invitationService := service.NewInvitationService(mockInvitationRepo, mockTTRRepo, mockUserRepo, notificationService, logger)

	captainID := uuid.New()
	inviterID := uuid.New()
	inviteeID := uuid.New()
	ttrID := uuid.New()

	ttr := &models.TTR{
		ID:            ttrID,
		CaptainUserID: captainID,
		MaxPlayers:    4,
	}

	mockTTRRepo.On("FindByID", ttrID).Return(ttr, nil)
	mockTTRRepo.On("IsCoCaptain", ttrID, inviterID).Return(false, nil)

	_, err := invitationService.CreateInvitation(ttrID, inviterID, inviteeID, nil)

	assert.Error(t, err)
	assert.Equal(t, "unauthorized: only captain or co-captain can send invitations", err.Error())
	mockTTRRepo.AssertExpectations(t)
}

func TestCreateInvitation_DuplicatePrevention(t *testing.T) {
	mockInvitationRepo := new(MockInvitationRepository)
	mockTTRRepo := new(MockTTRRepository)
	mockUserRepo := new(MockUserRepository)
	logger, _ := zap.NewDevelopment()
	notificationService := service.NewNotificationService(logger)
	invitationService := service.NewInvitationService(mockInvitationRepo, mockTTRRepo, mockUserRepo, notificationService, logger)

	captainID := uuid.New()
	inviteeID := uuid.New()
	ttrID := uuid.New()

	ttr := &models.TTR{
		ID:            ttrID,
		CaptainUserID: captainID,
		MaxPlayers:    4,
	}

	invitee := &models.User{
		ID:        inviteeID,
		Email:     "invitee@example.com",
		FirstName: "Jane",
		LastName:  "Doe",
	}

	existingInvitation := &models.Invitation{
		ID:            uuid.New(),
		TTRID:         ttrID,
		InviterUserID: captainID,
		InviteeUserID: inviteeID,
		Status:        models.InvitationStatusPending,
	}

	mockTTRRepo.On("FindByID", ttrID).Return(ttr, nil)
	mockTTRRepo.On("IsCoCaptain", ttrID, captainID).Return(false, nil)
	mockUserRepo.On("FindByID", inviteeID).Return(invitee, nil)
	mockTTRRepo.On("GetPlayers", ttrID).Return([]*models.TTRPlayer{}, nil)
	mockTTRRepo.On("IsPlayer", ttrID, inviteeID).Return(false, nil)
	mockInvitationRepo.On("FindByTTRAndInvitee", ttrID, inviteeID).Return(existingInvitation, nil)

	_, err := invitationService.CreateInvitation(ttrID, captainID, inviteeID, nil)

	assert.Error(t, err)
	assert.Equal(t, "pending invitation already exists for this user", err.Error())
	mockTTRRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockInvitationRepo.AssertExpectations(t)
}

func TestRespondToInvitation_AcceptJoinsTTR(t *testing.T) {
	mockInvitationRepo := new(MockInvitationRepository)
	mockTTRRepo := new(MockTTRRepository)
	mockUserRepo := new(MockUserRepository)
	logger, _ := zap.NewDevelopment()
	notificationService := service.NewNotificationService(logger)
	invitationService := service.NewInvitationService(mockInvitationRepo, mockTTRRepo, mockUserRepo, notificationService, logger)

	inviteeID := uuid.New()
	ttrID := uuid.New()
	invitationID := uuid.New()

	invitation := &models.Invitation{
		ID:            invitationID,
		TTRID:         ttrID,
		InviterUserID: uuid.New(),
		InviteeUserID: inviteeID,
		Status:        models.InvitationStatusPending,
		CreatedAt:     time.Now(),
	}

	ttr := &models.TTR{
		ID:         ttrID,
		MaxPlayers: 4,
	}

	mockInvitationRepo.On("FindByID", invitationID).Return(invitation, nil)
	mockTTRRepo.On("FindByID", ttrID).Return(ttr, nil)
	mockTTRRepo.On("GetPlayers", ttrID).Return([]*models.TTRPlayer{{UserID: uuid.New()}}, nil)
	mockTTRRepo.On("AddPlayer", ttrID, inviteeID, models.TTRPlayerStatusConfirmed).Return(nil)
	mockInvitationRepo.On("Update", mock.AnythingOfType("*models.Invitation")).Return(nil)
	mockInvitationRepo.On("FindByID", invitationID).Return(&models.Invitation{
		ID:            invitationID,
		TTRID:         ttrID,
		InviterUserID: invitation.InviterUserID,
		InviteeUserID: inviteeID,
		Status:        models.InvitationStatusYes,
		CreatedAt:     invitation.CreatedAt,
		RespondedAt:   &time.Time{},
	}, nil)

	result, err := invitationService.RespondToInvitation(invitationID, inviteeID, models.InvitationStatusYes)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.InvitationStatusYes, result.Status)
	mockInvitationRepo.AssertCalled(t, "Update", mock.AnythingOfType("*models.Invitation"))
	mockTTRRepo.AssertCalled(t, "AddPlayer", ttrID, inviteeID, models.TTRPlayerStatusConfirmed)
	mockInvitationRepo.AssertExpectations(t)
	mockTTRRepo.AssertExpectations(t)
}

func TestRespondToInvitation_WhenTTRFull(t *testing.T) {
	mockInvitationRepo := new(MockInvitationRepository)
	mockTTRRepo := new(MockTTRRepository)
	mockUserRepo := new(MockUserRepository)
	logger, _ := zap.NewDevelopment()
	notificationService := service.NewNotificationService(logger)
	invitationService := service.NewInvitationService(mockInvitationRepo, mockTTRRepo, mockUserRepo, notificationService, logger)

	inviteeID := uuid.New()
	ttrID := uuid.New()
	invitationID := uuid.New()

	invitation := &models.Invitation{
		ID:            invitationID,
		TTRID:         ttrID,
		InviterUserID: uuid.New(),
		InviteeUserID: inviteeID,
		Status:        models.InvitationStatusPending,
		CreatedAt:     time.Now(),
	}

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

	mockInvitationRepo.On("FindByID", invitationID).Return(invitation, nil)
	mockTTRRepo.On("FindByID", ttrID).Return(ttr, nil)
	mockTTRRepo.On("GetPlayers", ttrID).Return(players, nil)

	_, err := invitationService.RespondToInvitation(invitationID, inviteeID, models.InvitationStatusYes)

	assert.Error(t, err)
	assert.Equal(t, "TTR is full, cannot accept invitation", err.Error())
	mockInvitationRepo.AssertExpectations(t)
	mockTTRRepo.AssertExpectations(t)
}
