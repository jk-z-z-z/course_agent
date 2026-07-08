package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Fail(w, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}

	user, err := h.service.Register(r.Context(), strings.TrimSpace(req.Username), req.Password, strings.TrimSpace(req.Phone))
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Fail(w, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}

	token, expiredAt, user, err := h.service.Login(r.Context(), strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, map[string]interface{}{
		"token":     token,
		"expiredAt": expiredAt.Format(time.RFC3339),
		"user":      user,
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r.Header.Get("Authorization"))
	if err := h.service.Logout(r.Context(), token); err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, nil)
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := authcontext.UserID(r.Context())
	if !ok || userID == 0 {
		response.Fail(w, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	user, err := h.service.Me(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, user)
}

func (h *UserHandler) writeError(w http.ResponseWriter, err error) {
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
		response.Fail(w, status, codeErr.Code, codeErr.Message)
		return
	}
	response.Fail(w, http.StatusInternalServerError, 50000, err.Error())
}

func bearerToken(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}
