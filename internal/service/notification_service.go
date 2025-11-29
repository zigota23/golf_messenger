package service

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NotificationService struct {
	logger *zap.Logger
}

func NewNotificationService(logger *zap.Logger) *NotificationService {
	return &NotificationService{
		logger: logger,
	}
}

func (s *NotificationService) CreateNotification(userID uuid.UUID, notificationType string, title string, message string, targetType *string, targetID *uuid.UUID) error {
	s.logger.Info("Notification stub called",
		zap.String("user_id", userID.String()),
		zap.String("type", notificationType),
		zap.String("title", title),
		zap.String("message", message),
	)
	return nil
}
