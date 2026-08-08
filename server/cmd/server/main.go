// Package main 是职途 后端的入口
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zhitu/server/internal/config"
	"github.com/zhitu/server/internal/database"
	"github.com/zhitu/server/internal/handlers"
	"github.com/zhitu/server/internal/routers"
	"github.com/zhitu/server/internal/services"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	log.Printf("[bootstrap] config loaded (mode=%s, port=%d)", cfg.Server.Mode, cfg.Server.Port)

	// 2. 初始化数据库
	db, err := database.Init(&cfg.Database)
	if err != nil {
		log.Fatalf("init database: %v", err)
	}
	log.Printf("[bootstrap] database ready: %s", cfg.Database.Path)

	// 3. 初始化 services
	jwtSvc := services.NewJWTService(&cfg.JWT)
	authSvc := services.NewAuthService(db, &cfg.Admin)
	llmSvc := services.NewLLMService(&cfg.LLM)
	if llmSvc.IsConfigured() {
		log.Printf("[bootstrap] llm ready (model=%s, base=%s)", cfg.LLM.ChatModel, cfg.LLM.BaseURL)
	} else {
		log.Printf("[bootstrap] WARNING: llm not configured, ai features will be unavailable")
	}

	// 4. 初始化 handlers
	authHandler := handlers.NewAuthHandler(authSvc, jwtSvc)
	profileSvc := services.NewProfileService(db, llmSvc, &cfg.Storage)
	profileHandler := handlers.NewProfileHandler(profileSvc)
	resumeSvc := services.NewResumeService(db, profileSvc)
	resumeAISvc := services.NewResumeAIService(db, llmSvc, resumeSvc, profileSvc)
	resumeHandler := handlers.NewResumeHandler(resumeSvc, resumeAISvc)
	copilotSvc := services.NewResumeCopilotService(llmSvc, resumeSvc, profileSvc)
	copilotHandler := handlers.NewCopilotHandler(copilotSvc)
	browserStateSvc := services.NewBrowserStateService(db)
	browserStateHandler := handlers.NewBrowserStateHandler(browserStateSvc)
	interviewSvc := services.NewInterviewService(db, llmSvc, profileSvc, &cfg.Storage)
	interviewHandler := handlers.NewInterviewHandler(interviewSvc)
	deliverySvc := services.NewDeliveryService(db)
	deliveryHandler := handlers.NewDeliveryHandler(deliverySvc)
	adminSvc := services.NewAdminService(db)
	adminHandler := handlers.NewAdminHandler(adminSvc)

	// 5. 初始化路由
	engine := routers.New(routers.Deps{
		Config:              cfg,
		DB:                  db,
		JWTService:          jwtSvc,
		AuthService:         authSvc,
		LLMService:          llmSvc,
		AuthHandler:         authHandler,
		ProfileHandler:      profileHandler,
		ResumeHandler:       resumeHandler,
		CopilotHandler:      copilotHandler,
		BrowserStateHandler: browserStateHandler,
		InterviewHandler:    interviewHandler,
		DeliveryHandler:     deliveryHandler,
		AdminHandler:        adminHandler,
	})

	// 6. 启动 HTTP 服务（支持优雅关闭）
	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:     engine,
		ReadTimeout: 15 * time.Second,
		// AI/SSE handlers flush an initial status event before the model finishes.
		// Keep the connection writable for the configured upstream model timeout;
		// otherwise responses taking more than 15 seconds are silently truncated.
		WriteTimeout: 5 * time.Minute,
	}

	go func() {
		log.Printf("[server] listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// 7. 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[server] shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[server] forced shutdown: %v", err)
	}

	// 关闭数据库连接
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	log.Println("[server] exited")
}
