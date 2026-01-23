package cqrs

import (
	"context"
	"fmt"

	"github.com/yourusername/go-generic-event-driven/pkg/ddd"
)

// Command represents a CQRS command
type Command struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	AggregateID string      `json:"aggregate_id"`
	Data        interface{} `json:"data"`
}

func NewCommand(commandType, aggregateID string, data interface{}) *Command {
	return &Command{
		Type:        commandType,
		AggregateID: aggregateID,
		Data:        data,
	}
}

func (c *Command) GetID() string          { return c.ID }
func (c *Command) GetType() string        { return c.Type }
func (c *Command) GetData() interface{}   { return c.Data }
func (c *Command) GetAggregateID() string { return c.AggregateID }

// Query represents a CQRS query
type Query struct {
	ID   string      `json:"id"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewQuery(queryType string, data interface{}) *Query {
	return &Query{
		Type: queryType,
		Data: data,
	}
}

func (q *Query) GetID() string        { return q.ID }
func (q *Query) GetType() string      { return q.Type }
func (q *Query) GetData() interface{} { return q.Data }

// CommandResult represents the result of a command
type CommandResult struct {
	Events      []ddd.Event `json:"events"`
	AggregateID string      `json:"aggregate_id"`
	Version     int         `json:"version"`
}

// CommandBus handles command dispatching
type CommandBus interface {
	Register(commandType string, handler CommandHandler) error
	Dispatch(ctx context.Context, cmd *Command) (*CommandResult, error)
}

// QueryBus handles query dispatching
type QueryBus interface {
	Register(queryType string, handler QueryHandler) error
	Dispatch(ctx context.Context, query *Query) (interface{}, error)
}

// EventBus handles event publishing
type EventBus interface {
	Subscribe(eventType string, handler EventHandler) error
	Publish(ctx context.Context, event ddd.Event) error
}

// CommandHandler handles commands
type CommandHandler interface {
	Handle(ctx context.Context, cmd *Command) (*CommandResult, error)
}

// QueryHandler handles queries
type QueryHandler interface {
	Handle(ctx context.Context, query *Query) (interface{}, error)
}

// EventHandler handles events
type EventHandler interface {
	Handle(ctx context.Context, event ddd.Event) error
}

// InMemoryCommandBus implements an in-memory command bus
type InMemoryCommandBus struct {
	handlers map[string]CommandHandler
}

func NewInMemoryCommandBus() *InMemoryCommandBus {
	return &InMemoryCommandBus{
		handlers: make(map[string]CommandHandler),
	}
}

func (bus *InMemoryCommandBus) Register(commandType string, handler CommandHandler) error {
	if _, exists := bus.handlers[commandType]; exists {
		return fmt.Errorf("command handler for type %s already registered", commandType)
	}
	bus.handlers[commandType] = handler
	return nil
}

func (bus *InMemoryCommandBus) Dispatch(ctx context.Context, cmd *Command) (*CommandResult, error) {
	handler, exists := bus.handlers[cmd.Type]
	if !exists {
		return nil, fmt.Errorf("no handler registered for command type: %s", cmd.Type)
	}

	return handler.Handle(ctx, cmd)
}

// InMemoryQueryBus implements an in-memory query bus
type InMemoryQueryBus struct {
	handlers map[string]QueryHandler
}

func NewInMemoryQueryBus() *InMemoryQueryBus {
	return &InMemoryQueryBus{
		handlers: make(map[string]QueryHandler),
	}
}

func (bus *InMemoryQueryBus) Register(queryType string, handler QueryHandler) error {
	if _, exists := bus.handlers[queryType]; exists {
		return fmt.Errorf("query handler for type %s already registered", queryType)
	}
	bus.handlers[queryType] = handler
	return nil
}

func (bus *InMemoryQueryBus) Dispatch(ctx context.Context, query *Query) (interface{}, error) {
	handler, exists := bus.handlers[query.Type]
	if !exists {
		return nil, fmt.Errorf("no handler registered for query type: %s", query.Type)
	}

	return handler.Handle(ctx, query)
}

// InMemoryEventBus implements an in-memory event bus
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
}

func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
	if _, exists := bus.handlers[eventType]; !exists {
		bus.handlers[eventType] = make([]EventHandler, 0)
	}
	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
	return nil
}

func (bus *InMemoryEventBus) Publish(ctx context.Context, event ddd.Event) error {
	handlers, exists := bus.handlers[event.GetType()]
	if !exists {
		return nil // No handlers for this event type
	}

	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			return fmt.Errorf("event handler failed for event %s: %w", event.GetType(), err)
		}
	}

	return nil
}

// CQRS combines all buses
type CQRS struct {
	CommandBus CommandBus
	QueryBus   QueryBus
	EventBus   EventBus
}

func NewCQRS(commandBus CommandBus, queryBus QueryBus, eventBus EventBus) *CQRS {
	return &CQRS{
		CommandBus: commandBus,
		QueryBus:   queryBus,
		EventBus:   eventBus,
	}
}

// Helper methods for common operations
func (c *CQRS) RegisterCommandHandler(commandType string, handler CommandHandler) error {
	return c.CommandBus.Register(commandType, handler)
}

func (c *CQRS) RegisterQueryHandler(queryType string, handler QueryHandler) error {
	return c.QueryBus.Register(queryType, handler)
}

func (c *CQRS) SubscribeToEvent(eventType string, handler EventHandler) error {
	return c.EventBus.Subscribe(eventType, handler)
}

func (c *CQRS) ExecuteCommand(ctx context.Context, cmd *Command) (*CommandResult, error) {
	result, err := c.CommandBus.Dispatch(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Publish events from command result
	for _, event := range result.Events {
		if err := c.EventBus.Publish(ctx, event); err != nil {
			return nil, fmt.Errorf("failed to publish event %s: %w", event.GetType(), err)
		}
	}

	return result, nil
}

func (c *CQRS) ExecuteQuery(ctx context.Context, query *Query) (interface{}, error) {
	return c.QueryBus.Dispatch(ctx, query)
}
