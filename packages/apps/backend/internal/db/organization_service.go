package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organizationpreferences"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/jobs"
)

type OrganizationService struct {
	db   rez.Database
	jobs rez.JobService
}

func NewOrganizationService(db rez.Database, jobs rez.JobService) *OrganizationService {
	return &OrganizationService{db: db, jobs: jobs}
}

func (s *OrganizationService) Get(ctx context.Context, p predicate.Organization) (*ent.Organization, error) {
	query := s.db.Client(ctx).Organization.Query().
		Where(p).
		WithPreferences()
	return query.Only(ctx)
}

func (s *OrganizationService) Set(ctx context.Context, id uuid.UUID, setFn func(*ent.OrganizationMutation)) (*ent.Organization, error) {
	var res *ent.Organization
	return res, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.Organization, *ent.OrganizationMutation]
		if id == uuid.Nil {
			mutator = tx.Organization.Create()
		} else {
			mutator = tx.Organization.UpdateOneID(id)
		}

		setFn(mutator.Mutation())

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save: %w", saveErr)
		}
		res = saved.Unwrap()
		return nil
	})
}

var orgInitialSetupIntegrationSyncJob = jobs.SyncIntegrationEventsArgs{SyncReason: "org_initial_setup"}

func (s *OrganizationService) SetPreferences(ctx context.Context, orgId uuid.UUID, setFn func(*ent.OrganizationPreferencesMutation)) (*ent.OrganizationPreferences, error) {
	var prefs *ent.OrganizationPreferences
	return prefs, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		queryExisting := tx.OrganizationPreferences.Query().
			Where(organizationpreferences.OrganizationID(orgId))
		curr, queryErr := queryExisting.Only(ctx)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return queryErr
		}

		var mutator ent.EntityMutator[*ent.OrganizationPreferences, *ent.OrganizationPreferencesMutation]
		if curr != nil {
			mutator = curr.Update()
		} else {
			mutator = tx.OrganizationPreferences.Create().
				SetOrganizationID(orgId)
		}
		m := mutator.Mutation()
		setFn(m)

		updated, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save: %w", saveErr)
		}

		wasInitialSetup := !updated.InitialSetupAt.IsZero() && (curr == nil || curr.InitialSetupAt.IsZero())
		if wasInitialSetup {
			if _, jobErr := s.jobs.Insert(ctx, orgInitialSetupIntegrationSyncJob, nil); jobErr != nil {
				slog.Error("failed to insert org sync job", "error", jobErr)
			}
		}
		prefs = updated
		return nil
	})
}
