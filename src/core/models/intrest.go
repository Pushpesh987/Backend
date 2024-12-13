package models

import (
	"github.com/google/uuid"
)

type Interest struct {
	InterestID   uuid.UUID `gorm:"column:interest_id;type:uuid;primaryKey;not null" json:"interest_id"`
	InterestName string    `gorm:"column:interest_name;uniqueIndex;not null" json:"interest_name"`
}

func (Interest) TableName() string {
	return "interests"
}
