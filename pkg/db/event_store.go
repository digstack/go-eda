package db

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/yourusername/go-generic-event-driven/pkg/ddd"
)

// Connection represents a database connection
type Connection interface {
	Close() error
	IsConnected() bool
}

// NATSConnection represents a NATS connection
type NATSConnection struct {
	conn *nats.Conn
}

func NewNATSConnection(url string) (*NATSConnection, error) {
	nc, err := nats.Connect(url,
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(10),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			fmt.Printf("NATS disconnected: %v\n", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("NATS reconnected to %s\n", nc.ConnectedUrl())
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NATSConnection{conn: nc}, nil
}

func (c *NATSConnection) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

func (c *NATSConnection) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

func (c *NATSConnection) GetNATSConn() *nats.Conn {
	return c.conn
}

// EventStore represents an event store interface
type EventStore interface {
	SaveEvents(ctx context.Context, aggregateID string, events []ddd.EventWithMetadata, expectedVersion int) error
	GetEvents(ctx context.Context, aggregateID string) ([]ddd.EventWithMetadata, error)
	GetEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]ddd.EventWithMetadata, error)
	GetAllEvents(ctx context.Context) ([]ddd.EventWithMetadata, error)
}

// NATSEventStore implements event store using NATS JetStream
type NATSEventStore struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	ser    ddd.EventSerializer
	stream string
}

func NewNATSEventStore(conn *nats.Conn, serializer ddd.EventSerializer) *NATSEventStore {
	js, _ := conn.JetStream()

	if js != nil {
		_, err := js.StreamInfo("EVENTS")
		if err != nil {
			_, _ = js.AddStream(&nats.StreamConfig{
				Name:     "EVENTS",
				Subjects: []string{"EVENTS.*"},
			})
		}
	}

	return &NATSEventStore{
		conn:   conn,
		js:     js,
		ser:    serializer,
		stream: "EVENTS",
	}
}

func (s *NATSEventStore) SaveEvents(ctx context.Context, aggregateID string, events []ddd.EventWithMetadata, expectedVersion int) error {
	if s.js == nil {
		return fmt.Errorf("JetStream not available")
	}

	currentVersion, lastSubjectSeq, err := s.getAggregateState(aggregateID)
	if err != nil {
		return fmt.Errorf("failed to get aggregate state: %w", err)
	}
	if expectedVersion >= 0 && expectedVersion != currentVersion {
		if lastSubjectSeq > 0 && currentVersion == 0 {
			return fmt.Errorf("missing aggregate version header for %s", aggregateID)
		}
		return fmt.Errorf("unexpected version: expected %d, got %d", expectedVersion, currentVersion)
	}

	nextVersion := currentVersion + 1
	expectedSubjSeq := lastSubjectSeq
	for _, event := range events {
		data, err := s.ser.Serialize(event.Event)
		if err != nil {
			return fmt.Errorf("failed to serialize event: %w", err)
		}

		subject := fmt.Sprintf("%s.%s", s.stream, aggregateID)

		msg := &nats.Msg{
			Subject: subject,
			Data:    data,
			Header:  nats.Header{},
		}
		msg.Header.Set("Event-Type", event.Event.GetType())
		msg.Header.Set("Aggregate-ID", aggregateID)
		msg.Header.Set("Aggregate-Version", strconv.Itoa(nextVersion))
		nextVersion++

		var opts []nats.PubOpt
		if eventID := event.Event.GetID(); eventID != "" {
			opts = append(opts, nats.MsgId(eventID))
		}
		if expectedVersion >= 0 && expectedSubjSeq > 0 {
			opts = append(opts, nats.ExpectLastSequencePerSubject(expectedSubjSeq))
		}

		pa, err := s.js.PublishMsg(msg, opts...)
		if err != nil {
			return fmt.Errorf("failed to publish event to NATS: %w", err)
		}
		if expectedVersion >= 0 && pa != nil {
			expectedSubjSeq = pa.Sequence
		}
	}

	return nil
}

func (s *NATSEventStore) GetEvents(ctx context.Context, aggregateID string) ([]ddd.EventWithMetadata, error) {
	return s.GetEventsFromVersion(ctx, aggregateID, 0)
}

func (s *NATSEventStore) GetEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]ddd.EventWithMetadata, error) {
	if s.js == nil {
		return nil, fmt.Errorf("JetStream not available")
	}

	subject := fmt.Sprintf("%s.%s", s.stream, aggregateID)

	startSeq := uint64(fromVersion + 1)
	sub, err := s.js.PullSubscribe(subject, "", nats.BindStream(s.stream), nats.StartSequence(startSeq), nats.AckNone())
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to events: %w", err)
	}
	defer sub.Unsubscribe()

	var events []ddd.EventWithMetadata

	seen := 0
	for {
		msgs, err := sub.Fetch(50, nats.MaxWait(200*time.Millisecond))
		if err != nil {
			if err == nats.ErrTimeout {
				break
			}
			return nil, fmt.Errorf("failed to fetch messages: %w", err)
		}
		for _, msg := range msgs {
			eventType := msg.Header.Get("Event-Type")
			if eventType == "" {
				continue
			}

			version := parseVersionHeader(msg.Header.Get("Aggregate-Version"), seen+1)
			seen++
			if version <= fromVersion {
				continue
			}

			event, err := s.ser.Deserialize(eventType, msg.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to deserialize event: %w", err)
			}

			eventWithMetadata := ddd.EventWithMetadata{
				Event:       event,
				AggregateID: aggregateID,
				Version:     version,
				Metadata:    make(map[string]interface{}),
			}

			events = append(events, eventWithMetadata)
		}
	}

	return events, nil
}

func (s *NATSEventStore) GetAllEvents(ctx context.Context) ([]ddd.EventWithMetadata, error) {
	if s.js == nil {
		return nil, fmt.Errorf("JetStream not available")
	}

	var allEvents []ddd.EventWithMetadata

	consumerSub, err := s.js.PullSubscribe(s.stream+".*", "", nats.BindStream(s.stream), nats.StartSequence(1), nats.AckNone())
	if err != nil {
		return nil, fmt.Errorf("failed to create pull consumer: %w", err)
	}
	defer consumerSub.Unsubscribe()

	seen := 0
	for {
		msgs, err := consumerSub.Fetch(100, nats.MaxWait(200*time.Millisecond))
		if err != nil {
			if err == nats.ErrTimeout {
				break
			}
			return nil, fmt.Errorf("failed to fetch messages: %w", err)
		}
		for _, msg := range msgs {
			eventType := msg.Header.Get("Event-Type")
			if eventType == "" {
				continue
			}

			event, err := s.ser.Deserialize(eventType, msg.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to deserialize event: %w", err)
			}

			aggID := msg.Header.Get("Aggregate-ID")
			version := parseVersionHeader(msg.Header.Get("Aggregate-Version"), seen+1)
			seen++

			eventWithMetadata := ddd.EventWithMetadata{
				Event:       event,
				AggregateID: aggID,
				Version:     version,
				Metadata:    make(map[string]interface{}),
			}

			allEvents = append(allEvents, eventWithMetadata)
		}
	}

	return allEvents, nil
}

// InMemoryEventStore implements event store using in-memory storage
type InMemoryEventStore struct {
	events map[string][]ddd.EventWithMetadata
}

func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		events: make(map[string][]ddd.EventWithMetadata),
	}
}

func (s *InMemoryEventStore) SaveEvents(ctx context.Context, aggregateID string, events []ddd.EventWithMetadata, expectedVersion int) error {
	if _, exists := s.events[aggregateID]; !exists {
		s.events[aggregateID] = make([]ddd.EventWithMetadata, 0)
	}

	currentVersion := len(s.events[aggregateID])
	if expectedVersion >= 0 && expectedVersion != currentVersion {
		return fmt.Errorf("unexpected version: expected %d, got %d", expectedVersion, currentVersion)
	}

	nextVersion := currentVersion + 1
	for i := range events {
		events[i].AggregateID = aggregateID
		if events[i].Version == 0 {
			events[i].Version = nextVersion
		}
		nextVersion++
	}

	s.events[aggregateID] = append(s.events[aggregateID], events...)
	return nil
}

func (s *InMemoryEventStore) GetEvents(ctx context.Context, aggregateID string) ([]ddd.EventWithMetadata, error) {
	if events, exists := s.events[aggregateID]; exists {
		return events, nil
	}
	return make([]ddd.EventWithMetadata, 0), nil
}

func (s *InMemoryEventStore) GetEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]ddd.EventWithMetadata, error) {
	events, err := s.GetEvents(ctx, aggregateID)
	if err != nil {
		return nil, err
	}

	if fromVersion >= len(events) {
		return make([]ddd.EventWithMetadata, 0), nil
	}

	return events[fromVersion:], nil
}

func (s *InMemoryEventStore) GetAllEvents(ctx context.Context) ([]ddd.EventWithMetadata, error) {
	var allEvents []ddd.EventWithMetadata
	for _, events := range s.events {
		allEvents = append(allEvents, events...)
	}
	return allEvents, nil
}

func (s *NATSEventStore) getAggregateState(aggregateID string) (int, uint64, error) {
	subject := fmt.Sprintf("%s.%s", s.stream, aggregateID)
	msg, err := s.js.GetLastMsg(s.stream, subject)
	if err != nil {
		if err == nats.ErrMsgNotFound {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	version := parseVersionHeader(msg.Header.Get("Aggregate-Version"), 0)
	return version, msg.Sequence, nil
}

func parseVersionHeader(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return v
}
