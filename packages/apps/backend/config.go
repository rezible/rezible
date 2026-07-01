package rez

import (
	"cmp"
	"net/url"
	"os"
	"time"
)

func DefaultConfig() Config {
	return Config{
		App: AppConfig{
			DebugMode:       false,
			FrontendDomain:  "",
			FrontendApiPath: "/api",
			ApiDomain:       "",
			SingleTenant: AppSingleTenantConfig{
				Enabled: false,
				OrgName: "Default",
			},
		},
		AI: AiConfig{},
		HttpServer: HttpServerConfig{
			Host:     cmp.Or(os.Getenv("HOST"), "0.0.0.0"),
			Port:     cmp.Or(os.Getenv("PORT"), "7002"),
			BasePath: "",
			Auth:     HttpAuthConfig{},
		},
		Documents: DocumentsConfig{
			ServerUrl:             "http://localhost:7002",
			SessionTokenSecretHex: "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f",
			Proxy: DocumentsConfigServerProxy{
				Enabled: false,
				Host:    "localhost:7002",
			},
		},
		Postgres: PostgresConfig{
			Host:     "postgres",
			Port:     5432,
			Database: "rezible",
			SSLMode:  "require",
		},
		Telemetry: TelemetryConfig{
			ServiceName: cmp.Or(os.Getenv("OTEL_SERVICE_NAME"), "rezible"),
			Logging: LoggingConfig{
				Console: LoggingConsoleConfig{
					Enabled: true,
					Level:   "info",
					Json:    true,
				},
			},
			Tracing: TracingConfig{
				Enabled: false,
			},
			Metrics: MetricsConfig{
				Enabled:  false,
				Interval: 30 * time.Second,
			},
		},
		Integrations: IntegrationsConfig{},
	}
}

type Config struct {
	App          AppConfig          `cfg:"app"`
	AI           AiConfig           `cfg:"ai"`
	HttpServer   HttpServerConfig   `cfg:"http"`
	Documents    DocumentsConfig    `cfg:"documents"`
	Integrations IntegrationsConfig `cfg:"integrations"`
	Postgres     PostgresConfig     `cfg:"postgres"`
	Telemetry    TelemetryConfig    `cfg:"telemetry"`
}

type (
	AppConfig struct {
		DebugMode       bool                  `cfg:"debug_mode"`
		FrontendDomain  string                `cfg:"frontend_domain" validate:"required"`
		FrontendApiPath string                `cfg:"frontend_api_path" default:"/api"`
		ApiDomain       string                `cfg:"api_domain" validate:"required"`
		SingleTenant    AppSingleTenantConfig `cfg:"singletenant"`
	}
	AppSingleTenantConfig struct {
		Enabled bool   `cfg:"enabled"`
		OrgName string `cfg:"default_org_name"`
	}
)

func (a AppConfig) GetFrontendUrl(paths ...string) (*url.URL, error) {
	fePath, pathErr := url.JoinPath("/", paths...)
	if pathErr != nil {
		return nil, pathErr
	}
	return url.Parse("https://" + a.FrontendDomain + fePath)
}

type (
	AiConfig struct {
		Gemini AiConfigGemini `cfg:"gemini"`
	}
	AiConfigGemini struct {
		Enabled bool   `cfg:"enabled"`
		APIKey  string `cfg:"api_key"`
	}
)

type (
	HttpServerConfig struct {
		Host     string `cfg:"host"`
		Port     string `cfg:"port"`
		BasePath string `cfg:"base_path"`

		Auth HttpAuthConfig `cfg:"auth"`
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
	DocumentsConfig struct {
		ServerUrl             string                     `cfg:"server_url" validate:"required"`
		SessionTokenSecretHex string                     `cfg:"session_token_secret_hex" validate:"len=64"`
		Proxy                 DocumentsConfigServerProxy `cfg:"proxy"`
	}

	DocumentsConfigServerProxy struct {
		Enabled bool   `cfg:"enabled"`
		Host    string `cfg:"host"`
	}
)

type (
	IntegrationsConfig struct {
		Slack  IntegrationsConfigSlack  `cfg:"slack" validate:"omitempty"`
		Github IntegrationsConfigGithub `cfg:"github" validate:"omitempty"`
		Google IntegrationsConfigGoogle `cfg:"google" validate:"omitempty"`
	}
	IntegrationsConfigSlack struct {
		Agent     IntegrationsConfigSlackApp `cfg:"agent" validate:"omitempty"`
		Incidents IntegrationsConfigSlackApp `cfg:"incidents" validate:"omitempty"`
	}
	IntegrationsConfigSlackApp struct {
		Enabled bool `cfg:"enabled"`

		OAuthClientId     string `cfg:"oauth_client_id" validate:"required_if=Enabled true"`
		OAuthClientSecret string `cfg:"oauth_client_secret" validate:"required_if=Enabled true"`

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
