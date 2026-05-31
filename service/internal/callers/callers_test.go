package callers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestCreateReturnsPlaintextOnceAndStoresHash(t *testing.T) {
	store := NewMemoryStore()

	result, err := store.Create("worker", "background worker")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.PlaintextAPIKey == "" {
		t.Fatal("expected plaintext API key")
	}
	if result.Caller.APIKeyHash == "" || result.Caller.APIKeyHash == result.PlaintextAPIKey {
		t.Fatalf("stored hash = %q, plaintext = %q", result.Caller.APIKeyHash, result.PlaintextAPIKey)
	}

	caller, ok := store.Authenticate(result.PlaintextAPIKey)
	if !ok {
		t.Fatal("expected plaintext API key to authenticate")
	}
	if caller.ID != result.Caller.ID {
		t.Fatalf("caller ID = %s, want %s", caller.ID, result.Caller.ID)
	}

	apiKey, err := store.RevealAPIKey(result.Caller.ID)
	if err != nil {
		t.Fatalf("RevealAPIKey() error = %v", err)
	}
	if apiKey != result.PlaintextAPIKey {
		t.Fatalf("revealed api key = %q, want original plaintext", apiKey)
	}
}

func TestDisableCallerRejectsAuthentication(t *testing.T) {
	store := NewMemoryStore()
	result, err := store.Create("worker", "background worker")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := store.Disable(result.Caller.ID); err != nil {
		t.Fatalf("Disable() error = %v", err)
	}
	if _, ok := store.Authenticate(result.PlaintextAPIKey); ok {
		t.Fatal("expected disabled caller authentication to fail")
	}
}

func TestMemoryStoreListsUpdatesAndDeletesCallers(t *testing.T) {
	store := NewMemoryStore()
	result, err := store.Create("worker", "background worker")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	callers := store.List()
	if len(callers) != 1 {
		t.Fatalf("List() length = %d, want 1", len(callers))
	}
	if callers[0].ID != result.Caller.ID {
		t.Fatalf("List()[0].ID = %q, want %q", callers[0].ID, result.Caller.ID)
	}
	if !callers[0].CreatedAt.Equal(callers[0].UpdatedAt) {
		t.Fatalf("created caller timestamps differ: %#v", callers[0])
	}

	name := "worker-renamed"
	description := "renamed worker"
	status := StatusDisabled
	updated, err := store.Update(result.Caller.ID, UpdateRequest{Name: &name, Description: &description, Status: &status})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != name || updated.Description != description || updated.Status != StatusDisabled {
		t.Fatalf("updated caller = %#v", updated)
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) && !updated.UpdatedAt.Equal(updated.CreatedAt) {
		t.Fatalf("updated_at = %s, created_at = %s", updated.UpdatedAt, updated.CreatedAt)
	}
	if _, ok := store.Authenticate(result.PlaintextAPIKey); ok {
		t.Fatal("expected disabled updated caller authentication to fail")
	}

	if err := store.Delete(result.Caller.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if callers := store.List(); len(callers) != 0 {
		t.Fatalf("List() length after delete = %d, want 0", len(callers))
	}
}

func TestRegisterRoutesExposeAPIKeyCRUD(t *testing.T) {
	store := NewMemoryStore()
	app := fiber.New()
	RegisterRoutes(app, store)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", strings.NewReader(`{
		"name": "worker",
		"description": "background worker",
		"status": "disabled"
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createResp, err := app.Test(createReq)
	if err != nil {
		t.Fatalf("create app.Test() error = %v", err)
	}
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResp.StatusCode, http.StatusCreated)
	}
	var createBody CreateResult
	if err := json.NewDecoder(createResp.Body).Decode(&createBody); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createBody.Caller.Status != StatusDisabled {
		t.Fatalf("created status = %q, want %q", createBody.Caller.Status, StatusDisabled)
	}

	listResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/api-keys", nil))
	if err != nil {
		t.Fatalf("list app.Test() error = %v", err)
	}
	var listBody struct {
		Callers []Caller `json:"callers"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listBody.Callers) != 1 || listBody.Callers[0].Name != "worker" {
		t.Fatalf("list callers = %#v", listBody.Callers)
	}
	if apiKey, ok := callerJSONField(listBody.Callers[0], "api_key"); ok && apiKey != "" {
		t.Fatalf("list response exposed api_key = %q", apiKey)
	}

	secretReq := httptest.NewRequest(http.MethodGet, "/api/v1/api-keys/"+createBody.Caller.ID+"/secret", nil)
	secretResp, err := app.Test(secretReq)
	if err != nil {
		t.Fatalf("secret app.Test() error = %v", err)
	}
	if secretResp.StatusCode != http.StatusOK {
		t.Fatalf("secret status = %d, want %d", secretResp.StatusCode, http.StatusOK)
	}
	var secretBody struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(secretResp.Body).Decode(&secretBody); err != nil {
		t.Fatalf("decode secret response: %v", err)
	}
	if secretBody.APIKey != createBody.PlaintextAPIKey {
		t.Fatalf("secret api key = %q, want created plaintext", secretBody.APIKey)
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/api-keys/"+createBody.Caller.ID, strings.NewReader(`{
		"name": "worker-2",
		"status": "active"
	}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp, err := app.Test(updateReq)
	if err != nil {
		t.Fatalf("update app.Test() error = %v", err)
	}
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d, want %d", updateResp.StatusCode, http.StatusOK)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/api-keys/"+createBody.Caller.ID, nil)
	deleteResp, err := app.Test(deleteReq)
	if err != nil {
		t.Fatalf("delete app.Test() error = %v", err)
	}
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("delete status = %d, want %d", deleteResp.StatusCode, http.StatusOK)
	}
}

func callerJSONField(caller Caller, field string) (string, bool) {
	body, _ := json.Marshal(caller)
	values := map[string]string{}
	_ = json.Unmarshal(body, &values)
	value, ok := values[field]
	return value, ok
}
