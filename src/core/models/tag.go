package models

type Tag struct {
    ID   int    `gorm:"primaryKey;autoIncrement" json:"id"`
    Tag  string `gorm:"unique;not null" json:"tag"`
}
