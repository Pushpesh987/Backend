package models

import (
	"time"

	"github.com/google/uuid"
)

// Connection represents a connection between two users
type Connection struct {
	ID           int       `gorm:"column:id;type:int;primaryKey;autoIncrement" json:"id"`
	UserID       uuid.UUID `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	ConnectionID uuid.UUID `gorm:"column:connection_id;type:uuid;not null" json:"connection_id"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName sets the table name for the Connection model
func (Connection) TableName() string {
	return "connections"
}
