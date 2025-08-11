package middleware

import (
	"net/http"
	"sync"
	"time"

	"it-integration-service/internal/domain"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter maneja el rate limiting por IP
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      int
	burst    int
}

// NewRateLimiter crea un nuevo rate limiter
func NewRateLimiter(rps, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}
}

// getLimiter obtiene o crea un limiter para una IP específica
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.rps), rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// cleanupLimiters limpia limiters antiguos para evitar memory leaks
func (rl *RateLimiter) cleanupLimiters() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// En una implementación real, aquí limpiarías limiters que no se han usado
	// en un período de tiempo específico
	// Por ahora, mantenemos todos los limiters
}

// RateLimit middleware para limitar requests por IP
func RateLimit(rps, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, burst)

	// Iniciar cleanup periódico
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanupLimiters()
		}
	}()

	return func(c *gin.Context) {
		ip := getClientIP(c)
		limiter := limiter.getLimiter(ip)

		if !limiter.Allow() {
			// Registrar métrica de rate limit
			UpdateRateLimitMetrics(c.FullPath(), ip)

			c.JSON(http.StatusTooManyRequests, domain.APIResponse{
				Code:    "RATE_LIMIT_EXCEEDED",
				Message: "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// WebhookRateLimit middleware específico para webhooks
func WebhookRateLimit(rps, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, burst)

	return func(c *gin.Context) {
		ip := getClientIP(c)
		limiter := limiter.getLimiter(ip)

		if !limiter.Allow() {
			// Registrar métrica de rate limit para webhooks
			UpdateRateLimitMetrics("webhook", ip)

			c.JSON(http.StatusTooManyRequests, domain.APIResponse{
				Code:    "WEBHOOK_RATE_LIMIT_EXCEEDED",
				Message: "Webhook rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantRateLimit middleware para rate limiting por tenant
func TenantRateLimit(rps, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, burst)

	return func(c *gin.Context) {
		tenantID := getTenantID(c)
		if tenantID == "unknown" {
			c.Next()
			return
		}

		limiter := limiter.getLimiter(tenantID)

		if !limiter.Allow() {
			// Registrar métrica de rate limit por tenant
			UpdateRateLimitMetrics("tenant", tenantID)

			c.JSON(http.StatusTooManyRequests, domain.APIResponse{
				Code:    "TENANT_RATE_LIMIT_EXCEEDED",
				Message: "Tenant rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientIP obtiene la IP real del cliente
func getClientIP(c *gin.Context) string {
	// Verificar headers de proxy
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := c.GetHeader("X-Client-IP"); ip != "" {
		return ip
	}

	// Usar IP remota como fallback
	return c.ClientIP()
}
