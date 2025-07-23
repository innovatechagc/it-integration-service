# Template de Microservicio Go

Template estandarizado para crear microservicios en Go que se despliegan en GCP Cloud Run. Incluye configuración para desarrollo local, testing, QA y producción.

## 🚀 Características

- **Framework**: Gin para HTTP server
- **Logging**: Zap logger estructurado
- **Métricas**: Prometheus integrado
- **Secretos**: Integración con HashiCorp Vault
- **Documentación**: Swagger/OpenAPI
- **Testing**: Tests unitarios y de integración
- **Docker**: Multi-stage builds optimizados
- **CI/CD**: Configuración para diferentes entornos

## 📁 Estructura del Proyecto

```
├── cmd/                    # Comandos de la aplicación
├── internal/              # Código interno de la aplicación
│   ├── config/           # Configuración
│   ├── handlers/         # Handlers HTTP
│   ├── middleware/       # Middleware personalizado
│   └── services/         # Lógica de negocio
├── pkg/                  # Paquetes reutilizables
│   ├── logger/          # Logger personalizado
│   └── vault/           # Cliente de Vault
├── scripts/             # Scripts de inicialización
├── monitoring/          # Configuración de monitoreo
├── .env.*              # Archivos de configuración por entorno
├── docker-compose.yml  # Desarrollo local
├── Dockerfile         # Imagen de producción
└── Makefile          # Comandos de automatización
```

## 🛠️ Configuración Inicial

### 1. Clonar y configurar el proyecto

```bash
# Clonar el template
git clone <repository-url>
cd microservice-template

# Copiar configuración de ejemplo
cp .env.example .env.local

# Instalar dependencias
make deps
```

### 2. Configurar variables de entorno

Edita `.env.local` con tus configuraciones:

```bash
# Configuración básica
ENVIRONMENT=development
PORT=8080
LOG_LEVEL=debug

# Base de datos
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=microservice_dev

# Vault (comentado para desarrollo inicial)
# VAULT_ADDR=http://localhost:8200
# VAULT_TOKEN=dev-token
```

## 🚀 Desarrollo Local

### Opción 1: Ejecutar directamente

```bash
# Compilar y ejecutar
make build
make run

# O directamente
go run .
```

### Opción 2: Con Docker Compose (Recomendado)

```bash
# Levantar todos los servicios (app, postgres, vault, redis, prometheus)
make docker-dev

# Detener servicios
make docker-down
```

Servicios disponibles:
- **API**: http://localhost:8080
- **Swagger**: http://localhost:8080/swagger/index.html
- **Prometheus**: http://localhost:9090
- **Vault**: http://localhost:8200

## 🧪 Testing

```bash
# Ejecutar tests
make test

# Tests con cobertura
make test-coverage

# Tests con Docker
make docker-test

# Linting
make lint
```

## 📊 Endpoints Disponibles

### Health Checks
- `GET /api/v1/health` - Estado del servicio
- `GET /api/v1/ready` - Readiness check

### Métricas
- `GET /metrics` - Métricas de Prometheus

### Documentación
- `GET /swagger/index.html` - Documentación Swagger

## 🔧 Configuración por Entornos

### Desarrollo Local
- Archivo: `.env.local`
- Base de datos: PostgreSQL local
- Vault: Opcional (comentado por defecto)
- Logs: Debug level

### Testing/QA
- Archivo: `.env.test`
- Base de datos: PostgreSQL de testing
- Vault: Instancia de testing
- Logs: Info level

### Producción
- Archivo: `.env.production`
- Variables desde GCP Secret Manager o Vault
- SSL requerido para BD
- Logs: Warn level

## 🐳 Docker

### Desarrollo
```bash
# Construir imagen
make docker-build

# Ejecutar contenedor
make docker-run
```

### Testing
```bash
# Ejecutar tests en contenedor
make docker-test
```

## ☁️ Despliegue en GCP Cloud Run

### Preparación
1. Configurar gcloud CLI
2. Habilitar Cloud Run API
3. Configurar Container Registry

### Deploy a Staging
```bash
# Build y push de imagen
docker build -t gcr.io/PROJECT_ID/microservice-template:latest .
docker push gcr.io/PROJECT_ID/microservice-template:latest

# Deploy
make deploy-staging
```

### Deploy a Producción
```bash
make deploy-prod
```

## 🔐 Manejo de Secretos

### Con Vault (Recomendado)
```go
// Ejemplo de uso
vaultClient, err := vault.NewClient(cfg.VaultConfig)
secrets, err := vaultClient.GetSecret("secret/myapp/database")
password := secrets["password"].(string)
```

### Variables de Entorno
Para desarrollo local, usar archivos `.env.*`

## 📈 Monitoreo y Métricas

### Métricas Disponibles
- `http_requests_total` - Total de requests HTTP
- `http_request_duration_seconds` - Duración de requests

### Prometheus
Configuración en `monitoring/prometheus.yml`

## 🔄 Personalización del Template

### 1. Cambiar nombre del módulo
Actualizar en `go.mod`:
```go
module github.com/company/tu-microservicio
```

### 2. Agregar nuevos endpoints
```go
// En internal/handlers/handlers.go
api.GET("/tu-endpoint", h.TuHandler)
```

### 3. Agregar servicios externos
```go
// En internal/services/
type ExternalService interface {
    CallAPI() error
}
```

### 4. Configurar base de datos
Descomentar y configurar en:
- `internal/config/config.go`
- Scripts de migración en `scripts/`

## 📝 Comandos Útiles

```bash
# Ver todos los comandos disponibles
make help

# Desarrollo
make deps          # Instalar dependencias
make build         # Compilar
make run           # Ejecutar
make test          # Tests
make lint          # Linting
make format        # Formatear código

# Docker
make docker-build  # Construir imagen
make docker-dev    # Entorno completo
make docker-test   # Tests en Docker

# Documentación
make swagger       # Generar docs Swagger
```

## 🤝 Contribución

1. Fork el proyecto
2. Crear feature branch (`git checkout -b feature/nueva-funcionalidad`)
3. Commit cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push al branch (`git push origin feature/nueva-funcionalidad`)
5. Crear Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para detalles.

## 🆘 Soporte

Para preguntas o problemas:
1. Revisar la documentación
2. Buscar en issues existentes
3. Crear nuevo issue con detalles del problema

---

**Nota**: Este template incluye ejemplos comentados para facilitar el desarrollo. Descomenta y configura según las necesidades de tu microservicio.