package models

import (
	"time"
	"github.com/google/uuid"
)

type QuizAttempt struct {
	AttemptID   int       `gorm:"column:attempt_id;type:serial;primaryKey" json:"attempt_id"`
	UserID      uuid.UUID `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	QuestionID  int       `gorm:"column:question_id;type:int;not null" json:"question_id"`
	IsCorrect   bool      `gorm:"column:is_correct;type:boolean" json:"is_correct"`
	AttemptedAt time.Time `gorm:"column:attempted_at;type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"attempted_at"`
}

func (QuizAttempt) TableName() string {
	return "quiz_attempts"
}
