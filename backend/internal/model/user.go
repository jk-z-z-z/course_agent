package model

import "time"

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"size:64;not null;uniqueIndex"`
	PasswordHash string    `gorm:"size:255;not null"`
	Phone        string    `gorm:"size:32"`
	Status       string    `gorm:"size:16;not null;default:active"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string {
	return "users"
}
