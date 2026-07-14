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

type StudyPlanHandler struct {
	service *service.StudyPlanService
}

func NewStudyPlanHandler(service *service.StudyPlanService) *StudyPlanHandler {
	return &StudyPlanHandler{service: service}
}

func (h *StudyPlanHandler) List(c *gin.Context) {
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
	data, err := h.service.ListPlans(c.Request.Context(), userID, courseID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *StudyPlanHandler) GetDetail(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	planID, ok := parseUintParam(c, "planId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	data, err := h.service.GetPlan(c.Request.Context(), userID, courseID, planID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *StudyPlanHandler) Generate(c *gin.Context) {
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
	var req dto.GenerateStudyPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	data, err := h.service.GeneratePlan(c.Request.Context(), userID, courseID, strings.TrimSpace(req.Goal), req.DailyMinutes)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *StudyPlanHandler) UpdateItem(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	planID, ok := parseUintParam(c, "planId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	itemID, ok := parseUintParam(c, "itemId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	var req dto.UpdateStudyPlanItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	data, err := h.service.UpdatePlanItemStatus(c.Request.Context(), userID, courseID, planID, itemID, strings.TrimSpace(req.Status))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *StudyPlanHandler) writeError(c *gin.Context, err error) {
	if codeErr, ok := err.(*apperrors.CodeError); ok {
		status := http.StatusBadRequest
		switch codeErr.Code {
		case apperrors.ErrUnauthorized.Code, apperrors.ErrSessionExpired.Code:
			status = http.StatusUnauthorized
		case apperrors.ErrForbidden.Code:
			status = http.StatusForbidden
		case apperrors.ErrCourseNotFound.Code:
			status = http.StatusNotFound
		case apperrors.ErrStudyPlanUnavailable.Code:
			status = http.StatusServiceUnavailable
		}
		response.Fail(c, status, codeErr.Code, codeErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
}
