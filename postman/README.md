# IT Integration Service - Guía de Postman

Esta guía te ayudará a configurar y usar la colección de Postman para probar todas las funcionalidades del servicio de integración de mensajería.

## 📁 Archivos Incluidos

- `IT-Integration-Service.postman_collection.json` - Colección completa con todos los endpoints
- `IT-Integration-Service.postman_environment.json` - Variables de entorno
- `README.md` - Esta guía

## 🚀 Configuración Inicial

### 1. Importar en Postman

1. Abre Postman
2. Haz clic en "Import" 
3. Arrastra los archivos JSON o selecciónalos:
   - `IT-Integration-Service.postman_collection.json`
   - `IT-Integration-Service.postman_environment.json`

### 2. Configurar Variables de Entorno

Después de importar, configura las siguientes variables en el environment:

#### Variables Básicas
- `base_url`: URL del servicio (default: `http://localhost:8080`)
- `tenant_id`: ID de tu tenant (ej: `tenant_demo_123`)

#### Variables de Telegram
- `telegram_bot_token`: Token de tu bot de Telegram
- `telegram_chat_id`: ID del chat para pruebas

#### Variables de WhatsApp
- `whatsapp_access_token`: Token de acceso de Meta
- `whatsapp_phone_id`: ID del número de teléfono
- `whatsapp_business_id`: ID de la cuenta de negocio
- `whatsapp_test_recipient`: Número de teléfono para pruebas (formato: 573001234567)

## 📋 Estructura de la Colección

### 1. Health & Status
- **Health Check**: Verifica que el servicio esté funcionando
- **Readiness Check**: Verifica que el servicio esté listo para recibir tráfico

### 2. Channel Management
- **Get All Channels**: Lista todas las integraciones activas
- **Get Channel by ID**: Obtiene detalles de una integración específica
- **Create Channel**: Registra una nueva integración
- **Update Channel**: Actualiza una integración existente
- **Delete Channel**: Elimina/desactiva una integración

### 3. Telegram Setup
- **Get Bot Info**: Obtiene información del bot
- **Setup Telegram Integration**: Configura la integración completa
- **Get Webhook Info**: Verifica el estado del webhook
- **Set Webhook**: Configura el webhook
- **Delete Webhook**: Elimina el webhook
- **Send Test Message**: Envía un mensaje de prueba

### 4. WhatsApp Setup
- **Get Business Info**: Obtiene información de la cuenta de negocio
- **Get Phone Number Info**: Verifica el número de teléfono
- **Setup WhatsApp Integration**: Configura la integración completa
- **Send Test Message**: Envía un mensaje de prueba
- **Validate Webhook**: Valida el webhook (usado por Meta)

### 5. Message Operations
- **Send Single Message**: Envía un mensaje individual
- **Send Media Message**: Envía un mensaje con multimedia
- **Broadcast Message**: Envía mensajes masivos a múltiples plataformas

### 6. Message History
- **Get Inbound Messages**: Obtiene mensajes recibidos
- **Get Outbound Messages**: Obtiene mensajes enviados
- **Get Chat History**: Obtiene conversación con un usuario específico

### 7. Webhooks
- Endpoints para recibir webhooks de todas las plataformas
- Incluye verificaciones y ejemplos de payloads

## 🔧 Flujo de Configuración Recomendado

### Para Telegram:

1. **Obtener información del bot**
   ```
   GET /api/v1/integrations/telegram/bot-info?bot_token={{telegram_bot_token}}
   ```

2. **Configurar integración completa**
   ```
   POST /api/v1/integrations/telegram/setup
   ```

3. **Enviar mensaje de prueba**
   ```
   POST /api/v1/integrations/telegram/test-message
   ```

### Para WhatsApp:

1. **Verificar información del negocio**
   ```
   GET /api/v1/integrations/whatsapp/business-info
   ```

2. **Verificar número de teléfono**
   ```
   GET /api/v1/integrations/whatsapp/phone-info
   ```

3. **Configurar integración completa**
   ```
   POST /api/v1/integrations/whatsapp/setup
   ```

4. **Enviar mensaje de prueba**
   ```
   POST /api/v1/integrations/whatsapp/test-message
   ```

## 📨 Ejemplos de Uso

### Envío de Mensaje Simple
```json
{
  "channel_id": "channel_abc123",
  "recipient": "573001234567",
  "content": {
    "type": "text",
    "text": "¡Hola! Este es un mensaje de prueba."
  }
}
```

### Envío de Mensaje con Media
```json
{
  "channel_id": "channel_abc123",
  "recipient": "573001234567",
  "content": {
    "type": "media",
    "text": "Aquí tienes una imagen:",
    "media": {
      "url": "https://example.com/image.jpg",
      "caption": "Imagen de ejemplo",
      "mime_type": "image/jpeg"
    }
  }
}
```

### Mensaje Masivo (Broadcast)
```json
{
  "tenant_id": "tenant_demo_123",
  "platforms": ["whatsapp", "telegram"],
  "recipients": [
    "573001234567",
    "573009876543",
    "987654321"
  ],
  "content": {
    "type": "text",
    "text": "📢 Mensaje masivo: ¡Hola a todos!"
  }
}
```

## 🔐 Autenticación y Seguridad

- Los tokens de acceso se manejan como variables de entorno
- Los webhooks incluyen validación de signatures
- Todos los endpoints requieren tenant_id para aislamiento de datos

## 🐛 Troubleshooting

### Errores Comunes:

1. **"tenant_id is required"**
   - Asegúrate de configurar la variable `tenant_id` en el environment

2. **"Invalid bot token"** (Telegram)
   - Verifica que el `telegram_bot_token` sea correcto
   - Usa el endpoint "Get Bot Info" para validar

3. **"Failed to verify phone number"** (WhatsApp)
   - Verifica que el `whatsapp_access_token` y `whatsapp_phone_id` sean correctos
   - Usa el endpoint "Get Phone Number Info" para validar

4. **"Channel not found"**
   - Primero crea una integración usando los endpoints de setup
   - Guarda el `channel_id` devuelto en las variables de entorno

## 📊 Monitoreo

Usa los endpoints de Health para monitorear el servicio:

- `/api/v1/health` - Estado general del servicio
- `/api/v1/ready` - Disponibilidad para recibir tráfico

## 🔄 Webhooks de Prueba

Para probar los webhooks localmente, puedes usar herramientas como:
- ngrok para exponer tu localhost
- Postman Mock Server
- Webhook.site para debugging

## 📞 Soporte

Si encuentras problemas:
1. Verifica que el servicio esté corriendo (`Health Check`)
2. Revisa las variables de entorno
3. Consulta los logs del servicio
4. Usa los endpoints de "test-message" para validar configuraciones

---

¡Listo para probar todas las integraciones! 🚀