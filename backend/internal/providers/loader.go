package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent/tenant"
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
	"github.com/rezible/rezible/internal/providers/jira"
	"github.com/rezible/rezible/internal/providers/oauth2"
	"github.com/rezible/rezible/internal/providers/saml"
	"github.com/rezible/rezible/internal/providers/slack"
)

var (
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
	TenantName string                    `json:"tenant_name"`
	Configs    []providerConfigFileEntry `json:"configs"`
}

type providerConfigFileEntry struct {
	Type         providerconfig.ProviderType `json:"type"`
	ProviderName string                      `json:"provider_name"`
	Disabled     bool                        `json:"disabled"`
	Config       json.RawMessage             `json:"config"`
}

func readProviderConfigFile(filename string) (*providerConfigFile, error) {
	f, openErr := os.Open(filename)
	if openErr != nil {
		return nil, fmt.Errorf("opening file: %w", openErr)
	}
	defer f.Close()
	fileContents, readErr := io.ReadAll(f)
	if readErr != nil {
		return nil, fmt.Errorf("reading file: %w", readErr)
	}

	var cfg providerConfigFile
	if cfgErr := json.Unmarshal(fileContents, &cfg); cfgErr != nil {
		return nil, fmt.Errorf("unmarshalling file: %w", cfgErr)
	}
	return &cfg, nil
}

func makeUpsertProviderConfigTx(ctx context.Context, cfgs []providerConfigFileEntry) func(tx *ent.Tx) error {
	return func(tx *ent.Tx) error {
		for _, c := range cfgs {
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
	}
}

func LoadConfigFromFile(ctx context.Context, client *ent.Client, fileName string) error {
	cfg, cfgErr := readProviderConfigFile(fileName)
	if cfgErr != nil {
		return cfgErr
	}

	tnt, tenantErr := client.Tenant.Query().Where(tenant.Name(cfg.TenantName)).Only(ctx)
	if ent.IsNotFound(tenantErr) {
		tnt, tenantErr = client.Tenant.Create().SetName(cfg.TenantName).Save(ctx)
	}
	if tenantErr != nil {
		return fmt.Errorf("querying tenant %q: %w", cfg.TenantName, tenantErr)
	}
	tenantCtx := access.TenantContext(ctx, access.RoleSystem, tnt.ID)

	return ent.WithTx(tenantCtx, client, makeUpsertProviderConfigTx(tenantCtx, cfg.Configs))
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

func (l *Loader) GetLanguageModelProvider(ctx context.Context) (rez.LanguageModelProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeAi)
	if cfgErr != nil {
		return nil, cfgErr
	}
	switch pCfg.Name {
	case "anthropic":
		return loadProviderCtx(ctx, anthropic.NewClaudeLanguageModelProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid ai model provider config: %s", pCfg.Name)
	}
}

func (l *Loader) GetChatProvider(ctx context.Context) (rez.ChatProvider, error) {
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

func (l *Loader) GetOncallDataProvider(ctx context.Context) (rez.OncallDataProvider, error) {
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

func (l *Loader) GetAlertDataProvider(ctx context.Context) (rez.AlertDataProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeAlerts)
	if cfgErr != nil {
		return nil, cfgErr
	}

	var prov rez.AlertDataProvider
	var provErr error
	switch pCfg.Name {
	case "fake":
		prov, provErr = loadProvider(fakeprovider.NewAlertDataProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid alerts data provider: %s", pCfg.Name)
	}

	if prov != nil && provErr == nil {
		l.updateWebhooks("alerts", prov.GetWebhooks())
	}

	return prov, provErr
}

func (l *Loader) GetIncidentDataProvider(ctx context.Context) (rez.IncidentDataProvider, error) {
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

func (l *Loader) GetUserDataProvider(ctx context.Context) (rez.UserDataProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeUsers)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch pCfg.Name {
	case "slack":
		return loadProvider(slack.NewUserDataProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid user data provider: %s", pCfg.Name)
	}
}

func (l *Loader) GetTeamDataProvider(ctx context.Context) (rez.TeamDataProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeTeams)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch pCfg.Name {
	case "slack":
		return loadProvider(slack.NewTeamDataProvider, pCfg)
	case "fake":
		return loadProvider(fakeprovider.NewTeamsDataProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid team data provider: %s", pCfg.Name)
	}
}

func (l *Loader) GetSystemComponentsDataProvider(ctx context.Context) (rez.SystemComponentsDataProvider, error) {
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

func (l *Loader) GetAuthSessionProvider(ctx context.Context) (rez.AuthSessionProvider, error) {
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

func (l *Loader) GetTicketDataProvider(ctx context.Context) (rez.TicketDataProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeTickets)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch pCfg.Name {
	case "jira":
		return loadProviderCtx(ctx, jira.NewTicketDataProvider, pCfg)
	case "fake":
		return loadProvider(fakeprovider.NewTicketDataProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid ticket data provider: %s", pCfg.Name)
	}
}

func (l *Loader) GetPlaybookDataProvider(ctx context.Context) (rez.PlaybookDataProvider, error) {
	pCfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypePlaybooks)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch pCfg.Name {
	case "fake":
		return loadProvider(fakeprovider.NewPlaybookDataProvider, pCfg)
	default:
		return nil, fmt.Errorf("invalid playbooks data provider: %s", pCfg.Name)
	}
}
