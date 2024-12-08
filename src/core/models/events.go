package models

import (
	"github.com/google/uuid"
	"time"
)

type Event struct {
	ID                   uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID               uuid.UUID `json:"user_id" gorm:"type:uuid"`
	Title                string    `json:"title" gorm:"type:varchar(255);not null"`
	Theme                string    `json:"theme" gorm:"type:varchar(255)"`
	Description          string    `json:"description" gorm:"type:text"`
	Date                 time.Time `json:"date" gorm:"type:timestamp;not null"`
	Location             string    `json:"location" gorm:"type:varchar(255)"`
	EntryFee             float64   `json:"entry_fee" gorm:"type:decimal(10,2)"`
	PrizePool            float64   `json:"prize_pool" gorm:"type:decimal(10,2)"`
	Media                string    `json:"media" gorm:"type:varchar(255)"`
	RegistrationDeadline time.Time `json:"registration_deadline" gorm:"type:date"`
	OrganizerName        string    `json:"organizer_name" gorm:"type:varchar(255)"`
	OrganizerContact     string    `json:"organizer_contact" gorm:"type:varchar(50)"`
	Tags                 string    `json:"tags" gorm:"type:varchar(255)"`
	AttendeeCount        int       `json:"attendee_count" gorm:"type:int"`
	Status               string    `json:"status" gorm:"type:varchar(20);not null"`
}
type Workshop struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID           uuid.UUID `json:"user_id" gorm:"type:uuid"`
	Title            string    `json:"title" gorm:"type:varchar(255);not null"`
	Description      string    `json:"description" gorm:"type:text"`
	Date             time.Time `json:"date" gorm:"type:timestamp;not null"`
	Location         string    `json:"location" gorm:"type:varchar(255)"`
	Media            string    `json:"media" gorm:"type:varchar(255)"`
	EntryFee         float64   `json:"entry_fee" gorm:"type:decimal(10,2)"`
	Duration         string    `gorm:"column:duration"`
	InstructorInfo   string    `json:"instructor_info" gorm:"type:varchar(255)"`
	Tags             string    `json:"tags" gorm:"type:varchar(255)"`
	ParticipantLimit int       `json:"participant_limit" gorm:"type:int"`
	Status           string    `json:"status" gorm:"type:varchar(20);not null"`
	RegistrationLink string    `json:"registration_link" gorm:"type:varchar(255)"`
}
type Project struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid"`
	Title       string    `json:"title" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Domain      string    `json:"domain" gorm:"type:varchar(255)"`
	StartDate   time.Time `json:"start_date" gorm:"type:date;not null"`
	EndDate     time.Time `json:"end_date" gorm:"type:date"`
	Location    string    `json:"location" gorm:"type:varchar(255)"`
	Media       string    `json:"media" gorm:"type:varchar(255)"`
	Tags        string    `json:"tags" gorm:"type:varchar(255)"`
	TeamMembers string    `json:"team_members" gorm:"type:text"`
	Status      string    `json:"status" gorm:"type:varchar(20);not null"`
	Sponsors    string    `json:"sponsors" gorm:"type:varchar(255)"`
	ProjectLink string    `json:"project_link" gorm:"type:varchar(255)"`
	Goals       string    `json:"goals" gorm:"type:text"`
}
