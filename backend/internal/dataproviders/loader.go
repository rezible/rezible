package dataproviders

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providerconfig"
	fakeprovider "github.com/rezible/rezible/internal/dataproviders/fake"
	"github.com/rezible/rezible/internal/dataproviders/grafana"
	"github.com/rezible/rezible/internal/dataproviders/jira"
	"github.com/rezible/rezible/internal/dataproviders/slack"
)

type ProviderLoader struct {
	config rez.ProviderConfigService
}

func NewProviderLoader(config rez.ProviderConfigService) *ProviderLoader {
	return &ProviderLoader{
		config: config,
	}
}

func (l *ProviderLoader) listEnabledConfigs(ctx context.Context, t providerconfig.ProviderType) (ent.ProviderConfigs, error) {
	return l.config.ListProviderConfigs(ctx, rez.ListProviderConfigsParams{
		ProviderType: t,
		Enabled:      true,
	})
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

func (l *ProviderLoader) GetUserDataProviders(ctx context.Context) ([]rez.UserDataProvider, error) {
	cfgs, cfgsErr := l.listEnabledConfigs(ctx, providerconfig.ProviderTypeUsers)
	if cfgsErr != nil {
		return nil, cfgsErr
	}

	provs := make([]rez.UserDataProvider, len(cfgs))
	for i, cfg := range cfgs {
		var pErr error
		switch cfg.ProviderID {
		case "slack":
			provs[i], pErr = loadProvider(slack.NewUserDataProvider, cfg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetTeamDataProviders(ctx context.Context) ([]rez.TeamDataProvider, error) {
	cfgs, cfgsErr := l.listEnabledConfigs(ctx, providerconfig.ProviderTypeTeams)
	if cfgsErr != nil {
		return nil, cfgsErr
	}

	provs := make([]rez.TeamDataProvider, len(cfgs))
	for i, cfg := range cfgs {
		var pErr error
		switch cfg.ProviderID {
		case "slack":
			provs[i], pErr = loadProvider(slack.NewTeamDataProvider, cfg)
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewTeamsDataProvider, cfg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetOncallDataProviders(ctx context.Context) ([]rez.OncallDataProvider, error) {
	cfgs, cfgsErr := l.listEnabledConfigs(ctx, providerconfig.ProviderTypeOncall)
	if cfgsErr != nil {
		return nil, cfgsErr
	}

	provs := make([]rez.OncallDataProvider, len(cfgs))
	for i, cfg := range cfgs {
		var pErr error
		switch cfg.ProviderID {
		case "grafana":
			provs[i], pErr = loadProvider(grafana.NewOncallDataProvider, cfg)
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewOncallDataProvider, cfg)
		default:
			pErr = fmt.Errorf("unknown provider: %s", cfg.ProviderID)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetAlertDataProviders(ctx context.Context) ([]rez.AlertDataProvider, error) {
	cfgs, cfgsErr := l.listEnabledConfigs(ctx, providerconfig.ProviderTypeAlerts)
	if cfgsErr != nil {
		return nil, cfgsErr
	}

	provs := make([]rez.AlertDataProvider, len(cfgs))
	for i, cfg := range cfgs {
		var pErr error
		switch cfg.ProviderID {
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewAlertDataProvider, cfg)
		default:
			pErr = fmt.Errorf("unknown provider: %s", cfg.ProviderID)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetIncidentDataProviders(ctx context.Context) ([]rez.IncidentDataProvider, error) {
	cfgs, cfgsErr := l.listEnabledConfigs(ctx, providerconfig.ProviderTypeIncidents)
	if cfgsErr != nil {
		return nil, cfgsErr
	}

	provs := make([]rez.IncidentDataProvider, len(cfgs))
	for i, cfg := range cfgs {
		var pErr error
		switch cfg.ProviderID {
		case "grafana":
			provs[i], pErr = loadProvider(grafana.NewIncidentDataProvider, cfg)
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewIncidentDataProvider, cfg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetSystemComponentsDataProviders(ctx context.Context) ([]rez.SystemComponentsDataProvider, error) {
	cfgs, cfgsErr := l.listEnabledConfigs(ctx, providerconfig.ProviderTypeSystemComponents)
	if cfgsErr != nil {
		return nil, cfgsErr
	}

	provs := make([]rez.SystemComponentsDataProvider, len(cfgs))
	for i, cfg := range cfgs {
		var pErr error
		switch cfg.ProviderID {
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewSystemComponentsDataProvider, cfg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetTicketDataProviders(ctx context.Context) ([]rez.TicketDataProvider, error) {
	cfgs, cfgsErr := l.listEnabledConfigs(ctx, providerconfig.ProviderTypeTickets)
	if cfgsErr != nil {
		return nil, cfgsErr
	}

	provs := make([]rez.TicketDataProvider, len(cfgs))
	for i, cfg := range cfgs {
		var pErr error
		switch cfg.ProviderID {
		case "jira":
			provs[i], pErr = loadProviderCtx(ctx, jira.NewTicketDataProvider, cfg)
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewTicketDataProvider, cfg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetPlaybookDataProviders(ctx context.Context) ([]rez.PlaybookDataProvider, error) {
	cfgs, cfgsErr := l.listEnabledConfigs(ctx, providerconfig.ProviderTypePlaybooks)
	if cfgsErr != nil {
		return nil, cfgsErr
	}

	provs := make([]rez.PlaybookDataProvider, len(cfgs))
	for i, cfg := range cfgs {
		var pErr error
		switch cfg.ProviderID {
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewPlaybookDataProvider, cfg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}
