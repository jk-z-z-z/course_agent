package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	einoOpenAI "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

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

type unavailableClient struct {
	reason string
}

func (c *unavailableClient) Ask(context.Context, AskRequest) (*AskResponse, error) {
	return nil, fmt.Errorf("agent unavailable: %s", c.reason)
}

type EinoClient struct {
	chatModel *einoOpenAI.ChatModel
}

func NewClient(ctx context.Context, cfg Config) (Client, error) {
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

func (c *EinoClient) Ask(ctx context.Context, request AskRequest) (*AskResponse, error) {
	sources := retrieveSources(request.Question, request.Materials)
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

	result := &AskResponse{
		Answer:  strings.TrimSpace(resp.Content),
		Sources: sources,
	}
	if resp.ResponseMeta != nil && resp.ResponseMeta.Usage != nil {
		result.TokenUsage = resp.ResponseMeta.Usage.TotalTokens
	}
	return result, nil
}

func buildSystemPrompt(request AskRequest, sources []Source) string {
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

var tokenPattern = regexp.MustCompile(`[\p{Han}]{1,12}|[a-zA-Z0-9_]{2,}`)

type scoredSource struct {
	Source
	score int
}

func retrieveSources(question string, materials []MaterialDocument) []Source {
	keywords := extractKeywords(question)
	scored := make([]scoredSource, 0)
	for _, material := range materials {
		snippet, score, err := extractSnippet(material, question, keywords)
		if err != nil || score <= 0 || strings.TrimSpace(snippet) == "" {
			continue
		}
		scored = append(scored, scoredSource{
			Source: Source{MaterialNodeID: material.MaterialNodeID, FileName: material.FileName, Snippet: snippet},
			score:  score,
		})
	}
	sort.SliceStable(scored, func(i, j int) bool { return scored[i].score > scored[j].score })
	limit := 4
	if len(scored) < limit {
		limit = len(scored)
	}
	result := make([]Source, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, scored[i].Source)
	}
	return result
}

func extractKeywords(question string) []string {
	matches := tokenPattern.FindAllString(strings.ToLower(question), -1)
	keywords := make([]string, 0, len(matches))
	seen := make(map[string]struct{})
	for _, match := range matches {
		trimmed := strings.TrimSpace(match)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		keywords = append(keywords, trimmed)
	}
	return keywords
}

func extractSnippet(material MaterialDocument, question string, keywords []string) (string, int, error) {
	if !isTextLike(material) {
		return "", 0, nil
	}
	contentBytes, err := os.ReadFile(material.StoragePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", 0, nil
		}
		return "", 0, err
	}
	content := normalizeContent(string(contentBytes))
	if content == "" {
		return "", 0, nil
	}
	lines := strings.Split(content, "\n")
	bestLine := ""
	bestScore := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		score := scoreLine(strings.ToLower(line), strings.ToLower(question), keywords)
		if score > bestScore {
			bestScore = score
			bestLine = line
		}
	}
	if bestScore == 0 {
		bestLine = truncateText(content, 220)
		bestScore = 1
	}
	return truncateText(bestLine, 240), bestScore, nil
}

func scoreLine(line, question string, keywords []string) int {
	score := 0
	if question != "" && strings.Contains(line, question) {
		score += 8
	}
	for _, keyword := range keywords {
		if strings.Contains(line, keyword) {
			score += 3
		}
	}
	return score
}

func isTextLike(material MaterialDocument) bool {
	mimeType := strings.ToLower(material.MimeType)
	if strings.HasPrefix(mimeType, "text/") || strings.Contains(mimeType, "json") || strings.Contains(mimeType, "xml") {
		return true
	}
	name := strings.ToLower(material.FileName)
	switch {
	case strings.HasSuffix(name, ".txt"), strings.HasSuffix(name, ".md"), strings.HasSuffix(name, ".markdown"), strings.HasSuffix(name, ".csv"), strings.HasSuffix(name, ".json"), strings.HasSuffix(name, ".yaml"), strings.HasSuffix(name, ".yml"), strings.HasSuffix(name, ".log"), strings.HasSuffix(name, ".html"), strings.HasSuffix(name, ".htm"):
		return true
	default:
		return false
	}
}

func normalizeContent(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	return strings.TrimSpace(content)
}

func truncateText(text string, limit int) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit]) + "..."
}
