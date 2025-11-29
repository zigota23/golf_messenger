package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/golf_messenger/internal/models"
	"gorm.io/gorm"
)

type InvitationRepository interface {
	Create(invitation *models.Invitation) error
	FindByID(id uuid.UUID) (*models.Invitation, error)
	FindReceivedByUserID(userID uuid.UUID) ([]*models.Invitation, error)
	FindSentByUserID(userID uuid.UUID) ([]*models.Invitation, error)
	Update(invitation *models.Invitation) error
	Delete(id uuid.UUID) error
	FindByTTRAndInvitee(ttrID uuid.UUID, inviteeUserID uuid.UUID) (*models.Invitation, error)
}

type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) Create(invitation *models.Invitation) error {
	if err := r.db.Create(invitation).Error; err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}
	return nil
}

func (r *invitationRepository) FindByID(id uuid.UUID) (*models.Invitation, error) {
	var invitation models.Invitation
	if err := r.db.
		Preload("TTR").
		Preload("TTR.CaptainUser").
		Preload("InviterUser").
		Preload("InviteeUser").
		Where("id = ?", id).
		First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find invitation by ID: %w", err)
	}
	return &invitation, nil
}

func (r *invitationRepository) FindReceivedByUserID(userID uuid.UUID) ([]*models.Invitation, error) {
	var invitations []*models.Invitation

	if err := r.db.
		Preload("TTR").
		Preload("TTR.CaptainUser").
		Preload("InviterUser").
		Preload("InviteeUser").
		Where("invitee_user_id = ?", userID).
		Order("created_at DESC").
		Find(&invitations).Error; err != nil {
		return nil, fmt.Errorf("failed to find received invitations: %w", err)
	}

	return invitations, nil
}

func (r *invitationRepository) FindSentByUserID(userID uuid.UUID) ([]*models.Invitation, error) {
	var invitations []*models.Invitation

	if err := r.db.
		Preload("TTR").
		Preload("TTR.CaptainUser").
		Preload("InviterUser").
		Preload("InviteeUser").
		Where("inviter_user_id = ?", userID).
		Order("created_at DESC").
		Find(&invitations).Error; err != nil {
		return nil, fmt.Errorf("failed to find sent invitations: %w", err)
	}

	return invitations, nil
}

func (r *invitationRepository) Update(invitation *models.Invitation) error {
	if err := r.db.Save(invitation).Error; err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}
	return nil
}

func (r *invitationRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&models.Invitation{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete invitation: %w", err)
	}
	return nil
}

func (r *invitationRepository) FindByTTRAndInvitee(ttrID uuid.UUID, inviteeUserID uuid.UUID) (*models.Invitation, error) {
	var invitation models.Invitation
	if err := r.db.
		Where("ttr_id = ? AND invitee_user_id = ?", ttrID, inviteeUserID).
		First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find invitation by TTR and invitee: %w", err)
	}
	return &invitation, nil
}
