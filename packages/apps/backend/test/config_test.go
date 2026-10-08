package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigIgnoresAmbientApplicationSettings(t *testing.T) {
	t.Setenv("HOST", "ambient-host")
	t.Setenv("PORT", "9999")
	t.Setenv("OTEL_SERVICE_NAME", "ambient-service")
	t.Setenv("HTTP__AUTH__ENABLE_DEV_SKIP_MODE", "true")
	t.Setenv("APP__SINGLETENANT__ENABLED", "true")
	t.Setenv("AI__GEMINI__ENABLED", "true")
	t.Setenv("AI__GEMINI__API_KEY", "ambient-key")
	t.Setenv("INTEGRATIONS__GITHUB__ENABLED", "true")
	t.Setenv("INTEGRATIONS__SLACK__AGENT__ENABLED", "true")
	t.Setenv("REDIS__ENABLED", "true")
	t.Setenv("TELEMETRY__TRACING__ENABLED", "true")
	t.Setenv("TELEMETRY__METRICS__ENABLED", "true")
	t.Setenv("TELEMETRY__TRACING__CAPTURE_AI_CONTENT", "true")

	cfg, configErr := loadConfig(t.Context(), nil)
	require.NoError(t, configErr)

	require.Equal(t, "127.0.0.1", cfg.HttpServer.Host)
	require.Equal(t, "0", cfg.HttpServer.Port)
	require.Equal(t, "rezible-test", cfg.Telemetry.ServiceName)

	require.False(t, cfg.HttpServer.Auth.EnableDevSkipMode)
	require.False(t, cfg.App.SingleTenant.Enabled)
	require.False(t, cfg.AI.Gemini.Enabled)
	require.Empty(t, cfg.AI.Gemini.APIKey)
	require.False(t, cfg.Integrations.Github.Enabled)
	require.False(t, cfg.Integrations.Slack.Agent.Enabled)
	require.False(t, cfg.Redis.Enabled)
	require.False(t, cfg.Telemetry.Tracing.Enabled)
	require.False(t, cfg.Telemetry.Metrics.Enabled)
	require.False(t, cfg.Telemetry.Tracing.CaptureAiContent)
}

func TestConfigAppliesExplicitOverrides(t *testing.T) {
	overrides := map[string]any{
		"app.frontend_domain": "override.test",
	}

	cfg, configErr := loadConfig(t.Context(), overrides)
	require.NoError(t, configErr)

	require.Equal(t, "override.test", cfg.App.FrontendDomain)
	require.Equal(t, map[string]any{"app.frontend_domain": "override.test"}, overrides)
}

func TestConfigRequiresTestDatabaseCoordinates(t *testing.T) {
	variables := []string{
		"POSTGRES_TEST_HOST",
		"POSTGRES_TEST_PORT",
		"POSTGRES_TEST_DB",
	}
	for _, variable := range variables {
		t.Run(variable, func(t *testing.T) {
			t.Setenv(variable, "")

			_, configErr := loadConfig(t.Context(), nil)

			require.ErrorContains(t, configErr, "just test backend")
		})
	}
}

func TestConfigRejectsInvalidPort(t *testing.T) {
	t.Setenv("POSTGRES_TEST_PORT", "65536")

	_, configErr := loadConfig(t.Context(), nil)

	require.ErrorContains(t, configErr, "POSTGRES_TEST_PORT")
}

func TestConfigValidatesOverrides(t *testing.T) {
	overrides := map[string]any{
		"app.frontend_domain": "",
	}

	_, configErr := loadConfig(t.Context(), overrides)

	require.ErrorContains(t, configErr, "FrontendDomain")
}
