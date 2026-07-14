package vo

import "time"

type StudyPlanItemVO struct {
	ID               uint64    `json:"id"`
	PlanID           uint64    `json:"planId"`
	DayIndex         int       `json:"dayIndex"`
	PlanDate         time.Time `json:"planDate"`
	Title            string    `json:"title"`
	TasksText        string    `json:"tasksText"`
	SuggestedMinutes int       `json:"suggestedMinutes"`
	MaterialNodeIDs  []uint64  `json:"materialNodeIds"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type StudyPlanVO struct {
	ID               uint64            `json:"id"`
	CourseID         uint64            `json:"courseId"`
	UserID           uint64            `json:"userId"`
	Goal             string            `json:"goal"`
	DeadlineDate     time.Time         `json:"deadlineDate"`
	DailyMinutes     int               `json:"dailyMinutes"`
	Status           string            `json:"status"`
	GeneratedSummary string            `json:"generatedSummary"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	Items            []StudyPlanItemVO `json:"items"`
}

type StudyPlanSummaryVO struct {
	ID               uint64    `json:"id"`
	CourseID         uint64    `json:"courseId"`
	UserID           uint64    `json:"userId"`
	Goal             string    `json:"goal"`
	DeadlineDate     time.Time `json:"deadlineDate"`
	DailyMinutes     int       `json:"dailyMinutes"`
	Status           string    `json:"status"`
	GeneratedSummary string    `json:"generatedSummary"`
	ItemCount        int       `json:"itemCount"`
	DoneItemCount    int       `json:"doneItemCount"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
