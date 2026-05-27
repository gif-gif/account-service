package service_test

import (
	"os"
	"strings"
	"testing"
)

func TestDockerfileMakesAppHomeWritableForKiroCli(t *testing.T) {
	data, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatalf("ReadFile(Dockerfile) error = %v", err)
	}
	dockerfile := string(data)

	for _, dir := range []string{"/app/logs", "/app/.aws/sso/cache", "/app/.cache", "/app/.config", "/app/.local/share"} {
		if !strings.Contains(dockerfile, dir) {
			t.Fatalf("Dockerfile does not create %s", dir)
		}
	}
	if !strings.Contains(dockerfile, "chown -R app:app /app") {
		t.Fatal("Dockerfile must chown /app so kiro-cli can create its database/config as the app user")
	}
}
