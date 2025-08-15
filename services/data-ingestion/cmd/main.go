package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bhanukaranwal/UrbanZen/services/data-ingestion/internal/config"
	"github.com/bhanukaranwal/UrbanZen/services/data-ingestion/internal/handlers"
	"github.com/bhanukaranwal/UrbanZen/services/data-ingestion/internal/processors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title Data Ingestion Service
// @version 1.0
// @description Real-time data ingestion service for UrbanZen IoT Platform
// @host localhost:8082
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

	// Initialize processors
	kafkaProcessor := processors.NewKafkaProcessor(cfg, logger)
	timeseriesProcessor := processors.NewTimeseriesProcessor(cfg, logger)

	// Start background processors
	go kafkaProcessor.Start(context.Background())
	go timeseriesProcessor.Start(context.Background())

	// Initialize handlers
	dataHandler := handlers.NewDataHandler(cfg, logger, kafkaProcessor, timeseriesProcessor)

	// Setup router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "data-ingestion",
			"version": "1.0.0",
		})
	})

	// API routes
	v1 := r.Group("/api/v1")
	{
		// Data ingestion endpoints
		data := v1.Group("/data")
		{
			data.POST("/ingest", dataHandler.IngestData)
			data.POST("/batch", dataHandler.IngestBatch)
			data.GET("/streams", dataHandler.ListStreams)
			data.GET("/streams/:id/metrics", dataHandler.GetStreamMetrics)
		}

		// Real-time data endpoints
		realtime := v1.Group("/realtime")
		{
			realtime.GET("/device/:id", dataHandler.GetRealtimeData)
			realtime.GET("/metrics", dataHandler.GetRealtimeMetrics)
		}

		// Processing endpoints
		processing := v1.Group("/processing")
		{
			processing.GET("/status", dataHandler.GetProcessingStatus)
			processing.POST("/rules", dataHandler.CreateProcessingRule)
			processing.GET("/rules", dataHandler.ListProcessingRules)
		}
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		logger.Info("Starting Data Ingestion service", zap.String("port", cfg.Port))
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

	// Stop processors
	kafkaProcessor.Stop()
	timeseriesProcessor.Stop()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}