package vo

import "time"

type CourseVO struct {
	ID                uint64    `json:"id"`
	CourseCode        string    `json:"courseCode"`
	CourseName        string    `json:"courseName"`
	CourseDescription string    `json:"courseDescription"`
	OwnerUserID       uint64    `json:"ownerUserId"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	MyRole            string    `json:"myRole,omitempty"`
}

type CourseMemberVO struct {
	ID         uint64    `json:"id"`
	CourseID   uint64    `json:"courseId"`
	UserID     uint64    `json:"userId"`
	Role       string    `json:"role"`
	JoinStatus string    `json:"joinStatus"`
	JoinedAt   time.Time `json:"joinedAt"`
}
