package cqrs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/go-generic-event-driven/pkg/ddd"
)

type testEvent struct {
	ID        string
	Type      string
	Timestamp time.Time
	Data      interface{}
}

func (e *testEvent) GetID() string           { return e.ID }
func (e *testEvent) GetType() string         { return e.Type }
func (e *testEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *testEvent) GetData() interface{}    { return e.Data }

func newTestEvent(id, eventType string) ddd.Event {
	return &testEvent{ID: id, Type: eventType, Timestamp: time.Now()}
}

type testCommandHandler struct {
	result *CommandResult
	err    error
}

func (h *testCommandHandler) Handle(ctx context.Context, cmd *Command) (*CommandResult, error) {
	return h.result, h.err
}

type testQueryHandler struct {
	result interface{}
	err    error
}

func (h *testQueryHandler) Handle(ctx context.Context, q *Query) (interface{}, error) {
	return h.result, h.err
}

type testEventHandler struct {
	err error
}

func (h *testEventHandler) Handle(ctx context.Context, event ddd.Event) error {
	return h.err
}

func TestInMemoryCommandBus_RegisterDuplicate(t *testing.T) {
	bus := NewInMemoryCommandBus()
	h := &testCommandHandler{}

	if err := bus.Register("cmd.type", h); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := bus.Register("cmd.type", h); err == nil {
		t.Fatalf("expected duplicate registration error")
	}
}

func TestInMemoryCommandBus_DispatchNoHandler(t *testing.T) {
	bus := NewInMemoryCommandBus()
	_, err := bus.Dispatch(context.Background(), &Command{Type: "missing"})
	if err == nil {
		t.Fatalf("expected error for missing handler")
	}
}

func TestInMemoryCommandBus_DispatchCallsHandler(t *testing.T) {
	bus := NewInMemoryCommandBus()
	result := &CommandResult{Events: []ddd.Event{newTestEvent("e1", "evt")}}
	h := &testCommandHandler{result: result}

	if err := bus.Register("cmd.type", h); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := bus.Dispatch(context.Background(), &Command{Type: "cmd.type"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || len(got.Events) != 1 {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestInMemoryQueryBus_RegisterDuplicate(t *testing.T) {
	bus := NewInMemoryQueryBus()
	h := &testQueryHandler{}

	if err := bus.Register("query.type", h); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := bus.Register("query.type", h); err == nil {
		t.Fatalf("expected duplicate registration error")
	}
}

func TestInMemoryQueryBus_DispatchNoHandler(t *testing.T) {
	bus := NewInMemoryQueryBus()
	_, err := bus.Dispatch(context.Background(), &Query{Type: "missing"})
	if err == nil {
		t.Fatalf("expected error for missing handler")
	}
}

func TestInMemoryQueryBus_DispatchCallsHandler(t *testing.T) {
	bus := NewInMemoryQueryBus()
	h := &testQueryHandler{result: "ok"}

	if err := bus.Register("query.type", h); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := bus.Dispatch(context.Background(), &Query{Type: "query.type"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ok" {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestInMemoryEventBus_PublishNoHandlers(t *testing.T) {
	bus := NewInMemoryEventBus()
	if err := bus.Publish(context.Background(), newTestEvent("e1", "evt")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInMemoryEventBus_PublishHandlerError(t *testing.T) {
	bus := NewInMemoryEventBus()
	h := &testEventHandler{err: errors.New("boom")}

	if err := bus.Subscribe("evt", h); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := bus.Publish(context.Background(), newTestEvent("e1", "evt")); err == nil {
		t.Fatalf("expected handler error")
	}
}

func TestInMemoryEventBus_PublishHandlerOK(t *testing.T) {
	bus := NewInMemoryEventBus()
	h := &testEventHandler{}

	if err := bus.Subscribe("evt", h); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := bus.Publish(context.Background(), newTestEvent("e1", "evt")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
