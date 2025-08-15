package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/bhanukaranwal/urbanzen/internal/notification"
	"github.com/bhanukaranwal/urbanzen/internal/config"
	"github.com/bhanukaranwal/urbanzen/pkg/logger"
	"github.com/bhanukaranwal/urbanzen/pkg/database"
	"github.com/bhanukaranwal/urbanzen/pkg/kafka"
)

func main() {
	// Initialize logger
	log := logger.New("notification-service")
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}
	
	// Initialize database connection
	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()
	
	// Initialize Redis
	redis, err := database.NewRedis(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Redis", "error", err)
	}
	defer redis.Close()
	
	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, "notification-service-group")
	if err != nil {
		log.Fatal("Failed to create Kafka consumer", "error", err)
	}
	defer consumer.Close()
	
	// Initialize notification service
	notificationService := notification.NewService(db, redis, consumer, cfg, log)
	
	// Start the service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go notificationService.Start(ctx)
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Info("Shutting down notification service...")
	cancel()
}