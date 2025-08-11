# Integración Tawk.to

## 📋 Descripción

Esta integración permite conectar Tawk.to como proveedor de webchat para el servicio de integración. Tawk.to es una plataforma de chat en vivo profesional que ofrece funcionalidades avanzadas como bots, analytics, y una experiencia de usuario superior.

## 🚀 Características

### ✅ Funcionalidades Implementadas

- **Configuración de Integración**: Setup completo de Tawk.to con validación de credenciales
- **Gestión de Configuración**: Obtener y actualizar configuraciones de Tawk.to
- **Webhooks**: Recepción y procesamiento de webhooks de Tawk.to
- **Analytics**: Obtención de estadísticas y métricas de chat
- **Sesiones**: Gestión de sesiones de chat activas
- **Validación de Firmas**: Seguridad con HMAC SHA256 para webhooks
- **Normalización de Mensajes**: Conversión a formato estándar del sistema

### 🔧 Configuración

#### Variables de Entorno

```bash
# Tawk.to Configuration
TAWKTO_API_KEY=your_tawkto_api_key_here
TAWKTO_BASE_URL=https://api.tawk.to
TAWKTO_WEBHOOK_SECRET=your_tawkto_webhook_secret_here
TAWKTO_WIDGET_ID=your_tawkto_widget_id_here
TAWKTO_PROPERTY_ID=your_tawkto_property_id_here
TAWKTO_VERIFY_TOKEN=your_tawkto_verify_token_here
```

#### Configuración de Tawk.to

```json
{
  "widget_id": "your_widget_id",
  "property_id": "your_property_id", 
  "api_key": "your_api_key",
  "base_url": "https://api.tawk.to",
  "custom_css": "optional_custom_css",
  "custom_js": "optional_custom_js",
  "greeting": "optional_greeting_message",
  "offline_msg": "optional_offline_message"
}
```

## 📡 Endpoints

### 1. Setup de Integración

**POST** `/api/v1/integrations/tawkto/setup`

Configura una nueva integración de Tawk.to para un tenant.

```bash
curl -X POST http://localhost:8082/api/v1/integrations/tawkto/setup \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "your-tenant-id",
    "config": {
      "widget_id": "your-widget-id",
      "property_id": "your-property-id",
      "api_key": "your-api-key",
      "base_url": "https://api.tawk.to"
    }
  }'
```

**Respuesta Exitosa:**
```json
{
  "success": true,
  "data": {
    "integration_id": "generated-id",
    "status": "active",
    "message": "Integración Tawk.to configurada exitosamente"
  }
}
```

### 2. Obtener Configuración

**GET** `/api/v1/integrations/tawkto/config/:tenant_id`

Obtiene la configuración actual de Tawk.to para un tenant.

```bash
curl -X GET http://localhost:8082/api/v1/integrations/tawkto/config/your-tenant-id
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "widget_id": "your-widget-id",
    "property_id": "your-property-id",
    "api_key": "your-api-key",
    "base_url": "https://api.tawk.to",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 3. Actualizar Configuración

**PUT** `/api/v1/integrations/tawkto/config/:tenant_id`

Actualiza la configuración de Tawk.to para un tenant.

```bash
curl -X PUT http://localhost:8082/api/v1/integrations/tawkto/config/your-tenant-id \
  -H "Content-Type: application/json" \
  -d '{
    "greeting": "¡Hola! ¿En qué puedo ayudarte?",
    "custom_css": ".tawk-min-container { background: #007bff; }"
  }'
```

### 4. Analytics

**GET** `/api/v1/integrations/tawkto/analytics/:tenant_id`

Obtiene analytics de Tawk.to para un período específico.

```bash
curl -X GET "http://localhost:8082/api/v1/integrations/tawkto/analytics/your-tenant-id?start_date=2024-01-01&end_date=2024-01-31"
```

**Parámetros:**
- `start_date`: Fecha de inicio (YYYY-MM-DD)
- `end_date`: Fecha de fin (YYYY-MM-DD)

### 5. Sesiones de Chat

**GET** `/api/v1/integrations/tawkto/sessions/:tenant_id`

Obtiene las sesiones de chat activas de Tawk.to.

```bash
curl -X GET "http://localhost:8082/api/v1/integrations/tawkto/sessions/your-tenant-id?limit=50"
```

**Parámetros:**
- `limit`: Número máximo de sesiones a retornar (default: 50)

### 6. Webhook

**POST** `/api/v1/integrations/webhooks/tawkto`

Endpoint para recibir webhooks de Tawk.to.

```bash
curl -X POST http://localhost:8082/api/v1/integrations/webhooks/tawkto \
  -H "Content-Type: application/json" \
  -H "X-Tawk-Signature: sha256=..." \
  -d '{
    "event": "chat_message",
    "timestamp": 1640995200,
    "visitor": {
      "id": "visitor-123",
      "name": "Juan Pérez",
      "email": "juan@example.com"
    },
    "chat": {
      "id": "chat-456",
      "session": "session-789",
      "status": "active",
      "messages": [
        {
          "id": "msg-001",
          "type": "text",
          "content": "Hola, necesito ayuda",
          "sender": "visitor",
          "timestamp": "2024-01-01T12:00:00Z"
        }
      ]
    }
  }'
```

## 🔐 Seguridad

### Validación de Webhooks

Los webhooks de Tawk.to se validan usando HMAC SHA256:

```go
// Calcular firma esperada
h := hmac.New(sha256.New, []byte(webhookSecret))
h.Write(payload)
expectedSignature := "sha256=" + hex.EncodeToString(h.Sum(nil))

// Comparar con firma recibida
if signature != expectedSignature {
    return fmt.Errorf("firma inválida")
}
```

### Encriptación de Datos Sensibles

Los datos sensibles como API keys se encriptan en la base de datos usando AES-GCM.

## 📊 Normalización de Mensajes

Los mensajes de Tawk.to se normalizan al formato estándar del sistema:

```go
type NormalizedMessage struct {
    Platform   domain.Platform        `json:"platform"`
    Sender     string                 `json:"sender"`
    Recipient  string                 `json:"recipient"`
    Content    *domain.MessageContent `json:"content"`
    Timestamp  int64                  `json:"timestamp"`
    MessageID  string                 `json:"message_id"`
    TenantID   string                 `json:"tenant_id"`
    ChannelID  string                 `json:"channel_id"`
    RawPayload json.RawMessage        `json:"raw_payload"`
}
```

### Mapeo de Campos

| Tawk.to | NormalizedMessage |
|---------|-------------------|
| `visitor.id` | `Recipient` |
| `visitor.name` | `Sender` (si es visitor) |
| `chat.id` | `ChannelID` |
| `messages[].content` | `Content.Text` |
| `messages[].type` | `Content.Type` |
| `messages[].timestamp` | `Timestamp` |
| `messages[].id` | `MessageID` |

## 🚀 Despliegue

### Variables de Entorno en Producción

```yaml
# cloudrun-production.yaml
- name: TAWKTO_API_KEY
  valueFrom:
    secretKeyRef:
      key: latest
      name: it-tawkto-api-key
- name: TAWKTO_BASE_URL
  value: "https://api.tawk.to"
- name: TAWKTO_WEBHOOK_SECRET
  valueFrom:
    secretKeyRef:
      key: latest
      name: it-tawkto-webhook-secret
- name: TAWKTO_WIDGET_ID
  valueFrom:
    secretKeyRef:
      key: latest
      name: it-tawkto-widget-id
- name: TAWKTO_PROPERTY_ID
  valueFrom:
    secretKeyRef:
      key: latest
      name: it-tawkto-property-id
- name: TAWKTO_VERIFY_TOKEN
  valueFrom:
    secretKeyRef:
      key: latest
      name: it-tawkto-verify-token
```

### Configuración de Secrets en Google Cloud

```bash
# Crear secrets para Tawk.to
echo -n "your-tawkto-api-key" | gcloud secrets create it-tawkto-api-key --data-file=-
echo -n "your-tawkto-webhook-secret" | gcloud secrets create it-tawkto-webhook-secret --data-file=-
echo -n "your-tawkto-widget-id" | gcloud secrets create it-tawkto-widget-id --data-file=-
echo -n "your-tawkto-property-id" | gcloud secrets create it-tawkto-property-id --data-file=-
echo -n "your-tawkto-verify-token" | gcloud secrets create it-tawkto-verify-token --data-file=-
```

## 🧪 Testing

### Probar Setup de Integración

```bash
# Test con credenciales de prueba
curl -X POST http://localhost:8082/api/v1/integrations/tawkto/setup \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test-tenant",
    "config": {
      "widget_id": "test-widget",
      "property_id": "test-property", 
      "api_key": "test-key",
      "base_url": "https://api.tawk.to"
    }
  }'
```

### Probar Webhook

```bash
# Simular webhook de Tawk.to
curl -X POST http://localhost:8082/api/v1/integrations/webhooks/tawkto \
  -H "Content-Type: application/json" \
  -d '{
    "event": "chat_message",
    "timestamp": 1640995200,
    "visitor": {
      "id": "test-visitor",
      "name": "Test User"
    },
    "chat": {
      "id": "test-chat",
      "session": "test-session",
      "status": "active",
      "messages": [
        {
          "id": "test-msg",
          "type": "text",
          "content": "Test message",
          "sender": "visitor",
          "timestamp": "2024-01-01T12:00:00Z"
        }
      ]
    }
  }'
```

## 📈 Monitoreo

### Métricas Disponibles

- **Webhook Processing**: Tiempo de procesamiento de webhooks
- **API Calls**: Llamadas a la API de Tawk.to
- **Error Rates**: Tasa de errores por endpoint
- **Integration Status**: Estado de las integraciones activas

### Logs

Los logs incluyen:
- Configuración de integraciones
- Procesamiento de webhooks
- Errores de API
- Validación de credenciales

## 🔄 Flujo de Integración

1. **Setup**: Configurar integración con credenciales de Tawk.to
2. **Validación**: Verificar credenciales con API de Tawk.to
3. **Webhook Setup**: Configurar webhook en Tawk.to automáticamente
4. **Recepción**: Recibir mensajes vía webhook
5. **Normalización**: Convertir a formato estándar
6. **Forwarding**: Enviar al servicio de mensajería
7. **Analytics**: Obtener estadísticas y métricas

## 🎯 Ventajas de Tawk.to

### Para el Negocio
- ✅ **Chat profesional** desde el día 1
- ✅ **Analytics avanzados** incluidos
- ✅ **Soporte técnico** disponible
- ✅ **Escalabilidad** automática

### Para el Desarrollo
- ✅ **Integración rápida** (1-2 días)
- ✅ **API robusta** y documentada
- ✅ **Webhooks confiables**
- ✅ **Menos código** para mantener

### Para el Usuario Final
- ✅ **Experiencia de chat** profesional
- ✅ **Funcionalidades avanzadas** (archivos, emojis, etc.)
- ✅ **Disponibilidad** 24/7
- ✅ **Integración** perfecta con el sistema

## 📞 Soporte

Para soporte técnico con Tawk.to:
- **Documentación**: https://developer.tawk.to/
- **API Reference**: https://developer.tawk.to/api/
- **Webhook Guide**: https://developer.tawk.to/webhooks/

---

**Estado**: ✅ **Implementado y Funcionando**
**Última Actualización**: Enero 2024
**Versión**: 1.0.0
