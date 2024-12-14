package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID          int       `gorm:"column:id;type:serial;primaryKey" json:"id"`
	CommunityID int       `gorm:"column:community_id;type:int;not null" json:"community_id"`
	UserID      uuid.UUID `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	Message     string    `gorm:"column:message;type:text;not null" json:"message"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}
func (Message) TableName() string {
	return "messages"
}