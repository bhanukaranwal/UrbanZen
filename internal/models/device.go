package models

import (
	"time"
	"github.com/google/uuid"
)

type Device struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Type        string                 `json:"type" db:"type"`
	Location    Location               `json:"location" db:"location"`
	Status      string                 `json:"status" db:"status"`
	LastSeen    time.Time              `json:"last_seen" db:"last_seen"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

type DeviceData struct {
	DeviceID    string                 `json:"device_id"`
	DeviceType  string                 `json:"device_type"`
	Timestamp   time.Time              `json:"timestamp"`
	Location    Location               `json:"location"`
	Metrics     map[string]interface{} `json:"metrics"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type User struct {
	ID                  uuid.UUID              `json:"id" db:"id"`
	Username            string                 `json:"username" db:"username"`
	Email               string                 `json:"email" db:"email"`
	PasswordHash        string                 `json:"-" db:"password_hash"`
	FirstName           string                 `json:"first_name" db:"first_name"`
	LastName            string                 `json:"last_name" db:"last_name"`
	Role                string                 `json:"role" db:"role"`
	Phone               string                 `json:"phone" db:"phone"`
	Address             string                 `json:"address" db:"address"`
	IsActive            bool                   `json:"is_active" db:"is_active"`
	EmailVerified       bool                   `json:"email_verified" db:"email_verified"`
	NotificationPrefs   map[string]interface{} `json:"notification_preferences" db:"notification_preferences"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
}

type Alert struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Type        string                 `json:"type" db:"type"`
	Severity    string                 `json:"severity" db:"severity"`
	Title       string                 `json:"title" db:"title"`
	Message     string                 `json:"message" db:"message"`
	DeviceID    string                 `json:"device_id,omitempty" db:"device_id"`
	UserID      *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	Acknowledged bool                  `json:"acknowledged" db:"acknowledged"`
	Resolved    bool                   `json:"resolved" db:"resolved"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

type Notification struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	UserID      uuid.UUID              `json:"user_id" db:"user_id"`
	Type        string                 `json:"type" db:"type"`
	Title       string                 `json:"title" db:"title"`
	Message     string                 `json:"message" db:"message"`
	Priority    string                 `json:"priority" db:"priority"`
	Channels    []string               `json:"channels" db:"channels"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty" db:"scheduled_at"`
	Status      string                 `json:"status" db:"status"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}