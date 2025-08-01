# 🚀 Desarrollo Local - Integration Service

Guía completa para configurar y probar el servicio de integraciones localmente.

## 🏗️ Configuración Inicial

### 1. Prerrequisitos

```bash
# Instalar dependencias del sistema
- Docker & Docker Compose
- Go 1.21+
- Make
- curl & jq (para testing)
```

### 2. Setup del Proyecto

```bash
# Clonar y configurar
git clone <repository>
cd integration-service

# Configuración inicial automática
make setup

# O manualmente:
cp .env.example .env.local
go mod download
```

### 3. Configurar Variables de Entorno

Edita `.env.local` con tus configuraciones:

```bash
# Configuración básica
ENVIRONMENT=development
PORT=8080
LOG_LEVEL=debug

# Tokens de prueba para integraciones
META_APP_ID=your_test_app_id
META_APP_SECRET=your_test_app_secret
TWILIO_ACCOUNT_SID=your_test_account_sid
TELEGRAM_BOT_TOKEN=your_test_bot_token
```

## 🔧 Comandos de Desarrollo

### Entorno de Desarrollo

```bash
# Iniciar todos los servicios (app + DB + herramientas)
make dev

# Solo para testing con simuladores
make dev-test

# Ver logs en tiempo real
make dev-logs

# Detener servicios
make dev-down
```

### Testing

```bash
# Tests unitarios
make test

# Tests con coverage
make test-coverage

# Tests de integración completos
make test-integration

# Test específico de una plataforma
./scripts/test-integrations.sh whatsapp
./scripts/test-integrations.sh telegram
./scripts/test-integrations.sh all
```

### Herramientas de Desarrollo

```bash
# Formatear código
make fmt

# Linter
make lint

# Verificar health check
make health

# Ver estado de servicios
make status
```

## 🔗 URLs Importantes

Una vez iniciado el entorno de desarrollo:

- **API Principal**: http://localhost:8080
- **Health Check**: http://localhost:8080/api/v1/health
- **Webhook Simulator**: http://localhost:8081
- **Prometheus**: http://localhost:9090
- **Vault**: http://localhost:8200 (token: `dev-token`)
- **PostgreSQL**: localhost:5432 (user: `postgres`, pass: `postgres`)

## 🧪 Testing de Integraciones

### 1. Simulador Web de Webhooks

Abre http://localhost:8081 para acceder al simulador interactivo que incluye:

- ✅ WhatsApp (Meta) - Mensajes de texto, imágenes, estados
- ✅ Telegram - Mensajes y comandos
- ✅ Messenger - Mensajes de Facebook
- ✅ Instagram - Mensajes directos
- ✅ Herramientas de testing (health check, listar integraciones)

### 2. Testing por Línea de Comandos

```bash
# Test completo de todas las integraciones
./scripts/test-integrations.sh all

# Test específico por plataforma
./scripts/test-integrations.sh whatsapp
./scripts/test-integrations.sh telegram

# Test de funcionalidades específicas
./scripts/test-integrations.sh health
./scripts/test-integrations.sh crud
./scripts/test-integrations.sh send
```

### 3. Testing Manual con curl

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Listar integraciones
curl http://localhost:8080/api/v1/integrations

# Webhook de WhatsApp
curl -X POST http://localhost:8080/webhooks/whatsapp/meta \
  -H "Content-Type: application/json" \
  -d '{
    "object": "whatsapp_business_account",
    "entry": [{
      "changes": [{
        "value": {
          "messages": [{
            "from": "1234567890",
            "text": {"body": "Hello World"},
            "type": "text"
          }]
        }
      }]
    }]
  }'
```

## 🗄️ Base de Datos

### Estructura de Datos de Prueba

El entorno incluye datos de prueba automáticos:

```sql
-- Usuarios de prueba
test@example.com (admin)
dev@example.com (developer)

-- Integraciones de prueba
- WhatsApp Meta (tenant-test-1)
- Telegram (tenant-test-1)  
- WhatsApp Twilio (tenant-test-2)
```

### Comandos de Base de Datos

```bash
# Resetear base de datos
make db-reset

# Conectar a PostgreSQL
docker exec -it integration-service_postgres_1 psql -U postgres -d microservice_dev

# Ver logs de base de datos
docker-compose logs postgres
```

## 🌐 Exposición de Webhooks (Ngrok)

Para probar webhooks reales desde plataformas externas:

```bash
# Configurar token de ngrok
export NGROK_AUTHTOKEN=your_ngrok_token

# Exponer webhooks públicamente
make ngrok

# Ver dashboard de ngrok
open http://localhost:4040
```

Las URLs públicas estarán disponibles en formato:
- `https://abc123.ngrok.io/webhooks/whatsapp/meta`
- `https://abc123.ngrok.io/webhooks/telegram`

## 📊 Monitoreo y Debugging

### Logs

```bash
# Logs de la aplicación
make logs

# Logs específicos de un servicio
docker-compose logs -f postgres
docker-compose logs -f vault
```

### Métricas

```bash
# Abrir Prometheus
make metrics

# Ver métricas directamente
curl http://localhost:8080/metrics
```

### Debugging

```bash
# Verificar conectividad de servicios
make health

# Estado de todos los contenedores
make status

# Información del proyecto
make info
```

## 🔧 Configuración de Integraciones

### WhatsApp (Meta)

1. Crear app en Meta for Developers
2. Configurar webhook URL: `https://your-ngrok-url.ngrok.io/webhooks/whatsapp/meta`
3. Agregar verify token en `.env.local`
4. Suscribirse a eventos: `messages`, `message_deliveries`

### Telegram

1. Crear bot con @BotFather
2. Configurar webhook: `https://api.telegram.org/bot<TOKEN>/setWebhook?url=https://your-ngrok-url.ngrok.io/webhooks/telegram`
3. Agregar bot token en `.env.local`

### Twilio

1. Configurar cuenta de Twilio
2. Configurar webhook URL en la consola de Twilio
3. Agregar credenciales en `.env.local`

## 🚨 Troubleshooting

### Problemas Comunes

**Puerto 8080 ocupado:**
```bash
# Cambiar puerto en .env.local
PORT=8081

# O matar proceso que usa el puerto
lsof -ti:8080 | xargs kill -9
```

**Base de datos no conecta:**
```bash
# Resetear contenedores
make clean-all
make dev
```

**Webhooks no llegan:**
```bash
# Verificar ngrok
curl http://localhost:4040/api/tunnels

# Verificar logs
make logs
```

### Logs de Debug

```bash
# Habilitar logs detallados
export LOG_LEVEL=debug

# Ver logs en tiempo real con filtros
docker-compose logs -f app | grep "webhook"
docker-compose logs -f app | grep "ERROR"
```

## 📝 Desarrollo de Nuevas Integraciones

### 1. Agregar Nueva Plataforma

```go
// En internal/domain/entities.go
const (
    PlatformNewPlatform Platform = "newplatform"
)

// En internal/services/
// Implementar handlers específicos
```

### 2. Testing de Nueva Integración

```bash
# Agregar casos de prueba en scripts/test-integrations.sh
test_newplatform_webhooks() {
    # Implementar tests específicos
}
```

### 3. Configuración

```bash
# Agregar variables en .env.local
NEWPLATFORM_API_KEY=your_api_key
NEWPLATFORM_SECRET=your_secret
```

## 🎯 Flujo de Trabajo Recomendado

1. **Iniciar entorno**: `make dev`
2. **Abrir simulador**: http://localhost:8081
3. **Desarrollar feature**
4. **Probar con simulador web**
5. **Ejecutar tests**: `./scripts/test-integrations.sh all`
6. **Probar con ngrok** (si necesario)
7. **Commit y push**

## 📚 Recursos Adicionales

- [Documentación de WhatsApp Business API](https://developers.facebook.com/docs/whatsapp)
- [Documentación de Telegram Bot API](https://core.telegram.org/bots/api)
- [Documentación de Twilio](https://www.twilio.com/docs)
- [Docker Compose Reference](https://docs.docker.com/compose/)

---

¿Necesitas ayuda? Revisa los logs con `make logs` o ejecuta `make info` para ver el estado del sistema.