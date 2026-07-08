package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"course_agent_backend/internal/authcontext"
	apperrors "course_agent_backend/internal/errors"
	"course_agent_backend/internal/dto"
	"course_agent_backend/internal/response"
	"course_agent_backend/internal/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}

	user, err := h.service.Register(c.Request.Context(), strings.TrimSpace(req.Username), req.Password, strings.TrimSpace(req.Phone))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}

	token, expiredAt, user, err := h.service.Login(c.Request.Context(), strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, map[string]interface{}{
		"token":     token,
		"expiredAt": expiredAt.Format(time.RFC3339),
		"user":      user,
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
	token := bearerToken(c.GetHeader("Authorization"))
	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	user, err := h.service.Me(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) writeError(c *gin.Context, err error) {
	if codeErr, ok := err.(*apperrors.CodeError); ok {
		status := http.StatusBadRequest
		switch codeErr.Code {
		case apperrors.ErrUnauthorized.Code, apperrors.ErrSessionExpired.Code:
			status = http.StatusUnauthorized
		case apperrors.ErrForbidden.Code:
			status = http.StatusForbidden
		case apperrors.ErrUserNotFound.Code:
			status = http.StatusNotFound
		}
		response.Fail(c, status, codeErr.Code, codeErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
}

func bearerToken(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}
