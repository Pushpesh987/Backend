package models

import (
	"github.com/google/uuid"
	"time"
)

const TableNameUser = "users"

// User represents the user entity in the system.
type User struct {
	ID              uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	FirstName       string    `gorm:"column:first_name;type:text;not null" json:"first_name"`
	LastName        string    `gorm:"column:last_name;type:text;not null" json:"last_name"`
	Username        string    `gorm:"column:username;type:text;unique;not null" json:"username"`
	Email           string    `gorm:"column:email;type:text;unique;not null" json:"email"`
	Password        string    `gorm:"column:password;type:text;not null" json:"-"`
	Gender          string    `gorm:"column:gender;type:text;not null" json:"gender"`
	ProfilePhotoURL string    `gorm:"column:profile_photo_url;type:text" json:"profile_photo_url"` // New field for profile photo URL
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName returns the database table name for the User model.
func (*User) TableName() string {
	return TableNameUser
}
