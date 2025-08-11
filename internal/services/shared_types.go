package services

import (
	"encoding/json"
)

// MetaAPIResponse representa una respuesta de la API de Meta/Facebook
type MetaAPIResponse struct {
	Data  json.RawMessage `json:"data,omitempty"`
	Error *MetaAPIError   `json:"error,omitempty"`
}

// MetaAPIError representa un error de la API de Meta/Facebook
type MetaAPIError struct {
	Message   string `json:"message"`
	Type      string `json:"type"`
	Code      int    `json:"code"`
	ErrorData struct {
		MessagingProduct string `json:"messaging_product"`
		Details          string `json:"details"`
	} `json:"error_data,omitempty"`
}

// FacebookAPIResponse representa una respuesta de la API de Facebook (alias para compatibilidad)
type FacebookAPIResponse = MetaAPIResponse

// FacebookAPIError representa un error de la API de Facebook (alias para compatibilidad)
type FacebookAPIError = MetaAPIError

// WebhookSubscription representa una suscripción de webhook genérica
type WebhookSubscription struct {
	Object string   `json:"object"`
	Fields []string `json:"fields"`
}
