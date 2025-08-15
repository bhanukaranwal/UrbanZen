package gateway

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/bhanukaranwal/urbanzen/internal/config"
	"github.com/bhanukaranwal/urbanzen/internal/middleware"
	"github.com/bhanukaranwal/urbanzen/pkg/logger"
)

type Gateway struct {
	config *config.Config
	logger logger.Logger
}

func New(cfg *config.Config, log logger.Logger) *Gateway {
	return &Gateway{
		config: cfg,
		logger: log,
	}
}

func (g *Gateway) Login(c *gin.Context) {
	var loginReq struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual user authentication
	// For now, return a mock response
	if loginReq.Username == "admin" && loginReq.Password == "admin123" {
		token, err := middleware.GenerateToken("1", loginReq.Username, "admin", g.config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":       "1",
				"username": loginReq.Username,
				"role":     "admin",
			},
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
}

func (g *Gateway) Logout(c *gin.Context) {
	// TODO: Implement token blacklisting
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (g *Gateway) RefreshToken(c *gin.Context) {
	// TODO: Implement token refresh logic
	c.JSON(http.StatusOK, gin.H{"message": "Token refreshed"})
}

func (g *Gateway) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"id":       userID,
		"username": username,
		"role":     role,
	})
}

func (g *Gateway) ListDevices(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	deviceType := c.Query("type")

	// TODO: Implement actual device listing from database
	devices := []gin.H{
		{
			"id":        "device-001",
			"name":      "Water Sensor #1",
			"type":      "water_sensor",
			"status":    "active",
			"location":  gin.H{"latitude": 28.6139, "longitude": 77.2090},
			"last_seen": "2024-01-15T10:30:00Z",
		},
		{
			"id":        "device-002",
			"name":      "Smart Meter #1",
			"type":      "electricity_meter",
			"status":    "active",
			"location":  gin.H{"latitude": 28.6129, "longitude": 77.2080},
			"last_seen": "2024-01-15T10:29:00Z",
		},
	}

	// Filter by type if specified
	if deviceType != "" {
		filtered := []gin.H{}
		for _, device := range devices {
			if device["type"] == deviceType {
				filtered = append(filtered, device)
			}
		}
		devices = filtered
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(devices),
		},
	})
}

func (g *Gateway) CreateDevice(c *gin.Context) {
	var device struct {
		Name     string  `json:"name" binding:"required"`
		Type     string  `json:"type" binding:"required"`
		Latitude float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
	}

	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual device creation
	c.JSON(http.StatusCreated, gin.H{
		"id":       "device-new-001",
		"name":     device.Name,
		"type":     device.Type,
		"status":   "active",
		"location": gin.H{"latitude": device.Latitude, "longitude": device.Longitude},
		"message":  "Device created successfully",
	})
}

func (g *Gateway) GetDevice(c *gin.Context) {
	deviceID := c.Param("id")

	// TODO: Implement actual device retrieval
	c.JSON(http.StatusOK, gin.H{
		"id":        deviceID,
		"name":      "Water Sensor #1",
		"type":      "water_sensor",
		"status":    "active",
		"location":  gin.H{"latitude": 28.6139, "longitude": 77.2090},
		"last_seen": "2024-01-15T10:30:00Z",
		"metrics": gin.H{
			"flow_rate": 25.5,
			"pressure":  3.2,
			"ph_level":  7.1,
		},
	})
}

func (g *Gateway) UpdateDevice(c *gin.Context) {
	deviceID := c.Param("id")

	var updateReq struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual device update
	c.JSON(http.StatusOK, gin.H{
		"id":      deviceID,
		"message": "Device updated successfully",
	})
}

func (g *Gateway) DeleteDevice(c *gin.Context) {
	deviceID := c.Param("id")

	// TODO: Implement actual device deletion
	c.JSON(http.StatusOK, gin.H{
		"message": "Device " + deviceID + " deleted successfully",
	})
}

func (g *Gateway) GetWaterConsumption(c *gin.Context) {
	// TODO: Implement actual water consumption data
	c.JSON(http.StatusOK, gin.H{
		"daily_consumption":   245.5,
		"monthly_consumption": 7250.0,
		"unit":               "liters",
		"last_updated":       "2024-01-15T10:30:00Z",
	})
}

func (g *Gateway) GetWaterQuality(c *gin.Context) {
	// TODO: Implement actual water quality data
	c.JSON(http.StatusOK, gin.H{
		"ph_level":     7.1,
		"turbidity":    1.2,
		"chlorine":     0.5,
		"quality_index": 85,
		"status":       "good",
		"last_updated": "2024-01-15T10:30:00Z",
	})
}

func (g *Gateway) GetElectricityConsumption(c *gin.Context) {
	// TODO: Implement actual electricity consumption data
	c.JSON(http.StatusOK, gin.H{
		"daily_consumption":   15.5,
		"monthly_consumption": 450.0,
		"unit":               "kWh",
		"current_load":       2.3,
		"last_updated":       "2024-01-15T10:30:00Z",
	})
}

func (g *Gateway) GetGridStatus(c *gin.Context) {
	// TODO: Implement actual grid status data
	c.JSON(http.StatusOK, gin.H{
		"status":       "stable",
		"load":         78.5,
		"voltage":      230.2,
		"frequency":    50.1,
		"outages":      0,
		"last_updated": "2024-01-15T10:30:00Z",
	})
}