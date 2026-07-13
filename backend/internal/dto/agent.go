package dto

type CreateConversationRequest struct {
	Title string `json:"title"`
}

type AskAgentRequest struct {
	ConversationID uint64 `json:"conversationId"`
	Question       string `json:"question"`
}
