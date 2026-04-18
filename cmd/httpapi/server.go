package httpapi

import (
	"context"
	"errors"
	"time"

	"socrati/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// NewApp builds the Fiber app with default middleware and routes.
func NewApp(log logger.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:           5 * time.Second,
		DisableStartupMessage: true,
	})

	// Order matters: RequestID first → InjectRequestID wraps response →
	// RequestLogger logs after handler completes.
	app.Use(RequestID())
	app.Use(InjectRequestID())
	app.Use(RequestLogger(log))

	app.Get("/health", healthHandler)
	app.Post("/ask", askHandler)

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
