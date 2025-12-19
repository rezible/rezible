package datasync

import (
	"context"
	"fmt"
	"time"

	"github.com/rezible/rezible/ent/integration"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/jobs"
)

type Syncer struct {
	db *ent.Client
	pl rez.DataProviderLoader
}

func NewIntegrationsSyncer(db *ent.Client, pl rez.DataProviderLoader) *Syncer {
	return &Syncer{db: db, pl: pl}
}

func (s *Syncer) MakeSyncAllTenantIntegrationsDataPeriodicJob() jobs.PeriodicJob {
	return jobs.PeriodicJob{
		ConstructorFunc: func() jobs.InsertJobParams {
			return jobs.InsertJobParams{
				Args: &jobs.SyncIntegrationsData{},
				Uniqueness: &jobs.JobUniquenessOpts{
					ByState: jobs.NonCompletedJobStates,
				},
			}
		},
		Interval: time.Hour,
		Opts:     &jobs.PeriodicJobOpts{RunOnStart: true},
	}
}

func (s *Syncer) SyncAllTenantIntegrationsData(ctx context.Context, args jobs.SyncIntegrationsData) error {
	tenants, tenantsErr := s.db.Tenant.Query().All(ctx)
	if tenantsErr != nil {
		return fmt.Errorf("querying tenants: %w", tenantsErr)
	}

	for _, tenant := range tenants {
		tenantCtx := access.TenantSystemContext(ctx, tenant.ID)
		integrationsQuery := s.db.Integration.Query().
			Where(integration.Enabled(true))

		intgs, intgsErr := integrationsQuery.All(tenantCtx)
		if intgsErr != nil {
			return fmt.Errorf("querying integrations: %w", intgsErr)
		}

		if args.Hard {
			s.db.ProviderSyncHistory.Delete().ExecX(tenantCtx)
		}

		if syncErr := s.SyncIntegrationsData(tenantCtx, intgs); syncErr != nil {
			log.Error().
				Err(syncErr).
				Int("tenantID", tenant.ID).
				Msg("failed to sync integrations data")
		}
	}
	return nil
}

func (s *Syncer) SyncIntegrationsData(ctx context.Context, intgs ent.Integrations) error {
	names := make([]string, len(intgs))
	for i, intg := range intgs {
		names[i] = intg.Name
	}

	usersProviders, usersErr := s.pl.GetUserDataProviders(ctx, intgs)
	if usersErr != nil {
		log.Error().Err(usersErr).Msg("failed to load user data providers")
	} else if len(usersProviders) > 0 {
		for _, prov := range usersProviders {
			if syncErr := syncUsers(ctx, s.db, prov); syncErr != nil {
				return fmt.Errorf("user provider (name): %w", syncErr)
			}
		}
	}

	teamsProviders, teamsErr := s.pl.GetTeamDataProviders(ctx, intgs)
	if teamsErr != nil {
		log.Error().Err(teamsErr).Msg("failed to load teams data providers")
	} else if len(teamsProviders) > 0 {
		for _, teams := range teamsProviders {
			if syncErr := syncTeams(ctx, s.db, teams); syncErr != nil {
				return fmt.Errorf("teams: %w", syncErr)
			}
		}
	}

	oncallProviders, oncallErr := s.pl.GetOncallDataProviders(ctx, intgs)
	if oncallErr != nil {
		log.Error().Err(oncallErr).Msg("failed to load oncall data providers")
	} else if len(oncallProviders) > 0 {
		for _, oncall := range oncallProviders {
			if syncErr := syncOncallRosters(ctx, s.db, oncall); syncErr != nil {
				return fmt.Errorf("oncall rosters: %w", syncErr)
			}
			if syncErr := syncOncallShifts(ctx, s.db, oncall); syncErr != nil {
				return fmt.Errorf("oncall shifts: %w", syncErr)
			}
		}
	}

	componentsProviders, componentsErr := s.pl.GetSystemComponentsDataProviders(ctx, intgs)
	if componentsErr != nil {
		log.Error().Err(componentsErr).Msg("failed to load components data providers")
	} else if len(componentsProviders) > 0 {
		for _, components := range componentsProviders {
			if syncErr := syncSystemComponents(ctx, s.db, components); syncErr != nil {
				return fmt.Errorf("system components: %w", syncErr)
			}
		}
	}

	alertsProviders, alertsErr := s.pl.GetAlertDataProviders(ctx, intgs)
	if alertsErr != nil {
		log.Error().Err(alertsErr).Msg("failed to load alerts data providers")
	} else if len(alertsProviders) > 0 {
		for _, alerts := range alertsProviders {
			if syncErr := syncAlerts(ctx, s.db, alerts); syncErr != nil {
				return fmt.Errorf("alerts: %w", syncErr)
			}
			if syncErr := syncAlertInstances(ctx, s.db, alerts); syncErr != nil {
				return fmt.Errorf("alert instances: %w", syncErr)
			}
		}
	}

	playbooksProviders, playbooksErr := s.pl.GetPlaybookDataProviders(ctx, intgs)
	if playbooksErr != nil {
		log.Error().Err(playbooksErr).Msg("failed to load playbooks data providers")
	} else if len(playbooksProviders) > 0 {
		for _, playbooks := range playbooksProviders {
			if syncErr := syncPlaybooks(ctx, s.db, playbooks); syncErr != nil {
				return fmt.Errorf("playbooks: %w", syncErr)
			}
		}
	}

	incidentsProviders, incidentsErr := s.pl.GetIncidentDataProviders(ctx, intgs)
	if incidentsErr != nil {
		log.Error().Err(incidentsErr).Msg("failed to load incidents data providers")
	} else if len(incidentsProviders) > 0 {
		for _, incidents := range incidentsProviders {
			if syncErr := syncIncidentRoles(ctx, s.db, incidents); syncErr != nil {
				return fmt.Errorf("incident roles: %w", syncErr)
			}
			if syncErr := syncIncidents(ctx, s.db, incidents); syncErr != nil {
				return fmt.Errorf("incidents: %w", syncErr)
			}
		}
	}

	return nil
}
