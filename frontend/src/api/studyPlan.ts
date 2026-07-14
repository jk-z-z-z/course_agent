import { request } from './http'
import type { GenerateStudyPlanPayload, StudyPlanItemStatus, StudyPlanSummaryVO, StudyPlanVO } from '@/types/studyPlan'

function authHeaders(token: string) {
  return {
    Authorization: `Bearer ${token}`,
  }
}

export function listStudyPlans(token: string, courseId: number) {
  return request<StudyPlanSummaryVO[]>(`/api/courses/${courseId}/study-plans`, {
    headers: authHeaders(token),
  })
}

export function getStudyPlan(token: string, courseId: number, planId: number) {
  return request<StudyPlanVO>(`/api/courses/${courseId}/study-plans/${planId}`, {
    headers: authHeaders(token),
  })
}

export function generateStudyPlan(token: string, courseId: number, payload: GenerateStudyPlanPayload) {
  return request<StudyPlanVO>(`/api/courses/${courseId}/study-plans/generate`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  })
}

export function updateStudyPlanItemStatus(
  token: string,
  courseId: number,
  planId: number,
  itemId: number,
  status: StudyPlanItemStatus,
) {
  return request<StudyPlanVO>(`/api/courses/${courseId}/study-plans/${planId}/items/${itemId}`, {
    method: 'PATCH',
    headers: authHeaders(token),
    body: JSON.stringify({ status }),
  })
}
