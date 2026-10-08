package test

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"os"
	"strconv"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/koanf"
)

// loadConfig deliberately reads only the repository test runner's database settings.
func loadConfig(ctx context.Context, overrides map[string]any) (rez.Config, error) {
	host := os.Getenv("POSTGRES_TEST_HOST")
	portText := os.Getenv("POSTGRES_TEST_PORT")
	database := os.Getenv("POSTGRES_TEST_DB")
	if host == "" || portText == "" || database == "" {
		return rez.Config{}, fmt.Errorf(
			"POSTGRES_TEST_HOST, POSTGRES_TEST_PORT and POSTGRES_TEST_DB are required; run just test backend",
		)
	}

	port, portErr := strconv.ParseUint(portText, 10, 16)
	if portErr != nil || port == 0 {
		return rez.Config{}, fmt.Errorf("POSTGRES_TEST_PORT must be a port between 1 and 65535")
	}

	adminRole := cmp.Or(os.Getenv("POSTGRES_ADMIN_USER"), "postgres")
	appRole := cmp.Or(os.Getenv("POSTGRES_APP_USER"), "rez_app")

	values := map[string]any{
		"postgres.host":                host,
		"postgres.port":                uint16(port),
		"postgres.database":            database,
		"postgres.role_admin.name":     adminRole,
		"postgres.role_admin.password": adminRole,
		"postgres.role_app.name":       appRole,
		"postgres.role_app.password":   appRole,
		"postgres.sslmode":             "disable",

		"app.frontend_domain":   "app.test",
		"app.api_domain":        "app.test",
		"app.frontend_api_path": "/api",

		"http.host":                    "127.0.0.1",
		"http.port":                    "0",
		"http.base_path":               "/api",
		"http.auth.session_secret":     strings.Repeat("s", 32),
		"http.auth.oidc.issuer":        "https://issuer.invalid",
		"http.auth.oidc.client_id":     "test-client",
		"http.auth.oidc.client_secret": "test-secret",

		"documents.server_url":      "https://documents.invalid",
		"documents.session_key_hex": strings.Repeat("ab", 32),

		"telemetry.service_name":            "rezible-test",
		"telemetry.logging.console.enabled": false,

		"ai.agents.max_workers":    1,
		"ai.agents.worker_timeout": 30 * time.Second,

		// River otherwise finds retried and newly due jobs only on its 1s fetch poll.
		"jobs.fetch_poll_interval": 50 * time.Millisecond,
		"jobs.fetch_cooldown":      50 * time.Millisecond,
	}
	maps.Copy(values, overrides)
	options := koanf.Options{
		LoadEnvironment: false,
		SkipValidation:  false,
		Overrides:       values,
	}
	return koanf.LoadConfig(ctx, options)
}
