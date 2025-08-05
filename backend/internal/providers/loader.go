package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providerconfig"
	"github.com/rezible/rezible/ent/tenant"
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
}

func NewProviderLoader(client *ent.ProviderConfigClient) *Loader {
	l := &Loader{client: client}

	return l
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
	cfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeAi)
	if cfgErr != nil {
		return nil, cfgErr
	}
	switch cfg.Name {
	case "anthropic":
		return loadProviderCtx(ctx, anthropic.NewClaudeLanguageModelProvider, cfg)
	}
	return nil, fmt.Errorf("invalid ai model provider config: %s", cfg.Name)
}

func (l *Loader) GetOncallDataProvider(ctx context.Context) (rez.OncallDataProvider, error) {
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

func (l *Loader) GetAlertDataProvider(ctx context.Context) (rez.AlertDataProvider, error) {
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

func (l *Loader) GetIncidentDataProvider(ctx context.Context) (rez.IncidentDataProvider, error) {
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

func (l *Loader) GetUserDataProvider(ctx context.Context) (rez.UserDataProvider, error) {
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

func (l *Loader) GetTeamDataProvider(ctx context.Context) (rez.TeamDataProvider, error) {
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

func (l *Loader) GetSystemComponentsDataProvider(ctx context.Context) (rez.SystemComponentsDataProvider, error) {
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

func (l *Loader) GetAuthSessionProvider(ctx context.Context) (rez.AuthSessionProvider, error) {
	cfg, cfgErr := l.loadConfig(ctx, providerconfig.ProviderTypeAuthSession)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.Name {
	case "saml":
		return loadProviderCtx(ctx, saml.NewAuthSessionProvider, cfg)
	case "oauth2":
		return loadProvider(oauth2.NewAuthSessionProvider, cfg)
	}
	return nil, fmt.Errorf("invalid auth session provider: %s", cfg.Name)
}

func (l *Loader) GetTicketDataProvider(ctx context.Context) (rez.TicketDataProvider, error) {
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

func (l *Loader) GetPlaybookDataProvider(ctx context.Context) (rez.PlaybookDataProvider, error) {
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
