package models

import "time"
import "github.com/google/uuid"

// Share represents a user's post share with an origin and recipient user.
type Share struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	FromUserID uuid.UUID `json:"from_user_id" gorm:"type:uuid;not null"`
	ToUserID   uuid.UUID `json:"to_user_id" gorm:"type:uuid;not null"`
	PostID     uuid.UUID `json:"post_id" gorm:"type:uuid;not null"`
	SharedAt   time.Time `json:"shared_at" gorm:"autoCreateTime"`
}
