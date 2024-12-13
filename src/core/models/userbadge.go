package models

import "github.com/google/uuid"

type UserBadge struct {
	UserID  uuid.UUID `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	BadgeID int       `gorm:"column:badge_id;type:int;not null" json:"badge_id"`
}

func (UserBadge) TableName() string {
	return "user_badges"
}
