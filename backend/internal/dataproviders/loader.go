package dataproviders

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	fakeprovider "github.com/rezible/rezible/internal/dataproviders/fake"
	"github.com/rezible/rezible/internal/dataproviders/grafana"
	"github.com/rezible/rezible/internal/dataproviders/jira"
	"github.com/rezible/rezible/internal/dataproviders/slack"
)

type ProviderLoader struct {
	integrations rez.IntegrationsService
}

func NewProviderLoader(is rez.IntegrationsService) *ProviderLoader {
	return &ProviderLoader{integrations: is}
}

func (l *ProviderLoader) listEnabledIntegrations(ctx context.Context, t integration.IntegrationType) (ent.Integrations, error) {
	return l.integrations.ListIntegrations(ctx, rez.ListIntegrationsParams{
		Type:    t,
		Enabled: true,
	})
}

func loadProviderCtx[C any, P any](ctx context.Context, constructorFn func(context.Context, C) (P, error), intg *ent.Integration) (P, error) {
	return loadProvider(func(c C) (P, error) {
		return constructorFn(ctx, c)
	}, intg)
}

func loadProvider[C any, P any](constructorFn func(C) (P, error), intg *ent.Integration) (P, error) {
	var cfg C
	var p P
	if jsonErr := json.Unmarshal(intg.Config, &cfg); jsonErr != nil {
		return p, fmt.Errorf("failed to unmarshal integration config: %w", jsonErr)
	}
	return constructorFn(cfg)
}

func (l *ProviderLoader) GetUserDataProviders(ctx context.Context) ([]rez.UserDataProvider, error) {
	intgs, intgsErr := l.listEnabledIntegrations(ctx, integration.IntegrationTypeUsers)
	if intgsErr != nil {
		return nil, intgsErr
	}

	provs := make([]rez.UserDataProvider, len(intgs))
	for i, intg := range intgs {
		var pErr error
		switch intg.Name {
		case "slack":
			provs[i], pErr = loadProvider(slack.NewUserDataProvider, intg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetTeamDataProviders(ctx context.Context) ([]rez.TeamDataProvider, error) {
	intgs, intgsErr := l.listEnabledIntegrations(ctx, integration.IntegrationTypeTeams)
	if intgsErr != nil {
		return nil, intgsErr
	}

	provs := make([]rez.TeamDataProvider, len(intgs))
	for i, intg := range intgs {
		var pErr error
		switch intg.Name {
		case "slack":
			provs[i], pErr = loadProvider(slack.NewTeamDataProvider, intg)
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewTeamsDataProvider, intg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetOncallDataProviders(ctx context.Context) ([]rez.OncallDataProvider, error) {
	intgs, intgsErr := l.listEnabledIntegrations(ctx, integration.IntegrationTypeOncall)
	if intgsErr != nil {
		return nil, intgsErr
	}

	provs := make([]rez.OncallDataProvider, len(intgs))
	for i, intg := range intgs {
		var pErr error
		switch intg.Name {
		case "grafana":
			provs[i], pErr = loadProvider(grafana.NewOncallDataProvider, intg)
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewOncallDataProvider, intg)
		default:
			pErr = fmt.Errorf("unknown provider: %s", intg.Name)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetAlertDataProviders(ctx context.Context) ([]rez.AlertDataProvider, error) {
	intgs, intgsErr := l.listEnabledIntegrations(ctx, integration.IntegrationTypeAlerts)
	if intgsErr != nil {
		return nil, intgsErr
	}

	provs := make([]rez.AlertDataProvider, len(intgs))
	for i, intg := range intgs {
		var pErr error
		switch intg.Name {
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewAlertDataProvider, intg)
		default:
			pErr = fmt.Errorf("unknown provider: %s", intg.Name)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetIncidentDataProviders(ctx context.Context) ([]rez.IncidentDataProvider, error) {
	intgs, intgsErr := l.listEnabledIntegrations(ctx, integration.IntegrationTypeIncidents)
	if intgsErr != nil {
		return nil, intgsErr
	}

	provs := make([]rez.IncidentDataProvider, len(intgs))
	for i, intg := range intgs {
		var pErr error
		switch intg.Name {
		case "grafana":
			provs[i], pErr = loadProvider(grafana.NewIncidentDataProvider, intg)
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewIncidentDataProvider, intg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetSystemComponentsDataProviders(ctx context.Context) ([]rez.SystemComponentsDataProvider, error) {
	intgs, intgsErr := l.listEnabledIntegrations(ctx, integration.IntegrationTypeSystemComponents)
	if intgsErr != nil {
		return nil, intgsErr
	}

	provs := make([]rez.SystemComponentsDataProvider, len(intgs))
	for i, intg := range intgs {
		var pErr error
		switch intg.Name {
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewSystemComponentsDataProvider, intg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetTicketDataProviders(ctx context.Context) ([]rez.TicketDataProvider, error) {
	intgs, intgsErr := l.listEnabledIntegrations(ctx, integration.IntegrationTypeTickets)
	if intgsErr != nil {
		return nil, intgsErr
	}

	provs := make([]rez.TicketDataProvider, len(intgs))
	for i, intg := range intgs {
		var pErr error
		switch intg.Name {
		case "jira":
			provs[i], pErr = loadProviderCtx(ctx, jira.NewTicketDataProvider, intg)
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewTicketDataProvider, intg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}

func (l *ProviderLoader) GetPlaybookDataProviders(ctx context.Context) ([]rez.PlaybookDataProvider, error) {
	intgs, intgsErr := l.listEnabledIntegrations(ctx, integration.IntegrationTypePlaybooks)
	if intgsErr != nil {
		return nil, intgsErr
	}

	provs := make([]rez.PlaybookDataProvider, len(intgs))
	for i, intg := range intgs {
		var pErr error
		switch intg.Name {
		case "fake":
			provs[i], pErr = loadProvider(fakeprovider.NewPlaybookDataProvider, intg)
		}
		if pErr != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", pErr)
		}
	}
	return provs, nil
}
