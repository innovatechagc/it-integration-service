module github.com/company/microservice-template

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/joho/godotenv v1.5.1
	github.com/hashicorp/vault/api v1.10.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
	go.uber.org/zap v1.26.0
	github.com/prometheus/client_golang v1.17.0
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/files v1.0.1
	github.com/swaggo/swag v1.16.2
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/testcontainers/testcontainers-go v0.26.0
	github.com/golang-migrate/migrate/v4 v4.17.0
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	github.com/redis/go-redis/v9 v9.3.0
)