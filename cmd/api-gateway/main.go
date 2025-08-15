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

    "github.com/gin-gonic/gin"
    "github.com/bhanukaranwal/UrbanZen/internal/config"
    "github.com/bhanukaranwal/UrbanZen/internal/gateway"
    "github.com/bhanukaranwal/UrbanZen/internal/middleware"
    "github.com/bhanukaranwal/UrbanZen/pkg/logger"
)

func main() {
    // Initialize logger
    logger := logger.New("api-gateway")
    
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }

    // Initialize Gin router
    if cfg.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }
    
    router := gin.New()
    
    // Add middlewares
    router.Use(gin.Recovery())
    router.Use(middleware.Logger(logger))
    router.Use(middleware.CORS(cfg))
    router.Use(middleware.Security())
    router.Use(middleware.RateLimiter(cfg))

    // Initialize gateway
    gw := gateway.New(cfg, logger)
    
    // Setup routes
    v1 := router.Group("/api/v1")
    {
        // Authentication routes
        auth := v1.Group("/auth")
        {
            auth.POST("/login", gw.Login)
            auth.POST("/logout", gw.Logout)
            auth.POST("/refresh", gw.RefreshToken)
            auth.GET("/me", middleware.AuthRequired(cfg), gw.GetProfile)
        }
        
        // Device management routes
        devices := v1.Group("/devices")
        devices.Use(middleware.AuthRequired(cfg))
        {
            devices.GET("", gw.ListDevices)
            devices.POST("", gw.CreateDevice)
            devices.GET("/:id", gw.GetDevice)
            devices.PUT("/:id", gw.UpdateDevice)
            devices.DELETE("/:id", gw.DeleteDevice)
        }
        
        // Utility services routes
        utilities := v1.Group("/utilities")
        utilities.Use(middleware.AuthRequired(cfg))
        {
            water := utilities.Group("/water")
            {
                water.GET("/consumption", gw.GetWaterConsumption)
                water.GET("/quality", gw.GetWaterQuality)
            }
            
            electricity := utilities.Group("/electricity")
            {
                electricity.GET("/consumption", gw.GetElectricityConsumption)
                electricity.GET("/grid-status", gw.GetGridStatus)
            }
        }
    }
    
    // Health check endpoint
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status":    "healthy",
            "timestamp": time.Now().Unix(),
            "version":   cfg.Version,
        })
    })
    
    // Setup HTTP server
    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
        Handler: router,
    }
    
    // Start server in a goroutine
    go func() {
        logger.Info("Starting API Gateway on port", cfg.Server.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Failed to start server:", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    logger.Info("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    logger.Info("Server exited")
}