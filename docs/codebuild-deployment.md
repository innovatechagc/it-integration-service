# Despliegue a Google Cloud Run

Esta guía explica cómo configurar despliegues automáticos a Google Cloud Run usando tanto AWS CodeBuild como Google Cloud Build.

## Opciones de Despliegue

### Opción 1: Google Cloud Build (Recomendado)
```
GitHub/GitLab → Google Cloud Build → Google Container Registry → Google Cloud Run
```

### Opción 2: AWS CodeBuild
```
GitHub/GitLab → AWS CodeBuild → Google Container Registry → Google Cloud Run
```

## Google Cloud Build (cloudbuild.yaml)

### Configuración Rápida

1. **Habilitar APIs necesarias:**
```bash
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable containerregistry.googleapis.com
```

2. **Configurar permisos para Cloud Build:**
```bash
# Obtener el número del proyecto
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format="value(projectNumber)")

# Dar permisos a Cloud Build para desplegar en Cloud Run
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$PROJECT_NUMBER@cloudbuild.gserviceaccount.com" \
    --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$PROJECT_NUMBER@cloudbuild.gserviceaccount.com" \
    --role="roles/iam.serviceAccountUser"
```

3. **Crear trigger desde repositorio:**
```bash
# Para GitHub
gcloud builds triggers create github \
    --repo-name=tu-repositorio \
    --repo-owner=tu-usuario \
    --branch-pattern=".*" \
    --build-config=cloudbuild.yaml

# Para GitLab o otros
gcloud builds triggers create cloud-source-repositories \
    --repo=tu-repositorio \
    --branch-pattern=".*" \
    --build-config=cloudbuild.yaml
```

### Características del cloudbuild.yaml

- **Tests automáticos** antes del despliegue
- **Build multi-arquitectura** optimizado
- **Despliegue condicional** según rama (main → producción, otras → staging)
- **Health checks** automáticos post-despliegue
- **Logging centralizado** en Cloud Logging
- **Artefactos** guardados en Cloud Storage

---

## AWS CodeBuild (buildspec.yml / codebuild.yaml)

## Configuración Inicial

### 1. Prerrequisitos

- AWS CLI configurado con permisos administrativos
- Google Cloud SDK instalado y configurado
- Proyecto en Google Cloud Platform
- Repositorio de código en GitHub/GitLab

### 2. Ejecutar Script de Configuración

```bash
# Configurar variables de entorno
export GCP_PROJECT_ID="tu-proyecto-gcp"
export GCP_REGION="us-central1"
export AWS_REGION="us-east-1"

# Ejecutar script de configuración
./scripts/setup-codebuild.sh
```

Este script:
- Crea parámetros en AWS Parameter Store
- Configura Service Account en GCP con permisos necesarios
- Guarda credenciales en AWS Secrets Manager
- Crea rol IAM para CodeBuild

### 3. Crear Proyecto CodeBuild

En la consola de AWS CodeBuild:

1. **Nombre del proyecto**: `microservice-template`
2. **Fuente**: GitHub/GitLab (configurar webhook)
3. **Entorno**:
   - Imagen: `aws/codebuild/amazonlinux2-x86_64-standard:4.0`
   - Tipo de compilación: Linux
   - Rol de servicio: `CodeBuildServiceRole-microservice-template`
4. **Buildspec**: Usar `buildspec.yml` del repositorio

## Configuración del buildspec.yml

El archivo `buildspec.yml` maneja:

### Variables de Entorno

```yaml
env:
  variables:
    GO_VERSION: "1.21"
    DOCKER_BUILDKIT: "1"
  parameter-store:
    PROJECT_ID: "/microservice-template/gcp/project-id"
    REGION: "/microservice-template/gcp/region"
    SERVICE_NAME: "/microservice-template/service-name"
  secrets-manager:
    GCP_SERVICE_ACCOUNT_KEY: "microservice-template/gcp-service-account"
```

### Fases del Build

1. **Install**: Instala Go, Docker y Google Cloud SDK
2. **Pre-build**: Configura autenticación con GCP y prepara variables
3. **Build**: Ejecuta tests, construye imagen Docker y la sube a GCR
4. **Post-build**: Despliega a Cloud Run según la rama

## Estrategia de Despliegue

### Ramas y Entornos

- **main/master**: Despliega a producción usando `deploy/cloudrun-production.yaml`
- **Otras ramas**: Despliega a staging usando `deploy/cloudrun-staging.yaml`

### Tags de Imagen

- Cada commit genera una imagen con tag del hash del commit (7 caracteres)
- También se actualiza el tag `latest`
- Formato: `gcr.io/PROJECT_ID/SERVICE_NAME:COMMIT_HASH`

## Configuración de Cloud Run

### Producción (`deploy/cloudrun-production.yaml`)

- **Escalado**: 2-100 instancias
- **Recursos**: 2 CPU, 1GB RAM
- **Concurrencia**: 100 requests por instancia
- **Timeout**: 300 segundos

### Staging (`deploy/cloudrun-staging.yaml`)

- **Escalado**: 1-10 instancias
- **Recursos**: 1 CPU, 512MB RAM
- **Concurrencia**: 100 requests por instancia
- **Timeout**: 300 segundos

## Secretos y Variables

### Secretos en Google Secret Manager

Los siguientes secretos deben existir en Google Secret Manager:

**Producción:**
- `jwt-secret-prod`
- `db-password-prod`
- `vault-token-prod`
- `external-api-key-prod`

**Staging:**
- `jwt-secret-staging`
- `db-password-staging`
- `vault-token-staging`
- `external-api-key-staging`

### Crear Secretos

```bash
# Ejemplo para crear secretos
gcloud secrets create jwt-secret-prod --data-file=jwt-secret.txt
gcloud secrets create db-password-prod --data-file=db-password.txt
```

## Monitoreo y Logs

### Ver Logs de CodeBuild

```bash
# Listar grupos de logs
aws logs describe-log-groups --log-group-name-prefix /aws/codebuild/microservice-template

# Ver logs específicos
aws logs get-log-events --log-group-name /aws/codebuild/microservice-template --log-stream-name <stream-name>
```

### Ver Logs de Cloud Run

```bash
# Logs del servicio
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=microservice-template" --limit 50

# Logs en tiempo real
gcloud logging tail "resource.type=cloud_run_revision AND resource.labels.service_name=microservice-template"
```

## Comandos Útiles

### Ejecutar Build Manualmente

```bash
aws codebuild start-build --project-name microservice-template
```

### Verificar Despliegue

```bash
# Estado del servicio
gcloud run services describe microservice-template --region=us-central1

# URL del servicio
gcloud run services describe microservice-template --region=us-central1 --format="value(status.url)"
```

### Rollback

```bash
# Listar revisiones
gcloud run revisions list --service=microservice-template --region=us-central1

# Hacer rollback a una revisión específica
gcloud run services update-traffic microservice-template --to-revisions=REVISION_NAME=100 --region=us-central1
```

## Troubleshooting

### Errores Comunes

1. **Error de autenticación con GCP**
   - Verificar que el Service Account tenga los permisos correctos
   - Revisar que la clave esté correctamente guardada en Secrets Manager

2. **Error al subir imagen a GCR**
   - Verificar que el proyecto GCP tenga habilitada la Container Registry API
   - Confirmar que el Service Account tenga permisos de Storage Admin

3. **Error al desplegar en Cloud Run**
   - Verificar que Cloud Run API esté habilitada
   - Confirmar que los secretos existan en Google Secret Manager

### Debug

```bash
# Verificar configuración de CodeBuild
aws codebuild batch-get-projects --names microservice-template

# Verificar parámetros
aws ssm get-parameter --name /microservice-template/gcp/project-id

# Verificar secretos
aws secretsmanager describe-secret --secret-id microservice-template/gcp-service-account
```

## Seguridad

- Las credenciales de GCP se almacenan encriptadas en AWS Secrets Manager
- El Service Account de GCP tiene permisos mínimos necesarios
- Las imágenes Docker se construyen sin privilegios root
- Los secretos de aplicación se gestionan a través de Google Secret Manager

## Costos

- **CodeBuild**: Facturación por minuto de build
- **Cloud Run**: Facturación por requests y tiempo de CPU
- **Container Registry**: Almacenamiento de imágenes
- **Secrets Manager**: Almacenamiento de secretos

Para optimizar costos:
- Usar cache de dependencias en CodeBuild
- Configurar correctamente el escalado automático en Cloud Run
- Limpiar imágenes antiguas en GCR regularmente