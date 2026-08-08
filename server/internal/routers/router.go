// Package routers 注册所有 HTTP 路由
package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhitu/server/internal/config"
	"github.com/zhitu/server/internal/handlers"
	"github.com/zhitu/server/internal/middleware"
	"github.com/zhitu/server/internal/services"
	"github.com/zhitu/server/internal/utils"
	"gorm.io/gorm"
)

// Deps 路由初始化所需依赖
type Deps struct {
	Config           *config.Config
	DB               *gorm.DB
	JWTService       *services.JWTService
	AuthService      *services.AuthService
	LLMService       *services.LLMService
	AuthHandler      *handlers.AuthHandler
	ProfileHandler   *handlers.ProfileHandler
	ResumeHandler    *handlers.ResumeHandler
	CopilotHandler   *handlers.CopilotHandler
	InterviewHandler *handlers.InterviewHandler
	DeliveryHandler  *handlers.DeliveryHandler
	AdminHandler     *handlers.AdminHandler
}

// New 构造 gin 引擎并注册全部路由
func New(deps Deps) *gin.Engine {
	gin.SetMode(deps.Config.Server.Mode)

	r := gin.New()
	// 全局中间件
	r.Use(middleware.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORS(deps.Config.Server.AllowOrigins))

	// 健康检查
	r.GET("/health", healthCheck)
	r.GET("/", healthCheck)

	// API v1 根
	api := r.Group("/api")

	// ---------- 认证模块（无需登录） ----------
	auth := api.Group("/auth")
	{
		auth.POST("/register", deps.AuthHandler.Register)
		auth.POST("/login", deps.AuthHandler.Login)
		auth.POST("/admin/login", deps.AuthHandler.AdminLogin)

		// 需登录的认证相关接口
		authAuthed := auth.Group("")
		authAuthed.Use(middleware.JWTAuth(deps.JWTService))
		{
			authAuthed.POST("/change-password", deps.AuthHandler.ChangePassword)
			authAuthed.GET("/me", deps.AuthHandler.Me)
		}
	}

	// ---------- 管理员模块（需管理员身份） ----------
	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth(deps.JWTService), middleware.RequireAdmin())
	{
		admin.GET("/ping", func(c *gin.Context) {
			email, _ := c.Get(middleware.ContextEmail)
			utils.OK(c, gin.H{"role": "admin", "email": email})
		})

		// 仪表盘
		admin.GET("/stats", deps.AdminHandler.GetStats)

		// 用户管理
		admin.GET("/users", deps.AdminHandler.ListUsers)
		admin.GET("/users/:id", deps.AdminHandler.GetUser)
		admin.PATCH("/users/:id/status", deps.AdminHandler.ToggleUserStatus)
		admin.POST("/users/:id/reset-password", deps.AdminHandler.ResetPassword)

		// 投递管理
		admin.GET("/deliveries", deps.AdminHandler.ListDeliveries)
		admin.GET("/deliveries/funnel", deps.AdminHandler.GetFunnel)
	}

	// ---------- 业务模块（需登录，普通用户与管理员均可访问） ----------
	v1 := api.Group("/v1")
	v1.Use(middleware.JWTAuth(deps.JWTService))
	{
		v1.GET("/_ping", func(c *gin.Context) {
			utils.OK(c, gin.H{"message": "auth ok", "module": "v1"})
		})

		// ---------- 用户档案 ----------
		profile := v1.Group("/profile")
		{
			profile.GET("", deps.ProfileHandler.GetProfile)
			profile.PUT("", deps.ProfileHandler.UpdateProfile)
			profile.GET("/completion", deps.ProfileHandler.GetCompletion)
			profile.POST("/parse-resume", deps.ProfileHandler.ParseResume)

			// 子资源 CRUD
			registerSubRoutes(profile, "educations", deps.ProfileHandler.ListEducations, deps.ProfileHandler.CreateEducation, deps.ProfileHandler.UpdateEducation, deps.ProfileHandler.DeleteEducation)
			registerSubRoutes(profile, "works", deps.ProfileHandler.ListWorks, deps.ProfileHandler.CreateWork, deps.ProfileHandler.UpdateWork, deps.ProfileHandler.DeleteWork)
			registerSubRoutes(profile, "projects", deps.ProfileHandler.ListProjects, deps.ProfileHandler.CreateProject, deps.ProfileHandler.UpdateProject, deps.ProfileHandler.DeleteProject)
			registerSubRoutes(profile, "skills", deps.ProfileHandler.ListSkills, deps.ProfileHandler.CreateSkill, deps.ProfileHandler.UpdateSkill, deps.ProfileHandler.DeleteSkill)
			registerSubRoutes(profile, "honors", deps.ProfileHandler.ListHonors, deps.ProfileHandler.CreateHonor, deps.ProfileHandler.UpdateHonor, deps.ProfileHandler.DeleteHonor)
			registerSubRoutes(profile, "practices", deps.ProfileHandler.ListPractices, deps.ProfileHandler.CreatePractice, deps.ProfileHandler.UpdatePractice, deps.ProfileHandler.DeletePractice)
		}

		// ---------- 简历模块 ----------
		resumes := v1.Group("/resumes")
		{
			resumes.GET("", deps.ResumeHandler.List)
			resumes.POST("", deps.ResumeHandler.Create)
			resumes.GET("/:id", deps.ResumeHandler.Get)
			resumes.PUT("/:id", deps.ResumeHandler.Update)
			resumes.DELETE("/:id", deps.ResumeHandler.Delete)

			// 版本管理
			resumes.GET("/:id/versions", deps.ResumeHandler.ListVersions)
			resumes.POST("/:id/versions", deps.ResumeHandler.CreateVersion)
			resumes.GET("/:id/versions/:vid", deps.ResumeHandler.GetVersion)
			resumes.POST("/:id/rollback/:vid", deps.ResumeHandler.RollbackVersion)

			// AI 操作
			resumes.POST("/:id/ai/generate", deps.ResumeHandler.AIGenerate)
			resumes.POST("/:id/ai/polish", deps.ResumeHandler.AIPolish)
			resumes.POST("/:id/ai/score", deps.ResumeHandler.AIScore)
			resumes.POST("/:id/ai/jd-match", deps.ResumeHandler.AIJDMatch)

			// 同步档案
			resumes.POST("/:id/sync-profile", deps.ResumeHandler.SyncProfile)
		}

		// ---------- 求职 Copilot ----------
		copilot := v1.Group("/copilot")
		{
			copilot.POST("/chat", deps.CopilotHandler.Chat)
			copilot.POST("/apply", deps.CopilotHandler.Apply)
		}

		// ---------- 模拟面试 ----------
		interviews := v1.Group("/interviews")
		{
			interviews.GET("", deps.InterviewHandler.List)
			interviews.POST("", deps.InterviewHandler.Create)
			interviews.GET("/:id", deps.InterviewHandler.Get)
			interviews.POST("/:id/start", deps.InterviewHandler.Start)
			interviews.PATCH("/:id/mode", deps.InterviewHandler.SetMode)
			interviews.POST("/:id/resume", deps.InterviewHandler.AttachResume)
			interviews.POST("/:id/messages", deps.InterviewHandler.SendMessage)
			interviews.POST("/:id/transcribe", deps.InterviewHandler.TranscribeVoice)
			interviews.POST("/:id/voice", deps.InterviewHandler.SendVoice)
			interviews.GET("/:id/tts/:msgId", deps.InterviewHandler.GetTTS)
			interviews.POST("/:id/end", deps.InterviewHandler.End)
			interviews.POST("/:id/cancel", deps.InterviewHandler.Cancel)
			interviews.GET("/:id/report", deps.InterviewHandler.GetReport)
			interviews.GET("/:id/scores", deps.InterviewHandler.GetScores)
		}

		// ---------- 投递看板 ----------
		// 注意：stats / funnel 必须注册在 :id 之前，避免被当作 id 匹配
		deliveries := v1.Group("/deliveries")
		{
			deliveries.GET("", deps.DeliveryHandler.List)
			deliveries.POST("", deps.DeliveryHandler.Create)
			deliveries.GET("/stats", deps.DeliveryHandler.GetStats)
			deliveries.GET("/funnel", deps.DeliveryHandler.GetFunnel)

			deliveries.GET("/:id", deps.DeliveryHandler.Get)
			deliveries.PUT("/:id", deps.DeliveryHandler.Update)
			deliveries.DELETE("/:id", deps.DeliveryHandler.Delete)
			deliveries.PATCH("/:id/status", deps.DeliveryHandler.ChangeStatus)

			// 面试轮次
			deliveries.GET("/:id/rounds", deps.DeliveryHandler.ListRounds)
			deliveries.POST("/:id/rounds", deps.DeliveryHandler.CreateRound)
			deliveries.PUT("/:id/rounds/:rid", deps.DeliveryHandler.UpdateRound)
			deliveries.DELETE("/:id/rounds/:rid", deps.DeliveryHandler.DeleteRound)

			// HR 反馈
			deliveries.GET("/:id/feedbacks", deps.DeliveryHandler.ListFeedbacks)
			deliveries.POST("/:id/feedbacks", deps.DeliveryHandler.CreateFeedback)
			deliveries.DELETE("/:id/feedbacks/:fid", deps.DeliveryHandler.DeleteFeedback)
		}
	}

	// 静态文件服务：/static/* 映射到存储目录，用于访问上传的音频、TTS、简历文件
	r.Static("/static", deps.Config.Storage.BaseDir)

	// 404 兜底
	r.NoRoute(func(c *gin.Context) {
		utils.NotFound(c, "route not found")
	})
	r.NoMethod(func(c *gin.Context) {
		utils.Fail(c, http.StatusMethodNotAllowed, utils.CodeInvalidParams, "method not allowed")
	})

	return r
}

// healthCheck 健康检查端点
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "zhitu-server",
	})
}

// registerSubRoutes 注册子资源的标准 CRUD 路由
//
//	GET    /{resource}      列表
//	POST   /{resource}      创建
//	PUT    /{resource}/:id  更新
//	DELETE /{resource}/:id  删除
func registerSubRoutes(g *gin.RouterGroup, resource string, list, create, update, del gin.HandlerFunc) {
	g.GET("/"+resource, list)
	g.POST("/"+resource, create)
	g.PUT("/"+resource+"/:id", update)
	g.DELETE("/"+resource+"/:id", del)
}
