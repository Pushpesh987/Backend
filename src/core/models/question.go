package models

import (
	"encoding/json"
	"time"
)

type Question struct {
	QuestionID    int             `gorm:"column:question_id;type:serial;primaryKey" json:"question_id"`
	QuestionText  string          `gorm:"column:question_text;type:text;not null" json:"question_text"`
	Options       json.RawMessage `gorm:"column:options;type:jsonb;not null" json:"options"` // Updated type
	CorrectAnswer string          `gorm:"column:correct_answer;type:text;not null" json:"correct_answer"`
	Difficulty    string          `gorm:"column:difficulty;type:difficulty_enum;not null" json:"difficulty"`
	Points        int             `gorm:"column:points;type:int;default:0;not null" json:"points"`
	Multiplier    float64         `gorm:"column:multiplier;type:float8;default:1.0;not null" json:"multiplier"`
	QuestionType  string          `gorm:"column:question_type;type:varchar(10);not null" json:"question_type"`
	CreatedAt     time.Time       `gorm:"column:created_at;type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName maps the struct to the questions table
func (Question) TableName() string {
	return "questions"
}
