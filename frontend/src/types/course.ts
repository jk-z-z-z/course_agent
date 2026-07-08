export type CourseRole = 'owner' | 'teacher' | 'student'
export type CourseStatus = 'active' | 'archived' | 'deleted'
export type CourseJoinStatus = 'active' | 'removed'

export interface CourseVO {
  id: number
  courseCode: string
  courseName: string
  courseDescription: string
  ownerUserId: number
  status: CourseStatus
  createdAt: string
  updatedAt: string
  myRole?: CourseRole
}

export interface CourseMemberVO {
  id: number
  courseId: number
  userId: number
  username: string
  role: CourseRole
  joinStatus: CourseJoinStatus
  joinedAt: string
}

export interface CreateCoursePayload {
  courseCode: string
  courseName: string
  courseDescription: string
}

export interface UpdateCoursePayload {
  courseName: string
  courseDescription: string
}

export interface AddCourseMemberPayload {
  username: string
  role: Exclude<CourseRole, 'owner'>
}

export interface UpdateCourseMemberPayload {
  role: Exclude<CourseRole, 'owner'>
}
