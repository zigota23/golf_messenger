package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	NotificationTypeInvitation      = "INVITATION"
	NotificationTypeTTRUpdate       = "TTR_UPDATE"
	NotificationTypeNewMessage      = "NEW_MESSAGE"
	NotificationTypeTTRCancelled    = "TTR_CANCELLED"
	NotificationTypePlayerJoined    = "PLAYER_JOINED"
	NotificationTypeCoCaptainAdded  = "CO_CAPTAIN_ADDED"
)

type Notification struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Type       string     `gorm:"type:varchar(100);not null" json:"type"`
	Title      string     `gorm:"type:varchar(255);not null" json:"title"`
	Message    string     `gorm:"type:text;not null" json:"message"`
	TargetType *string    `gorm:"type:varchar(50)" json:"target_type,omitempty"`
	TargetID   *uuid.UUID `gorm:"type:uuid" json:"target_id,omitempty"`
	IsRead     bool       `gorm:"default:false" json:"is_read"`
	CreatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
	User       *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (n *Notification) TableName() string {
	return "notifications"
}
