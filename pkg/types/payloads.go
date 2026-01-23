package types

import (
	"time"

	"github.com/google/uuid"
)

// Generic payload types for common operations

// CreatePayload represents a generic create operation payload
type CreatePayload struct {
	Data interface{} `json:"data"`
}

// UpdatePayload represents a generic update operation payload
type UpdatePayload struct {
	ID   string      `json:"id"`
	Data interface{} `json:"data"`
}

// DeletePayload represents a generic delete operation payload
type DeletePayload struct {
	ID string `json:"id"`
}

// GetPayload represents a generic get operation payload
type GetPayload struct {
	ID string `json:"id"`
}

// ListPayload represents a generic list operation payload
type ListPayload struct {
	Filter   map[string]interface{} `json:"filter,omitempty"`
	Sort     map[string]interface{} `json:"sort,omitempty"`
	Page     int                    `json:"page,omitempty"`
	PageSize int                    `json:"page_size,omitempty"`
}

// SearchPayload represents a generic search operation payload
type SearchPayload struct {
	Query    string                 `json:"query"`
	Filter   map[string]interface{} `json:"filter,omitempty"`
	Sort     map[string]interface{} `json:"sort,omitempty"`
	Page     int                    `json:"page,omitempty"`
	PageSize int                    `json:"page_size,omitempty"`
}

// Response represents a generic response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ListResponse represents a generic list response
type ListResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Total   int         `json:"total"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error"`
	Code    string      `json:"code,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// Metadata represents generic metadata
type Metadata struct {
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Version    int                    `json:"version"`
	CreatedBy  string                 `json:"created_by,omitempty"`
	UpdatedBy  string                 `json:"updated_by,omitempty"`
	Tags       map[string]string      `json:"tags,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Entity represents a generic entity with metadata
type Entity struct {
	ID       string   `json:"id"`
	Metadata Metadata `json:"metadata"`
}

// NewEntity creates a new entity with default metadata
func NewEntity(id string) *Entity {
	now := time.Now()
	return &Entity{
		ID: id,
		Metadata: Metadata{
			CreatedAt:  now,
			UpdatedAt:  now,
			Version:    1,
			Tags:       make(map[string]string),
			Properties: make(map[string]interface{}),
		},
	}
}

// NewEntityWithCreator creates a new entity with creator information
func NewEntityWithCreator(id, creatorID string) *Entity {
	entity := NewEntity(id)
	entity.Metadata.CreatedBy = creatorID
	entity.Metadata.UpdatedBy = creatorID
	return entity
}

// UpdateMetadata updates the entity metadata
func (e *Entity) UpdateMetadata(updatedBy string) {
	e.Metadata.UpdatedAt = time.Now()
	e.Metadata.Version++
	e.Metadata.UpdatedBy = updatedBy
}

// AddTag adds a tag to the entity
func (e *Entity) AddTag(key, value string) {
	if e.Metadata.Tags == nil {
		e.Metadata.Tags = make(map[string]string)
	}
	e.Metadata.Tags[key] = value
}

// SetProperty sets a property on the entity
func (e *Entity) SetProperty(key string, value interface{}) {
	if e.Metadata.Properties == nil {
		e.Metadata.Properties = make(map[string]interface{})
	}
	e.Metadata.Properties[key] = value
}

// GetProperty gets a property from the entity
func (e *Entity) GetProperty(key string) (interface{}, bool) {
	if e.Metadata.Properties == nil {
		return nil, false
	}
	value, exists := e.Metadata.Properties[key]
	return value, exists
}

// Command and Query types

// Command represents a generic command
type Command struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   interface{}            `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewCommand creates a new command
func NewCommand(commandType string, payload interface{}) *Command {
	return &Command{
		ID:        uuid.New().String(),
		Type:      commandType,
		Timestamp: time.Now(),
		Payload:   payload,
		Metadata:  make(map[string]interface{}),
	}
}

// Query represents a generic query
type Query struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   interface{}            `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewQuery creates a new query
func NewQuery(queryType string, payload interface{}) *Query {
	return &Query{
		ID:        uuid.New().String(),
		Type:      queryType,
		Timestamp: time.Now(),
		Payload:   payload,
		Metadata:  make(map[string]interface{}),
	}
}

// Event represents a generic event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   interface{}            `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewEvent creates a new event
func NewEvent(eventType string, payload interface{}) *Event {
	return &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   payload,
		Metadata:  make(map[string]interface{}),
	}
}

// Helper functions for creating common payloads

// NewCreatePayload creates a create payload
func NewCreatePayload(data interface{}) *CreatePayload {
	return &CreatePayload{Data: data}
}

// NewUpdatePayload creates an update payload
func NewUpdatePayload(id string, data interface{}) *UpdatePayload {
	return &UpdatePayload{ID: id, Data: data}
}

// NewDeletePayload creates a delete payload
func NewDeletePayload(id string) *DeletePayload {
	return &DeletePayload{ID: id}
}

// NewGetPayload creates a get payload
func NewGetPayload(id string) *GetPayload {
	return &GetPayload{ID: id}
}

// NewListPayload creates a list payload
func NewListPayload(page, pageSize int) *ListPayload {
	return &ListPayload{
		Page:     page,
		PageSize: pageSize,
		Filter:   make(map[string]interface{}),
		Sort:     make(map[string]interface{}),
	}
}

// NewSearchPayload creates a search payload
func NewSearchPayload(query string, page, pageSize int) *SearchPayload {
	return &SearchPayload{
		Query:    query,
		Page:     page,
		PageSize: pageSize,
		Filter:   make(map[string]interface{}),
		Sort:     make(map[string]interface{}),
	}
}
