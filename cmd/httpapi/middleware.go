package httpapi

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"socrati/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDKey is the locals key for the per-request ID.
const RequestIDKey = "request_id"

// RequestID assigns or reuses an X-Request-ID and stores it in locals.
// Must run before InjectRequestID and RequestLogger.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqID := c.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Set("X-Request-ID", reqID)
		c.Locals(RequestIDKey, reqID)
		return c.Next()
	}
}

// InjectRequestID rewrites JSON object responses to include "request_id".
// Non-JSON responses or non-object payloads (arrays, scalars) pass through unchanged.
func InjectRequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			return err
		}

		ct := string(c.Response().Header.ContentType())
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return nil
		}

		body := c.Response().Body()
		trimmed := bytes.TrimSpace(body)
		if len(trimmed) == 0 || trimmed[0] != '{' {
			return nil
		}

		var obj map[string]any
		if err := json.Unmarshal(trimmed, &obj); err != nil {
			return nil
		}

		reqID, _ := c.Locals(RequestIDKey).(string)
		obj["request_id"] = reqID

		patched, err := json.Marshal(obj)
		if err != nil {
			return nil
		}
		c.Response().SetBody(patched)
		c.Response().Header.SetContentLength(len(patched))
		return nil
	}
}

// RequestLogger logs method, path, status and latency_ms for each request.
func RequestLogger(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		reqID, _ := c.Locals(RequestIDKey).(string)
		log.Info("http_request",
			"request_id", reqID,
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", latency.Milliseconds(),
		)
		return err
	}
}
