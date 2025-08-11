# 🧪 Tests - IT Integration Service

## 📁 Estructura de Tests

```
tests/
├── integration/                    # Tests de integración
│   └── health_integration_test.go  # Tests de health checks
├── webhook-simulator/              # Simulador de webhooks
│   └── index.html                  # Interfaz web para testing
├── nginx.conf                      # Configuración nginx para simulador
└── README.md                       # Este archivo
```

## 🚀 Cómo Ejecutar los Tests

### **Tests de Integración**
```bash
# Ejecutar todos los tests de integración
go test ./tests/integration/...

# Ejecutar tests específicos
go test ./tests/integration/ -v

# Ejecutar con coverage
go test ./tests/integration/ -cover
```

### **Tests Unitarios**
```bash
# Ejecutar tests unitarios de servicios
go test ./internal/services/...

# Ejecutar tests unitarios de handlers
go test ./internal/handlers/...

# Ejecutar tests unitarios de repository
go test ./internal/repository/...
```

## 🔗 Webhook Simulator

### **¿Qué es?**
El Webhook Simulator es una herramienta web que permite simular webhooks de diferentes plataformas de mensajería para testing.

### **Plataformas Soportadas:**
- ✅ **WhatsApp** (Meta)
- ✅ **Telegram**
- ✅ **Messenger** (Facebook)
- ✅ **Instagram**

### **Cómo Usar:**

#### **1. Iniciar el Simulador**
```bash
# Opción 1: Con nginx (recomendado)
docker run -d \
  --name webhook-simulator \
  -p 8081:80 \
  -v $(pwd)/tests/webhook-simulator:/usr/share/nginx/html \
  -v $(pwd)/tests/nginx.conf:/etc/nginx/nginx.conf \
  nginx:alpine

# Opción 2: Con Python (simple)
cd tests/webhook-simulator
python3 -m http.server 8081
```

#### **2. Acceder al Simulador**
Abrir en el navegador: `http://localhost:8081`

#### **3. Configurar Webhook URLs**
- **WhatsApp**: `http://localhost:8080/api/v1/integrations/webhooks/whatsapp`
- **Telegram**: `http://localhost:8080/api/v1/integrations/webhooks/telegram`
- **Messenger**: `http://localhost:8080/api/v1/integrations/webhooks/messenger`
- **Instagram**: `http://localhost:8080/api/v1/integrations/webhooks/instagram`

#### **4. Enviar Webhooks de Prueba**
1. Seleccionar la plataforma
2. Modificar el payload JSON si es necesario
3. Hacer clic en "Enviar Webhook"
4. Ver la respuesta del servidor

## 🧪 Tipos de Tests

### **1. Tests Unitarios**
- **Ubicación**: `internal/*/..._test.go`
- **Propósito**: Probar funciones individuales
- **Ejecución**: Rápida, sin dependencias externas

### **2. Tests de Integración**
- **Ubicación**: `tests/integration/`
- **Propósito**: Probar integración entre componentes
- **Ejecución**: Requiere contenedores de test (PostgreSQL, Redis, Vault)

### **3. Tests Manuales**
- **Herramienta**: Webhook Simulator
- **Propósito**: Probar integraciones reales con plataformas
- **Ejecución**: Manual, requiere servicio corriendo

## 🔧 Configuración de TestContainers

Los tests de integración usan TestContainers para:
- **PostgreSQL**: Base de datos de test
- **Redis**: Cache de test
- **Vault**: Gestión de secretos de test

### **Requisitos:**
- Docker instalado y corriendo
- Go 1.19+
- TestContainers Go

## 📊 Coverage de Tests

### **Objetivos de Coverage:**
- **Services**: >80%
- **Handlers**: >70%
- **Repository**: >90%
- **Integration**: >60%

### **Generar Reporte de Coverage:**
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## 🚨 Troubleshooting

### **Problema**: Tests de integración fallan
**Solución**: Verificar que Docker esté corriendo y que los puertos estén libres

### **Problema**: Webhook simulator no responde
**Solución**: Verificar que el servicio esté corriendo en puerto 8080

### **Problema**: Tests unitarios fallan
**Solución**: Verificar dependencias con `go mod tidy`

## 📝 Notas Importantes

1. **Los tests E2E complejos fueron eliminados** para simplificar el mantenimiento
2. **El webhook simulator es la herramienta principal** para testing manual
3. **Los tests de integración se enfocan** en health checks y funcionalidad básica
4. **Para testing completo** usar el webhook simulator con el servicio corriendo 