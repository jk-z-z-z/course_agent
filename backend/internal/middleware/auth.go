package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"course_agent_backend/internal/authcontext"
	apperrors "course_agent_backend/internal/errors"
	"course_agent_backend/internal/response"
	"course_agent_backend/internal/service"
)

type AuthMiddleware struct {
	service *service.UserService
}

func NewAuthMiddleware(service *service.UserService) *AuthMiddleware {
	return &AuthMiddleware{service: service}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		userID, err := m.service.ResolveToken(c.Request.Context(), token)
		if err != nil {
			if codeErr, ok := err.(*apperrors.CodeError); ok {
				response.Fail(c, http.StatusUnauthorized, codeErr.Code, codeErr.Message)
				c.Abort()
				return
			}
			response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
			c.Abort()
			return
		}

		c.Request = c.Request.WithContext(authcontext.WithUserID(c.Request.Context(), userID))
		c.Next()
	}
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
