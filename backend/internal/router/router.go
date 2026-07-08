package router

import (
	"net/http"

	"course_agent_backend/internal/handler"
	"course_agent_backend/internal/middleware"
)

func New(userHandler *handler.UserHandler, auth *middleware.AuthMiddleware) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /api/users/register", userHandler.Register)
	mux.HandleFunc("POST /api/users/login", userHandler.Login)
	mux.Handle("POST /api/users/logout", auth.RequireAuth(http.HandlerFunc(userHandler.Logout)))
	mux.Handle("GET /api/users/me", auth.RequireAuth(http.HandlerFunc(userHandler.Me)))
	return mux
}
