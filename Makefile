# Makefile para el template de microservicio Go

.PHONY: help build run test clean docker-build docker-run docker-test lint format deps upgrade-deps

# Variables
APP_NAME=microservice-template
DOCKER_IMAGE=$(APP_NAME):latest
DOCKER_TEST_IMAGE=$(APP_NAME):test
MIGRATION_DIR=./migrations

help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Instalar dependencias
	go mod download
	go mod tidy

upgrade-deps: ## Actualizar dependencias
	go get -u ./...
	go mod tidy

build: ## Compilar la aplicación
	go build -o bin/$(APP_NAME) .

run: ## Ejecutar la aplicación localmente
	go run .

run-prod: ## Ejecutar con configuración de producción
	ENVIRONMENT=production go run .

test: ## Ejecutar tests unitarios
	go test ./internal/... -v -race -coverprofile=coverage.out

test-integration: ## Ejecutar tests de integración
	go test ./tests/integration/... -v -race

test-e2e: ## Ejecutar tests end-to-end
	go test ./tests/e2e/... -v -race

test-all: ## Ejecutar todos los tests
	go test ./... -v -race -coverprofile=coverage.out

test-coverage: test-all ## Ejecutar tests y mostrar cobertura
	go tool cover -html=coverage.out -o coverage.html
	@echo "Reporte de cobertura generado en coverage.html"

lint: ## Ejecutar linter
	golangci-lint run

format: ## Formatear código
	go fmt ./...
	goimports -w .

clean: ## Limpiar archivos generados
	rm -rf bin/
	rm -f coverage.out coverage.html

# Database migrations
migrate-create: ## Crear nueva migración (uso: make migrate-create NAME=create_users_table)
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)

migrate-up: ## Ejecutar migraciones de base de datos
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" up

migrate-down: ## Revertir última migración
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" down 1

migrate-force: ## Forzar versión de migración (uso: make migrate-force VERSION=1)
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" force $(VERSION)

seed: ## Ejecutar seeds de base de datos
	@echo "Ejecutando seeds..."
	go run scripts/seed.go

# Integration Service specific commands
init-integration-db: ## Inicializar base de datos para integration service
	@echo "Inicializando base de datos de integración..."
	psql $(DATABASE_URL) -f scripts/init-integration.sql

test-webhooks: ## Probar webhooks localmente
	@echo "Probando webhook de WhatsApp..."
	curl -X POST http://localhost:8080/api/v1/integrations/webhooks/whatsapp \
		-H "Content-Type: application/json" \
		-d '{"entry":[{"changes":[{"value":{"messages":[{"id":"test","from":"123","timestamp":"1640995200","text":{"body":"test"},"type":"text"}],"metadata":{"phone_number_id":"123"}}}]}]}'

test-send-message: ## Probar envío de mensaje
	@echo "Probando envío de mensaje..."
	curl -X POST http://localhost:8080/api/v1/integrations/send \
		-H "Content-Type: application/json" \
		-d '{"channel_id":"test-channel","recipient":"573001112233","content":{"type":"text","text":"Hello from Integration Service!"}}'

create-test-channel: ## Crear canal de prueba
	@echo "Creando canal de prueba..."
	curl -X POST http://localhost:8080/api/v1/integrations/channels \
		-H "Content-Type: application/json" \
		-d '{"tenant_id":"test-tenant","platform":"whatsapp","provider":"meta","access_token":"test-token","webhook_url":"https://test.com/webhook","config":{"phone_number_id":"123456789"}}'

# Docker commands
docker-build: ## Construir imagen Docker
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Ejecutar con Docker
	docker run -p 8080:8080 --env-file .env.local $(DOCKER_IMAGE)

docker-test: ## Ejecutar tests con Docker
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit

docker-dev: ## Levantar entorno de desarrollo completo
	docker-compose up --build

docker-down: ## Detener contenedores
	docker-compose down -v

docker-logs: ## Ver logs de contenedores
	docker-compose logs -f

# GCP Cloud Run deployment
gcp-build: ## Construir imagen en GCP
	gcloud builds submit --tag gcr.io/$(PROJECT_ID)/$(APP_NAME):latest

deploy-staging: ## Deploy a staging en Cloud Run
	gcloud run deploy $(APP_NAME)-staging \
		--image gcr.io/$(PROJECT_ID)/$(APP_NAME):latest \
		--platform managed \
		--region us-central1 \
		--allow-unauthenticated \
		--set-env-vars ENVIRONMENT=test \
		--set-secrets="JWT_SECRET=jwt-secret:latest,DB_PASSWORD=db-password:latest"

deploy-prod: ## Deploy a producción en Cloud Run
	gcloud run deploy $(APP_NAME) \
		--image gcr.io/$(PROJECT_ID)/$(APP_NAME):latest \
		--platform managed \
		--region us-central1 \
		--allow-unauthenticated \
		--set-env-vars ENVIRONMENT=production \
		--set-secrets="JWT_SECRET=jwt-secret-prod:latest,DB_PASSWORD=db-password-prod:latest"

# Swagger
swagger: ## Generar documentación Swagger
	swag init -g main.go

# Security
security-scan: ## Ejecutar escaneo de seguridad
	gosec ./...

# Performance
benchmark: ## Ejecutar benchmarks
	go test -bench=. -benchmem ./...

# Monitoring
metrics: ## Ver métricas de la aplicación
	curl http://localhost:8080/metrics

health: ## Verificar health de la aplicación
	curl http://localhost:8080/api/v1/health

# Development helpers
dev-setup: deps docker-dev migrate-up seed ## Setup completo para desarrollo
	@echo "Entorno de desarrollo listo!"

dev-reset: docker-down clean dev-setup ## Reset completo del entorno de desarrollo