package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/golf_messenger/internal/models"
	"gorm.io/gorm"
)

type TTRRepository interface {
	Create(ttr *models.TTR) error
	FindByID(id uuid.UUID) (*models.TTR, error)
	FindAll(limit int, offset int, status string) ([]*models.TTR, error)
	Update(ttr *models.TTR) error
	Delete(id uuid.UUID) error
	FindUpcomingByUserID(userID uuid.UUID) ([]*models.TTR, error)
	FindPastByUserID(userID uuid.UUID) ([]*models.TTR, error)
	AddCoCaptain(ttrID uuid.UUID, userID uuid.UUID) error
	RemoveCoCaptain(ttrID uuid.UUID, userID uuid.UUID) error
	IsCoCaptain(ttrID uuid.UUID, userID uuid.UUID) (bool, error)
	AddPlayer(ttrID uuid.UUID, userID uuid.UUID, status string) error
	RemovePlayer(ttrID uuid.UUID, userID uuid.UUID) error
	GetPlayers(ttrID uuid.UUID) ([]*models.TTRPlayer, error)
	IsPlayer(ttrID uuid.UUID, userID uuid.UUID) (bool, error)
}

type ttrRepository struct {
	db *gorm.DB
}

func NewTTRRepository(db *gorm.DB) TTRRepository {
	return &ttrRepository{db: db}
}

func (r *ttrRepository) Create(ttr *models.TTR) error {
	if err := r.db.Create(ttr).Error; err != nil {
		return fmt.Errorf("failed to create ttr: %w", err)
	}
	return nil
}

func (r *ttrRepository) FindByID(id uuid.UUID) (*models.TTR, error) {
	var ttr models.TTR
	if err := r.db.
		Preload("CreatedByUser").
		Preload("CaptainUser").
		Preload("CoCaptains.User").
		Preload("Players.User").
		Where("id = ?", id).
		First(&ttr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find ttr by ID: %w", err)
	}
	return &ttr, nil
}

func (r *ttrRepository) FindAll(limit int, offset int, status string) ([]*models.TTR, error) {
	var ttrs []*models.TTR
	query := r.db.
		Preload("CreatedByUser").
		Preload("CaptainUser").
		Preload("CoCaptains.User").
		Preload("Players.User")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.
		Limit(limit).
		Offset(offset).
		Order("tee_date ASC, tee_time ASC").
		Find(&ttrs).Error; err != nil {
		return nil, fmt.Errorf("failed to find all ttrs: %w", err)
	}

	return ttrs, nil
}

func (r *ttrRepository) Update(ttr *models.TTR) error {
	if err := r.db.Save(ttr).Error; err != nil {
		return fmt.Errorf("failed to update ttr: %w", err)
	}
	return nil
}

func (r *ttrRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&models.TTR{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete ttr: %w", err)
	}
	return nil
}

func (r *ttrRepository) FindUpcomingByUserID(userID uuid.UUID) ([]*models.TTR, error) {
	var ttrs []*models.TTR

	now := time.Now()

	if err := r.db.
		Preload("CreatedByUser").
		Preload("CaptainUser").
		Preload("CoCaptains.User").
		Preload("Players.User").
		Joins("LEFT JOIN ttr_players ON ttrs.id = ttr_players.ttr_id").
		Joins("LEFT JOIN ttr_co_captains ON ttrs.id = ttr_co_captains.ttr_id").
		Where("ttrs.tee_date >= ? AND (ttrs.captain_user_id = ? OR ttr_players.user_id = ? OR ttr_co_captains.user_id = ?)",
			now, userID, userID, userID).
		Group("ttrs.id").
		Order("ttrs.tee_date ASC, ttrs.tee_time ASC").
		Find(&ttrs).Error; err != nil {
		return nil, fmt.Errorf("failed to find upcoming ttrs: %w", err)
	}

	return ttrs, nil
}

func (r *ttrRepository) FindPastByUserID(userID uuid.UUID) ([]*models.TTR, error) {
	var ttrs []*models.TTR

	now := time.Now()

	if err := r.db.
		Preload("CreatedByUser").
		Preload("CaptainUser").
		Preload("CoCaptains.User").
		Preload("Players.User").
		Joins("LEFT JOIN ttr_players ON ttrs.id = ttr_players.ttr_id").
		Joins("LEFT JOIN ttr_co_captains ON ttrs.id = ttr_co_captains.ttr_id").
		Where("ttrs.tee_date < ? AND (ttrs.captain_user_id = ? OR ttr_players.user_id = ? OR ttr_co_captains.user_id = ?)",
			now, userID, userID, userID).
		Group("ttrs.id").
		Order("ttrs.tee_date DESC, ttrs.tee_time DESC").
		Find(&ttrs).Error; err != nil {
		return nil, fmt.Errorf("failed to find past ttrs: %w", err)
	}

	return ttrs, nil
}

func (r *ttrRepository) AddCoCaptain(ttrID uuid.UUID, userID uuid.UUID) error {
	coCaptain := &models.TTRCoCaptain{
		TTRID:  ttrID,
		UserID: userID,
	}

	if err := r.db.Create(coCaptain).Error; err != nil {
		return fmt.Errorf("failed to add co-captain: %w", err)
	}

	return nil
}

func (r *ttrRepository) RemoveCoCaptain(ttrID uuid.UUID, userID uuid.UUID) error {
	if err := r.db.
		Where("ttr_id = ? AND user_id = ?", ttrID, userID).
		Delete(&models.TTRCoCaptain{}).Error; err != nil {
		return fmt.Errorf("failed to remove co-captain: %w", err)
	}

	return nil
}

func (r *ttrRepository) IsCoCaptain(ttrID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.TTRCoCaptain{}).
		Where("ttr_id = ? AND user_id = ?", ttrID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check co-captain status: %w", err)
	}

	return count > 0, nil
}

func (r *ttrRepository) AddPlayer(ttrID uuid.UUID, userID uuid.UUID, status string) error {
	player := &models.TTRPlayer{
		TTRID:  ttrID,
		UserID: userID,
		Status: status,
	}

	if err := r.db.Create(player).Error; err != nil {
		return fmt.Errorf("failed to add player: %w", err)
	}

	return nil
}

func (r *ttrRepository) RemovePlayer(ttrID uuid.UUID, userID uuid.UUID) error {
	if err := r.db.
		Where("ttr_id = ? AND user_id = ?", ttrID, userID).
		Delete(&models.TTRPlayer{}).Error; err != nil {
		return fmt.Errorf("failed to remove player: %w", err)
	}

	return nil
}

func (r *ttrRepository) GetPlayers(ttrID uuid.UUID) ([]*models.TTRPlayer, error) {
	var players []*models.TTRPlayer

	if err := r.db.
		Preload("User").
		Where("ttr_id = ?", ttrID).
		Find(&players).Error; err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	return players, nil
}

func (r *ttrRepository) IsPlayer(ttrID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.TTRPlayer{}).
		Where("ttr_id = ? AND user_id = ?", ttrID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check player status: %w", err)
	}

	return count > 0, nil
}
