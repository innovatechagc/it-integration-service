-- Integration Service Database Schema

-- Channel Integrations table
CREATE TABLE IF NOT EXISTS channel_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    platform VARCHAR(50) NOT NULL CHECK (platform IN ('whatsapp', 'messenger', 'instagram', 'telegram', 'webchat')),
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('meta', 'twilio', '360dialog', 'custom')),
    access_token TEXT NOT NULL, -- Encrypted
    webhook_url VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled', 'error')),
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for channel_integrations
CREATE INDEX IF NOT EXISTS idx_channel_integrations_tenant_id ON channel_integrations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_channel_integrations_platform ON channel_integrations(platform);
CREATE INDEX IF NOT EXISTS idx_channel_integrations_status ON channel_integrations(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_integrations_tenant_platform ON channel_integrations(tenant_id, platform, provider);

-- Inbound Messages table (for logging/debugging)
CREATE TABLE IF NOT EXISTS inbound_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed BOOLEAN DEFAULT FALSE
);

-- Indexes for inbound_messages
CREATE INDEX IF NOT EXISTS idx_inbound_messages_platform ON inbound_messages(platform);
CREATE INDEX IF NOT EXISTS idx_inbound_messages_processed ON inbound_messages(processed);
CREATE INDEX IF NOT EXISTS idx_inbound_messages_received_at ON inbound_messages(received_at);

-- Outbound Message Logs table
CREATE TABLE IF NOT EXISTS outbound_message_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channel_integrations(id) ON DELETE CASCADE,
    recipient VARCHAR(255) NOT NULL,
    content JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK (status IN ('sent', 'failed', 'queued')),
    response JSONB DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for outbound_message_logs
CREATE INDEX IF NOT EXISTS idx_outbound_logs_channel_id ON outbound_message_logs(channel_id);
CREATE INDEX IF NOT EXISTS idx_outbound_logs_status ON outbound_message_logs(status);
CREATE INDEX IF NOT EXISTS idx_outbound_logs_timestamp ON outbound_message_logs(timestamp);

-- Update trigger for channel_integrations
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_channel_integrations_updated_at 
    BEFORE UPDATE ON channel_integrations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Sample data for development (optional)
INSERT INTO channel_integrations (tenant_id, platform, provider, access_token, webhook_url, config) 
VALUES 
    ('tenant-1', 'whatsapp', 'meta', 'encrypted-token-1', 'https://api.example.com/webhooks/whatsapp', '{"phone_number_id": "123456789", "business_id": "987654321"}'),
    ('tenant-1', 'messenger', 'meta', 'encrypted-token-2', 'https://api.example.com/webhooks/messenger', '{"page_id": "123456789"}'),
    ('tenant-2', 'telegram', 'custom', 'encrypted-token-3', 'https://api.example.com/webhooks/telegram', '{"bot_token": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"}')
ON CONFLICT DO NOTHING;