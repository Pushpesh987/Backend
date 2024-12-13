package models

import (
	"github.com/google/uuid"
)

type Skill struct {
	SkillID   uuid.UUID `gorm:"column:skill_id;type:uuid;primaryKey;not null" json:"skill_id"`
	SkillName string    `gorm:"column:skill_name;uniqueIndex;not null" json:"skill_name"`
}

func (Skill) TableName() string {
	return "skills"
}
