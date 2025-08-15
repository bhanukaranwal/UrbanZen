package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/bhanukaranwal/urbanzen/internal/config"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

type visitor struct {
	requests int
	lastSeen time.Time
}

func RateLimiter(cfg *config.Config) gin.HandlerFunc {
	limiter := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     cfg.Security.RateLimitPerMin,
		window:   time.Minute,
	}

	// Clean up old visitors every 10 minutes
	go limiter.cleanup()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		if !limiter.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists {
		rl.visitors[ip] = &visitor{
			requests: 1,
			lastSeen: now,
		}
		return true
	}

	// Reset counter if window has passed
	if now.Sub(v.lastSeen) > rl.window {
		v.requests = 1
		v.lastSeen = now
		return true
	}

	if v.requests >= rl.rate {
		return false
	}

	v.requests++
	v.lastSeen = now
	return true
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(10 * time.Minute)
		
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}