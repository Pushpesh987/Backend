package models

import "github.com/google/uuid"

// Like represents a like on a post
type Like struct {
	UserID uuid.UUID `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	PostID uuid.UUID `gorm:"column:post_id;type:uuid;not null" json:"post_id"`
}

// TableName sets the table name for the Like model
func (Like) TableName() string {
	return "likes"
}