# Go Generic Event-Driven Boilerplate

A reusable, generic Go boilerplate for event-driven applications using CQRS, Domain-Driven Design (DDD), and event sourcing patterns.

## 🎯 Purpose

This boilerplate provides **generic interfaces and base implementations** without any business logic, making it reusable across different domains and applications.

## 📦 Architecture

```
pkg/
├── ddd/          # Generic DDD interfaces and base implementations
├── cqrs/         # Generic CQRS pattern implementation
├── db/           # Generic database interfaces (NATS, in-memory)
├── logger/       # Generic logging interfaces
├── module/       # Generic module system
└── types/        # Generic types and payloads
```

## 🚀 Features Validated ✅

- ✅ **Generic CQRS pattern** - Command/Query separation with in-memory buses
- ✅ **Generic DDD interfaces** - Aggregates, Events, Repositories without business logic
- ✅ **Generic module system** - Register and organize your domain modules
- ✅ **Generic event handling** - Event-driven architecture with handlers
- ✅ **Generic types and payloads** - Reusable entities, commands, queries
- ✅ **Generic logging** - Pluggable logging system
- ✅ **Reusable without business logic** - Clean separation from domain specifics

## 📋 Quick Start

### Installation

```bash
go mod init your-project
go get github.com/yourusername/go-generic-event-driven
```

### Basic Usage

```go
package main

import (
    "context"
    "github.com/yourusername/go-generic-event-driven/pkg/cqrs"
    "github.com/yourusername/go-generic-event-driven/pkg/module"
    "github.com/yourusername/go-generic-event-driven/pkg/logger"
)

// Define your domain types
type User struct {
    types.Entity
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Create your handlers
type CreateUserHandler struct {
    users map[string]*User
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *cqrs.Command) (*cqrs.CommandResult, error) {
    // Your business logic here
    return &cqrs.CommandResult{
        Events: []ddd.Event{event},
        AggregateID: userID,
        Version: 1,
    }, nil
}

// Setup your application
func main() {
    app := module.NewApplication(logger.NewStandardLogger())
    
    // Create and register your module
    userModule := module.NewModule("user")
    userModule.RegisterCommandHandler("CreateUser", &CreateUserHandler{users})
    
    app.RegisterModule(userModule)
    app.Start(context.Background())
    defer app.Stop()
    
    // Use CQRS
    cmd := cqrs.NewCommand("CreateUser", "user-1", map[string]interface{}{
        "name": "John Doe",
        "email": "john@example.com",
    })
    
    result, _ := app.CQRS.ExecuteCommand(context.Background(), cmd)
    fmt.Println("User created:", result.AggregateID)
}
```

## 🧪 Test the Boilerplate

```bash
git clone https://github.com/yourusername/go-generic-event-driven
cd go-generic-event-driven
go run examples/user_example.go
```

## 🔄 Architecture Principles

### Generic by Design
- **NO business logic** - Only patterns and infrastructure
- **NO domain-specific types** - Use your own User, Product, Order, etc.
- **NO business constraints** - Flexible architecture for any domain

### Extensible by Default
- **Pluggable persistence** - NATS, in-memory, or your own implementation
- **Custom handlers** - Implement CommandHandler, QueryHandler, EventHandler interfaces
- **Modular organization** - Separate your domain into logical modules

### Production Ready
- **Event sourcing support** - Full event store implementation
- **Type safety** - Generic interfaces with compile-time checking
- **Logging** - Built-in structured logging
- **Testing friendly** - In-memory implementations for unit tests

## 📁 Project Structure When Using Boilerplate

```
your-project/
├── go.mod
├── main.go
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── modules/
│   │   ├── user/
│   │   │   ├── user.go          # Domain entity
│   │   │   ├── handlers.go      # CQRS handlers
│   │   │   └── events.go        # Domain events
│   │   └── product/
│   │       ├── product.go
│   │       ├── handlers.go
│   │       └── events.go
│   └── application/
│       └── app.go              # Application setup
└── pkg/
    └── shared/                 # Your shared types
```

## 🎯 What Makes It Generic

| Feature | Generic Boilerplate | Domain-Specific |
|---------|-------------------|------------------|
| Reusability | ✅ High - works for any domain | ❌ Low - tied to specific business |
| Business Logic | ❌ None - pure patterns | ✅ Included - domain rules |
| Extension | ✅ Easy - implement interfaces | ⚠️ Limited - coupled to domain |
| Testing | ✅ Simple - in-memory impl | ⚠️ Complex - domain dependencies |

## 🛠️ Core Interfaces You'll Use

### CQRS Handlers
```go
type CommandHandler interface {
    Handle(ctx context.Context, cmd *Command) (*CommandResult, error)
}

type QueryHandler interface {
    Handle(ctx context.Context, query *Query) (interface{}, error)
}

type EventHandler interface {
    Handle(ctx context.Context, event Event) error
}
```

### DDD Building Blocks
```go
type Event interface {
    GetID() string
    GetType() string
    GetTimestamp() time.Time
    GetData() interface{}
}

type Aggregate interface {
    GetID() string
    GetVersion() int
    GetUncommittedEvents() []Event
    MarkEventsAsCommitted()
}
```

## 📚 Patterns Implemented

1. **Command Query Responsibility Segregation (CQRS)**
2. **Event Sourcing**
3. **Domain-Driven Design (DDD)**
4. **Module Architecture**
5. **Event-Driven Architecture**

## 🚀 Next Steps

1. **Fork and customize** for your specific needs
2. **Add your domain modules** using the generic interfaces
3. **Choose your persistence** (NATS for production, in-memory for development)
4. **Build your application** on top of the solid foundation

## 📝 Dependencies

- `github.com/nats-io/nats.go` - NATS messaging (optional)
- `github.com/google/uuid` - UUID generation

## 🤝 Contributing

This is a **generic boilerplate** - contributions should maintain the generic nature. Add infrastructure, not business logic.

---

**Built for developers who want solid patterns without business constraints.** 🎯