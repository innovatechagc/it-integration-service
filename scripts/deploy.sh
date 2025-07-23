#!/bin/bash

# Script de despliegue para GCP Cloud Run
set -e

# Variables
PROJECT_ID=${PROJECT_ID:-"your-project-id"}
SERVICE_NAME=${SERVICE_NAME:-"microservice-template"}
REGION=${REGION:-"us-central1"}
ENVIRONMENT=${ENVIRONMENT:-"staging"}

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Iniciando despliegue para ${ENVIRONMENT}${NC}"

# Verificar que gcloud esté configurado
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}❌ gcloud CLI no está instalado${NC}"
    exit 1
fi

# Verificar autenticación
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
    echo -e "${RED}❌ No hay cuentas autenticadas en gcloud${NC}"
    echo "Ejecuta: gcloud auth login"
    exit 1
fi

# Configurar proyecto
echo -e "${YELLOW}📋 Configurando proyecto: ${PROJECT_ID}${NC}"
gcloud config set project ${PROJECT_ID}

# Habilitar APIs necesarias
echo -e "${YELLOW}🔧 Habilitando APIs necesarias${NC}"
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable secretmanager.googleapis.com

# Construir imagen
echo -e "${YELLOW}🏗️ Construyendo imagen Docker${NC}"
gcloud builds submit --tag gcr.io/${PROJECT_ID}/${SERVICE_NAME}:latest

# Crear secretos si no existen (solo para staging/desarrollo)
if [ "$ENVIRONMENT" = "staging" ]; then
    echo -e "${YELLOW}🔐 Verificando secretos para staging${NC}"
    
    # JWT Secret
    if ! gcloud secrets describe jwt-secret-staging &>/dev/null; then
        echo "Creando jwt-secret-staging..."
        echo -n "$(openssl rand -base64 32)" | gcloud secrets create jwt-secret-staging --data-file=-
    fi
    
    # DB Password
    if ! gcloud secrets describe db-password-staging &>/dev/null; then
        echo "Creando db-password-staging..."
        echo -n "staging-db-password" | gcloud secrets create db-password-staging --data-file=-
    fi
    
    # Vault Token
    if ! gcloud secrets describe vault-token-staging &>/dev/null; then
        echo "Creando vault-token-staging..."
        echo -n "staging-vault-token" | gcloud secrets create vault-token-staging --data-file=-
    fi
    
    # External API Key
    if ! gcloud secrets describe external-api-key-staging &>/dev/null; then
        echo "Creando external-api-key-staging..."
        echo -n "staging-api-key" | gcloud secrets create external-api-key-staging --data-file=-
    fi
fi

# Desplegar según el entorno
if [ "$ENVIRONMENT" = "production" ]; then
    echo -e "${YELLOW}🚀 Desplegando a producción${NC}"
    gcloud run services replace deploy/cloudrun-production.yaml --region=${REGION}
else
    echo -e "${YELLOW}🚀 Desplegando a staging${NC}"
    gcloud run services replace deploy/cloudrun-staging.yaml --region=${REGION}
fi

# Obtener URL del servicio
SERVICE_URL=$(gcloud run services describe ${SERVICE_NAME}${ENVIRONMENT:+-$ENVIRONMENT} --region=${REGION} --format="value(status.url)")

echo -e "${GREEN}✅ Despliegue completado${NC}"
echo -e "${GREEN}🌐 URL del servicio: ${SERVICE_URL}${NC}"

# Verificar health check
echo -e "${YELLOW}🏥 Verificando health check${NC}"
if curl -f "${SERVICE_URL}/api/v1/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Health check exitoso${NC}"
else
    echo -e "${RED}❌ Health check falló${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 Despliegue completado exitosamente${NC}"