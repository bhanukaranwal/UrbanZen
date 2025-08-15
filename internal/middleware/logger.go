package middleware

import (
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/bhanukaranwal/UrbanZen/pkg/logger"
)

func Logger(log logger.Logger) gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        log.Info(
            "method", param.Method,
            "path", param.Path,
            "status", param.StatusCode,
            "latency", param.Latency,
            "ip", param.ClientIP,
            "user_agent", param.Request.UserAgent(),
        )
        return ""
    })
}

func Security() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Next()
    }
}