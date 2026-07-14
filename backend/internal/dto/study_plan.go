package dto

type GenerateStudyPlanRequest struct {
	Goal         string `json:"goal"`
	DeadlineDate string `json:"deadlineDate"`
	DailyMinutes int    `json:"dailyMinutes"`
}

type UpdateStudyPlanItemRequest struct {
	Status string `json:"status"`
}
