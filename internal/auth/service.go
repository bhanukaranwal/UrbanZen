package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/bhanukaranwal/urbanzen/internal/models"
	"github.com/bhanukaranwal/urbanzen/pkg/database"
	"github.com/bhanukaranwal/urbanzen/pkg/logger"
)

type Service struct {
	db     *database.PostgresDB
	redis  *database.RedisClient
	config *Config
	logger logger.Logger
}

type Config struct {
	JWTSecret           string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
	PasswordMinLength   int
	MaxLoginAttempts    int
	LockoutDuration     time.Duration
	RequireMFA          bool
}

type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	SessionID   string   `json:"session_id"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty"`
}

type LoginResponse struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	ExpiresIn    int64            `json:"expires_in"`
	User         *models.UserInfo `json:"user"`
}

func NewService(db *database.PostgresDB, redis *database.RedisClient, 
	config *Config, logger logger.Logger) *Service {
	return &Service{
		db:     db,
		redis:  redis,
		config: config,
		logger: logger,
	}
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Check rate limiting
	if err := s.checkRateLimit(ctx, req.Username); err != nil {
		return nil, err
	}
	
	// Get user from database
	user, err := s.getUserByUsername(ctx, req.Username)
	if err != nil {
		s.incrementFailedAttempts(ctx, req.Username)
		return nil, fmt.Errorf("invalid credentials")
	}
	
	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, fmt.Errorf("account locked until %v", user.LockedUntil)
	}
	
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.incrementFailedAttempts(ctx, req.Username)
		return nil, fmt.Errorf("invalid credentials")
	}
	
	// Check MFA if required
	if s.config.RequireMFA && user.MFAEnabled {
		if req.MFACode == "" {
			return nil, fmt.Errorf("MFA code required")
		}
		
		if !s.verifyMFACode(ctx, user.ID, req.MFACode) {
			s.incrementFailedAttempts(ctx, req.Username)
			return nil, fmt.Errorf("invalid MFA code")
		}
	}
	
	// Reset failed attempts
	s.resetFailedAttempts(ctx, req.Username)
	
	// Generate tokens
	sessionID := uuid.New().String()
	accessToken, err := s.generateAccessToken(user, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	
	refreshToken, err := s.generateRefreshToken(user.ID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	
	// Store session
	if err := s.storeSession(ctx, sessionID, user.ID); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}
	
	// Update last login
	s.updateLastLogin(ctx, user.ID)
	
	// Log successful login
	s.logger.Info("User logged in successfully", 
		"user_id", user.ID, 
		"username", user.Username,
		"session_id", sessionID,
	)
	
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.AccessTokenExpiry.Seconds()),
		User: &models.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
		},
	}, nil
}

func (s *Service) generateAccessToken(user *models.User, sessionID string) (string, error) {
	permissions, err := s.getUserPermissions(context.Background(), user.ID)
	if err != nil {
		return "", err
	}
	
	claims := &Claims{
		UserID:      user.ID,
		Username:    user.Username,
		Role:        user.Role,
		Permissions: permissions,
		SessionID:   sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "urbanzen-auth",
			Subject:   user.ID,
			ID:        uuid.New().String(),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *Service) generateRefreshToken(userID, sessionID string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	
	refreshToken := base64.URLEncoding.EncodeToString(tokenBytes)
	
	// Store refresh token with expiry
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	value := fmt.Sprintf("%s:%s", userID, sessionID)
	
	err := s.redis.Set(context.Background(), key, value, s.config.RefreshTokenExpiry)
	return refreshToken, err
}

func (s *Service) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	
	// Check if session is still valid
	if !s.isSessionValid(ctx, claims.SessionID, claims.UserID) {
		return nil, fmt.Errorf("session expired")
	}
	
	return claims, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// Get user and session from refresh token
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	value, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}
	
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid refresh token format")
	}
	
	userID, sessionID := parts[0], parts[1]
	
	// Get user
	user, err := s.getUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Generate new tokens
	newAccessToken, err := s.generateAccessToken(user, sessionID)
	if err != nil {
		return nil, err
	}
	
	newRefreshToken, err := s.generateRefreshToken(userID, sessionID)
	if err != nil {
		return nil, err
	}
	
	// Invalidate old refresh token
	s.redis.Del(ctx, key)
	
	return &LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.config.AccessTokenExpiry.Seconds()),
		User: &models.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
		},
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	// Invalidate session
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	
	// Get refresh token to invalidate it too
	if sessionData, err := s.redis.Get(ctx, sessionKey); err == nil {
		var session models.Session
		if err := json.Unmarshal([]byte(sessionData), &session); err == nil {
			refreshKey := fmt.Sprintf("refresh_token:%s", session.RefreshToken)
			s.redis.Del(ctx, refreshKey)
		}
	}
	
	return s.redis.Del(ctx, sessionKey)
}

func (s *Service) checkRateLimit(ctx context.Context, username string) error {
	key := fmt.Sprintf("login_attempts:%s", username)
	attempts, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil // No previous attempts
	}
	
	if attempts >= fmt.Sprintf("%d", s.config.MaxLoginAttempts) {
		return fmt.Errorf("too many login attempts, try again later")
	}
	
	return nil
}

func (s *Service) incrementFailedAttempts(ctx context.Context, username string) {
	key := fmt.Sprintf("login_attempts:%s", username)
	s.redis.Incr(ctx, key)
	s.redis.Expire(ctx, key, s.config.LockoutDuration)
}

func (s *Service) resetFailedAttempts(ctx context.Context, username string) {
	key := fmt.Sprintf("login_attempts:%s", username)
	s.redis.Del(ctx, key)
}

// Role-Based Access Control (RBAC) Implementation
func (s *Service) HasPermission(ctx context.Context, userID, permission string) bool {
	permissions, err := s.getUserPermissions(ctx, userID)
	if err != nil {
		return false
	}
	
	for _, p := range permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	
	return false
}

func (s *Service) getUserPermissions(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT DISTINCT p.name
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE u.id = $1 AND u.is_active = true
	`
	
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var permissions []string
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			continue
		}
		permissions = append(permissions, permission)
	}
	
	return permissions, nil
}