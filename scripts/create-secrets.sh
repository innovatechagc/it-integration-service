#!/bin/bash

# Script para crear los secretos faltantes en Google Secret Manager
# Uso: ./scripts/create-secrets.sh

set -e

PROJECT_ID=$(gcloud config get-value project)

echo "🔐 Creando secretos faltantes en Google Secret Manager..."
echo "Proyecto: $PROJECT_ID"

# Crear secretos para producción
echo "📝 Creando secretos de producción..."

# Vault token (puedes usar un token dummy si no usas Vault)
echo "dummy-vault-token-prod" | gcloud secrets create vault-token-prod --data-file=-

# External API key (puedes usar una key dummy)
echo "dummy-external-api-key-prod" | gcloud secrets create external-api-key-prod --data-file=-

# Crear secretos para staging
echo "📝 Creando secretos de staging..."

# JWT para staging (puedes usar el mismo que producción o uno diferente)
echo "staging-jwt-secret-$(date +%s)" | gcloud secrets create jwt-secret-staging --data-file=-

# DB password para staging
echo "staging-db-password-$(date +%s)" | gcloud secrets create db-password-staging --data-file=-

# Vault token para staging
echo "dummy-vault-token-staging" | gcloud secrets create vault-token-staging --data-file=-

# External API key para staging
echo "dummy-external-api-key-staging" | gcloud secrets create external-api-key-staging --data-file=-

echo "✅ Secretos creados exitosamente!"
echo ""
echo "📋 Secretos disponibles:"
gcloud secrets list --filter="name:vault-token OR name:external-api-key OR name:jwt-secret-staging OR name:db-password-staging"

echo ""
echo "⚠️  IMPORTANTE: Actualiza los valores de los secretos con datos reales:"
echo "gcloud secrets versions add vault-token-prod --data-file=vault-token.txt"
echo "gcloud secrets versions add external-api-key-prod --data-file=api-key.txt"