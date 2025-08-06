package otter

import (
	"context"

	rez "github.com/rezible/rezible"
)

type ProviderLoader struct {
	pl rez.ProviderLoader
}

func NewCachedProviderLoader(pl rez.ProviderLoader) *ProviderLoader {
	return &ProviderLoader{pl: pl}
}

func (p ProviderLoader) GetLanguageModelProvider(ctx context.Context) (rez.LanguageModelProvider, error) {
	return p.pl.GetLanguageModelProvider(ctx)
}

func (p ProviderLoader) GetIncidentDataProvider(ctx context.Context) (rez.IncidentDataProvider, error) {
	return p.pl.GetIncidentDataProvider(ctx)
}

func (p ProviderLoader) GetOncallDataProvider(ctx context.Context) (rez.OncallDataProvider, error) {
	return p.pl.GetOncallDataProvider(ctx)
}

func (p ProviderLoader) GetSystemComponentsDataProvider(ctx context.Context) (rez.SystemComponentsDataProvider, error) {
	return p.pl.GetSystemComponentsDataProvider(ctx)
}

func (p ProviderLoader) GetTeamDataProvider(ctx context.Context) (rez.TeamDataProvider, error) {
	return p.pl.GetTeamDataProvider(ctx)
}

func (p ProviderLoader) GetUserDataProvider(ctx context.Context) (rez.UserDataProvider, error) {
	return p.pl.GetUserDataProvider(ctx)
}

func (p ProviderLoader) GetTicketDataProvider(ctx context.Context) (rez.TicketDataProvider, error) {
	return p.pl.GetTicketDataProvider(ctx)
}

func (p ProviderLoader) GetAlertDataProvider(ctx context.Context) (rez.AlertDataProvider, error) {
	return p.pl.GetAlertDataProvider(ctx)
}

func (p ProviderLoader) GetPlaybookDataProvider(ctx context.Context) (rez.PlaybookDataProvider, error) {
	return p.pl.GetPlaybookDataProvider(ctx)
}
