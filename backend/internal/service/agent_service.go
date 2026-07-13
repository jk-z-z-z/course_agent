package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"

	agentruntime "course_agent_backend/internal/agent"
	apperrors "course_agent_backend/internal/errors"
	"course_agent_backend/internal/model"
	"course_agent_backend/internal/repository"
	"course_agent_backend/internal/vo"
)

type AgentService struct {
	courseRepo   *repository.CourseRepository
	agentRepo    *repository.AgentRepository
	materialRepo *repository.MaterialRepository
	runtime      agentruntime.Client
}

func NewAgentService(courseRepo *repository.CourseRepository, agentRepo *repository.AgentRepository, materialRepo *repository.MaterialRepository, runtime agentruntime.Client) *AgentService {
	return &AgentService{courseRepo: courseRepo, agentRepo: agentRepo, materialRepo: materialRepo, runtime: runtime}
}

func (s *AgentService) GetAgent(ctx context.Context, userID, courseID uint64) (*vo.CourseAgentVO, error) {
	_, _, err := s.requireCourseMember(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	agentModel, err := s.agentRepo.GetCourseAgentByCourseID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrAgentNotFound
		}
		return nil, err
	}
	result := toCourseAgentVO(agentModel, false)
	return &result, nil
}

func (s *AgentService) CreateConversation(ctx context.Context, userID, courseID uint64, title string) (*vo.AgentConversationVO, error) {
	_, _, err := s.requireCourseMember(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	agentModel, err := s.agentRepo.GetCourseAgentByCourseID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrAgentNotFound
		}
		return nil, err
	}
	conversation := &model.AgentConversation{
		CourseID:          courseID,
		AgentID:           agentModel.ID,
		UserID:            userID,
		ConversationTitle: strings.TrimSpace(title),
	}
	if conversation.ConversationTitle == "" {
		conversation.ConversationTitle = "新会话"
	}
	if err := s.agentRepo.CreateConversation(ctx, conversation); err != nil {
		return nil, err
	}
	result := toConversationVO(conversation)
	return &result, nil
}

func (s *AgentService) ListConversations(ctx context.Context, userID, courseID uint64) ([]vo.AgentConversationVO, error) {
	_, role, err := s.requireCourseMember(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	var conversations []model.AgentConversation
	if role == "owner" || role == "teacher" {
		conversations, err = s.agentRepo.ListConversationsByCourseID(ctx, courseID)
	} else {
		conversations, err = s.agentRepo.ListConversationsByCourseIDAndUserID(ctx, courseID, userID)
	}
	if err != nil {
		return nil, err
	}
	result := make([]vo.AgentConversationVO, 0, len(conversations))
	for _, conversation := range conversations {
		result = append(result, toConversationVO(&conversation))
	}
	return result, nil
}

func (s *AgentService) GetConversationDetail(ctx context.Context, userID, courseID, conversationID uint64) (*vo.AgentConversationDetailVO, error) {
	conversation, _, err := s.requireConversationAccess(ctx, userID, courseID, conversationID)
	if err != nil {
		return nil, err
	}
	messages, err := s.agentRepo.ListMessagesByConversationID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	messageIDs := make([]uint64, 0, len(messages))
	for _, message := range messages {
		messageIDs = append(messageIDs, message.ID)
	}
	sources, err := s.agentRepo.ListSourcesByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, err
	}
	nodes, err := s.materialRepo.ListActiveNodesByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	fileNames := make(map[uint64]string)
	for _, node := range nodes {
		fileNames[node.ID] = node.NodeName
	}
	byMessageID := make(map[uint64][]vo.AgentMessageSourceVO)
	for _, source := range sources {
		byMessageID[source.MessageID] = append(byMessageID[source.MessageID], vo.AgentMessageSourceVO{
			ID:             source.ID,
			MessageID:      source.MessageID,
			MaterialNodeID: source.MaterialNodeID,
			FileName:       fileNames[source.MaterialNodeID],
			SnippetText:    source.SnippetText,
			CreatedAt:      source.CreatedAt,
		})
	}
	messageVOs := make([]vo.AgentMessageVO, 0, len(messages))
	for _, message := range messages {
		messageSources := byMessageID[message.ID]
		messageVOs = append(messageVOs, vo.AgentMessageVO{
			ID:                 message.ID,
			ConversationID:     message.ConversationID,
			SenderType:         message.SenderType,
			MessageContent:     message.MessageContent,
			TokenUsage:         message.TokenUsage,
			CreatedAt:          message.CreatedAt,
			Sources:            messageSources,
			RetrievedMaterials: messageSources,
		})
	}
	result := &vo.AgentConversationDetailVO{Conversation: toConversationVO(conversation), Messages: messageVOs}
	return result, nil
}

func (s *AgentService) Ask(ctx context.Context, userID, courseID, conversationID uint64, question string) (*vo.AgentAskResultVO, error) {
	result, err := s.ask(ctx, userID, courseID, conversationID, question, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AgentService) AskStream(ctx context.Context, userID, courseID, conversationID uint64, question string, onEvent func(agentruntime.StreamEvent) error) (*vo.AgentAskResultVO, error) {
	result, err := s.ask(ctx, userID, courseID, conversationID, question, onEvent)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AgentService) ask(ctx context.Context, userID, courseID, conversationID uint64, question string, onEvent func(agentruntime.StreamEvent) error) (*vo.AgentAskResultVO, error) {
	conversation, role, err := s.requireConversationAccess(ctx, userID, courseID, conversationID)
	if err != nil {
		return nil, err
	}
	_ = role
	agentModel, err := s.agentRepo.GetCourseAgentByCourseID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrAgentNotFound
		}
		return nil, err
	}
	if agentModel.Status != "enabled" {
		return nil, apperrors.ErrAgentDisabled
	}
	trimmedQuestion := strings.TrimSpace(question)
	if trimmedQuestion == "" {
		return nil, apperrors.ErrInvalidParameter
	}

	historyMessages, err := s.agentRepo.ListMessagesByConversationID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	history := make([]agentruntime.ConversationTurn, 0, len(historyMessages))
	for _, message := range historyMessages {
		history = append(history, agentruntime.ConversationTurn{Role: message.SenderType, Content: message.MessageContent})
	}

	materialNodes, err := s.materialRepo.ListActiveNodesByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	materials := make([]agentruntime.MaterialDocument, 0)
	for _, node := range materialNodes {
		if node.NodeType != "file" || strings.TrimSpace(node.StoragePath) == "" {
			continue
		}
		materials = append(materials, agentruntime.MaterialDocument{
			MaterialNodeID: node.ID,
			FileName:       node.NodeName,
			StoragePath:    node.StoragePath,
			MimeType:       node.MimeType,
		})
	}

	answer, err := s.runtime.AskStream(ctx, agentruntime.AskRequest{
		AgentName:      agentModel.AgentName,
		PromptTemplate: agentModel.PromptTemplate,
		Question:       trimmedQuestion,
		History:        history,
		Materials:      materials,
	}, onEvent)
	if err != nil {
		publicMessage := summarizeAgentRuntimeError(err)
		log.Printf("agent ask failed: course_id=%d conversation_id=%d user_id=%d agent_name=%q base_error=%v", courseID, conversationID, userID, agentModel.AgentName, err)
		return nil, fmt.Errorf("%w: %v", apperrors.New(apperrors.ErrAgentUnavailable.Code, publicMessage), err)
	}

	retrievedMaterials := answer.RetrievedMaterials
	if len(retrievedMaterials) == 0 {
		retrievedMaterials = answer.Sources
	}

	result := &vo.AgentAskResultVO{
		ConversationID:     conversationID,
		Question:           trimmedQuestion,
		Answer:             answer.Answer,
		Sources:            make([]vo.AgentMessageSourceVO, 0, len(retrievedMaterials)),
		RetrievedMaterials: make([]vo.AgentRetrievedMaterialVO, 0, len(retrievedMaterials)),
	}
	for _, source := range retrievedMaterials {
		sourceVO := vo.AgentMessageSourceVO{
			MaterialNodeID: source.MaterialNodeID,
			FileName:       source.FileName,
			SnippetText:    source.Snippet,
		}
		result.Sources = append(result.Sources, sourceVO)
		result.RetrievedMaterials = append(result.RetrievedMaterials, sourceVO)
	}

	err = s.agentRepo.Transaction(ctx, func(tx *gorm.DB) error {
		userMessage := &model.AgentMessage{ConversationID: conversationID, SenderType: "user", MessageContent: trimmedQuestion}
		if err := s.agentRepo.CreateMessageTx(tx, userMessage); err != nil {
			return err
		}
		agentMessage := &model.AgentMessage{ConversationID: conversationID, SenderType: "agent", MessageContent: answer.Answer, TokenUsage: answer.TokenUsage}
		if err := s.agentRepo.CreateMessageTx(tx, agentMessage); err != nil {
			return err
		}
		sourceModels := make([]model.AgentMessageSource, 0, len(retrievedMaterials))
		for _, source := range retrievedMaterials {
			sourceModels = append(sourceModels, model.AgentMessageSource{MessageID: agentMessage.ID, MaterialNodeID: source.MaterialNodeID, MaterialVersionID: source.VersionID, SnippetText: source.Snippet, CreatedAt: time.Now()})
		}
		if err := s.agentRepo.CreateMessageSourcesTx(tx, sourceModels); err != nil {
			return err
		}
		if strings.TrimSpace(conversation.ConversationTitle) == "" || conversation.ConversationTitle == "新会话" {
			conversation.ConversationTitle = truncateConversationTitle(trimmedQuestion)
		}
		conversation.UpdatedAt = time.Now()
		if err := s.agentRepo.UpdateConversationTx(tx, conversation); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AgentService) requireCourseMember(ctx context.Context, userID, courseID uint64) (*model.Course, string, error) {
	course, err := s.courseRepo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", apperrors.ErrCourseNotFound
		}
		return nil, "", err
	}
	if course.Status != "active" {
		return nil, "", apperrors.ErrCourseNotFound
	}
	member, err := s.courseRepo.GetMember(ctx, courseID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", apperrors.ErrForbidden
		}
		return nil, "", err
	}
	if member.JoinStatus != "active" {
		return nil, "", apperrors.ErrForbidden
	}
	return course, member.Role, nil
}

func (s *AgentService) requireConversationAccess(ctx context.Context, userID, courseID, conversationID uint64) (*model.AgentConversation, string, error) {
	_, role, err := s.requireCourseMember(ctx, userID, courseID)
	if err != nil {
		return nil, "", err
	}
	conversation, err := s.agentRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", apperrors.ErrConversationNotFound
		}
		return nil, "", err
	}
	if conversation.CourseID != courseID {
		return nil, "", apperrors.ErrConversationNotFound
	}
	if role == "student" && conversation.UserID != userID {
		return nil, "", apperrors.ErrForbidden
	}
	return conversation, role, nil
}

func toCourseAgentVO(agentModel *model.CourseAgent, includePrompt bool) vo.CourseAgentVO {
	result := vo.CourseAgentVO{
		ID:             agentModel.ID,
		CourseID:       agentModel.CourseID,
		AgentName:      agentModel.AgentName,
		Status:         agentModel.Status,
		RetrievalScope: agentModel.RetrievalScope,
		CreatedBy:      agentModel.CreatedBy,
		CreatedAt:      agentModel.CreatedAt,
		UpdatedAt:      agentModel.UpdatedAt,
	}
	if includePrompt {
		result.PromptTemplate = agentModel.PromptTemplate
	}
	return result
}

func toConversationVO(conversation *model.AgentConversation) vo.AgentConversationVO {
	return vo.AgentConversationVO{
		ID:                conversation.ID,
		CourseID:          conversation.CourseID,
		AgentID:           conversation.AgentID,
		UserID:            conversation.UserID,
		ConversationTitle: conversation.ConversationTitle,
		CreatedAt:         conversation.CreatedAt,
		UpdatedAt:         conversation.UpdatedAt,
	}
}

func truncateConversationTitle(question string) string {
	runes := []rune(strings.TrimSpace(question))
	if len(runes) <= 24 {
		return string(runes)
	}
	return string(runes[:24]) + "..."
}

func summarizeAgentRuntimeError(err error) string {
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case strings.Contains(message, "incorrect api key provided"):
		return "Agent API Key 无效，请检查百炼密钥配置"
	case strings.Contains(message, "free quota has been exhausted"):
		return "Agent 调用额度已耗尽，请检查百炼额度或计费配置"
	case strings.Contains(message, "status code: 401"):
		return "Agent 鉴权失败，请检查百炼 API Key"
	case strings.Contains(message, "status code: 403"):
		return "Agent 当前无可用调用权限，请检查额度、计费或模型授权"
	case strings.Contains(message, "does not support tool calling"):
		return "当前模型不支持工具调用，请更换支持 tool calling 的模型"
	case strings.Contains(message, "context deadline exceeded"):
		return "Agent 响应超时，请稍后重试"
	case strings.Contains(message, "context canceled"):
		return "Agent 请求已取消"
	default:
		return apperrors.ErrAgentUnavailable.Message
	}
}
