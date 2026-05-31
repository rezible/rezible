package rez

import (
	"os"
	"time"
)

func DefaultConfig() Config {
	envOr := func(key, fallback string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return fallback
	}
	return Config{
		App: AppConfig{
			FrontendUrl: "",
			ApiUrl:      "",
			SingleTenant: AppSingleTenantConfig{
				Enabled: false,
				OrgName: "Default",
			},
		},
		HttpServer: HttpServerConfig{
			Host:     envOr("HOST", "0.0.0.0"),
			Port:     envOr("PORT", "7002"),
			BasePath: "",
			Auth:     HttpAuthConfig{},
			DocumentsProxy: HttpServerDocumentsProxyConfig{
				ProxyHost: "localhost:7002",
			},
		},
		Postgres: PostgresConfig{
			Host:     "postgres",
			Port:     5432,
			Database: "rezible",
			SSLMode:  "require",
		},
		Telemetry: TelemetryConfig{
			ServiceName: envOr("OTEL_SERVICE_NAME", "rezible"),
			Logging: LoggingConfig{
				Console: LoggingConsoleConfig{
					Enabled: true,
					Level:   "info",
					Json:    true,
				},
			},
			Tracing: TracingConfig{
				Enabled: true,
			},
			Metrics: MetricsConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
			},
		},
		Integrations: IntegrationsConfig{},
	}
}

type Config struct {
	App          AppConfig          `cfg:"app"`
	HttpServer   HttpServerConfig   `cfg:"http"`
	Integrations IntegrationsConfig `cfg:"integrations"`
	Postgres     PostgresConfig     `cfg:"postgres"`
	Telemetry    TelemetryConfig    `cfg:"telemetry"`
}

type (
	AppConfig struct {
		DebugMode    bool                  `cfg:"debug_mode"`
		FrontendUrl  string                `cfg:"frontend_url" validate:"required"`
		ApiUrl       string                `cfg:"api_url" validate:"required"`
		SingleTenant AppSingleTenantConfig `cfg:"singletenant"`
	}
	AppSingleTenantConfig struct {
		Enabled bool   `cfg:"enabled"`
		OrgName string `cfg:"default_org_name"`
	}
)

type (
	HttpServerConfig struct {
		Host     string `cfg:"host"`
		Port     string `cfg:"port"`
		BasePath string `cfg:"base_path"`

		Auth HttpAuthConfig `cfg:"auth"`

		DocumentsProxy HttpServerDocumentsProxyConfig `cfg:"documents_proxy"`
	}

	HttpServerDocumentsProxyConfig struct {
		Enabled   bool   `cfg:"enabled"`
		ProxyHost string `cfg:"proxy_host"`
	}

	HttpAuthConfig struct {
		SessionSecret []byte             `cfg:"session_secret" validate:"required"`
		Oidc          HttpAuthOidcConfig `cfg:"oidc"`
	}

	HttpAuthOidcConfig struct {
		Issuer       string `cfg:"issuer" validate:"required"`
		ClientID     string `cfg:"client_id" validate:"required"`
		ClientSecret string `cfg:"client_secret" validate:"required"`
		RedirectUrl  string `cfg:"redirect_url"`
	}
)

type (
	IntegrationsConfig struct {
		Slack  IntegrationsConfigSlack  `cfg:"slack" validate:"omitempty"`
		Github IntegrationsConfigGithub `cfg:"github" validate:"omitempty"`
		Google IntegrationsConfigGoogle `cfg:"google" validate:"omitempty"`
	}
	IntegrationsConfigSlack struct {
		Enabled bool `cfg:"enabled"`

		OAuthClientId     string `cfg:"client_id" validate:"required_if=Enabled true"`
		OAuthClientSecret string `cfg:"client_secret" validate:"required_if=Enabled true"`

		WebhookSigningSecret string `cfg:"webhook_signing_secret" validate:"required_if=EnableSocketMode false"`

		EnableSocketMode bool   `cfg:"socketmode_enabled"`
		AppToken         string `cfg:"app_token" validate:"required_if=EnableSocketMode true"`
		BotToken         string `cfg:"bot_token" validate:"required_if=EnableSocketMode true"`
	}
	IntegrationsConfigGithub struct {
		Enabled       bool   `cfg:"enabled"`
		WebhookSecret string `cfg:"webhook_secret" validate:"required"`
		App           struct {
			AppID         int64  `cfg:"app_id"`
			ClientID      string `cfg:"client_id"`
			ClientSecret  string `cfg:"client_secret"`
			PrivateKeyPEM string `cfg:"private_key_pem"`
		} `cfg:"app"`
	}
	IntegrationsConfigGoogle struct {
	}
)

type (
	PostgresConfig struct {
		Host         string             `cfg:"host"`
		Port         uint16             `cfg:"port"`
		Database     string             `cfg:"database"`
		AppRole      PostgresRoleConfig `cfg:"role_app"`
		AdminRole    PostgresRoleConfig `cfg:"role_admin"`
		SSLMode      string             `cfg:"sslmode"`
		PoolMaxConns int32              `cfg:"pool_max_conns"`
	}
	PostgresRoleConfig struct {
		Name     string `cfg:"name"`
		Password string `cfg:"password"`
	}
)

type (
	TelemetryConfig struct {
		ServiceName string        `cfg:"service_name"`
		Logging     LoggingConfig `cfg:"logging"`
		Tracing     TracingConfig `cfg:"tracing"`
		Metrics     MetricsConfig `cfg:"metrics"`
	}

	LoggingConfig struct {
		Console   LoggingConsoleConfig `cfg:"console"`
		OTel      LoggingOtelConfig    `cfg:"otel"`
		AddSource bool                 `cfg:"add_source"`
	}

	LoggingConsoleConfig struct {
		Enabled bool   `cfg:"enabled"`
		Level   string `cfg:"level"`
		Json    bool   `cfg:"json"`
		Color   bool   `cfg:"color"`
	}

	LoggingOtelConfig struct {
		Enabled bool `cfg:"enabled"`
	}

	TracingConfig struct {
		Enabled bool   `cfg:"enabled"`
		Level   string `cfg:"level"`
	}

	MetricsConfig struct {
		Enabled  bool          `cfg:"enabled"`
		Interval time.Duration `cfg:"interval"`
	}
)
