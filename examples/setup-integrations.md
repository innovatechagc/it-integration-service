# 🚀 Configuración de Integraciones - Telegram y WhatsApp

## 📱 **Telegram Bot Setup**

### 1. Crear Bot en Telegram
1. Ve a [@BotFather](https://t.me/botfather) en Telegram
2. Envía `/newbot`
3. Sigue las instrucciones para crear tu bot
4. Guarda el **Token** que te proporciona

### 2. Configurar Webhook
```bash
# Reemplaza YOUR_BOT_TOKEN con el token de tu bot
curl -X POST "https://api.telegram.org/botYOUR_BOT_TOKEN/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "http://localhost:8080/api/v1/integrations/webhooks/telegram",
    "allowed_updates": ["message", "edited_message"]
  }'
```

### 3. Crear Integración en la API
```bash
curl -X POST "http://localhost:8080/api/v1/integrations/channels" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "default",
    "platform": "telegram",
    "name": "Mi Bot de Telegram",
    "status": "active",
    "config": {
      "bot_token": "YOUR_BOT_TOKEN",
      "webhook_url": "http://localhost:8080/api/v1/integrations/webhooks/telegram"
    }
  }'
```

## 📱 **WhatsApp Business API Setup**

### 1. Configurar Meta Developer Account
1. Ve a [Meta for Developers](https://developers.facebook.com/)
2. Crea una nueva app
3. Agrega el producto "WhatsApp Business API"
4. Configura el webhook

### 2. Configurar Webhook
```bash
# Verificar webhook (GET request)
curl "http://localhost:8080/api/v1/integrations/webhooks/whatsapp?hub.mode=subscribe&hub.challenge=CHALLENGE_ACCEPTED&hub.verify_token=YOUR_VERIFY_TOKEN"
```

### 3. Crear Integración en la API
```bash
curl -X POST "http://localhost:8080/api/v1/integrations/channels" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "default",
    "platform": "whatsapp",
    "name": "Mi WhatsApp Business",
    "status": "active",
    "config": {
      "access_token": "YOUR_ACCESS_TOKEN",
      "phone_number_id": "YOUR_PHONE_NUMBER_ID",
      "webhook_url": "http://localhost:8080/api/v1/integrations/webhooks/whatsapp",
      "verify_token": "YOUR_VERIFY_TOKEN"
    }
  }'
```

## 🔧 **Configuración Local**

### 1. Variables de Entorno
Crea un archivo `.env` en `it-integration-service/`:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=itapp

# Server
PORT=8080
ENVIRONMENT=development

# Telegram (reemplaza con tus valores)
TELEGRAM_BOT_TOKEN=your_bot_token_here

# WhatsApp (reemplaza con tus valores)
WHATSAPP_ACCESS_TOKEN=your_access_token_here
WHATSAPP_PHONE_NUMBER_ID=your_phone_number_id_here
WHATSAPP_VERIFY_TOKEN=your_verify_token_here
```

### 2. Ejecutar Migraciones
```bash
cd it-migration
go run main.go
```

### 3. Levantar Servicios
```bash
# Integration Service
cd it-integration-service
go run main.go

# Flutter App
cd it-app
flutter run
```

## 🧪 **Testing**

### 1. Verificar Integraciones
```bash
# Listar canales configurados
curl http://localhost:8080/api/v1/integrations/channels?tenant_id=default
```

### 2. Enviar Mensaje de Prueba
```bash
curl -X POST "http://localhost:8080/api/v1/integrations/send" \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "telegram",
    "recipient_id": "YOUR_USER_ID",
    "message": "¡Hola! Este es un mensaje de prueba.",
    "message_type": "text"
  }'
```

### 3. Verificar Mensajes Entrantes
```bash
# Obtener mensajes entrantes
curl http://localhost:8080/api/v1/integrations/messages/inbound

# Obtener historial de chat específico
curl http://localhost:8080/api/v1/integrations/chat/telegram/YOUR_USER_ID
```

## 📱 **Flutter App**

### 1. Verificar Configuración
El archivo `it-app/lib/config/runtime_config.dart` ya está configurado para apuntar a:
- `http://localhost:8080/api/v1/integrations`

### 2. Probar en la App
1. Abre la app Flutter
2. Ve a la pantalla de mensajería
3. Deberías ver las conversaciones si hay mensajes en la base de datos

## 🐛 **Troubleshooting**

### Problema: No aparecen mensajes en Flutter
**Solución:**
1. Verifica que el integration service esté corriendo
2. Verifica que haya mensajes en la base de datos
3. Verifica la configuración de URLs en Flutter

### Problema: Webhooks no funcionan
**Solución:**
1. Usa ngrok para exponer tu localhost: `ngrok http 8080`
2. Actualiza las URLs de webhook con la URL de ngrok
3. Verifica que los tokens sean correctos

### Problema: Error de CORS
**Solución:**
1. Verifica que el middleware CORS esté configurado
2. Asegúrate de que Flutter esté usando la URL correcta 