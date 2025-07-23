#!/bin/bash

# Script para inicializar Vault con secretos de desarrollo
echo "Inicializando Vault con secretos de desarrollo..."

# Esperar a que Vault esté disponible
sleep 5

# Configurar secretos de ejemplo
vault kv put secret/microservice \
    db_password="postgres" \
    api_key="dev-api-key" \
    jwt_secret="dev-jwt-secret"

vault kv put secret/microservice-test \
    db_password="postgres_test" \
    api_key="test-api-key" \
    jwt_secret="test-jwt-secret"

echo "Secretos de desarrollo configurados en Vault"