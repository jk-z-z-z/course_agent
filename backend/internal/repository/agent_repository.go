package repository

import (
	"context"

	"gorm.io/gorm"

	"course_agent_backend/internal/model"
)

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *AgentRepository) CreateCourseAgent(ctx context.Context, agent *model.CourseAgent) error {
	return r.db.WithContext(ctx).Create(agent).Error
}

func (r *AgentRepository) CreateCourseAgentTx(tx *gorm.DB, agent *model.CourseAgent) error {
	return tx.Create(agent).Error
}

func (r *AgentRepository) GetCourseAgentByCourseID(ctx context.Context, courseID uint64) (*model.CourseAgent, error) {
	var agent model.CourseAgent
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseID).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) CreateConversation(ctx context.Context, conversation *model.AgentConversation) error {
	return r.db.WithContext(ctx).Create(conversation).Error
}

func (r *AgentRepository) CreateConversationTx(tx *gorm.DB, conversation *model.AgentConversation) error {
	return tx.Create(conversation).Error
}

func (r *AgentRepository) ListConversationsByCourseID(ctx context.Context, courseID uint64) ([]model.AgentConversation, error) {
	var conversations []model.AgentConversation
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseID).Order("updated_at DESC, id DESC").Find(&conversations).Error; err != nil {
		return nil, err
	}
	return conversations, nil
}

func (r *AgentRepository) ListConversationsByCourseIDAndUserID(ctx context.Context, courseID, userID uint64) ([]model.AgentConversation, error) {
	var conversations []model.AgentConversation
	if err := r.db.WithContext(ctx).Where("course_id = ? AND user_id = ?", courseID, userID).Order("updated_at DESC, id DESC").Find(&conversations).Error; err != nil {
		return nil, err
	}
	return conversations, nil
}

func (r *AgentRepository) GetConversationByID(ctx context.Context, conversationID uint64) (*model.AgentConversation, error) {
	var conversation model.AgentConversation
	if err := r.db.WithContext(ctx).First(&conversation, conversationID).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *AgentRepository) UpdateConversation(ctx context.Context, conversation *model.AgentConversation) error {
	return r.db.WithContext(ctx).Save(conversation).Error
}

func (r *AgentRepository) UpdateConversationTx(tx *gorm.DB, conversation *model.AgentConversation) error {
	return tx.Save(conversation).Error
}

func (r *AgentRepository) CreateMessage(ctx context.Context, message *model.AgentMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *AgentRepository) CreateMessageTx(tx *gorm.DB, message *model.AgentMessage) error {
	return tx.Create(message).Error
}

func (r *AgentRepository) ListMessagesByConversationID(ctx context.Context, conversationID uint64) ([]model.AgentMessage, error) {
	var messages []model.AgentMessage
	if err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at ASC, id ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *AgentRepository) CreateMessageSourcesTx(tx *gorm.DB, sources []model.AgentMessageSource) error {
	if len(sources) == 0 {
		return nil
	}
	return tx.Create(&sources).Error
}

func (r *AgentRepository) ListSourcesByMessageIDs(ctx context.Context, messageIDs []uint64) ([]model.AgentMessageSource, error) {
	if len(messageIDs) == 0 {
		return []model.AgentMessageSource{}, nil
	}
	var sources []model.AgentMessageSource
	if err := r.db.WithContext(ctx).Where("message_id IN ?", messageIDs).Order("id ASC").Find(&sources).Error; err != nil {
		return nil, err
	}
	return sources, nil
}
