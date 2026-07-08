package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"course_agent_backend/internal/authcontext"
	"course_agent_backend/internal/dto"
	apperrors "course_agent_backend/internal/errors"
	"course_agent_backend/internal/response"
	"course_agent_backend/internal/service"
)

type CourseHandler struct {
	service *service.CourseService
}

func NewCourseHandler(service *service.CourseService) *CourseHandler {
	return &CourseHandler{service: service}
}

func (h *CourseHandler) CreateCourse(c *gin.Context) {
	var req dto.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	course, err := h.service.CreateCourse(c.Request.Context(), userID, req.CourseCode, req.CourseName, req.CourseDescription)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, course)
}

func (h *CourseHandler) ListCourses(c *gin.Context) {
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	courses, err := h.service.ListCourses(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, courses)
}

func (h *CourseHandler) GetCourse(c *gin.Context) {
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
	course, err := h.service.GetCourseDetail(c.Request.Context(), userID, courseID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, course)
}

func (h *CourseHandler) UpdateCourse(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	var req dto.UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	course, err := h.service.UpdateCourse(c.Request.Context(), userID, courseID, req.CourseName, req.CourseDescription)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, course)
}

func (h *CourseHandler) DeleteCourse(c *gin.Context) {
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
	if err := h.service.DeleteCourse(c.Request.Context(), userID, courseID); err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *CourseHandler) ListMembers(c *gin.Context) {
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
	members, err := h.service.ListMembers(c.Request.Context(), userID, courseID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, members)
}

func (h *CourseHandler) AddMember(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	var req dto.AddCourseMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	member, err := h.service.AddMember(c.Request.Context(), userID, courseID, strings.TrimSpace(req.UserIdentifier), strings.TrimSpace(req.Role))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, member)
}

func (h *CourseHandler) UpdateMember(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	memberID, ok := parseUintParam(c, "memberId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	var req dto.UpdateCourseMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	member, err := h.service.UpdateMemberRole(c.Request.Context(), userID, courseID, memberID, strings.TrimSpace(req.Role))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, member)
}

func (h *CourseHandler) DeleteMember(c *gin.Context) {
	courseID, ok := parseUintParam(c, "courseId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	memberID, ok := parseUintParam(c, "memberId")
	if !ok {
		response.Fail(c, http.StatusBadRequest, apperrors.ErrInvalidParameter.Code, apperrors.ErrInvalidParameter.Message)
		return
	}
	userID, ok := authcontext.UserID(c.Request.Context())
	if !ok || userID == 0 {
		response.Fail(c, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code, apperrors.ErrUnauthorized.Message)
		return
	}
	if err := h.service.RemoveMember(c.Request.Context(), userID, courseID, memberID); err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *CourseHandler) writeError(c *gin.Context, err error) {
	if codeErr, ok := err.(*apperrors.CodeError); ok {
		status := http.StatusBadRequest
		switch codeErr.Code {
		case apperrors.ErrUnauthorized.Code, apperrors.ErrSessionExpired.Code:
			status = http.StatusUnauthorized
		case apperrors.ErrForbidden.Code:
			status = http.StatusForbidden
		case apperrors.ErrUserNotFound.Code, apperrors.ErrCourseNotFound.Code, apperrors.ErrCourseMemberNotFound.Code:
			status = http.StatusNotFound
		}
		response.Fail(c, status, codeErr.Code, codeErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
}

func parseUintParam(c *gin.Context, key string) (uint64, bool) {
	value := strings.TrimSpace(c.Param(key))
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}
