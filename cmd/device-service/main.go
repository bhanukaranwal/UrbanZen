package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/bhanukaranwal/urbanzen/internal/device"
	"github.com/bhanukaranwal/urbanzen/internal/config"
	"github.com/bhanukaranwal/urbanzen/pkg/logger"
	"github.com/bhanukaranwal/urbanzen/pkg/database"
	"github.com/bhanukaranwal/urbanzen/pkg/kafka"
)

func main() {
	// Initialize logger
	log := logger.New("device-service")
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}
	
	// Initialize database connections
	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL", "error", err)
	}
	defer db.Close()
	
	tsdb, err := database.NewTimescaleDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to TimescaleDB", "error", err)
	}
	defer tsdb.Close()
	
	// Initialize Kafka producer and consumer
	producer, err := kafka.NewProducer(cfg.Kafka.Brokers)
	if err != nil {
		log.Fatal("Failed to create Kafka producer", "error", err)
	}
	defer producer.Close()
	
	consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, "device-service-group")
	if err != nil {
		log.Fatal("Failed to create Kafka consumer", "error", err)
	}
	defer consumer.Close()
	
	// Initialize device service
	deviceService := device.NewService(db, tsdb, producer, consumer, log)
	
	// Start the service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go deviceService.Start(ctx)
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Info("Shutting down device service...")
	cancel()
}