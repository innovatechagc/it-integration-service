package services

import (
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"it-integration-service/pkg/logger"
)

// HealthService maneja los health checks del servicio
type HealthService struct {
	db     *sql.DB
	logger logger.Logger
}

// NewHealthService crea una nueva instancia del servicio de health
func NewHealthService(db *sql.DB, logger logger.Logger) HealthService {
	return HealthService{
		db:     db,
		logger: logger,
	}
}

// HealthStatus representa el estado de salud del servicio
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Checks    map[string]interface{} `json:"checks,omitempty"`
}

// SystemInfo representa información del sistema
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	Memory       struct {
		Alloc      uint64 `json:"alloc"`
		TotalAlloc uint64 `json:"total_alloc"`
		Sys        uint64 `json:"sys"`
		NumGC      uint32 `json:"num_gc"`
	} `json:"memory"`
}

// DatabaseHealth representa el estado de salud de la base de datos
type DatabaseHealth struct {
	Status    string        `json:"status"`
	Latency   time.Duration `json:"latency"`
	Connections struct {
		Open  int `json:"open"`
		InUse int `json:"in_use"`
		Idle  int `json:"idle"`
	} `json:"connections"`
	Error string `json:"error,omitempty"`
}

// ExternalServiceHealth representa el estado de salud de servicios externos
type ExternalServiceHealth struct {
	MessagingService struct {
		Status    string        `json:"status"`
		Latency   time.Duration `json:"latency"`
		Error     string        `json:"error,omitempty"`
	} `json:"messaging_service"`
	Vault struct {
		Status    string        `json:"status"`
		Latency   time.Duration `json:"latency"`
		Error     string        `json:"error,omitempty"`
	} `json:"vault"`
}

var startTime = time.Now()

// CheckHealth verifica el estado general del servicio
func (s *HealthService) CheckHealth() *HealthStatus {
	status := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).String(),
		Service:   "it-integration-service",
		Version:   "1.0.0",
		Checks:    make(map[string]interface{}),
	}

	// Verificar base de datos
	dbHealth := s.checkDatabaseHealth()
	status.Checks["database"] = dbHealth
	if dbHealth.Status != "healthy" {
		status.Status = "degraded"
	}

	// Verificar servicios externos
	externalHealth := s.checkExternalServicesHealth()
	status.Checks["external_services"] = externalHealth
	if externalHealth.MessagingService.Status != "healthy" || externalHealth.Vault.Status != "healthy" {
		status.Status = "degraded"
	}

	// Verificar sistema
	systemInfo := s.getSystemInfo()
	status.Checks["system"] = systemInfo

	// Verificar integraciones activas
	integrationsHealth := s.checkIntegrationsHealth()
	status.Checks["integrations"] = integrationsHealth

	// Si hay errores críticos, marcar como unhealthy
	if dbHealth.Status == "unhealthy" {
		status.Status = "unhealthy"
	}

	return status
}

// CheckReadiness verifica si el servicio está listo para recibir tráfico
func (s *HealthService) CheckReadiness() *HealthStatus {
	status := &HealthStatus{
		Status:    "ready",
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).String(),
		Service:   "it-integration-service",
		Version:   "1.0.0",
		Checks:    make(map[string]interface{}),
	}

	// Verificar que la base de datos esté disponible
	dbHealth := s.checkDatabaseHealth()
	status.Checks["database"] = dbHealth
	if dbHealth.Status != "healthy" {
		status.Status = "not_ready"
	}

	// Verificar que los servicios críticos estén disponibles
	externalHealth := s.checkExternalServicesHealth()
	status.Checks["external_services"] = externalHealth
	if externalHealth.MessagingService.Status != "healthy" {
		status.Status = "not_ready"
	}

	return status
}

// checkDatabaseHealth verifica el estado de la base de datos
func (s *HealthService) checkDatabaseHealth() *DatabaseHealth {
	health := &DatabaseHealth{
		Status: "healthy",
	}

	start := time.Now()
	
	// Verificar conexión
	if err := s.db.Ping(); err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		health.Latency = time.Since(start)
		return health
	}

	health.Latency = time.Since(start)

	// Obtener estadísticas de conexiones
	stats := s.db.Stats()
	health.Connections.Open = stats.OpenConnections
	health.Connections.InUse = stats.InUse
	health.Connections.Idle = stats.Idle

	// Verificar que no haya demasiadas conexiones abiertas
	if stats.OpenConnections > 100 {
		health.Status = "degraded"
		health.Error = "too many open connections"
	}

	return health
}

// checkExternalServicesHealth verifica el estado de servicios externos
func (s *HealthService) checkExternalServicesHealth() *ExternalServiceHealth {
	health := &ExternalServiceHealth{}

	// Verificar servicio de mensajería
	health.MessagingService = s.checkMessagingService()

	// Verificar Vault
	health.Vault = s.checkVaultService()

	return health
}

// checkMessagingService verifica el estado del servicio de mensajería
func (s *HealthService) checkMessagingService() struct {
	Status  string        `json:"status"`
	Latency time.Duration `json:"latency"`
	Error   string        `json:"error,omitempty"`
} {
	result := struct {
		Status  string        `json:"status"`
		Latency time.Duration `json:"latency"`
		Error   string        `json:"error,omitempty"`
	}{
		Status: "healthy",
	}

	// En una implementación real, esto haría una llamada HTTP al servicio de mensajería
	// Por ahora, simulamos la verificación
	start := time.Now()
	
	// Simular llamada HTTP
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:8081/api/v1/health")
	
	result.Latency = time.Since(start)
	
	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			result.Status = "degraded"
			result.Error = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		}
	}

	return result
}

// checkVaultService verifica el estado del servicio Vault
func (s *HealthService) checkVaultService() struct {
	Status  string        `json:"status"`
	Latency time.Duration `json:"latency"`
	Error   string        `json:"error,omitempty"`
} {
	result := struct {
		Status  string        `json:"status"`
		Latency time.Duration `json:"latency"`
		Error   string        `json:"error,omitempty"`
	}{
		Status: "healthy",
	}

	// En una implementación real, esto verificaría la conexión a Vault
	// Por ahora, simulamos la verificación
	start := time.Now()
	
	// Simular verificación de Vault
	time.Sleep(10 * time.Millisecond) // Simular latencia
	
	result.Latency = time.Since(start)
	
	// Por ahora, asumimos que Vault está disponible
	// En producción, esto haría una llamada real a Vault

	return result
}

// getSystemInfo obtiene información del sistema
func (s *HealthService) getSystemInfo() *SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := &SystemInfo{
		GoVersion:    runtime.Version(),
		Architecture: runtime.GOARCH,
		OS:           runtime.GOOS,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
	}

	info.Memory.Alloc = m.Alloc
	info.Memory.TotalAlloc = m.TotalAlloc
	info.Memory.Sys = m.Sys
	info.Memory.NumGC = m.NumGC

	return info
}

// checkIntegrationsHealth verifica el estado de las integraciones
func (s *HealthService) checkIntegrationsHealth() map[string]interface{} {
	health := make(map[string]interface{})

	// En una implementación real, esto consultaría la base de datos
	// para obtener estadísticas de las integraciones
	health["total_integrations"] = 5
	health["active_integrations"] = 4
	health["error_integrations"] = 1
	health["platforms"] = map[string]int{
		"whatsapp":  2,
		"telegram":  1,
		"messenger": 1,
		"instagram": 1,
	}

	return health
}

// CheckLiveness verifica si el servicio está vivo
func (s *HealthService) CheckLiveness() *HealthStatus {
	status := &HealthStatus{
		Status:    "alive",
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).String(),
		Service:   "it-integration-service",
		Version:   "1.0.0",
	}

	// Verificación básica de que el proceso está ejecutándose
	// Si llegamos aquí, el proceso está vivo
	return status
}
