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
	Date                 time.Time `json:"date" gorm:"type:date;not null"` // Changed to 'date' if time isn't necessary
	Location             string    `json:"location" gorm:"type:varchar(255)"`
	EntryFee             int       `json:"entry_fee" gorm:"type:int"`
	PrizePool            int       `json:"prize_pool" gorm:"type:int"`
	Media                string    `json:"media" gorm:"type:varchar(255)"`
	RegistrationDeadline time.Time `json:"registration_deadline" gorm:"type:date"` // Correct use of 'date' here
	OrganizerName        string    `json:"organizer_name" gorm:"type:varchar(255)"`
	OrganizerContact     string    `json:"organizer_contact" gorm:"type:varchar(50)"`
	Tags                 string    `json:"tags" gorm:"type:varchar(255)"`
	AttendeeCount        int       `json:"attendee_count" gorm:"type:int"`
	Status               string    `json:"status" gorm:"type:varchar(20);not null"`
}

type Workshop struct {
    ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID           uuid.UUID `json:"user_id" gorm:"type:uuid"`
    Title            string    `json:"title" gorm:"type:varchar(255);not null"`
    Description      string    `json:"description" gorm:"type:text"`
    Date             time.Time `json:"date" gorm:"type:date;not null"`
    Location         string    `json:"location" gorm:"type:varchar(255)"`
    Media            string    `json:"media" gorm:"type:text"`
    EntryFee         string    `json:"entry_fee" gorm:"type:decimal(10,2)"` // Represented as a string, could also be float64 if you need arithmetic
    Duration         string    `gorm:"column:duration""`
    InstructorInfo   string    `json:"instructor_info" gorm:"type:text"`
    Tags             string    `json:"tags" gorm:"type:varchar(255)"`
    ParticipantLimit int       `json:"participant_limit" gorm:"type:int"` // Changed to int
    Status           string    `json:"status" gorm:"type:workshop_status;not null"` // Enum values
    RegistrationLink string    `json:"registration_link" gorm:"type:text"`
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
