package httpapi

import "github.com/gofiber/fiber/v2"

type askRequest struct {
	Question string `json:"question"`
}

type askResponse struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// askHandler is a mock implementation that echoes the question back with a stub answer.
func askHandler(c *fiber.Ctx) error {
	var req askRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
	}
	if req.Question == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "question is required"})
	}

	return c.Status(fiber.StatusOK).JSON(askResponse{
		Question: req.Question,
		Answer:   "mock answer: " + req.Question,
	})
}
