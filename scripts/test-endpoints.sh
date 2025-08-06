#!/bin/bash

# Script para probar todos los endpoints del servicio de integración
# Asegúrate de que el servicio esté corriendo en localhost:8080

BASE_URL="http://localhost:8080"
TENANT_ID="tenant_demo_123"

echo "🚀 Iniciando pruebas de endpoints del IT Integration Service"
echo "=================================================="

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Función para hacer requests y mostrar resultados
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo "Method: $method"
    echo "Endpoint: $endpoint"
    
    if [ -n "$data" ]; then
        echo "Data: $data"
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X $method \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X $method \
            "$BASE_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✅ SUCCESS (HTTP $http_code)${NC}"
    else
        echo -e "${RED}❌ FAILED (HTTP $http_code)${NC}"
    fi
    
    echo "Response: $body"
    echo "----------------------------------------"
}

# 1. Health Checks
echo -e "\n${YELLOW}=== HEALTH CHECKS ===${NC}"
test_endpoint "GET" "/api/v1/health" "" "Health Check"
test_endpoint "GET" "/api/v1/ready" "" "Readiness Check"

# 2. Channel Management
echo -e "\n${YELLOW}=== CHANNEL MANAGEMENT ===${NC}"
test_endpoint "GET" "/api/v1/integrations/channels?tenant_id=$TENANT_ID" "" "Get All Channels"

# Crear un canal de prueba
channel_data='{
  "tenant_id": "'$TENANT_ID'",
  "platform": "whatsapp",
  "provider": "meta",
  "access_token": "test_token_123",
  "webhook_url": "https://example.com/webhook",
  "status": "active",
  "config": {
    "phone_number_id": "123456789",
    "business_account_id": "987654321"
  }
}'

test_endpoint "POST" "/api/v1/integrations/channels" "$channel_data" "Create Channel"

# 3. Message Operations
echo -e "\n${YELLOW}=== MESSAGE OPERATIONS ===${NC}"

# Envío de mensaje simple
message_data='{
  "channel_id": "mock-channel-1",
  "recipient": "573001234567",
  "content": {
    "type": "text",
    "text": "¡Hola! Este es un mensaje de prueba desde IT App Chat."
  }
}'

test_endpoint "POST" "/api/v1/integrations/send" "$message_data" "Send Single Message"

# Envío de mensaje con media
media_message_data='{
  "channel_id": "mock-channel-1",
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

test_endpoint "POST" "/api/v1/integrations/send" "$media_message_data" "Send Media Message"

# Broadcast message
broadcast_data='{
  "tenant_id": "'$TENANT_ID'",
  "platforms": ["whatsapp", "telegram"],
  "recipients": ["573001234567", "573009876543"],
  "content": {
    "type": "text",
    "text": "📢 Mensaje masivo: ¡Hola a todos! Este es un mensaje enviado a múltiples plataformas."
  }
}'

test_endpoint "POST" "/api/v1/integrations/broadcast" "$broadcast_data" "Broadcast Message"

# 4. Message History
echo -e "\n${YELLOW}=== MESSAGE HISTORY ===${NC}"
test_endpoint "GET" "/api/v1/integrations/messages/inbound?platform=whatsapp&limit=10" "" "Get Inbound Messages"
test_endpoint "GET" "/api/v1/integrations/messages/outbound?platform=telegram&limit=10" "" "Get Outbound Messages"
test_endpoint "GET" "/api/v1/integrations/chat/whatsapp/user_123" "" "Get Chat History"

# 5. Telegram Setup (requiere token real para funcionar completamente)
echo -e "\n${YELLOW}=== TELEGRAM SETUP ===${NC}"
if [ -n "$TELEGRAM_BOT_TOKEN" ]; then
    test_endpoint "GET" "/api/v1/integrations/telegram/bot-info?bot_token=$TELEGRAM_BOT_TOKEN" "" "Get Telegram Bot Info"
    
    telegram_setup_data='{
      "bot_token": "'$TELEGRAM_BOT_TOKEN'",
      "webhook_url": "https://your-domain.com/api/v1/integrations/webhooks/telegram",
      "tenant_id": "'$TENANT_ID'"
    }'
    
    test_endpoint "POST" "/api/v1/integrations/telegram/setup" "$telegram_setup_data" "Setup Telegram Integration"
else
    echo -e "${YELLOW}⚠️  TELEGRAM_BOT_TOKEN not set, skipping Telegram tests${NC}"
fi

# 6. WhatsApp Setup (requiere tokens reales para funcionar completamente)
echo -e "\n${YELLOW}=== WHATSAPP SETUP ===${NC}"
if [ -n "$WHATSAPP_ACCESS_TOKEN" ] && [ -n "$WHATSAPP_PHONE_ID" ]; then
    test_endpoint "GET" "/api/v1/integrations/whatsapp/phone-info?access_token=$WHATSAPP_ACCESS_TOKEN&phone_number_id=$WHATSAPP_PHONE_ID" "" "Get WhatsApp Phone Info"
    
    whatsapp_setup_data='{
      "access_token": "'$WHATSAPP_ACCESS_TOKEN'",
      "phone_number_id": "'$WHATSAPP_PHONE_ID'",
      "business_account_id": "'$WHATSAPP_BUSINESS_ID'",
      "webhook_url": "https://your-domain.com/api/v1/integrations/webhooks/whatsapp",
      "tenant_id": "'$TENANT_ID'"
    }'
    
    test_endpoint "POST" "/api/v1/integrations/whatsapp/setup" "$whatsapp_setup_data" "Setup WhatsApp Integration"
else
    echo -e "${YELLOW}⚠️  WhatsApp tokens not set, skipping WhatsApp tests${NC}"
fi

# 7. Webhook Tests (simulados)
echo -e "\n${YELLOW}=== WEBHOOK TESTS ===${NC}"

# Webhook de WhatsApp (verificación)
test_endpoint "GET" "/api/v1/integrations/webhooks/whatsapp?hub.mode=subscribe&hub.verify_token=test-token&hub.challenge=12345" "" "WhatsApp Webhook Verification"

# Webhook de Telegram (mensaje simulado)
telegram_webhook_data='{
  "update_id": 123456789,
  "message": {
    "message_id": 1,
    "from": {
      "id": 987654321,
      "is_bot": false,
      "first_name": "Usuario",
      "username": "usuario_test"
    },
    "chat": {
      "id": 987654321,
      "first_name": "Usuario",
      "type": "private"
    },
    "date": 1234567890,
    "text": "Hola bot, este es un mensaje de prueba"
  }
}'

test_endpoint "POST" "/api/v1/integrations/webhooks/telegram" "$telegram_webhook_data" "Telegram Webhook"

# Webhook de Webchat
webchat_webhook_data='{
  "type": "message",
  "user_id": "webchat_user_123",
  "session_id": "session_abc123",
  "message": {
    "text": "Hola, necesito ayuda con mi pedido",
    "timestamp": "2024-01-15T10:30:00Z"
  },
  "metadata": {
    "page_url": "https://example.com/contact",
    "user_agent": "Mozilla/5.0...",
    "ip_address": "192.168.1.1"
  }
}'

test_endpoint "POST" "/api/v1/integrations/webhooks/webchat" "$webchat_webhook_data" "Webchat Webhook"

echo -e "\n${GREEN}🎉 Pruebas completadas!${NC}"
echo "=================================================="
echo -e "${YELLOW}Notas:${NC}"
echo "- Para pruebas completas de Telegram, configura: export TELEGRAM_BOT_TOKEN='tu_token'"
echo "- Para pruebas completas de WhatsApp, configura:"
echo "  export WHATSAPP_ACCESS_TOKEN='tu_token'"
echo "  export WHATSAPP_PHONE_ID='tu_phone_id'"
echo "  export WHATSAPP_BUSINESS_ID='tu_business_id'"
echo "- Algunos endpoints pueden fallar si la base de datos no está configurada (modo mock activo)"
echo "- Los webhooks reales requieren configuración de signatures y tokens de verificación"