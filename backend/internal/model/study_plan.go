package model

import "time"

type StudyPlan struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"`
	CourseID         uint64    `gorm:"not null;index"`
	UserID           uint64    `gorm:"not null;index"`
	Goal             string    `gorm:"type:text;not null"`
	DeadlineDate     time.Time `gorm:"type:date;not null"`
	DailyMinutes     int       `gorm:"not null"`
	Status           string    `gorm:"size:16;not null;default:active;index"`
	GeneratedSummary string    `gorm:"type:text"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (StudyPlan) TableName() string {
	return "study_plans"
}

type StudyPlanItem struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"`
	PlanID           uint64    `gorm:"not null;index"`
	DayIndex         int       `gorm:"not null"`
	PlanDate         time.Time `gorm:"type:date;not null"`
	Title            string    `gorm:"size:255;not null"`
	TasksText        string    `gorm:"type:text;not null"`
	SuggestedMinutes int       `gorm:"not null"`
	MaterialNodeIDs  string    `gorm:"type:text"`
	Status           string    `gorm:"size:16;not null;default:pending;index"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (StudyPlanItem) TableName() string {
	return "study_plan_items"
}
