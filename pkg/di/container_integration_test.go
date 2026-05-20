package di

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/go-generic-event-driven/pkg/cqrs"
	"github.com/yourusername/go-generic-event-driven/pkg/db"
	"github.com/yourusername/go-generic-event-driven/pkg/logger"
)

// TestContainerIntegration tests the container with real services
func TestContainerIntegration(t *testing.T) {
	container := NewContainer()

	// Register real services
	container.Register("logger", logger.NewStandardLoggerWithPrefix("TEST"))
	container.Register("eventStore", db.NewInMemoryEventStore())
	container.Register("commandBus", cqrs.NewInMemoryCommandBus())
	container.Register("queryBus", cqrs.NewInMemoryQueryBus())
	container.Register("eventBus", cqrs.NewInMemoryEventBus())

	// Test that we can retrieve all services
	log, err := container.Get("logger")
	assert.NoError(t, err)
	assert.NotNil(t, log)

	eventStore, err := container.Get("eventStore")
	assert.NoError(t, err)
	assert.NotNil(t, eventStore)

	commandBus, err := container.Get("commandBus")
	assert.NoError(t, err)
	assert.NotNil(t, commandBus)

	queryBus, err := container.Get("queryBus")
	assert.NoError(t, err)
	assert.NotNil(t, queryBus)

	eventBus, err := container.Get("eventBus")
	assert.NoError(t, err)
	assert.NotNil(t, eventBus)
}

// TestContainerWithCQRS tests the container with CQRS system
func TestContainerWithCQRS(t *testing.T) {
	container := NewContainer()

	// Register services
	container.Register("logger", logger.NewStandardLoggerWithPrefix("CQRS-TEST"))
	container.Register("eventStore", db.NewInMemoryEventStore())
	container.Register("commandBus", cqrs.NewInMemoryCommandBus())
	container.Register("queryBus", cqrs.NewInMemoryQueryBus())
	container.Register("eventBus", cqrs.NewInMemoryEventBus())

	// Get services
	commandBus := container.MustGet("commandBus").(cqrs.CommandBus)
	queryBus := container.MustGet("queryBus").(cqrs.QueryBus)
	eventBus := container.MustGet("eventBus").(cqrs.EventBus)

	// Create CQRS system
	cqrsSystem := cqrs.NewCQRS(commandBus, queryBus, eventBus)
	
	assert.NotNil(t, cqrsSystem)
	assert.NotNil(t, cqrsSystem.CommandBus)
	assert.NotNil(t, cqrsSystem.QueryBus)
	assert.NotNil(t, cqrsSystem.EventBus)
}

// TestContainerLazyLoading tests lazy loading of services
func TestContainerLazyLoading(t *testing.T) {
	container := NewContainer()
	
	callCount := 0
	factory := func() interface{} {
		callCount++
		return logger.NewStandardLoggerWithPrefix("LAZY")
	}
	
	container.RegisterLazy("lazyLogger", factory)
	
	// Factory should not be called yet
	assert.Equal(t, 0, callCount)
	
	// Get the service - factory should be called
	service, err := container.GetLazy("lazyLogger")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.NotNil(t, service)
	
	// Get again - factory should not be called again
	service2, err := container.GetLazy("lazyLogger")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount) // Still 1
	assert.Equal(t, service, service2)
}

// TestContainerSingleton tests singleton behavior
func TestContainerSingleton(t *testing.T) {
	container := NewContainer()
	
	callCount := 0
	factory := func() interface{} {
		callCount++
		return logger.NewStandardLoggerWithPrefix("SINGLETON")
	}
	
	// Register singleton multiple times
	container.RegisterSingleton("singletonLogger", factory)
	container.RegisterSingleton("singletonLogger", factory)
	container.RegisterSingleton("singletonLogger", factory)
	
	// Factory should be called only once
	assert.Equal(t, 1, callCount)
	
	// Get the service
	service1, err := container.Get("singletonLogger")
	assert.NoError(t, err)
	assert.NotNil(t, service1)
	
	service2, err := container.Get("singletonLogger")
	assert.NoError(t, err)
	assert.Equal(t, service1, service2)
}