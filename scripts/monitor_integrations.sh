#!/bin/bash

# Script de monitoreo para integraciones de mensajes
# Verifica el estado de puertos, endpoints y conectividad

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Función para mostrar mensajes
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_header() {
    echo -e "${PURPLE}[MONITOR]${NC} $1"
}

# Cargar variables de entorno
if [ -f .env.local ]; then
    source .env.local
fi

clear
echo "🔍 Monitor de Integraciones de Mensajes"
echo "======================================="
echo "Fecha: $(date)"
echo ""

# Función para verificar puerto
check_port() {
    local port=$1
    local service=$2
    
    if nc -z localhost $port 2>/dev/null; then
        log_success "Puerto $port ($service): ✅ ACTIVO"
        return 0
    else
        log_warning "Puerto $port ($service): ⚠️  INACTIVO"
        return 1
    fi
}

# Función para verificar endpoint HTTP
check_endpoint() {
    local url=$1
    local name=$2
    local expected_status=${3:-200}
    local method=${4:-GET}
    
    local status
    if [ "$method" = "POST" ]; then
        status=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d '{"test": true}' "$url" 2>/dev/null || echo "000")
    else
        status=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    fi
    
    if [ "$status" = "$expected_status" ]; then
        log_success "$name: ✅ ACTIVO (Status: $status)"
        return 0
    elif [ "$status" = "000" ]; then
        log_error "$name: ❌ NO ACCESIBLE"
        return 1
    else
        log_warning "$name: ⚠️  RESPUESTA INESPERADA (Status: $status)"
        return 1
    fi
}

# 1. Verificar servicios Docker
log_header "1. ESTADO DE SERVICIOS DOCKER"
echo "================================"

if docker info > /dev/null 2>&1; then
    log_success "Docker: ✅ CORRIENDO"
    
    # Verificar contenedores
    if docker-compose ps | grep -q "app.*Up"; then
        log_success "Contenedor app: ✅ CORRIENDO"
    else
        log_warning "Contenedor app: ⚠️  DETENIDO"
    fi
    
    if docker-compose ps | grep -q "postgres.*Up"; then
        log_success "Contenedor postgres: ✅ CORRIENDO"
    else
        log_warning "Contenedor postgres: ⚠️  DETENIDO (usar --profile local-db si es necesario)"
    fi
else
    log_error "Docker: ❌ NO DISPONIBLE"
fi

echo ""

# 2. Verificar puertos
log_header "2. ESTADO DE PUERTOS"
echo "====================="

check_port 8080 "Servicio Principal"
check_port 5433 "PostgreSQL Local"

echo ""

# 3. Verificar endpoints principales
log_header "3. ENDPOINTS PRINCIPALES"
echo "========================="

BASE_URL="http://localhost:8080"

check_endpoint "$BASE_URL/api/v1/health" "Health Check"
check_endpoint "$BASE_URL/api/v1/integrations/channels" "API Channels" "400"

echo ""

# 4. Verificar webhooks
log_header "4. WEBHOOKS DE INTEGRACIÓN"
echo "==========================="

check_endpoint "$BASE_URL/api/v1/integrations/webhooks/telegram" "Telegram Webhook" "500" "POST"
check_endpoint "$BASE_URL/api/v1/integrations/webhooks/whatsapp" "WhatsApp Webhook" "403"
check_endpoint "$BASE_URL/api/v1/integrations/webhooks/messenger" "Messenger Webhook" "403"
check_endpoint "$BASE_URL/api/v1/integrations/webhooks/instagram" "Instagram Webhook" "403"
check_endpoint "$BASE_URL/api/v1/integrations/webhooks/webchat" "Webchat Webhook" "500" "POST"

echo ""

# 5. Verificar conectividad de base de datos
log_header "5. CONECTIVIDAD DE BASE DE DATOS"
echo "================================="

if [ -n "$DB_HOST" ] && [ -n "$DB_PORT" ]; then
    if nc -z $DB_HOST $DB_PORT 2>/dev/null; then
        log_success "Base de datos externa ($DB_HOST:$DB_PORT): ✅ ACCESIBLE"
    else
        log_warning "Base de datos externa ($DB_HOST:$DB_PORT): ⚠️  NO ACCESIBLE"
    fi
else
    log_warning "Configuración de BD externa no encontrada"
fi

echo ""

# 6. Verificar configuración de túnel (si existe)
log_header "6. CONFIGURACIÓN DE TÚNEL"
echo "=========================="

if [ -n "$WEBHOOK_BASE_URL" ] && [[ "$WEBHOOK_BASE_URL" == *"loca.lt"* ]]; then
    TUNNEL_URL="$WEBHOOK_BASE_URL"
    log_info "Túnel configurado: $TUNNEL_URL"
    
    # Verificar accesibilidad del túnel
    if curl -s -m 10 "$TUNNEL_URL/api/v1/health" > /dev/null 2>&1; then
        log_success "Túnel: ✅ ACCESIBLE"
        
        # Verificar webhooks a través del túnel
        echo "Verificando webhooks a través del túnel:"
        check_endpoint "$TUNNEL_URL/api/v1/integrations/webhooks/telegram" "  Telegram (Túnel)" "400"
        check_endpoint "$TUNNEL_URL/api/v1/integrations/webhooks/whatsapp" "  WhatsApp (Túnel)" "403"
    else
        log_warning "Túnel: ⚠️  NO ACCESIBLE"
    fi
else
    log_info "No hay túnel configurado (usando localhost)"
fi

echo ""

# 7. Verificar logs recientes
log_header "7. LOGS RECIENTES"
echo "=================="

if docker-compose ps | grep -q "app.*Up"; then
    log_info "Últimas 5 líneas de logs del servicio:"
    docker-compose logs --tail=5 app | sed 's/^/  /'
else
    log_warning "Servicio no está corriendo, no hay logs disponibles"
fi

echo ""

# 8. Resumen de configuración
log_header "8. RESUMEN DE CONFIGURACIÓN"
echo "============================"

echo "Puerto del servicio: ${PORT:-8080}"
echo "Entorno: ${ENVIRONMENT:-development}"
echo "Base de datos: ${DB_HOST:-localhost}:${DB_PORT:-5432}"
echo "URL de webhooks: ${WEBHOOK_BASE_URL:-http://localhost:8080}"
echo ""

# Mostrar URLs importantes
echo "📋 URLs IMPORTANTES:"
echo "===================="
echo "Health Check: $BASE_URL/api/v1/health"
echo "API Docs: $BASE_URL/swagger/index.html (si está habilitado)"
echo "Métricas: $BASE_URL/metrics (si está habilitado)"
echo ""
echo "Webhooks:"
echo "  Telegram: $BASE_URL/api/v1/integrations/webhooks/telegram"
echo "  WhatsApp: $BASE_URL/api/v1/integrations/webhooks/whatsapp"
echo "  Messenger: $BASE_URL/api/v1/integrations/webhooks/messenger"
echo "  Instagram: $BASE_URL/api/v1/integrations/webhooks/instagram"
echo ""

# Opción para monitoreo continuo
echo ""
read -p "¿Quieres activar monitoreo continuo cada 30 segundos? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Iniciando monitoreo continuo (Ctrl+C para salir)..."
    echo ""
    
    while true; do
        sleep 30
        clear
        echo "🔄 Monitoreo Continuo - $(date)"
        echo "================================"
        
        # Solo verificar lo esencial en modo continuo
        check_port 8080 "Servicio Principal"
        check_endpoint "$BASE_URL/api/v1/health" "Health Check"
        
        if [ -n "$WEBHOOK_BASE_URL" ] && [[ "$WEBHOOK_BASE_URL" == *"loca.lt"* ]]; then
            check_endpoint "$WEBHOOK_BASE_URL/api/v1/health" "Túnel"
        fi
        
        echo ""
        echo "Presiona Ctrl+C para salir del monitoreo continuo"
    done
fi

log_success "Monitoreo completado!"