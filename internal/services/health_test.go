package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthService_CheckHealth(t *testing.T) {
	service := NewHealthService()
	
	result := service.CheckHealth()
	
	assert.Equal(t, "healthy", result["status"])
	assert.Equal(t, "microservice-template", result["service"])
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