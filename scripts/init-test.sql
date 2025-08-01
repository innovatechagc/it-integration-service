-- Inicialización de base de datos para testing
-- Este script se ejecuta automáticamente cuando se inicia el contenedor de PostgreSQL

-- Crear extensiones necesarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Tabla de integraciones de canal
CREATE TABLE IF NOT EXISTS channel_integrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    platform VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    access_token TEXT, -- Encriptado
    webhook_url TEXT,
    status VARCHAR(20) DEFAULT 'active',
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT unique_tenant_platform UNIQUE(tenant_id, platform, provider)
);

-- Tabla de mensajes entrantes (para logs/debug)
CREATE TABLE IF NOT EXISTS inbound_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    platform VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed BOOLEAN DEFAULT FALSE,
    
    INDEX idx_inbound_platform (platform),
    INDEX idx_inbound_received_at (received_at),
    INDEX idx_inbound_processed (processed)
);

-- Tabla de logs de mensajes salientes
CREATE TABLE IF NOT EXISTS outbound_message_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id UUID NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    content JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'queued',
    response JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (channel_id) REFERENCES channel_integrations(id) ON DELETE CASCADE,
    INDEX idx_outbound_channel (channel_id),
    INDEX idx_outbound_status (status),
    INDEX idx_outbound_timestamp (timestamp)
);

-- Tabla de usuarios (para auditoría)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    roles TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de auditoría
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_audit_user (user_id),
    INDEX idx_audit_action (action),
    INDEX idx_audit_created_at (created_at)
);

-- Función para actualizar updated_at automáticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para updated_at
CREATE TRIGGER update_channel_integrations_updated_at 
    BEFORE UPDATE ON channel_integrations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Datos de prueba para desarrollo
INSERT INTO users (id, email, name, roles) VALUES 
    ('550e8400-e29b-41d4-a716-446655440000', 'test@example.com', 'Test User', ARRAY['admin']),
    ('550e8400-e29b-41d4-a716-446655440001', 'dev@example.com', 'Developer', ARRAY['developer'])
ON CONFLICT (email) DO NOTHING;

-- Integraciones de prueba
INSERT INTO channel_integrations (id, tenant_id, platform, provider, webhook_url, config) VALUES 
    ('660e8400-e29b-41d4-a716-446655440000', 'tenant-test-1', 'whatsapp', 'meta', 'http://localhost:8080/webhooks/whatsapp/meta', '{"phone_number_id": "test123", "business_account_id": "test456"}'),
    ('660e8400-e29b-41d4-a716-446655440001', 'tenant-test-1', 'telegram', 'custom', 'http://localhost:8080/webhooks/telegram', '{"bot_token": "test_bot_token"}'),
    ('660e8400-e29b-41d4-a716-446655440002', 'tenant-test-2', 'whatsapp', 'twilio', 'http://localhost:8080/webhooks/whatsapp/twilio', '{"account_sid": "test_sid", "auth_token": "test_token"}')
ON CONFLICT (tenant_id, platform, provider) DO NOTHING;

-- Mensajes de prueba
INSERT INTO inbound_messages (platform, payload, processed) VALUES 
    ('whatsapp', '{"from": "+1234567890", "text": "Hello World", "timestamp": "2024-01-01T10:00:00Z"}', true),
    ('telegram', '{"chat_id": 123456, "text": "Test message", "date": 1704110400}', false);

INSERT INTO outbound_message_logs (channel_id, recipient, content, status) VALUES 
    ('660e8400-e29b-41d4-a716-446655440000', '+1234567890', '{"type": "text", "text": "Welcome message"}', 'sent'),
    ('660e8400-e29b-41d4-a716-446655440001', '123456', '{"type": "text", "text": "Bot response"}', 'queued');

COMMIT;