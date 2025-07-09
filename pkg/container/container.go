package container

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/sirupsen/logrus"
)

// Container represents a dependency injection container
type Container struct {
	services  map[string]interface{}
	factories map[string]func() interface{}
	mutex     sync.RWMutex
}

// NewContainerWithLogger creates a new container with a pre-configured logger
func NewContainerWithLogger(logger *logrus.Logger) *Container {
	c := NewContainer()
	c.Register("logger", logger)
	return c
}

// NewContainer creates a new dependency injection container
func NewContainer() *Container {
	return &Container{
		services:  make(map[string]interface{}),
		factories: make(map[string]func() interface{}),
	}
}

// Register registers a service instance in the container
func (c *Container) Register(name string, service interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.services[name] = service
}

// RegisterFactory registers a factory function for creating services
func (c *Container) RegisterFactory(name string, factory func() interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.factories[name] = factory
}

// Get retrieves a service from the container
func (c *Container) Get(name string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Check if service instance exists
	if service, exists := c.services[name]; exists {
		return service, nil
	}

	// Check if factory exists
	if factory, exists := c.factories[name]; exists {
		service := factory()
		c.services[name] = service // Cache the instance
		return service, nil
	}

	return nil, fmt.Errorf("service '%s' not found in container", name)
}

// GetTyped retrieves a service from the container with type assertion
func GetTyped[T any](c *Container, name string) (T, error) {
	var zero T
	service, err := c.Get(name)
	if err != nil {
		return zero, err
	}

	typed, ok := service.(T)
	if !ok {
		return zero, fmt.Errorf("service '%s' is not of expected type %v", name, reflect.TypeOf(zero))
	}

	return typed, nil
}

// Has checks if a service is registered in the container
func (c *Container) Has(name string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, hasService := c.services[name]
	_, hasFactory := c.factories[name]
	return hasService || hasFactory
}

// Remove removes a service from the container
func (c *Container) Remove(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.services, name)
	delete(c.factories, name)
}

// Clear removes all services from the container
func (c *Container) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.services = make(map[string]interface{})
	c.factories = make(map[string]func() interface{})
}

// ListServices returns a list of all registered service names
func (c *Container) ListServices() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var names []string
	for name := range c.services {
		names = append(names, name)
	}
	for name := range c.factories {
		if _, exists := c.services[name]; !exists {
			names = append(names, name)
		}
	}
	return names
}
