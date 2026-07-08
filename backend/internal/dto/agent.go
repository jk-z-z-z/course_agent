package dto

type UpdateCourseAgentRequest struct {
	AgentName      string `json:"agentName"`
	PromptTemplate string `json:"promptTemplate"`
	Status         string `json:"status"`
	RetrievalScope string `json:"retrievalScope"`
}

type CreateConversationRequest struct {
	Title string `json:"title"`
}

type AskAgentRequest struct {
	ConversationID uint64 `json:"conversationId"`
	Question       string `json:"question"`
}
