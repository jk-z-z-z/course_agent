package dto

type GenerateStudyPlanRequest struct {
	Goal         string `json:"goal"`
	DailyMinutes int    `json:"dailyMinutes"`
}

type UpdateStudyPlanItemRequest struct {
	Status string `json:"status"`
}
