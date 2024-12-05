package models

import (
	"time"
	"github.com/google/uuid"
)

// Comment represents a comment on a post
type Comment struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	PostID    uuid.UUID `gorm:"column:post_id;type:uuid;not null" json:"post_id"`
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	Content   string    `gorm:"column:content;type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName sets the table name for the Comment model
func (Comment) TableName() string {
	return "comments"
}
