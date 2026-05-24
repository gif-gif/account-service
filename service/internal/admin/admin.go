package admin

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"account-service/service/internal/security"

	"github.com/gofiber/fiber/v3"
)

const SessionCookieName = "account_admin_session"

type User struct {
	ID           string
	Username     string
	PasswordHash string
	Status       string
}

type Session struct {
	Token     string
	TokenHash string
	User      User
	ExpiresAt time.Time
}

type MemoryStore struct {
	mu       sync.Mutex
	ttl      time.Duration
	users    map[string]User
	sessions map[string]Session
}

func NewMemoryStore(ttl time.Duration) *MemoryStore {
	return &MemoryStore{
		ttl:      ttl,
		users:    map[string]User{},
		sessions: map[string]Session{},
	}
}

func (store *MemoryStore) AddUser(user User) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.users[user.Username] = user
}

func (store *MemoryStore) FindUser(username string) (User, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	user, ok := store.users[username]
	return user, ok
}

func (store *MemoryStore) CreateSession(user User) (Session, error) {
	token, err := randomToken()
	if err != nil {
		return Session{}, err
	}
	session := Session{
		Token:     token,
		TokenHash: hashSessionToken(token),
		User:      user,
		ExpiresAt: time.Now().Add(store.ttl),
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	store.sessions[session.TokenHash] = session
	return session, nil
}

func (store *MemoryStore) FindSession(token string) (Session, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	session, ok := store.sessions[hashSessionToken(token)]
	if !ok || time.Now().After(session.ExpiresAt) {
		return Session{}, false
	}
	return session, true
}

func (store *MemoryStore) DeleteSession(token string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.sessions, hashSessionToken(token))
}

type Service struct {
	store  *MemoryStore
	secret string
	secure bool
}

func NewService(store *MemoryStore, secret string, secure bool) *Service {
	return &Service{store: store, secret: secret, secure: secure}
}

func RegisterRoutes(app *fiber.App, service *Service) {
	app.Post("/api/v1/admin/login", service.login)
	app.Get("/api/v1/admin/me", service.me)
	app.Post("/api/v1/admin/logout", service.logout)
}

func (service *Service) login(c fiber.Ctx) error {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind().Body(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "invalid_request", "message": "Invalid login request"}})
	}

	user, ok := service.store.FindUser(request.Username)
	if !ok || user.Status != "active" || !security.VerifyPassword(request.Password, user.PasswordHash) {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": fiber.Map{"code": "unauthorized", "message": "Invalid username or password"}})
	}

	session, err := service.store.CreateSession(user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"code": "internal_error", "message": "Failed to create session"}})
	}

	c.Cookie(&fiber.Cookie{
		Name:     SessionCookieName,
		Value:    session.Token,
		HTTPOnly: true,
		Secure:   service.secure,
		SameSite: "Lax",
		Expires:  session.ExpiresAt,
	})

	return c.Status(http.StatusOK).JSON(fiber.Map{"user": publicUser(user)})
}

func (service *Service) me(c fiber.Ctx) error {
	session, ok := service.currentSession(c)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": fiber.Map{"code": "unauthorized", "message": "Admin session is required"}})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"user": publicUser(session.User)})
}

func (service *Service) logout(c fiber.Ctx) error {
	token := c.Cookies(SessionCookieName)
	if token != "" {
		service.store.DeleteSession(token)
	}
	c.Cookie(&fiber.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   service.secure,
		SameSite: "Lax",
		Expires:  time.Unix(0, 0),
	})
	return c.Status(http.StatusOK).JSON(fiber.Map{"ok": true})
}

func (service *Service) currentSession(c fiber.Ctx) (Session, bool) {
	token := c.Cookies(SessionCookieName)
	if token == "" {
		return Session{}, false
	}
	return service.store.FindSession(token)
}

func publicUser(user User) fiber.Map {
	return fiber.Map{"id": user.ID, "username": user.Username}
}

func randomToken() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
