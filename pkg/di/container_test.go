package di

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestContainer(t *testing.T) {
	container := NewContainer()

	// Test registration and retrieval
	logger := &MockLogger{}
	container.Register("logger", logger)

	retrieved, err := container.Get("logger")
	assert.NoError(t, err)
	assert.Equal(t, logger, retrieved)

	// Test GetOrDefault
	defaultLogger := &MockLogger{}
	result := container.GetOrDefault("nonexistent", defaultLogger)
	assert.Equal(t, defaultLogger, result)

	// Test MustGet with existing service
	result = container.MustGet("logger")
	assert.Equal(t, logger, result)

	// Test MustGet with non-existing service (should panic)
	assert.Panics(t, func() {
		container.MustGet("nonexistent")
	})
}

func TestSingleton(t *testing.T) {
	container := NewContainer()
	
	callCount := 0
	factory := func() interface{} {
		callCount++
		return &MockLogger{}
	}
	
	container.RegisterSingleton("logger", factory)
	container.RegisterSingleton("logger", factory) // Should not call factory again
	
	assert.Equal(t, 1, callCount)
}

func TestLazyRegistration(t *testing.T) {
	container := NewContainer()
	
	callCount := 0
	factory := func() interface{} {
		callCount++
		return &MockLogger{}
	}
	
	container.RegisterLazy("logger", factory)
	
	// Factory should not be called yet
	assert.Equal(t, 0, callCount)
	
	// Get the service
	service, err := container.GetLazy("logger")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
	
	// Get again - factory should not be called again
	service2, err := container.GetLazy("logger")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.Equal(t, service, service2)
}

type MockLogger struct{}

func (m *MockLogger) Log(message string) {}
