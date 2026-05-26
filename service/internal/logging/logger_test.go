package logging

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewOutputUsesConsoleInDevelopment(t *testing.T) {
	out, path, err := NewOutput("development", t.TempDir(), time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewOutput() error = %v", err)
	}
	defer out.Close()

	if out != os.Stdout {
		t.Fatalf("output = %#v, want stdout", out)
	}
	if path != "" {
		t.Fatalf("path = %q, want empty path for console logging", path)
	}
}

func TestNewOutputWritesDateNamedFileOutsideDevelopment(t *testing.T) {
	dir := t.TempDir()
	out, path, err := NewOutput("production", dir, time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewOutput() error = %v", err)
	}

	if path != filepath.Join(dir, "2026-05-26.log") {
		t.Fatalf("path = %q, want date-named log path", path)
	}
	if _, err := out.Write([]byte("hello\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "hello\n" {
		t.Fatalf("log file = %q, want hello line", string(data))
	}
}
