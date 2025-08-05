#!/bin/bash

# Script para probar el envío de mensajes reales
# Reemplaza los valores con datos reales

echo "🧪 Probando envío de mensajes..."

# Enviar mensaje a Telegram
echo "📱 Enviando mensaje a Telegram..."
curl -X POST "http://localhost:8080/api/v1/integrations/send" \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "telegram",
    "recipient_id": "YOUR_TELEGRAM_USER_ID",
    "message": "¡Hola! Este es un mensaje de prueba desde la API.",
    "message_type": "text"
  }'

echo -e "\n"

# Enviar mensaje a WhatsApp
echo "📱 Enviando mensaje a WhatsApp..."
curl -X POST "http://localhost:8080/api/v1/integrations/send" \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "whatsapp",
    "recipient_id": "YOUR_WHATSAPP_PHONE_NUMBER",
    "message": "¡Hola! Este es un mensaje de prueba desde la API.",
    "message_type": "text"
  }'

echo -e "\n"

# Verificar mensajes entrantes
echo "📥 Verificando mensajes entrantes..."
curl -s "http://localhost:8080/api/v1/integrations/messages/inbound?tenant_id=tenant1" | jq '.'

echo -e "\n"

# Verificar canales configurados
echo "🔗 Verificando canales configurados..."
curl -s "http://localhost:8080/api/v1/integrations/channels?tenant_id=tenant1" | jq '.data[].platform' 