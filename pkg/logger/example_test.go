package logger_test

import (
	"fmt"

	"socrati/pkg/logger"
)

// Example demonstrates basic logger usage
func ExampleNew() {
	log := logger.New("text", "info")
	log.Info("application started", "version", "1.0.0")
	// Output varies due to timestamp
}

// Example demonstrates using mock logger in tests
func ExampleNewMock() {
	// Create a mock logger for testing
	mock := logger.NewMock()

	// Use it like a regular logger
	mock.Info("processing request", "id", 123)
	mock.Error("error occurred", logger.String("error", "connection failed"))

	// Assert on logged messages
	if mock.HasMessage("info", "processing request") {
		fmt.Println("Info message logged")
	}

	// Check count
	fmt.Printf("Total messages: %d\n", mock.Count())

	// Output:
	// Info message logged
	// Total messages: 2
}

// Example demonstrates logger with attributes
func ExampleLogger_With() {
	log := logger.New("text", "info")

	// Create a logger with pre-set attributes
	serviceLog := log.With("service", "bridge", "version", "1.0")

	// All subsequent logs will include these attributes
	serviceLog.Info("service started")
	serviceLog.Info("processing request")
	// Output varies due to timestamp
}

// Example demonstrates dependency injection pattern
func ExampleLogger_dependencyInjection() {
	// This demonstrates how to use Logger interface in your services
	type MyService struct {
		logger logger.Logger
	}

	// NewMyService accepts Logger interface, not concrete type
	NewMyService := func(log logger.Logger) *MyService {
		return &MyService{
			logger: log.With("service", "example"),
		}
	}

	// Can use mock logger in tests
	mockLogger := logger.NewMock()
	testService := NewMyService(mockLogger)
	testService.logger.Info("test service")

	// Verify mock logged the message
	if mockLogger.HasMessage("info", "test service") {
		fmt.Println("Test service logged correctly")
	}

	// In production, use real logger:
	// realLogger := logger.New("text", "info")
	// prodService := NewMyService(realLogger)

	// Output:
	// Test service logged correctly
}
