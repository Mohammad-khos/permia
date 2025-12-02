package main

import (
	"log"

	"Permia/api-gateway/internal/config"
	"Permia/api-gateway/internal/handler"
	"Permia/api-gateway/internal/middleware"
	"Permia/api-gateway/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Initialize Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Load Configuration
	cfg, err := config.LoadConfig("../../../deployment/.env")
	if err != nil {
		sugar.Fatalf("Failed to load configuration: %v", err)
	}

	sugar.Infof("🌉 API Gateway starting on port %s", cfg.Port)
	sugar.Infof("Core Service: %s", cfg.CoreServiceURL)
	sugar.Infof("Bot Service: %s", cfg.BotServiceURL)

	// Set Gin mode based on environment
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize Router
	router := gin.New()

	// Apply Global Middleware
	router.Use(middleware.LoggingMiddleware(sugar))
	router.Use(middleware.CORSMiddleware(cfg.AllowOrigins))
	router.Use(middleware.AuthenticationMiddleware(sugar))
	router.Use(middleware.TraefikLabelMiddleware())
	router.Use(gin.Recovery())

	// Initialize Services
	proxyService := service.NewProxyService(cfg.CoreServiceURL, cfg.BotServiceURL, sugar)
	apiHandler := handler.NewHandler(proxyService, sugar)

	// Health Check Endpoint (public)
	router.GET("/health", apiHandler.Health)
	router.GET("/api/v1/health", apiHandler.Health)

	// Core Service Routes
	coreGroup := router.Group("/api/v1")
	{
		// Public routes
		coreGroup.GET("/products", apiHandler.ProxyCoreAPI)
		coreGroup.POST("/auth/login", apiHandler.ProxyCoreAPI)
		coreGroup.POST("/auth/register", apiHandler.ProxyCoreAPI)

		// Protected routes
		coreGroup.GET("/profile", apiHandler.ProxyCoreAPI)
		coreGroup.GET("/orders", apiHandler.ProxyCoreAPI)
		coreGroup.POST("/orders", apiHandler.ProxyCoreAPI)
		coreGroup.POST("/payments", apiHandler.ProxyCoreAPI)
		coreGroup.GET("/payments/:id", apiHandler.ProxyCoreAPI)
		coreGroup.POST("/admin/*path", apiHandler.ProxyCoreAPI)
	}

	// 404 Handler
	router.NoRoute(apiHandler.NotFound)

	// Start Server
	if err := router.Run(":" + cfg.Port); err != nil {
		sugar.Fatalf("Failed to start server: %v", err)
	}
}
