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

func TestEntrypointFixesMountedKiroCliDataDirectoryOwnership(t *testing.T) {
	data, err := os.ReadFile("docker-entrypoint.sh")
	if err != nil {
		t.Fatalf("ReadFile(docker-entrypoint.sh) error = %v", err)
	}
	entrypoint := string(data)

	for _, dir := range []string{"/app/logs", "/app/.aws/sso/cache", "/app/.cache", "/app/.config", "/app/.local/share"} {
		if !strings.Contains(entrypoint, dir) {
			t.Fatalf("docker-entrypoint.sh must repair ownership for mounted directory %s", dir)
		}
	}
	if !strings.Contains(entrypoint, "chown -R app:app") {
		t.Fatal("docker-entrypoint.sh must recursively chown mounted kiro-cli data directories before gosu app")
	}
}

func TestEntrypointUsesAppHomeForKiroCliRuntimeData(t *testing.T) {
	data, err := os.ReadFile("docker-entrypoint.sh")
	if err != nil {
		t.Fatalf("ReadFile(docker-entrypoint.sh) error = %v", err)
	}
	entrypoint := string(data)

	for _, env := range []string{
		"HOME=/app",
		"XDG_CACHE_HOME=/app/.cache",
		"XDG_CONFIG_HOME=/app/.config",
		"XDG_DATA_HOME=/app/.local/share",
	} {
		if !strings.Contains(entrypoint, env) {
			t.Fatalf("docker-entrypoint.sh must set %s so kiro-cli does not write under /root", env)
		}
	}
}
