import { API_BASE_URL, request } from './http'
import type {
  AgentAskResultVO,
  AgentConversationDetailVO,
  AgentConversationVO,
  AskAgentPayload,
  CourseAgentStreamCompleteEvent,
  CourseAgentStreamDeltaEvent,
  CourseAgentVO,
  CreateAgentConversationPayload,
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

interface StreamHandlers {
  onDelta?: (payload: CourseAgentStreamDeltaEvent) => void
  onComplete?: (payload: CourseAgentStreamCompleteEvent) => void
  onDone?: (payload: AgentAskResultVO) => void
}

export async function streamCourseAgent(
  token: string,
  courseId: number,
  payload: AskAgentPayload,
  handlers: StreamHandlers = {},
) {
  const response = await fetch(`${API_BASE_URL}/api/courses/${courseId}/agent/ask/stream`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(token),
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    const failure = await response.json().catch(() => null)
    throw new Error(failure?.message ?? '流式请求失败')
  }
  if (!response.body) {
    throw new Error('浏览器不支持流式响应')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { value, done } = await reader.read()
    buffer += decoder.decode(value ?? new Uint8Array(), { stream: !done })

    let boundary = buffer.indexOf('\n\n')
    while (boundary >= 0) {
      const block = buffer.slice(0, boundary)
      buffer = buffer.slice(boundary + 2)
      processSSEBlock(block, handlers)
      boundary = buffer.indexOf('\n\n')
    }

    if (done) {
      break
    }
  }
}

function processSSEBlock(block: string, handlers: StreamHandlers) {
  const lines = block.split('\n')
  let event = 'message'
  const dataLines: string[] = []

  for (const rawLine of lines) {
    const line = rawLine.trimEnd()
    if (line.startsWith('event:')) {
      event = line.slice(6).trim()
      continue
    }
    if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).trim())
    }
  }

  if (!dataLines.length) return
  const payload = JSON.parse(dataLines.join('\n'))

  switch (event) {
    case 'delta':
      handlers.onDelta?.(payload as CourseAgentStreamDeltaEvent)
      return
    case 'complete':
      handlers.onComplete?.(payload as CourseAgentStreamCompleteEvent)
      return
    case 'done':
      handlers.onDone?.(payload as AgentAskResultVO)
      return
    case 'error':
      throw new Error((payload as { message?: string }).message ?? '流式请求失败')
    default:
      return
  }
}
