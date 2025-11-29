package integration

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

type MockTTRRepository struct {
	mock.Mock
	ttrs      map[uuid.UUID]*models.TTR
	players   map[uuid.UUID]map[uuid.UUID]*models.TTRPlayer
	coCaptains map[uuid.UUID]map[uuid.UUID]*models.TTRCoCaptain
}

func NewMockTTRRepository() *MockTTRRepository {
	return &MockTTRRepository{
		ttrs:      make(map[uuid.UUID]*models.TTR),
		players:   make(map[uuid.UUID]map[uuid.UUID]*models.TTRPlayer),
		coCaptains: make(map[uuid.UUID]map[uuid.UUID]*models.TTRCoCaptain),
	}
}

func (m *MockTTRRepository) Create(ttr *models.TTR) error {
	if ttr.ID == uuid.Nil {
		ttr.ID = uuid.New()
	}
	m.ttrs[ttr.ID] = ttr
	m.players[ttr.ID] = make(map[uuid.UUID]*models.TTRPlayer)
	m.coCaptains[ttr.ID] = make(map[uuid.UUID]*models.TTRCoCaptain)
	return nil
}

func (m *MockTTRRepository) FindByID(id uuid.UUID) (*models.TTR, error) {
	if ttr, exists := m.ttrs[id]; exists {
		ttrCopy := *ttr
		if playerMap, ok := m.players[id]; ok {
			players := make([]models.TTRPlayer, 0, len(playerMap))
			for _, p := range playerMap {
				players = append(players, *p)
			}
			ttrCopy.Players = players
		}
		if ccMap, ok := m.coCaptains[id]; ok {
			coCaptains := make([]models.TTRCoCaptain, 0, len(ccMap))
			for _, cc := range ccMap {
				coCaptains = append(coCaptains, *cc)
			}
			ttrCopy.CoCaptains = coCaptains
		}
		return &ttrCopy, nil
	}
	return nil, nil
}

func (m *MockTTRRepository) FindAll(limit int, offset int, status string) ([]*models.TTR, error) {
	result := make([]*models.TTR, 0)
	for _, ttr := range m.ttrs {
		if status == "" || ttr.Status == status {
			result = append(result, ttr)
		}
	}
	return result, nil
}

func (m *MockTTRRepository) Update(ttr *models.TTR) error {
	m.ttrs[ttr.ID] = ttr
	return nil
}

func (m *MockTTRRepository) Delete(id uuid.UUID) error {
	delete(m.ttrs, id)
	return nil
}

func (m *MockTTRRepository) FindUpcomingByUserID(userID uuid.UUID) ([]*models.TTR, error) {
	return nil, nil
}

func (m *MockTTRRepository) FindPastByUserID(userID uuid.UUID) ([]*models.TTR, error) {
	return nil, nil
}

func (m *MockTTRRepository) AddCoCaptain(ttrID uuid.UUID, userID uuid.UUID) error {
	if _, ok := m.coCaptains[ttrID]; !ok {
		m.coCaptains[ttrID] = make(map[uuid.UUID]*models.TTRCoCaptain)
	}
	m.coCaptains[ttrID][userID] = &models.TTRCoCaptain{
		TTRID:      ttrID,
		UserID:     userID,
		AssignedAt: time.Now(),
	}
	return nil
}

func (m *MockTTRRepository) RemoveCoCaptain(ttrID uuid.UUID, userID uuid.UUID) error {
	if ccMap, ok := m.coCaptains[ttrID]; ok {
		delete(ccMap, userID)
	}
	return nil
}

func (m *MockTTRRepository) IsCoCaptain(ttrID uuid.UUID, userID uuid.UUID) (bool, error) {
	if ccMap, ok := m.coCaptains[ttrID]; ok {
		_, exists := ccMap[userID]
		return exists, nil
	}
	return false, nil
}

func (m *MockTTRRepository) AddPlayer(ttrID uuid.UUID, userID uuid.UUID, status string) error {
	if _, ok := m.players[ttrID]; !ok {
		m.players[ttrID] = make(map[uuid.UUID]*models.TTRPlayer)
	}
	m.players[ttrID][userID] = &models.TTRPlayer{
		TTRID:    ttrID,
		UserID:   userID,
		JoinedAt: time.Now(),
		Status:   status,
	}
	return nil
}

func (m *MockTTRRepository) RemovePlayer(ttrID uuid.UUID, userID uuid.UUID) error {
	if playerMap, ok := m.players[ttrID]; ok {
		delete(playerMap, userID)
	}
	return nil
}

func (m *MockTTRRepository) GetPlayers(ttrID uuid.UUID) ([]*models.TTRPlayer, error) {
	result := make([]*models.TTRPlayer, 0)
	if playerMap, ok := m.players[ttrID]; ok {
		for _, player := range playerMap {
			result = append(result, player)
		}
	}
	return result, nil
}

func (m *MockTTRRepository) IsPlayer(ttrID uuid.UUID, userID uuid.UUID) (bool, error) {
	if playerMap, ok := m.players[ttrID]; ok {
		_, exists := playerMap[userID]
		return exists, nil
	}
	return false, nil
}

type MockUserRepository struct {
	users map[uuid.UUID]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (m *MockUserRepository) Create(user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	return nil, nil
}

func (m *MockUserRepository) Update(user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Search(query string, limit int, offset int) ([]*models.User, error) {
	return nil, nil
}

type MockInvitationRepository struct {
	invitations map[uuid.UUID]*models.Invitation
}

func NewMockInvitationRepository() *MockInvitationRepository {
	return &MockInvitationRepository{
		invitations: make(map[uuid.UUID]*models.Invitation),
	}
}

func (m *MockInvitationRepository) Create(invitation *models.Invitation) error {
	if invitation.ID == uuid.Nil {
		invitation.ID = uuid.New()
	}
	m.invitations[invitation.ID] = invitation
	return nil
}

func (m *MockInvitationRepository) FindByID(id uuid.UUID) (*models.Invitation, error) {
	if inv, exists := m.invitations[id]; exists {
		return inv, nil
	}
	return nil, nil
}

func (m *MockInvitationRepository) FindReceivedByUserID(userID uuid.UUID) ([]*models.Invitation, error) {
	return nil, nil
}

func (m *MockInvitationRepository) FindSentByUserID(userID uuid.UUID) ([]*models.Invitation, error) {
	return nil, nil
}

func (m *MockInvitationRepository) Update(invitation *models.Invitation) error {
	m.invitations[invitation.ID] = invitation
	return nil
}

func (m *MockInvitationRepository) Delete(id uuid.UUID) error {
	delete(m.invitations, id)
	return nil
}

func (m *MockInvitationRepository) FindByTTRAndInvitee(ttrID uuid.UUID, inviteeUserID uuid.UUID) (*models.Invitation, error) {
	for _, inv := range m.invitations {
		if inv.TTRID == ttrID && inv.InviteeUserID == inviteeUserID && inv.Status == models.InvitationStatusPending {
			return inv, nil
		}
	}
	return nil, nil
}

func TestTTRCompleteFlow(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockTTRRepo := NewMockTTRRepository()
	mockUserRepo := NewMockUserRepository()
	mockInvitationRepo := NewMockInvitationRepository()

	notificationService := service.NewNotificationService(logger)
	ttrService := service.NewTTRService(mockTTRRepo, mockUserRepo, logger)
	invitationService := service.NewInvitationService(mockInvitationRepo, mockTTRRepo, mockUserRepo, notificationService, logger)

	captainID := uuid.New()
	captain := &models.User{
		ID:        captainID,
		Email:     "captain@example.com",
		FirstName: "Captain",
		LastName:  "Smith",
	}
	mockUserRepo.Create(captain)

	coCaptainID := uuid.New()
	coCaptain := &models.User{
		ID:        coCaptainID,
		Email:     "cocaptain@example.com",
		FirstName: "CoCaptain",
		LastName:  "Jones",
	}
	mockUserRepo.Create(coCaptain)

	playerID := uuid.New()
	player := &models.User{
		ID:        playerID,
		Email:     "player@example.com",
		FirstName: "Player",
		LastName:  "Brown",
	}
	mockUserRepo.Create(player)

	courseName := "Pebble Beach"
	courseLocation := "California"
	teeDate := time.Now().Add(24 * time.Hour)
	teeTime := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)
	maxPlayers := 4
	notes := "Fun round"

	ttr, err := ttrService.CreateTTR(captainID, courseName, &courseLocation, teeDate, teeTime, maxPlayers, &notes)
	assert.NoError(t, err)
	assert.NotNil(t, ttr)
	assert.Equal(t, captainID, ttr.CaptainUserID)
	t.Logf("Step 1: TTR created with ID: %s", ttr.ID)

	err = ttrService.AddCoCaptain(ttr.ID, captainID, coCaptainID)
	assert.NoError(t, err)
	t.Logf("Step 2: Co-captain added")

	message := "Join us for golf!"
	invitation, err := invitationService.CreateInvitation(ttr.ID, captainID, playerID, &message)
	assert.NoError(t, err)
	assert.NotNil(t, invitation)
	assert.Equal(t, models.InvitationStatusPending, invitation.Status)
	t.Logf("Step 3: Invitation sent to player")

	respondedInvitation, err := invitationService.RespondToInvitation(invitation.ID, playerID, models.InvitationStatusYes)
	assert.NoError(t, err)
	assert.Equal(t, models.InvitationStatusYes, respondedInvitation.Status)
	t.Logf("Step 4: Player accepted invitation")

	players, err := ttrService.GetPlayers(ttr.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(players))
	t.Logf("Step 5: Verified player was added to TTR (total players: %d)", len(players))

	err = ttrService.UpdatePlayerStatus(ttr.ID, captainID, playerID, models.TTRPlayerStatusMaybe)
	assert.NoError(t, err)
	t.Logf("Step 6: Captain updated player status to MAYBE")

	err = ttrService.LeaveTTR(ttr.ID, playerID)
	assert.NoError(t, err)
	t.Logf("Step 7: Player left TTR")

	playersAfterLeave, err := ttrService.GetPlayers(ttr.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(playersAfterLeave))
	t.Logf("Step 8: Verified player was removed (remaining players: %d)", len(playersAfterLeave))
}
