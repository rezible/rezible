package integrations

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"

	fakeprovider "github.com/rezible/rezible/internal/fake"
	"github.com/rezible/rezible/internal/grafana"
	"github.com/rezible/rezible/internal/jira"
	"github.com/rezible/rezible/internal/slack"
)

// TODO: do these properly
func CheckConfigValid(intg *ent.Integration) bool {
	return true
}

func GetEnabledDataKinds(intg *ent.Integration) []string {
	if intg.Name == "slack" {
		return []string{"users", "chat"}
	}
	return []string{}
}

func GetUserDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.UserDataProvider, error) {
	var provs []rez.UserDataProvider
	for _, intg := range intgs {
		var prov rez.UserDataProvider
		var pErr error
		switch intg.Name {
		case "slack":
			prov, pErr = slack.NewUserDataProvider(intg)
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

func GetTeamDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.TeamDataProvider, error) {
	var provs []rez.TeamDataProvider
	for _, intg := range intgs {
		var prov rez.TeamDataProvider
		var pErr error
		switch intg.Name {
		//case "slack":
		//	prov, pErr = loadProvider(slack.NewTeamDataProvider, intg)
		case "fake":
			prov, pErr = fakeprovider.NewTeamsDataProvider(intg)
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

func GetOncallDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.OncallDataProvider, error) {
	var provs []rez.OncallDataProvider
	for _, intg := range intgs {
		var prov rez.OncallDataProvider
		var pErr error
		switch intg.Name {
		case "grafana":
			prov, pErr = grafana.NewOncallDataProvider(intg)
		case "fake":
			prov, pErr = fakeprovider.NewOncallDataProvider(intg)
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

func GetAlertDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.AlertDataProvider, error) {
	var provs []rez.AlertDataProvider
	for _, intg := range intgs {
		var prov rez.AlertDataProvider
		var pErr error
		switch intg.Name {
		case "fake":
			prov, pErr = fakeprovider.NewAlertDataProvider(intg)
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

func GetIncidentDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.IncidentDataProvider, error) {
	var provs []rez.IncidentDataProvider
	for _, intg := range intgs {
		var prov rez.IncidentDataProvider
		var pErr error
		switch intg.Name {
		case "grafana":
			prov, pErr = grafana.NewIncidentDataProvider(intg)
		case "fake":
			prov, pErr = fakeprovider.NewIncidentDataProvider(intg)
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

func GetSystemComponentsDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.SystemComponentsDataProvider, error) {
	var provs []rez.SystemComponentsDataProvider
	for _, intg := range intgs {
		var prov rez.SystemComponentsDataProvider
		var pErr error
		switch intg.Name {
		case "fake":
			prov, pErr = fakeprovider.NewSystemComponentsDataProvider(intg)
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

func GetTicketDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.TicketDataProvider, error) {
	var provs []rez.TicketDataProvider
	for _, intg := range intgs {
		var prov rez.TicketDataProvider
		var pErr error
		switch intg.Name {
		case "jira":
			prov, pErr = jira.NewTicketDataProvider(ctx, intg)
		case "fake":
			prov, pErr = fakeprovider.NewTicketDataProvider(intg)
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

func GetPlaybookDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.PlaybookDataProvider, error) {
	var provs []rez.PlaybookDataProvider
	for _, intg := range intgs {
		var prov rez.PlaybookDataProvider
		var pErr error
		switch intg.Name {
		case "fake":
			prov, pErr = fakeprovider.NewPlaybookDataProvider(intg)
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
