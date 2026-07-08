package router

import (
	"github.com/gin-gonic/gin"

	"course_agent_backend/internal/handler"
	"course_agent_backend/internal/middleware"
)

func New(userHandler *handler.UserHandler, auth *middleware.AuthMiddleware, mode string) *gin.Engine {
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})

	api := r.Group("/api")
	users := api.Group("/users")
	{
		users.POST("/register", userHandler.Register)
		users.POST("/login", userHandler.Login)
		users.POST("/logout", auth.RequireAuth(), userHandler.Logout)
		users.GET("/me", auth.RequireAuth(), userHandler.Me)
	}

	return r
}
