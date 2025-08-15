package device

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/bhanukaranwal/urbanzen/pkg/logger"
	"github.com/bhanukaranwal/urbanzen/pkg/database"
	"github.com/bhanukaranwal/urbanzen/pkg/kafka"
	"github.com/bhanukaranwal/urbanzen/internal/models"
)

type Service struct {
	db       *database.PostgresDB
	tsdb     *database.TimescaleDB
	producer *kafka.Producer
	consumer *kafka.Consumer
	logger   logger.Logger
}

func NewService(db *database.PostgresDB, tsdb *database.TimescaleDB, 
	producer *kafka.Producer, consumer *kafka.Consumer, log logger.Logger) *Service {
	return &Service{
		db:       db,
		tsdb:     tsdb,
		producer: producer,
		consumer: consumer,
		logger:   log,
	}
}

func (s *Service) Start(ctx context.Context) error {
	// Start consuming device data
	go s.consumeDeviceData(ctx)
	
	// Start device health monitoring
	go s.monitorDeviceHealth(ctx)
	
	// Start command processing
	go s.processCommands(ctx)
	
	s.logger.Info("Device service started")
	
	<-ctx.Done()
	return nil
}

func (s *Service) consumeDeviceData(ctx context.Context) {
	topics := []string{"device-data", "device-telemetry"}
	
	for {
		select {
		case <-ctx.Done():
			return
		default:
			messages, err := s.consumer.ConsumeMessages(topics, time.Second*5)
			if err != nil {
				s.logger.Error("Failed to consume messages", "error", err)
				continue
			}
			
			for _, msg := range messages {
				s.processDeviceMessage(msg)
			}
		}
	}
}

func (s *Service) processDeviceMessage(msg *kafka.Message) {
	var deviceData models.DeviceData
	if err := json.Unmarshal(msg.Value, &deviceData); err != nil {
		s.logger.Error("Failed to unmarshal device data", "error", err)
		return
	}
	
	// Validate device data
	if err := s.validateDeviceData(&deviceData); err != nil {
		s.logger.Error("Invalid device data", "error", err, "device_id", deviceData.DeviceID)
		return
	}
	
	// Store in TimescaleDB
	if err := s.storeDeviceData(&deviceData); err != nil {
		s.logger.Error("Failed to store device data", "error", err)
		return
	}
	
	// Process analytics
	s.processAnalytics(&deviceData)
	
	// Check for anomalies
	if anomaly := s.detectAnomaly(&deviceData); anomaly != nil {
		s.handleAnomaly(anomaly)
	}
	
	s.logger.Debug("Processed device data", "device_id", deviceData.DeviceID)
}

func (s *Service) validateDeviceData(data *models.DeviceData) error {
	if data.DeviceID == "" {
		return fmt.Errorf("device ID is required")
	}
	
	if data.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	
	if len(data.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	
	return nil
}

func (s *Service) storeDeviceData(data *models.DeviceData) error {
	query := `
		INSERT INTO device_telemetry (device_id, timestamp, device_type, location, metrics, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	metricsJSON, _ := json.Marshal(data.Metrics)
	metadataJSON, _ := json.Marshal(data.Metadata)
	
	_, err := s.tsdb.Exec(query, 
		data.DeviceID, 
		data.Timestamp, 
		data.DeviceType, 
		fmt.Sprintf("POINT(%f %f)", data.Location.Longitude, data.Location.Latitude),
		metricsJSON,
		metadataJSON,
	)
	
	return err
}

func (s *Service) processAnalytics(data *models.DeviceData) {
	// Send to analytics service for processing
	analyticsData := map[string]interface{}{
		"device_id":   data.DeviceID,
		"device_type": data.DeviceType,
		"timestamp":   data.Timestamp,
		"metrics":     data.Metrics,
		"location":    data.Location,
	}
	
	message, _ := json.Marshal(analyticsData)
	s.producer.ProduceMessage("analytics-data", data.DeviceID, message)
}

func (s *Service) detectAnomaly(data *models.DeviceData) *models.Anomaly {
	// Simple anomaly detection based on thresholds
	for metric, value := range data.Metrics {
		switch data.DeviceType {
		case "water_sensor":
			if metric == "flow_rate" && value.(float64) > 1000 {
				return &models.Anomaly{
					DeviceID:    data.DeviceID,
					Type:        "high_flow_rate",
					Severity:    "critical",
					Description: "Extremely high water flow rate detected",
					Timestamp:   time.Now(),
					Value:       value,
				}
			}
		case "electricity_meter":
			if metric == "current" && value.(float64) > 100 {
				return &models.Anomaly{
					DeviceID:    data.DeviceID,
					Type:        "high_current",
					Severity:    "warning",
					Description: "High electrical current detected",
					Timestamp:   time.Now(),
					Value:       value,
				}
			}
		}
	}
	
	return nil
}

func (s *Service) handleAnomaly(anomaly *models.Anomaly) {
	// Store anomaly
	s.storeAnomaly(anomaly)
	
	// Send alert
	alert := map[string]interface{}{
		"type":        "anomaly_detected",
		"device_id":   anomaly.DeviceID,
		"severity":    anomaly.Severity,
		"description": anomaly.Description,
		"timestamp":   anomaly.Timestamp,
	}
	
	message, _ := json.Marshal(alert)
	s.producer.ProduceMessage("alerts", anomaly.DeviceID, message)
	
	s.logger.Warn("Anomaly detected", 
		"device_id", anomaly.DeviceID,
		"type", anomaly.Type,
		"severity", anomaly.Severity,
	)
}

func (s *Service) storeAnomaly(anomaly *models.Anomaly) error {
	query := `
		INSERT INTO anomalies (device_id, type, severity, description, timestamp, value, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := s.db.Exec(query,
		anomaly.DeviceID,
		anomaly.Type,
		anomaly.Severity,
		anomaly.Description,
		anomaly.Timestamp,
		anomaly.Value,
		"{}",
	)
	
	return err
}

func (s *Service) monitorDeviceHealth(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkDeviceHealth()
		}
	}
}

func (s *Service) checkDeviceHealth() {
	// Check for devices that haven't sent data recently
	query := `
		SELECT device_id, MAX(timestamp) as last_seen
		FROM device_telemetry
		GROUP BY device_id
		HAVING MAX(timestamp) < NOW() - INTERVAL '10 minutes'
	`
	
	rows, err := s.tsdb.Query(query)
	if err != nil {
		s.logger.Error("Failed to check device health", "error", err)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var deviceID string
		var lastSeen time.Time
		
		if err := rows.Scan(&deviceID, &lastSeen); err != nil {
			continue
		}
		
		// Send offline alert
		alert := map[string]interface{}{
			"type":      "device_offline",
			"device_id": deviceID,
			"last_seen": lastSeen,
			"severity":  "warning",
		}
		
		message, _ := json.Marshal(alert)
		s.producer.ProduceMessage("alerts", deviceID, message)
	}
}

func (s *Service) processCommands(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			messages, err := s.consumer.ConsumeMessages([]string{"device-commands"}, time.Second*5)
			if err != nil {
				continue
			}
			
			for _, msg := range messages {
				s.processDeviceCommand(msg)
			}
		}
	}
}

func (s *Service) processDeviceCommand(msg *kafka.Message) {
	var command models.DeviceCommand
	if err := json.Unmarshal(msg.Value, &command); err != nil {
		s.logger.Error("Failed to unmarshal device command", "error", err)
		return
	}
	
	// Validate and execute command
	if err := s.executeCommand(&command); err != nil {
		s.logger.Error("Failed to execute command", "error", err, "device_id", command.DeviceID)
		return
	}
	
	s.logger.Info("Command executed", "device_id", command.DeviceID, "command", command.Command)
}

func (s *Service) executeCommand(command *models.DeviceCommand) error {
	// In a real implementation, this would send the command to the actual device
	// For now, we'll just log it and store the command history
	
	query := `
		INSERT INTO device_commands (device_id, command, parameters, timestamp, status)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	parametersJSON, _ := json.Marshal(command.Parameters)
	
	_, err := s.db.Exec(query,
		command.DeviceID,
		command.Command,
		parametersJSON,
		time.Now(),
		"executed",
	)
	
	return err
}