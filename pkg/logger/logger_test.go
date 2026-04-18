package logger

import (
	"context"
	"errors"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		format string
		level  string
	}{
		{"json debug", "json", "debug"},
		{"json info", "json", "info"},
		{"text debug", "text", "debug"},
		{"text info", "text", "info"},
		{"text warn", "text", "warn"},
		{"text error", "text", "error"},
		{"unknown level defaults to info", "text", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.format, tt.level)
			if l == nil {
				t.Fatal("expected non-nil logger")
			}
			// Verify it's a SlogLogger implementation
			if _, ok := l.(*SlogLogger); !ok {
				t.Fatal("expected SlogLogger implementation")
			}
		})
	}
}

func TestLoggerWith(t *testing.T) {
	l := New("text", "debug")
	l2 := l.With("component", "test", "version", "1.0")
	if l2 == nil {
		t.Fatal("expected non-nil logger from With")
	}
	l2.Debug("test with component")
}

func TestLoggerWithGroup(t *testing.T) {
	l := New("text", "debug")
	l2 := l.WithGroup("app")
	if l2 == nil {
		t.Fatal("expected non-nil logger from WithGroup")
	}
	l2.Info("test with group")
}

func TestLogLevels(t *testing.T) {
	l := New("text", "debug")

	// Test all log levels
	l.Debug("debug message", "key", "value")
	l.Info("info message", "key", "value")
	l.Warn("warn message", "key", "value")
	l.Error("error message", "key", "value")
}

func TestLogContext(t *testing.T) {
	l := New("text", "debug")
	ctx := context.Background()

	// Test context-aware logging
	l.DebugContext(ctx, "debug message", "key", "value")
	l.InfoContext(ctx, "info message", "key", "value")
	l.WarnContext(ctx, "warn message", "key", "value")
	l.ErrorContext(ctx, "error message", "key", "value")
}

func TestHelperFunctions(t *testing.T) {
	l := New("text", "debug")

	// Test all helper functions
	l.Info("test helpers",
		String("str", "value"),
		Int("int", 42),
		Int64("int64", int64(123)),
		Bool("bool", true),
		Any("any", map[string]string{"key": "value"}),
		Err(errors.New("test error")),
	)
}

func TestChaining(t *testing.T) {
	l := New("json", "debug")

	// Test chaining With
	l2 := l.With("service", "bridge")
	l3 := l2.With("handler", "deposit")

	if l3 == nil {
		t.Fatal("expected non-nil logger after chaining")
	}

	l3.Info("chained logger test", "id", 123)
}
