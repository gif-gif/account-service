package admin

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"account-service/service/internal/security"

	"github.com/gofiber/fiber/v3"
)

const SessionCookieName = "account_admin_session"
const accessTokenType = "access"
const refreshTokenType = "refresh"

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

type PublicUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type AuthResponse struct {
	User         PublicUser `json:"user"`
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
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
	store      *MemoryStore
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(store *MemoryStore, secret string, accessTTL time.Duration, refreshTTL time.Duration) *Service {
	return &Service{store: store, secret: secret, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func RegisterRoutes(app *fiber.App, service *Service) {
	app.Post("/api/v1/admin/login", service.LoginHandler())
	app.Post("/api/v1/admin/refresh", service.refresh)
	app.Get("/api/v1/admin/me", service.me)
	app.Post("/api/v1/admin/logout", service.logout)
}

func (service *Service) LoginHandler() fiber.Handler {
	return service.login
}

func (service *Service) CurrentSession(c fiber.Ctx) (Session, bool) {
	return service.currentSession(c)
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

	response, err := service.authResponse(user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"code": "internal_error", "message": "Failed to create token"}})
	}

	return c.Status(http.StatusOK).JSON(response)
}

func (service *Service) refresh(c fiber.Ctx) error {
	var request struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.Bind().Body(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "invalid_request", "message": "Invalid refresh request"}})
	}

	user, ok := service.userFromToken(request.RefreshToken, refreshTokenType)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": fiber.Map{"code": "unauthorized", "message": "Refresh token is invalid or expired"}})
	}
	response, err := service.authResponse(user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"code": "internal_error", "message": "Failed to create token"}})
	}

	return c.Status(http.StatusOK).JSON(response)
}

func (service *Service) me(c fiber.Ctx) error {
	session, ok := service.currentSession(c)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": fiber.Map{"code": "unauthorized", "message": "Access token is required"}})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"user": publicUser(session.User)})
}

func (service *Service) logout(c fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{"ok": true})
}

func (service *Service) currentSession(c fiber.Ctx) (Session, bool) {
	token := bearerToken(c)
	if token == "" {
		return Session{}, false
	}
	claims, ok := service.verifyToken(token, accessTokenType)
	if !ok {
		return Session{}, false
	}
	user, ok := service.store.FindUser(claims.Username)
	if !ok || user.Status != "active" || user.ID != claims.Subject {
		return Session{}, false
	}
	return Session{
		Token:     token,
		TokenHash: hashSessionToken(token),
		User:      user,
		ExpiresAt: time.Unix(claims.ExpiresAt, 0),
	}, true
}

func (service *Service) authResponse(user User) (AuthResponse, error) {
	accessToken, err := service.signToken(user, accessTokenType, service.accessTTL)
	if err != nil {
		return AuthResponse{}, err
	}
	refreshToken, err := service.signToken(user, refreshTokenType, service.refreshTTL)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{
		User:         publicUser(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (service *Service) userFromToken(token string, tokenType string) (User, bool) {
	claims, ok := service.verifyToken(token, tokenType)
	if !ok {
		return User{}, false
	}
	user, ok := service.store.FindUser(claims.Username)
	if !ok || user.Status != "active" || user.ID != claims.Subject {
		return User{}, false
	}
	return user, true
}

type tokenClaims struct {
	Subject   string `json:"sub"`
	Username  string `json:"username"`
	TokenType string `json:"typ"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	ID        string `json:"jti"`
}

func (service *Service) signToken(user User, tokenType string, ttl time.Duration) (string, error) {
	header, err := json.Marshal(fiber.Map{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	jti, err := randomToken()
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims, err := json.Marshal(tokenClaims{
		Subject:   user.ID,
		Username:  user.Username,
		TokenType: tokenType,
		ExpiresAt: now.Add(ttl).Unix(),
		IssuedAt:  now.Unix(),
		ID:        jti,
	})
	if err != nil {
		return "", err
	}
	signingInput := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(claims)
	signature := service.sign(signingInput)
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (service *Service) verifyToken(token string, tokenType string) (tokenClaims, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return tokenClaims{}, false
	}
	signingInput := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return tokenClaims{}, false
	}
	if !hmac.Equal(signature, service.sign(signingInput)) {
		return tokenClaims{}, false
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenClaims{}, false
	}
	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return tokenClaims{}, false
	}
	if claims.TokenType != tokenType || time.Now().Unix() >= claims.ExpiresAt {
		return tokenClaims{}, false
	}
	return claims, true
}

func (service *Service) sign(value string) []byte {
	mac := hmac.New(sha256.New, []byte(service.secret))
	mac.Write([]byte(value))
	return mac.Sum(nil)
}

func bearerToken(c fiber.Ctx) string {
	auth := c.Get("Authorization")
	token, ok := strings.CutPrefix(auth, "Bearer ")
	if !ok {
		return ""
	}
	return strings.TrimSpace(token)
}

func publicUser(user User) PublicUser {
	return PublicUser{ID: user.ID, Username: user.Username}
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
