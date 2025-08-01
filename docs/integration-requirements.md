# 📋 Requerimientos de Integración - Checklist

Este documento detalla qué necesitas implementar o configurar para que el servicio de integraciones funcione completamente.

## 🔧 Configuraciones Pendientes

### 1. Base de Datos - Repositorios

**Estado**: ❌ Pendiente  
**Ubicación**: `internal/repositories/`

Necesitas implementar los repositorios reales para:

```go
// internal/repositories/channel_repository.go
type ChannelRepository interface {
    GetByTenant(ctx context.Context, tenantID string) ([]*domain.ChannelIntegration, error)
    GetByID(ctx context.Context, id string) (*domain.ChannelIntegration, error)
    Create(ctx context.Context, channel *domain.ChannelIntegration) error
    Update(ctx context.Context, channel *domain.ChannelIntegration) error
    Delete(ctx context.Context, id string) error
}

// internal/repositories/message_repository.go
type InboundMessageRepository interface {
    Create(ctx context.Context, message *domain.InboundMessage) error
    GetByPlatform(ctx context.Context, platform domain.Platform) ([]*domain.InboundMessage, error)
}

type OutboundMessageRepository interface {
    Create(ctx context.Context, log *domain.OutboundMessageLog) error
    GetByChannel(ctx context.Context, channelID string) ([]*domain.OutboundMessageLog, error)
}
```

### 2. Servicios de Mensajería - Implementaciones Reales

**Estado**: ❌ Pendiente  
**Ubicación**: `internal/services/provider_impl.go`

Actualmente tienes mocks. Necesitas implementar:

#### WhatsApp (Meta)
```go
func (s *MessagingProviderService) SendWhatsAppMessage(ctx context.Context, config map[string]interface{}, recipient string, content domain.MessageContent) error {
    // Implementar llamada real a Graph API
    // POST https://graph.facebook.com/v18.0/{phone_number_id}/messages
}
```

#### Telegram
```go
func (s *MessagingProviderService) SendTelegramMessage(ctx context.Context, config map[string]interface{}, recipient string, content domain.MessageContent) error {
    // Implementar llamada real a Telegram Bot API
    // POST https://api.telegram.org/bot{token}/sendMessage
}
```

#### Twilio WhatsApp
```go
func (s *MessagingProviderService) SendTwilioWhatsAppMessage(ctx context.Context, config map[string]interface{}, recipient string, content domain.MessageContent) error {
    // Implementar llamada real a Twilio API
    // POST https://api.twilio.com/2010-04-01/Accounts/{AccountSid}/Messages.json
}
```

### 3. Validación de Webhooks

**Estado**: ❌ Pendiente  
**Ubicación**: `internal/services/webhook_impl.go`

Implementar validación de firmas:

```go
func (s *WebhookService) ValidateWhatsAppSignature(payload []byte, signature string, secret string) bool {
    // Validar X-Hub-Signature-256 de Meta
    expectedSignature := "sha256=" + hmac.SHA256(payload, secret)
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (s *WebhookService) ValidateTelegramWebhook(payload []byte, secretToken string) bool {
    // Validar secret token de Telegram si está configurado
}
```

### 4. Configuración de Secrets

**Estado**: ❌ Pendiente  
**Ubicación**: `.env.local` y configuración externa

Necesitas configurar tokens reales:

```bash
# Meta/Facebook
META_APP_ID=tu_app_id_real
META_APP_SECRET=tu_app_secret_real
META_VERIFY_TOKEN=tu_verify_token

# Telegram
TELEGRAM_BOT_TOKEN=tu_bot_token_real

# Twilio
TWILIO_ACCOUNT_SID=tu_account_sid
TWILIO_AUTH_TOKEN=tu_auth_token

# 360Dialog
DIALOG_360_API_KEY=tu_api_key
```

## 🚀 Funcionalidades que Ya Funcionan

### ✅ Estructura Base
- Modelos de dominio completos
- Handlers HTTP implementados
- Rutas configuradas
- Middleware básico

### ✅ Health Checks
- `/api/v1/health` - Estado del servicio
- `/api/v1/ready` - Readiness check

### ✅ CRUD de Integraciones
- Crear, leer, actualizar, eliminar canales
- Validación de datos de entrada
- Respuestas estructuradas

### ✅ Webhooks Endpoints
- WhatsApp, Telegram, Messenger, Instagram, Webchat
- Procesamiento de payloads
- Logging estructurado

## 🔨 Próximos Pasos Recomendados

### Paso 1: Implementar Repositorios
```bash
# Crear archivos de repositorio
touch internal/repositories/channel_repository.go
touch internal/repositories/message_repository.go
touch internal/repositories/implementations/postgres_channel_repo.go
```

### Paso 2: Configurar Base de Datos
```bash
# Ejecutar migraciones (si no están aplicadas)
# Las tablas ya están definidas en scripts/init-test.sql
```

### Paso 3: Implementar Proveedores Reales
```bash
# Actualizar provider_impl.go con llamadas HTTP reales
# Agregar manejo de errores específicos de cada proveedor
```

### Paso 4: Testing con Tokens Reales
```bash
# Configurar tokens de desarrollo en .env.local
# Probar con Postman usando la colección creada
```

## 📝 Casos de Prueba Prioritarios

### 1. Flujo Completo WhatsApp
1. Crear integración WhatsApp Meta
2. Configurar webhook en Meta Developer Console
3. Enviar mensaje desde WhatsApp → Recibir webhook
4. Enviar respuesta desde API → Usuario recibe mensaje

### 2. Flujo Completo Telegram
1. Crear bot con @BotFather
2. Configurar webhook del bot
3. Enviar mensaje al bot → Recibir webhook
4. Enviar respuesta desde API → Usuario recibe mensaje

### 3. Testing de Errores
1. Tokens inválidos
2. Webhooks con firmas incorrectas
3. Payloads malformados
4. Rate limiting

## 🔍 Herramientas de Debug

### Logs Estructurados
```bash
# Ver logs en tiempo real
make dev-logs

# Filtrar por tipo de evento
docker-compose logs app | grep "webhook"
docker-compose logs app | grep "ERROR"
```

### Webhook Testing
```bash
# Usar simulador web
open http://localhost:8081

# Usar script de testing
./scripts/test-integrations.sh whatsapp
./scripts/test-integrations.sh telegram
```

### Postman Collection
- Importar `postman/Integration-Service.postman_collection.json`
- Configurar environment con `postman/Integration-Service.postman_environment.json`
- Ejecutar tests en orden: Health → Create Channel → Send Message → Webhook

## 🚨 Consideraciones de Seguridad

### Validación de Webhooks
- ✅ Implementar validación de firmas HMAC
- ✅ Verificar tokens de verificación
- ✅ Rate limiting por IP/tenant
- ✅ Sanitización de payloads

### Manejo de Secrets
- ✅ No logear tokens en logs
- ✅ Encriptar access_tokens en base de datos
- ✅ Rotar tokens periódicamente
- ✅ Usar HTTPS en producción

### Monitoreo
- ✅ Alertas por errores de webhook
- ✅ Métricas de latencia y throughput
- ✅ Logs de auditoría para cambios de configuración

## 📊 Métricas Importantes

### KPIs del Servicio
- Webhooks procesados/minuto
- Latencia promedio de procesamiento
- Tasa de error por plataforma
- Mensajes enviados exitosamente
- Uptime del servicio

### Dashboards Recomendados
- Estado de integraciones por tenant
- Volumen de mensajes por plataforma
- Errores y reintentos
- Performance de APIs externas