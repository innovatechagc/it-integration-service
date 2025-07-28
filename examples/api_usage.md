# Integration Service API Usage Examples

## 1. Crear una Integración de WhatsApp

```bash
curl -X POST http://localhost:8080/api/v1/integrations/channels \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-123",
    "platform": "whatsapp",
    "provider": "meta",
    "access_token": "your-whatsapp-access-token",
    "webhook_url": "https://your-app.com/webhooks/whatsapp",
    "config": {
      "phone_number_id": "123456789",
      "business_id": "987654321"
    }
  }'
```

## 2. Listar Integraciones por Tenant

```bash
curl -X GET "http://localhost:8080/api/v1/integrations/channels?tenant_id=tenant-123"
```

## 3. Enviar Mensaje por WhatsApp

```bash
curl -X POST http://localhost:8080/api/v1/integrations/send \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "channel-uuid-here",
    "recipient": "573001112233",
    "content": {
      "type": "text",
      "text": "¡Hola! Bienvenido a nuestro servicio de atención al cliente."
    }
  }'
```

## 4. Configurar Webhook de WhatsApp (Meta)

### Verificación inicial (GET)
Meta enviará una verificación GET a tu webhook:
```
GET /api/v1/integrations/webhooks/whatsapp?hub.mode=subscribe&hub.verify_token=tu-token&hub.challenge=123456
```

### Recibir mensajes (POST)
```bash
# Ejemplo de payload que Meta enviará a tu webhook
curl -X POST http://localhost:8080/api/v1/integrations/webhooks/whatsapp \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=signature-here" \
  -d '{
    "entry": [{
      "changes": [{
        "value": {
          "messages": [{
            "id": "wamid.123",
            "from": "573001112233",
            "timestamp": "1640995200",
            "text": {
              "body": "Hola, necesito ayuda"
            },
            "type": "text"
          }],
          "metadata": {
            "phone_number_id": "123456789"
          }
        }
      }]
    }]
  }'
```

## 5. Configurar Integración de Telegram

```bash
curl -X POST http://localhost:8080/api/v1/integrations/channels \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-123",
    "platform": "telegram",
    "provider": "custom",
    "access_token": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
    "webhook_url": "https://your-app.com/webhooks/telegram",
    "config": {
      "bot_token": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
    }
  }'
```

## 6. Webhook de Telegram

```bash
curl -X POST http://localhost:8080/api/v1/integrations/webhooks/telegram \
  -H "Content-Type: application/json" \
  -d '{
    "message": {
      "message_id": 123,
      "from": {
        "id": 987654321,
        "username": "usuario123"
      },
      "chat": {
        "id": 987654321
      },
      "date": 1640995200,
      "text": "Hola bot!"
    }
  }'
```

## 7. Configurar Integración de Messenger

```bash
curl -X POST http://localhost:8080/api/v1/integrations/channels \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-123",
    "platform": "messenger",
    "provider": "meta",
    "access_token": "your-page-access-token",
    "webhook_url": "https://your-app.com/webhooks/messenger",
    "config": {
      "page_id": "123456789"
    }
  }'
```

## 8. Webhook de Messenger

```bash
curl -X POST http://localhost:8080/api/v1/integrations/webhooks/messenger \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=signature-here" \
  -d '{
    "entry": [{
      "messaging": [{
        "sender": {
          "id": "987654321"
        },
        "recipient": {
          "id": "123456789"
        },
        "timestamp": 1640995200,
        "message": {
          "mid": "mid.123",
          "text": "Hola desde Messenger"
        }
      }]
    }]
  }'
```

## 9. Actualizar Integración

```bash
curl -X PATCH http://localhost:8080/api/v1/integrations/channels/channel-uuid-here \
  -H "Content-Type: application/json" \
  -d '{
    "status": "disabled",
    "config": {
      "phone_number_id": "new-phone-number-id"
    }
  }'
```

## 10. Eliminar Integración

```bash
curl -X DELETE http://localhost:8080/api/v1/integrations/channels/channel-uuid-here
```

## Respuestas de la API

### Respuesta Exitosa
```json
{
  "code": "SUCCESS",
  "message": "Operation completed successfully",
  "data": {
    "id": "channel-uuid",
    "tenant_id": "tenant-123",
    "platform": "whatsapp",
    "provider": "meta",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Respuesta de Error
```json
{
  "code": "INVALID_REQUEST",
  "message": "tenant_id is required",
  "data": null
}
```

## Códigos de Estado HTTP

- `200 OK` - Operación exitosa
- `201 Created` - Recurso creado exitosamente
- `400 Bad Request` - Solicitud inválida
- `404 Not Found` - Recurso no encontrado
- `500 Internal Server Error` - Error interno del servidor

## Plataformas Soportadas

- **WhatsApp**: Meta Business API, 360Dialog, Twilio
- **Messenger**: Meta Graph API
- **Instagram**: Meta Graph API
- **Telegram**: Bot API
- **Webchat**: API personalizada

## Configuración de Webhooks por Plataforma

### WhatsApp (Meta)
- URL: `https://tu-dominio.com/api/v1/integrations/webhooks/whatsapp`
- Verificación: Token personalizado
- Firma: X-Hub-Signature-256

### Messenger/Instagram (Meta)
- URL: `https://tu-dominio.com/api/v1/integrations/webhooks/messenger`
- URL: `https://tu-dominio.com/api/v1/integrations/webhooks/instagram`
- Verificación: hub.verify_token
- Firma: X-Hub-Signature-256

### Telegram
- URL: `https://tu-dominio.com/api/v1/integrations/webhooks/telegram`
- Configuración: `https://api.telegram.org/bot<token>/setWebhook`
- Sin firma (validación por token)

### Webchat
- URL: `https://tu-dominio.com/api/v1/integrations/webhooks/webchat`
- Configuración personalizada
- Firma opcional