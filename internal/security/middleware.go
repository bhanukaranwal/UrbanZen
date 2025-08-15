package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/bhanukaranwal/urbanzen/internal/auth"
	"github.com/bhanukaranwal/urbanzen/pkg/logger"
)

type SecurityConfig struct {
	EnableCSRF          bool
	EnableRateLimit     bool
	RateLimitPerMinute  int
	EnableCORS          bool
	AllowedOrigins      []string
	RequireHTTPS        bool
	EnableHSTS          bool
	EnableContentTypes  bool
	EnableXSSProtection bool
}

type Middleware struct {
	config     *
