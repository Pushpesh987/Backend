package models

import "github.com/google/uuid"

type PostTag struct {
    ID     int       `gorm:"primaryKey"`
    PostID uuid.UUID `gorm:"column:post_id"`
    TagID  int       `gorm:"column:tag_id"`
}

