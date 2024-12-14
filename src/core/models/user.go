package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                      uuid.UUID `gorm:"column:id;type:uuid;primaryKey;not null" json:"id"`
	FirstName               string    `gorm:"column:first_name;type:text;not null" json:"first_name"`
	LastName                string    `gorm:"column:last_name;type:text;not null" json:"last_name"`
	Username                string    `gorm:"column:username;type:text;unique;not null" json:"username"`
	ProfilePhotoURL         string    `gorm:"column:profile_pic_url;type:text;not null" json:"profile_photo_url"`
	ProfilePhotoSize        int       `gorm:"column:profile_pic_size;type:int;not null;default:0" json:"profile_photo_size"`
	ProfilePhotoStoragePath string    `gorm:"column:profile_pic_storage_path;type:text;not null" json:"profile_photo_storage_path"`
	LocationID              uuid.UUID `gorm:"column:location_id;type:uuid;not null" json:"location_id"`
	EducationLevel          string    `gorm:"-" json:"education_level"`
	EducationLevelID        uuid.UUID `gorm:"column:education_level_id;type:uuid;not null" json:"education_level_id"`
	FieldOfStudyID          uuid.UUID `gorm:"column:field_of_study_id;type:uuid;not null" json:"field_of_study_id"`
	CollegeNameID           uuid.UUID `gorm:"column:college_name_id;type:uuid;not null" json:"college_name_id"`
	Age                     int       `gorm:"column:age;type:int;not null;default:0" json:"age"`
	Dob                     time.Time `gorm:"column:dob;type:date;not null" json:"dob"`
	Gender                  string    `gorm:"column:gender;type:text;not null;default:''" json:"gender"`
	Phone                   string    `gorm:"column:phone;type:text;unique;not null" json:"phone"`
	Email                   string    `gorm:"column:email;type:text;unique;not null" json:"email"`
	AuthID                  uuid.UUID `gorm:"column:auth_id;type:uuid;unique" json:"auth_id"`
	CreatedAt               time.Time `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
