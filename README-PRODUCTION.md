# 🚀 Guía de Configuración para Producción

## 📋 **Resumen del Estado Actual**

### ✅ **Lo que ya está funcionando:**
- ✅ Integration Service corriendo en puerto 8080
- ✅ Base de datos con integraciones configuradas (Telegram y WhatsApp)
- ✅ API endpoints funcionando correctamente
- ✅ Flutter app configurada para conectarse al backend
- ✅ Webhooks configurados con URLs de ngrok

### 🔧 **Lo que necesitas hacer:**

## **PASO 1: Configurar ngrok para Webhooks**

### 1.1 Instalar ngrok
```bash
# Descargar ngrok
wget https://bin.equinox.io/c/bNyj1mQVY4c/ngrok-v3-stable-linux-amd64.tgz
tar xvzf ngrok-v3-stable-linux-amd64.tgz

# O usar snap
sudo snap install ngrok
```

### 1.2 Exponer el puerto 8080
```bash
ngrok http 8080
```

### 1.3 Obtener la URL de ngrok
```bash
curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url'
```

## **PASO 2: Configurar Webhooks con URLs Reales**

### 2.1 Actualizar webhook de Telegram
```bash
# Reemplaza YOUR_NGROK_URL con tu URL de ngrok
curl -X POST "https://api.telegram.org/bot8253543805:AAFjcX5L_LRTgl_MP9j0k_D4056ar2XtZw4/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://YOUR_NGROK_URL/api/v1/integrations/webhooks/telegram",
    "allowed_updates": ["message", "edited_message"]
  }'
```

### 2.2 Actualizar webhook de WhatsApp
```bash
# En Meta Developer Console, actualiza la URL del webhook con:
https://YOUR_NGROK_URL/api/v1/integrations/webhooks/whatsapp
```

### 2.3 Actualizar base de datos
```sql
-- Actualizar URLs de webhook en la base de datos
UPDATE channel_integrations 
SET webhook_url = 'https://YOUR_NGROK_URL/api/v1/integrations/webhooks/telegram'
WHERE platform = 'telegram' AND tenant_id = 'tenant1';

UPDATE channel_integrations 
SET webhook_url = 'https://YOUR_NGROK_URL/api/v1/integrations/webhooks/whatsapp'
WHERE platform = 'whatsapp' AND tenant_id = 'tenant1';
```

## **PASO 3: Probar Integraciones**

### 3.1 Probar Telegram
1. Busca tu bot en Telegram: `@it_app_chat_bot`
2. Envía un mensaje al bot
3. Verifica que el webhook reciba el mensaje:
```bash
curl -s "http://localhost:8080/api/v1/integrations/messages/inbound?tenant_id=tenant1" | jq '.'
```

### 3.2 Probar WhatsApp
1. Envía un mensaje al número de WhatsApp configurado
2. Verifica que el webhook reciba el mensaje:
```bash
curl -s "http://localhost:8080/api/v1/integrations/messages/inbound?tenant_id=tenant1" | jq '.'
```

### 3.3 Enviar mensajes de respuesta
```bash
# Enviar mensaje a Telegram (reemplaza USER_ID con el ID real)
curl -X POST "http://localhost:8080/api/v1/integrations/send" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "23d8d953-a571-45df-95de-f5aecb5b0b93",
    "recipient": "USER_ID",
    "content": {
      "type": "text",
      "text": "¡Hola! Gracias por tu mensaje."
    }
  }'

# Enviar mensaje a WhatsApp (reemplaza PHONE_NUMBER con el número real)
curl -X POST "http://localhost:8080/api/v1/integrations/send" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "42ef8faa-571c-4fe8-9fbe-7531ad05a72d",
    "recipient": "PHONE_NUMBER",
    "content": {
      "type": "text",
      "text": "¡Hola! Gracias por tu mensaje."
    }
  }'
```

## **PASO 4: Probar Flutter App**

### 4.1 Ejecutar Flutter
```bash
cd it-app
flutter run
```

### 4.2 Verificar conexión
1. Abre la app Flutter
2. Ve a la pantalla de mensajería
3. Deberías ver las conversaciones si hay mensajes en la base de datos

### 4.3 Debugging
Si no aparecen mensajes:
```bash
# Verificar que la API responda
curl -s "http://localhost:8080/api/v1/health"

# Verificar canales
curl -s "http://localhost:8080/api/v1/integrations/channels?tenant_id=tenant1"

# Verificar mensajes
curl -s "http://localhost:8080/api/v1/integrations/messages/inbound?tenant_id=tenant1"
```

## **PASO 5: Configuración para Producción**

### 5.1 Variables de Entorno
Crea un archivo `.env` en `it-integration-service/`:
```env
# Database
DB_HOST=your-production-db-host
DB_PORT=5432
DB_USER=your-db-user
DB_PASSWORD=your-secure-password
DB_NAME=itapp

# Server
PORT=8080
ENVIRONMENT=production

# Telegram
TELEGRAM_BOT_TOKEN=8253543805:AAFjcX5L_LRTgl_MP9j0k_D4056ar2XtZw4

# WhatsApp
WHATSAPP_ACCESS_TOKEN=EAAtY7ZAsoerwBPL5tNn0Kmuq0j4jXMS88My30y6BjMP0Df4Rqoz3ZBcgJFQOqvx0Oa0FiUoZBFF0ALaah5jCZCh2ej2WgnWHQzXWONG5QwnIiVJnV6ljmlnU3pPXlZAiMZAcrjuZBBajfviNAYobb7CxctomZCBZBLVmR5cPrmDouhzpfoCqsI1J1b3l1XvDnLUkTUw2PpqUKEz17ZAUmNYiynM25pCsedJVOdFWbhNSJXrzoWCA7J4Lm6CoGBl4nvvgZDZD
WHATSAPP_PHONE_NUMBER_ID=764957900026580
WHATSAPP_VERIFY_TOKEN=itapp_whatsapp_verify_2024
```

### 5.2 Deployment
```bash
# Construir para producción
go build -o bin/integration-service .

# Ejecutar en producción
./bin/integration-service
```

### 5.3 URLs de Producción
Actualiza las URLs en `it-app/lib/config/runtime_config.dart`:
```dart
static const String _prodBaseUrl = 'https://your-production-domain.com/api/v1';
static const String _prodAuthUrl = 'https://auth.your-domain.com';
static const String _prodUserUrl = 'https://user.your-domain.com/api/v1';
```

## **🧪 Testing Completo**

### Test 1: Verificar API
```bash
# Health check
curl http://localhost:8080/api/v1/health

# Canales
curl http://localhost:8080/api/v1/integrations/channels?tenant_id=tenant1

# Mensajes
curl http://localhost:8080/api/v1/integrations/messages/inbound?tenant_id=tenant1
```

### Test 2: Enviar mensaje de prueba
```bash
# Telegram
curl -X POST "http://localhost:8080/api/v1/integrations/send" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "23d8d953-a571-45df-95de-f5aecb5b0b93",
    "recipient": "YOUR_TELEGRAM_USER_ID",
    "content": {
      "type": "text",
      "text": "Mensaje de prueba desde la API"
    }
  }'
```

### Test 3: Verificar Flutter
1. Ejecutar `flutter run`
2. Navegar a la pantalla de mensajería
3. Verificar que aparezcan las conversaciones

## **🐛 Troubleshooting**

### Problema: Webhooks no funcionan
**Solución:**
1. Verificar que ngrok esté corriendo
2. Verificar que la URL de ngrok sea HTTPS
3. Verificar que el puerto 8080 esté expuesto

### Problema: No aparecen mensajes en Flutter
**Solución:**
1. Verificar que haya mensajes en la base de datos
2. Verificar que el tenant_id sea correcto
3. Verificar la configuración de URLs en Flutter

### Problema: Error de CORS
**Solución:**
1. Verificar que el middleware CORS esté configurado
2. Verificar que Flutter esté usando la URL correcta

## **📞 Soporte**

Si tienes problemas:
1. Revisa los logs del integration service
2. Verifica que todos los servicios estén corriendo
3. Prueba los endpoints individualmente
4. Verifica la configuración de la base de datos 