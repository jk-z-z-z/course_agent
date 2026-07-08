export type AgentStatus = 'enabled' | 'disabled'
export type AgentRetrievalScope = 'course_all'
export type AgentSenderType = 'user' | 'agent'

export interface CourseAgentVO {
  id: number
  courseId: number
  agentName: string
  promptTemplate?: string
  status: AgentStatus
  retrievalScope: AgentRetrievalScope
  createdBy: number
  createdAt: string
  updatedAt: string
}

export interface UpdateCourseAgentPayload {
  agentName: string
  promptTemplate: string
  status: AgentStatus
  retrievalScope: AgentRetrievalScope
}

export interface AgentConversationVO {
  id: number
  courseId: number
  agentId: number
  userId: number
  conversationTitle: string
  createdAt: string
  updatedAt: string
}

export interface CreateAgentConversationPayload {
  title: string
}

export interface AgentMessageSourceVO {
  id?: number
  messageId?: number
  materialNodeId: number
  fileName?: string
  snippetText: string
  createdAt?: string
}

export interface AgentMessageVO {
  id: number
  conversationId: number
  senderType: AgentSenderType
  messageContent: string
  tokenUsage: number
  createdAt: string
  sources?: AgentMessageSourceVO[]
}

export interface AgentConversationDetailVO {
  conversation: AgentConversationVO
  messages: AgentMessageVO[]
}

export interface AskAgentPayload {
  conversationId: number
  question: string
}

export interface AgentAskResultVO {
  conversationId: number
  question: string
  answer: string
  sources: AgentMessageSourceVO[]
}

export interface CourseAgentStreamDeltaEvent {
  content: string
}

export interface CourseAgentStreamCompleteEvent {
  answer: string
  sources: AgentMessageSourceVO[]
  tokenUsage: number
}
