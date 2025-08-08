# Makefile para el servicio de integraciones

.PHONY: help build test run clean docker-build docker-run dev test-integration

# Variables
APP_NAME=integration-service
DOCKER_IMAGE=gcr.io/$(PROJECT_ID)/$(APP_NAME)
GO_VERSION=1.21

help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Desarrollo local
dev: ## Iniciar entorno de desarrollo con hot reload
	@echo "🚀 Iniciando entorno de desarrollo..."
	docker-compose -f docker-compose.yml up --build app

dev-simple: ## Ejecutar aplicación directamente (sin Docker)
	@echo "🚀 Ejecutando aplicación localmente..."
	@echo "Modo desarrollo: usando mocks, sin base de datos"
	go run main-dev.go

dev-test: ## Iniciar entorno de testing
	@echo "🧪 Iniciando entorno de testing..."
	docker-compose -f docker-compose.test.yml up --build

dev-down: ## Detener entorno de desarrollo
	@echo "🛑 Deteniendo entorno de desarrollo..."
	docker-compose -f docker-compose.yml down
	docker-compose -f docker-compose.test.yml down

dev-logs: ## Ver logs del entorno de desarrollo
	docker-compose -f docker-compose.yml logs -f app

# Testing
test: ## Ejecutar tests unitarios
	@echo "🧪 Ejecutando tests unitarios..."
	go test ./... -v -race -coverprofile=coverage.out

test-coverage: test ## Ejecutar tests y mostrar coverage
	@echo "📊 Generando reporte de coverage..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Reporte generado en coverage.html"

test-integration: ## Ejecutar tests de integración
	@echo "🔗 Ejecutando tests de integración..."
	docker-compose -f docker-compose.test.yml up -d
	@sleep 10
	go test ./tests/integration/... -v -tags=integration
	docker-compose -f docker-compose.test.yml down

# Build
build: ## Compilar la aplicación
	@echo "🔨 Compilando aplicación..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(APP_NAME) .

build-local: ## Compilar para desarrollo local
	@echo "🔨 Compilando para desarrollo local..."
	go build -o bin/$(APP_NAME) .

# Docker
docker-build: ## Construir imagen Docker
	@echo "🐳 Construyendo imagen Docker..."
	docker build -t $(APP_NAME):latest .

docker-run: ## Ejecutar contenedor Docker
	@echo "🐳 Ejecutando contenedor Docker..."
	docker run -p 8080:8080 --env-file .env.local $(APP_NAME):latest

# Herramientas de desarrollo
fmt: ## Formatear código
	@echo "✨ Formateando código..."
	go fmt ./...

lint: ## Ejecutar linter
	@echo "🔍 Ejecutando linter..."
	golangci-lint run

deps: ## Descargar dependencias
	@echo "📦 Descargando dependencias..."
	go mod download
	go mod tidy

# Base de datos
db-migrate: ## Ejecutar migraciones de base de datos
	@echo "🗄️ Ejecutando migraciones..."
	@echo "Las migraciones se manejan en el servicio de migraciones"

db-reset: ## Resetear base de datos de desarrollo
	@echo "🗄️ Reseteando base de datos..."
	docker-compose -f docker-compose.yml down -v
	docker-compose -f docker-compose.yml up -d postgres
	@sleep 5
	docker-compose -f docker-compose.yml up -d

# Herramientas de integración
webhook-simulator: ## Abrir simulador de webhooks
	@echo "🔗 Abriendo simulador de webhooks..."
	@echo "Simulador disponible en: http://localhost:8081"
	docker-compose -f docker-compose.test.yml up -d webhook-simulator

ngrok: ## Exponer webhooks con ngrok (requiere NGROK_AUTHTOKEN)
	@echo "🌐 Exponiendo webhooks con ngrok..."
	docker-compose -f docker-compose.test.yml --profile ngrok up -d ngrok
	@echo "Dashboard de ngrok: http://localhost:4040"

# Monitoreo
metrics: ## Ver métricas de Prometheus
	@echo "📊 Métricas disponibles en: http://localhost:9090"
	docker-compose -f docker-compose.yml --profile monitoring up -d prometheus

logs: ## Ver logs de la aplicación
	@echo "📋 Viendo logs..."
	docker-compose -f docker-compose.yml logs -f app

# Limpieza
clean: ## Limpiar archivos generados
	@echo "🧹 Limpiando archivos generados..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker system prune -f

clean-all: clean ## Limpieza completa incluyendo volúmenes
	@echo "🧹 Limpieza completa..."
	docker-compose -f docker-compose.yml down -v
	docker-compose -f docker-compose.test.yml down -v
	docker volume prune -f

# Despliegue
deploy-staging: ## Desplegar a staging
	@echo "🚀 Desplegando a staging..."
	gcloud builds submit --config cloudbuild.yaml --substitutions=_ENV=staging

deploy-prod: ## Desplegar a producción
	@echo "🚀 Desplegando a producción..."
	gcloud builds submit --config cloudbuild.yaml --substitutions=_ENV=production

# Utilidades
health: ## Verificar health check
	@echo "❤️ Verificando health check..."
	curl -f http://localhost:8080/api/v1/health || echo "Servicio no disponible"

test-endpoints: ## Probar todos los endpoints
	@echo "🔍 Probando endpoints..."
	@if pgrep -f "integration-service" > /dev/null || curl -s http://localhost:8080/api/v1/health > /dev/null 2>&1; then \
		chmod +x scripts/test-endpoints.sh; \
		./scripts/test-endpoints.sh; \
	else \
		echo "❌ El servicio no está corriendo. Ejecuta 'make dev' o 'make dev-simple' primero."; \
	fi

status: ## Ver estado de los servicios
	@echo "📊 Estado de los servicios:"
	docker-compose -f docker-compose.yml ps

# Configuración inicial
setup: ## Configuración inicial del proyecto
	@echo "⚙️ Configuración inicial..."
	@echo "1. Copiando archivos de configuración..."
	@if [ ! -f .env.local ]; then cp .env.example .env.local; fi
	@echo "2. Descargando dependencias..."
	go mod download
	@echo "3. Instalando herramientas de desarrollo..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Configuración completada"
	@echo ""
	@echo "Próximos pasos:"
	@echo "1. Editar .env.local con tus configuraciones"
	@echo "2. Ejecutar 'make dev' para iniciar el entorno de desarrollo"
	@echo "3. Abrir http://localhost:8081 para el simulador de webhooks"

# Información del proyecto
info: ## Mostrar información del proyecto
	@echo "📋 Información del proyecto:"
	@echo "Nombre: $(APP_NAME)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Docker Image: $(DOCKER_IMAGE)"
	@echo ""
	@echo "URLs importantes:"
	@echo "- API: http://localhost:8080"
	@echo "- Health Check: http://localhost:8080/api/v1/health"
	@echo "- Webhook Simulator: http://localhost:8081"
	@echo "- Prometheus: http://localhost:9090 (con --profile monitoring)"