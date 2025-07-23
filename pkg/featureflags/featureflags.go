package featureflags

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/company/microservice-template/pkg/logger"
)

// FeatureFlag representa una feature flag
type FeatureFlag struct {
	Key         string                 `json:"key"`
	Enabled     bool                   `json:"enabled"`
	Percentage  int                    `json:"percentage"` // 0-100
	Rules       []Rule                 `json:"rules"`
	Metadata    map[string]interface{} `json:"metadata"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Rule representa una regla de feature flag
type Rule struct {
	Attribute string      `json:"attribute"`
	Operator  string      `json:"operator"` // eq, ne, in, not_in
	Value     interface{} `json:"value"`
}

// Context representa el contexto de evaluación
type Context struct {
	UserID     string                 `json:"user_id"`
	Email      string                 `json:"email"`
	Attributes map[string]interface{} `json:"attributes"`
}

// Client interface para feature flags
type Client interface {
	IsEnabled(ctx context.Context, flagKey string, evalCtx Context) bool
	GetVariation(ctx context.Context, flagKey string, evalCtx Context, defaultValue interface{}) interface{}
	RefreshFlags(ctx context.Context) error
	Close() error
}

// InMemoryClient implementación en memoria para desarrollo
type InMemoryClient struct {
	flags  map[string]FeatureFlag
	mutex  sync.RWMutex
	logger logger.Logger
}

func NewInMemoryClient(logger logger.Logger) Client {
	client := &InMemoryClient{
		flags:  make(map[string]FeatureFlag),
		logger: logger,
	}
	
	// Flags por defecto para desarrollo
	client.setDefaultFlags()
	
	return client
}

func (c *InMemoryClient) setDefaultFlags() {
	defaultFlags := []FeatureFlag{
		{
			Key:       "new_user_onboarding",
			Enabled:   true,
			Percentage: 100,
			UpdatedAt: time.Now(),
		},
		{
			Key:       "advanced_analytics",
			Enabled:   false,
			Percentage: 0,
			UpdatedAt: time.Now(),
		},
		{
			Key:       "beta_features",
			Enabled:   true,
			Percentage: 10, // Solo 10% de usuarios
			Rules: []Rule{
				{
					Attribute: "user_type",
					Operator:  "eq",
					Value:     "beta_tester",
				},
			},
			UpdatedAt: time.Now(),
		},
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for _, flag := range defaultFlags {
		c.flags[flag.Key] = flag
	}
}

func (c *InMemoryClient) IsEnabled(ctx context.Context, flagKey string, evalCtx Context) bool {
	c.mutex.RLock()
	flag, exists := c.flags[flagKey]
	c.mutex.RUnlock()

	if !exists {
		c.logger.Debug("Feature flag not found", "flag_key", flagKey)
		return false
	}

	if !flag.Enabled {
		return false
	}

	// Evaluar reglas
	if len(flag.Rules) > 0 {
		if !c.evaluateRules(flag.Rules, evalCtx) {
			return false
		}
	}

	// Evaluar porcentaje
	if flag.Percentage < 100 {
		return c.evaluatePercentage(flagKey, evalCtx.UserID, flag.Percentage)
	}

	return true
}

func (c *InMemoryClient) GetVariation(ctx context.Context, flagKey string, evalCtx Context, defaultValue interface{}) interface{} {
	if !c.IsEnabled(ctx, flagKey, evalCtx) {
		return defaultValue
	}

	c.mutex.RLock()
	flag, exists := c.flags[flagKey]
	c.mutex.RUnlock()

	if !exists {
		return defaultValue
	}

	// Si hay metadata con variaciones, devolverla
	if variation, exists := flag.Metadata["variation"]; exists {
		return variation
	}

	return defaultValue
}

func (c *InMemoryClient) evaluateRules(rules []Rule, evalCtx Context) bool {
	for _, rule := range rules {
		var value interface{}
		
		// Obtener valor del contexto
		switch rule.Attribute {
		case "user_id":
			value = evalCtx.UserID
		case "email":
			value = evalCtx.Email
		default:
			if v, exists := evalCtx.Attributes[rule.Attribute]; exists {
				value = v
			}
		}

		if !c.evaluateRule(rule, value) {
			return false
		}
	}
	return true
}

func (c *InMemoryClient) evaluateRule(rule Rule, value interface{}) bool {
	switch rule.Operator {
	case "eq":
		return value == rule.Value
	case "ne":
		return value != rule.Value
	case "in":
		if slice, ok := rule.Value.([]interface{}); ok {
			for _, v := range slice {
				if v == value {
					return true
				}
			}
		}
		return false
	case "not_in":
		if slice, ok := rule.Value.([]interface{}); ok {
			for _, v := range slice {
				if v == value {
					return false
				}
			}
		}
		return true
	}
	return false
}

func (c *InMemoryClient) evaluatePercentage(flagKey, userID string, percentage int) bool {
	// Hash simple para distribución consistente
	hash := 0
	for _, char := range flagKey + userID {
		hash = hash*31 + int(char)
	}
	
	if hash < 0 {
		hash = -hash
	}
	
	return (hash % 100) < percentage
}

func (c *InMemoryClient) RefreshFlags(ctx context.Context) error {
	c.logger.Info("Refreshing feature flags")
	// En una implementación real, aquí cargarías desde una fuente externa
	return nil
}

func (c *InMemoryClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.flags = make(map[string]FeatureFlag)
	return nil
}

// LaunchDarklyClient implementación con LaunchDarkly (comentada)
/*
type LaunchDarklyClient struct {
	client ldclient.LDClient
	logger logger.Logger
}

func NewLaunchDarklyClient(sdkKey string, logger logger.Logger) (Client, error) {
	config := ld.DefaultConfig
	client, err := ldclient.MakeClient(sdkKey, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create LaunchDarkly client: %w", err)
	}

	return &LaunchDarklyClient{
		client: client,
		logger: logger,
	}, nil
}

func (c *LaunchDarklyClient) IsEnabled(ctx context.Context, flagKey string, evalCtx Context) bool {
	user := lduser.NewUser(evalCtx.UserID).
		Email(evalCtx.Email)
	
	for key, value := range evalCtx.Attributes {
		user = user.Custom(key, ldvalue.String(fmt.Sprintf("%v", value)))
	}

	return c.client.BoolVariation(flagKey, user, false)
}
*/