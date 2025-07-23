package testing

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainers struct {
	PostgresContainer testcontainers.Container
	VaultContainer    testcontainers.Container
	RedisContainer    testcontainers.Container
}

func SetupTestContainers(ctx context.Context) (*TestContainers, error) {
	containers := &TestContainers{}

	// PostgreSQL Container
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "test_db",
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}
	containers.PostgresContainer = postgresContainer

	// Vault Container
	vaultReq := testcontainers.ContainerRequest{
		Image:        "vault:latest",
		ExposedPorts: []string{"8200/tcp"},
		Env: map[string]string{
			"VAULT_DEV_ROOT_TOKEN_ID":      "test-token",
			"VAULT_DEV_LISTEN_ADDRESS":     "0.0.0.0:8200",
		},
		Cmd: []string{"vault", "server", "-dev"},
		WaitingFor: wait.ForHTTP("/v1/sys/health").WithPort("8200/tcp").WithStartupTimeout(30 * time.Second),
	}

	vaultContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: vaultReq,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start vault container: %w", err)
	}
	containers.VaultContainer = vaultContainer

	// Redis Container
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(30 * time.Second),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}
	containers.RedisContainer = redisContainer

	return containers, nil
}

func (tc *TestContainers) GetPostgresConnectionString(ctx context.Context) (string, error) {
	host, err := tc.PostgresContainer.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := tc.PostgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("postgres://test_user:test_password@%s:%s/test_db?sslmode=disable", host, port.Port()), nil
}

func (tc *TestContainers) GetVaultAddress(ctx context.Context) (string, error) {
	host, err := tc.VaultContainer.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := tc.VaultContainer.MappedPort(ctx, "8200")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

func (tc *TestContainers) GetRedisAddress(ctx context.Context) (string, error) {
	host, err := tc.RedisContainer.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := tc.RedisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}

func (tc *TestContainers) Cleanup(ctx context.Context) error {
	if tc.PostgresContainer != nil {
		if err := tc.PostgresContainer.Terminate(ctx); err != nil {
			return err
		}
	}
	if tc.VaultContainer != nil {
		if err := tc.VaultContainer.Terminate(ctx); err != nil {
			return err
		}
	}
	if tc.RedisContainer != nil {
		if err := tc.RedisContainer.Terminate(ctx); err != nil {
			return err
		}
	}
	return nil
}