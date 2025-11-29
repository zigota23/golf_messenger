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

type InvitationService struct {
	invitationRepo      repository.InvitationRepository
	ttrRepo             repository.TTRRepository
	userRepo            repository.UserRepository
	notificationService *NotificationService
	logger              *zap.Logger
}

func NewInvitationService(
	invitationRepo repository.InvitationRepository,
	ttrRepo repository.TTRRepository,
	userRepo repository.UserRepository,
	notificationService *NotificationService,
	logger *zap.Logger,
) *InvitationService {
	return &InvitationService{
		invitationRepo:      invitationRepo,
		ttrRepo:             ttrRepo,
		userRepo:            userRepo,
		notificationService: notificationService,
		logger:              logger,
	}
}

func (s *InvitationService) CreateInvitation(ttrID uuid.UUID, inviterUserID uuid.UUID, inviteeUserID uuid.UUID, message *string) (*models.Invitation, error) {
	ttr, err := s.ttrRepo.FindByID(ttrID)
	if err != nil {
		return nil, fmt.Errorf("failed to find TTR: %w", err)
	}
	if ttr == nil {
		return nil, errors.New("TTR not found")
	}

	isCaptain := ttr.CaptainUserID == inviterUserID
	isCoCaptain, err := s.ttrRepo.IsCoCaptain(ttrID, inviterUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check co-captain status: %w", err)
	}

	if !isCaptain && !isCoCaptain {
		return nil, errors.New("unauthorized: only captain or co-captain can send invitations")
	}

	inviteeUser, err := s.userRepo.FindByID(inviteeUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find invitee user: %w", err)
	}
	if inviteeUser == nil {
		return nil, errors.New("invitee user not found")
	}

	players, err := s.ttrRepo.GetPlayers(ttrID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}
	if len(players) >= ttr.MaxPlayers {
		return nil, errors.New("TTR is full")
	}

	isAlreadyPlayer, err := s.ttrRepo.IsPlayer(ttrID, inviteeUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check player status: %w", err)
	}
	if isAlreadyPlayer {
		return nil, errors.New("invitee is already a player in this TTR")
	}

	existingInvitation, err := s.invitationRepo.FindByTTRAndInvitee(ttrID, inviteeUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing invitation: %w", err)
	}
	if existingInvitation != nil && existingInvitation.Status == models.InvitationStatusPending {
		return nil, errors.New("pending invitation already exists for this user")
	}

	invitation := &models.Invitation{
		TTRID:         ttrID,
		InviterUserID: inviterUserID,
		InviteeUserID: inviteeUserID,
		Status:        models.InvitationStatusPending,
		Message:       message,
	}

	if err := s.invitationRepo.Create(invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	targetType := "invitation"
	notifTitle := "New TTR Invitation"
	notifMessage := fmt.Sprintf("You have been invited to join a tee time at %s", ttr.CourseName)
	if err := s.notificationService.CreateNotification(inviteeUserID, "invitation_received", notifTitle, notifMessage, &targetType, &invitation.ID); err != nil {
		s.logger.Error("Failed to create notification", zap.Error(err))
	}

	createdInvitation, err := s.invitationRepo.FindByID(invitation.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created invitation: %w", err)
	}

	return createdInvitation, nil
}

func (s *InvitationService) RespondToInvitation(invitationID uuid.UUID, inviteeUserID uuid.UUID, status string) (*models.Invitation, error) {
	validStatuses := map[string]bool{
		models.InvitationStatusYes:   true,
		models.InvitationStatusNo:    true,
		models.InvitationStatusMaybe: true,
	}
	if !validStatuses[status] {
		return nil, errors.New("invalid invitation status")
	}

	invitation, err := s.invitationRepo.FindByID(invitationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find invitation: %w", err)
	}
	if invitation == nil {
		return nil, errors.New("invitation not found")
	}

	if invitation.InviteeUserID != inviteeUserID {
		return nil, errors.New("unauthorized: you can only respond to your own invitations")
	}

	if invitation.Status != models.InvitationStatusPending {
		return nil, errors.New("invitation has already been responded to")
	}

	now := time.Now()
	invitation.Status = status
	invitation.RespondedAt = &now

	if status == models.InvitationStatusYes {
		ttr, err := s.ttrRepo.FindByID(invitation.TTRID)
		if err != nil {
			return nil, fmt.Errorf("failed to find TTR: %w", err)
		}
		if ttr == nil {
			return nil, errors.New("TTR not found")
		}

		players, err := s.ttrRepo.GetPlayers(invitation.TTRID)
		if err != nil {
			return nil, fmt.Errorf("failed to get players: %w", err)
		}
		if len(players) >= ttr.MaxPlayers {
			return nil, errors.New("TTR is full, cannot accept invitation")
		}

		if err := s.ttrRepo.AddPlayer(invitation.TTRID, inviteeUserID, models.TTRPlayerStatusConfirmed); err != nil {
			return nil, fmt.Errorf("failed to add player to TTR: %w", err)
		}
	}

	if err := s.invitationRepo.Update(invitation); err != nil {
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	updatedInvitation, err := s.invitationRepo.FindByID(invitationID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated invitation: %w", err)
	}

	return updatedInvitation, nil
}

func (s *InvitationService) GetInvitation(id uuid.UUID) (*models.Invitation, error) {
	invitation, err := s.invitationRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}
	if invitation == nil {
		return nil, errors.New("invitation not found")
	}
	return invitation, nil
}

func (s *InvitationService) GetUserInvitations(userID uuid.UUID, received bool) ([]*models.Invitation, error) {
	var invitations []*models.Invitation
	var err error

	if received {
		invitations, err = s.invitationRepo.FindReceivedByUserID(userID)
	} else {
		invitations, err = s.invitationRepo.FindSentByUserID(userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user invitations: %w", err)
	}

	return invitations, nil
}

func (s *InvitationService) CancelInvitation(invitationID uuid.UUID, userID uuid.UUID) error {
	invitation, err := s.invitationRepo.FindByID(invitationID)
	if err != nil {
		return fmt.Errorf("failed to find invitation: %w", err)
	}
	if invitation == nil {
		return errors.New("invitation not found")
	}

	if invitation.InviterUserID != userID {
		return errors.New("unauthorized: only the inviter can cancel the invitation")
	}

	if invitation.Status != models.InvitationStatusPending {
		return errors.New("only pending invitations can be canceled")
	}

	invitation.Status = models.InvitationStatusCanceled

	if err := s.invitationRepo.Update(invitation); err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	return nil
}
