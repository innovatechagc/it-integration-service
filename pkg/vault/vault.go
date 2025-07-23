package vault

import (
	"fmt"

	"github.com/company/microservice-template/internal/config"
	"github.com/hashicorp/vault/api"
)

type Client interface {
	GetSecret(path string) (map[string]interface{}, error)
	GetSecretValue(path, key string) (string, error)
}

type vaultClient struct {
	client *api.Client
	path   string
}

func NewClient(cfg config.VaultConfig) (Client, error) {
	config := api.DefaultConfig()
	config.Address = cfg.Address
	
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}
	
	if cfg.Token != "" {
		client.SetToken(cfg.Token)
	}
	
	return &vaultClient{
		client: client,
		path:   cfg.Path,
	}, nil
}

func (v *vaultClient) GetSecret(path string) (map[string]interface{}, error) {
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from vault: %w", err)
	}
	
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}
	
	return secret.Data, nil
}

func (v *vaultClient) GetSecretValue(path, key string) (string, error) {
	data, err := v.GetSecret(path)
	if err != nil {
		return "", err
	}
	
	value, exists := data[key]
	if !exists {
		return "", fmt.Errorf("key %s not found in secret at path %s", key, path)
	}
	
	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value for key %s is not a string", key)
	}
	
	return strValue, nil
}

// Ejemplo de uso comentado:
/*
// Para obtener un secreto completo:
secrets, err := vaultClient.GetSecret("secret/myapp/database")
if err != nil {
    log.Fatal(err)
}

// Para obtener un valor específico:
dbPassword, err := vaultClient.GetSecretValue("secret/myapp/database", "password")
if err != nil {
    log.Fatal(err)
}
*/