package agent

import (
	"context"

	runtimepkg "course_agent_backend/internal/agent/runtime"
	agenttypes "course_agent_backend/internal/agent/types"
)

type (
	Config           = agenttypes.Config
	MaterialDocument = agenttypes.MaterialDocument
	ConversationTurn = agenttypes.ConversationTurn
	Source           = agenttypes.Source
	AskRequest       = agenttypes.AskRequest
	AskResponse      = agenttypes.AskResponse
	Client           = agenttypes.Client
)

func NewClient(ctx context.Context, cfg Config) (Client, error) {
	return runtimepkg.NewClient(ctx, cfg)
}
