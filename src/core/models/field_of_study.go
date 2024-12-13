package models

import (
	"github.com/google/uuid"
)

type FieldOfStudy struct {
	ID   uuid.UUID `gorm:"column:id;type:uuid;primaryKey;not null" json:"id"`
	Name string    `gorm:"column:field_name;uniqueIndex;not null" json:"name"`
}

func (FieldOfStudy) TableName() string {
	return "fields_of_study"
}
