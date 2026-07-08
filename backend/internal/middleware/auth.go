package middleware

import (
	"net/http"
	"strings"

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

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		userID, err := m.service.ResolveToken(r.Context(), token)
		if err != nil {
			if codeErr, ok := err.(*apperrors.CodeError); ok {
				response.Fail(w, http.StatusUnauthorized, codeErr.Code, codeErr.Message)
				return
			}
			response.Fail(w, http.StatusInternalServerError, 50000, err.Error())
			return
		}

		next.ServeHTTP(w, r.WithContext(authcontext.WithUserID(r.Context(), userID)))
	})
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
