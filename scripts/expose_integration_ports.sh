#!/bin/bash

# Script para exponer puertos de integración de mensajes
# Telegram y WhatsApp

set -e

echo "🚀 Configurando puertos de integración para Telegram y WhatsApp..."

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Verificar si Docker está corriendo
if ! docker info > /dev/null 2>&1; then
    log_error "Docker no está corriendo. Por favor inicia Docker primero."
    exit 1
fi

# Verificar si el servicio está corriendo
log_info "Verificando estado del servicio..."
if docker-compose ps | grep -q "app.*Up"; then
    log_success "Servicio principal está corriendo"
else
    log_warning "Servicio principal no está corriendo. Iniciando..."
    docker-compose up -d app
fi

# Mostrar puertos expuestos
log_info "Puertos expuestos para integraciones:"
echo ""
echo "📱 TELEGRAM:"
echo "   Webhook URL: http://localhost:8080/api/v1/integrations/webhooks/telegram"
echo "   Método: POST"
echo "   Puerto: 8080"
echo ""
echo "💬 WHATSAPP:"
echo "   Webhook URL: http://localhost:8080/api/v1/integrations/webhooks/whatsapp"
echo "   Método: POST/GET (verificación)"
echo "   Puerto: 8080"
echo ""
echo "🌐 OTROS ENDPOINTS DISPONIBLES:"
echo "   Messenger: http://localhost:8080/api/v1/integrations/webhooks/messenger"
echo "   Instagram: http://localhost:8080/api/v1/integrations/webhooks/instagram"
echo "   Webchat: http://localhost:8080/api/v1/integrations/webhooks/webchat"
echo ""
echo "🔧 GESTIÓN DE CANALES:"
echo "   GET    /api/v1/integrations/channels - Listar integraciones"
echo "   POST   /api/v1/integrations/channels - Crear integración"
echo "   GET    /api/v1/integrations/channels/{id} - Obtener integración"
echo "   PATCH  /api/v1/integrations/channels/{id} - Actualizar integración"
echo "   DELETE /api/v1/integrations/channels/{id} - Eliminar integración"
echo ""
echo "📨 MENSAJERÍA:"
echo "   POST /api/v1/integrations/send - Enviar mensaje"
echo "   GET  /api/v1/integrations/messages/inbound - Mensajes entrantes"
echo "   GET  /api/v1/integrations/chat/{platform}/{user_id} - Historial de chat"
echo ""

# Verificar conectividad de puertos
log_info "Verificando conectividad de puertos..."

# Puerto principal (8080)
if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/health | grep -q "200"; then
    log_success "Puerto 8080 - Servicio principal: ✅ ACTIVO"
else
    log_warning "Puerto 8080 - Servicio principal: ⚠️  NO RESPONDE"
fi

# Puerto de base de datos (5433)
if nc -z localhost 5433 2>/dev/null; then
    log_success "Puerto 5433 - PostgreSQL: ✅ ACTIVO"
else
    log_warning "Puerto 5433 - PostgreSQL: ⚠️  NO ACTIVO (usar --profile local-db si necesitas BD local)"
fi

echo ""
log_info "Configuración de túnel para webhooks externos:"
echo ""
echo "Para recibir webhooks de Telegram y WhatsApp desde internet:"
echo ""
echo "1. Instalar localtunnel (si no lo tienes):"
echo "   npm install -g localtunnel"
echo ""
echo "2. Crear túnel para el puerto 8080:"
echo "   lt --port 8080 --subdomain tu-subdominio-unico"
echo ""
echo "3. Configurar webhooks con la URL del túnel:"
echo "   Telegram: https://tu-subdominio-unico.loca.lt/api/v1/integrations/webhooks/telegram"
echo "   WhatsApp: https://tu-subdominio-unico.loca.lt/api/v1/integrations/webhooks/whatsapp"
echo ""

# Mostrar logs en tiempo real (opcional)
read -p "¿Quieres ver los logs del servicio en tiempo real? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Mostrando logs del servicio (Ctrl+C para salir)..."
    docker-compose logs -f app
fi

log_success "Configuración de puertos completada!"