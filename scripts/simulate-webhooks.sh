#!/bin/bash

echo "🧪 Simulando webhooks y probando el flujo completo..."

# URL base del integration service
BASE_URL="http://localhost:8080/api/v1/integrations"

echo "📋 Paso 1: Verificar que el servicio esté funcionando..."
curl -s "$BASE_URL/../health" | jq '.'

echo -e "\n📋 Paso 2: Verificar canales configurados..."
curl -s "$BASE_URL/channels?tenant_id=tenant1" | jq '.data[] | {platform, status, id}'

echo -e "\n📋 Paso 3: Simular webhook de Telegram..."
curl -X POST "$BASE_URL/webhooks/telegram" \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 123456789,
    "message": {
      "message_id": 1,
      "from": {
        "id": 123456789,
        "first_name": "Usuario",
        "last_name": "Prueba",
        "username": "usuario_prueba"
      },
      "chat": {
        "id": 123456789,
        "type": "private"
      },
      "date": 1691251200,
      "text": "Hola, este es un mensaje de prueba desde Telegram"
    }
  }'

echo -e "\n📋 Paso 4: Simular webhook de WhatsApp..."
curl -X POST "$BASE_URL/webhooks/whatsapp" \
  -H "Content-Type: application/json" \
  -d '{
    "object": "whatsapp_business_account",
    "entry": [
      {
        "id": "297284031622102",
        "changes": [
          {
            "value": {
              "messaging_product": "whatsapp",
              "metadata": {
                "display_phone_number": "573188827146",
                "phone_number_id": "764957900026580"
              },
              "contacts": [
                {
                  "profile": {
                    "name": "Usuario WhatsApp"
                  },
                  "wa_id": "573188827146"
                }
              ],
              "messages": [
                {
                  "from": "573188827146",
                  "id": "wamid.HBgMNTczMTg4ODI3MTQ2FQIAEhgUMjAyNTA4MDUyMDAwMDAwMDAwMDAwAA==",
                  "timestamp": "1691251200",
                  "text": {
                    "body": "Hola, este es un mensaje de prueba desde WhatsApp"
                  },
                  "type": "text"
                }
              ]
            },
            "field": "messages"
          }
        ]
      }
    ]
  }'

echo -e "\n📋 Paso 5: Verificar mensajes entrantes..."
curl -s "$BASE_URL/messages/inbound?tenant_id=tenant1" | jq '.'

echo -e "\n📋 Paso 6: Enviar mensaje de respuesta a Telegram..."
curl -X POST "$BASE_URL/send" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "23d8d953-a571-45df-95de-f5aecb5b0b93",
    "recipient": "123456789",
    "content": {
      "type": "text",
      "text": "¡Hola! Gracias por tu mensaje. Este es un mensaje de respuesta automática."
    }
  }' | jq '.'

echo -e "\n📋 Paso 7: Enviar mensaje de respuesta a WhatsApp..."
curl -X POST "$BASE_URL/send" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "42ef8faa-571c-4fe8-9fbe-7531ad05a72d",
    "recipient": "573188827146",
    "content": {
      "type": "text",
      "text": "¡Hola! Gracias por tu mensaje. Este es un mensaje de respuesta automática."
    }
  }' | jq '.'

echo -e "\n📋 Paso 8: Verificar historial de chat..."
echo "Historial de Telegram:"
curl -s "$BASE_URL/chat/telegram/123456789?tenant_id=tenant1" | jq '.'

echo -e "\nHistorial de WhatsApp:"
curl -s "$BASE_URL/chat/whatsapp/573188827146?tenant_id=tenant1" | jq '.'

echo -e "\n✅ Prueba completada!" 