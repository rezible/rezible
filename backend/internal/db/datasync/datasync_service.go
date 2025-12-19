package datasync

import (
	"context"
	"fmt"
	"time"

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

func (s *Syncer) MakeSyncIntegrationsDataPeriodicJob() jobs.PeriodicJob {
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

func (s *Syncer) SyncIntegrationsData(ctx context.Context, args jobs.SyncIntegrationsData) error {
	tenants, tenantsErr := s.db.Tenant.Query().All(ctx)
	if tenantsErr != nil {
		return fmt.Errorf("querying tenants: %w", tenantsErr)
	}

	for _, tenant := range tenants {
		tenantCtx := access.TenantSystemContext(ctx, tenant.ID)
		if syncErr := s.syncProviderData(tenantCtx, args.Hard); syncErr != nil {
			log.Error().
				Err(syncErr).
				Int("tenantID", tenant.ID).
				Msg("failed to sync provider data")
		}
	}
	return nil
}

func (s *Syncer) syncProviderData(ctx context.Context, hard bool) error {
	if hard {
		s.db.ProviderSyncHistory.Delete().ExecX(ctx)
	}

	if usersErr := s.SyncUserData(ctx); usersErr != nil {
		return fmt.Errorf("syncing users: %w", usersErr)
	}

	teamsProviders, teamsErr := s.pl.GetTeamDataProviders(ctx)
	if teamsErr == nil {
		for _, teams := range teamsProviders {
			if syncErr := syncTeams(ctx, s.db, teams); syncErr != nil {
				return fmt.Errorf("teams: %w", syncErr)
			}
		}
	}

	oncallProviders, oncallErr := s.pl.GetOncallDataProviders(ctx)
	if oncallErr == nil {
		for _, oncall := range oncallProviders {
			if syncErr := syncOncallRosters(ctx, s.db, oncall); syncErr != nil {
				return fmt.Errorf("oncall rosters: %w", syncErr)
			}
			if syncErr := syncOncallShifts(ctx, s.db, oncall); syncErr != nil {
				return fmt.Errorf("oncall shifts: %w", syncErr)
			}
		}
	}

	componentsProviders, componentsErr := s.pl.GetSystemComponentsDataProviders(ctx)
	if componentsErr == nil {
		for _, components := range componentsProviders {
			if syncErr := syncSystemComponents(ctx, s.db, components); syncErr != nil {
				return fmt.Errorf("system components: %w", syncErr)
			}
		}
	}

	alertsProviders, alertsErr := s.pl.GetAlertDataProviders(ctx)
	if alertsErr == nil {
		for _, alerts := range alertsProviders {
			if syncErr := syncAlerts(ctx, s.db, alerts); syncErr != nil {
				return fmt.Errorf("alerts: %w", syncErr)
			}
			if syncErr := syncAlertInstances(ctx, s.db, alerts); syncErr != nil {
				return fmt.Errorf("alert instances: %w", syncErr)
			}
		}
	}

	playbooksProviders, playbooksErr := s.pl.GetPlaybookDataProviders(ctx)
	if playbooksErr == nil {
		for _, playbooks := range playbooksProviders {
			if syncErr := syncPlaybooks(ctx, s.db, playbooks); syncErr != nil {
				return fmt.Errorf("playbooks: %w", syncErr)
			}
		}
	}

	incidentsProviders, incidentsErr := s.pl.GetIncidentDataProviders(ctx)
	if incidentsErr == nil {
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

func (s *Syncer) SyncUserData(ctx context.Context) error {
	usersProviders, usersErr := s.pl.GetUserDataProviders(ctx)
	if usersErr == nil {
		for _, prov := range usersProviders {
			if syncErr := syncUsers(ctx, s.db, prov); syncErr != nil {
				return fmt.Errorf("user provider (name): %w", syncErr)
			}
		}
	}
	return nil
}
