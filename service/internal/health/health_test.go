package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestLiveReturnsOK(t *testing.T) {
	app := fiber.New()
	Register(app, CheckerFunc(func(context.Context) error {
		return nil
	}))

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestReadyReturnsOKWhenDatabaseCheckSucceeds(t *testing.T) {
	app := fiber.New()
	Register(app, CheckerFunc(func(context.Context) error {
		return nil
	}))

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestReadyReturnsUnavailableWhenDatabaseCheckFails(t *testing.T) {
	app := fiber.New()
	databaseUnavailable := errors.New("database unavailable")
	Register(app, CheckerFunc(func(context.Context) error {
		return databaseUnavailable
	}))

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
}
