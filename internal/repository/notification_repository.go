package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/golf_messenger/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notification *models.Notification) error
	FindByID(id uuid.UUID) (*models.Notification, error)
	FindByUserID(userID uuid.UUID, limit int, offset int) ([]*models.Notification, error)
	FindUnreadByUserID(userID uuid.UUID) ([]*models.Notification, error)
	MarkAsRead(id uuid.UUID) error
	MarkAllAsRead(userID uuid.UUID) error
	Delete(id uuid.UUID) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *models.Notification) error {
	if err := r.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

func (r *notificationRepository) FindByID(id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	if err := r.db.Where("id = ?", id).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}
	return &notification, nil
}

func (r *notificationRepository) FindByUserID(userID uuid.UUID, limit int, offset int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	if err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to find notifications: %w", err)
	}
	return notifications, nil
}

func (r *notificationRepository) FindUnreadByUserID(userID uuid.UUID) ([]*models.Notification, error) {
	var notifications []*models.Notification
	if err := r.db.
		Where("user_id = ? AND is_read = ?", userID, false).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to find unread notifications: %w", err)
	}
	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(id uuid.UUID) error {
	if err := r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error; err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(userID uuid.UUID) error {
	if err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error; err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

func (r *notificationRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&models.Notification{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	return nil
}
