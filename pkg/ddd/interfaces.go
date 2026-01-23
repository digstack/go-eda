package ddd

import (
	"time"
)

// Generic interfaces for Domain-Driven Design

// Event represents a domain event
type Event interface {
	GetID() string
	GetType() string
	GetTimestamp() time.Time
	GetData() interface{}
}

// Aggregate represents a DDD aggregate
type Aggregate interface {
	GetID() string
	GetVersion() int
	GetUncommittedEvents() []Event
	MarkEventsAsCommitted()
}

// Entity represents a DDD entity
type Entity interface {
	GetID() string
}

// ValueObject represents a DDD value object
type ValueObject interface {
	Equals(other ValueObject) bool
}

// Repository represents a DDD repository
type Repository[T Aggregate] interface {
	Save(aggregate T) error
	FindByID(id string) (T, error)
	Find(specification Specification[T]) ([]T, error)
	Delete(id string) error
}

// Specification represents a DDD specification pattern
type Specification[T any] interface {
	IsSatisfiedBy(candidate T) bool
	And(other Specification[T]) Specification[T]
	Or(other Specification[T]) Specification[T]
	Not() Specification[T]
}

// DomainService represents a DDD domain service
type DomainService interface {
	Execute(ctx DomainContext) error
}

// DomainContext provides context for domain operations
type DomainContext interface {
	GetCorrelationID() string
	GetUserID() string
	GetTenantID() string
	GetValue(key string) interface{}
	SetValue(key string, value interface{})
}

// EventBus represents an event bus for domain events
type EventBus interface {
	Publish(event Event) error
	Subscribe(eventType string, handler EventHandler) error
}

// EventHandler handles domain events
type EventHandler interface {
	Handle(ctx DomainContext, event Event) error
}

// Command represents a domain command
type Command interface {
	GetID() string
	GetType() string
	GetData() interface{}
	GetAggregateID() string
}

// CommandHandler handles domain commands
type CommandHandler interface {
	Handle(ctx DomainContext, cmd Command) ([]Event, error)
}

// Query represents a domain query
type Query interface {
	GetID() string
	GetType() string
	GetData() interface{}
}

// QueryHandler handles domain queries
type QueryHandler interface {
	Handle(ctx DomainContext, query Query) (interface{}, error)
}
