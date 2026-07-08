package runtime

import (
	"context"
	"fmt"
	"strings"

	einoOpenAI "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	"course_agent_backend/internal/agent/retrieval"
	agenttypes "course_agent_backend/internal/agent/types"
)

type unavailableClient struct {
	reason string
}

func (c *unavailableClient) Ask(context.Context, agenttypes.AskRequest) (*agenttypes.AskResponse, error) {
	return nil, fmt.Errorf("agent unavailable: %s", c.reason)
}

type EinoClient struct {
	chatModel *einoOpenAI.ChatModel
}

func NewClient(ctx context.Context, cfg agenttypes.Config) (agenttypes.Client, error) {
	if strings.TrimSpace(cfg.APIKey) == "" || strings.TrimSpace(cfg.Model) == "" {
		return &unavailableClient{reason: "missing API key or model configuration"}, nil
	}
	chatModel, err := einoOpenAI.NewChatModel(ctx, &einoOpenAI.ChatModelConfig{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		BaseURL: cfg.BaseURL,
	})
	if err != nil {
		return nil, err
	}
	return &EinoClient{chatModel: chatModel}, nil
}

func (c *EinoClient) Ask(ctx context.Context, request agenttypes.AskRequest) (*agenttypes.AskResponse, error) {
	sources := retrieval.RetrieveSources(request.Question, request.Materials)
	messages := make([]*schema.Message, 0, 2+len(request.History))
	messages = append(messages, schema.SystemMessage(buildSystemPrompt(request, sources)))
	for _, turn := range request.History {
		switch turn.Role {
		case "agent":
			messages = append(messages, schema.AssistantMessage(turn.Content, nil))
		default:
			messages = append(messages, schema.UserMessage(turn.Content))
		}
	}
	messages = append(messages, schema.UserMessage(request.Question))

	resp, err := c.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}

	result := &agenttypes.AskResponse{
		Answer:  strings.TrimSpace(resp.Content),
		Sources: sources,
	}
	if resp.ResponseMeta != nil && resp.ResponseMeta.Usage != nil {
		result.TokenUsage = resp.ResponseMeta.Usage.TotalTokens
	}
	return result, nil
}

func buildSystemPrompt(request agenttypes.AskRequest, sources []agenttypes.Source) string {
	var builder strings.Builder
	builder.WriteString(strings.TrimSpace(request.PromptTemplate))
	builder.WriteString("\n\n")
	builder.WriteString("你当前扮演的课程 Agent 名称：")
	builder.WriteString(request.AgentName)
	builder.WriteString("。\n")
	builder.WriteString("你只能基于下方给出的课程资料片段回答。如果资料片段不足以支持答案，要明确说明资料不足，不要编造。\n")
	builder.WriteString("回答时尽量引用对应资料名称。\n\n")
	if len(sources) == 0 {
		builder.WriteString("当前没有可用于回答的课程资料片段。\n")
		return builder.String()
	}
	builder.WriteString("课程资料片段：\n")
	for idx, source := range sources {
		builder.WriteString(fmt.Sprintf("[%d] 文件：%s\n%s\n\n", idx+1, source.FileName, source.Snippet))
	}
	return builder.String()
}
