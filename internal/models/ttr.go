package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TTRStatusOpen      = "OPEN"
	TTRStatusConfirmed = "CONFIRMED"
	TTRStatusCancelled = "CANCELLED"
	TTRStatusCompleted = "COMPLETED"
)

const (
	TTRPlayerStatusConfirmed = "CONFIRMED"
	TTRPlayerStatusMaybe     = "MAYBE"
	TTRPlayerStatusDeclined  = "DECLINED"
)

type TTR struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CourseName      string          `gorm:"type:varchar(255);not null" json:"course_name"`
	CourseLocation  *string         `gorm:"type:varchar(255)" json:"course_location,omitempty"`
	TeeDate         time.Time       `gorm:"type:date;not null" json:"tee_date"`
	TeeTime         time.Time       `gorm:"type:time;not null" json:"tee_time"`
	MaxPlayers      int             `gorm:"default:4" json:"max_players"`
	CreatedByUserID uuid.UUID       `gorm:"type:uuid;not null" json:"created_by_user_id"`
	CaptainUserID   uuid.UUID       `gorm:"type:uuid;not null" json:"captain_user_id"`
	Status          string          `gorm:"type:varchar(50);default:'OPEN'" json:"status"`
	Notes           *string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt       time.Time       `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
	CreatedByUser   *User           `gorm:"foreignKey:CreatedByUserID" json:"created_by_user,omitempty"`
	CaptainUser     *User           `gorm:"foreignKey:CaptainUserID" json:"captain_user,omitempty"`
	CoCaptains      []TTRCoCaptain  `gorm:"foreignKey:TTRID" json:"co_captains,omitempty"`
	Players         []TTRPlayer     `gorm:"foreignKey:TTRID" json:"players,omitempty"`
}

func (t *TTR) TableName() string {
	return "ttrs"
}

type TTRCoCaptain struct {
	TTRID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"ttr_id"`
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	AssignedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"assigned_at"`
	User       *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (t *TTRCoCaptain) TableName() string {
	return "ttr_co_captains"
}

type TTRPlayer struct {
	TTRID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"ttr_id"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	JoinedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	Status   string    `gorm:"type:varchar(50);default:'CONFIRMED'" json:"status"`
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (t *TTRPlayer) TableName() string {
	return "ttr_players"
}
