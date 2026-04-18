package httpapi

import (
	"context"
	"errors"
	"time"

	"socrati/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDKey is the locals key for the per-request ID.
const RequestIDKey = "request_id"

// NewApp builds the Fiber app with /health and request logging middleware.
func NewApp(log logger.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:           5 * time.Second,
		DisableStartupMessage: true,
	})

	app.Use(RequestLogger(log))

	app.Get("/health", healthHandler)

	return app
}

// Run starts the Fiber app and blocks until ctx is cancelled, then shuts down gracefully.
func Run(ctx context.Context, app *fiber.App, addr string, log logger.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		log.Info("http server started", "addr", addr)
		if err := app.Listen(addr); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	case <-ctx.Done():
		log.Info("shutting down http server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return app.ShutdownWithContext(shutdownCtx)
	}
}

func healthHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}

// RequestLogger injects a request_id and logs latency_ms for each request.
func RequestLogger(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqID := c.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Set("X-Request-ID", reqID)
		c.Locals(RequestIDKey, reqID)

		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

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
