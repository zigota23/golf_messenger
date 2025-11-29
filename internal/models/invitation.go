package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	InvitationStatusPending  = "PENDING"
	InvitationStatusYes      = "YES"
	InvitationStatusNo       = "NO"
	InvitationStatusMaybe    = "MAYBE"
	InvitationStatusCanceled = "CANCELED"
)

type Invitation struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	TTRID         uuid.UUID  `gorm:"type:uuid;not null" json:"ttr_id"`
	InviterUserID uuid.UUID  `gorm:"type:uuid;not null" json:"inviter_user_id"`
	InviteeUserID uuid.UUID  `gorm:"type:uuid;not null" json:"invitee_user_id"`
	Status        string     `gorm:"type:varchar(50);default:'PENDING'" json:"status"`
	Message       *string    `gorm:"type:text" json:"message,omitempty"`
	CreatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	RespondedAt   *time.Time `json:"responded_at,omitempty"`
	TTR           *TTR       `gorm:"foreignKey:TTRID" json:"ttr,omitempty"`
	InviterUser   *User      `gorm:"foreignKey:InviterUserID" json:"inviter_user,omitempty"`
	InviteeUser   *User      `gorm:"foreignKey:InviteeUserID" json:"invitee_user,omitempty"`
}

func (i *Invitation) TableName() string {
	return "invitations"
}
