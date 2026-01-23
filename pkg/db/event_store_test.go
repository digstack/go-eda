package db

import (
	"context"
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
	return &testEvent{
		ID:        id,
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"id": id},
	}
}

func TestInMemoryEventStore_SaveAndGetEvents(t *testing.T) {
	store := NewInMemoryEventStore()
	ctx := context.Background()

	events := []ddd.EventWithMetadata{
		{Event: newTestEvent("e1", "type.a"), AggregateID: "agg-1", Version: 1},
		{Event: newTestEvent("e2", "type.b"), AggregateID: "agg-1", Version: 2},
	}

	if err := store.SaveEvents(ctx, "agg-1", events, 0); err != nil {
		t.Fatalf("SaveEvents error: %v", err)
	}

	got, err := store.GetEvents(ctx, "agg-1")
	if err != nil {
		t.Fatalf("GetEvents error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0].Event.GetID() != "e1" || got[1].Event.GetID() != "e2" {
		t.Fatalf("unexpected event order: %s, %s", got[0].Event.GetID(), got[1].Event.GetID())
	}
}

func TestInMemoryEventStore_GetEventsFromVersion(t *testing.T) {
	store := NewInMemoryEventStore()
	ctx := context.Background()

	events := []ddd.EventWithMetadata{
		{Event: newTestEvent("e1", "type.a"), AggregateID: "agg-1", Version: 1},
		{Event: newTestEvent("e2", "type.b"), AggregateID: "agg-1", Version: 2},
		{Event: newTestEvent("e3", "type.c"), AggregateID: "agg-1", Version: 3},
	}

	if err := store.SaveEvents(ctx, "agg-1", events, 0); err != nil {
		t.Fatalf("SaveEvents error: %v", err)
	}

	got, err := store.GetEventsFromVersion(ctx, "agg-1", 1)
	if err != nil {
		t.Fatalf("GetEventsFromVersion error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0].Event.GetID() != "e2" || got[1].Event.GetID() != "e3" {
		t.Fatalf("unexpected events: %s, %s", got[0].Event.GetID(), got[1].Event.GetID())
	}

	got, err = store.GetEventsFromVersion(ctx, "agg-1", 3)
	if err != nil {
		t.Fatalf("GetEventsFromVersion error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 events, got %d", len(got))
	}
}

func TestInMemoryEventStore_GetAllEvents(t *testing.T) {
	store := NewInMemoryEventStore()
	ctx := context.Background()

	if err := store.SaveEvents(ctx, "agg-1", []ddd.EventWithMetadata{
		{Event: newTestEvent("e1", "type.a"), AggregateID: "agg-1", Version: 1},
		{Event: newTestEvent("e2", "type.b"), AggregateID: "agg-1", Version: 2},
	}, 0); err != nil {
		t.Fatalf("SaveEvents error: %v", err)
	}

	if err := store.SaveEvents(ctx, "agg-2", []ddd.EventWithMetadata{
		{Event: newTestEvent("e3", "type.c"), AggregateID: "agg-2", Version: 1},
	}, 0); err != nil {
		t.Fatalf("SaveEvents error: %v", err)
	}

	all, err := store.GetAllEvents(ctx)
	if err != nil {
		t.Fatalf("GetAllEvents error: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 events, got %d", len(all))
	}

	found := map[string]bool{}
	for _, ev := range all {
		found[ev.Event.GetID()] = true
	}
	for _, id := range []string{"e1", "e2", "e3"} {
		if !found[id] {
			t.Fatalf("missing event id: %s", id)
		}
	}
}
