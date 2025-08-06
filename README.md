# IT Integration Service

Servicio de integración para múltiples plataformas de mensajería incluyendo WhatsApp, Telegram, Messenger, Instagram y Webchat.

## 🚀 Características

- **Múltiples Plataformas**: WhatsApp, Telegram, Messenger, Instagram, Webchat
- **Múltiples Proveedores**: Meta, Twilio, 360Dialog, Custom
- **Gestión de Canales**: CRUD completo para integraciones
- **Envío de Mensajes**: Individual y masivo (broadcast)
- **Webhooks**: Recepción y procesamiento de mensajes entrantes
- **Historial**: Consulta de mensajes entrantes, salientes y conversaciones
- **Configuración Automática**: Setup asistido para cada plataforma
- **Observabilidad**: Métricas, logs y health checks
- **Modo Mock**: Funciona sin base de datos para desarrollo

## � Requisitos

- Go 1.21+
- PostgreSQL 13+ (opcional, tiene modo mock)
- Docker (opcional)

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

### 3. Configurar el servicio
```bash
cp config.example.yaml config.yaml
# Edita config.yaml con tus configuraciones
```

### 4. Configurar variables de entorno
```bash
cp env.example .env
# Edita .env con tus tokens y configuraciones
```

### 5. Compilar y ejecutar
```bash
# Compilar
make build

# Ejecutar
make run

# O directamente
go run main.go
```

## 🐳 Docker

```bash
# Construir imagen
docker build -t it-integration-service .

# Ejecutar con docker-compose
docker-compose up -d
```

## 📚 Documentación de API

Una vez que el servicio esté ejecutándose, puedes acceder a:

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/api/v1/health
- **Métricas**: http://localhost:8080/metrics

## 🔧 Configuración de Plataformas

### Telegram

1. Crear un bot con @BotFather
2. Obtener el token del bot
3. Usar los endpoints de setup:

```bash
# Obtener información del bot
curl "http://localhost:8080/api/v1/integrations/telegram/bot-info?bot_token=YOUR_TOKEN"

# Configurar integración completa
curl -X POST "http://localhost:8080/api/v1/integrations/telegram/setup" \
  -H "Content-Type: application/json" \
  -d '{
    "bot_token": "YOUR_TOKEN",
    "webhook_url": "https://your-domain.com/api/v1/integrations/webhooks/telegram",
    "tenant_id": "your_tenant_id"
  }'
```

### WhatsApp

1. Configurar WhatsApp Business API
2. Obtener access token y phone number ID
3. Usar los endpoints de setup:

```bash
# Verificar número de teléfono
curl "http://localhost:8080/api/v1/integrations/whatsapp/phone-info?access_token=YOUR_TOKEN&phone_number_id=YOUR_PHONE_ID"

# Configurar integración completa
curl -X POST "http://localhost:8080/api/v1/integrations/whatsapp/setup" \
  -H "Content-Type: application/json" \
  -d '{
    "access_token": "YOUR_TOKEN",
    "phone_number_id": "YOUR_PHONE_ID",
    "business_account_id": "YOUR_BUSINESS_ID",
    "webhook_url": "https://your-domain.com/api/v1/integrations/webhooks/whatsapp",
    "tenant_id": "your_tenant_id"
  }'
```

## 📨 Envío de Mensajes

### Mensaje Simple
```bash
curl -X POST "http://localhost:8080/api/v1/integrations/send" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "your_channel_id",
    "recipient": "573001234567",
    "content": {
      "type": "text",
      "text": "¡Hola! Este es un mensaje de prueba."
    }
  }'
```

### Mensaje con Media
```bash
curl -X POST "http://localhost:8080/api/v1/integrations/send" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "your_channel_id",
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
  }'
```

### Mensaje Masivo (Broadcast)
```bash
curl -X POST "http://localhost:8080/api/v1/integrations/broadcast" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "your_tenant_id",
    "platforms": ["whatsapp", "telegram"],
    "recipients": ["573001234567", "573009876543"],
    "content": {
      "type": "text",
      "text": "📢 Mensaje masivo: ¡Hola a todos!"
    }
  }'
```

## 📊 Consulta de Mensajes

### Mensajes Entrantes
```bash
curl "http://localhost:8080/api/v1/integrations/messages/inbound?platform=whatsapp&limit=50"
```

### Mensajes Salientes
```bash
curl "http://localhost:8080/api/v1/integrations/messages/outbound?platform=telegram&limit=50"
```

### Historial de Conversación
```bash
curl "http://localhost:8080/api/v1/integrations/chat/whatsapp/user_123"
```

## 🔗 Webhooks

El servicio expone webhooks para recibir mensajes de todas las plataformas:

- **WhatsApp**: `POST /api/v1/integrations/webhooks/whatsapp`
- **Telegram**: `POST /api/v1/integrations/webhooks/telegram`
- **Messenger**: `POST /api/v1/integrations/webhooks/messenger`
- **Instagram**: `POST /api/v1/integrations/webhooks/instagram`
- **Webchat**: `POST /api/v1/integrations/webhooks/webchat`

### Configurar Webhooks

Para cada plataforma, configura la URL del webhook en su respectiva consola:

- **Telegram**: Usar el endpoint `/telegram/webhook` del servicio
- **WhatsApp**: Configurar en Meta Developer Console
- **Messenger**: Configurar en Facebook Developer Console
- **Instagram**: Configurar en Facebook Developer Console

## 🧪 Pruebas

### Ejecutar todas las pruebas
```bash
make test
```

### Probar endpoints manualmente
```bash
# Asegúrate de que el servicio esté corriendo
./scripts/test-endpoints.sh
```

### Con Postman
1. Importa los archivos de la carpeta `postman/`
2. Configura las variables de entorno
3. Ejecuta las colecciones

## 🔧 Desarrollo

### Estructura del Proyecto
```
├── cmd/                    # Comandos CLI
├── internal/
│   ├── config/            # Configuración
│   ├── domain/            # Entidades y repositorios
│   ├── handlers/          # Handlers HTTP
│   ├── middleware/        # Middlewares
│   ├── repository/        # Implementaciones de repositorios
│   ├── services/          # Lógica de negocio
│   └── usecase/           # Casos de uso
├── pkg/                   # Paquetes compartidos
├── scripts/               # Scripts de utilidad
├── postman/               # Colecciones de Postman
└── docs/                  # Documentación
```

### Agregar Nueva Plataforma

1. Agregar constante en `internal/domain/entities.go`
2. Implementar handler en `internal/handlers/`
3. Implementar servicio en `internal/services/`
4. Agregar rutas en `internal/handlers/handlers.go`
5. Actualizar tests y documentación

### Modo Mock

El servicio puede funcionar sin base de datos para desarrollo:

```yaml
# config.yaml
features:
  database_enabled: false
```

En este modo, todos los datos se simulan en memoria.

## 🚀 Despliegue

### Variables de Entorno Requeridas

```bash
# Básicas
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info

# Base de datos (opcional si database_enabled=false)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=it_integration_db

# Tokens de plataformas (según necesites)
TELEGRAM_BOT_TOKEN=your_telegram_token
WHATSAPP_ACCESS_TOKEN=your_whatsapp_token
WHATSAPP_PHONE_ID=your_phone_id
WHATSAPP_BUSINESS_ID=your_business_id
```

### Docker Compose

```yaml
version: '3.8'
services:
  it-integration-service:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - ENVIRONMENT=production
      - DB_HOST=postgres
    depends_on:
      - postgres
  
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: it_integration_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## 📈 Monitoreo

### Health Checks
- `GET /api/v1/health` - Estado general
- `GET /api/v1/ready` - Disponibilidad

### Métricas
- `GET /metrics` - Métricas de Prometheus

### Logs
Los logs se estructuran en JSON para facilitar el análisis:

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "message": "Message sent successfully",
  "platform": "whatsapp",
  "recipient": "573001234567",
  "channel_id": "channel_123"
}
```

## 🔒 Seguridad

- Validación de signatures en webhooks
- Encriptación de tokens sensibles
- Rate limiting configurable
- CORS configurable
- Autenticación básica para Swagger

## 🤝 Contribuir

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver `LICENSE` para más detalles.

## 📞 Soporte

- Documentación: `/docs`
- Issues: GitHub Issues
- API Docs: `/swagger/index.html`

---

¡Listo para integrar todas tus plataformas de mensajería! 🚀