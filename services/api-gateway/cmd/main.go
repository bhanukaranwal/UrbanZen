package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bhanukaranwal/UrbanZen/services/api-gateway/internal/config"
	"github.com/bhanukaranwal/UrbanZen/services/api-gateway/internal/router"
	"github.com/bhanukaranwal/UrbanZen/services/api-gateway/pkg/logger"
	"github.com/gin-gonic/gin"
)

// @title UrbanZen API Gateway
// @version 1.0
// @description Government-Grade IoT Smart City Management Platform API Gateway
// @termsOfService https://urbanzen.gov.in/terms

// @contact.name API Support
// @contact.url https://urbanzen.gov.in/support
// @contact.email api-support@urbanzen.gov.in

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Initialize configuration
	cfg := config.Load()

	// Initialize logger
	zapLogger := logger.NewZapLogger(cfg.LogLevel)
	defer zapLogger.Sync()

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	r := router.SetupRouter(cfg, zapLogger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		zapLogger.Sugar().Infof("Starting API Gateway server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Sugar().Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Sugar().Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Sugar().Fatalf("Server forced to shutdown: %v", err)
	}

	zapLogger.Sugar().Info("Server exited")
}