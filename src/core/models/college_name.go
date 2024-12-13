package models

import (
	"github.com/google/uuid"
)

type College struct {
	ID          uuid.UUID `gorm:"column:id;type:uuid;primaryKey;not null" json:"id"`
	CollegeName string    `gorm:"column:college_name;type:text;uniqueIndex;not null" json:"college_name"`
}

func (College) TableName() string {
	return "colleges"
}
