package audit

import (
	"context"
	"testing"
)

func TestMemoryWriterRecordsEvent(t *testing.T) {
	writer := NewMemoryWriter()

	err := writer.Record(context.Background(), Event{
		ActorType:    "admin",
		ActorID:      "admin-id",
		Action:       "account.update",
		ResourceType: "account",
		ResourceID:   "account-id",
		RequestID:    "request-id",
		IPAddress:    "127.0.0.1",
		UserAgent:    "test",
		Metadata: map[string]any{
			"status": "active",
		},
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	events := writer.Events()
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].Action != "account.update" {
		t.Fatalf("Action = %q, want account.update", events[0].Action)
	}
}

func TestRedactsSensitiveMetadata(t *testing.T) {
	metadata := RedactMetadata(map[string]any{
		"password":      "plain-password",
		"access_token":  "access-token",
		"refresh_token": "refresh-token",
		"api_key":       "api-key",
		"nested": map[string]any{
			"password": "nested-password",
		},
		"status": "active",
	})

	for _, key := range []string{"password", "access_token", "refresh_token", "api_key"} {
		if metadata[key] != RedactedValue {
			t.Fatalf("%s = %v, want redacted", key, metadata[key])
		}
	}
	nested, ok := metadata["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested = %#v, want map", metadata["nested"])
	}
	if nested["password"] != RedactedValue {
		t.Fatalf("nested password = %v, want redacted", nested["password"])
	}
	if metadata["status"] != "active" {
		t.Fatalf("status = %v, want active", metadata["status"])
	}
}
