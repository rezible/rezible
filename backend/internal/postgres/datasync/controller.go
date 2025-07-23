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
	args := &jobs.SyncProviderData{
		Users:            true,
		Teams:            true,
		Incidents:        true,
		Oncall:           true,
		OncallEvents:     true,
		SystemComponents: true,
	}

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

	if args.Teams {
		teams, teamsErr := s.pl.LoadTeamDataProvider(ctx)
		if teamsErr != nil {
			return fmt.Errorf("failed to load teams data provider: %w", teamsErr)
		}
		if syncErr := syncTeams(ctx, s.db, teams); syncErr != nil {
			return fmt.Errorf("teams: %w", syncErr)
		}
	}

	if args.Users {
		users, provErr := s.pl.LoadUserDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load users data provider: %w", provErr)
		}
		if syncErr := syncUsers(ctx, s.db, users); syncErr != nil {
			return fmt.Errorf("users: %w", syncErr)
		}
	}

	if args.Oncall {
		oncall, provErr := s.pl.LoadOncallDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		if syncErr := syncOncallRosters(ctx, s.db, oncall); syncErr != nil {
			return fmt.Errorf("oncall rosters: %w", syncErr)
		}
		if syncErr := syncOncallShifts(ctx, s.db, oncall); syncErr != nil {
			return fmt.Errorf("oncall shifts: %w", syncErr)
		}
	}

	if args.OncallEvents {
		alerts, provErr := s.pl.LoadAlertDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall events data provider: %w", provErr)
		}
		if syncErr := syncOncallEvents(ctx, s.db, alerts); syncErr != nil {
			return fmt.Errorf("oncall events: %w", syncErr)
		}
	}

	if args.Incidents {
		prov, provErr := s.pl.LoadIncidentDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		if syncErr := syncIncidentRoles(ctx, s.db, prov); syncErr != nil {
			return fmt.Errorf("incident roles: %w", syncErr)
		}
		if syncErr := syncIncidents(ctx, s.db, prov); syncErr != nil {
			return fmt.Errorf("incidents: %w", syncErr)
		}
	}

	if args.SystemComponents {
		prov, provErr := s.pl.LoadSystemComponentsDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		if syncErr := syncSystemComponents(ctx, s.db, prov); syncErr != nil {
			return fmt.Errorf("system components: %w", syncErr)
		}
	}

	return nil
}
