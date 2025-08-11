# IT Integration Service

Servicio de **integración y configuración** para múltiples plataformas de redes sociales incluyendo WhatsApp, Telegram, Messenger, Instagram y Webchat.

## 🎯 **Propósito**

Este servicio se encarga **únicamente** de:
- **Configurar integraciones** con plataformas de redes sociales
- **Gestionar canales** de comunicación
- **Recibir webhooks** de mensajes entrantes
- **Validar tokens** y configuraciones
- **Reenviar mensajes** al servicio de mensajería

**NO maneja el envío de mensajes** - esa funcionalidad está en el servicio de mensajería separado.

## 🚀 Características

- **Configuración de Plataformas**: WhatsApp, Telegram, Messenger, Instagram, Webchat
- **Gestión de Canales**: CRUD completo para integraciones
- **Webhooks**: Recepción y procesamiento de mensajes entrantes
- **Validación**: Verificación de tokens y configuraciones
- **Setup Asistido**: Configuración automática para cada plataforma
- **Observabilidad**: Métricas, logs y health checks
- **Multi-tenant**: Soporte para múltiples clientes/empresas

## 📋 Endpoints Principales

### 🔧 Gestión de Canales
- `GET /api/v1/integrations/channels` - Listar canales
- `POST /api/v1/integrations/channels` - Crear canal
- `GET /api/v1/integrations/channels/:id` - Obtener canal
- `PATCH /api/v1/integrations/channels/:id` - Actualizar canal
- `DELETE /api/v1/integrations/channels/:id` - Eliminar canal

### 🔗 Setup de Plataformas
- `GET /api/v1/integrations/telegram/bot-info` - Info del bot
- `POST /api/v1/integrations/telegram/setup` - Configurar Telegram
- `GET /api/v1/integrations/whatsapp/business-info` - Info de WhatsApp
- `POST /api/v1/integrations/whatsapp/setup` - Configurar WhatsApp
- `GET /api/v1/integrations/messenger/page-info` - Info de Messenger
- `POST /api/v1/integrations/messenger/setup` - Configurar Messenger

### 📥 Webhooks (Recepción)
- `POST /api/v1/integrations/webhooks/whatsapp` - Webhook WhatsApp
- `POST /api/v1/integrations/webhooks/telegram` - Webhook Telegram
- `POST /api/v1/integrations/webhooks/messenger` - Webhook Messenger
- `POST /api/v1/integrations/webhooks/instagram` - Webhook Instagram
- `POST /api/v1/integrations/webhooks/webchat` - Webhook Webchat

### 📊 Validación
- `GET /api/v1/integrations/messages/inbound` - Validar mensajes entrantes
- `GET /api/v1/health` - Health check
- `GET /api/v1/ready` - Readiness check

## 🏗️ Arquitectura

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Redes Sociales │    │ Integration      │    │ Messaging       │
│   (WhatsApp,     │───▶│ Service          │───▶│ Service         │
│    Telegram,     │    │ (Este servicio)  │    │ (Servicio       │
│    etc.)         │    │                  │    │  separado)      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Flujo de Integración:
1. **Configuración** → Crear canales para cada plataforma
2. **Validación** → Verificar tokens y configuraciones
3. **Webhooks** → Recibir mensajes entrantes
4. **Procesamiento** → Normalizar y reenviar al servicio de mensajería

## 🛠️ Instalación

### 1. Clonar el repositorio
```bash
git clone <repository-url>
cd it-integration-service
```

### 2. Instalar dependencias
```bash
go mod download
```

### 3. Configurar variables de entorno
```bash
cp env.example .env.local
# Editar .env.local con tus configuraciones
```

### 4. Ejecutar
```bash
# Desarrollo
make dev-simple

# Docker
make dev
```

## 📚 Documentación de API

Una vez que el servicio esté ejecutándose:
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/api/v1/health

## 🔧 Configuración de Plataformas

### Telegram
```bash
# Obtener información del bot
curl "http://localhost:8080/api/v1/integrations/telegram/bot-info?bot_token=YOUR_TOKEN"

# Configurar integración
curl -X POST "http://localhost:8080/api/v1/integrations/telegram/setup" \
  -H "Content-Type: application/json" \
  -d '{
    "bot_token": "YOUR_TOKEN",
    "webhook_url": "https://your-domain.com/api/v1/integrations/webhooks/telegram",
    "tenant_id": "your_tenant_id"
  }'
```

### WhatsApp
```bash
# Obtener información del negocio
curl "http://localhost:8080/api/v1/integrations/whatsapp/business-info?access_token=YOUR_TOKEN"

# Configurar integración
curl -X POST "http://localhost:8080/api/v1/integrations/whatsapp/setup" \
  -H "Content-Type: application/json" \
  -d '{
    "access_token": "YOUR_TOKEN",
    "phone_number_id": "YOUR_PHONE_ID",
    "webhook_url": "https://your-domain.com/api/v1/integrations/webhooks/whatsapp",
    "tenant_id": "your_tenant_id"
  }'
```

## ⚠️ Notas Importantes

1. **Este servicio NO envía mensajes** - Solo configura integraciones
2. **Los mensajes se reenvían** al servicio de mensajería separado
3. **Los webhooks son solo para recepción** de mensajes entrantes
4. **La validación de tokens** es para verificar configuraciones
5. **Multi-tenant** - Cada cliente tiene su propio tenant_id

## 🔄 Integración con Servicio de Mensajería

Este servicio se integra con el servicio de mensajería a través de:
- **Webhooks** que reenvían mensajes normalizados
- **Configuración de canales** que el servicio de mensajería consulta
- **Validación de integraciones** antes de permitir envío

El servicio de mensajería es responsable de:
- Envío de mensajes individuales y masivos
- Gestión de colas y reintentos
- Templates y contenido de mensajes
- Métricas de envío