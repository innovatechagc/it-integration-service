package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Métricas de HTTP
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status", "platform"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "platform"},
	)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// Métricas de Webhooks
	webhookRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_requests_total",
			Help: "Total number of webhook requests",
		},
		[]string{"platform", "status", "tenant_id"},
	)

	webhookProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "webhook_processing_duration_seconds",
			Help:    "Webhook processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"platform", "tenant_id"},
	)

	webhookPayloadSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "webhook_payload_size_bytes",
			Help:    "Webhook payload size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 2, 10),
		},
		[]string{"platform"},
	)

	// Métricas de Integraciones
	integrationsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "integrations_total",
			Help: "Total number of integrations by platform and status",
		},
		[]string{"platform", "status", "tenant_id"},
	)

	integrationSetupDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "integration_setup_duration_seconds",
			Help:    "Integration setup duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"platform", "tenant_id"},
	)

	// Métricas de Base de Datos
	databaseConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "Database connection pool statistics",
		},
		[]string{"state"},
	)

	databaseQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Métricas de Servicios Externos
	externalServiceRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "external_service_requests_total",
			Help: "Total number of external service requests",
		},
		[]string{"service", "method", "status"},
	)

	externalServiceDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "external_service_duration_seconds",
			Help:    "External service request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	// Métricas de Errores
	errorRate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors by type",
		},
		[]string{"type", "platform", "tenant_id"},
	)

	// Métricas de Rate Limiting
	rateLimitHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"endpoint", "ip"},
	)
)

func init() {
	// Registrar métricas
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		httpRequestsInFlight,
		webhookRequestsTotal,
		webhookProcessingDuration,
		webhookPayloadSize,
		integrationsTotal,
		integrationSetupDuration,
		databaseConnections,
		databaseQueryDuration,
		externalServiceRequests,
		externalServiceDuration,
		errorRate,
		rateLimitHits,
	)
}

// Metrics middleware para métricas de HTTP
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Incrementar requests en vuelo
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		// Procesar request
		c.Next()

		// Registrar métricas
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		// Determinar plataforma basado en la ruta
		platform := getPlatformFromPath(path)

		httpRequestsTotal.WithLabelValues(method, path, status, platform).Inc()
		httpRequestDuration.WithLabelValues(method, path, platform).Observe(duration)

		// Registrar errores
		if c.Writer.Status() >= 400 {
			errorType := "http_error"
			if c.Writer.Status() >= 500 {
				errorType = "server_error"
			}
			errorRate.WithLabelValues(errorType, platform, getTenantID(c)).Inc()
		}
	}
}

// WebhookMetrics middleware específico para métricas de webhooks
func WebhookMetrics(platform string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		tenantID := getTenantID(c)

		// Registrar tamaño del payload
		if c.Request.ContentLength > 0 {
			webhookPayloadSize.WithLabelValues(platform).Observe(float64(c.Request.ContentLength))
		}

		// Procesar request
		c.Next()

		// Registrar métricas de webhook
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		webhookRequestsTotal.WithLabelValues(platform, status, tenantID).Inc()
		webhookProcessingDuration.WithLabelValues(platform, tenantID).Observe(duration)

		// Registrar errores de webhook
		if c.Writer.Status() >= 400 {
			errorRate.WithLabelValues("webhook_error", platform, tenantID).Inc()
		}
	}
}

// DatabaseMetrics middleware para métricas de base de datos
func DatabaseMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Este middleware se puede usar para registrar métricas de base de datos
		// cuando se ejecuten operaciones de DB
		c.Next()
	}
}

// ExternalServiceMetrics middleware para métricas de servicios externos
func ExternalServiceMetrics(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method

		// Procesar request
		c.Next()

		// Registrar métricas
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		externalServiceRequests.WithLabelValues(service, method, status).Inc()
		externalServiceDuration.WithLabelValues(service, method).Observe(duration)
	}
}

// MetricsHandler retorna el handler de Prometheus
func MetricsHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// Helper functions
func getPlatformFromPath(path string) string {
	if len(path) == 0 {
		return "unknown"
	}

	// Extraer plataforma de la ruta
	if len(path) > 20 && path[:20] == "/api/v1/integrations" {
		// Buscar plataforma en la ruta
		platforms := []string{"whatsapp", "telegram", "messenger", "instagram", "webchat"}
		for _, platform := range platforms {
			if len(path) > 20+len(platform) && path[20:20+len(platform)] == platform {
				return platform
			}
		}
	}

	return "api"
}

func getTenantID(c *gin.Context) string {
	// Intentar obtener tenant_id de diferentes fuentes
	if tenantID := c.Query("tenant_id"); tenantID != "" {
		return tenantID
	}
	if tenantID := c.Param("tenant_id"); tenantID != "" {
		return tenantID
	}
	if tenantID := c.GetHeader("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}
	return "unknown"
}

// UpdateIntegrationMetrics actualiza métricas de integraciones
func UpdateIntegrationMetrics(platform, status, tenantID string) {
	integrationsTotal.WithLabelValues(platform, status, tenantID).Inc()
}

// UpdateDatabaseMetrics actualiza métricas de base de datos
func UpdateDatabaseMetrics(operation, table string, duration time.Duration) {
	databaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// UpdateDatabaseConnections actualiza métricas de conexiones de base de datos
func UpdateDatabaseConnections(open, inUse, idle int) {
	databaseConnections.WithLabelValues("open").Set(float64(open))
	databaseConnections.WithLabelValues("in_use").Set(float64(inUse))
	databaseConnections.WithLabelValues("idle").Set(float64(idle))
}

// UpdateIntegrationSetupMetrics actualiza métricas de configuración de integraciones
func UpdateIntegrationSetupMetrics(platform, tenantID string, duration time.Duration) {
	integrationSetupDuration.WithLabelValues(platform, tenantID).Observe(duration.Seconds())
}

// UpdateRateLimitMetrics actualiza métricas de rate limiting
func UpdateRateLimitMetrics(endpoint, ip string) {
	rateLimitHits.WithLabelValues(endpoint, ip).Inc()
}