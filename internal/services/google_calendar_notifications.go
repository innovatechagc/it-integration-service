package services

import (
	"context"
	"fmt"
	"time"

	"it-integration-service/internal/domain"
	"it-integration-service/pkg/logger"
)

// NotificationService maneja las notificaciones automáticas para eventos de Google Calendar
type NotificationService struct {
	logger logger.Logger
	// TODO: Agregar clientes de servicios de mensajería existentes
	// whatsappClient *WhatsAppClient
	// telegramClient *TelegramClient
	// emailClient    *EmailClient
	// smsClient      *SMSClient
}

// NewNotificationService crea una nueva instancia del servicio de notificaciones
func NewNotificationService(logger logger.Logger) *NotificationService {
	return &NotificationService{
		logger: logger,
	}
}

// NotificationRequest representa una solicitud de notificación
type NotificationRequest struct {
	EventID           string                `json:"event_id"`
	TenantID          string                `json:"tenant_id"`
	ChannelID         string                `json:"channel_id"`
	EventSummary      string                `json:"event_summary"`
	EventDescription  string                `json:"event_description"`
	EventLocation     string                `json:"event_location"`
	StartTime         time.Time             `json:"start_time"`
	EndTime           time.Time             `json:"end_time"`
	Attendees         []CalendarAttendee    `json:"attendees"`
	NotificationType  NotificationType      `json:"notification_type"`
	ReminderMinutes   int                   `json:"reminder_minutes"`
	CustomMessage     string                `json:"custom_message,omitempty"`
}

// NotificationType define los tipos de notificación
type NotificationType string

const (
	NotificationTypeReminder    NotificationType = "reminder"
	NotificationTypeConfirmation NotificationType = "confirmation"
	NotificationTypeUpdate      NotificationType = "update"
	NotificationTypeCancellation NotificationType = "cancellation"
)

// NotificationChannel define los canales de notificación
type NotificationChannel string

const (
	NotificationChannelWhatsApp NotificationChannel = "whatsapp"
	NotificationChannelTelegram NotificationChannel = "telegram"
	NotificationChannelEmail    NotificationChannel = "email"
	NotificationChannelSMS      NotificationChannel = "sms"
)

// NotificationResult representa el resultado de una notificación
type NotificationResult struct {
	Success     bool                   `json:"success"`
	Channel     NotificationChannel    `json:"channel"`
	Recipient   string                 `json:"recipient"`
	MessageID   string                 `json:"message_id,omitempty"`
	Error       string                 `json:"error,omitempty"`
	SentAt      time.Time              `json:"sent_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendEventReminder envía recordatorios para un evento
func (s *NotificationService) SendEventReminder(ctx context.Context, req *NotificationRequest) ([]*NotificationResult, error) {
	s.logger.Info("Enviando recordatorio de evento", map[string]interface{}{
		"event_id":          req.EventID,
		"notification_type": req.NotificationType,
		"reminder_minutes":  req.ReminderMinutes,
		"attendees_count":   len(req.Attendees),
	})

	var results []*NotificationResult

	// Procesar cada asistente
	for _, attendee := range req.Attendees {
		// Determinar canales de notificación para este asistente
		channels := s.determineNotificationChannels(attendee)

		for _, channel := range channels {
			result := s.sendNotification(ctx, req, attendee, channel)
			results = append(results, result)
		}
	}

	s.logger.Info("Recordatorios enviados", map[string]interface{}{
		"event_id":      req.EventID,
		"total_sent":    len(results),
		"success_count": s.countSuccessfulResults(results),
	})

	return results, nil
}

// SendEventConfirmation envía confirmaciones de asistencia
func (s *NotificationService) SendEventConfirmation(ctx context.Context, req *NotificationRequest) ([]*NotificationResult, error) {
	s.logger.Info("Enviando confirmación de evento", map[string]interface{}{
		"event_id": req.EventID,
		"attendees_count": len(req.Attendees),
	})

	var results []*NotificationResult

	for _, attendee := range req.Attendees {
		channels := s.determineNotificationChannels(attendee)

		for _, channel := range channels {
			result := s.sendNotification(ctx, req, attendee, channel)
			results = append(results, result)
		}
	}

	return results, nil
}

// SendEventUpdate envía notificaciones de actualización de evento
func (s *NotificationService) SendEventUpdate(ctx context.Context, req *NotificationRequest) ([]*NotificationResult, error) {
	s.logger.Info("Enviando notificación de actualización de evento", map[string]interface{}{
		"event_id": req.EventID,
	})

	var results []*NotificationResult

	for _, attendee := range req.Attendees {
		channels := s.determineNotificationChannels(attendee)

		for _, channel := range channels {
			result := s.sendNotification(ctx, req, attendee, channel)
			results = append(results, result)
		}
	}

	return results, nil
}

// SendEventCancellation envía notificaciones de cancelación de evento
func (s *NotificationService) SendEventCancellation(ctx context.Context, req *NotificationRequest) ([]*NotificationResult, error) {
	s.logger.Info("Enviando notificación de cancelación de evento", map[string]interface{}{
		"event_id": req.EventID,
	})

	var results []*NotificationResult

	for _, attendee := range req.Attendees {
		channels := s.determineNotificationChannels(attendee)

		for _, channel := range channels {
			result := s.sendNotification(ctx, req, attendee, channel)
			results = append(results, result)
		}
	}

	return results, nil
}

// ScheduleReminders programa recordatorios automáticos para un evento
func (s *NotificationService) ScheduleReminders(ctx context.Context, event *domain.CalendarEvent, reminderMinutes []int) error {
	s.logger.Info("Programando recordatorios automáticos", map[string]interface{}{
		"event_id":         event.ID,
		"reminder_minutes": reminderMinutes,
	})

	for _, minutes := range reminderMinutes {
		reminderTime := event.StartTime.Add(-time.Duration(minutes) * time.Minute)
		
		// Solo programar si el recordatorio es en el futuro
		if reminderTime.After(time.Now()) {
			go s.scheduleReminder(ctx, event, minutes, reminderTime)
		}
	}

	return nil
}

// ProcessWebhookNotification procesa notificaciones de webhook y envía alertas
func (s *NotificationService) ProcessWebhookNotification(ctx context.Context, notification *domain.WebhookNotification) error {
	s.logger.Info("Procesando notificación de webhook", map[string]interface{}{
		"resource_id":  notification.ResourceID,
		"resource_uri": notification.ResourceURI,
		"state":        notification.State,
	})

	// TODO: Implementar lógica específica según el tipo de notificación
	// - Evento creado: enviar confirmaciones
	// - Evento actualizado: enviar notificaciones de cambio
	// - Evento cancelado: enviar notificaciones de cancelación

	return nil
}

// Helper methods

// determineNotificationChannels determina los canales de notificación para un asistente
func (s *NotificationService) determineNotificationChannels(attendee domain.CalendarAttendee) []NotificationChannel {
	var channels []NotificationChannel

	// Lógica para determinar canales basada en preferencias del asistente
	// Por ahora, usar canales por defecto
	if attendee.Email != "" {
		channels = append(channels, NotificationChannelEmail)
	}

	// TODO: Agregar lógica para determinar WhatsApp/Telegram basada en configuración
	// if attendee.HasWhatsApp {
	//     channels = append(channels, NotificationChannelWhatsApp)
	// }
	// if attendee.HasTelegram {
	//     channels = append(channels, NotificationChannelTelegram)
	// }

	return channels
}

// sendNotification envía una notificación por un canal específico
func (s *NotificationService) sendNotification(ctx context.Context, req *NotificationRequest, attendee domain.CalendarAttendee, channel NotificationChannel) *NotificationResult {
	result := &NotificationResult{
		Channel:   channel,
		Recipient: attendee.Email,
		SentAt:    time.Now(),
	}

	message := s.buildNotificationMessage(req, attendee, channel)

	switch channel {
	case NotificationChannelEmail:
		result = s.sendEmailNotification(ctx, attendee.Email, message, req)
	case NotificationChannelWhatsApp:
		result = s.sendWhatsAppNotification(ctx, attendee.Email, message, req)
	case NotificationChannelTelegram:
		result = s.sendTelegramNotification(ctx, attendee.Email, message, req)
	case NotificationChannelSMS:
		result = s.sendSMSNotification(ctx, attendee.Email, message, req)
	default:
		result.Success = false
		result.Error = fmt.Sprintf("canal no soportado: %s", channel)
	}

	return result
}

// buildNotificationMessage construye el mensaje de notificación
func (s *NotificationService) buildNotificationMessage(req *NotificationRequest, attendee domain.CalendarAttendee, channel NotificationChannel) string {
	var message string

	switch req.NotificationType {
	case NotificationTypeReminder:
		message = s.buildReminderMessage(req, attendee, channel)
	case NotificationTypeConfirmation:
		message = s.buildConfirmationMessage(req, attendee, channel)
	case NotificationTypeUpdate:
		message = s.buildUpdateMessage(req, attendee, channel)
	case NotificationTypeCancellation:
		message = s.buildCancellationMessage(req, attendee, channel)
	}

	if req.CustomMessage != "" {
		message += "\n\n" + req.CustomMessage
	}

	return message
}

// buildReminderMessage construye mensaje de recordatorio
func (s *NotificationService) buildReminderMessage(req *NotificationRequest, attendee domain.CalendarAttendee, channel NotificationChannel) string {
	timeStr := req.StartTime.Format("15:04")
	dateStr := req.StartTime.Format("02/01/2006")

	switch channel {
	case NotificationChannelWhatsApp, NotificationChannelTelegram:
		return fmt.Sprintf("🔔 *Recordatorio de evento*\n\n"+
			"*%s*\n"+
			"📅 %s a las %s\n"+
			"📍 %s\n\n"+
			"Te recordamos que tienes este evento en %d minutos.",
			req.EventSummary, dateStr, timeStr, req.EventLocation, req.ReminderMinutes)
	case NotificationChannelEmail:
		return fmt.Sprintf("Recordatorio de evento: %s\n\n"+
			"Fecha: %s\n"+
			"Hora: %s\n"+
			"Ubicación: %s\n\n"+
			"Este evento comienza en %d minutos.",
			req.EventSummary, dateStr, timeStr, req.EventLocation, req.ReminderMinutes)
	default:
		return fmt.Sprintf("Recordatorio: %s - %s a las %s", req.EventSummary, dateStr, timeStr)
	}
}

// buildConfirmationMessage construye mensaje de confirmación
func (s *NotificationService) buildConfirmationMessage(req *NotificationRequest, attendee domain.CalendarAttendee, channel NotificationChannel) string {
	timeStr := req.StartTime.Format("15:04")
	dateStr := req.StartTime.Format("02/01/2006")

	switch channel {
	case NotificationChannelWhatsApp, NotificationChannelTelegram:
		return fmt.Sprintf("✅ *Evento confirmado*\n\n"+
			"*%s*\n"+
			"📅 %s a las %s\n"+
			"📍 %s\n\n"+
			"Tu evento ha sido confirmado.",
			req.EventSummary, dateStr, timeStr, req.EventLocation)
	case NotificationChannelEmail:
		return fmt.Sprintf("Evento confirmado: %s\n\n"+
			"Fecha: %s\n"+
			"Hora: %s\n"+
			"Ubicación: %s\n\n"+
			"Tu evento ha sido confirmado exitosamente.",
			req.EventSummary, dateStr, timeStr, req.EventLocation)
	default:
		return fmt.Sprintf("Confirmado: %s - %s a las %s", req.EventSummary, dateStr, timeStr)
	}
}

// buildUpdateMessage construye mensaje de actualización
func (s *NotificationService) buildUpdateMessage(req *NotificationRequest, attendee domain.CalendarAttendee, channel NotificationChannel) string {
	timeStr := req.StartTime.Format("15:04")
	dateStr := req.StartTime.Format("02/01/2006")

	switch channel {
	case NotificationChannelWhatsApp, NotificationChannelTelegram:
		return fmt.Sprintf("🔄 *Evento actualizado*\n\n"+
			"*%s*\n"+
			"📅 %s a las %s\n"+
			"📍 %s\n\n"+
			"Tu evento ha sido actualizado.",
			req.EventSummary, dateStr, timeStr, req.EventLocation)
	case NotificationChannelEmail:
		return fmt.Sprintf("Evento actualizado: %s\n\n"+
			"Fecha: %s\n"+
			"Hora: %s\n"+
			"Ubicación: %s\n\n"+
			"Tu evento ha sido actualizado.",
			req.EventSummary, dateStr, timeStr, req.EventLocation)
	default:
		return fmt.Sprintf("Actualizado: %s - %s a las %s", req.EventSummary, dateStr, timeStr)
	}
}

// buildCancellationMessage construye mensaje de cancelación
func (s *NotificationService) buildCancellationMessage(req *NotificationRequest, attendee domain.CalendarAttendee, channel NotificationChannel) string {
	timeStr := req.StartTime.Format("15:04")
	dateStr := req.StartTime.Format("02/01/2006")

	switch channel {
	case NotificationChannelWhatsApp, NotificationChannelTelegram:
		return fmt.Sprintf("❌ *Evento cancelado*\n\n"+
			"*%s*\n"+
			"📅 %s a las %s\n\n"+
			"Tu evento ha sido cancelado.",
			req.EventSummary, dateStr, timeStr)
	case NotificationChannelEmail:
		return fmt.Sprintf("Evento cancelado: %s\n\n"+
			"Fecha: %s\n"+
			"Hora: %s\n\n"+
			"Tu evento ha sido cancelado.",
			req.EventSummary, dateStr, timeStr)
	default:
		return fmt.Sprintf("Cancelado: %s - %s a las %s", req.EventSummary, dateStr, timeStr)
	}
}

// sendEmailNotification envía notificación por email
func (s *NotificationService) sendEmailNotification(ctx context.Context, recipient, message string, req *NotificationRequest) *NotificationResult {
	result := &NotificationResult{
		Channel:   NotificationChannelEmail,
		Recipient: recipient,
		SentAt:    time.Now(),
	}

	// TODO: Integrar con servicio de email existente
	s.logger.Info("Enviando notificación por email", map[string]interface{}{
		"recipient": recipient,
		"event_id":  req.EventID,
	})

	// Simulación de envío exitoso
	result.Success = true
	result.MessageID = fmt.Sprintf("email_%s_%d", req.EventID, time.Now().Unix())

	return result
}

// sendWhatsAppNotification envía notificación por WhatsApp
func (s *NotificationService) sendWhatsAppNotification(ctx context.Context, recipient, message string, req *NotificationRequest) *NotificationResult {
	result := &NotificationResult{
		Channel:   NotificationChannelWhatsApp,
		Recipient: recipient,
		SentAt:    time.Now(),
	}

	// TODO: Integrar con servicio de WhatsApp existente
	s.logger.Info("Enviando notificación por WhatsApp", map[string]interface{}{
		"recipient": recipient,
		"event_id":  req.EventID,
	})

	// Simulación de envío exitoso
	result.Success = true
	result.MessageID = fmt.Sprintf("whatsapp_%s_%d", req.EventID, time.Now().Unix())

	return result
}

// sendTelegramNotification envía notificación por Telegram
func (s *NotificationService) sendTelegramNotification(ctx context.Context, recipient, message string, req *NotificationRequest) *NotificationResult {
	result := &NotificationResult{
		Channel:   NotificationChannelTelegram,
		Recipient: recipient,
		SentAt:    time.Now(),
	}

	// TODO: Integrar con servicio de Telegram existente
	s.logger.Info("Enviando notificación por Telegram", map[string]interface{}{
		"recipient": recipient,
		"event_id":  req.EventID,
	})

	// Simulación de envío exitoso
	result.Success = true
	result.MessageID = fmt.Sprintf("telegram_%s_%d", req.EventID, time.Now().Unix())

	return result
}

// sendSMSNotification envía notificación por SMS
func (s *NotificationService) sendSMSNotification(ctx context.Context, recipient, message string, req *NotificationRequest) *NotificationResult {
	result := &NotificationResult{
		Channel:   NotificationChannelSMS,
		Recipient: recipient,
		SentAt:    time.Now(),
	}

	// TODO: Integrar con servicio de SMS existente
	s.logger.Info("Enviando notificación por SMS", map[string]interface{}{
		"recipient": recipient,
		"event_id":  req.EventID,
	})

	// Simulación de envío exitoso
	result.Success = true
	result.MessageID = fmt.Sprintf("sms_%s_%d", req.EventID, time.Now().Unix())

	return result
}

// scheduleReminder programa un recordatorio para ejecutarse en el futuro
func (s *NotificationService) scheduleReminder(ctx context.Context, event *domain.CalendarEvent, minutes int, reminderTime time.Time) {
	// Calcular tiempo de espera
	waitTime := time.Until(reminderTime)
	if waitTime <= 0 {
		return
	}

	// Esperar hasta el momento del recordatorio
	time.Sleep(waitTime)

	// Enviar recordatorio
	req := &NotificationRequest{
		EventID:          event.ID,
		TenantID:         event.TenantID,
		ChannelID:        event.ChannelID,
		EventSummary:     event.Summary,
		EventDescription: event.Description,
		EventLocation:    event.Location,
		StartTime:        event.StartTime,
		EndTime:          event.EndTime,
		Attendees:        event.Attendees,
		NotificationType: NotificationTypeReminder,
		ReminderMinutes:  minutes,
	}

	_, err := s.SendEventReminder(ctx, req)
	if err != nil {
		s.logger.Error("Error enviando recordatorio programado", err, map[string]interface{}{
			"event_id": event.ID,
			"minutes":  minutes,
		})
	}
}

// countSuccessfulResults cuenta los resultados exitosos
func (s *NotificationService) countSuccessfulResults(results []*NotificationResult) int {
	count := 0
	for _, result := range results {
		if result.Success {
			count++
		}
	}
	return count
}
