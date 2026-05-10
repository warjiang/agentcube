package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/volcano-sh/agentcube/pkg/kernelhub"
)

type Config struct {
	Port          int
	Workspace     string
	MemoryLimit   int64
	MaxSessions   int
	SessionTTL    time.Duration
	StateStore    string
	RedisAddr     string
	RedisPassword string
}

type Server struct {
	engine *gin.Engine
	config Config
	manager *kernelhub.Manager
	auth   *kernelhub.AuthManager
}

func NewServer(config Config) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())

	s := &Server{
		engine: engine,
		config: config,
	}

	s.auth = kernelhub.NewAuthManager()
	if err := s.auth.LoadPublicKeyFromEnv(); err != nil {
		return nil, fmt.Errorf("failed to load auth public key: %v", err)
	}

	stateStore := kernelhub.NewMemoryStateStore()

	s.manager = kernelhub.NewManager(
		kernelhub.WithMemoryLimit(config.MemoryLimit),
		kernelhub.WithStateStore(stateStore),
		kernelhub.WithMaxSessions(config.MaxSessions),
		kernelhub.WithSessionTTL(config.SessionTTL),
	)

	s.registerRoutes()

	return s, nil
}

func (s *Server) registerRoutes() {
	s.engine.GET("/health", s.healthHandler)

	handlers := kernelhub.NewHandlers(s.manager)

	api := s.engine.Group("/api")
	api.Use(s.auth.AuthMiddleware())
	{
		api.POST("/sessions", handlers.CreateSession)
		api.GET("/sessions/:session_id", handlers.GetSession)
		api.DELETE("/sessions/:session_id", handlers.DeleteSession)
		api.GET("/sessions", handlers.ListSessions)

		api.POST("/sessions/:session_id/start", handlers.StartKernel)
		api.POST("/sessions/:session_id/execute", handlers.ExecuteCode)
		api.POST("/sessions/:session_id/interrupt", handlers.InterruptKernel)
		api.POST("/sessions/:session_id/restart", handlers.RestartKernel)
		api.POST("/sessions/:session_id/stop", handlers.StopKernel)
		api.GET("/sessions/:session_id/state", handlers.GetState)
		api.POST("/sessions/:session_id/state", handlers.RestoreState)
		api.GET("/sessions/:session_id/kernel", handlers.GetKernelInfo)
	}
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	klog.Infof("KernelHub server starting on %s", addr)

	return s.engine.Run(addr)
}

func (s *Server) healthHandler(c *gin.Context) {
	sessions := s.manager.ListSessions()

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "KernelHub",
		"version": "0.1.0",
		"sessions": len(sessions),
	})
}

func (s *Server) StartCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.manager.CleanupIdle(ctx)
		}
	}
}
