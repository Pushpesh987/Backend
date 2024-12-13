package models

import (
	"github.com/google/uuid"
)

type EducationLevel struct {
	ID   uuid.UUID `gorm:"column:id;type:uuid;primaryKey;not null" json:"id"`
	Name string    `gorm:"column:level_name;uniqueIndex;not null" json:"name"`
}

func (EducationLevel) TableName() string {
	return "education_levels"
}
