package dto

type CreateCourseRequest struct {
	CourseCode        string `json:"courseCode"`
	CourseName        string `json:"courseName"`
	CourseDescription string `json:"courseDescription"`
}

type UpdateCourseRequest struct {
	CourseName        string `json:"courseName"`
	CourseDescription string `json:"courseDescription"`
}
