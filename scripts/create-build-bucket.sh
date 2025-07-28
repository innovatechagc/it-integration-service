#!/bin/bash

# Script para crear bucket de artifacts de Cloud Build
# Uso: ./scripts/create-build-bucket.sh

set -e

PROJECT_ID=$(gcloud config get-value project)
BUCKET_NAME="${PROJECT_ID}-build-artifacts"

echo "🪣 Creando bucket de artifacts: gs://$BUCKET_NAME"

# Crear bucket
gsutil mb -p $PROJECT_ID gs://$BUCKET_NAME

# Configurar lifecycle para limpiar artifacts antiguos (30 días)
cat > /tmp/lifecycle.json << EOF
{
  "lifecycle": {
    "rule": [
      {
        "action": {"type": "Delete"},
        "condition": {"age": 30}
      }
    ]
  }
}
EOF

gsutil lifecycle set /tmp/lifecycle.json gs://$BUCKET_NAME
rm /tmp/lifecycle.json

echo "✅ Bucket creado exitosamente: gs://$BUCKET_NAME"
echo "📋 Para habilitar artifacts, descomenta la sección en cloudbuild.yaml"