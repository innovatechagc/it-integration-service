# Configuración de Telegram Bot

Este documento explica cómo configurar un bot de Telegram usando la API del servicio, sin necesidad de scripts externos.

## Pasos para configurar un bot de Telegram

### 1. Crear el bot en Telegram

1. Ve a [@BotFather](https://t.me/BotFather) en Telegram
2. Envía `/newbot`
3. Sigue las instrucciones para crear tu bot
4. Copia el token que te proporciona (formato: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

### 2. Validar el token (opcional)

Antes de configurar completamente el bot, puedes validar que el token sea correcto:

```bash
curl -X POST http://localhost:8080/api/v1/integrations/telegram/validate \
  -H "Content-Type: application/json" \
  -d '{
    "bot_token": "TU_BOT_TOKEN_AQUI"
  }'
```

Respuesta esperada:
```json
{
  "code": "SUCCESS",
  "message": "Bot token is valid",
  "data": {
    "id": 123456789,
    "is_bot": true,
    "first_name": "Mi Bot",
    "username": "mi_bot",
    "can_join_groups": true,
    "can_read_all_group_messages": false,
    "supports_inline_queries": false
  }
}
```

### 3. Configurar el bot completo

Tienes dos opciones para configurar el bot:

#### Opción A: Configuración manual (recomendada para desarrollo)

```bash
curl -X POST http://localhost:8080/api/v1/integrations/telegram/setup \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "default",
    "bot_token": "TU_BOT_TOKEN_AQUI",
    "webhook_url": "https://tu-dominio.com/api/v1/integrations/webhooks/telegram",
    "description": "Bot de atención al cliente"
  }'
```

**Nota sobre webhook_url:**
- En desarrollo local, puedes usar ngrok: `https://abc123.ngrok.io/api/v1/integrations/webhooks/telegram`
- En producción, usa tu dominio real: `https://tu-dominio.com/api/v1/integrations/webhooks/telegram`

#### Opción B: Configuración desde variables de entorno (recomendada para producción)

Si ya tienes configuradas las variables de entorno `TELEGRAM_BOT_TOKEN` y `TELEGRAM_DEFAULT_WEBHOOK_URL` (o `BASE_URL`), puedes usar:

```bash
curl -X POST http://localhost:8080/api/v1/integrations/telegram/setup-from-config \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "default",
    "description": "Bot de atención al cliente"
  }'
```

Esta opción es más segura porque no expones el token en la petición HTTP.

Respuesta esperada:
```json
{
  "code": "SUCCESS",
  "message": "Telegram bot configured successfully",
  "data": {
    "channel_id": "uuid-generado",
    "bot_info": {
      "id": 123456789,
      "is_bot": true,
      "first_name": "Mi Bot",
      "username": "mi_bot",
      "can_join_groups": true,
      "can_read_all_group_messages": false,
      "supports_inline_queries": false
    },
    "webhook_info": {
      "url": "https://tu-dominio.com/api/v1/integrations/webhooks/telegram",
      "has_custom_certificate": false,
      "pending_update_count": 0,
      "max_connections": 40,
      "allowed_updates": ["message", "callback_query"]
    },
    "status": "success",
    "message": "Telegram bot @mi_bot configured successfully"
  }
}
```

### 4. Verificar la integración

Puedes verificar que la integración se creó correctamente:

```bash
curl "http://localhost:8080/api/v1/integrations/channels?tenant_id=default"
```

### 5. Probar el bot

1. Busca tu bot en Telegram por su username (ej: @mi_bot)
2. Envía `/start` o cualquier mensaje
3. Verifica en los logs del servicio que llegan los webhooks
4. Revisa los mensajes entrantes en la API:

```bash
curl "http://localhost:8080/api/v1/integrations/messages/inbound?platform=telegram"
```

## Enviar mensajes

Una vez configurado, puedes enviar mensajes usando el channel_id obtenido:

```bash
curl -X POST http://localhost:8080/api/v1/integrations/send \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "uuid-del-canal",
    "recipient": "123456789",
    "content": {
      "type": "text",
      "text": "¡Hola! Este es un mensaje desde la API."
    }
  }'
```

## Gestión del webhook

### Eliminar webhook

Si necesitas eliminar el webhook (por ejemplo, para cambiar la URL):

```bash
curl -X POST http://localhost:8080/api/v1/integrations/telegram/webhook/remove \
  -H "Content-Type: application/json" \
  -d '{
    "bot_token": "TU_BOT_TOKEN_AQUI"
  }'
```

### Reconfigurar webhook

Para cambiar la URL del webhook, simplemente ejecuta el setup nuevamente con la nueva URL.

## Variables de entorno

Asegúrate de tener estas variables en tu `.env.local`:

```env
# Telegram (opcional, para secrets adicionales)
TELEGRAM_WEBHOOK_SECRET=tu-secret-opcional

# Base de datos (requerida)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=tu-password
DB_NAME=integration_service
DB_SSL_MODE=disable

# Servicio de mensajería (opcional)
MESSAGING_SERVICE_URL=http://localhost:8081
```

## Troubleshooting

### Error: "invalid bot token"
- Verifica que el token sea correcto
- Asegúrate de que el bot no haya sido eliminado en BotFather

### Error: "failed to configure webhook"
- Verifica que la URL del webhook sea accesible desde internet
- En desarrollo local, usa ngrok para exponer tu puerto
- Asegúrate de que la URL termine en `/api/v1/integrations/webhooks/telegram`

### Error: "failed to create integration"
- Verifica la conexión a la base de datos
- Revisa que el tenant_id sea válido
- Comprueba los logs del servicio para más detalles

### No llegan mensajes
- Verifica que el webhook esté configurado correctamente
- Revisa los logs del servicio
- Comprueba que el bot no esté en modo polling (solo webhook)
- Asegúrate de que la URL del webhook sea HTTPS en producción