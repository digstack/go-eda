# Go Generic Event-Driven Boilerplate

Modern Go boilerplate for **event-driven**, **event-sourced**, **CQRS**-based applications. No business logic — only patterns, primitives and infrastructure you can compose.

## Why this exists

The same building blocks (typed event store, OCC, idempotent publish, typed CQRS, projections, observability) keep getting reinvented per service. This repo factors them out into reusable packages with a small surface, generics, and tests.

## Layout

```
pkg/
├── ddd/        # generic Aggregate[ID], EventEnvelope[ID], Clock, payload contract
├── db/         # event stores: JetStreamStore[ID] (NATS v2), InMemoryStore[ID]; SnapshotStore
├── cqrs/       # typed CommandHandler[C,R] / QueryHandler[Q,R], middleware chain, TypedEventBus
├── di/         # type-safe Registry: Provide[T] / Resolve[T], scopes, lifecycle hooks
├── obs/        # slog logger, Meter/Tracer interfaces, cqrs middlewares (logging/metrics/tracing)
├── outbox/     # transactional outbox pattern: Store interface + Relay + in-mem impl
├── projection/ # event-store projections: Projector + Manager (catch-up + live + checkpoint)
├── processmanager/ # event-driven process manager (state machine + OCC + idempotency)
├── logger/     # Logger interface + slog adapter
├── module/     # legacy module registry (string-keyed, kept for compatibility)
└── types/      # legacy generic payload structs (kept for compatibility)
examples/
└── banking/    # runnable example exercising every modern layer
```

The `pkg/cqrs`, `pkg/ddd` and `pkg/di` packages contain both a **legacy** API (string-keyed, `interface{}`) and a **typed** API (generics). The typed API is the recommended one; the legacy one is preserved while consumers migrate.

## Quick start

```bash
go run ./examples/banking
```

Expected output ends with: `account=<uuid> balance=350 (expected 350)`.

## Building blocks

### Event envelopes (pkg/ddd)

Every event flows through `EventEnvelope[ID]` — a stable, JSON-serializable shape with:

- `EventID` (idempotency key) and `EventType` (stable kind string)
- aggregate identity (`AggregateID`, `AggregateType`, `AggregateVersion`)
- correlation/causation/tenant IDs
- a typed `Payload` implementing `ddd.EventPayload` (i.e. `EventKind() string`)

Aggregates embed `BaseAggregateRoot[ID]` and use `ddd.Raise(...)` to emit events; replay goes through `ddd.LoadFromHistory(...)`.

### Event store (pkg/db)

Two implementations behind one mental model:

| Store | Use case | Notes |
|---|---|---|
| `InMemoryStore[ID]` | tests, local dev | OCC enforced, contiguous versions, optional `Subscribe` |
| `JetStreamStore[ID]` | production | NATS JetStream v2 API, OCC via `ExpectLastSequencePerSubject`, idempotent publish via `Nats-Msg-Id`, ordered consumer replay |

Snapshots are first-class: `SnapshotStore[ID]` with an in-memory impl and a KV (NATS JetStream) impl.

### CQRS (pkg/cqrs)

Typed handlers, dispatched by Go type — no string lookup at call sites:

```go
cqrs.RegisterCommandHandler[OpenAccountCmd, OpenAccountRes](router, openHandler)
res, err := cqrs.Execute[OpenAccountCmd, OpenAccountRes](ctx, router, OpenAccountCmd{Owner: "Alice"})
```

Middlewares compose with `cqrs.Chain(...)`. Bundled: `RecoveryMiddleware`, `TimeoutMiddleware`, `RetryMiddleware`. From `pkg/obs`: `LoggingMiddleware`, `MetricsMiddleware`, `TracingMiddleware`.

Typed errors: `ErrNotFound`, `ErrConcurrencyConflict`, `ErrValidation`, `ErrTimeout`, `ErrHandlerPanic`, `ErrNoHandler`.

### Typed DI (pkg/di)

```go
r := di.New()
di.Provide(r, func(_ *di.Resolver) (logger.Logger, error) { return logger.NewJSONSlogLogger(slog.LevelInfo), nil })
di.Provide(r, func(rv *di.Resolver) (*UserService, error) {
    log := di.MustFrom[logger.Logger](rv)
    return &UserService{log: log}, nil
})
svc := di.MustResolve[*UserService](r)
```

Features:

- generics — no casts, no string keys
- lifetimes: `Singleton`, `Transient`, `Scoped` (per-request / per-tenant via `Registry.NewScope()`)
- qualifiers: `ProvideTagged` / `ResolveTagged` for multiple impls of the same interface
- cycle detection at resolve time
- lifecycle hooks: services implementing `di.Lifecycle` get `OnStart`/`OnStop` ordered by build order (leaves → roots; stop reversed)

### Outbox (pkg/outbox)

Transactional outbox: enqueue an envelope inside your application transaction, then a `Relay` worker drains pending records and publishes them through a `Publisher`. The `Store` interface is storage-agnostic — implement it on top of SQL `SELECT ... FOR UPDATE SKIP LOCKED`, MongoDB, or anything else. An in-memory `Store` is shipped for tests.

### Projections (pkg/projection)

A `Manager[ID]` drives a `Projector` through two phases under one API: **catch-up** (replay history from the last checkpoint) then **live tail** (subscribe to new events). `CheckpointStore` persists `Checkpoint[ID]` (last-seen version per aggregate); in-memory implementation provided.

### Process manager (pkg/processmanager)

A generic, idempotent state machine: declare a `Definition[ID,S]` with an `InstanceID` derivation function, an `Initial` state, and a routing table of `Handlers` per event kind. The `Engine` loads/initializes the instance, runs the handler, persists with optimistic concurrency, and executes the returned `Effects`. Replay-safe by `LastEventID`.

### Observability (pkg/obs)

- `Meter` / `Tracer` interfaces with `Nop*` defaults — plug Prometheus or OpenTelemetry adapters when ready
- ready-made middlewares for the typed CQRS dispatch path
- `logger.SlogLogger` bridges `*slog.Logger` to the boilerplate `Logger` interface

## Tests

```bash
go test -race ./...
```

The JetStream-backed store has its own integration test that requires a local NATS instance (skipped by default; see `pkg/db`).

## Status

See `REVIEW_NOTES.md` for the historical audit and what is now resolved.

## License

See LICENSE.
