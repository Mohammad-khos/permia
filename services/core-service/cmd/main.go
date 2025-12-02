package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"Permia/core-service/internal/config"
	"Permia/core-service/internal/handler"
	"Permia/core-service/internal/infrastructure/brocard"
	"Permia/core-service/internal/infrastructure/zarinpal"
	"Permia/core-service/internal/repository"
	"Permia/core-service/internal/service"
	"Permia/core-service/migration"
	"Permia/pkg/logger"
	"Permia/pkg/response"

	"Permia/core-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. راه اندازی Logger
	appLogger := logger.NewLogger()
	defer appLogger.Sync()

	// 2. لود کردن تنظیمات (از پوشه deployment)
	cfg, err := config.Load()
	if err != nil {
		appLogger.Fatal("Failed to load configuration", zap.Error(err))
	}
	appLogger.Info("Configuration loaded", zap.String("env", cfg.AppEnv))

	// 3. اتصال به دیتابیس
	// نکته: تابع NewPostgresDB هنوز از os.Getenv استفاده می‌کند که مشکلی نیست
	// چون config.Load مقادیر .env را در سیستم ست کرده است.
	db, err := repository.NewPostgresDB()
	if err != nil {
		// لاگ کردن اطلاعات دیتابیس برای دیباگ (بدون پسورد)
		appLogger.Fatal("Database connection failed",
			zap.String("host", cfg.Database.Host),
			zap.String("port", cfg.Database.Port),
			zap.Error(err),
		)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 4. اجرای مایگریشن‌ها
	migration.Up(db)

	// 5. تزریق وابستگی‌ها (Dependency Injection)
	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	//--- VCC Provider ---
	brocardClient := brocard.NewBrocardClient()

	// --- Services ---
	// Initialize Zarinpal client
	zarinpalClient := zarinpal.NewZarinpalClient(appLogger)

	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(orderRepo, userRepo, productRepo, accountRepo, brocardClient, db)
	paymentService := service.NewPaymentService(paymentRepo, orderRepo, userRepo, zarinpalClient, db, appLogger)
	adminService := service.NewAdminService(accountRepo, orderRepo, productRepo, db, appLogger)

	// --- Handlers ---
	userHandler := handler.NewUserHandler(userService)
	productHandler := handler.NewProductHandler(productService)
	orderHandler := handler.NewOrderHandler(orderService, userService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	adminHandler := handler.NewAdminHandler(adminService, productService)

	// 6. تنظیمات Gin
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// میدلورهای اختصاصی
	r.Use(GinZapLogger(appLogger), GinRecovery(appLogger, true))

	// 7. تعریف روت‌ها
	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			response.Success(c, "OK", "Core Service is healthy 🚀")
		})

		// User Routes
		users := api.Group("/users")
		{
			users.POST("/auth", userHandler.AuthUser)
			users.GET("/:id/balance", userHandler.GetBalance)
		}

		// Product Routes
		api.GET("/products", productHandler.ListProducts)

		// Payment Routes
		payment := api.Group("/payment")
		{
			payment.POST("/charge", paymentHandler.ChargePayment)
			payment.GET("/verify", paymentHandler.VerifyPayment)
		}

		// Order Routes
		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders", orderHandler.GetOrders)
		api.GET("/orders/:id", orderHandler.GetOrderByID)

		// Admin Routes
		admin := api.Group("/admin")
		admin.Use(middleware.AdminAuth(appLogger)) // Apply security middleware
		{
			// Inventory Management
			admin.POST("/inventory", adminHandler.CreateInventory)
			admin.GET("/inventory/stats", adminHandler.GetInventoryStats)

			// Order Management
			admin.GET("/orders", adminHandler.GetOrderStats)
			admin.POST("/orders/:id/complete", adminHandler.CompleteOrder)

			// Product Management
			admin.POST("/products", adminHandler.CreateProduct)
			admin.PUT("/products/:sku", adminHandler.UpdateProduct)
		}
	}

	// 8. اجرای سرور
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	appLogger.Info("Starting Core Service", zap.String("address", serverAddr))

	if err := r.Run(serverAddr); err != nil {
		appLogger.Fatal("Server startup failed", zap.Error(err))
	}
}

// =============================================================================
// Middleware Functions
// =============================================================================

func GinZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				logger.Error(e)
			}
		} else {
			logger.Info(path,
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.Duration("latency", latency),
			)
		}
	}
}

func GinRecovery(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					c.Error(err.(error))
					c.Abort()
					return
				}

				if stack {
					logger.Error("[Recovery from panic]",
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					logger.Error("[Recovery from panic]",
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}

				response.Error(c, http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		c.Next()
	}
}
