package datasync

import (
	"context"
	"fmt"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/jobs"
	"time"

	"github.com/rezible/rezible/ent"
)

type SyncController struct {
	db *ent.Client
	pl rez.ProviderLoader
}

var syncInterval = time.Hour

func NewSyncController(db *ent.Client, pl rez.ProviderLoader) *SyncController {
	return &SyncController{db: db, pl: pl}
}

func (s *SyncController) RegisterPeriodicSyncJob(j rez.JobsService) error {
	args := &jobs.SyncProviderData{}

	opts := &jobs.InsertOpts{
		Uniqueness: &jobs.UniquenessOpts{
			ByState: jobs.NonCompletedJobStates,
		},
	}

	job := jobs.NewPeriodicJob(
		jobs.PeriodicInterval(syncInterval),
		func() (jobs.JobArgs, *jobs.InsertOpts) {
			return args, opts
		},
		&jobs.PeriodicJobOpts{
			RunOnStart: true,
		},
	)

	j.RegisterPeriodicJob(job)

	return jobs.RegisterWorkerFunc(s.SyncData)
}

func (s *SyncController) SyncData(ctx context.Context, args jobs.SyncProviderData) error {
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
