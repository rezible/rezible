package integrations

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/google"

	fakeprovider "github.com/rezible/rezible/internal/fake"
	"github.com/rezible/rezible/internal/grafana"
	"github.com/rezible/rezible/internal/jira"
	"github.com/rezible/rezible/internal/slack"
)

var (
	packageMap        = map[string]rez.IntegrationPackage{}
	packageSetupFuncs = []rez.SetupPackageFunc{
		fakeprovider.SetupIntegration,
		slack.SetupIntegration,
		google.SetupIntegration,
	}
)

func Setup(ctx context.Context, svcs *rez.Services) error {
	packageMap = make(map[string]rez.IntegrationPackage)
	for _, pkgFn := range packageSetupFuncs {
		pkg, pkgErr := pkgFn(ctx, svcs)
		if pkgErr != nil {
			return fmt.Errorf("setup integration: %w", pkgErr)
		}
		packageMap[pkg.Name()] = pkg
	}
	return nil
}

func GetAvailable() []rez.IntegrationPackage {
	enabled := make([]rez.IntegrationPackage, 0)
	for _, pkg := range packageMap {
		if pkg.Enabled() {
			enabled = append(enabled, pkg)
		}
	}
	return enabled
}

func GetPackage(name string) (rez.IntegrationPackage, error) {
	p, valid := packageMap[name]
	if !valid {
		return nil, fmt.Errorf("unknown integration package: %s", name)
	}
	return p, nil
}

func GetDataProviders[T any](intgs ent.Integrations, iFn func(rez.IntegrationPackage, *ent.Integration) (bool, T, error)) ([]T, error) {
	var provs []T
	for _, intg := range intgs {
		if p, valid := packageMap[intg.Name]; valid {
			if supported, prov, pErr := iFn(p, intg); supported {
				if pErr != nil {
					return nil, fmt.Errorf("loading data provider: %w", pErr)
				}
				provs = append(provs, prov)
			}
		}
	}
	return provs, nil
}

func GetUserDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.UserDataProvider, error) {
	type integrationWithUserDataProvider interface {
		MakeUserDataProvider(context.Context, *ent.Integration) (rez.UserDataProvider, error)
	}

	provFn := func(p rez.IntegrationPackage, i *ent.Integration) (bool, rez.UserDataProvider, error) {
		if dpi, ok := p.(integrationWithUserDataProvider); ok {
			prov, pErr := dpi.MakeUserDataProvider(ctx, i)
			return true, prov, pErr
		}
		return false, nil, nil
	}
	return GetDataProviders[rez.UserDataProvider](intgs, provFn)
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
