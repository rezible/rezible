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
	im "github.com/rezible/rezible/ent/incidentmilestone"
	ira "github.com/rezible/rezible/ent/incidentroleassignment"
	"github.com/rezible/rezible/ent/schema"
	"github.com/rezible/rezible/ent/sequencecounter"
	"github.com/rezible/rezible/ent/situation"
	"github.com/rezible/rezible/pkg/errs"
)

var (
	errIncidentNeedsTitle    = errs.New(errs.CodeInvalidInput, "An incident needs a title.")
	errIncidentManagementOff = errs.New(errs.CodeConflict, "Incident management is off for this organization.")
	errIncidentResolved      = errs.New(errs.CodeConflict, "This incident is resolved. Its response can't change.")
	errResolutionNotLatest   = errs.New(errs.CodeConflict, "The resolution can't be before an impact or mitigation.")
)

// IncidentManagementService runs incident response: declaring, milestones and roles. Every write checks that
// incident management is on, inside its transaction.
type IncidentManagementService struct {
	db             rez.Database
	msgs           rez.MessageQueue
	orgs           rez.OrganizationService
	situations     rez.SituationService
	retrospectives rez.RetrospectiveService
	details        rez.IncidentDetailsService
}

func NewIncidentManagementService(db rez.Database, msgs rez.MessageQueue, orgs rez.OrganizationService, situations rez.SituationService, retrospectives rez.RetrospectiveService, details rez.IncidentDetailsService) (*IncidentManagementService, error) {
	svc := &IncidentManagementService{
		db:             db,
		msgs:           msgs,
		orgs:           orgs,
		situations:     situations,
		retrospectives: retrospectives,
		details:        details,
	}
	return svc, nil
}

func (s *IncidentManagementService) checkEnabled(ctx context.Context) error {
	enabled, enabledErr := s.orgs.IsIncidentManagementEnabled(ctx)
	if enabledErr != nil {
		return enabledErr
	}
	if !enabled {
		return errIncidentManagementOff
	}
	return nil
}

func (s *IncidentManagementService) Declare(ctx context.Context, params rez.DeclareIncidentParams) (*ent.Incident, error) {
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return nil, errIncidentNeedsTitle
	}
	if params.DeclaredByID == uuid.Nil {
		return nil, fmt.Errorf("%w: an incident needs its declarer", errs.ErrInvalidInput)
	}
	id, txErr := ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (uuid.UUID, error) {
		orgID, orgErr := tx.Organization.Query().OnlyID(ctx)
		if orgErr != nil {
			return uuid.Nil, fmt.Errorf("query organization: %w", orgErr)
		}
		// Turning the preference off takes the same lock, so no incident is declared after it is off.
		if lockErr := s.db.AcquireTxLocks(ctx, "incident_management", orgID.String()); lockErr != nil {
			return uuid.Nil, fmt.Errorf("lock incident management: %w", lockErr)
		}
		if enabledErr := s.checkEnabled(ctx); enabledErr != nil {
			return uuid.Nil, enabledErr
		}

		number, numberErr := tx.SequenceCounter.Next(ctx, sequencecounter.KindIncident)
		if numberErr != nil {
			return uuid.Nil, fmt.Errorf("allocate incident number: %w", numberErr)
		}
		create := tx.Incident.Create().
			SetNumber(number).
			SetTitle(title).
			SetSummary(params.Summary).
			SetNillableSeverityID(params.SeverityID).
			SetNillableTypeID(params.TypeID).
			SetOpenedAt(time.Now()).
			SetDeclaredByID(params.DeclaredByID).
			AddFieldSelectionIDs(params.FieldOptionIDs...).
			AddTagAssignmentIDs(params.TagIDs...)
		created, createErr := create.Save(ctx)
		if createErr != nil {
			return uuid.Nil, fmt.Errorf("create incident: %w", createErr)
		}

		links := rez.IncidentSituationLinkChanges{Added: params.SituationIDs}
		if syncErr := s.situations.SyncIncidentLinks(ctx, created.ID, links); syncErr != nil {
			return uuid.Nil, fmt.Errorf("link incident situations: %w", syncErr)
		}
		if publishErr := s.msgs.Publish(ctx, rez.EventOnIncidentUpdated{Created: true, IncidentId: created.ID}); publishErr != nil {
			return uuid.Nil, fmt.Errorf("publish incident update: %w", publishErr)
		}
		return created.ID, nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return s.details.GetIncident(ctx, incident.ID(id))
}

func (s *IncidentManagementService) AddMilestone(ctx context.Context, params rez.AddIncidentMilestoneParams) (*ent.IncidentMilestone, error) {
	if kindErr := im.KindValidator(params.Kind); kindErr != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidInput, kindErr)
	}
	timestamp := params.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	added := &ent.IncidentMilestone{ID: uuid.New(), IncidentID: params.IncidentID, Kind: params.Kind, Timestamp: timestamp}
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.IncidentMilestone, error) {
		if enabledErr := s.checkEnabled(ctx); enabledErr != nil {
			return nil, enabledErr
		}
		before, lockErr := s.lockIncidentMilestones(ctx, tx, params.IncidentID)
		if lockErr != nil {
			return nil, lockErr
		}
		after := append(ent.IncidentMilestones{added}, before...)
		if checkErr := s.checkMilestoneChange(before, after); checkErr != nil {
			return nil, checkErr
		}

		create := tx.IncidentMilestone.Create().
			SetID(added.ID).
			SetIncidentID(params.IncidentID).
			SetKind(params.Kind).
			SetTimestamp(timestamp).
			SetDescription(params.Description).
			SetNillableUserID(params.UserID).
			SetSource(params.Source)
		created, createErr := create.Save(ctx)
		if createErr != nil {
			return nil, fmt.Errorf("create incident milestone: %w", createErr)
		}
		if afterErr := s.afterMilestoneChange(ctx, before, after, created, true); afterErr != nil {
			return nil, afterErr
		}
		return created, nil
	})
}

func (s *IncidentManagementService) UpdateMilestone(ctx context.Context, id uuid.UUID, params rez.UpdateIncidentMilestoneParams) (*ent.IncidentMilestone, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.IncidentMilestone, error) {
		if enabledErr := s.checkEnabled(ctx); enabledErr != nil {
			return nil, enabledErr
		}
		incidentID, queryErr := tx.IncidentMilestone.Query().Where(im.ID(id)).QueryIncident().OnlyID(ctx)
		if queryErr != nil {
			return nil, fmt.Errorf("query milestone incident: %w", queryErr)
		}
		before, lockErr := s.lockIncidentMilestones(ctx, tx, incidentID)
		if lockErr != nil {
			return nil, lockErr
		}
		after := make(ent.IncidentMilestones, len(before))
		for i, milestone := range before {
			after[i] = milestone
			if milestone.ID == id && params.Timestamp != nil {
				changed := *milestone
				changed.Timestamp = *params.Timestamp
				after[i] = &changed
			}
		}
		if checkErr := s.checkMilestoneChange(before, after); checkErr != nil {
			return nil, checkErr
		}

		update := tx.IncidentMilestone.UpdateOneID(id).
			SetNillableTimestamp(params.Timestamp).
			SetNillableDescription(params.Description)
		updated, updateErr := update.Save(ctx)
		if updateErr != nil {
			return nil, fmt.Errorf("update incident milestone: %w", updateErr)
		}
		if afterErr := s.afterMilestoneChange(ctx, before, after, updated, false); afterErr != nil {
			return nil, afterErr
		}
		return updated, nil
	})
}

// lockIncidentMilestones locks the incident row, so milestone writes to one incident are checked in turn, and
// returns its milestones.
func (s *IncidentManagementService) lockIncidentMilestones(ctx context.Context, tx *ent.Client, incidentID uuid.UUID) (ent.IncidentMilestones, error) {
	lockIncident := tx.Incident.Query().
		Where(incident.ID(incidentID)).
		ForUpdate()
	if _, lockErr := lockIncident.Only(ctx); lockErr != nil {
		return nil, fmt.Errorf("lock incident: %w", lockErr)
	}
	milestones, queryErr := tx.IncidentMilestone.Query().Where(im.IncidentID(incidentID)).All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query incident milestones: %w", queryErr)
	}
	return milestones, nil
}

// checkMilestoneChange rejects a change to a resolved incident's response, and an impact or mitigation dated
// after the resolution.
func (s *IncidentManagementService) checkMilestoneChange(before, after ent.IncidentMilestones) error {
	if resolution := before.GetLatest(im.KindResolution); resolution != nil {
		resolutions := 0
		for _, milestone := range after {
			if milestone.Kind == im.KindResolution {
				resolutions++
			}
		}
		if resolutions > 1 || !after.GetLatest(im.KindResolution).Timestamp.Equal(resolution.Timestamp) {
			return errIncidentResolved
		}
	}
	if resolution := after.GetLatest(im.KindResolution); resolution != nil {
		latest := after.GetLatest(im.KindImpact, im.KindMitigation)
		if latest != nil && latest.Timestamp.After(resolution.Timestamp) {
			return errResolutionNotLatest
		}
	}
	return nil
}

// afterMilestoneChange stores the incident's response state, creates the retrospective when the change
// resolves the incident, and refreshes what is shown about its linked situations, open or closed. It lets the
// linked situations re-evaluate, and publishes the milestone.
func (s *IncidentManagementService) afterMilestoneChange(ctx context.Context, before, after ent.IncidentMilestones, milestone *ent.IncidentMilestone, created bool) error {
	state, resolvedAt := s.responseState(after)
	update := s.db.Client(ctx).Incident.UpdateOneID(milestone.IncidentID).
		SetResponseState(state).
		SetNillableResolvedAt(resolvedAt)
	if updateErr := update.Exec(ctx); updateErr != nil {
		return fmt.Errorf("update incident response state: %w", updateErr)
	}
	if state == incident.ResponseStateResolved && before.GetLatest(im.KindResolution) == nil {
		if _, retroErr := s.retrospectives.CreateForIncident(ctx, milestone.IncidentID); retroErr != nil {
			return fmt.Errorf("create incident retrospective: %w", retroErr)
		}
		queryLinked := s.db.Client(ctx).Situation.Query().
			Where(situation.HasIncidentsWith(incident.ID(milestone.IncidentID)))
		linkedIDs, linkedErr := queryLinked.IDs(ctx)
		if linkedErr != nil {
			return fmt.Errorf("query incident situations: %w", linkedErr)
		}
		// Resolving lets closed situations show the resolution; evaluation may publish nothing for them.
		for _, situationID := range linkedIDs {
			if publishErr := s.msgs.Publish(ctx, rez.EventOnSituationUpdated{SituationID: situationID}); publishErr != nil {
				return fmt.Errorf("publish situation update: %w", publishErr)
			}
		}
	}
	if syncErr := s.situations.SyncIncidentLinks(ctx, milestone.IncidentID, rez.IncidentSituationLinkChanges{}); syncErr != nil {
		return fmt.Errorf("sync incident situations: %w", syncErr)
	}
	updatedEvent := rez.EventOnIncidentMilestoneUpdated{
		Created:     created,
		IncidentId:  milestone.IncidentID,
		MilestoneId: milestone.ID,
	}
	if publishErr := s.msgs.Publish(ctx, updatedEvent); publishErr != nil {
		return fmt.Errorf("publish incident milestone update: %w", publishErr)
	}
	return nil
}

// responseState is resolved once there is a resolution milestone, at its time. Otherwise the newest impact or
// mitigation milestone decides: mitigated for a mitigation, started for an impact or when there is neither.
func (s *IncidentManagementService) responseState(milestones ent.IncidentMilestones) (incident.ResponseState, *time.Time) {
	if resolution := milestones.GetLatest(im.KindResolution); resolution != nil {
		return incident.ResponseStateResolved, &resolution.Timestamp
	}
	latest := milestones.GetLatest(im.KindImpact, im.KindMitigation)
	if latest != nil && latest.Kind == im.KindMitigation {
		return incident.ResponseStateMitigated, nil
	}
	return incident.ResponseStateStarted, nil
}

// SetRoleAssignment assigns the role to the user. Assigning a role the user already holds returns the existing
// assignment and publishes nothing.
func (s *IncidentManagementService) SetRoleAssignment(ctx context.Context, incidentID uuid.UUID, roleID uuid.UUID, userID uuid.UUID) (*ent.IncidentRoleAssignment, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.IncidentRoleAssignment, error) {
		if enabledErr := s.checkEnabled(ctx); enabledErr != nil {
			return nil, enabledErr
		}
		// The incident row lock keeps concurrent assignments within the role's limit.
		lockIncident := tx.Incident.Query().
			Where(incident.ID(incidentID)).
			ForUpdate()
		if _, lockErr := lockIncident.Only(ctx); lockErr != nil {
			return nil, fmt.Errorf("lock incident: %w", lockErr)
		}
		role, roleErr := tx.IncidentRole.Get(schema.IncludeArchived(ctx), roleID)
		if roleErr != nil {
			return nil, fmt.Errorf("get incident role: %w", roleErr)
		}
		if role.ArchiveTime != nil {
			return nil, errs.New(errs.CodeInvalidInput, fmt.Sprintf("%s is archived.", role.Name))
		}

		queryExisting := tx.IncidentRoleAssignment.Query().
			Where(ira.IncidentID(incidentID), ira.RoleID(roleID), ira.UserID(userID)).
			WithRole().
			WithUser()
		existing, existingErr := queryExisting.Only(ctx)
		if existingErr == nil {
			return existing, nil
		}
		if !ent.IsNotFound(existingErr) {
			return nil, fmt.Errorf("query role assignment: %w", existingErr)
		}

		if role.MaxAssignees != nil {
			queryHolders := tx.IncidentRoleAssignment.Query().
				Where(ira.IncidentID(incidentID), ira.RoleID(roleID))
			holders, countErr := queryHolders.Count(ctx)
			if countErr != nil {
				return nil, fmt.Errorf("count role assignments: %w", countErr)
			}
			if holders >= *role.MaxAssignees {
				people := "people"
				if *role.MaxAssignees == 1 {
					people = "person"
				}
				return nil, errs.New(errs.CodeConflict, fmt.Sprintf("%s can be held by %d %s in an incident.", role.Name, *role.MaxAssignees, people))
			}
		}

		create := tx.IncidentRoleAssignment.Create().
			SetIncidentID(incidentID).
			SetRoleID(roleID).
			SetUserID(userID)
		created, createErr := create.Save(ctx)
		if createErr != nil {
			return nil, fmt.Errorf("create role assignment: %w", createErr)
		}
		if publishErr := s.msgs.Publish(ctx, rez.EventOnIncidentUpdated{IncidentId: incidentID}); publishErr != nil {
			return nil, fmt.Errorf("publish incident update: %w", publishErr)
		}
		return tx.IncidentRoleAssignment.Query().
			Where(ira.ID(created.ID)).
			WithRole().
			WithUser().
			Only(ctx)
	})
}

func (s *IncidentManagementService) RemoveRoleAssignment(ctx context.Context, assignmentID uuid.UUID) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if enabledErr := s.checkEnabled(ctx); enabledErr != nil {
			return enabledErr
		}
		assignment, queryErr := tx.IncidentRoleAssignment.Get(ctx, assignmentID)
		if queryErr != nil {
			return fmt.Errorf("get role assignment: %w", queryErr)
		}
		if deleteErr := tx.IncidentRoleAssignment.DeleteOneID(assignmentID).Exec(ctx); deleteErr != nil {
			return fmt.Errorf("delete role assignment: %w", deleteErr)
		}
		if publishErr := s.msgs.Publish(ctx, rez.EventOnIncidentUpdated{IncidentId: assignment.IncidentID}); publishErr != nil {
			return fmt.Errorf("publish incident update: %w", publishErr)
		}
		return nil
	})
}
