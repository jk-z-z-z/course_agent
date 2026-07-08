import { request } from './http'
import type {
  AgentAskResultVO,
  AgentConversationDetailVO,
  AgentConversationVO,
  AskAgentPayload,
  CourseAgentVO,
  CreateAgentConversationPayload,
  UpdateCourseAgentPayload,
} from '@/types/agent'

function authHeaders(token: string) {
  return {
    Authorization: `Bearer ${token}`,
  }
}

export function getCourseAgent(token: string, courseId: number) {
  return request<CourseAgentVO>(`/api/courses/${courseId}/agent`, {
    headers: authHeaders(token),
  })
}

export function updateCourseAgent(token: string, courseId: number, payload: UpdateCourseAgentPayload) {
  return request<CourseAgentVO>(`/api/courses/${courseId}/agent`, {
    method: 'PUT',
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  })
}

export function listAgentConversations(token: string, courseId: number) {
  return request<AgentConversationVO[]>(`/api/courses/${courseId}/agent/conversations`, {
    headers: authHeaders(token),
  })
}

export function createAgentConversation(token: string, courseId: number, payload: CreateAgentConversationPayload) {
  return request<AgentConversationVO>(`/api/courses/${courseId}/agent/conversations`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  })
}

export function getAgentConversation(token: string, courseId: number, conversationId: number) {
  return request<AgentConversationDetailVO>(`/api/courses/${courseId}/agent/conversations/${conversationId}`, {
    headers: authHeaders(token),
  })
}

export function askCourseAgent(token: string, courseId: number, payload: AskAgentPayload) {
  return request<AgentAskResultVO>(`/api/courses/${courseId}/agent/ask`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  })
}
