package datasync

import (
	"context"
	"fmt"
	"iter"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentroleassignment"
)

func syncIncidents(ctx context.Context, db *ent.Client, prov rez.IncidentDataProvider) error {
	b := &incidentBatcher{db: db, provider: prov}
	s := newBatchedDataSyncer[*ent.Incident](db, "incidents", b)
	return s.Sync(ctx)
}

type incidentBatcher struct {
	db       *ent.Client
	provider rez.IncidentDataProvider

	slugs              *slugTracker
	deletedIncidentIds []uuid.UUID
	provIdRoleMap      map[string]*ent.IncidentRole
	provIdSeverityMap  map[string]*ent.IncidentSeverity
}

func (b *incidentBatcher) setup(ctx context.Context) error {
	b.deletedIncidentIds = make([]uuid.UUID, 0)
	b.slugs = newSlugTracker()

	b.provIdRoleMap = make(map[string]*ent.IncidentRole)
	roles, rolesErr := b.db.IncidentRole.Query().All(ctx)
	if rolesErr != nil {
		return fmt.Errorf("incident roles: %w", rolesErr)
	}
	for _, r := range roles {
		role := r
		b.provIdRoleMap[role.ProviderID] = role
	}

	b.provIdSeverityMap = make(map[string]*ent.IncidentSeverity)
	severities, sevsErr := b.db.IncidentSeverity.Query().All(ctx)
	if sevsErr != nil {
		return fmt.Errorf("incident severities: %w", rolesErr)
	}
	for _, s := range severities {
		sev := s
		b.provIdSeverityMap[sev.ProviderID] = sev
	}

	return nil
}

func (b *incidentBatcher) pullData(ctx context.Context) iter.Seq2[*ent.Incident, error] {
	return b.provider.PullIncidents(ctx)
}

func (b *incidentBatcher) queryDbIncidents(ctx context.Context, provIds []string) ([]*ent.Incident, error) {
	dbIncidents, incErr := b.db.Incident.Query().
		Where(incident.ProviderIDIn(provIds...)).
		WithRoleAssignments().
		All(ctx)
	if incErr != nil && !ent.IsNotFound(incErr) {
		return nil, incErr
	}
	return dbIncidents, nil
}

func (b *incidentBatcher) getDeletionMutations() []ent.Mutation {
	if len(b.deletedIncidentIds) > 0 {
		fmt.Printf("delete incidents: %v\n", b.deletedIncidentIds)
	}
	return nil
}

func (b *incidentBatcher) createBatchMutations(ctx context.Context, batch []*ent.Incident) ([]ent.Mutation, error) {
	provIds := make([]string, len(batch))
	for i, p := range batch {
		provIds[i] = p.ProviderID
	}

	dbIncidents, dbIncErr := b.queryDbIncidents(ctx, provIds)
	if dbIncErr != nil {
		return nil, fmt.Errorf("querying db incidents: %w", dbIncErr)
	}

	dbProvIdMap := make(map[string]*ent.Incident)
	syncIds := mapset.NewSet[string](provIds...)
	deletedIds := mapset.NewSet[uuid.UUID]()

	lastSyncTime := time.Time{}
	for _, dbInc := range dbIncidents {
		inc := dbInc
		needsSync := dbInc.ModifiedAt.IsZero() || inc.ModifiedAt.After(lastSyncTime)
		if !needsSync {
			syncIds.Remove(inc.ProviderID)
		} else {
			deletedIds.Add(inc.ID)
			dbProvIdMap[inc.ProviderID] = inc
		}
	}
	if syncIds.IsEmpty() {
		return nil, nil
	}

	var muts []ent.Mutation
	for _, provInc := range batch {
		dbInc, incExists := dbProvIdMap[provInc.ProviderID]
		if incExists {
			deletedIds.Remove(dbInc.ID)
		}

		incId, incMut, incSyncErr := b.syncIncident(ctx, dbInc, provInc)
		if incSyncErr != nil {
			return nil, fmt.Errorf("incident: %w", incSyncErr)
		} else if incMut != nil {
			muts = append(muts, incMut)
		}

		provInc.ID = incId
		assnMuts, assnsErr := b.syncIncidentRoleAssignments(ctx, dbInc, provInc)
		if assnsErr != nil {
			return nil, fmt.Errorf("role assignments: %w", assnsErr)
		} else if assnMuts != nil {
			muts = append(muts, assnMuts...)
		}
	}

	b.deletedIncidentIds = append(b.deletedIncidentIds, deletedIds.ToSlice()...)

	return muts, nil
}

func (b *incidentBatcher) makeIncidentSlugCountFn(ctx context.Context) func(prefix string) (int, error) {
	return func(prefix string) (int, error) {
		return b.db.Incident.Query().Where(incident.SlugHasPrefix(prefix)).Count(ctx)
	}
}

func (b *incidentBatcher) syncIncident(ctx context.Context, db, prov *ent.Incident) (uuid.UUID, *ent.IncidentMutation, error) {
	var m *ent.IncidentMutation
	var incId uuid.UUID
	if db == nil {
		incId = uuid.New()
		m = b.db.Incident.Create().SetID(incId).Mutation()
	} else {
		incId = db.ID
		m = b.db.Incident.UpdateOneID(incId).Mutation()

		// TODO: get provider mapping support for fields
		needsSync := db.Title != prov.Title
		if !needsSync {
			return incId, nil, nil
		}
	}

	m.SetTitle(prov.Title)
	m.SetSummary(prov.Summary)
	m.SetProviderID(prov.ProviderID)
	m.SetChatChannelID(prov.ChatChannelID)
	m.SetOpenedAt(prov.OpenedAt)
	m.SetModifiedAt(prov.ModifiedAt)
	m.SetClosedAt(prov.ClosedAt)

	slug, slugErr := b.slugs.generateUnique(prov.Title, b.makeIncidentSlugCountFn(ctx))
	if slugErr != nil {
		return incId, nil, fmt.Errorf("failed to create unique incident slug: %w", slugErr)
	}
	m.SetSlug(slug)

	return incId, m, nil
}

func (b *incidentBatcher) userRoleAssignmentKey(userId, roleId uuid.UUID) string {
	return userId.String() + roleId.String()
}

func (b *incidentBatcher) syncIncidentRoleAssignments(ctx context.Context, dbInc, provInc *ent.Incident) ([]ent.Mutation, error) {
	deletedIds := mapset.NewSet[uuid.UUID]()
	dbAssns := make(map[string]*ent.IncidentRoleAssignment)
	if dbInc != nil {
		assns, queryErr := dbInc.Edges.RoleAssignmentsOrErr()
		if queryErr != nil {
			return nil, fmt.Errorf("incident role assignments edge: %w", queryErr)
		}
		for _, assn := range assns {
			a := assn
			dbAssns[b.userRoleAssignmentKey(a.UserID, a.RoleID)] = a
			deletedIds.Add(a.ID)
		}
	}

	var muts []ent.Mutation
	for _, assn := range provInc.Edges.RoleAssignments {
		a := assn
		a.IncidentID = provInc.ID

		provRole := a.Edges.Role
		provUser := a.Edges.User

		role, roleExists := b.provIdRoleMap[provRole.ProviderID]
		if !roleExists {
			log.Warn().Str("id", provRole.ProviderID).Msg("failed to lookup incident role")
			return nil, fmt.Errorf("missing db role for id: %s", provRole.ProviderID)
		}

		usr, usrErr := lookupProviderUser(ctx, b.db, provUser)
		if usrErr != nil {
			// log.Warn().Str("email", provUser.Email).Msg("failed to lookup incident role user by email")
			continue
		}

		dbAssn, exists := dbAssns[b.userRoleAssignmentKey(usr.ID, role.ID)]
		if exists {
			deletedIds.Remove(dbAssn.ID)
		} else {
			create := b.db.IncidentRoleAssignment.Create().
				SetIncidentID(provInc.ID).
				SetUserID(usr.ID).
				SetRoleID(role.ID)
			muts = append(muts, create.Mutation())
		}
	}

	if !deletedIds.IsEmpty() {
		var mut ent.IncidentRoleAssignmentMutation
		mut.SetOp(ent.OpDelete)
		mut.Where(incidentroleassignment.IDIn(deletedIds.ToSlice()...))
		muts = append(muts, &mut)
	}

	return muts, nil
}

func (b *incidentBatcher) syncIncidentSeverity(ctx context.Context, dbInc, provInc *ent.Incident) (*ent.IncidentSeverityMutation, error) {
	return nil, nil
}

func (b *incidentBatcher) syncIncidentType(ctx context.Context, dbInc, provInc *ent.Incident) (*ent.IncidentTypeMutation, error) {
	return nil, nil
}

func (b *incidentBatcher) syncIncidentEvents(ctx context.Context, dbInc, provInc *ent.Incident) (*ent.IncidentEventMutation, error) {
	return nil, nil
}
