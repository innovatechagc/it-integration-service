-- Script para actualizar URLs de webhook con ngrok
-- Reemplaza YOUR_NGROK_URL con tu URL de ngrok

-- Actualizar webhook de Telegram
UPDATE channel_integrations 
SET webhook_url = 'https://your-ngrok-url.ngrok.io/api/v1/integrations/webhooks/telegram'
WHERE platform = 'telegram' AND tenant_id = 'tenant1';

-- Actualizar webhook de WhatsApp
UPDATE channel_integrations 
SET webhook_url = 'https://your-ngrok-url.ngrok.io/api/v1/integrations/webhooks/whatsapp'
WHERE platform = 'whatsapp' AND tenant_id = 'tenant1';

-- Verificar los cambios
SELECT platform, webhook_url FROM channel_integrations WHERE tenant_id = 'tenant1'; 