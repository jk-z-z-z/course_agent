package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

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

func (h *AgentHandler) UpdateAgent(c *gin.Context) {
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
	var req dto.UpdateCourseAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	data, err := h.service.UpdateAgent(c.Request.Context(), userID, courseID, strings.TrimSpace(req.AgentName), strings.TrimSpace(req.PromptTemplate), strings.TrimSpace(req.Status), strings.TrimSpace(req.RetrievalScope))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
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
	if codeErr, ok := err.(*apperrors.CodeError); ok {
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
