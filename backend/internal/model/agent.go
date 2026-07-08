package model

import "time"

type CourseAgent struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement"`
	CourseID       uint64    `gorm:"not null;uniqueIndex"`
	AgentName      string    `gorm:"size:128;not null"`
	PromptTemplate string    `gorm:"type:text"`
	Status         string    `gorm:"size:16;not null;default:enabled"`
	RetrievalScope string    `gorm:"size:16;not null;default:course_all"`
	CreatedBy      uint64    `gorm:"not null;index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (CourseAgent) TableName() string {
	return "course_agents"
}

type AgentConversation struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement"`
	CourseID          uint64    `gorm:"not null;index"`
	AgentID           uint64    `gorm:"not null;index"`
	UserID            uint64    `gorm:"not null;index"`
	ConversationTitle string    `gorm:"size:255"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (AgentConversation) TableName() string {
	return "agent_conversations"
}

type AgentMessage struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement"`
	ConversationID uint64    `gorm:"not null;index"`
	SenderType     string    `gorm:"size:16;not null"`
	MessageContent string    `gorm:"type:text;not null"`
	TokenUsage     int       `gorm:"not null;default:0"`
	CreatedAt      time.Time
}

func (AgentMessage) TableName() string {
	return "agent_messages"
}

type AgentMessageSource struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement"`
	MessageID         uint64    `gorm:"not null;index"`
	MaterialNodeID    uint64    `gorm:"not null;index"`
	MaterialVersionID *uint64   `gorm:"index"`
	SnippetText       string    `gorm:"type:text"`
	CreatedAt         time.Time
}

func (AgentMessageSource) TableName() string {
	return "agent_message_sources"
}
