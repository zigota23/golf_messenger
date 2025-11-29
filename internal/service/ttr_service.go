package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/golf_messenger/internal/models"
	"github.com/yourusername/golf_messenger/internal/repository"
	"go.uber.org/zap"
)

type TTRService struct {
	ttrRepo  repository.TTRRepository
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewTTRService(ttrRepo repository.TTRRepository, userRepo repository.UserRepository, logger *zap.Logger) *TTRService {
	return &TTRService{
		ttrRepo:  ttrRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *TTRService) CreateTTR(userID uuid.UUID, courseName string, courseLocation *string, teeDate time.Time, teeTime time.Time, maxPlayers int, notes *string) (*models.TTR, error) {
	if maxPlayers <= 0 {
		return nil, errors.New("max_players must be greater than 0")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	ttr := &models.TTR{
		CourseName:      courseName,
		CourseLocation:  courseLocation,
		TeeDate:         teeDate,
		TeeTime:         teeTime,
		MaxPlayers:      maxPlayers,
		CreatedByUserID: userID,
		CaptainUserID:   userID,
		Status:          models.TTRStatusOpen,
		Notes:           notes,
	}

	if err := s.ttrRepo.Create(ttr); err != nil {
		return nil, fmt.Errorf("failed to create TTR: %w", err)
	}

	if err := s.ttrRepo.AddPlayer(ttr.ID, userID, models.TTRPlayerStatusConfirmed); err != nil {
		return nil, fmt.Errorf("failed to add captain as player: %w", err)
	}

	createdTTR, err := s.ttrRepo.FindByID(ttr.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created TTR: %w", err)
	}

	return createdTTR, nil
}

func (s *TTRService) GetTTR(id uuid.UUID) (*models.TTR, error) {
	ttr, err := s.ttrRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get TTR: %w", err)
	}
	if ttr == nil {
		return nil, errors.New("TTR not found")
	}
	return ttr, nil
}

func (s *TTRService) UpdateTTR(ttrID uuid.UUID, userID uuid.UUID, courseName *string, courseLocation *string, teeDate *time.Time, teeTime *time.Time, maxPlayers *int, status *string, notes *string) (*models.TTR, error) {
	canManage, err := s.canManageTTR(ttrID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !canManage {
		return nil, errors.New("unauthorized: only captain or co-captain can update TTR")
	}

	ttr, err := s.ttrRepo.FindByID(ttrID)
	if err != nil {
		return nil, fmt.Errorf("failed to find TTR: %w", err)
	}
	if ttr == nil {
		return nil, errors.New("TTR not found")
	}

	if courseName != nil {
		ttr.CourseName = *courseName
	}
	if courseLocation != nil {
		ttr.CourseLocation = courseLocation
	}
	if teeDate != nil {
		ttr.TeeDate = *teeDate
	}
	if teeTime != nil {
		ttr.TeeTime = *teeTime
	}
	if maxPlayers != nil {
		if *maxPlayers <= 0 {
			return nil, errors.New("max_players must be greater than 0")
		}
		ttr.MaxPlayers = *maxPlayers
	}
	if status != nil {
		ttr.Status = *status
	}
	if notes != nil {
		ttr.Notes = notes
	}

	if err := s.ttrRepo.Update(ttr); err != nil {
		return nil, fmt.Errorf("failed to update TTR: %w", err)
	}

	updatedTTR, err := s.ttrRepo.FindByID(ttrID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated TTR: %w", err)
	}

	return updatedTTR, nil
}

func (s *TTRService) DeleteTTR(ttrID uuid.UUID, userID uuid.UUID) error {
	isCaptain, err := s.isCaptain(ttrID, userID)
	if err != nil {
		return fmt.Errorf("failed to check captain status: %w", err)
	}
	if !isCaptain {
		return errors.New("unauthorized: only captain can delete TTR")
	}

	if err := s.ttrRepo.Delete(ttrID); err != nil {
		return fmt.Errorf("failed to delete TTR: %w", err)
	}

	return nil
}

func (s *TTRService) SearchTTRs(limit int, offset int, status string) ([]*models.TTR, error) {
	ttrs, err := s.ttrRepo.FindAll(limit, offset, status)
	if err != nil {
		return nil, fmt.Errorf("failed to search TTRs: %w", err)
	}
	return ttrs, nil
}

func (s *TTRService) AddCoCaptain(ttrID uuid.UUID, captainUserID uuid.UUID, coCaptainUserID uuid.UUID) error {
	isCaptain, err := s.isCaptain(ttrID, captainUserID)
	if err != nil {
		return fmt.Errorf("failed to check captain status: %w", err)
	}
	if !isCaptain {
		return errors.New("unauthorized: only captain can add co-captains")
	}

	coCaptainUser, err := s.userRepo.FindByID(coCaptainUserID)
	if err != nil {
		return fmt.Errorf("failed to find co-captain user: %w", err)
	}
	if coCaptainUser == nil {
		return errors.New("co-captain user not found")
	}

	isAlreadyCoCaptain, err := s.ttrRepo.IsCoCaptain(ttrID, coCaptainUserID)
	if err != nil {
		return fmt.Errorf("failed to check co-captain status: %w", err)
	}
	if isAlreadyCoCaptain {
		return errors.New("user is already a co-captain")
	}

	if err := s.ttrRepo.AddCoCaptain(ttrID, coCaptainUserID); err != nil {
		return fmt.Errorf("failed to add co-captain: %w", err)
	}

	return nil
}

func (s *TTRService) RemoveCoCaptain(ttrID uuid.UUID, captainUserID uuid.UUID, coCaptainUserID uuid.UUID) error {
	isCaptain, err := s.isCaptain(ttrID, captainUserID)
	if err != nil {
		return fmt.Errorf("failed to check captain status: %w", err)
	}
	if !isCaptain {
		return errors.New("unauthorized: only captain can remove co-captains")
	}

	if err := s.ttrRepo.RemoveCoCaptain(ttrID, coCaptainUserID); err != nil {
		return fmt.Errorf("failed to remove co-captain: %w", err)
	}

	return nil
}

func (s *TTRService) JoinTTR(ttrID uuid.UUID, userID uuid.UUID) error {
	ttr, err := s.ttrRepo.FindByID(ttrID)
	if err != nil {
		return fmt.Errorf("failed to find TTR: %w", err)
	}
	if ttr == nil {
		return errors.New("TTR not found")
	}

	playerCount, err := s.getPlayerCount(ttrID)
	if err != nil {
		return fmt.Errorf("failed to get player count: %w", err)
	}
	if playerCount >= ttr.MaxPlayers {
		return errors.New("TTR is full")
	}

	isAlreadyPlayer, err := s.ttrRepo.IsPlayer(ttrID, userID)
	if err != nil {
		return fmt.Errorf("failed to check player status: %w", err)
	}
	if isAlreadyPlayer {
		return errors.New("user is already a player")
	}

	if err := s.ttrRepo.AddPlayer(ttrID, userID, models.TTRPlayerStatusConfirmed); err != nil {
		return fmt.Errorf("failed to join TTR: %w", err)
	}

	return nil
}

func (s *TTRService) LeaveTTR(ttrID uuid.UUID, userID uuid.UUID) error {
	ttr, err := s.ttrRepo.FindByID(ttrID)
	if err != nil {
		return fmt.Errorf("failed to find TTR: %w", err)
	}
	if ttr == nil {
		return errors.New("TTR not found")
	}

	if ttr.CaptainUserID == userID {
		return errors.New("captain cannot leave TTR")
	}

	if err := s.ttrRepo.RemovePlayer(ttrID, userID); err != nil {
		return fmt.Errorf("failed to leave TTR: %w", err)
	}

	return nil
}

func (s *TTRService) UpdatePlayerStatus(ttrID uuid.UUID, managerUserID uuid.UUID, playerUserID uuid.UUID, status string) error {
	canManage, err := s.canManageTTR(ttrID, managerUserID)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !canManage {
		return errors.New("unauthorized: only captain or co-captain can update player status")
	}

	validStatuses := map[string]bool{
		models.TTRPlayerStatusConfirmed: true,
		models.TTRPlayerStatusMaybe:     true,
		models.TTRPlayerStatusDeclined:  true,
	}
	if !validStatuses[status] {
		return errors.New("invalid player status")
	}

	players, err := s.ttrRepo.GetPlayers(ttrID)
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	var found bool
	for _, player := range players {
		if player.UserID == playerUserID {
			found = true
			break
		}
	}

	if !found {
		return errors.New("player not found in TTR")
	}

	if err := s.ttrRepo.RemovePlayer(ttrID, playerUserID); err != nil {
		return fmt.Errorf("failed to remove player: %w", err)
	}

	if err := s.ttrRepo.AddPlayer(ttrID, playerUserID, status); err != nil {
		return fmt.Errorf("failed to add player with new status: %w", err)
	}

	return nil
}

func (s *TTRService) GetPlayers(ttrID uuid.UUID) ([]*models.TTRPlayer, error) {
	players, err := s.ttrRepo.GetPlayers(ttrID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}
	return players, nil
}

func (s *TTRService) isCaptain(ttrID uuid.UUID, userID uuid.UUID) (bool, error) {
	ttr, err := s.ttrRepo.FindByID(ttrID)
	if err != nil {
		return false, fmt.Errorf("failed to find TTR: %w", err)
	}
	if ttr == nil {
		return false, errors.New("TTR not found")
	}
	return ttr.CaptainUserID == userID, nil
}

func (s *TTRService) isCoCaptain(ttrID uuid.UUID, userID uuid.UUID) (bool, error) {
	return s.ttrRepo.IsCoCaptain(ttrID, userID)
}

func (s *TTRService) canManageTTR(ttrID uuid.UUID, userID uuid.UUID) (bool, error) {
	isCaptain, err := s.isCaptain(ttrID, userID)
	if err != nil {
		return false, err
	}
	if isCaptain {
		return true, nil
	}

	isCoCaptain, err := s.isCoCaptain(ttrID, userID)
	if err != nil {
		return false, err
	}
	return isCoCaptain, nil
}

func (s *TTRService) getPlayerCount(ttrID uuid.UUID) (int, error) {
	players, err := s.ttrRepo.GetPlayers(ttrID)
	if err != nil {
		return 0, fmt.Errorf("failed to get players: %w", err)
	}
	return len(players), nil
}
