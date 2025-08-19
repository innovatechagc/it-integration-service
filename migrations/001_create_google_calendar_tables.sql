-- Migración para crear tablas de Google Calendar
-- Ejecutar: psql -d your_database -f 001_create_google_calendar_tables.sql

-- Tabla para integraciones de Google Calendar
CREATE TABLE IF NOT EXISTS google_calendar_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    channel_id VARCHAR(255) UNIQUE NOT NULL,
    calendar_type VARCHAR(50) NOT NULL CHECK (calendar_type IN ('personal', 'work', 'shared')),
    calendar_id VARCHAR(255) NOT NULL,
    calendar_name VARCHAR(500) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_expiry TIMESTAMP WITH TIME ZONE NOT NULL,
    webhook_channel VARCHAR(255),
    webhook_resource TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'disabled' CHECK (status IN ('active', 'disabled', 'error')),
    config JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Tabla para eventos de calendario
CREATE TABLE IF NOT EXISTS calendar_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    channel_id VARCHAR(255) NOT NULL,
    google_id VARCHAR(255) NOT NULL,
    calendar_id VARCHAR(255) NOT NULL,
    summary VARCHAR(500) NOT NULL,
    description TEXT,
    location VARCHAR(1000),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    all_day BOOLEAN NOT NULL DEFAULT FALSE,
    attendees JSONB,
    recurrence JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'tentative', 'cancelled')),
    visibility VARCHAR(50) NOT NULL DEFAULT 'default' CHECK (visibility IN ('default', 'public', 'private')),
    reminders JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Índices para google_calendar_integrations
CREATE INDEX IF NOT EXISTS idx_google_calendar_integrations_tenant_id ON google_calendar_integrations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_google_calendar_integrations_channel_id ON google_calendar_integrations(channel_id);
CREATE INDEX IF NOT EXISTS idx_google_calendar_integrations_calendar_type ON google_calendar_integrations(calendar_type);
CREATE INDEX IF NOT EXISTS idx_google_calendar_integrations_status ON google_calendar_integrations(status);
CREATE INDEX IF NOT EXISTS idx_google_calendar_integrations_token_expiry ON google_calendar_integrations(token_expiry);
CREATE INDEX IF NOT EXISTS idx_google_calendar_integrations_created_at ON google_calendar_integrations(created_at);

-- Índices para calendar_events
CREATE INDEX IF NOT EXISTS idx_calendar_events_tenant_id ON calendar_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_channel_id ON calendar_events(channel_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_google_id ON calendar_events(google_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_calendar_id ON calendar_events(calendar_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_start_time ON calendar_events(start_time);
CREATE INDEX IF NOT EXISTS idx_calendar_events_end_time ON calendar_events(end_time);
CREATE INDEX IF NOT EXISTS idx_calendar_events_status ON calendar_events(status);
CREATE INDEX IF NOT EXISTS idx_calendar_events_created_at ON calendar_events(created_at);

-- Índice compuesto para búsquedas por canal y rango de fechas
CREATE INDEX IF NOT EXISTS idx_calendar_events_channel_date_range ON calendar_events(channel_id, start_time, end_time);

-- Índice para búsquedas por tenant y fecha
CREATE INDEX IF NOT EXISTS idx_calendar_events_tenant_date ON calendar_events(tenant_id, start_time DESC);

-- Índice GIN para búsquedas en campos JSON
CREATE INDEX IF NOT EXISTS idx_calendar_events_attendees_gin ON calendar_events USING GIN (attendees);
CREATE INDEX IF NOT EXISTS idx_calendar_events_recurrence_gin ON calendar_events USING GIN (recurrence);
CREATE INDEX IF NOT EXISTS idx_calendar_events_reminders_gin ON calendar_events USING GIN (reminders);

-- Restricciones de integridad referencial
ALTER TABLE calendar_events 
ADD CONSTRAINT fk_calendar_events_channel_id 
FOREIGN KEY (channel_id) REFERENCES google_calendar_integrations(channel_id) ON DELETE CASCADE;

-- Trigger para actualizar updated_at automáticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Aplicar trigger a google_calendar_integrations
CREATE TRIGGER update_google_calendar_integrations_updated_at 
    BEFORE UPDATE ON google_calendar_integrations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Aplicar trigger a calendar_events
CREATE TRIGGER update_calendar_events_updated_at 
    BEFORE UPDATE ON calendar_events 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comentarios para documentación
COMMENT ON TABLE google_calendar_integrations IS 'Almacena las integraciones de Google Calendar por tenant';
COMMENT ON TABLE calendar_events IS 'Almacena los eventos de Google Calendar sincronizados';

COMMENT ON COLUMN google_calendar_integrations.calendar_type IS 'Tipo de calendario: personal, work, shared';
COMMENT ON COLUMN google_calendar_integrations.access_token IS 'Token de acceso encriptado';
COMMENT ON COLUMN google_calendar_integrations.refresh_token IS 'Token de refresh encriptado';
COMMENT ON COLUMN google_calendar_integrations.webhook_channel IS 'ID del canal de webhook de Google';
COMMENT ON COLUMN google_calendar_integrations.webhook_resource IS 'URI del recurso monitoreado por el webhook';

COMMENT ON COLUMN calendar_events.google_id IS 'ID único del evento en Google Calendar';
COMMENT ON COLUMN calendar_events.attendees IS 'Array JSON de asistentes al evento';
COMMENT ON COLUMN calendar_events.recurrence IS 'Configuración de recurrencia del evento';
COMMENT ON COLUMN calendar_events.reminders IS 'Array JSON de recordatorios del evento';

-- Vistas útiles para consultas comunes
CREATE OR REPLACE VIEW active_google_calendar_integrations AS
SELECT 
    gci.*,
    COUNT(ce.id) as total_events
FROM google_calendar_integrations gci
LEFT JOIN calendar_events ce ON gci.channel_id = ce.channel_id
WHERE gci.status = 'active'
GROUP BY gci.id, gci.tenant_id, gci.channel_id, gci.calendar_type, gci.calendar_id, 
         gci.calendar_name, gci.access_token, gci.refresh_token, gci.token_expiry, 
         gci.webhook_channel, gci.webhook_resource, gci.status, gci.config, 
         gci.created_at, gci.updated_at;

CREATE OR REPLACE VIEW upcoming_calendar_events AS
SELECT 
    ce.*,
    gci.calendar_name,
    gci.calendar_type
FROM calendar_events ce
JOIN google_calendar_integrations gci ON ce.channel_id = gci.channel_id
WHERE ce.start_time >= NOW()
  AND ce.status = 'confirmed'
ORDER BY ce.start_time ASC;

-- Función para limpiar eventos antiguos (opcional)
CREATE OR REPLACE FUNCTION cleanup_old_calendar_events(days_to_keep INTEGER DEFAULT 365)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM calendar_events 
    WHERE end_time < NOW() - INTERVAL '1 day' * days_to_keep
      AND status = 'cancelled';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Función para obtener estadísticas de eventos por tenant
CREATE OR REPLACE FUNCTION get_calendar_events_stats(p_tenant_id VARCHAR)
RETURNS TABLE(
    total_events BIGINT,
    upcoming_events BIGINT,
    past_events BIGINT,
    cancelled_events BIGINT,
    total_integrations BIGINT,
    active_integrations BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(ce.id)::BIGINT as total_events,
        COUNT(CASE WHEN ce.start_time >= NOW() THEN 1 END)::BIGINT as upcoming_events,
        COUNT(CASE WHEN ce.end_time < NOW() THEN 1 END)::BIGINT as past_events,
        COUNT(CASE WHEN ce.status = 'cancelled' THEN 1 END)::BIGINT as cancelled_events,
        COUNT(DISTINCT gci.channel_id)::BIGINT as total_integrations,
        COUNT(DISTINCT CASE WHEN gci.status = 'active' THEN gci.channel_id END)::BIGINT as active_integrations
    FROM google_calendar_integrations gci
    LEFT JOIN calendar_events ce ON gci.channel_id = ce.channel_id
    WHERE gci.tenant_id = p_tenant_id;
END;
$$ LANGUAGE plpgsql;
