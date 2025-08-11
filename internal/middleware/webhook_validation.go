package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"it-integration-service/internal/config"
	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

type WebhookValidationMiddleware struct {
	config *config.Config
	logger logger.Logger
}

func NewWebhookValidationMiddleware(cfg *config.Config, logger logger.Logger) *WebhookValidationMiddleware {
	return &WebhookValidationMiddleware{
		config: cfg,
		logger: logger,
	}
}

// ValidateWebhookSignature valida la firma HMAC de los webhooks de Meta (WhatsApp, Messenger, Instagram)
func (m *WebhookValidationMiddleware) ValidateWebhookSignature(platform string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Para verificación de webhook (GET request), no validar firma
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		// Obtener el secret para la plataforma
		secret, exists := m.config.Integration.WebhookSecrets[platform]
		if !exists || secret == "" {
			m.logger.Error("Webhook secret not configured for platform", map[string]interface{}{
				"platform": platform,
			})
			c.JSON(http.StatusInternalServerError, domain.APIResponse{
				Code:    "CONFIGURATION_ERROR",
				Message: "Webhook secret not configured",
			})
			c.Abort()
			return
		}

		// Leer el body completo
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			m.logger.Error("Failed to read request body", map[string]interface{}{
				"platform": platform,
				"error":    err.Error(),
			})
			c.JSON(http.StatusBadRequest, domain.APIResponse{
				Code:    "INVALID_REQUEST",
				Message: "Failed to read request body",
			})
			c.Abort()
			return
		}

		// Restaurar el body para que otros handlers puedan leerlo
		c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

		// Obtener la firma del header
		signature := c.GetHeader("X-Hub-Signature-256")
		if signature == "" {
			m.logger.Error("Missing webhook signature", map[string]interface{}{
				"platform": platform,
			})
			c.JSON(http.StatusUnauthorized, domain.APIResponse{
				Code:    "UNAUTHORIZED",
				Message: "Missing webhook signature",
			})
			c.Abort()
			return
		}

		// Validar la firma
		if !m.validateHMACSignature(body, signature, secret) {
			m.logger.Error("Invalid webhook signature", map[string]interface{}{
				"platform": platform,
			})
			c.JSON(http.StatusUnauthorized, domain.APIResponse{
				Code:    "UNAUTHORIZED",
				Message: "Invalid webhook signature",
			})
			c.Abort()
			return
		}

		m.logger.Info("Webhook signature validated successfully", map[string]interface{}{
			"platform": platform,
		})

		c.Next()
	}
}

// ValidateWebhookVerification valida el token de verificación para configuración de webhooks
func (m *WebhookValidationMiddleware) ValidateWebhookVerification(platform string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Solo validar en requests GET (verificación de webhook)
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		mode := c.Query("hub.mode")
		if mode != "subscribe" {
			c.JSON(http.StatusBadRequest, domain.APIResponse{
				Code:    "INVALID_REQUEST",
				Message: "Invalid hub.mode",
			})
			c.Abort()
			return
		}

		token := c.Query("hub.verify_token")
		if token == "" {
			c.JSON(http.StatusBadRequest, domain.APIResponse{
				Code:    "INVALID_REQUEST",
				Message: "Missing hub.verify_token",
			})
			c.Abort()
			return
		}

		// Obtener el token de verificación configurado
		expectedToken, exists := m.config.Integration.WebhookVerifyTokens[platform]
		if !exists || expectedToken == "" {
			m.logger.Error("Webhook verify token not configured for platform", map[string]interface{}{
				"platform": platform,
			})
			c.JSON(http.StatusInternalServerError, domain.APIResponse{
				Code:    "CONFIGURATION_ERROR",
				Message: "Webhook verify token not configured",
			})
			c.Abort()
			return
		}

		// Validar el token
		if token != expectedToken {
			m.logger.Error("Invalid webhook verify token", map[string]interface{}{
				"platform": platform,
			})
			c.JSON(http.StatusForbidden, domain.APIResponse{
				Code:    "FORBIDDEN",
				Message: "Invalid webhook verify token",
			})
			c.Abort()
			return
		}

		// Responder con el challenge
		challenge := c.Query("hub.challenge")
		if challenge != "" {
			c.String(http.StatusOK, challenge)
		} else {
			c.JSON(http.StatusOK, domain.APIResponse{
				Code:    "SUCCESS",
				Message: "Webhook verification successful",
			})
		}

		m.logger.Info("Webhook verification successful", map[string]interface{}{
			"platform": platform,
		})

		c.Abort() // No continuar con otros handlers
	}
}

// ValidateTelegramWebhook valida webhooks de Telegram (no usa HMAC, solo secret token opcional)
func (m *WebhookValidationMiddleware) ValidateTelegramWebhook() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Telegram no usa HMAC, pero puede usar un secret token
		secretToken := c.GetHeader("X-Telegram-Bot-Api-Secret-Token")
		if secretToken != "" {
			expectedToken, exists := m.config.Integration.WebhookSecrets["telegram"]
			if exists && expectedToken != "" && secretToken != expectedToken {
				m.logger.Error("Invalid Telegram secret token")
				c.JSON(http.StatusUnauthorized, domain.APIResponse{
					Code:    "UNAUTHORIZED",
					Message: "Invalid secret token",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// validateHMACSignature valida una firma HMAC SHA256
func (m *WebhookValidationMiddleware) validateHMACSignature(payload []byte, signature, secret string) bool {
	// Remover prefijo "sha256=" si existe
	if strings.HasPrefix(signature, "sha256=") {
		signature = signature[7:]
	}

	// Calcular la firma esperada
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Comparar firmas de manera segura
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
