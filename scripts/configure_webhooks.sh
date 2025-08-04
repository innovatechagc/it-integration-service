#!/bin/bash

# Script para configurar webhooks de Telegram y WhatsApp
# Automatiza la configuración de URLs de webhook

set -e

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

# Cargar variables de entorno
if [ -f .env.local ]; then
    source .env.local
fi

echo "🔧 Configurador de Webhooks para Integraciones"
echo "=============================================="
echo ""

# Solicitar URL base del túnel
read -p "Ingresa la URL base de tu túnel (ej: https://tu-subdominio.loca.lt): " TUNNEL_URL

if [ -z "$TUNNEL_URL" ]; then
    log_error "URL del túnel es requerida"
    exit 1
fi

# Remover trailing slash si existe
TUNNEL_URL=${TUNNEL_URL%/}

echo ""
log_info "Configurando webhooks con URL base: $TUNNEL_URL"
echo ""

# Configurar Telegram
echo "📱 CONFIGURACIÓN DE TELEGRAM"
echo "=============================="

if [ -z "$TELEGRAM_BOT_TOKEN" ] || [ "$TELEGRAM_BOT_TOKEN" = "your_test_bot_token" ]; then
    read -p "Ingresa tu Telegram Bot Token: " TELEGRAM_BOT_TOKEN
fi

if [ -n "$TELEGRAM_BOT_TOKEN" ]; then
    TELEGRAM_WEBHOOK_URL="$TUNNEL_URL/api/v1/integrations/webhooks/telegram"
    
    log_info "Configurando webhook de Telegram..."
    log_info "URL: $TELEGRAM_WEBHOOK_URL"
    
    # Configurar webhook de Telegram
    RESPONSE=$(curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook" \
        -H "Content-Type: application/json" \
        -d "{\"url\":\"$TELEGRAM_WEBHOOK_URL\"}")
    
    if echo "$RESPONSE" | grep -q '"ok":true'; then
        log_success "Webhook de Telegram configurado correctamente"
        
        # Obtener información del webhook
        WEBHOOK_INFO=$(curl -s "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo")
        echo "Información del webhook:"
        echo "$WEBHOOK_INFO" | jq '.' 2>/dev/null || echo "$WEBHOOK_INFO"
    else
        log_error "Error configurando webhook de Telegram:"
        echo "$RESPONSE"
    fi
else
    log_warning "Token de Telegram no configurado, saltando..."
fi

echo ""

# Configurar WhatsApp (Meta)
echo "💬 CONFIGURACIÓN DE WHATSAPP"
echo "============================="

if [ -z "$META_APP_ID" ] || [ "$META_APP_ID" = "your_test_app_id" ]; then
    read -p "Ingresa tu Meta App ID: " META_APP_ID
fi

if [ -z "$META_APP_SECRET" ] || [ "$META_APP_SECRET" = "your_test_app_secret" ]; then
    read -p "Ingresa tu Meta App Secret: " META_APP_SECRET
fi

if [ -z "$META_VERIFY_TOKEN" ] || [ "$META_VERIFY_TOKEN" = "your_webhook_verify_token" ]; then
    read -p "Ingresa tu Meta Verify Token: " META_VERIFY_TOKEN
fi

if [ -n "$META_APP_ID" ] && [ -n "$META_APP_SECRET" ]; then
    WHATSAPP_WEBHOOK_URL="$TUNNEL_URL/api/v1/integrations/webhooks/whatsapp"
    
    log_info "Configurando webhook de WhatsApp..."
    log_info "URL: $WHATSAPP_WEBHOOK_URL"
    log_info "Verify Token: $META_VERIFY_TOKEN"
    
    echo ""
    log_warning "Para WhatsApp, debes configurar manualmente en Meta Developer Console:"
    echo "1. Ve a https://developers.facebook.com/apps/$META_APP_ID/webhooks/"
    echo "2. Configura el webhook con:"
    echo "   - Callback URL: $WHATSAPP_WEBHOOK_URL"
    echo "   - Verify Token: $META_VERIFY_TOKEN"
    echo "3. Suscríbete a los eventos: messages, messaging_postbacks"
    
else
    log_warning "Credenciales de Meta no configuradas, saltando..."
fi

echo ""

# Actualizar archivo .env.local con las nuevas configuraciones
log_info "Actualizando archivo .env.local..."

# Crear backup del archivo original
cp .env.local .env.local.backup

# Actualizar variables
if [ -n "$TELEGRAM_BOT_TOKEN" ]; then
    sed -i "s/TELEGRAM_BOT_TOKEN=.*/TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN/" .env.local
fi

if [ -n "$META_APP_ID" ]; then
    sed -i "s/META_APP_ID=.*/META_APP_ID=$META_APP_ID/" .env.local
fi

if [ -n "$META_APP_SECRET" ]; then
    sed -i "s/META_APP_SECRET=.*/META_APP_SECRET=$META_APP_SECRET/" .env.local
fi

if [ -n "$META_VERIFY_TOKEN" ]; then
    sed -i "s/META_VERIFY_TOKEN=.*/META_VERIFY_TOKEN=$META_VERIFY_TOKEN/" .env.local
fi

# Actualizar WEBHOOK_BASE_URL
sed -i "s|WEBHOOK_BASE_URL=.*|WEBHOOK_BASE_URL=$TUNNEL_URL|" .env.local

log_success "Archivo .env.local actualizado"

echo ""
echo "🧪 PRUEBAS DE CONECTIVIDAD"
echo "=========================="

# Probar endpoint de salud
log_info "Probando endpoint de salud..."
if curl -s "$TUNNEL_URL/health" | grep -q "ok"; then
    log_success "Endpoint de salud: ✅ ACTIVO"
else
    log_warning "Endpoint de salud: ⚠️  NO RESPONDE"
fi

# Probar endpoints de webhook
log_info "Probando endpoints de webhook..."

# Telegram
TELEGRAM_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$TUNNEL_URL/api/v1/integrations/webhooks/telegram" \
    -H "Content-Type: application/json" \
    -d '{"test": true}')

if [ "$TELEGRAM_STATUS" = "200" ] || [ "$TELEGRAM_STATUS" = "400" ]; then
    log_success "Webhook Telegram: ✅ ACCESIBLE"
else
    log_warning "Webhook Telegram: ⚠️  NO ACCESIBLE (Status: $TELEGRAM_STATUS)"
fi

# WhatsApp (verificación GET)
WHATSAPP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$TUNNEL_URL/api/v1/integrations/webhooks/whatsapp?hub.mode=subscribe&hub.verify_token=$META_VERIFY_TOKEN&hub.challenge=123")

if [ "$WHATSAPP_STATUS" = "200" ] || [ "$WHATSAPP_STATUS" = "403" ]; then
    log_success "Webhook WhatsApp: ✅ ACCESIBLE"
else
    log_warning "Webhook WhatsApp: ⚠️  NO ACCESIBLE (Status: $WHATSAPP_STATUS)"
fi

echo ""
echo "📋 RESUMEN DE CONFIGURACIÓN"
echo "==========================="
echo "Telegram Webhook: $TUNNEL_URL/api/v1/integrations/webhooks/telegram"
echo "WhatsApp Webhook: $TUNNEL_URL/api/v1/integrations/webhooks/whatsapp"
echo "Messenger Webhook: $TUNNEL_URL/api/v1/integrations/webhooks/messenger"
echo "Instagram Webhook: $TUNNEL_URL/api/v1/integrations/webhooks/instagram"
echo ""
echo "Verify Token (Meta): $META_VERIFY_TOKEN"
echo ""

log_success "Configuración de webhooks completada!"
echo ""
log_info "Recuerda reiniciar el servicio para aplicar los cambios:"
echo "docker-compose restart app"