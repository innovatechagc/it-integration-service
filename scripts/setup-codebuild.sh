#!/bin/bash

# Script para configurar CodeBuild para despliegue en Cloud Run
# Uso: ./scripts/setup-codebuild.sh

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Configurando CodeBuild para despliegue en Cloud Run${NC}"

# Verificar dependencias
command -v aws >/dev/null 2>&1 || { echo -e "${RED}❌ AWS CLI no está instalado${NC}" >&2; exit 1; }
command -v gcloud >/dev/null 2>&1 || { echo -e "${RED}❌ Google Cloud SDK no está instalado${NC}" >&2; exit 1; }

# Variables de configuración (personalizar según tu proyecto)
PROJECT_NAME="microservice-template"
GCP_PROJECT_ID="${GCP_PROJECT_ID:-your-gcp-project-id}"
GCP_REGION="${GCP_REGION:-us-central1}"
AWS_REGION="${AWS_REGION:-us-east-1}"

echo -e "${YELLOW}📋 Configuración:${NC}"
echo "  - Proyecto GCP: $GCP_PROJECT_ID"
echo "  - Región GCP: $GCP_REGION"
echo "  - Región AWS: $AWS_REGION"
echo "  - Nombre del proyecto: $PROJECT_NAME"

# 1. Crear parámetros en AWS Parameter Store
echo -e "${GREEN}📝 Creando parámetros en AWS Parameter Store...${NC}"

aws ssm put-parameter \
    --name "/$PROJECT_NAME/gcp/project-id" \
    --value "$GCP_PROJECT_ID" \
    --type "String" \
    --overwrite \
    --region "$AWS_REGION" || echo "Parámetro project-id ya existe"

aws ssm put-parameter \
    --name "/$PROJECT_NAME/gcp/region" \
    --value "$GCP_REGION" \
    --type "String" \
    --overwrite \
    --region "$AWS_REGION" || echo "Parámetro region ya existe"

aws ssm put-parameter \
    --name "/$PROJECT_NAME/service-name" \
    --value "$PROJECT_NAME" \
    --type "String" \
    --overwrite \
    --region "$AWS_REGION" || echo "Parámetro service-name ya existe"

# 2. Crear service account en GCP y obtener la clave
echo -e "${GREEN}🔑 Configurando Service Account en GCP...${NC}"

SERVICE_ACCOUNT_NAME="codebuild-deployer"
SERVICE_ACCOUNT_EMAIL="$SERVICE_ACCOUNT_NAME@$GCP_PROJECT_ID.iam.gserviceaccount.com"

# Crear service account si no existe
gcloud iam service-accounts create $SERVICE_ACCOUNT_NAME \
    --display-name="CodeBuild Deployer" \
    --description="Service account para despliegues desde AWS CodeBuild" \
    --project=$GCP_PROJECT_ID 2>/dev/null || echo "Service account ya existe"

# Asignar roles necesarios
gcloud projects add-iam-policy-binding $GCP_PROJECT_ID \
    --member="serviceAccount:$SERVICE_ACCOUNT_EMAIL" \
    --role="roles/run.admin"

gcloud projects add-iam-policy-binding $GCP_PROJECT_ID \
    --member="serviceAccount:$SERVICE_ACCOUNT_EMAIL" \
    --role="roles/storage.admin"

gcloud projects add-iam-policy-binding $GCP_PROJECT_ID \
    --member="serviceAccount:$SERVICE_ACCOUNT_EMAIL" \
    --role="roles/iam.serviceAccountUser"

# Crear y descargar clave del service account
KEY_FILE="/tmp/gcp-service-account-key.json"
gcloud iam service-accounts keys create $KEY_FILE \
    --iam-account=$SERVICE_ACCOUNT_EMAIL \
    --project=$GCP_PROJECT_ID

# Codificar la clave en base64 y guardarla en Secrets Manager
KEY_BASE64=$(base64 -w 0 $KEY_FILE)

echo -e "${GREEN}🔐 Guardando credenciales en AWS Secrets Manager...${NC}"

aws secretsmanager create-secret \
    --name "$PROJECT_NAME/gcp-service-account" \
    --description "GCP Service Account key for CodeBuild deployment" \
    --secret-string "$KEY_BASE64" \
    --region "$AWS_REGION" 2>/dev/null || \
aws secretsmanager update-secret \
    --secret-id "$PROJECT_NAME/gcp-service-account" \
    --secret-string "$KEY_BASE64" \
    --region "$AWS_REGION"

# Limpiar archivo temporal
rm -f $KEY_FILE

# 3. Crear rol IAM para CodeBuild
echo -e "${GREEN}👤 Creando rol IAM para CodeBuild...${NC}"

ROLE_NAME="CodeBuildServiceRole-$PROJECT_NAME"
TRUST_POLICY='{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codebuild.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}'

# Crear rol
aws iam create-role \
    --role-name "$ROLE_NAME" \
    --assume-role-policy-document "$TRUST_POLICY" \
    --region "$AWS_REGION" 2>/dev/null || echo "Rol ya existe"

# Política personalizada para el proyecto
POLICY_DOCUMENT='{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ssm:GetParameter",
        "ssm:GetParameters"
      ],
      "Resource": [
        "arn:aws:ssm:*:*:parameter/'$PROJECT_NAME'/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": [
        "arn:aws:secretsmanager:*:*:secret:'$PROJECT_NAME'/*"
      ]
    }
  ]
}'

POLICY_NAME="CodeBuildPolicy-$PROJECT_NAME"

aws iam create-policy \
    --policy-name "$POLICY_NAME" \
    --policy-document "$POLICY_DOCUMENT" \
    --region "$AWS_REGION" 2>/dev/null || echo "Política ya existe"

# Obtener ARN de la cuenta
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
POLICY_ARN="arn:aws:iam::$ACCOUNT_ID:policy/$POLICY_NAME"

# Adjuntar políticas al rol
aws iam attach-role-policy \
    --role-name "$ROLE_NAME" \
    --policy-arn "$POLICY_ARN"

aws iam attach-role-policy \
    --role-name "$ROLE_NAME" \
    --policy-arn "arn:aws:iam::aws:policy/AWSCodeBuildDeveloperAccess"

echo -e "${GREEN}✅ Configuración completada!${NC}"
echo -e "${YELLOW}📋 Próximos pasos:${NC}"
echo "1. Crear el proyecto CodeBuild en la consola de AWS"
echo "2. Usar el rol: arn:aws:iam::$ACCOUNT_ID:role/$ROLE_NAME"
echo "3. Configurar el webhook para GitHub/GitLab si es necesario"
echo "4. Ejecutar un build de prueba"

echo -e "${GREEN}🔧 Comandos útiles:${NC}"
echo "- Ver logs: aws logs describe-log-groups --log-group-name-prefix /aws/codebuild/$PROJECT_NAME"
echo "- Ejecutar build: aws codebuild start-build --project-name $PROJECT_NAME"