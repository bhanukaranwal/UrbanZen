package handlers

import (
	"net/http"
	"strconv"

	"github.com/bhanukaranwal/UrbanZen/services/device-mgmt/internal/config"
	"github.com/bhanukaranwal/UrbanZen/services/device-mgmt/internal/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DeviceHandler struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewDeviceHandler(cfg *config.Config, logger *zap.Logger) *DeviceHandler {
	return &DeviceHandler{
		cfg:    cfg,
		logger: logger,
	}
}

// ListDevices handles GET /devices
// @Summary List all devices
// @Description Get a paginated list of IoT devices
// @Tags devices
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Param type query string false "Filter by device type"
// @Param status query string false "Filter by device status"
// @Success 200 {object} models.DeviceListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices [get]
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	deviceType := c.Query("type")
	status := c.Query("status")

	h.logger.Info("Listing devices",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("type", deviceType),
		zap.String("status", status),
	)

	// Mock data for demonstration
	devices := []models.Device{
		{
			ID:                 1,
			DeviceID:           "WM001",
			DeviceTypeID:       1,
			Name:               "Water Meter - Sector 15",
			Status:             "active",
			ConnectivityStatus: "connected",
			Location: &models.Point{
				Lat: 28.4595,
				Lng: 77.0266,
			},
			Address: stringPtr("Sector 15, Block A, Gurgaon"),
			Configuration: models.JSON{
				"measurement_interval": 60,
				"alert_threshold":      1000,
			},
			DeviceType: &models.DeviceType{
				ID:           1,
				Name:         "Smart Water Meter",
				Category:     "water",
				Manufacturer: stringPtr("AquaTech Solutions"),
			},
		},
		{
			ID:                 2,
			DeviceID:           "EM001",
			DeviceTypeID:       2,
			Name:               "Electricity Meter - Sector 16",
			Status:             "active",
			ConnectivityStatus: "connected",
			Location: &models.Point{
				Lat: 28.4605,
				Lng: 77.0276,
			},
			Address: stringPtr("Sector 16, Block B, Gurgaon"),
			Configuration: models.JSON{
				"measurement_interval": 30,
				"power_factor_alert":   0.85,
			},
			DeviceType: &models.DeviceType{
				ID:           2,
				Name:         "Smart Electricity Meter",
				Category:     "electricity",
				Manufacturer: stringPtr("PowerSense Systems"),
			},
		},
	}

	response := models.DeviceListResponse{
		Devices: devices,
		Pagination: models.Pagination{
			CurrentPage:  page,
			TotalPages:   1,
			TotalItems:   len(devices),
			ItemsPerPage: limit,
		},
	}

	c.JSON(http.StatusOK, response)
}

// CreateDevice handles POST /devices
// @Summary Create a new device
// @Description Register a new IoT device in the system
// @Tags devices
// @Accept json
// @Produce json
// @Param device body models.CreateDeviceRequest true "Device data"
// @Success 201 {object} models.Device
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices [post]
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var req models.CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	h.logger.Info("Creating device", zap.String("device_id", req.DeviceID))

	// Mock creation - in real implementation, this would save to database
	device := models.Device{
		ID:           3,
		DeviceID:     req.DeviceID,
		DeviceTypeID: req.DeviceTypeID,
		Name:         req.Name,
		Description:  req.Description,
		Location:     req.Location,
		Address:      req.Address,
		WardID:       req.WardID,
		ZoneID:       req.ZoneID,
		Status:       "inactive",
		ConnectivityStatus: "disconnected",
		Configuration: req.Configuration,
		Metadata:     req.Metadata,
	}

	c.JSON(http.StatusCreated, device)
}

// GetDevice handles GET /devices/:id
// @Summary Get device details
// @Description Get detailed information about a specific device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} models.Device
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices/{id} [get]
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	deviceID := c.Param("id")
	
	h.logger.Info("Getting device", zap.String("device_id", deviceID))

	// Mock device data
	device := models.Device{
		ID:           1,
		DeviceID:     deviceID,
		DeviceTypeID: 1,
		Name:         "Water Meter - Sector 15",
		Status:       "active",
		ConnectivityStatus: "connected",
		Location: &models.Point{
			Lat: 28.4595,
			Lng: 77.0266,
		},
		Address: stringPtr("Sector 15, Block A, Gurgaon"),
		Configuration: models.JSON{
			"measurement_interval": 60,
			"alert_threshold":      1000,
		},
	}

	c.JSON(http.StatusOK, device)
}

// UpdateDevice handles PUT /devices/:id
// @Summary Update device
// @Description Update device information and configuration
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param device body models.UpdateDeviceRequest true "Updated device data"
// @Success 200 {object} models.Device
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices/{id} [put]
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	deviceID := c.Param("id")
	
	var req models.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	h.logger.Info("Updating device", zap.String("device_id", deviceID))

	// Mock update response
	device := models.Device{
		ID:           1,
		DeviceID:     deviceID,
		DeviceTypeID: 1,
		Name:         getStringValue(req.Name, "Water Meter - Sector 15"),
		Status:       getStringValue(req.Status, "active"),
		ConnectivityStatus: getStringValue(req.ConnectivityStatus, "connected"),
	}

	c.JSON(http.StatusOK, device)
}

// DeleteDevice handles DELETE /devices/:id
// @Summary Delete device
// @Description Remove a device from the system
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices/{id} [delete]
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	deviceID := c.Param("id")
	
	h.logger.Info("Deleting device", zap.String("device_id", deviceID))

	c.Status(http.StatusNoContent)
}

// SendCommand handles POST /devices/:id/command
// @Summary Send command to device
// @Description Send a command to an IoT device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param command body models.DeviceCommandRequest true "Command data"
// @Success 202 {object} models.DeviceCommand
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices/{id}/command [post]
func (h *DeviceHandler) SendCommand(c *gin.Context) {
	deviceID := c.Param("id")
	
	var req models.DeviceCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	h.logger.Info("Sending command to device",
		zap.String("device_id", deviceID),
		zap.String("command", req.Command),
	)

	// Mock command response
	command := models.DeviceCommand{
		ID:          1,
		DeviceID:    deviceID,
		CommandID:   "cmd-" + deviceID + "-001",
		CommandType: req.Command,
		CommandData: req.Parameters,
		Status:      "pending",
	}

	c.JSON(http.StatusAccepted, command)
}

// GetDeviceStatus handles GET /devices/:id/status
// @Summary Get device status
// @Description Get current status and connectivity information for a device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices/{id}/status [get]
func (h *DeviceHandler) GetDeviceStatus(c *gin.Context) {
	deviceID := c.Param("id")
	
	h.logger.Info("Getting device status", zap.String("device_id", deviceID))

	status := gin.H{
		"device_id":           deviceID,
		"status":              "active",
		"connectivity_status": "connected",
		"last_seen":           "2024-01-01T12:00:00Z",
		"battery_level":       85,
		"signal_strength":     -65,
		"firmware_version":    "1.2.3",
		"uptime":              "72h15m30s",
	}

	c.JSON(http.StatusOK, status)
}

// GetDeviceTelemetry handles GET /devices/:id/telemetry
// @Summary Get device telemetry
// @Description Get telemetry data for a specific device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param start_time query string false "Start time (ISO 8601)"
// @Param end_time query string false "End time (ISO 8601)"
// @Param metrics query string false "Comma-separated list of metrics"
// @Success 200 {array} models.DeviceTelemetry
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /devices/{id}/telemetry [get]
func (h *DeviceHandler) GetDeviceTelemetry(c *gin.Context) {
	deviceID := c.Param("id")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	metrics := c.Query("metrics")

	h.logger.Info("Getting device telemetry",
		zap.String("device_id", deviceID),
		zap.String("start_time", startTime),
		zap.String("end_time", endTime),
		zap.String("metrics", metrics),
	)

	// Mock telemetry data
	telemetry := []models.DeviceTelemetry{
		{
			DeviceID:     deviceID,
			MetricName:   "flow_rate",
			MetricValue:  15.5,
			Unit:         stringPtr("L/min"),
			QualityScore: 0.98,
		},
		{
			DeviceID:     deviceID,
			MetricName:   "pressure",
			MetricValue:  2.1,
			Unit:         stringPtr("bar"),
			QualityScore: 0.95,
		},
	}

	c.JSON(http.StatusOK, telemetry)
}

// Placeholder handlers for device types and firmware
func (h *DeviceHandler) ListDeviceTypes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Device types endpoint"})
}

func (h *DeviceHandler) CreateDeviceType(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create device type endpoint"})
}

func (h *DeviceHandler) ListFirmware(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Firmware list endpoint"})
}

func (h *DeviceHandler) UploadFirmware(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Upload firmware endpoint"})
}

func (h *DeviceHandler) DeployFirmware(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Deploy firmware endpoint"})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func getStringValue(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}