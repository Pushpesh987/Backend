package models

import (
	"time"
)

// Community struct maps to the communities table
type Community struct {
	ID          int       `gorm:"column:id;type:serial;primaryKey" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (Community) TableName() string {
	return "communities"
}