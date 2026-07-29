package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// AuthMiddleware validates JWT tokens from the Authorization header.
func AuthMiddleware(adminSvc *service.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		claims, err := adminSvc.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("session_id", claims.SessionID)
		c.Next()
	}
}

// OptionalAuth extracts user info from token if present, but doesn't block.
func OptionalAuth(adminSvc *service.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Next()
			return
		}
		claims, err := adminSvc.ValidateToken(parts[1])
		if err == nil && claims != nil {
			c.Set("admin_id", claims.AdminID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
		}
		c.Next()
	}
}

// RoleMiddleware checks that the authenticated user has the required role.
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		roleStr := role.(string)
		for _, r := range roles {
			if strings.EqualFold(roleStr, r) {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

// CORSMiddleware handles cross-origin requests.
// In production, set the CORS_ORIGIN env var to your panel domain.
func CORSMiddleware() gin.HandlerFunc {
	allowedOrigin := os.Getenv("CORS_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// If a specific origin is configured, only allow that origin
		if allowedOrigin != "*" {
			if origin == allowedOrigin {
				c.Header("Access-Control-Allow-Origin", allowedOrigin)
				c.Header("Vary", "Origin")
			} else if origin == "" {
				// Same-origin request (no Origin header) — allow without CORS headers
				if c.Request.Method == "OPTIONS" {
					c.AbortWithStatus(http.StatusNoContent)
					return
				}
				c.Next()
				return
			} else {
				// Origin not allowed
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// RateLimiter provides in-memory rate limiting with named zones.
type RateLimiter struct {
	mu        sync.Mutex
	requests  map[string]map[string]int
	windows   map[string]time.Duration
	limits    map[string]int
	lastClean map[string]time.Time
}

// RateLimitConfig defines a named rate limit zone.
type RateLimitConfig struct {
	Name   string
	Limit  int
	Window time.Duration
}

// NewRateLimiter creates a new rate limiter with multiple named zones.
// If no configs provided, defaults to a single 100 req/min zone.
func NewRateLimiter(configs ...RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		requests:  make(map[string]map[string]int),
		windows:   make(map[string]time.Duration),
		limits:    make(map[string]int),
		lastClean: make(map[string]time.Time),
	}

	if len(configs) == 0 {
		configs = []RateLimitConfig{
			{Name: "default", Limit: 100, Window: time.Minute},
		}
	}

	for _, cfg := range configs {
		rl.limits[cfg.Name] = cfg.Limit
		rl.windows[cfg.Name] = cfg.Window
		rl.requests[cfg.Name] = make(map[string]int)
		rl.lastClean[cfg.Name] = time.Now()
	}

	return rl
}

// Middleware returns a Gin handler that rate limits by IP.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rl.mu.Lock()

		ip := c.ClientIP()
		zone := "default"

		// Determine zone based on request path
		path := c.Request.URL.Path
		switch {
		case strings.Contains(path, "/login") || strings.Contains(path, "/auth/refresh"):
			zone = "auth"
		case strings.Contains(path, "/sub/") || strings.Contains(path, "/sub-group/"):
			zone = "subscription"
		default:
			zone = "api"
		}

		// Fall back to default zone if specific zone not configured
		if _, ok := rl.limits[zone]; !ok {
			zone = "default"
		}

		// Per-zone cleanup: only reset this zone if its window has elapsed
		now := time.Now()
		if lastClean, ok := rl.lastClean[zone]; ok && now.Sub(lastClean) > rl.windows[zone] {
			rl.requests[zone] = make(map[string]int)
			rl.lastClean[zone] = now
		}

		// Check limit BEFORE incrementing (strict enforcement)
		count := rl.requests[zone][ip]
		limit := rl.limits[zone]
		window := rl.windows[zone]

		if count >= limit {
			rl.mu.Unlock()
			c.Header("Retry-After", fmt.Sprintf("%.0f", window.Seconds()))
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": int(window.Seconds()),
				"zone":        zone,
				"limit":       limit,
			})
			return
		}

		// Increment counter AFTER successful check
		rl.requests[zone][ip] = count + 1
		rl.mu.Unlock()

		// Set rate limit headers
		remaining := limit - (count + 1)
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Zone", zone)

		c.Next()
	}
}

// RequestIDMiddleware adds a unique ID to each request.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("vort-%d", time.Now().UnixNano())
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// FederationKeyMiddleware authenticates incoming federation requests from remote panels.
// Validates the key against a provided validator function.
func FederationKeyMiddleware(validateKey func(string) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check X-Federation-Key header first
		fedKey := c.GetHeader("X-Federation-Key")
		if fedKey != "" && validateKey(fedKey) {
			c.Set("federation_source", "remote")
			c.Next()
			return
		}
		// Fall back to Bearer token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if validateKey(token) {
				c.Set("federation_source", "token")
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid federation key"})
	}
}

// SecurityHeadersMiddleware sets security-related HTTP headers.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
