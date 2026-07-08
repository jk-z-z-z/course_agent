package model

import "time"

type Course struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement"`
	CourseCode         string    `gorm:"size:64;not null;uniqueIndex"`
	CourseName         string    `gorm:"size:128;not null"`
	CourseDescription  string    `gorm:"type:text"`
	OwnerUserID        uint64    `gorm:"not null;index"`
	Status             string    `gorm:"size:16;not null;default:active"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (Course) TableName() string {
	return "courses"
}

type CourseMember struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	CourseID   uint64    `gorm:"not null;uniqueIndex:idx_course_user"`
	UserID     uint64    `gorm:"not null;uniqueIndex:idx_course_user"`
	Role       string    `gorm:"size:16;not null"`
	JoinStatus string    `gorm:"size:16;not null;default:active"`
	JoinedAt   time.Time `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (CourseMember) TableName() string {
	return "course_members"
}
