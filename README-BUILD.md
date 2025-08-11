# Guía de Build del IT Integration Service

## 📁 Estructura Simplificada

El proyecto ahora tiene **un solo archivo principal** para mantener la simplicidad:

### 🔧 `main.go` (Versión Única)
- **Base de datos:** PostgreSQL real
- **Repositorios:** Implementaciones reales
- **Uso:** Producción y desarrollo completo
- **Configuración:** Completa con todas las funcionalidades

## 🚀 Comandos de Ejecución

### Desarrollo Local
```bash
# Ejecutar directamente
go run main.go

# O usar el comando del Makefile
make dev-simple
```

### Docker
```bash
make dev
```

## 🔨 Comandos de Build

### Build de Producción
```bash
# Compilar aplicación
make build
# o
go build -o bin/integration-service .
```

### Build Local
```bash
# Compilar para desarrollo local
make build-local
# o
go build -o bin/integration-service .
```

## 📊 Características

| Aspecto | Estado |
|---------|--------|
| **Base de datos** | ✅ PostgreSQL real |
| **Repositorios** | ✅ Implementaciones reales |
| **Configuración** | ✅ Completa |
| **Webhooks** | ✅ Funcionales |
| **Documentación** | ✅ Swagger disponible |
| **Métricas** | ✅ Prometheus |
| **Logs** | ✅ Estructurados |

## 🎯 Ventajas de la Estructura Única

1. **Simplicidad:** Un solo archivo para mantener
2. **Consistencia:** Mismo comportamiento en todos los entornos
3. **Datos reales:** Siempre trabajas con datos reales
4. **Menos confusión:** No hay que elegir entre versiones
5. **Mantenimiento:** Más fácil de mantener y debuggear

## ⚠️ Requisitos

### Base de Datos
- PostgreSQL configurado y accesible
- Variables de entorno configuradas en `.env.local`

### Variables de Entorno
```bash
# Copiar el archivo de ejemplo
cp env.example .env.local

# Configurar las variables necesarias
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=itapp
```

## 🔧 Troubleshooting

### Error: "database connection failed"
```bash
# Verificar que PostgreSQL esté corriendo
sudo systemctl status postgresql

# Verificar variables de entorno
cat .env.local
```

### Error: "port already in use"
```bash
# Detener otros procesos
pkill -f integration-service
# o cambiar puerto en .env
PORT=8081
```

### Error: "module not found"
```bash
# Descargar dependencias
make deps
# o
go mod download
go mod tidy
```

## 📝 Notas de Desarrollo

- **Siempre usa datos reales** - no hay modo mock
- **Configura la base de datos** antes de ejecutar
- **Los webhooks funcionan** con datos reales
- **La documentación Swagger** está disponible en `/swagger/index.html`
- **Las métricas** están disponibles en `/metrics` 