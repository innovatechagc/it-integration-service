# 🚀 Guía Completa para Postman - IT Integration Service

## 📁 Archivos para Importar

### 1. Colección Principal
- **Archivo**: `postman/IT-Integration-Service.postman_collection.json`
- **Contiene**: 50+ endpoints organizados en carpetas
- **Funcionalidades**: Health checks, gestión de canales, setup de plataformas, envío de mensajes, webhooks

### 2. Variables de Entorno
- **Archivo**: `postman/IT-Integration-Service.postman_environment.json`
- **Contiene**: Variables predefinidas para todas las configuraciones

## 🔧 Configuración Paso a Paso

### Paso 1: Importar en Postman
1. Abre Postman
2. Clic en "Import"
3. Arrastra ambos archivos JSON o selecciónalos
4. Confirma la importación

### Paso 2: Configurar Variables de Entorno
Después de importar, configura estas variables en el environment:

#### Variables Básicas (Obligatorias)
```
base_url = http://localhost:8080
tenant_id = tenant_demo_123
```

#### Variables de Telegram (Opcionales)
```
telegram_bot_token = TU_BOT_TOKEN_AQUI
telegram_chat_id = TU_CHAT_ID_AQUI
```

#### Variables de WhatsApp (Opcionales)
```
whatsapp_access_token = TU_ACCESS_TOKEN_AQUI
whatsapp_phone_id = TU_PHONE_ID_AQUI
whatsapp_business_id = TU_BUSINESS_ID_AQUI
whatsapp_test_recipient = 573001234567
```

## 🚀 Iniciar el Servicio

### Opción 1: Desarrollo Simple (Recomendado)
```bash
# Configuración inicial
make setup

# Ejecutar el servicio
make dev-simple
```

### Opción 2: Con Docker
```bash
make dev
```

### Verificar que funciona
```bash
# Verificar health check
curl http://localhost:8080/api/v1/health

# O usar Postman con el endpoint "Health Check"
```

## 📋 Flujo de Pruebas Recomendado

### 1. Verificar Servicio
- **Health Check**: Confirma que el servicio está corriendo
- **Readiness Check**: Confirma que está listo para recibir tráfico

### 2. Gestión de Canales
- **Get All Channels**: Ver integraciones existentes (inicialmente vacío)
- **Create Channel**: Crear una integración manual

### 3. Setup Automático de Plataformas

#### Para Telegram:
1. **Get Bot Info**: Verificar tu bot token
2. **Setup Telegram Integration**: Configuración automática completa
3. **Send Test Message**: Probar envío

#### Para WhatsApp:
1. **Get Phone Number Info**: Verificar tu configuración
2. **Setup WhatsApp Integration**: Configuración automática completa
3. **Send Test Message**: Probar envío

### 4. Envío de Mensajes
- **Send Single Message**: Mensaje individual
- **Send Media Message**: Mensaje con imagen/archivo
- **Broadcast Message**: Mensaje masivo a múltiples plataformas

### 5. Consulta de Historial
- **Get Inbound Messages**: Mensajes recibidos
- **Get Outbound Messages**: Mensajes enviados
- **Get Chat History**: Conversación completa con un usuario

### 6. Webhooks (Simulación)
- **WhatsApp Webhook**: Simular mensaje entrante de WhatsApp
- **Telegram Webhook**: Simular mensaje entrante de Telegram
- **Webchat Webhook**: Simular mensaje del chat web

## 🔑 Configuración de Tokens

### Telegram Bot Token
1. Habla con @BotFather en Telegram
2. Crea un nuevo bot: `/newbot`
3. Copia el token que te da
4. Pégalo en `telegram_bot_token`

### WhatsApp Tokens
1. Ve a [Meta for Developers](https://developers.facebook.com/)
2. Crea una app de WhatsApp Business
3. Obtén:
   - Access Token (temporal o permanente)
   - Phone Number ID
   - Business Account ID
4. Pégalos en las variables correspondientes

## 📊 Ejemplos de Uso

### Crear Canal Manualmente
```json
{
  "tenant_id": "tenant_demo_123",
  "platform": "whatsapp",
  "provider": "meta",
  "access_token": "tu_token_aqui",
  "webhook_url": "https://tu-dominio.com/api/v1/integrations/webhooks/whatsapp",
  "status": "active",
  "config": {
    "phone_number_id": "123456789",
    "business_account_id": "987654321"
  }
}
```

### Enviar Mensaje Simple
```json
{
  "channel_id": "tu_channel_id",
  "recipient": "573001234567",
  "content": {
    "type": "text",
    "text": "¡Hola! Este es un mensaje de prueba."
  }
}
```

### Mensaje Masivo
```json
{
  "tenant_id": "tenant_demo_123",
  "platforms": ["whatsapp", "telegram"],
  "recipients": ["573001234567", "573009876543"],
  "content": {
    "type": "text",
    "text": "📢 Mensaje masivo para todos!"
  }
}
```

## 🐛 Solución de Problemas

### Error: "Service not available"
- Verifica que el servicio esté corriendo: `make health`
- Reinicia el servicio: `make dev-simple`

### Error: "tenant_id is required"
- Configura la variable `tenant_id` en el environment de Postman

### Error: "Invalid bot token" (Telegram)
- Verifica que el token sea correcto con "Get Bot Info"
- Asegúrate de que el bot esté activo

### Error: "Channel not found"
- Primero crea una integración usando los endpoints de setup
- O crea un canal manualmente con "Create Channel"

### Modo Mock Activo
Si ves respuestas como "mock-channel-1", significa que:
- El servicio está funcionando en modo mock (sin base de datos)
- Esto es normal para desarrollo
- Los datos se simulan en memoria

## 🔄 Flujo Completo de Ejemplo

### 1. Configurar Telegram
```bash
# En Postman:
1. "Get Bot Info" con tu token
2. "Setup Telegram Integration" 
3. Guarda el channel_id devuelto
4. "Send Test Message"
```

### 2. Enviar Mensajes
```bash
# Usar el channel_id obtenido en el paso anterior
1. "Send Single Message"
2. "Send Media Message" 
3. "Broadcast Message"
```

### 3. Ver Historial
```bash
1. "Get Inbound Messages"
2. "Get Outbound Messages"
3. "Get Chat History"
```

## 📈 Monitoreo

### Endpoints de Salud
- `GET /api/v1/health` - Estado general
- `GET /api/v1/ready` - Disponibilidad

### Logs del Servicio
```bash
# Ver logs en tiempo real
make logs

# O si usas dev-simple
# Los logs aparecen en la consola
```

## 🎯 Casos de Uso Principales

### 1. Bot de Atención al Cliente
- Configurar WhatsApp y Telegram
- Recibir mensajes vía webhooks
- Responder automáticamente
- Mantener historial de conversaciones

### 2. Notificaciones Masivas
- Configurar múltiples plataformas
- Usar broadcast para enviar a todos
- Monitorear entregas exitosas/fallidas

### 3. Integración con CRM
- Recibir webhooks de mensajes
- Procesar y reenviar a sistema principal
- Mantener logs de todas las interacciones

## 🔐 Seguridad

### Variables Sensibles
- Nunca compartas tus tokens reales
- Usa variables de entorno en producción
- Los tokens en Postman son solo para desarrollo

### Webhooks
- En producción, configura URLs HTTPS
- Valida signatures de webhooks
- Usa tokens de verificación únicos

---

## ✅ Checklist de Configuración

- [ ] Servicio corriendo (`make dev-simple`)
- [ ] Colección importada en Postman
- [ ] Environment configurado
- [ ] Variables básicas configuradas (`base_url`, `tenant_id`)
- [ ] Health check exitoso
- [ ] Al menos una plataforma configurada (Telegram o WhatsApp)
- [ ] Mensaje de prueba enviado exitosamente

¡Listo para integrar todas tus plataformas de mensajería! 🚀