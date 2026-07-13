import { request } from './http'
import type {
  AddCourseMemberPayload,
  CourseMemberVO,
  CourseVO,
  CreateCoursePayload,
  UpdateCourseMemberPayload,
  UpdateCoursePayload,
} from '@/types/course'

function authHeaders(token: string) {
  return {
    Authorization: `Bearer ${token}`,
  }
}

export function listCourses(token: string) {
  return request<CourseVO[]>('/api/courses', {
    headers: authHeaders(token),
  })
}

export function listDiscoverableCourses(token: string) {
  return request<CourseVO[]>('/api/courses/discover', {
    headers: authHeaders(token),
  })
}

export function getCourse(token: string, courseId: number) {
  return request<CourseVO>(`/api/courses/${courseId}`, {
    headers: authHeaders(token),
  })
}

export function createCourse(token: string, payload: CreateCoursePayload) {
  return request<CourseVO>('/api/courses', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  })
}

export function updateCourse(token: string, courseId: number, payload: UpdateCoursePayload) {
  return request<CourseVO>(`/api/courses/${courseId}`, {
    method: 'PUT',
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  })
}

export function deleteCourse(token: string, courseId: number) {
  return request<null>(`/api/courses/${courseId}`, {
    method: 'DELETE',
    headers: authHeaders(token),
  })
}

export function joinCourse(token: string, courseId: number) {
  return request<CourseMemberVO>(`/api/courses/${courseId}/join`, {
    method: 'POST',
    headers: authHeaders(token),
  })
}

export function listCourseMembers(token: string, courseId: number) {
  return request<CourseMemberVO[]>(`/api/courses/${courseId}/members`, {
    headers: authHeaders(token),
  })
}

export function addCourseMember(token: string, courseId: number, payload: AddCourseMemberPayload) {
  return request<CourseMemberVO>(`/api/courses/${courseId}/members`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  })
}

export function updateCourseMember(token: string, courseId: number, memberId: number, payload: UpdateCourseMemberPayload) {
  return request<CourseMemberVO>(`/api/courses/${courseId}/members/${memberId}`, {
    method: 'PUT',
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  })
}

export function deleteCourseMember(token: string, courseId: number, memberId: number) {
  return request<null>(`/api/courses/${courseId}/members/${memberId}`, {
    method: 'DELETE',
    headers: authHeaders(token),
  })
}
