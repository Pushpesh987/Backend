package models

import (
	"github.com/google/uuid"
)

type Location struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;primaryKey;not null" json:"id"`
	LocationName string    `gorm:"column:name;uniqueIndex;not null" json:"location_name"`
}

func (Location) TableName() string {
	return "locations"
}

