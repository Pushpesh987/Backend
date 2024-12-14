
package models

import (
	"time"

	"github.com/google/uuid"
)

// CommunityMember struct maps to the community_members table
type CommunityMember struct {
	ID          int       `gorm:"column:id;type:serial;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	CommunityID int       `gorm:"column:community_id;type:int;not null" json:"community_id"`
	JoinedAt    time.Time `gorm:"column:joined_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"joined_at"`
}

func (CommunityMember) TableName() string {
	return "community_members"
}
