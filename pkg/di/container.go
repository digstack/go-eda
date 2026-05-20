package di

import (
	"fmt"
	"sync"
)

// Container is a simple dependency injection container
type Container struct {
	services map[string]interface{}
	mu       sync.RWMutex
}

// NewContainer creates a new DI container
func NewContainer() *Container {
	return &Container{
		services: make(map[string]interface{}),
	}
}

// Register adds a service to the container
func (c *Container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

// Get retrieves a service from the container
func (c *Container) Get(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	service, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}
	return service, nil
}

// GetOrDefault retrieves a service or returns a default value
func (c *Container) GetOrDefault(name string, defaultValue interface{}) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	service, exists := c.services[name]
	if !exists {
		return defaultValue
	}
	return service
}

// MustGet retrieves a service or panics if not found
func (c *Container) MustGet(name string) interface{} {
	service, err := c.Get(name)
	if err != nil {
		panic(err)
	}
	return service
}

// RegisterSingleton registers a service that will be created only once
func (c *Container) RegisterSingleton(name string, factory func() interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if _, exists := c.services[name]; !exists {
		c.services[name] = factory()
	}
}

// RegisterLazy registers a factory function that will be called when the service is first requested
func (c *Container) RegisterLazy(name string, factory func() interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = factory
}

// GetLazy retrieves a service, calling the factory if it's a lazy registration
func (c *Container) GetLazy(name string) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	service, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}
	
	// If it's a factory function, call it
	if factory, ok := service.(func() interface{}); ok {
		service = factory()
		c.services[name] = service // Cache the result
	}
	
	return service, nil
}