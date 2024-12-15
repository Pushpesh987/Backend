package models

import (
	"time"

	"github.com/google/uuid"
)

type IotLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"` // Automatically set by GORM
	Location  string    `gorm:"type:varchar(255)" json:"location,omitempty"`
}

