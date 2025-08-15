package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Claims represents JWT claims
type Claims struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthService handles authentication operations
type AuthService struct {
	db        *sql.DB
	redisDB   *redis.Client
	jwtSecret string
	logger    *zap.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(db *sql.DB, redisDB *redis.Client, jwtSecret string, logger *zap.Logger) *AuthService {
	return &AuthService{
		db:        db,
		redisDB:   redisDB,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// JWTMiddleware validates JWT tokens
func (a *AuthService) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Check if token is blacklisted
		ctx := context.Background()
		blacklisted, err := a.redisDB.Get(ctx, "blacklist:"+tokenString).Result()
		if err == nil && blacklisted == "true" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(a.jwtSecret), nil
		})

		if err != nil {
			a.logger.Error("JWT validation error", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// RoleMiddleware checks if user has required role
func (a *AuthService) RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		role := userRole.(string)
		
		// Check if user has required role
		hasRole := false
		for _, requiredRole := range requiredRoles {
			if role == requiredRole || role == "super_admin" {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateToken generates a JWT token for a user
func (a *AuthService) GenerateToken(userID int64, email, role, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "urbanzen",
			Subject:   strconv.FormatInt(userID, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.jwtSecret))
}

// RevokeToken blacklists a token
func (a *AuthService) RevokeToken(tokenString string) error {
	ctx := context.Background()
	
	// Parse token to get expiration time
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.jwtSecret), nil
	})
	
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return fmt.Errorf("invalid token claims")
	}

	// Calculate TTL for blacklist entry
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil // Token already expired
	}

	// Add token to blacklist
	return a.redisDB.Set(ctx, "blacklist:"+tokenString, "true", ttl).Err()
}

// ValidateAPIKey validates API key for service-to-service communication
func (a *AuthService) ValidateAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		// Validate API key against database
		var serviceID int64
		var serviceName string
		err := a.db.QueryRow("SELECT id, name FROM api_keys WHERE key_hash = $1 AND active = true", apiKey).
			Scan(&serviceID, &serviceName)
		
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			} else {
				a.logger.Error("API key validation error", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
			c.Abort()
			return
		}

		// Store service information in context
		c.Set("service_id", serviceID)
		c.Set("service_name", serviceName)

		c.Next()
	}
}