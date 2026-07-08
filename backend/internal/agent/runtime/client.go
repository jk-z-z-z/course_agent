package runtime

import (
	"context"
	"fmt"
	"io"
	"strings"

	einoOpenAI "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	"course_agent_backend/internal/agent/tools"
	agenttypes "course_agent_backend/internal/agent/types"
)

type unavailableClient struct {
	reason string
}

func (c *unavailableClient) Ask(context.Context, agenttypes.AskRequest) (*agenttypes.AskResponse, error) {
	return nil, fmt.Errorf("agent unavailable: %s", c.reason)
}

func (c *unavailableClient) AskStream(context.Context, agenttypes.AskRequest, func(agenttypes.StreamEvent) error) (*agenttypes.AskResponse, error) {
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
	return c.AskStream(ctx, request, nil)
}

func (c *EinoClient) AskStream(ctx context.Context, request agenttypes.AskRequest, onEvent func(agenttypes.StreamEvent) error) (*agenttypes.AskResponse, error) {
	reactAgent, agentOptions, collectSources, err := c.newReActAgent(ctx, request)
	if err != nil {
		return nil, err
	}

	stream, err := reactAgent.Stream(ctx, buildMessages(request), agentOptions...)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	var answerBuilder strings.Builder
	result := &agenttypes.AskResponse{}
	for {
		chunk, recvErr := stream.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return nil, recvErr
		}
		if chunk == nil {
			continue
		}
		if chunk.ResponseMeta != nil && chunk.ResponseMeta.Usage != nil {
			result.TokenUsage = chunk.ResponseMeta.Usage.TotalTokens
		}
		content := chunk.Content
		if content == "" {
			continue
		}
		answerBuilder.WriteString(content)
		if onEvent != nil {
			if err := onEvent(agenttypes.StreamEvent{Type: agenttypes.StreamEventDelta, Content: content}); err != nil {
				return nil, err
			}
		}
	}

	result.Answer = strings.TrimSpace(answerBuilder.String())
	result.Sources = collectSources()
	if onEvent != nil {
		if err := onEvent(agenttypes.StreamEvent{Type: agenttypes.StreamEventComplete, Answer: result.Answer, Sources: result.Sources, TokenUsage: result.TokenUsage}); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (c *EinoClient) newReActAgent(ctx context.Context, request agenttypes.AskRequest) (*react.Agent, []agent.AgentOption, func() []agenttypes.Source, error) {
	courseTools, collectSources, err := tools.BuildCourseTools(request.Materials)
	if err != nil {
		return nil, nil, nil, err
	}
	agentOptions, err := react.WithTools(ctx, courseTools...)
	if err != nil {
		return nil, nil, nil, err
	}
	toolCallingModel, ok := any(c.chatModel).(model.ToolCallingChatModel)
	if !ok {
		return nil, nil, nil, fmt.Errorf("chat model does not support tool calling")
	}
	reactAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: toolCallingModel,
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			_ = ctx
			messages := make([]*schema.Message, 0, len(input)+1)
			messages = append(messages, schema.SystemMessage(buildSystemPrompt(request)))
			messages = append(messages, input...)
			return messages
		},
		MaxStep: 12,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return reactAgent, agentOptions, collectSources, nil
}

func buildMessages(request agenttypes.AskRequest) []*schema.Message {
	messages := make([]*schema.Message, 0, len(request.History)+1)
	for _, turn := range request.History {
		switch turn.Role {
		case "agent":
			messages = append(messages, schema.AssistantMessage(turn.Content, nil))
		default:
			messages = append(messages, schema.UserMessage(turn.Content))
		}
	}
	messages = append(messages, schema.UserMessage(request.Question))
	return messages
}

func buildSystemPrompt(request agenttypes.AskRequest) string {
	var builder strings.Builder
	builder.WriteString(strings.TrimSpace(request.PromptTemplate))
	builder.WriteString("\n\n")
	builder.WriteString("你当前扮演的课程 Agent 名称：")
	builder.WriteString(request.AgentName)
	builder.WriteString("。\n")
	builder.WriteString("你只能回答当前课程资料范围内的问题，不能使用课程外的信息。\n")
	builder.WriteString("回答前优先使用 search_course_materials 和 read_course_material 工具核对资料；如果没有足够依据，直接说明资料不足。\n")
	builder.WriteString("回答中尽量点明引用的资料文件名，不要编造未在资料中出现的事实。")
	return builder.String()
}
