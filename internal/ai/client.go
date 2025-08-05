package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"it-integration-service/pkg/logger"
)

// AIClient interface genérico para clientes de IA
type AIClient interface {
	GenerateResponse(ctx context.Context, prompt string, options ...Option) (*Response, error)
	GenerateChatResponse(ctx context.Context, messages []Message, options ...Option) (*Response, error)
	Close() error
}

// Message representa un mensaje en una conversación
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// Response representa la respuesta de la IA
type Response struct {
	Content      string                 `json:"content"`
	TokensUsed   int                    `json:"tokens_used"`
	Model        string                 `json:"model"`
	FinishReason string                 `json:"finish_reason"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Option para configurar requests
type Option func(*RequestConfig)

type RequestConfig struct {
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
}

// OpenAIClient implementación para OpenAI
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     logger.Logger
}

func NewOpenAIClient(apiKey string, logger logger.Logger) AIClient {
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (c *OpenAIClient) GenerateResponse(ctx context.Context, prompt string, options ...Option) (*Response, error) {
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	return c.GenerateChatResponse(ctx, messages, options...)
}

func (c *OpenAIClient) GenerateChatResponse(ctx context.Context, messages []Message, options ...Option) (*Response, error) {
	config := &RequestConfig{
		Model:       "gpt-3.5-turbo",
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        1.0,
	}

	// Aplicar opciones
	for _, option := range options {
		option(config)
	}

	// Preparar request
	requestBody := map[string]interface{}{
		"model":       config.Model,
		"messages":    messages,
		"max_tokens":  config.MaxTokens,
		"temperature": config.Temperature,
		"top_p":       config.TopP,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Crear HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Ejecutar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// Parsear respuesta
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	response := &Response{
		Content:      openAIResp.Choices[0].Message.Content,
		TokensUsed:   openAIResp.Usage.TotalTokens,
		Model:        openAIResp.Model,
		FinishReason: openAIResp.Choices[0].FinishReason,
		Metadata:     make(map[string]interface{}),
	}

	c.logger.Info("AI response generated", 
		"model", response.Model, 
		"tokens_used", response.TokensUsed,
		"finish_reason", response.FinishReason)

	return response, nil
}

func (c *OpenAIClient) Close() error {
	return nil
}

// MockAIClient para testing y desarrollo
type MockAIClient struct {
	responses []string
	index     int
	logger    logger.Logger
}

func NewMockAIClient(responses []string, logger logger.Logger) AIClient {
	return &MockAIClient{
		responses: responses,
		index:     0,
		logger:    logger,
	}
}

func (c *MockAIClient) GenerateResponse(ctx context.Context, prompt string, options ...Option) (*Response, error) {
	if len(c.responses) == 0 {
		return &Response{
			Content:      "Esta es una respuesta mock para el prompt: " + prompt,
			TokensUsed:   50,
			Model:        "mock-model",
			FinishReason: "stop",
			Metadata:     make(map[string]interface{}),
		}, nil
	}

	response := c.responses[c.index%len(c.responses)]
	c.index++

	return &Response{
		Content:      response,
		TokensUsed:   len(response) / 4, // Aproximación simple
		Model:        "mock-model",
		FinishReason: "stop",
		Metadata:     make(map[string]interface{}),
	}, nil
}

func (c *MockAIClient) GenerateChatResponse(ctx context.Context, messages []Message, options ...Option) (*Response, error) {
	lastMessage := ""
	if len(messages) > 0 {
		lastMessage = messages[len(messages)-1].Content
	}
	return c.GenerateResponse(ctx, lastMessage, options...)
}

func (c *MockAIClient) Close() error {
	return nil
}

// Opciones de configuración
func WithModel(model string) Option {
	return func(config *RequestConfig) {
		config.Model = model
	}
}

func WithMaxTokens(maxTokens int) Option {
	return func(config *RequestConfig) {
		config.MaxTokens = maxTokens
	}
}

func WithTemperature(temperature float64) Option {
	return func(config *RequestConfig) {
		config.Temperature = temperature
	}
}

func WithTopP(topP float64) Option {
	return func(config *RequestConfig) {
		config.TopP = topP
	}
}