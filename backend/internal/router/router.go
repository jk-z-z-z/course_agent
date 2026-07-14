package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"course_agent_backend/internal/handler"
	"course_agent_backend/internal/middleware"
)

func New(userHandler *handler.UserHandler, courseHandler *handler.CourseHandler, materialHandler *handler.MaterialHandler, agentHandler *handler.AgentHandler, studyPlanHandler *handler.StudyPlanHandler, auth *middleware.AuthMiddleware, mode string) *gin.Engine {
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

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

	courses := api.Group("/courses")
	courses.Use(auth.RequireAuth())
	{
		courses.GET("/discover", courseHandler.ListDiscoverableCourses)
		courses.POST("", courseHandler.CreateCourse)
		courses.GET("", courseHandler.ListCourses)
		courses.GET("/:courseId", courseHandler.GetCourse)
		courses.PUT("/:courseId", courseHandler.UpdateCourse)
		courses.DELETE("/:courseId", courseHandler.DeleteCourse)
		courses.POST("/:courseId/join", courseHandler.JoinCourse)
		courses.GET("/:courseId/members", courseHandler.ListMembers)
		courses.POST("/:courseId/members", courseHandler.AddMember)
		courses.PUT("/:courseId/members/:memberId", courseHandler.UpdateMember)
		courses.DELETE("/:courseId/members/:memberId", courseHandler.DeleteMember)
		courses.GET("/:courseId/materials/tree", materialHandler.GetTree)
		courses.POST("/:courseId/materials/folders", materialHandler.CreateFolder)
		courses.POST("/:courseId/materials/upload", materialHandler.UploadFile)
		courses.GET("/:courseId/materials/:nodeId", materialHandler.GetDetail)
		courses.GET("/:courseId/materials/:nodeId/preview", materialHandler.PreviewFile)
		courses.GET("/:courseId/materials/:nodeId/download", materialHandler.DownloadFile)
		courses.PUT("/:courseId/materials/:nodeId", materialHandler.UpdateNode)
		courses.DELETE("/:courseId/materials/:nodeId", materialHandler.DeleteNode)
		courses.GET("/:courseId/agent", agentHandler.GetAgent)
		courses.POST("/:courseId/agent/conversations", agentHandler.CreateConversation)
		courses.GET("/:courseId/agent/conversations", agentHandler.ListConversations)
		courses.GET("/:courseId/agent/conversations/:conversationId", agentHandler.GetConversation)
		courses.POST("/:courseId/agent/ask", agentHandler.Ask)
		courses.POST("/:courseId/agent/ask/stream", agentHandler.AskStream)
		courses.GET("/:courseId/study-plans", studyPlanHandler.List)
		courses.POST("/:courseId/study-plans/generate", studyPlanHandler.Generate)
		courses.GET("/:courseId/study-plans/:planId", studyPlanHandler.GetDetail)
		courses.PATCH("/:courseId/study-plans/:planId/items/:itemId", studyPlanHandler.UpdateItem)
	}

	return r
}
