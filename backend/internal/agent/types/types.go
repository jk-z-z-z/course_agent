package types

import "context"

type Config struct {
	BaseURL string
	APIKey  string
	Model   string
}

type MaterialDocument struct {
	MaterialNodeID uint64
	FileName       string
	StoragePath    string
	MimeType       string
}

type ConversationTurn struct {
	Role    string
	Content string
}

type Source struct {
	MaterialNodeID uint64
	FileName       string
	Snippet        string
	VersionID      *uint64
}

type AskRequest struct {
	AgentName      string
	PromptTemplate string
	Question       string
	History        []ConversationTurn
	Materials      []MaterialDocument
}

type AskResponse struct {
	Answer     string
	Sources    []Source
	TokenUsage int
}

type Client interface {
	Ask(ctx context.Context, request AskRequest) (*AskResponse, error)
}
