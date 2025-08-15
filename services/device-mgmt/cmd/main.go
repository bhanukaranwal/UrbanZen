package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bhanukaranwal/UrbanZen/services/device-mgmt/internal/config"
	"github.com/bhanukaranwal/UrbanZen/services/device-mgmt/internal/handlers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title Device Management Service
// @version 1.0
// @description IoT Device Management Service for UrbanZen Platform
// @host localhost:8081
// @BasePath /api/v1

func main() {
	// Initialize configuration
	cfg := config.Load()

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize handlers
	deviceHandler := handlers.NewDeviceHandler(cfg, logger)

	// Setup router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "device-management",
			"version": "1.0.0",
		})
	})

	// API routes
	v1 := r.Group("/api/v1")
	{
		devices := v1.Group("/devices")
		{
			devices.GET("/", deviceHandler.ListDevices)
			devices.POST("/", deviceHandler.CreateDevice)
			devices.GET("/:id", deviceHandler.GetDevice)
			devices.PUT("/:id", deviceHandler.UpdateDevice)
			devices.DELETE("/:id", deviceHandler.DeleteDevice)
			devices.POST("/:id/command", deviceHandler.SendCommand)
			devices.GET("/:id/status", deviceHandler.GetDeviceStatus)
			devices.GET("/:id/telemetry", deviceHandler.GetDeviceTelemetry)
		}

		// Device types
		deviceTypes := v1.Group("/device-types")
		{
			deviceTypes.GET("/", deviceHandler.ListDeviceTypes)
			deviceTypes.POST("/", deviceHandler.CreateDeviceType)
		}

		// Firmware management
		firmware := v1.Group("/firmware")
		{
			firmware.GET("/", deviceHandler.ListFirmware)
			firmware.POST("/", deviceHandler.UploadFirmware)
			firmware.POST("/:id/deploy", deviceHandler.DeployFirmware)
		}
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		logger.Info("Starting Device Management service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}