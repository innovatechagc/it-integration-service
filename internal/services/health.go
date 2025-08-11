package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// === IMPLEMENTACIÓN ===
type HealthService interface {
	CheckHealth() map[string]interface{}
	CheckReadiness() map[string]interface{}
}

type healthService struct {
	startTime time.Time
}

func NewHealthService() HealthService {
	return &healthService{startTime: time.Now()}
}

func (s *healthService) CheckHealth() map[string]interface{} {
	return map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(s.startTime).String(),
		"service":   "it-integration-service",
		"version":   "1.0.0",
	}
}

func (s *healthService) CheckReadiness() map[string]interface{} {
	ready := true
	checks := make(map[string]bool)

	return map[string]interface{}{
		"ready":     ready,
		"timestamp": time.Now().UTC(),
		"checks":    checks,
	}
}

// === TESTS ===
func TestHealthService_CheckHealth(t *testing.T) {
	service := NewHealthService()
	result := service.CheckHealth()

	assert.Equal(t, "healthy", result["status"])
	assert.Equal(t, "it-integration-service", result["service"])
	assert.Equal(t, "1.0.0", result["version"])
	assert.NotNil(t, result["timestamp"])
	assert.NotNil(t, result["uptime"])
}

func TestHealthService_CheckReadiness(t *testing.T) {
	service := NewHealthService()
	result := service.CheckReadiness()

	assert.Equal(t, true, result["ready"])
	assert.NotNil(t, result["timestamp"])
	assert.NotNil(t, result["checks"])
}

func TestHealthService_UptimeIncreases(t *testing.T) {
	service := NewHealthService()

	result1 := service.CheckHealth()
	uptime1 := result1["uptime"].(string)

	time.Sleep(10 * time.Millisecond)

	result2 := service.CheckHealth()
	uptime2 := result2["uptime"].(string)

	assert.NotEqual(t, uptime1, uptime2, "uptime should increase over time")
}

func TestHealthService_TimestampIsRecent(t *testing.T) {
	service := NewHealthService()
	result := service.CheckHealth()
	timestamp := result["timestamp"].(time.Time)

	now := time.Now().UTC()
	diff := now.Sub(timestamp)

	assert.Less(t, diff, time.Minute, "timestamp should be recent")
	assert.Greater(t, diff, -time.Minute, "timestamp should not be in the future")
}
