package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)
type TextArray []string

func (ta *TextArray) Scan(value interface{}) error {
    return json.Unmarshal(value.([]byte), &ta)
}

func (ta TextArray) Value() (driver.Value, error) {
    return json.Marshal(ta)
}
// Post struct represents a post in the system
type Post struct {
	ID            uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID        uuid.UUID  `json:"user_id"`
	Content       string     `json:"content"`
	MediaURL      string     `json:"media_url,omitempty"`
	LikesCount    int        `json:"likes_count,omitempty"`
	CommentsCount int        `json:"comments_count,omitempty"`
	Tags          string  `json:"tags" gorm:"type:text[]"`
	CreatedAt     time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}



