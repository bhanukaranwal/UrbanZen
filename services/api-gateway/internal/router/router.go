package router

import (
	"database/sql"
	"net/http"

	"github.com/bhanukaranwal/UrbanZen/services/api-gateway/internal/auth"
	"github.com/bhanukaranwal/UrbanZen/services/api-gateway/internal/config"
	"github.com/bhanukaranwal/UrbanZen/services/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	_ "github.com/lib/pq"
)

// SetupRouter configures and returns the main router
func SetupRouter(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	r := gin.New()

	// Initialize database connections
	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	
	// Configure database connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.URL[8:], // Remove redis:// prefix
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize auth service
	authService := auth.NewAuthService(db, redisClient, cfg.JWT.Secret, logger)

	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Metrics())
	r.Use(middleware.CORS(cfg.Security.AllowedOrigins))
	r.Use(middleware.Security())
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "api-gateway",
			"version":   "1.0.0",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	// Metrics endpoint for Prometheus
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := v1.Group("/public")
		{
			setupPublicRoutes(public, cfg, logger)
		}

		// Authentication routes
		auth := v1.Group("/auth")
		{
			setupAuthRoutes(auth, authService, cfg, logger)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(authService.JWTMiddleware())
		{
			// Device management routes
			devices := protected.Group("/devices")
			{
				setupDeviceRoutes(devices, authService, cfg, logger)
			}

			// Data ingestion routes
			data := protected.Group("/data")
			{
				setupDataRoutes(data, authService, cfg, logger)
			}

			// Analytics routes
			analytics := protected.Group("/analytics")
			{
				setupAnalyticsRoutes(analytics, authService, cfg, logger)
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			{
				setupNotificationRoutes(notifications, authService, cfg, logger)
			}

			// User management routes
			users := protected.Group("/users")
			{
				setupUserRoutes(users, authService, cfg, logger)
			}

			// Billing routes
			billing := protected.Group("/billing")
			{
				setupBillingRoutes(billing, authService, cfg, logger)
			}

			// Reporting routes
			reports := protected.Group("/reports")
			{
				setupReportingRoutes(reports, authService, cfg, logger)
			}
		}

		// Admin routes (admin role required)
		admin := v1.Group("/admin")
		admin.Use(authService.JWTMiddleware())
		admin.Use(authService.RoleMiddleware("admin", "super_admin"))
		{
			setupAdminRoutes(admin, authService, cfg, logger)
		}

		// Service-to-service routes (API key required)
		internal := v1.Group("/internal")
		internal.Use(authService.ValidateAPIKey())
		{
			setupInternalRoutes(internal, cfg, logger)
		}
	}

	return r
}

// setupPublicRoutes configures public API routes
func setupPublicRoutes(router *gin.RouterGroup, cfg *config.Config, logger *zap.Logger) {
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "operational",
			"services": gin.H{
				"api_gateway":      "healthy",
				"device_mgmt":      "healthy",
				"data_ingestion":   "healthy",
				"analytics":        "healthy",
				"notification":     "healthy",
				"user_mgmt":        "healthy",
				"billing":          "healthy",
				"reporting":        "healthy",
			},
		})
	})

	router.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "UrbanZen IoT Smart City Platform",
			"version":     "1.0.0",
			"description": "Government-Grade IoT Smart City Management Platform",
			"contact":     "api-support@urbanzen.gov.in",
			"docs":        "/swagger/index.html",
		})
	})
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(router *gin.RouterGroup, authService *auth.AuthService, cfg *config.Config, logger *zap.Logger) {
	router.POST("/login", func(c *gin.Context) {
		// Implementation will be added when user management service is ready
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Authentication endpoint - implementation pending"})
	})

	router.POST("/logout", authService.JWTMiddleware(), func(c *gin.Context) {
		// Implementation will be added when user management service is ready
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Logout endpoint - implementation pending"})
	})

	router.POST("/refresh", authService.JWTMiddleware(), func(c *gin.Context) {
		// Implementation will be added when user management service is ready
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Token refresh endpoint - implementation pending"})
	})
}

// Placeholder route setup functions for microservices
func setupDeviceRoutes(router *gin.RouterGroup, authService *auth.AuthService, cfg *config.Config, logger *zap.Logger) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "device-management", "endpoint": "list-devices"})
	})
}

func setupDataRoutes(router *gin.RouterGroup, authService *auth.AuthService, cfg *config.Config, logger *zap.Logger) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "data-ingestion", "endpoint": "data-streams"})
	})
}

func setupAnalyticsRoutes(router *gin.RouterGroup, authService *auth.AuthService, cfg *config.Config, logger *zap.Logger) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "analytics", "endpoint": "analytics-data"})
	})
}

func setupNotificationRoutes(router *gin.RouterGroup, authService *auth.AuthService, cfg *config.Config, logger *zap.Logger) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "notification", "endpoint": "notifications"})
	})
}

func setupUserRoutes(router *gin.RouterGroup, authService *auth.AuthService, cfg *config.Config, logger *zap.Logger) {
	router.GET("/profile", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "user-management", "endpoint": "user-profile"})
	})
}

func setupBillingRoutes(router *gin.RouterGroup, authService *auth.AuthService, cfg *config.Config, logger *zap.Logger) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "billing", "endpoint": "billing-data"})
	})
}

func setupReportingRoutes(router *gin.RouterGroup, authService *auth.AuthService, cfg *config.Config, logger *zap.Logger) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "reporting", "endpoint": "reports"})
	})
}

func setupAdminRoutes(router *gin.RouterGroup, authService *auth.AuthService, cfg *config.Config, logger *zap.Logger) {
	router.GET("/stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "admin", "endpoint": "system-stats"})
	})
}

func setupInternalRoutes(router *gin.RouterGroup, cfg *config.Config, logger *zap.Logger) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "internal": true})
	})
}