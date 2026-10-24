package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/organizationpreferences"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/errs"
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
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.Organization, error) {
		var mutator ent.EntityMutator[*ent.Organization, *ent.OrganizationMutation]
		if id == uuid.Nil {
			mutator = tx.Organization.Create()
		} else {
			mutator = tx.Organization.UpdateOneID(id)
		}

		setFn(mutator.Mutation())

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save: %w", saveErr)
		}
		return saved, nil
	})
}

func (s *OrganizationService) CompleteOrgSetup(ctx context.Context, orgId uuid.UUID, params rez.CompleteOrgSetupParams) (*ent.Organization, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: organization name is required", errs.ErrInvalidInput)
	}

	timezone := strings.TrimSpace(params.Timezone)
	if timezone != "" {
		if timezone == "Local" {
			return nil, fmt.Errorf("%w: Local is not a valid organization timezone", errs.ErrInvalidInput)
		}
		if _, loadErr := time.LoadLocation(timezone); loadErr != nil {
			return nil, fmt.Errorf("%w: invalid organization timezone", errs.ErrInvalidInput)
		}
	}

	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.Organization, error) {
		if lockErr := s.db.AcquireTxLocks(ctx, "organization_setup", orgId.String()); lockErr != nil {
			return nil, fmt.Errorf("lock organization setup: %w", lockErr)
		}

		currOrg, queryOrgErr := s.Get(ctx, organization.ID(orgId))
		if queryOrgErr != nil {
			return nil, fmt.Errorf("query org: %w", queryOrgErr)
		}

		var prefsMut ent.EntityMutator[*ent.OrganizationPreferences, *ent.OrganizationPreferencesMutation]
		if currPrefs := currOrg.Edges.Preferences; currPrefs != nil {
			if !currPrefs.InitialSetupAt.IsZero() {
				return nil, fmt.Errorf("%w: organization setup already completed", errs.ErrConflict)
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
			return nil, fmt.Errorf("update preferences: %w", savePrefsErr)
		}

		updateOrg := currOrg.Update().
			SetName(name)
		if updateOrgErr := updateOrg.Exec(ctx); updateOrgErr != nil {
			return nil, fmt.Errorf("update org: %w", updateOrgErr)
		}

		updatedOrg, queryUpdatedOrgErr := s.Get(ctx, organization.ID(orgId))
		if queryUpdatedOrgErr != nil {
			return nil, fmt.Errorf("get org: %w", queryUpdatedOrgErr)
		}
		return updatedOrg, nil
	})
}

func (s *OrganizationService) SetPreferences(ctx context.Context, orgId uuid.UUID, setFn func(*ent.OrganizationPreferencesMutation)) (*ent.OrganizationPreferences, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.OrganizationPreferences, error) {
		queryExisting := tx.OrganizationPreferences.Query().
			Where(organizationpreferences.OrganizationID(orgId))
		curr, queryErr := queryExisting.Only(ctx)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return nil, queryErr
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
		if enabled, changed := m.EnableIncidentManagement(); changed {
			if checkErr := s.checkIncidentManagementChange(ctx, tx, orgId, enabled); checkErr != nil {
				return nil, checkErr
			}
		}

		updated, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save: %w", saveErr)
		}

		return updated, nil
	})
}

// IsIncidentManagementEnabled reads the preference without a lock. A missing preferences row means off.
func (s *OrganizationService) IsIncidentManagementEnabled(ctx context.Context) (bool, error) {
	prefs, queryErr := s.db.Client(ctx).OrganizationPreferences.Query().Only(ctx)
	if queryErr != nil {
		if ent.IsNotFound(queryErr) {
			return false, nil
		}
		return false, fmt.Errorf("query organization preferences: %w", queryErr)
	}
	return prefs.EnableIncidentManagement, nil
}

// checkIncidentManagementChange serializes a change of the preference with declaring incidents, which takes the
// same lock, and rejects turning it off while an incident is unresolved. The preferences row may not exist yet,
// so it cannot be the lock.
func (s *OrganizationService) checkIncidentManagementChange(ctx context.Context, tx *ent.Client, orgId uuid.UUID, enabled bool) error {
	if lockErr := s.db.AcquireTxLocks(ctx, "incident_management", orgId.String()); lockErr != nil {
		return fmt.Errorf("lock incident management: %w", lockErr)
	}
	if enabled {
		return nil
	}

	queryUnresolved := tx.Incident.Query().
		Where(incident.ResponseStateNEQ(incident.ResponseStateResolved))
	unresolved, countErr := queryUnresolved.Count(ctx)
	if countErr != nil {
		return fmt.Errorf("count unresolved incidents: %w", countErr)
	}
	if unresolved == 0 {
		return nil
	}
	message := fmt.Sprintf("Incident management can't be turned off while %d incidents are unresolved. Resolve them first.", unresolved)
	if unresolved == 1 {
		message = "Incident management can't be turned off while 1 incident is unresolved. Resolve it first."
	}
	return errs.New(errs.CodeConflict, message)
}
