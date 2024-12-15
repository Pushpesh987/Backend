package models

import (
	"time"
	"github.com/google/uuid"
)

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"` // Recipient of the notification
	Message   string    `gorm:"type:text;not null" json:"message"`  // Notification text
	Category  string    `gorm:"type:varchar(50);not null" json:"category"` // e.g., "connection", "iot", etc.
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"` // Timestamp of when the notification was created
	IsRead    bool      `gorm:"default:false" json:"is_read"` // Whether the notification has been read by the user
}

type NotificationTemplate struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Category    string `gorm:"type:varchar(50);not null" json:"category"`  // Category of the notification (e.g., "connection")
	TemplateText string `gorm:"type:text;not null" json:"template_text"` // Cryptic message template (e.g., "{user1} should connect with {user2}")
}
