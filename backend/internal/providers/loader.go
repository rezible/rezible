package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providerconfig"
	"github.com/rezible/rezible/internal/providers/anthropic"
	"github.com/rezible/rezible/internal/providers/fake"
	"github.com/rezible/rezible/internal/providers/grafana"
	"github.com/rezible/rezible/internal/providers/oauth2"
	"github.com/rezible/rezible/internal/providers/saml"
	"github.com/rezible/rezible/internal/providers/slack"
)

var (
	ErrNoProviderConfigured         = errors.New("no provider configured")
	ErrNoStoredConfigs              = errors.New("no stored configs")
	ErrMultipleEnabledStoredConfigs = errors.New("multiple stored configs enabled")
)

type Loader struct {
	client *ent.ProviderConfigClient

	webhookMux       *chi.Mux
	providerWebhooks map[string]rez.Webhooks
}

func NewProviderLoader(client *ent.ProviderConfigClient) *Loader {
	l := &Loader{client: client}

	l.providerWebhooks = make(map[string]rez.Webhooks)
	l.webhookMux = chi.NewMux()

	return l
}

func (l *Loader) WebhookHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.webhookMux.ServeHTTP(w, r)
	})
}

func (l *Loader) updateWebhooks(provKey string, hooks rez.Webhooks) {
	l.providerWebhooks[provKey] = hooks

	m := chi.NewMux()
	for _, provHooks := range l.providerWebhooks {
		for route, handler := range provHooks {
			m.Handle("/"+route, handler)
		}
	}

	l.webhookMux = m
}

type providerConfigFile struct {
	Configs []struct {
		Type         providerconfig.ProviderType `json:"type"`
		ProviderName string                      `json:"provider_name"`
		Disabled     bool                        `json:"disabled"`
		Config       json.RawMessage             `json:"config"`
	} `json:"configs"`
}

func LoadConfigFromFile(ctx context.Context, client *ent.Client, fileName string) error {
	f, openErr := os.Open(fileName)
	if openErr != nil {
		return fmt.Errorf("opening file: %w", openErr)
	}
	defer f.Close()
	fileContents, readErr := io.ReadAll(f)
	if readErr != nil {
		return fmt.Errorf("reading file: %w", readErr)
	}

	var cfg providerConfigFile
	if cfgErr := json.Unmarshal(fileContents, &cfg); cfgErr != nil {
		return fmt.Errorf("unmarshalling file: %w", cfgErr)
	}

	return ent.WithTx(ctx, client, func(tx *ent.Tx) error {
		for _, c := range cfg.Configs {
			log.Info().Str("name", c.ProviderName).Str("type", string(c.Type)).Msg("loading provider")
			upsert := tx.ProviderConfig.Create().
				SetProviderName(c.ProviderName).
				SetProviderType(c.Type).
				SetProviderConfig(c.Config).
				SetEnabled(!c.Disabled).
				SetUpdatedAt(time.Now()).
				OnConflictColumns(
					providerconfig.FieldProviderName,
					providerconfig.FieldProviderType).
				UpdateProviderConfig().
				UpdateUpdatedAt()

			if upsertErr := upsert.Exec(ctx); upsertErr != nil {
				return fmt.Errorf("upserting (%s %s): %w", string(c.Type), c.ProviderName, upsertErr)
			}
		}
		return nil
	})
}

func (l *Loader) LoadProviders(ctx context.Context) (*rez.Providers, error) {
	var provs rez.Providers
	var loadErr error

	provs.AiModel, loadErr = l.LoadAiModelProvider(ctx)
	if loadErr != nil {
		return nil, fmt.Errorf("ai model: %w", loadErr)
	}

	provs.AuthSession, loadErr = l.LoadAuthSessionProvider(ctx)
	if loadErr != nil {
		return nil, fmt.Errorf("auth: %w", loadErr)
	}

	provs.Chat, loadErr = l.LoadChatProvider(ctx)
	if loadErr != nil {
		return nil, fmt.Errorf("chat: %w", loadErr)
	}

	provs.AlertsData, loadErr = l.LoadAlertsDataProvider(ctx)
	if loadErr != nil {
		return nil, fmt.Errorf("alerts: %w", loadErr)
	}

	provs.IncidentData, loadErr = l.LoadIncidentDataProvider(ctx)
	if loadErr != nil {
		return nil, fmt.Errorf("incident data: %w", loadErr)
	}

	provs.OncallData, loadErr = l.LoadOncallDataProvider(ctx)
	if loadErr != nil {
		return nil, fmt.Errorf("oncall data: %w", loadErr)
	}

	provs.SystemComponentsData, loadErr = l.LoadSystemComponentsDataProvider(ctx)
	if loadErr != nil {
		return nil, fmt.Errorf("system components: %w", loadErr)
	}

	provs.TeamData, loadErr = l.LoadTeamDataProvider(ctx)
	if loadErr != nil {
		return nil, fmt.Errorf("team data: %w", loadErr)
	}

	provs.UserData, loadErr = l.LoadUserDataProvider(ctx)
	if loadErr != nil {
		return nil, fmt.Errorf("user data: %w", loadErr)
	}

	return &provs, nil
}

type loadedConfig struct {
	Name      string
	UpdatedAt time.Time
	RawConfig []byte
}

func loadProvider[C any, P any](constructorFn func(C) (P, error), lc *loadedConfig) (P, error) {
	var cfg C
	var p P
	if jsonErr := json.Unmarshal(lc.RawConfig, &cfg); jsonErr != nil {
		return p, fmt.Errorf("failed to unmarshal provider config: %w", jsonErr)
	}
	return constructorFn(cfg)
}

func loadProviderCtx[C any, P any](ctx context.Context, constructorFn func(ctx context.Context, cfg C) (P, error), lc *loadedConfig) (P, error) {
	constructorFnCtx := func(c C) (P, error) {
		return constructorFn(ctx, c)
	}
	return loadProvider(constructorFnCtx, lc)
}

func (l *Loader) loadConfig(ctx context.Context, t providerconfig.ProviderType) (*loadedConfig, error) {
	pc, queryErr := l.client.Query().
		Where(providerconfig.ProviderTypeEQ(t)).
		Where(providerconfig.EnabledEQ(true)).
		Only(ctx)
	if queryErr != nil {
		if ent.IsNotFound(queryErr) {
			return nil, ErrNoStoredConfigs
		} else if ent.IsNotSingular(queryErr) {
			return nil, ErrMultipleEnabledStoredConfigs
		}
		return nil, fmt.Errorf("failed to load %s provider config: %w", t, queryErr)
	}

	cfg := &loadedConfig{
		Name:      strings.ToLower(pc.ProviderName),
		UpdatedAt: pc.UpdatedAt,
		RawConfig: pc.ProviderConfig,
	}
	return cfg, nil
}

func (l *Loader) LoadAiModelProvider(ctx context.Context) (rez.AiModelProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeAi)
	if cfgErr != nil {
		return nil, cfgErr
	}
	switch pCfg.Name {
	case "anthropic":
		return loadProvider(anthropic.NewClaudeAiModelProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid ai model provider config: %s", pCfg.Name)
	}
}

func (l *Loader) LoadChatProvider(ctx context.Context) (rez.ChatProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeChat)
	if cfgErr != nil {
		return nil, cfgErr
	}

	var prov rez.ChatProvider
	var provErr error
	switch pCfg.Name {
	case "slack":
		prov, provErr = loadProvider(slack.NewChatProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid chat provider: %s", pCfg.Name)
	}

	if prov != nil && provErr == nil {
		l.updateWebhooks("chat", prov.GetWebhooks())
	}

	return prov, provErr
}

func (l *Loader) LoadOncallDataProvider(ctx context.Context) (rez.OncallDataProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeOncall)
	if cfgErr != nil {
		return nil, cfgErr
	}

	var prov rez.OncallDataProvider
	var provErr error
	switch pCfg.Name {
	case "grafana":
		prov, provErr = loadProvider(grafana.NewOncallDataProvider, pCfg)
	case "fake":
		prov, provErr = loadProvider(fakeprovider.NewOncallDataProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid oncall data provider: %s", pCfg.Name)
	}

	if prov != nil && provErr == nil {
		l.updateWebhooks("oncall", prov.GetWebhooks())
	}

	return prov, provErr
}

func (l *Loader) LoadAlertsDataProvider(ctx context.Context) (rez.AlertsDataProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeAlerts)
	if cfgErr != nil {
		return nil, cfgErr
	}

	var prov rez.AlertsDataProvider
	var provErr error
	switch pCfg.Name {
	case "grafana":
		prov, provErr = loadProvider(grafana.NewAlertsDataProvider, pCfg)
	case "fake":
		prov, provErr = loadProvider(fakeprovider.NewAlertsDataProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid alerts data provider: %s", pCfg.Name)
	}

	if prov != nil && provErr == nil {
		l.updateWebhooks("alerts", prov.GetWebhooks())
	}

	return prov, provErr
}

func (l *Loader) LoadIncidentDataProvider(ctx context.Context) (rez.IncidentDataProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeIncidents)
	if cfgErr != nil {
		return nil, cfgErr
	}

	var prov rez.IncidentDataProvider
	var provErr error
	switch pCfg.Name {
	case "grafana":
		prov, provErr = loadProvider(grafana.NewIncidentDataProvider, pCfg)
	case "fake":
		prov, provErr = loadProvider(fakeprovider.NewIncidentDataProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid incident data provider: %s", pCfg.Name)
	}

	if prov != nil && provErr == nil {
		l.updateWebhooks("incidents", prov.GetWebhooks())
	}
	return prov, provErr
}

// TODO: use one instance of DataProvider (eg slack)

func (l *Loader) LoadUserDataProvider(ctx context.Context) (rez.UserDataProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeUsers)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch pCfg.Name {
	case "slack":
		return loadProvider(slack.NewDataProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid user data provider: %s", pCfg.Name)
	}
}

func (l *Loader) LoadTeamDataProvider(ctx context.Context) (rez.TeamDataProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeTeams)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch pCfg.Name {
	case "slack":
		return loadProvider(slack.NewDataProvider, pCfg)
	case "fake":
		return loadProvider(fakeprovider.NewTeamsDataProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid user data provider: %s", pCfg.Name)
	}
}

func (l *Loader) LoadSystemComponentsDataProvider(ctx context.Context) (rez.SystemComponentsDataProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeSystemComponents)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch pCfg.Name {
	case "fake":
		return loadProvider(fakeprovider.NewSystemComponentsDataProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid system components data provider: %s", pCfg.Name)
	}
}

func (l *Loader) LoadAuthSessionProvider(ctx context.Context) (rez.AuthSessionProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeAuthSession)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch pCfg.Name {
	case "saml":
		return loadProviderCtx(ctx, saml.NewAuthSessionProvider, pCfg)
	case "oauth2":
		return loadProvider(oauth2.NewAuthSessionProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid auth session provider: %s", pCfg.Name)
	}
}
