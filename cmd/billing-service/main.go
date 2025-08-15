package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/gin-gonic/gin"
	"github.com/bhanukaranwal/urbanzen/internal/billing"
	"github.com/bhanukaranwal/urbanzen/internal/config"
	"github.com/bhanukaranwal/urbanzen/internal/middleware"
	"github.com/bhanukaranwal/urbanzen/pkg/logger"
	"github.com/bhanukaranwal/urbanzen/pkg/database"
)

func main() {
	// Initialize logger
	log := logger.New("billing-service")
	
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
	
	redis, err := database.NewRedis(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Redis", "error", err)
	}
	defer redis.Close()
	
	// Initialize billing service
	billingService := billing.NewService(db, tsdb, redis, cfg, log)
	
	// Setup HTTP router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.CORS())
	router.Use(middleware.Security())
	
	// Setup routes
	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthRequired())
	{
		bills := v1.Group("/bills")
		{
			bills.GET("", billingService.GetUserBills)
			bills.GET("/:id", billingService.GetBill)
			bills.POST("/:id/pay", billingService.ProcessPayment)
			bills.GET("/:id/download", billingService.DownloadBill)
		}
		
		consumption := v1.Group("/consumption")
		{
			consumption.GET("/water", billingService.GetWaterConsumption)
			consumption.GET("/electricity", billingService.GetElectricityConsumption)
			consumption.GET("/analytics", billingService.GetConsumptionAnalytics)
		}
		
		admin := v1.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.POST("/generate-bills", billingService.GenerateBills)
			admin.GET("/billing-reports", billingService.GetBillingReports)
			admin.POST("/rates", billingService.UpdateRates)
		}
	}
	
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	
	// Start server
	srv := &http.Server{
		Addr:    ":8082",
		Handler: router,
	}
	
	go func() {
		log.Info("Starting billing service", "port", 8082)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Info("Shutting down billing service...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}
	
	log.Info("Billing service exited")
}