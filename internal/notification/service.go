package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/bhanukaranwal/urbanzen/internal/config"
	"github.com/bhanukaranwal/urbanzen/internal/models"
	"github.com/bhanukaranwal/urbanzen/pkg/database"
	"github.com/bhanukaranwal/urbanzen/pkg/kafka"
	"github.com/bhanukaranwal/urbanzen/pkg/logger"
	"github.com/bhanukaranwal/urbanzen/pkg/notification/email"
	"github.com/bhanukaranwal/urbanzen/pkg/notification/sms"
	"github.com/bhanukaranwal/urbanzen/pkg/notification/push"
)

type Service struct {
	db          *database.PostgresDB
	redis       *database.RedisDB
	consumer    *kafka.Consumer
	config      *config.Config
	logger      logger.Logger
	emailSvc    *email.Service
	smsSvc      *sms.Service
	pushSvc     *push.Service
	channels    map[string]NotificationChannel
}

type NotificationChannel interface {
	Send(ctx context.Context, notification *models.Notification) error
	IsAvailable() bool
}

func NewService(db *database.PostgresDB, redis *database.RedisDB, 
	consumer *kafka.Consumer, cfg *config.Config, log logger.Logger) *Service {
	
	emailSvc := email.NewService(cfg.ExternalAPIs.EmailService, log)
	smsSvc := sms.NewService(cfg.ExternalAPIs.SMSGateway, log)
	pushSvc := push.NewService(cfg.Notifications.PushNotifications, log)
	
	channels := map[string]NotificationChannel{
		"email": emailSvc,
		"sms":   smsSvc,
		"push":  pushSvc,
	}
	
	return &Service{
		db:       db,
		redis:    redis,
		consumer: consumer,
		config:   cfg,
		logger:   log,
		emailSvc: emailSvc,
		smsSvc:   smsSvc,
		pushSvc:  pushSvc,
		channels: channels,
	}
}

func (s *Service) Start(ctx context.Context) error {
	// Start consuming notification requests
	go s.consumeNotifications(ctx)
	
	// Start notification scheduler
	go s.startScheduler(ctx)
	
	// Start delivery status processor
	go s.processDeliveryStatus(ctx)
	
	s.logger.Info("Notification service started")
	
	<-ctx.Done()
	return nil
}

func (s *Service) consumeNotifications(ctx context.Context) {
	topics := []string{"user-notifications", "system-alerts", "emergency-alerts"}
	
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
				s.processNotificationMessage(ctx, msg)
			}
		}
	}
}

func (s *Service) processNotificationMessage(ctx context.Context, msg *kafka.Message) {
	var notification models.Notification
	if err := json.Unmarshal(msg.Value, &notification); err != nil {
		s.logger.Error("Failed to unmarshal notification", "error", err)
		return
	}
	
	// Validate notification
	if err := s.validateNotification(&notification); err != nil {
		s.logger.Error("Invalid notification", "error", err)
		return
	}
	
	// Store notification
	if err := s.storeNotification(&notification); err != nil {
		s.logger.Error("Failed to store notification", "error", err)
		return
	}
	
	// Process notification based on priority and type
	switch notification.Priority {
	case "emergency":
		s.processEmergencyNotification(ctx, &notification)
	case "high":
		s.processHighPriorityNotification(ctx, &notification)
	default:
		s.processRegularNotification(ctx, &notification)
	}
}

func (s *Service) processEmergencyNotification(ctx context.Context, notification *models.Notification) {
	// Emergency notifications are sent immediately via all available channels
	channels := []string{"push", "sms", "email"}
	
	for _, channel := range channels {
		if svc, exists := s.channels[channel]; exists && svc.IsAvailable() {
			go func(ch string, svc NotificationChannel) {
				if err := svc.Send(ctx, notification); err != nil {
					s.logger.Error("Failed to send emergency notification", 
						"channel", ch, "error", err, "notification_id", notification.ID)
				} else {
					s.updateDeliveryStatus(notification.ID, ch, "delivered")
				}
			}(channel, svc)
		}
	}
}

func (s *Service) processHighPriorityNotification(ctx context.Context, notification *models.Notification) {
	// High priority notifications are sent via push and SMS first
	preferredChannels := []string{"push", "sms"}
	
	for _, channel := range preferredChannels {
		if svc, exists := s.channels[channel]; exists && svc.IsAvailable() {
			if err := svc.Send(ctx, notification); err != nil {
				s.logger.Error("Failed to send high priority notification", 
					"channel", channel, "error", err)
				continue
			}
			s.updateDeliveryStatus(notification.ID, channel, "delivered")
			return // Send via one channel successfully
		}
	}
	
	// Fallback to email if other channels fail
	if emailSvc, exists := s.channels["email"]; exists && emailSvc.IsAvailable() {
		if err := emailSvc.Send(ctx, notification); err != nil {
			s.logger.Error("Failed to send notification via email fallback", "error", err)
		} else {
			s.updateDeliveryStatus(notification.ID, "email", "delivered")
		}
	}
}

func (s *Service) processRegularNotification(ctx context.Context, notification *models.Notification) {
	// Regular notifications follow user preferences
	userPrefs, err := s.getUserNotificationPreferences(notification.UserID)
	if err != nil {
		s.logger.Error("Failed to get user preferences", "error", err, "user_id", notification.UserID)
		// Default to email
		userPrefs = map[string]bool{"email": true}
	}
	
	for channel, enabled := range userPrefs {
		if !enabled {
			continue
		}
		
		if svc, exists := s.channels[channel]; exists && svc.IsAvailable() {
			if err := svc.Send(ctx, notification); err != nil {
				s.logger.Error("Failed to send notification", 
					"channel", channel, "error", err)
				s.updateDeliveryStatus(notification.ID, channel, "failed")
			} else {
				s.updateDeliveryStatus(notification.ID, channel, "delivered")
			}
		}
	}
}

func (s *Service) storeNotification(notification *models.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, type, title, message, priority, channels, 
			metadata, scheduled_at, created_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	
	channelsJSON, _ := json.Marshal(notification.Channels)
	metadataJSON, _ := json.Marshal(notification.Metadata)
	
	_, err := s.db.Exec(query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.Priority,
		channelsJSON,
		metadataJSON,
		notification.ScheduledAt,
		time.Now(),
		"pending",
	)
	
	return err
}

func (s *Service) getUserNotificationPreferences(userID string) (map[string]bool, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("user_prefs:%s", userID)
	if cached, err := s.redis.Get(cacheKey); err == nil {
		var prefs map[string]bool
		if json.Unmarshal([]byte(cached), &prefs) == nil {
			return prefs, nil
		}
	}
	
	// Get from database
	query := `
		SELECT notification_preferences 
		FROM users 
		WHERE id = $1
	`
	
	var prefsJSON string
	err := s.db.QueryRow(query, userID).Scan(&prefsJSON)
	if err != nil {
		return nil, err
	}
	
	var prefs map[string]bool
	if err := json.Unmarshal([]byte(prefsJSON), &prefs); err != nil {
		return nil, err
	}
	
	// Cache for 1 hour
	prefsBytes, _ := json.Marshal(prefs)
	s.redis.SetEX(cacheKey, string(prefsBytes), time.Hour)
	
	return prefs, nil
}

func (s *Service) updateDeliveryStatus(notificationID, channel, status string) {
	query := `
		INSERT INTO notification_delivery_status (notification_id, channel, status, attempted_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (notification_id, channel) 
		DO UPDATE SET status = $2, attempted_at = $4
	`
	
	_, err := s.db.Exec(query, notificationID, channel, status, time.Now())
	if err != nil {
		s.logger.Error("Failed to update delivery status", "error", err)
	}
}

func (s *Service) startScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processScheduledNotifications(ctx)
		}
	}
}

func (s *Service) processScheduledNotifications(ctx context.Context) {
	query := `
		SELECT id, user_id, type, title, message, priority, channels, metadata
		FROM notifications
		WHERE scheduled_at <= NOW() AND status = 'pending'
		ORDER BY priority DESC, scheduled_at ASC
		LIMIT 100
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		s.logger.Error("Failed to query scheduled notifications", "error", err)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var notification models.Notification
		var channelsJSON, metadataJSON string
		
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&notification.Priority,
			&channelsJSON,
			&metadataJSON,
		)
		
		if err != nil {
			s.logger.Error("Failed to scan notification", "error", err)
			continue
		}
		
		json.Unmarshal([]byte(channelsJSON), &notification.Channels)
		json.Unmarshal([]byte(metadataJSON), &notification.Metadata)
		
		// Process the notification
		switch notification.Priority {
		case "emergency":
			s.processEmergencyNotification(ctx, &notification)
		case "high":
			s.processHighPriorityNotification(ctx, &notification)
		default:
			s.processRegularNotification(ctx, &notification)
		}
		
		// Update status to processing
		s.updateNotificationStatus(notification.ID, "processing")
	}
}

func (s *Service) updateNotificationStatus(notificationID, status string) {
	query := `UPDATE notifications SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := s.db.Exec(query, status, time.Now(), notificationID)
	if err != nil {
		s.logger.Error("Failed to update notification status", "error", err)
	}
}

func (s *Service) validateNotification(notification *models.Notification) error {
	if notification.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	
	if notification.Title == "" {
		return fmt.Errorf("title is required")
	}
	
	if notification.Message == "" {
		return fmt.Errorf("message is required")
	}
	
	if notification.Type == "" {
		return fmt.Errorf("type is required")
	}
	
	return nil
}

func (s *Service) processDeliveryStatus(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.retryFailedNotifications(ctx)
		}
	}
}

func (s *Service) retryFailedNotifications(ctx context.Context) {
	query := `
		SELECT n.id, n.user_id, n.type, n.title, n.message, n.priority, 
			   n.channels, n.metadata, nds.channel
		FROM notifications n
		JOIN notification_delivery_status nds ON n.id = nds.notification_id
		WHERE nds.status = 'failed' 
		AND nds.attempted_at < NOW() - INTERVAL '5 minutes'
		AND n.created_at > NOW() - INTERVAL '24 hours'
		ORDER BY n.priority DESC, n.created_at ASC
		LIMIT 50
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		s.logger.Error("Failed to query failed notifications", "error", err)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var notification models.Notification
		var channelsJSON, metadataJSON, failedChannel string
		
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&notification.Priority,
			&channelsJSON,
			&metadataJSON,
			&failedChannel,
		)
		
		if err != nil {
			continue
		}
		
		json.Unmarshal([]byte(channelsJSON), &notification.Channels)
		json.Unmarshal([]byte(metadataJSON), &notification.Metadata)
		
		// Retry with the failed channel
		if svc, exists := s.channels[failedChannel]; exists && svc.IsAvailable() {
			if err := svc.Send(ctx, &notification); err != nil {
				s.logger.Error("Retry failed", "channel", failedChannel, "error", err)
			} else {
				s.updateDeliveryStatus(notification.ID, failedChannel, "delivered")
			}
		}
	}
}