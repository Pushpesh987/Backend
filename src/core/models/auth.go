package models

import (
	"time"
	"github.com/google/uuid"
)

// Auth model represents the authentication data of a user.
type Auth struct {
	ID              uuid.UUID `gorm:"column:id;type:uuid;primaryKey;not null" json:"id"`
	Username        string    `gorm:"column:username;type:text;not null;unique" json:"username"`
	Password        string    `gorm:"column:password;type:text;not null" json:"password"`
	Email           string    `gorm:"column:email;type:text;not null;unique" json:"email"`  // Added email field
	LastSignInAt    time.Time `gorm:"column:last_sign_in_at;type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"last_sign_in_at"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName sets the table name for the Auth model
func (Auth) TableName() string {
	return "auth"
}
