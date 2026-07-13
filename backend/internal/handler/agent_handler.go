package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	agentruntime "course_agent_backend/internal/agent"
	"course_agent_backend/internal/authcontext"
	"course_agent_backend/internal/dto"
	apperrors "course_agent_backend/internal/errors"
	"course_agent_backend/internal/response"
	"course_agent_backend/internal/service"
)

type AgentHandler struct {
	service *service.AgentService
}

func NewAgentHandler(service *service.AgentService) *AgentHandler {
	return &AgentHandler{service: service}
}

func (h *AgentHandler) GetAgent(c *gin.Context) {
	h.withCourseUser(c, func(userID, courseID uint64) {
		data, err := h.service.GetAgent(c.Request.Context(), userID, courseID)
		if err != nil {
			h.writeError(c, err)
			return
		}
		response.Success(c, data)
	})
}

func (h *AgentHandler) CreateConversation(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	var req dto.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	data, err := h.service.CreateConversation(c.Request.Context(), userID, courseID, strings.TrimSpace(req.Title))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *AgentHandler) ListConversations(c *gin.Context) {
	h.withCourseUser(c, func(userID, courseID uint64) {
		data, err := h.service.ListConversations(c.Request.Context(), userID, courseID)
		if err != nil {
			h.writeError(c, err)
			return
		}
		response.Success(c, data)
	})
}

func (h *AgentHandler) GetConversation(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	conversationID, ok := parseUintParam(c, "conversationId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	data, err := h.service.GetConversationDetail(c.Request.Context(), userID, courseID, conversationID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *AgentHandler) Ask(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	var req dto.AskAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	data, err := h.service.Ask(c.Request.Context(), userID, courseID, req.ConversationID, strings.TrimSpace(req.Question))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *AgentHandler) AskStream(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	var req dto.AskAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	result, err := h.service.AskStream(c.Request.Context(), userID, courseID, req.ConversationID, strings.TrimSpace(req.Question), func(event agentruntime.StreamEvent) error {
		switch event.Type {
		case agentruntime.StreamEventDelta:
			return writeSSE(c, "delta", map[string]any{"content": event.Content})
		case agentruntime.StreamEventComplete:
			return writeSSE(c, "complete", map[string]any{
				"answer":             event.Answer,
				"sources":            event.Sources,
				"retrievedMaterials": event.RetrievedMaterials,
				"tokenUsage":         event.TokenUsage,
			})
		default:
			return nil
		}
	})
	if err != nil {
		_ = writeSSE(c, "error", map[string]any{"message": h.errorMessage(err)})
		return
	}
	_ = writeSSE(c, "done", result)
}

func (h *AgentHandler) withCourseUser(c *gin.Context, fn func(userID, courseID uint64)) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	fn(userID, courseID)
}

func (h *AgentHandler) writeError(c *gin.Context, err error) {
	var codeErr *apperrors.CodeError
	if errors.As(err, &codeErr) {
		status := http.StatusBadRequest
		switch codeErr.Code {
		case apperrors.ErrUnauthorized.Code, apperrors.ErrSessionExpired.Code:
			status = http.StatusUnauthorized
		case apperrors.ErrForbidden.Code:
			status = http.StatusForbidden
		case apperrors.ErrCourseNotFound.Code, apperrors.ErrAgentNotFound.Code, apperrors.ErrConversationNotFound.Code:
			status = http.StatusNotFound
		}
		response.Fail(c, status, codeErr.Code, codeErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
}

func (h *AgentHandler) errorMessage(err error) string {
	var codeErr *apperrors.CodeError
	if errors.As(err, &codeErr) {
		return codeErr.Message
	}
	return "请求失败"
}

func writeSSE(c *gin.Context, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := c.Writer.WriteString("event: " + event + "\n"); err != nil {
		return err
	}
	if _, err := c.Writer.WriteString("data: " + string(data) + "\n\n"); err != nil {
		return err
	}
	c.Writer.Flush()
	return nil
}
