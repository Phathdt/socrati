package logger

import (
	"context"
	"sync"
)

// MockLogger is a mock implementation of Logger for testing
type MockLogger struct {
	mu       sync.Mutex
	Messages []LogMessage
}

// LogMessage represents a logged message for testing
type LogMessage struct {
	Level   string
	Message string
	Args    []any
}

// NewMock creates a new MockLogger
func NewMock() *MockLogger {
	return &MockLogger{
		Messages: make([]LogMessage, 0),
	}
}

// Debug logs at debug level
func (m *MockLogger) Debug(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, LogMessage{Level: "debug", Message: msg, Args: args})
}

// Info logs at info level
func (m *MockLogger) Info(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, LogMessage{Level: "info", Message: msg, Args: args})
}

// Warn logs at warn level
func (m *MockLogger) Warn(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, LogMessage{Level: "warn", Message: msg, Args: args})
}

// Error logs at error level
func (m *MockLogger) Error(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, LogMessage{Level: "error", Message: msg, Args: args})
}

// Fatal logs at error level (does not exit in mock)
func (m *MockLogger) Fatal(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, LogMessage{Level: "fatal", Message: msg, Args: args})
}

// With returns a new MockLogger (same instance for simplicity)
func (m *MockLogger) With(args ...any) Logger {
	return m
}

// WithGroup returns a new MockLogger (same instance for simplicity)
func (m *MockLogger) WithGroup(name string) Logger {
	return m
}

// DebugContext logs at debug level with context
func (m *MockLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	m.Debug(msg, args...)
}

// InfoContext logs at info level with context
func (m *MockLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	m.Info(msg, args...)
}

// WarnContext logs at warn level with context
func (m *MockLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	m.Warn(msg, args...)
}

// ErrorContext logs at error level with context
func (m *MockLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	m.Error(msg, args...)
}

// Reset clears all logged messages
func (m *MockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = make([]LogMessage, 0)
}

// HasMessage checks if a message with the given level and content exists
func (m *MockLogger) HasMessage(level, msg string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, logMsg := range m.Messages {
		if logMsg.Level == level && logMsg.Message == msg {
			return true
		}
	}
	return false
}

// Count returns the number of logged messages
func (m *MockLogger) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Messages)
}
