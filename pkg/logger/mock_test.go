package logger

import (
	"context"
	"testing"
)

func TestMockLogger(t *testing.T) {
	mock := NewMock()

	// Test logging at different levels
	mock.Debug("debug message", "key", "value")
	mock.Info("info message", "key", "value")
	mock.Warn("warn message", "key", "value")
	mock.Error("error message", "key", "value")

	// Verify messages were logged
	if mock.Count() != 4 {
		t.Errorf("expected 4 messages, got %d", mock.Count())
	}

	// Verify specific messages
	if !mock.HasMessage("debug", "debug message") {
		t.Error("expected debug message not found")
	}
	if !mock.HasMessage("info", "info message") {
		t.Error("expected info message not found")
	}
	if !mock.HasMessage("warn", "warn message") {
		t.Error("expected warn message not found")
	}
	if !mock.HasMessage("error", "error message") {
		t.Error("expected error message not found")
	}
}

func TestMockLoggerReset(t *testing.T) {
	mock := NewMock()

	// Log some messages
	mock.Info("test message")
	if mock.Count() != 1 {
		t.Errorf("expected 1 message, got %d", mock.Count())
	}

	// Reset and verify
	mock.Reset()
	if mock.Count() != 0 {
		t.Errorf("expected 0 messages after reset, got %d", mock.Count())
	}
}

func TestMockLoggerContext(t *testing.T) {
	mock := NewMock()
	ctx := context.Background()

	// Test context-aware logging
	mock.DebugContext(ctx, "debug message")
	mock.InfoContext(ctx, "info message")
	mock.WarnContext(ctx, "warn message")
	mock.ErrorContext(ctx, "error message")

	if mock.Count() != 4 {
		t.Errorf("expected 4 messages, got %d", mock.Count())
	}
}

func TestMockLoggerWith(t *testing.T) {
	mock := NewMock()

	// Test With returns Logger interface
	logger := mock.With("component", "test")
	if logger == nil {
		t.Fatal("expected non-nil logger from With")
	}

	// Test WithGroup returns Logger interface
	logger = mock.WithGroup("app")
	if logger == nil {
		t.Fatal("expected non-nil logger from WithGroup")
	}
}

func TestMockLoggerFatal(t *testing.T) {
	mock := NewMock()

	// Test Fatal doesn't exit in mock
	mock.Fatal("fatal message")

	if !mock.HasMessage("fatal", "fatal message") {
		t.Error("expected fatal message not found")
	}
}
