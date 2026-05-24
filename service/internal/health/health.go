package health

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

type Checker interface {
	Check(context.Context) error
}

type CheckerFunc func(context.Context) error

func (fn CheckerFunc) Check(ctx context.Context) error {
	return fn(ctx)
}

func Register(app *fiber.App, checker Checker) {
	app.Get("/health/live", func(c fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/health/ready", func(c fiber.Ctx) error {
		if err := checker.Check(c.Context()); err != nil {
			return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"status": "unavailable"})
		}
		return c.Status(http.StatusOK).JSON(fiber.Map{"status": "ok"})
	})
}
