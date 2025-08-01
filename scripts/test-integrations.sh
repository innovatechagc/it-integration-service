#!/bin/bash

# Script para probar integraciones localmente
# Uso: ./scripts/test-integrations.sh [platform]

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
PLATFORM=${1:-"all"}

echo -e "${BLUE}🧪 Testing Integration Service${NC}"
echo -e "${BLUE}================================${NC}"

# Función para hacer requests
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo -e "${BLUE}$method $endpoint${NC}"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method \
            -H "Content-Type: application/json" \
            "$BASE_URL$endpoint")
    fi
    
    # Separar body y status code
    body=$(echo "$response" | head -n -1)
    status_code=$(echo "$response" | tail -n 1)
    
    if [[ $status_code -ge 200 && $status_code -lt 300 ]]; then
        echo -e "${GREEN}✅ Success ($status_code)${NC}"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        echo -e "${RED}❌ Failed ($status_code)${NC}"
        echo "$body"
    fi
}

# Test Health Check
test_health() {
    echo -e "\n${BLUE}🏥 Health Check${NC}"
    make_request "GET" "/api/v1/health" "" "Health Check"
}

# Test Integrations CRUD
test_integrations_crud() {
    echo -e "\n${BLUE}🔗 Integration Management${NC}"
    
    # List integrations
    make_request "GET" "/api/v1/integrations" "" "List all integrations"
    
    # Create integration
    local integration_data='{
        "tenant_id": "test-tenant-1",
        "platform": "whatsapp",
        "provider": "meta",
        "access_token": "test_token_123",
        "webhook_url": "http://localhost:8080/webhooks/whatsapp/meta",
        "config": {
            "phone_number_id": "123456789",
            "business_account_id": "987654321"
        }
    }'
    
    make_request "POST" "/api/v1/integrations" "$integration_data" "Create WhatsApp integration"
    
    # Get specific integration (assuming ID from previous response)
    make_request "GET" "/api/v1/integrations/test-tenant-1/whatsapp" "" "Get WhatsApp integration"
}

# Test WhatsApp Webhooks
test_whatsapp_webhooks() {
    echo -e "\n${BLUE}📱 WhatsApp Webhooks${NC}"
    
    # Text message webhook
    local whatsapp_text='{
        "object": "whatsapp_business_account",
        "entry": [{
            "id": "WHATSAPP_BUSINESS_ACCOUNT_ID",
            "changes": [{
                "value": {
                    "messaging_product": "whatsapp",
                    "metadata": {
                        "display_phone_number": "15550559999",
                        "phone_number_id": "PHONE_NUMBER_ID"
                    },
                    "messages": [{
                        "from": "16505551234",
                        "id": "wamid.test123",
                        "timestamp": "1669233778",
                        "text": {
                            "body": "Hello from test script!"
                        },
                        "type": "text"
                    }]
                },
                "field": "messages"
            }]
        }]
    }'
    
    make_request "POST" "/webhooks/whatsapp/meta" "$whatsapp_text" "WhatsApp text message webhook"
    
    # Status update webhook
    local whatsapp_status='{
        "object": "whatsapp_business_account",
        "entry": [{
            "id": "WHATSAPP_BUSINESS_ACCOUNT_ID",
            "changes": [{
                "value": {
                    "messaging_product": "whatsapp",
                    "metadata": {
                        "display_phone_number": "15550559999",
                        "phone_number_id": "PHONE_NUMBER_ID"
                    },
                    "statuses": [{
                        "id": "wamid.test123",
                        "status": "delivered",
                        "timestamp": "1669233778",
                        "recipient_id": "16505551234"
                    }]
                },
                "field": "messages"
            }]
        }]
    }'
    
    make_request "POST" "/webhooks/whatsapp/meta" "$whatsapp_status" "WhatsApp status update webhook"
}

# Test Telegram Webhooks
test_telegram_webhooks() {
    echo -e "\n${BLUE}✈️ Telegram Webhooks${NC}"
    
    local telegram_message='{
        "update_id": 123456789,
        "message": {
            "message_id": 1234,
            "from": {
                "id": 987654321,
                "is_bot": false,
                "first_name": "Test",
                "last_name": "User",
                "username": "testuser"
            },
            "chat": {
                "id": 987654321,
                "first_name": "Test",
                "last_name": "User",
                "username": "testuser",
                "type": "private"
            },
            "date": 1669233778,
            "text": "Hello from Telegram test!"
        }
    }'
    
    make_request "POST" "/webhooks/telegram" "$telegram_message" "Telegram message webhook"
}

# Test Message Sending
test_message_sending() {
    echo -e "\n${BLUE}📤 Message Sending${NC}"
    
    local send_message='{
        "channel_id": "660e8400-e29b-41d4-a716-446655440000",
        "recipient": "+1234567890",
        "content": {
            "type": "text",
            "text": "Hello! This is a test message from the integration service."
        }
    }'
    
    make_request "POST" "/api/v1/messages/send" "$send_message" "Send text message"
}

# Test Metrics
test_metrics() {
    echo -e "\n${BLUE}📊 Metrics${NC}"
    make_request "GET" "/metrics" "" "Prometheus metrics"
}

# Main execution
main() {
    echo -e "${YELLOW}Platform: $PLATFORM${NC}"
    echo -e "${YELLOW}Base URL: $BASE_URL${NC}"
    
    # Verificar que el servicio esté corriendo
    if ! curl -s "$BASE_URL/api/v1/health" > /dev/null; then
        echo -e "${RED}❌ Service is not running at $BASE_URL${NC}"
        echo -e "${YELLOW}💡 Run 'make dev' to start the service${NC}"
        exit 1
    fi
    
    # Ejecutar tests según la plataforma
    case $PLATFORM in
        "all")
            test_health
            test_integrations_crud
            test_whatsapp_webhooks
            test_telegram_webhooks
            test_message_sending
            test_metrics
            ;;
        "whatsapp")
            test_health
            test_whatsapp_webhooks
            ;;
        "telegram")
            test_health
            test_telegram_webhooks
            ;;
        "health")
            test_health
            ;;
        "crud")
            test_integrations_crud
            ;;
        "send")
            test_message_sending
            ;;
        *)
            echo -e "${RED}❌ Unknown platform: $PLATFORM${NC}"
            echo -e "${YELLOW}Available options: all, whatsapp, telegram, health, crud, send${NC}"
            exit 1
            ;;
    esac
    
    echo -e "\n${GREEN}🎉 Testing completed!${NC}"
}

# Verificar dependencias
if ! command -v curl &> /dev/null; then
    echo -e "${RED}❌ curl is required but not installed${NC}"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}⚠️ jq is not installed. JSON responses will not be formatted${NC}"
fi

main