#!/bin/bash

# Script para debuggear problemas de despliegue en Cloud Run
# Uso: ./scripts/debug-deployment.sh [service-name] [region]

set -e

SERVICE_NAME=${1:-"it-integration-service"}
REGION=${2:-"us-east1"}

echo "🔍 Debugging despliegue de Cloud Run..."
echo "Servicio: $SERVICE_NAME"
echo "Región: $REGION"
echo ""

# 1. Verificar estado del servicio
echo "📋 Estado del servicio:"
gcloud run services describe $SERVICE_NAME --region=$REGION --format="table(
  metadata.name,
  status.conditions[0].type,
  status.conditions[0].status,
  status.conditions[0].reason
)" || echo "❌ Servicio no encontrado"

echo ""

# 2. Ver revisiones
echo "📦 Revisiones del servicio:"
gcloud run revisions list --service=$SERVICE_NAME --region=$REGION --limit=3 --format="table(
  metadata.name,
  status.conditions[0].type,
  status.conditions[0].status,
  status.conditions[0].reason,
  spec.containers[0].image
)" || echo "❌ No se pudieron obtener revisiones"

echo ""

# 3. Ver logs recientes
echo "📝 Logs recientes (últimos 10 minutos):"
gcloud logging read "
  resource.type=cloud_run_revision AND 
  resource.labels.service_name=$SERVICE_NAME AND
  timestamp >= \"$(date -u -d '10 minutes ago' '+%Y-%m-%dT%H:%M:%SZ')\"
" --limit=20 --format="table(
  timestamp,
  severity,
  textPayload
)" || echo "❌ No se pudieron obtener logs"

echo ""

# 4. Verificar secretos
echo "🔐 Verificando secretos necesarios:"
SECRETS=("it-chatbot-jwt-password" "it-chatbot-db-password")

for secret in "${SECRETS[@]}"; do
  if gcloud secrets describe $secret >/dev/null 2>&1; then
    echo "✅ $secret - existe"
  else
    echo "❌ $secret - no encontrado"
  fi
done

echo ""

# 5. Verificar imagen en Container Registry
echo "🐳 Verificando imagen en Container Registry:"
PROJECT_ID=$(gcloud config get-value project)
IMAGE_NAME="gcr.io/$PROJECT_ID/$SERVICE_NAME"

if gcloud container images list --repository=$IMAGE_NAME >/dev/null 2>&1; then
  echo "✅ Imagen existe en GCR"
  echo "📋 Tags disponibles:"
  gcloud container images list-tags $IMAGE_NAME --limit=5 --format="table(tags,timestamp)"
else
  echo "❌ Imagen no encontrada en GCR"
fi

echo ""

# 6. Probar conectividad si el servicio está funcionando
echo "🌐 Probando conectividad:"
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --region=$REGION --format="value(status.url)" 2>/dev/null)

if [ ! -z "$SERVICE_URL" ]; then
  echo "URL del servicio: $SERVICE_URL"
  
  echo "Probando health check..."
  if curl -f "$SERVICE_URL/api/v1/health" --max-time 10 --silent; then
    echo "✅ Health check exitoso"
  else
    echo "❌ Health check falló"
    echo "Probando conectividad básica..."
    if curl -I "$SERVICE_URL" --max-time 10 --silent; then
      echo "✅ Servicio responde pero health check falló"
    else
      echo "❌ Servicio no responde"
    fi
  fi
else
  echo "❌ No se pudo obtener URL del servicio"
fi

echo ""
echo "🔧 Comandos útiles para más debugging:"
echo "# Ver logs en tiempo real:"
echo "gcloud logging tail \"resource.type=cloud_run_revision AND resource.labels.service_name=$SERVICE_NAME\""
echo ""
echo "# Ver configuración completa del servicio:"
echo "gcloud run services describe $SERVICE_NAME --region=$REGION"
echo ""
echo "# Ver detalles de la revisión más reciente:"
echo "gcloud run revisions describe \$(gcloud run revisions list --service=$SERVICE_NAME --region=$REGION --limit=1 --format=\"value(metadata.name)\") --region=$REGION"