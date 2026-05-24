package callers

import "testing"

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
