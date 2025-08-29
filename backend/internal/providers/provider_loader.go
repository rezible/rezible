package providers

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providerconfig"
	"github.com/rezible/rezible/internal/providers/fake"
	"github.com/rezible/rezible/internal/providers/grafana"
	"github.com/rezible/rezible/internal/providers/jira"
	"github.com/rezible/rezible/internal/providers/slack"
)

type ProviderLoader struct {
	config rez.ProviderConfigService
}

func NewProviderLoader(config rez.ProviderConfigService) *ProviderLoader {
	return &ProviderLoader{
		config: config,
	}
}

func loadProviderCtx[C any, P any](ctx context.Context, constructorFn func(context.Context, C) (P, error), pc *ent.ProviderConfig) (P, error) {
	return loadProvider(func(c C) (P, error) {
		return constructorFn(ctx, c)
	}, pc)
}

func loadProvider[C any, P any](constructorFn func(C) (P, error), pc *ent.ProviderConfig) (P, error) {
	var cfg C
	var p P
	if jsonErr := json.Unmarshal(pc.Config, &cfg); jsonErr != nil {
		return p, fmt.Errorf("failed to unmarshal provider config: %w", jsonErr)
	}
	return constructorFn(cfg)
}

func (l *ProviderLoader) GetOncallDataProvider(ctx context.Context) (rez.OncallDataProvider, error) {
	cfg, cfgErr := l.config.GetEnabledTypeConfig(ctx, providerconfig.ProviderTypeOncall)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.ProviderID {
	case "grafana":
		return loadProvider(grafana.NewOncallDataProvider, cfg)
	case "fake":
		return loadProvider(fakeprovider.NewOncallDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid oncall data provider: %s", cfg.ProviderID)
}

func (l *ProviderLoader) GetAlertDataProvider(ctx context.Context) (rez.AlertDataProvider, error) {
	cfg, cfgErr := l.config.GetEnabledTypeConfig(ctx, providerconfig.ProviderTypeAlerts)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.ProviderID {
	case "fake":
		return loadProvider(fakeprovider.NewAlertDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid alerts data provider: %s", cfg.ProviderID)
}

func (l *ProviderLoader) GetIncidentDataProvider(ctx context.Context) (rez.IncidentDataProvider, error) {
	cfg, cfgErr := l.config.GetEnabledTypeConfig(ctx, providerconfig.ProviderTypeIncidents)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.ProviderID {
	case "grafana":
		return loadProvider(grafana.NewIncidentDataProvider, cfg)
	case "fake":
		return loadProvider(fakeprovider.NewIncidentDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid incident data provider: %s", cfg.ProviderID)
}

func (l *ProviderLoader) GetUserDataProvider(ctx context.Context) (rez.UserDataProvider, error) {
	cfg, cfgErr := l.config.GetEnabledTypeConfig(ctx, providerconfig.ProviderTypeUsers)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.ProviderID {
	case "slack":
		return loadProvider(slack.NewUserDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid user data provider: %s", cfg.ProviderID)
}

func (l *ProviderLoader) GetTeamDataProvider(ctx context.Context) (rez.TeamDataProvider, error) {
	cfg, cfgErr := l.config.GetEnabledTypeConfig(ctx, providerconfig.ProviderTypeTeams)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.ProviderID {
	case "slack":
		return loadProvider(slack.NewTeamDataProvider, cfg)
	case "fake":
		return loadProvider(fakeprovider.NewTeamsDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid team data provider: %s", cfg.ProviderID)
}

func (l *ProviderLoader) GetSystemComponentsDataProvider(ctx context.Context) (rez.SystemComponentsDataProvider, error) {
	cfg, cfgErr := l.config.GetEnabledTypeConfig(ctx, providerconfig.ProviderTypeSystemComponents)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.ProviderID {
	case "fake":
		return loadProvider(fakeprovider.NewSystemComponentsDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid system components data provider: %s", cfg.ProviderID)
}

func (l *ProviderLoader) GetTicketDataProvider(ctx context.Context) (rez.TicketDataProvider, error) {
	cfg, cfgErr := l.config.GetEnabledTypeConfig(ctx, providerconfig.ProviderTypeTickets)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.ProviderID {
	case "jira":
		return loadProviderCtx(ctx, jira.NewTicketDataProvider, cfg)
	case "fake":
		return loadProvider(fakeprovider.NewTicketDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid ticket data provider: %s", cfg.ProviderID)
}

func (l *ProviderLoader) GetPlaybookDataProvider(ctx context.Context) (rez.PlaybookDataProvider, error) {
	cfg, cfgErr := l.config.GetEnabledTypeConfig(ctx, providerconfig.ProviderTypePlaybooks)
	if cfgErr != nil {
		return nil, cfgErr
	}

	switch cfg.ProviderID {
	case "fake":
		return loadProvider(fakeprovider.NewPlaybookDataProvider, cfg)
	}
	return nil, fmt.Errorf("invalid playbooks data provider: %s", cfg.ProviderID)
}
