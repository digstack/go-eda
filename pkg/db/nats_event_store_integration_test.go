//go:build integration

package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/yourusername/go-generic-event-driven/pkg/ddd"
)

type integEvent struct {
	ID        string
	Type      string
	Timestamp time.Time
	Data      interface{}
}

func (e *integEvent) GetID() string           { return e.ID }
func (e *integEvent) GetType() string         { return e.Type }
func (e *integEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *integEvent) GetData() interface{}    { return e.Data }

func newIntegEvent(id, eventType string) ddd.Event {
	return &integEvent{ID: id, Type: eventType, Timestamp: time.Now()}
}

func TestNATSEventStore_SaveAndReplay(t *testing.T) {
	ctx := context.Background()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	defer nc.Close()

	ser := ddd.NewJSONEventSerializer()
	ser.RegisterEventType("evt.type", func() ddd.Event { return &integEvent{} })

	store := NewNATSEventStore(nc, ser)
	aggID := "agg-" + uuid.NewString()
	subject := "EVENTS." + aggID

	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("JetStream context error: %v", err)
	}

	if msg, err := js.GetLastMsg("EVENTS", subject); err == nil {
		t.Fatalf("unexpected existing message seq=%d", msg.Sequence)
	} else if err != nats.ErrMsgNotFound {
		t.Fatalf("GetLastMsg error: %v", err)
	}

	events := []ddd.EventWithMetadata{
		{Event: newIntegEvent(uuid.NewString(), "evt.type")},
		{Event: newIntegEvent(uuid.NewString(), "evt.type")},
	}

	if err := store.SaveEvents(ctx, aggID, events, 0); err != nil {
		t.Fatalf("SaveEvents error: %v", err)
	}

	replayed, err := store.GetEventsFromVersion(ctx, aggID, 0)
	if err != nil {
		t.Fatalf("GetEventsFromVersion error: %v", err)
	}
	if len(replayed) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(replayed))
	}
}

func TestNATSEventStore_ExpectedVersion(t *testing.T) {
	ctx := context.Background()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	defer nc.Close()

	ser := ddd.NewJSONEventSerializer()
	ser.RegisterEventType("evt.type", func() ddd.Event { return &integEvent{} })

	store := NewNATSEventStore(nc, ser)
	aggID := "agg-" + uuid.NewString()

	if err := store.SaveEvents(ctx, aggID, []ddd.EventWithMetadata{{Event: newIntegEvent(uuid.NewString(), "evt.type")}}, 0); err != nil {
		t.Fatalf("SaveEvents error: %v", err)
	}

	// Expect version mismatch
	if err := store.SaveEvents(ctx, aggID, []ddd.EventWithMetadata{{Event: newIntegEvent(uuid.NewString(), "evt.type")}}, 0); err == nil {
		t.Fatalf("expected version error")
	}
}
