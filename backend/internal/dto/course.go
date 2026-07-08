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

type AddCourseMemberRequest struct {
	UserIdentifier string `json:"userIdentifier"`
	Role           string `json:"role"`
}

type UpdateCourseMemberRequest struct {
	Role string `json:"role"`
}
