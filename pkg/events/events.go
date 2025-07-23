package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/company/microservice-template/pkg/logger"
)

// Event representa un evento del sistema
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
}

// EventBus interface para el bus de eventos
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType string, handler EventHandler) error
	Close() error
}

// EventHandler función para manejar eventos
type EventHandler func(ctx context.Context, event Event) error

// InMemoryEventBus implementación en memoria para desarrollo
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	logger   logger.Logger
}

func NewInMemoryEventBus(logger logger.Logger) EventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
		logger:   logger,
	}
}

func (bus *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	handlers, exists := bus.handlers[event.Type]
	if !exists {
		bus.logger.Debug("No handlers for event type", "event_type", event.Type)
		return nil
	}

	bus.logger.Info("Publishing event", "event_id", event.ID, "event_type", event.Type)

	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h(ctx, event); err != nil {
				bus.logger.Error("Event handler failed", "event_id", event.ID, "error", err)
			}
		}(handler)
	}

	return nil
}

func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
	bus.logger.Info("Subscribed to event type", "event_type", eventType)
	return nil
}

func (bus *InMemoryEventBus) Close() error {
	bus.handlers = make(map[string][]EventHandler)
	return nil
}

// PubSubEventBus implementación con Google Pub/Sub (comentada para desarrollo)
/*
type PubSubEventBus struct {
	client *pubsub.Client
	topic  *pubsub.Topic
	logger logger.Logger
}

func NewPubSubEventBus(projectID, topicName string, logger logger.Logger) (EventBus, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	topic := client.Topic(topicName)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check topic existence: %w", err)
	}

	if !exists {
		topic, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			return nil, fmt.Errorf("failed to create topic: %w", err)
		}
	}

	return &PubSubEventBus{
		client: client,
		topic:  topic,
		logger: logger,
	}, nil
}

func (bus *PubSubEventBus) Publish(ctx context.Context, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	result := bus.topic.Publish(ctx, &pubsub.Message{
		Data: data,
		Attributes: map[string]string{
			"event_type": event.Type,
			"source":     event.Source,
		},
	})

	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	bus.logger.Info("Event published to Pub/Sub", "event_id", event.ID, "event_type", event.Type)
	return nil
}
*/

// EventFactory para crear eventos estándar
type EventFactory struct {
	source string
}

func NewEventFactory(source string) *EventFactory {
	return &EventFactory{source: source}
}

func (f *EventFactory) CreateUserEvent(eventType, userID string, data map[string]interface{}) Event {
	return Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    f.source,
		Data:      data,
		Timestamp: time.Now().UTC(),
		UserID:    userID,
	}
}

func (f *EventFactory) CreateSystemEvent(eventType string, data map[string]interface{}) Event {
	return Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    f.source,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}