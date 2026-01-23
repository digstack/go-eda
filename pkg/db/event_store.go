package db

import (
	"context"
	"fmt"

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
		nats.ReconnectWait(2),
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

	for _, event := range events {
		data, err := s.ser.Serialize(event.Event)
		if err != nil {
			return fmt.Errorf("failed to serialize event: %w", err)
		}

		subject := fmt.Sprintf("%s.%s", s.stream, aggregateID)

		_, err = s.js.Publish(subject, data)
		if err != nil {
			return fmt.Errorf("failed to publish event to NATS: %w", err)
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

	sub, err := s.js.SubscribeSync(subject)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to events: %w", err)
	}
	defer sub.Unsubscribe()

	var events []ddd.EventWithMetadata

	for {
		msg, err := sub.NextMsg(0)
		if err != nil {
			if err == nats.ErrTimeout {
				break
			}
			return nil, fmt.Errorf("failed to get next message: %w", err)
		}

		eventType := msg.Header.Get("Event-Type")
		if eventType == "" {
			continue
		}

		event, err := s.ser.Deserialize(eventType, msg.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize event: %w", err)
		}

		eventWithMetadata := ddd.EventWithMetadata{
			Event:       event,
			AggregateID: aggregateID,
			Version:     fromVersion + len(events) + 1,
			Metadata:    make(map[string]interface{}),
		}

		events = append(events, eventWithMetadata)
	}

	return events, nil
}

func (s *NATSEventStore) GetAllEvents(ctx context.Context) ([]ddd.EventWithMetadata, error) {
	if s.js == nil {
		return nil, fmt.Errorf("JetStream not available")
	}

	streamInfo, err := s.js.StreamInfo(s.stream)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream info: %w", err)
	}

	var allEvents []ddd.EventWithMetadata

	consumerSub, err := s.js.PullSubscribe(s.stream+".CONSUMER", "all-events-consumer")
	if err != nil {
		return nil, fmt.Errorf("failed to create pull consumer: %w", err)
	}
	defer consumerSub.Unsubscribe()

	for i := uint64(0); i < streamInfo.State.Msgs; i++ {
		msg, err := consumerSub.NextMsg(0)
		if err != nil {
			if err == nats.ErrTimeout {
				break
			}
			return nil, fmt.Errorf("failed to get message %d: %w", i, err)
		}

		eventType := msg.Header.Get("Event-Type")
		if eventType == "" {
			continue
		}

		event, err := s.ser.Deserialize(eventType, msg.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize event: %w", err)
		}

		eventWithMetadata := ddd.EventWithMetadata{
			Event:    event,
			Metadata: make(map[string]interface{}),
		}

		allEvents = append(allEvents, eventWithMetadata)
		msg.Ack()
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
