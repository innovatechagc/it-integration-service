package models

import (
	"time"
)

// PaymentRequest representa la solicitud de pago
type PaymentRequest struct {
	TransactionAmount float64        `json:"transaction_amount" binding:"required"`
	Token             string         `json:"token" binding:"required"`
	Description       string         `json:"description" binding:"required"`
	Installments      int            `json:"installments"`
	PaymentMethodID   string         `json:"payment_method_id" binding:"required"`
	Payer             Payer          `json:"payer" binding:"required"`
	ExternalReference string         `json:"external_reference"`
	NotificationURL   string         `json:"notification_url"`
	AdditionalInfo    AdditionalInfo `json:"additional_info,omitempty"`
}

// Payer representa la información del pagador
type Payer struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name,omitempty"`
}

// AdditionalInfo representa información adicional del pago
type AdditionalInfo struct {
	Items []Item `json:"items,omitempty"`
}

// Item representa un item del pago
type Item struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// PaymentResponse representa la respuesta de pago de Mercado Pago
type PaymentResponse struct {
	ID                        int64          `json:"id"`
	Status                    string         `json:"status"`
	StatusDetail              string         `json:"status_detail"`
	TransactionAmount         float64        `json:"transaction_amount"`
	TransactionAmountRefunded float64        `json:"transaction_amount_refunded"`
	CurrencyID                string         `json:"currency_id"`
	Description               string         `json:"description"`
	PaymentMethodID           string         `json:"payment_method_id"`
	PaymentTypeID             string         `json:"payment_type_id"`
	Installments              int            `json:"installments"`
	ExternalReference         string         `json:"external_reference"`
	DateCreated               time.Time      `json:"date_created"`
	DateLastUpdated           time.Time      `json:"date_last_updated"`
	Payer                     Payer          `json:"payer"`
	NotificationURL           string         `json:"notification_url"`
	AdditionalInfo            AdditionalInfo `json:"additional_info,omitempty"`
}

// WebhookNotification representa la notificación de webhook
type WebhookNotification struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

// PaymentStatus representa el estado del pago
type PaymentStatus struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

// ErrorResponse representa una respuesta de error
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}
