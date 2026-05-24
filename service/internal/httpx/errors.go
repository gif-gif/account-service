package httpx

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const requestIDKey = "request_id"

func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Locals(requestIDKey, requestID)
		c.Set("X-Request-ID", requestID)
		return c.Next()
	}
}

func RequestIDFromContext(c fiber.Ctx) string {
	requestID, _ := c.Locals(requestIDKey).(string)
	return requestID
}

func JSONError(c fiber.Ctx, status int, code string, message string) error {
	requestID := RequestIDFromContext(c)
	if requestID == "" {
		requestID = c.Get("X-Request-ID")
	}
	if requestID == "" {
		requestID = uuid.NewString()
		c.Set("X-Request-ID", requestID)
	}

	if status == 0 {
		status = http.StatusInternalServerError
	}
	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{
			"code":       code,
			"message":    message,
			"request_id": requestID,
		},
	})
}
