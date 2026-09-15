package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/organizationpreferences"
	"github.com/rezible/rezible/ent/predicate"
)

type OrganizationService struct {
	db   rez.Database
	jobs rez.JobService
}

func NewOrganizationService(db rez.Database, jobs rez.JobService) (*OrganizationService, error) {
	return &OrganizationService{db: db, jobs: jobs}, nil
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

func (s *OrganizationService) CompleteOrgSetup(ctx context.Context, orgId uuid.UUID, params rez.CompleteOrgSetupParams) (*ent.Organization, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: organization name is required", rez.ErrInvalidInput)
	}

	timezone := strings.TrimSpace(params.Timezone)
	if timezone != "" {
		if timezone == "Local" {
			return nil, fmt.Errorf("%w: Local is not a valid organization timezone", rez.ErrInvalidInput)
		}
		if _, loadErr := time.LoadLocation(timezone); loadErr != nil {
			return nil, fmt.Errorf("%w: invalid organization timezone", rez.ErrInvalidInput)
		}
	}

	var org *ent.Organization
	return org, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if lockErr := s.db.AcquireTxLocks(ctx, "organization_setup", orgId.String()); lockErr != nil {
			return fmt.Errorf("lock organization setup: %w", lockErr)
		}

		currOrg, queryOrgErr := s.Get(ctx, organization.ID(orgId))
		if queryOrgErr != nil {
			return fmt.Errorf("query org: %w", queryOrgErr)
		}

		var prefsMut ent.EntityMutator[*ent.OrganizationPreferences, *ent.OrganizationPreferencesMutation]
		if currPrefs := currOrg.Edges.Preferences; currPrefs != nil {
			if !currPrefs.InitialSetupAt.IsZero() {
				return fmt.Errorf("%w: organization setup already completed", rez.ErrConflict)
			}

			prefsMut = currPrefs.Update()
		} else {
			prefsMut = tx.OrganizationPreferences.Create().
				SetOrganizationID(orgId)
		}
		pm := prefsMut.Mutation()
		pm.SetInitialSetupAt(time.Now().UTC())
		if timezone == "" {
			pm.ClearTimezone()
		} else {
			pm.SetTimezone(timezone)
		}
		if savePrefsErr := prefsMut.Exec(ctx); savePrefsErr != nil {
			return fmt.Errorf("update preferences: %w", savePrefsErr)
		}

		updateOrg := currOrg.Update().
			SetName(name)
		if updateOrgErr := updateOrg.Exec(ctx); updateOrgErr != nil {
			return fmt.Errorf("update org: %w", updateOrgErr)
		}

		updatedOrg, queryUpdatedOrgErr := s.Get(ctx, organization.ID(orgId))
		if queryUpdatedOrgErr != nil {
			return fmt.Errorf("get org: %w", queryUpdatedOrgErr)
		}
		org = updatedOrg.Unwrap()
		return nil
	})
}

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

		prefs = updated
		return nil
	})
}
