package datasync

import (
	"context"
	"fmt"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/jobs"
	"time"
)

type ProviderSyncService struct {
	db *ent.Client
	pl rez.ProviderLoader
}

func NewProviderSyncService(db *ent.Client, pl rez.ProviderLoader) *ProviderSyncService {
	return &ProviderSyncService{db: db, pl: pl}
}

func (s *ProviderSyncService) MakeSyncProviderDataPeriodicJob() jobs.PeriodicJob {
	return jobs.PeriodicJob{
		ConstructorFunc: func() jobs.InsertJobParams {
			return jobs.InsertJobParams{
				Args: &jobs.SyncProviderData{},
				Uniqueness: &jobs.JobUniquenessOpts{
					ByState: jobs.NonCompletedJobStates,
				},
			}
		},
		Interval: time.Hour,
		Opts:     &jobs.PeriodicJobOpts{RunOnStart: true},
	}
}

func (s *ProviderSyncService) SyncProviderData(ctx context.Context, args jobs.SyncProviderData) error {
	if args.Hard {
		// TODO: maybe just pass a flag?
		s.db.ProviderSyncHistory.Delete().ExecX(ctx)
	}

	users, usersErr := s.pl.LoadUserDataProvider(ctx)
	if usersErr != nil {
		return fmt.Errorf("users data provider: %w", usersErr)
	}
	if syncErr := syncUsers(ctx, s.db, users); syncErr != nil {
		return fmt.Errorf("users: %w", syncErr)
	}

	teams, teamsErr := s.pl.LoadTeamDataProvider(ctx)
	if teamsErr != nil {
		return fmt.Errorf("teams data provider: %w", teamsErr)
	}
	if syncErr := syncTeams(ctx, s.db, teams); syncErr != nil {
		return fmt.Errorf("teams: %w", syncErr)
	}

	oncall, oncallErr := s.pl.LoadOncallDataProvider(ctx)
	if oncallErr != nil {
		return fmt.Errorf("oncall data provider: %w", oncallErr)
	}
	if syncErr := syncOncallRosters(ctx, s.db, oncall); syncErr != nil {
		return fmt.Errorf("oncall rosters: %w", syncErr)
	}
	if syncErr := syncOncallShifts(ctx, s.db, oncall); syncErr != nil {
		return fmt.Errorf("oncall shifts: %w", syncErr)
	}

	components, componentsErr := s.pl.LoadSystemComponentsDataProvider(ctx)
	if componentsErr != nil {
		return fmt.Errorf("system components data provider: %w", componentsErr)
	}
	if syncErr := syncSystemComponents(ctx, s.db, components); syncErr != nil {
		return fmt.Errorf("system components: %w", syncErr)
	}

	alerts, alertsErr := s.pl.LoadAlertDataProvider(ctx)
	if alertsErr != nil {
		return fmt.Errorf("alert data provider: %w", alertsErr)
	}
	if syncErr := syncAlerts(ctx, s.db, alerts); syncErr != nil {
		return fmt.Errorf("alerts: %w", syncErr)
	}

	playbooks, playbooksErr := s.pl.LoadPlaybookDataProvider(ctx)
	if playbooksErr != nil {
		return fmt.Errorf("playbooks data provider: %w", playbooksErr)
	}
	if syncErr := syncPlaybooks(ctx, s.db, playbooks); syncErr != nil {
		return fmt.Errorf("playbooks: %w", syncErr)
	}

	incidents, incidentsErr := s.pl.LoadIncidentDataProvider(ctx)
	if incidentsErr != nil {
		return fmt.Errorf("incidents data provider: %w", incidentsErr)
	}
	if syncErr := syncIncidentRoles(ctx, s.db, incidents); syncErr != nil {
		return fmt.Errorf("incident roles: %w", syncErr)
	}
	if syncErr := syncIncidents(ctx, s.db, incidents); syncErr != nil {
		return fmt.Errorf("incidents: %w", syncErr)
	}

	if syncErr := syncOncallEvents(ctx, s.db, alerts); syncErr != nil {
		return fmt.Errorf("oncall events: %w", syncErr)
	}

	return nil
}
