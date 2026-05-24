package audit

import (
	"context"
	"strings"
	"sync"
	"time"
)

const RedactedValue = "[REDACTED]"

type Event struct {
	ActorType    string
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	RequestID    string
	IPAddress    string
	UserAgent    string
	Metadata     map[string]any
	CreatedAt    time.Time
}

type Writer interface {
	Record(context.Context, Event) error
}

type MemoryWriter struct {
	mu     sync.Mutex
	events []Event
}

func NewMemoryWriter() *MemoryWriter {
	return &MemoryWriter{}
}

func (writer *MemoryWriter) Record(_ context.Context, event Event) error {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	event.Metadata = RedactMetadata(event.Metadata)
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	writer.events = append(writer.events, event)
	return nil
}

func (writer *MemoryWriter) Events() []Event {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	out := make([]Event, len(writer.events))
	copy(out, writer.events)
	return out
}

func RedactMetadata(metadata map[string]any) map[string]any {
	out := make(map[string]any, len(metadata))
	for key, value := range metadata {
		if isSensitiveKey(key) {
			out[key] = RedactedValue
			continue
		}
		nested, ok := value.(map[string]any)
		if ok {
			out[key] = RedactMetadata(nested)
			continue
		}
		out[key] = value
	}
	return out
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	switch normalized {
	case "password", "access_token", "refreshtoken", "refresh_token", "api_key", "apikey":
		return true
	default:
		return false
	}
}
