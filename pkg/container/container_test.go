package container

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Test interfaces and structs
type TestService interface {
	GetName() string
}

type ConcreteTestService struct {
	name string
}

func (s *ConcreteTestService) GetName() string {
	return s.name
}

type TestServiceWithDependency struct {
	dependency TestService
}

func (s *TestServiceWithDependency) GetDependency() TestService {
	return s.dependency
}

func TestNewContainer(t *testing.T) {
	container := NewContainer()
	assert.NotNil(t, container)
	assert.IsType(t, &Container{}, container)
	assert.NotNil(t, container.services)
	assert.NotNil(t, container.factories)
}

func TestNewContainerWithLogger(t *testing.T) {
	logger := logrus.New()
	container := NewContainerWithLogger(logger)
	assert.NotNil(t, container)
	assert.IsType(t, &Container{}, container)
	assert.NotNil(t, container.services)
	assert.NotNil(t, container.factories)
	// Verify logger is registered as a service
	retrievedLogger, err := container.Get("logger")
	assert.NoError(t, err)
	assert.Equal(t, logger, retrievedLogger)
}

func TestContainer_Register(t *testing.T) {
	container := NewContainer()
	service := &ConcreteTestService{name: "test"}

	// Test successful registration
	container.Register("test-service", service)
	assert.True(t, container.Has("test-service"))

	// Test duplicate registration (should allow overwriting)
	container.Register("test-service", service)
	assert.True(t, container.Has("test-service"))

	// Test registration with nil service
	container.Register("nil-service", nil)
	assert.True(t, container.Has("nil-service"))

	// Test registration with empty name
	container.Register("", service)
	assert.True(t, container.Has(""))
}

func TestContainer_RegisterFactory(t *testing.T) {
	container := NewContainer()
	factory := func() interface{} {
		return &ConcreteTestService{name: "factory-created"}
	}

	// Test successful factory registration
	container.RegisterFactory("factory-service", factory)
	assert.True(t, container.Has("factory-service"))

	// Test duplicate factory registration (should allow overwriting)
	container.RegisterFactory("factory-service", factory)
	assert.True(t, container.Has("factory-service"))

	// Test registration with nil factory
	container.RegisterFactory("nil-factory", nil)
	assert.True(t, container.Has("nil-factory"))

	// Test registration with empty name
	container.RegisterFactory("", factory)
	assert.True(t, container.Has(""))
}

func TestContainer_Get(t *testing.T) {
	container := NewContainer()
	service := &ConcreteTestService{name: "test"}
	container.Register("test-service", service)

	// Test successful retrieval
	retrieved, err := container.Get("test-service")
	assert.NoError(t, err)
	assert.Equal(t, service, retrieved)

	// Test retrieval of non-existent service
	retrieved, err = container.Get("non-existent")
	assert.Error(t, err)
	assert.Nil(t, retrieved)
		assert.Contains(t, err.Error(), "service 'non-existent' not found")

	// Test retrieval with empty name
	retrieved, err = container.Get("")
	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.Contains(t, err.Error(), "not found")
}

func TestContainer_GetTyped(t *testing.T) {
	container := NewContainer()
	logger := logrus.New()
	container.Register("logger", logger)

	// Test successful type assertion
	retrievedLogger, err := GetTyped[*logrus.Logger](container, "logger")
	assert.NoError(t, err)
	assert.Equal(t, logger, retrievedLogger)

	// Test type assertion failure
	_, err = GetTyped[string](container, "logger")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not of expected type")

	// Test non-existent service
	_, err = GetTyped[string](container, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestContainer_GetWithFactory(t *testing.T) {
	container := NewContainer()
	factory := func() interface{} {
		return &ConcreteTestService{name: "factory-created"}
	}
	container.RegisterFactory("factory-service", factory)

	// Test successful retrieval from factory
	retrieved, err := container.Get("factory-service")
	assert.NoError(t, err)
	assert.IsType(t, &ConcreteTestService{}, retrieved)
	assert.Equal(t, "factory-created", retrieved.(*ConcreteTestService).GetName())

	// Test that subsequent calls return the same cached instance
	retrieved2, err := container.Get("factory-service")
	assert.NoError(t, err)
	assert.Same(t, retrieved, retrieved2) // Same cached instance
}

func TestContainer_GetWithFactoryError(t *testing.T) {
	container := NewContainer()
	factory := func() interface{} {
		panic("factory error")
	}
	container.RegisterFactory("error-factory", factory)

	// Test factory panicking
	assert.Panics(t, func() {
		container.Get("error-factory")
	})
}

func TestContainer_Has(t *testing.T) {
	container := NewContainer()
	service := &ConcreteTestService{name: "test"}
	container.Register("test-service", service)

	factory := func() interface{} {
		return &ConcreteTestService{name: "factory-created"}
	}
	container.RegisterFactory("factory-service", factory)

	// Test existing service
	assert.True(t, container.Has("test-service"))

	// Test existing factory
	assert.True(t, container.Has("factory-service"))

	// Test non-existent service
	assert.False(t, container.Has("non-existent"))

	// Test empty name
	assert.False(t, container.Has(""))
}

func TestContainer_Remove(t *testing.T) {
	container := NewContainer()
	service := &ConcreteTestService{name: "test"}
	container.Register("test-service", service)

	factory := func() interface{} {
		return &ConcreteTestService{name: "factory-created"}
	}
	container.RegisterFactory("factory-service", factory)

	// Test removing existing service
	assert.True(t, container.Has("test-service"))
	container.Remove("test-service")
	assert.False(t, container.Has("test-service"))

	// Test removing existing factory
	assert.True(t, container.Has("factory-service"))
	container.Remove("factory-service")
	assert.False(t, container.Has("factory-service"))

	// Test removing non-existent service
	container.Remove("non-existent")
	assert.False(t, container.Has("non-existent"))

	// Test removing with empty name
	container.Remove("")
	assert.False(t, container.Has(""))
}

func TestContainer_Clear(t *testing.T) {
	container := NewContainer()
	service1 := &ConcreteTestService{name: "test1"}
	service2 := &ConcreteTestService{name: "test2"}
	container.Register("service1", service1)
	container.Register("service2", service2)

	factory := func() interface{} {
		return &ConcreteTestService{name: "factory-created"}
	}
	container.RegisterFactory("factory-service", factory)

	// Verify services exist
	assert.True(t, container.Has("service1"))
	assert.True(t, container.Has("service2"))
	assert.True(t, container.Has("factory-service"))

	// Clear all services
	container.Clear()

	// Verify all services are removed
	assert.False(t, container.Has("service1"))
	assert.False(t, container.Has("service2"))
	assert.False(t, container.Has("factory-service"))
}

func TestContainer_ListServices(t *testing.T) {
	container := NewContainer()
	service1 := &ConcreteTestService{name: "test1"}
	service2 := &ConcreteTestService{name: "test2"}
	container.Register("service1", service1)
	container.Register("service2", service2)

	factory := func() interface{} {
		return &ConcreteTestService{name: "factory-created"}
	}
	container.RegisterFactory("factory-service", factory)

	// Test listing services
	services := container.ListServices()
	assert.Len(t, services, 3)
	assert.Contains(t, services, "service1")
	assert.Contains(t, services, "service2")
	assert.Contains(t, services, "factory-service")

	// Test empty container
	container.Clear()
	services = container.ListServices()
	assert.Len(t, services, 0)
}

func TestContainer_ConcurrentAccess(t *testing.T) {
	container := NewContainer()
	service := &ConcreteTestService{name: "concurrent-test"}

	// Test concurrent registration and retrieval
	done := make(chan bool, 2)

	// Goroutine 1: Register service
	go func() {
		container.Register("concurrent-service", service)
		done <- true
	}()

	// Goroutine 2: Try to get service (might fail if not registered yet)
	go func() {
		// Wait a bit to ensure registration happens first
		// In real scenarios, proper synchronization would be used
		for i := 0; i < 100; i++ {
			if container.Has("concurrent-service") {
				retrieved, err := container.Get("concurrent-service")
				if err == nil {
					assert.Equal(t, service, retrieved)
					break
				}
			}
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}

func TestContainer_TypeAssertionError(t *testing.T) {
	container := NewContainer()
	service := "string-service" // Register a string instead of TestService
	container.Register("string-service", service)

	// Test type assertion failure
	_, err := GetTyped[TestService](container, "string-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not of expected type")
}