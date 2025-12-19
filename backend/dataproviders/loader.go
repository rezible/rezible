package dataproviders

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	fakeprovider "github.com/rezible/rezible/internal/dataproviders/fake"
	"github.com/rezible/rezible/internal/dataproviders/grafana"
	"github.com/rezible/rezible/internal/dataproviders/jira"
	"github.com/rezible/rezible/internal/dataproviders/slack"
)

type ProviderLoader struct {
}

func NewProviderLoader() *ProviderLoader {
	return &ProviderLoader{}
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

func (l *ProviderLoader) GetUserDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.UserDataProvider, error) {
	var provs []rez.UserDataProvider
	for _, intg := range intgs {
		var prov rez.UserDataProvider
		var pErr error
		switch intg.Name {
		case "slack":
			prov, pErr = loadProvider(slack.NewUserDataProvider, intg)
		default:
			continue
		}
		if pErr != nil {
			return nil, fmt.Errorf("loading provider: %w", pErr)
		}
		provs = append(provs, prov)
	}
	return provs, nil
}

func (l *ProviderLoader) GetTeamDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.TeamDataProvider, error) {
	var provs []rez.TeamDataProvider
	for _, intg := range intgs {
		var prov rez.TeamDataProvider
		var pErr error
		switch intg.Name {
		case "slack":
			prov, pErr = loadProvider(slack.NewTeamDataProvider, intg)
		case "fake":
			prov, pErr = loadProvider(fakeprovider.NewTeamsDataProvider, intg)
		default:
			continue
		}
		if pErr != nil {
			return nil, fmt.Errorf("loading provider: %w", pErr)
		}
		provs = append(provs, prov)
	}
	return provs, nil
}

func (l *ProviderLoader) GetOncallDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.OncallDataProvider, error) {
	var provs []rez.OncallDataProvider
	for _, intg := range intgs {
		var prov rez.OncallDataProvider
		var pErr error
		switch intg.Name {
		case "grafana":
			prov, pErr = loadProvider(grafana.NewOncallDataProvider, intg)
		case "fake":
			prov, pErr = loadProvider(fakeprovider.NewOncallDataProvider, intg)
		default:
			continue
		}
		if pErr != nil {
			return nil, fmt.Errorf("loading provider: %w", pErr)
		}
		provs = append(provs, prov)
	}
	return provs, nil
}

func (l *ProviderLoader) GetAlertDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.AlertDataProvider, error) {
	var provs []rez.AlertDataProvider
	for _, intg := range intgs {
		var prov rez.AlertDataProvider
		var pErr error
		switch intg.Name {
		case "fake":
			prov, pErr = loadProvider(fakeprovider.NewAlertDataProvider, intg)
		default:
			continue
		}
		if pErr != nil {
			return nil, fmt.Errorf("loading provider: %w", pErr)
		}
		provs = append(provs, prov)
	}
	return provs, nil
}

func (l *ProviderLoader) GetIncidentDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.IncidentDataProvider, error) {
	var provs []rez.IncidentDataProvider
	for _, intg := range intgs {
		var prov rez.IncidentDataProvider
		var pErr error
		switch intg.Name {
		case "grafana":
			prov, pErr = loadProvider(grafana.NewIncidentDataProvider, intg)
		case "fake":
			prov, pErr = loadProvider(fakeprovider.NewIncidentDataProvider, intg)
		default:
			continue
		}
		if pErr != nil {
			return nil, fmt.Errorf("loading provider: %w", pErr)
		}
		provs = append(provs, prov)
	}
	return provs, nil
}

func (l *ProviderLoader) GetSystemComponentsDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.SystemComponentsDataProvider, error) {
	var provs []rez.SystemComponentsDataProvider
	for _, intg := range intgs {
		var prov rez.SystemComponentsDataProvider
		var pErr error
		switch intg.Name {
		case "fake":
			prov, pErr = loadProvider(fakeprovider.NewSystemComponentsDataProvider, intg)
		default:
			continue
		}
		if pErr != nil {
			return nil, fmt.Errorf("loading provider: %w", pErr)
		}
		provs = append(provs, prov)
	}
	return provs, nil
}

func (l *ProviderLoader) GetTicketDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.TicketDataProvider, error) {
	var provs []rez.TicketDataProvider
	for _, intg := range intgs {
		var prov rez.TicketDataProvider
		var pErr error
		switch intg.Name {
		case "jira":
			prov, pErr = loadProviderCtx(ctx, jira.NewTicketDataProvider, intg)
		case "fake":
			prov, pErr = loadProvider(fakeprovider.NewTicketDataProvider, intg)
		default:
			continue
		}
		if pErr != nil {
			return nil, fmt.Errorf("loading provider: %w", pErr)
		}
		provs = append(provs, prov)
	}
	return provs, nil
}

func (l *ProviderLoader) GetPlaybookDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.PlaybookDataProvider, error) {
	var provs []rez.PlaybookDataProvider
	for _, intg := range intgs {
		var prov rez.PlaybookDataProvider
		var pErr error
		switch intg.Name {
		case "fake":
			prov, pErr = loadProvider(fakeprovider.NewPlaybookDataProvider, intg)
		default:
			continue
		}
		if pErr != nil {
			return nil, fmt.Errorf("loading provider: %w", pErr)
		}
		provs = append(provs, prov)
	}
	return provs, nil
}
