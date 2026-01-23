package ddd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BaseEvent provides a generic event implementation
type BaseEvent struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

func NewBaseEvent(eventType string, data interface{}) *BaseEvent {
	return &BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

func (e *BaseEvent) GetID() string           { return e.ID }
func (e *BaseEvent) GetType() string         { return e.Type }
func (e *BaseEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *BaseEvent) GetData() interface{}    { return e.Data }

// BaseAggregate provides a generic aggregate implementation
type BaseAggregate struct {
	ID                string
	Version           int
	UncommittedEvents []Event
}

func NewBaseAggregate(id string) *BaseAggregate {
	return &BaseAggregate{
		ID:                id,
		Version:           0,
		UncommittedEvents: make([]Event, 0),
	}
}

func (a *BaseAggregate) GetID() string                 { return a.ID }
func (a *BaseAggregate) GetVersion() int               { return a.Version }
func (a *BaseAggregate) GetUncommittedEvents() []Event { return a.UncommittedEvents }

func (a *BaseAggregate) AddEvent(event Event) {
	a.UncommittedEvents = append(a.UncommittedEvents, event)
	a.Version++
}

func (a *BaseAggregate) MarkEventsAsCommitted() {
	a.UncommittedEvents = make([]Event, 0)
}

// EventSerializer handles event serialization/deserialization
type EventSerializer interface {
	Serialize(event Event) ([]byte, error)
	Deserialize(eventType string, data []byte) (Event, error)
}

// JSONEventSerializer implements JSON serialization
type JSONEventSerializer struct {
	eventTypeRegistry map[string]func() Event
}

func NewJSONEventSerializer() *JSONEventSerializer {
	return &JSONEventSerializer{
		eventTypeRegistry: make(map[string]func() Event),
	}
}

func (s *JSONEventSerializer) RegisterEventType(eventType string, factory func() Event) {
	s.eventTypeRegistry[eventType] = factory
}

func (s *JSONEventSerializer) Serialize(event Event) ([]byte, error) {
	return json.Marshal(event)
}

func (s *JSONEventSerializer) Deserialize(eventType string, data []byte) (Event, error) {
	factory, exists := s.eventTypeRegistry[eventType]
	if !exists {
		// Fallback to BaseEvent
		var baseEvent BaseEvent
		if err := json.Unmarshal(data, &baseEvent); err != nil {
			return nil, fmt.Errorf("failed to deserialize base event: %w", err)
		}
		baseEvent.Type = eventType
		return &baseEvent, nil
	}

	event := factory()
	if err := json.Unmarshal(data, event); err != nil {
		return nil, fmt.Errorf("failed to deserialize event %s: %w", eventType, err)
	}

	return event, nil
}

// EventWithMetadata represents an event with additional metadata
type EventWithMetadata struct {
	Event       Event
	AggregateID string
	Version     int
	Metadata    map[string]interface{}
}

// Specification implementations

type AndSpecification[T any] struct {
	left, right Specification[T]
}

func (s *AndSpecification[T]) IsSatisfiedBy(candidate T) bool {
	return s.left.IsSatisfiedBy(candidate) && s.right.IsSatisfiedBy(candidate)
}

func (s *AndSpecification[T]) And(other Specification[T]) Specification[T] {
	return &AndSpecification[T]{left: s, right: other}
}

func (s *AndSpecification[T]) Or(other Specification[T]) Specification[T] {
	return &OrSpecification[T]{left: s, right: other}
}

func (s *AndSpecification[T]) Not() Specification[T] {
	return &NotSpecification[T]{spec: s}
}

type OrSpecification[T any] struct {
	left, right Specification[T]
}

func (s *OrSpecification[T]) IsSatisfiedBy(candidate T) bool {
	return s.left.IsSatisfiedBy(candidate) || s.right.IsSatisfiedBy(candidate)
}

func (s *OrSpecification[T]) And(other Specification[T]) Specification[T] {
	return &AndSpecification[T]{left: s, right: other}
}

func (s *OrSpecification[T]) Or(other Specification[T]) Specification[T] {
	return &OrSpecification[T]{left: s, right: other}
}

func (s *OrSpecification[T]) Not() Specification[T] {
	return &NotSpecification[T]{spec: s}
}

type NotSpecification[T any] struct {
	spec Specification[T]
}

func (s *NotSpecification[T]) IsSatisfiedBy(candidate T) bool {
	return !s.spec.IsSatisfiedBy(candidate)
}

func (s *NotSpecification[T]) And(other Specification[T]) Specification[T] {
	return &AndSpecification[T]{left: s, right: other}
}

func (s *NotSpecification[T]) Or(other Specification[T]) Specification[T] {
	return &OrSpecification[T]{left: s, right: other}
}

func (s *NotSpecification[T]) Not() Specification[T] {
	return s.spec
}

// Helper functions for creating specifications
func And[T any](left, right Specification[T]) Specification[T] {
	return &AndSpecification[T]{left: left, right: right}
}

func Or[T any](left, right Specification[T]) Specification[T] {
	return &OrSpecification[T]{left: left, right: right}
}

func Not[T any](spec Specification[T]) Specification[T] {
	return &NotSpecification[T]{spec: spec}
}
