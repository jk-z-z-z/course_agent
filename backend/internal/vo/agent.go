package vo

import "time"

type CourseAgentVO struct {
	ID             uint64    `json:"id"`
	CourseID       uint64    `json:"courseId"`
	AgentName      string    `json:"agentName"`
	PromptTemplate string    `json:"promptTemplate,omitempty"`
	Status         string    `json:"status"`
	RetrievalScope string    `json:"retrievalScope"`
	CreatedBy      uint64    `json:"createdBy"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type AgentConversationVO struct {
	ID                uint64    `json:"id"`
	CourseID          uint64    `json:"courseId"`
	AgentID           uint64    `json:"agentId"`
	UserID            uint64    `json:"userId"`
	ConversationTitle string    `json:"conversationTitle"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type AgentMessageSourceVO struct {
	ID             uint64    `json:"id"`
	MessageID      uint64    `json:"messageId"`
	MaterialNodeID uint64    `json:"materialNodeId"`
	FileName       string    `json:"fileName,omitempty"`
	SnippetText    string    `json:"snippetText"`
	CreatedAt      time.Time `json:"createdAt"`
}

type AgentMessageVO struct {
	ID             uint64                 `json:"id"`
	ConversationID uint64                 `json:"conversationId"`
	SenderType     string                 `json:"senderType"`
	MessageContent string                 `json:"messageContent"`
	TokenUsage     int                    `json:"tokenUsage"`
	CreatedAt      time.Time              `json:"createdAt"`
	Sources        []AgentMessageSourceVO `json:"sources,omitempty"`
}

type AgentConversationDetailVO struct {
	Conversation AgentConversationVO `json:"conversation"`
	Messages     []AgentMessageVO    `json:"messages"`
}

type AgentAskResultVO struct {
	ConversationID uint64                 `json:"conversationId"`
	Question       string                 `json:"question"`
	Answer         string                 `json:"answer"`
	Sources        []AgentMessageSourceVO `json:"sources"`
}
