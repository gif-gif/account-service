package httpx

import (
	"net/http"
	"strings"

	"account-service/service/internal/admin"
	"account-service/service/internal/callers"

	"github.com/gofiber/fiber/v3"
)

const callerKey = "api_caller"
const adminSessionKey = "admin_session"

type CallerStore interface {
	Authenticate(apiKey string) (callers.Caller, bool)
}

type AdminSessionService interface {
	CurrentSession(c fiber.Ctx) (admin.Session, bool)
}

func APIKeyAuth(store CallerStore) fiber.Handler {
	return func(c fiber.Ctx) error {
		auth := c.Get("Authorization")
		apiKey, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok || strings.TrimSpace(apiKey) == "" {
			return JSONError(c, http.StatusUnauthorized, "unauthorized", "API key is required")
		}
		caller, ok := store.Authenticate(strings.TrimSpace(apiKey))
		if !ok {
			return JSONError(c, http.StatusUnauthorized, "unauthorized", "Invalid API key")
		}
		c.Locals(callerKey, caller)
		return c.Next()
	}
}

func CallerFromContext(c fiber.Ctx) callers.Caller {
	caller, _ := c.Locals(callerKey).(callers.Caller)
	return caller
}

func AdminSession(service AdminSessionService) fiber.Handler {
	return func(c fiber.Ctx) error {
		session, ok := service.CurrentSession(c)
		if !ok {
			return JSONError(c, http.StatusUnauthorized, "unauthorized", "Admin session is required")
		}
		c.Locals(adminSessionKey, session)
		return c.Next()
	}
}

func CORS(allowedOrigins []string) fiber.Handler {
	allowed := map[string]bool{}
	for _, origin := range allowedOrigins {
		allowed[strings.TrimSpace(origin)] = true
	}

	return func(c fiber.Ctx) error {
		origin := c.Get("Origin")
		if allowed[origin] {
			c.Set("Access-Control-Allow-Origin", origin)
			c.Set("Access-Control-Allow-Credentials", "true")
			c.Set("Vary", "Origin")
		}
		if c.Method() == http.MethodOptions {
			c.Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
			c.Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID")
			return c.SendStatus(http.StatusNoContent)
		}
		return c.Next()
	}
}
