package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providerconfig"
	"github.com/rezible/rezible/ent/tenant"
	"github.com/rezible/rezible/internal/providers/fake"
	"github.com/rezible/rezible/internal/providers/grafana"
	"github.com/rezible/rezible/internal/providers/jira"
	"github.com/rezible/rezible/internal/providers/slack"
)

var (
	ErrNoStoredConfigs              = errors.New("no stored configs")
	ErrMultipleEnabledStoredConfigs = errors.New("multiple stored configs enabled")
)

type (
	providerConfig struct {
		Name      string
		UpdatedAt time.Time
		RawConfig []byte
	}
	providerConfigCache map[int]map[providerconfig.ProviderType]providerConfig

	ProviderLoader struct {
		client   *ent.ProviderConfigClient
		cfgCache providerConfigCache
	}

	TenantConfig struct {
		TenantName    string              `json:"tenant_name"`
		ConfigEntries []TenantConfigEntry `json:"configs"`
	}

	TenantConfigEntry struct {
		Type         providerconfig.ProviderType `json:"type"`
		ProviderName string                      `json:"provider_name"`
		Disabled     bool                        `json:"disabled"`
		Config       json.RawMessage             `json:"config"`
	}
)

func NewProviderLoader(client *ent.ProviderConfigClient) *ProviderLoader {
	return &ProviderLoader{
		client:   client,
		cfgCache: make(providerConfigCache),
	}
}

func fakeProviderConfigEntry(t providerconfig.ProviderType) TenantConfigEntry {
	return TenantConfigEntry{Type: t, ProviderName: "fake", Config: []byte("{}")}
}

func LoadDevConfig(ctx context.Context, client *ent.Client) error {
	// TODO: use fake oncall provider
	grafanaOncallRawConfig := fmt.Sprintf(`{"api_endpoint":"%s","api_token":"%s"}`,
		os.Getenv("GRAFANA_ONCALL_API_ENDPOINT"),
		os.Getenv("GRAFANA_ONCALL_API_TOKEN"))
	cfg := &TenantConfig{
		TenantName: "Rezible Test",
		ConfigEntries: []TenantConfigEntry{
			{
				Type:         providerconfig.ProviderTypeUsers,
				ProviderName: "slack",
				Config:       []byte("{}"),
			},
			{
				Type:         providerconfig.ProviderTypeTeams,
				ProviderName: "slack",
				Config:       []byte("{}"),
			},
			{
				Type:         providerconfig.ProviderTypeOncall,
				ProviderName: "grafana",
				Config:       []byte(grafanaOncallRawConfig),
			},
			fakeProviderConfigEntry(providerconfig.ProviderTypeIncidents),
			fakeProviderConfigEntry(providerconfig.ProviderTypeAlerts),
			fakeProviderConfigEntry(providerconfig.ProviderTypeTickets),
			fakeProviderConfigEntry(providerconfig.ProviderTypePlaybooks),
			fakeProviderConfigEntry(providerconfig.ProviderTypeSystemComponents),
		},
	}
	return LoadTenantConfig(ctx, client, cfg)
}

func createTenantContext(ctx context.Context, client *ent.Client, tenantName string) (context.Context, error) {
	tnt, tenantErr := client.Tenant.Query().Where(tenant.Name(tenantName)).Only(ctx)
	if ent.IsNotFound(tenantErr) {
		tnt, tenantErr = client.Tenant.Create().SetName(tenantName).Save(ctx)
	}
	if tenantErr != nil {
		return nil, fmt.Errorf("querying tenant %q: %w", tenantName, tenantErr)
	}
	return access.TenantSystemContext(ctx, tnt.ID), nil
}

func LoadTenantConfig(ctx context.Context, client *ent.Client, cfg *TenantConfig) error {
	tenantCtx, ctxErr := createTenantContext(ctx, client, cfg.TenantName)
	if ctxErr != nil {
		return ctxErr
	}

	return ent.WithTx(tenantCtx, client, func(tx *ent.Tx) error {
		for _, c := range cfg.ConfigEntries {
			log.Info().
				Str("name", c.ProviderName).
				Str("type", string(c.Type)).
				Msg("loading provider")

			upsert := tx.ProviderConfig.Create().
				SetProviderName(c.ProviderName).
				SetProviderType(c.Type).
				SetProviderConfig(c.Config).
				SetEnabled(!c.Disabled).
				SetUpdatedAt(time.Now()).
				OnConflictColumns(providerconfig.FieldProviderName, providerconfig.FieldProviderType).
				UpdateProviderConfig().
				UpdateUpdatedAt()

			if upsertErr := upsert.Exec(tenantCtx); upsertErr != nil {
				return fmt.Errorf("upserting (%s %s): %w", string(c.Type), c.ProviderName, upsertErr)
			}
		}
		return nil
	})
}

func loadProvider[C any, P any](constructorFn func(C) (P, error), lc *providerConfig) (P, error) {
	var cfg C
	var p P
	if jsonErr := json.Unmarshal(lc.RawConfig, &cfg); jsonErr != nil {
		return p, fmt.Errorf("failed to unmarshal provider config: %w", jsonErr)
	}
	return constructorFn(cfg)
}

func loadProviderCtx[C any, P any](ctx context.Context, constructorFn func(ctx context.Context, cfg C) (P, error), lc *providerConfig) (P, error) {
	constructorFnCtx := func(c C) (P, error) {
		return constructorFn(ctx, c)
	}
	return loadProvider(constructorFnCtx, lc)
}

func (l *ProviderLoader) loadProviderConfig(ctx context.Context, t providerconfig.ProviderType) (*providerConfig, error) {
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
	cfg := &providerConfig{
		Name:      strings.ToLower(pc.ProviderName),
		UpdatedAt: pc.UpdatedAt,
		RawConfig: pc.ProviderConfig,
	}
	return cfg, nil
}

func (l *ProviderLoader) loadCachedConfig(tenantId int, t providerconfig.ProviderType) *providerConfig {
	if _, cacheExists := l.cfgCache[tenantId]; !cacheExists {
		l.cfgCache[tenantId] = make(map[providerconfig.ProviderType]providerConfig)
	}
	if cached, exists := l.cfgCache[tenantId][t]; exists {
		return &cached
	}
	return nil
}

func (l *ProviderLoader) loadConfig(ctx context.Context, t providerconfig.ProviderType) (*providerConfig, error) {
	tenantId, idExists := access.GetContextTenantId(ctx)
	if idExists {
		if cached := l.loadCachedConfig(tenantId, t); cached != nil {
			return cached, nil
		}
	}

	cfg, loadErr := l.loadProviderConfig(ctx, t)
	if loadErr != nil {
		return nil, loadErr
	}

	if idExists {
		l.cfgCache[tenantId][t] = *cfg
	}

	return cfg, nil
}

func (l *ProviderLoader) GetOncallDataProvider(ctx context.Context) (rez.OncallDataProvider, error) {
	cfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeOncall)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.Name {
	case "grafana":
		return loadProvider(grafana.NewOncallDataProvider, cfg)
	case "fake":
		return loadProvider(fakeprovider.NewOncallDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid oncall data provider: %s", cfg.Name)
}

func (l *ProviderLoader) GetAlertDataProvider(ctx context.Context) (rez.AlertDataProvider, error) {
	cfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeAlerts)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.Name {
	case "fake":
		return loadProvider(fakeprovider.NewAlertDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid alerts data provider: %s", cfg.Name)
}

func (l *ProviderLoader) GetIncidentDataProvider(ctx context.Context) (rez.IncidentDataProvider, error) {
	cfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeIncidents)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.Name {
	case "grafana":
		return loadProvider(grafana.NewIncidentDataProvider, cfg)
	case "fake":
		return loadProvider(fakeprovider.NewIncidentDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid incident data provider: %s", cfg.Name)
}

func (l *ProviderLoader) GetUserDataProvider(ctx context.Context) (rez.UserDataProvider, error) {
	cfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeUsers)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.Name {
	case "slack":
		return loadProvider(slack.NewUserDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid user data provider: %s", cfg.Name)
}

func (l *ProviderLoader) GetTeamDataProvider(ctx context.Context) (rez.TeamDataProvider, error) {
	cfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeTeams)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.Name {
	case "slack":
		return loadProvider(slack.NewTeamDataProvider, cfg)
	case "fake":
		return loadProvider(fakeprovider.NewTeamsDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid team data provider: %s", cfg.Name)
}

func (l *ProviderLoader) GetSystemComponentsDataProvider(ctx context.Context) (rez.SystemComponentsDataProvider, error) {
	cfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeSystemComponents)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.Name {
	case "fake":
		return loadProvider(fakeprovider.NewSystemComponentsDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid system components data provider: %s", cfg.Name)
}

func (l *ProviderLoader) GetTicketDataProvider(ctx context.Context) (rez.TicketDataProvider, error) {
	cfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeTickets)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.Name {
	case "jira":
		return loadProviderCtx(ctx, jira.NewTicketDataProvider, cfg)
	case "fake":
		return loadProvider(fakeprovider.NewTicketDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid ticket data provider: %s", cfg.Name)
}

func (l *ProviderLoader) GetPlaybookDataProvider(ctx context.Context) (rez.PlaybookDataProvider, error) {
	cfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypePlaybooks)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.Name {
	case "fake":
		return loadProvider(fakeprovider.NewPlaybookDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid playbooks data provider: %s", cfg.Name)
}
