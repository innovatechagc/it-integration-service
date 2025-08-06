#!/bin/bash

# Cargar variables de entorno
source .env.local

# Mostrar configuración de DB para debug
echo "=== Configuración de Base de Datos ==="
echo "DB_HOST: $DB_HOST"
echo "DB_PORT: $DB_PORT"
echo "DB_NAME: $DB_NAME"
echo "DB_USER: $DB_USER"
echo "DB_SSL_MODE: $DB_SSL_MODE"
echo "======================================"

# Probar conexión primero
echo "Probando conexión a la base de datos..."
PGPASSWORD="$DB_PASSWORD" psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 'Conexión exitosa' as status;" || {
    echo "❌ Error: No se puede conectar a la base de datos"
    exit 1
}

echo "✅ Conexión a la base de datos exitosa"
echo "Iniciando servicio..."

# Ejecutar el servicio
export DB_HOST DB_PORT DB_NAME DB_USER DB_PASSWORD DB_SSL_MODE
export ENVIRONMENT LOG_LEVEL
export MESSAGING_SERVICE_URL
export PORT=8081

echo "Iniciando servicio en puerto $PORT..."
./bin/integration-service