package models

import (
	"github.com/google/uuid"
	"time"
)

// Post struct represents a post in the system
type Post struct {
	ID            uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID        uuid.UUID  `json:"user_id"`
	Content       string     `json:"content"`
	MediaURL      string     `json:"media_url,omitempty"`
	LikesCount    int        `json:"likes_count,omitempty"`
	CommentsCount int        `json:"comments_count,omitempty"`
	CreatedAt     time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}
