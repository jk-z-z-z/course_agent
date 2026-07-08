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

type MaterialHandler struct {
	service *service.MaterialService
}

func NewMaterialHandler(service *service.MaterialService) *MaterialHandler {
	return &MaterialHandler{service: service}
}

func (h *MaterialHandler) GetTree(c *gin.Context) {
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
	data, err := h.service.GetTree(c.Request.Context(), userID, courseID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *MaterialHandler) CreateFolder(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	var req dto.CreateMaterialFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	data, err := h.service.CreateFolder(c.Request.Context(), userID, courseID, req.ParentID, strings.TrimSpace(req.FolderName))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *MaterialHandler) GetDetail(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	nodeID, ok := parseUintParam(c, "nodeId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	data, err := h.service.GetDetail(c.Request.Context(), userID, courseID, nodeID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *MaterialHandler) UpdateNode(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	nodeID, ok := parseUintParam(c, "nodeId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	var req dto.UpdateMaterialNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	data, err := h.service.UpdateNode(c.Request.Context(), userID, courseID, nodeID, strings.TrimSpace(req.NodeName), req.SortIndex)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *MaterialHandler) DeleteNode(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	nodeID, ok := parseUintParam(c, "nodeId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	if err := h.service.DeleteNode(c.Request.Context(), userID, courseID, nodeID); err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *MaterialHandler) writeError(c *gin.Context, err error) {
	if codeErr, ok := err.(*apperrors.CodeError); ok {
		status := http.StatusBadRequest
		switch codeErr.Code {
		case apperrors.ErrUnauthorized.Code, apperrors.ErrSessionExpired.Code:
			status = http.StatusUnauthorized
		case apperrors.ErrForbidden.Code:
			status = http.StatusForbidden
		case apperrors.ErrCourseNotFound.Code, apperrors.ErrMaterialNotFound.Code, apperrors.ErrStorageSpaceNotFound.Code:
			status = http.StatusNotFound
		}
		response.Fail(c, status, codeErr.Code, codeErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
}
