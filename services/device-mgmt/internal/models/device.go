package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Device represents an IoT device
type Device struct {
	ID               int64           `json:"id" db:"id"`
	DeviceID         string          `json:"device_id" db:"device_id"`
	DeviceTypeID     int64           `json:"device_type_id" db:"device_type_id"`
	Name             string          `json:"name" db:"name"`
	Description      *string         `json:"description" db:"description"`
	Location         *Point          `json:"location" db:"location"`
	Address          *string         `json:"address" db:"address"`
	WardID           *int            `json:"ward_id" db:"ward_id"`
	ZoneID           *int            `json:"zone_id" db:"zone_id"`
	Status           string          `json:"status" db:"status"`
	ConnectivityStatus string        `json:"connectivity_status" db:"connectivity_status"`
	Configuration    JSON            `json:"configuration" db:"configuration"`
	Metadata         JSON            `json:"metadata" db:"metadata"`
	InstalledAt      *time.Time      `json:"installed_at" db:"installed_at"`
	LastSeen         *time.Time      `json:"last_seen" db:"last_seen"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
	
	// Joined fields
	DeviceType       *DeviceType     `json:"device_type,omitempty"`
}

// DeviceType represents a type of IoT device
type DeviceType struct {
	ID                  int64     `json:"id" db:"id"`
	Name                string    `json:"name" db:"name"`
	Category            string    `json:"category" db:"category"`
	Description         *string   `json:"description" db:"description"`
	Manufacturer        *string   `json:"manufacturer" db:"manufacturer"`
	Model               *string   `json:"model" db:"model"`
	Version             *string   `json:"version" db:"version"`
	Capabilities        JSON      `json:"capabilities" db:"capabilities"`
	ConfigurationSchema JSON      `json:"configuration_schema" db:"configuration_schema"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// DeviceCommand represents a command sent to a device
type DeviceCommand struct {
	ID           int64     `json:"id"`
	DeviceID     string    `json:"device_id"`
	CommandID    string    `json:"command_id"`
	CommandType  string    `json:"command_type"`
	CommandData  JSON      `json:"command_data"`
	Status       string    `json:"status"`
	ResponseData JSON      `json:"response_data,omitempty"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	ExecutedAt   *time.Time `json:"executed_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// DeviceTelemetry represents telemetry data from a device
type DeviceTelemetry struct {
	Time         time.Time `json:"time"`
	DeviceID     string    `json:"device_id"`
	MetricName   string    `json:"metric_name"`
	MetricValue  float64   `json:"metric_value"`
	Unit         *string   `json:"unit"`
	QualityScore float64   `json:"quality_score"`
	Metadata     JSON      `json:"metadata"`
}

// Point represents a geographic point
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// JSON represents a JSON field that can be stored in the database
type JSON map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, j)
}

// CreateDeviceRequest represents the request to create a new device
type CreateDeviceRequest struct {
	DeviceID      string  `json:"device_id" validate:"required"`
	DeviceTypeID  int64   `json:"device_type_id" validate:"required"`
	Name          string  `json:"name" validate:"required"`
	Description   *string `json:"description"`
	Location      *Point  `json:"location"`
	Address       *string `json:"address"`
	WardID        *int    `json:"ward_id"`
	ZoneID        *int    `json:"zone_id"`
	Configuration JSON    `json:"configuration"`
	Metadata      JSON    `json:"metadata"`
}

// UpdateDeviceRequest represents the request to update a device
type UpdateDeviceRequest struct {
	Name               *string `json:"name"`
	Description        *string `json:"description"`
	Location           *Point  `json:"location"`
	Address            *string `json:"address"`
	WardID             *int    `json:"ward_id"`
	ZoneID             *int    `json:"zone_id"`
	Status             *string `json:"status"`
	ConnectivityStatus *string `json:"connectivity_status"`
	Configuration      JSON    `json:"configuration"`
	Metadata           JSON    `json:"metadata"`
}

// DeviceCommandRequest represents a command to be sent to a device
type DeviceCommandRequest struct {
	Command    string `json:"command" validate:"required"`
	Parameters JSON   `json:"parameters"`
}

// DeviceListResponse represents the response for listing devices
type DeviceListResponse struct {
	Devices    []Device   `json:"devices"`
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination information
type Pagination struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	TotalItems  int `json:"total_items"`
	ItemsPerPage int `json:"items_per_page"`
}