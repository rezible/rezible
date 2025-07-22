package providers

import (
	"context"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentrole"
	"github.com/rezible/rezible/ent/incidentroleassignment"
)

type incidentDataSyncer struct {
	db       *ent.Client
	provider rez.IncidentDataProvider

	slugs             *slugTracker
	provIdRoleMap     map[string]*ent.IncidentRole
	provIdSeverityMap map[string]*ent.IncidentSeverity

	mutations []ent.Mutation
}

func newIncidentDataSyncer(db *ent.Client, prov rez.IncidentDataProvider) *incidentDataSyncer {
	return &incidentDataSyncer{db: db, provider: prov, slugs: newSlugTracker()}
}

func (ds *incidentDataSyncer) SyncProviderData(ctx context.Context) error {
	start := time.Now()

	lastSync := getLastSyncTime(ctx, ds.db, "incidents")
	if lastSync.Add(time.Minute * 30).After(start) {
		return nil
	}

	if rolesErr := ds.syncAllProviderIncidentRoles(ctx); rolesErr != nil {
		return fmt.Errorf("incident roles: %w", rolesErr)
	}
	rolesEnd := time.Now()

	if incidentsErr := ds.syncAllProviderIncidents(ctx); incidentsErr != nil {
		return fmt.Errorf("incidents: %w", incidentsErr)
	}
	log.Info().
		Int64("sync_roles_ms", rolesEnd.Sub(start).Milliseconds()).
		Int64("sync_incidents_ms", time.Since(rolesEnd).Milliseconds()).
		Msg("incidents data sync complete")

	return nil
}

func (ds *incidentDataSyncer) syncAllProviderIncidentRoles(ctx context.Context) error {
	var mutations []ent.Mutation

	start := time.Now()

	dbRoles, dbRolesErr := ds.db.IncidentRole.Query().All(ctx)
	if dbRolesErr != nil {
		return fmt.Errorf("listing incident roles: %w", dbRolesErr)
	}

	providerIdMap := make(map[string]*ent.IncidentRole)
	deletedIds := mapset.NewSet[uuid.UUID]()
	for _, role := range dbRoles {
		r := role
		providerIdMap[r.ProviderID] = r
		deletedIds.Add(role.ID)
	}

	provRoles, provRolesErr := ds.provider.GetRoles(ctx)
	if provRolesErr != nil {
		return fmt.Errorf("listing provider incident roles: %w", provRolesErr)
	}

	for _, r := range provRoles {
		prov := r
		db, exists := providerIdMap[prov.ProviderID]
		if exists {
			deletedIds.Remove(db.ID)
		}

		var mut *ent.IncidentRoleMutation
		needsSync := true
		if db == nil {
			mut = ds.db.IncidentRole.Create().SetID(uuid.New()).Mutation()
		} else {
			mut = ds.db.IncidentRole.UpdateOneID(db.ID).Mutation()
			// TODO: get provider mapping support for fields
			needsSync = db.Name != prov.Name || db.Required != prov.Required
		}
		mut.SetProviderID(prov.ProviderID)
		mut.SetName(prov.Name)
		mut.SetRequired(prov.Required)
		if needsSync {
			mutations = append(mutations, mut)
		}
	}

	if !deletedIds.IsEmpty() {
		var mut ent.IncidentRoleMutation
		mut.SetOp(ent.OpDelete)
		mut.Where(incidentrole.IDIn(deletedIds.ToSlice()...))
		mutations = append(mutations, &mut)
	}

	numMutations := len(mutations)
	if syncErr := applySyncMutations(ctx, ds.db, mutations); syncErr != nil {
		return fmt.Errorf("failed to sync incident roles: %w", syncErr)
	}

	if saveErr := saveSyncHistory(ctx, ds.db, start, numMutations, "incident_roles"); saveErr != nil {
		log.Error().Err(saveErr).Msg("failed to save incident roles data sync history")
	}

	return nil
}

func (ds *incidentDataSyncer) resetState() {
	ds.provIdRoleMap = make(map[string]*ent.IncidentRole)
	ds.provIdSeverityMap = make(map[string]*ent.IncidentSeverity)
	ds.slugs.reset()
	ds.mutations = make([]ent.Mutation, 0)
}

func (ds *incidentDataSyncer) loadFieldMaps(ctx context.Context) error {
	ds.resetState()

	roles, rolesErr := ds.db.IncidentRole.Query().All(ctx)
	if rolesErr != nil {
		return fmt.Errorf("incident roles: %w", rolesErr)
	}
	for _, r := range roles {
		role := r
		ds.provIdRoleMap[role.ProviderID] = role
	}

	severities, sevsErr := ds.db.IncidentSeverity.Query().All(ctx)
	if sevsErr != nil {
		return fmt.Errorf("incident severities: %w", rolesErr)
	}
	for _, s := range severities {
		sev := s
		ds.provIdSeverityMap[sev.ProviderID] = sev
	}

	return nil
}

func (ds *incidentDataSyncer) syncAllProviderIncidents(ctx context.Context) error {
	defer ds.resetState()
	if fieldsErr := ds.loadFieldMaps(ctx); fieldsErr != nil {
		return fmt.Errorf("loading incident fields: %w", fieldsErr)
	}

	var numMutations int
	syncBatch := func(batch []*ent.Incident) error {
		if len(batch) == 0 {
			return nil
		}
		syncErr := ds.createIncidentSyncMutations(ctx, batch)
		if syncErr != nil {
			return fmt.Errorf("building mutations: %w", syncErr)
		}

		if applyErr := applySyncMutations(ctx, ds.db, ds.mutations); applyErr != nil {
			return fmt.Errorf("applying mutations: %w", applyErr)
		}
		numMutations += len(ds.mutations)
		ds.mutations = make([]ent.Mutation, 0)

		return nil
	}

	start := time.Now()
	const BatchSize = 10
	var batch []*ent.Incident
	for inc, pullErr := range ds.provider.PullIncidents(ctx) {
		if pullErr != nil {
			return fmt.Errorf("pull: %w", pullErr)
		}

		batch = append(batch, inc)
		if len(batch) >= BatchSize {
			if syncErr := syncBatch(batch); syncErr != nil {
				return fmt.Errorf("sync: %w", syncErr)
			}
			batch = make([]*ent.Incident, 0)
		}
	}
	if syncErr := syncBatch(batch); syncErr != nil {
		return fmt.Errorf("sync: %w", syncErr)
	}

	if saveErr := saveSyncHistory(ctx, ds.db, start, numMutations, "incidents"); saveErr != nil {
		log.Error().Err(saveErr).Msg("failed to save incidents data sync history")
	}

	return nil
}

func (ds *incidentDataSyncer) queryDbIncidents(ctx context.Context, provIds []string) ([]*ent.Incident, error) {
	dbIncidents, incErr := ds.db.Incident.Query().
		Where(incident.ProviderIDIn(provIds...)).
		All(ctx)
	if incErr != nil && !ent.IsNotFound(incErr) {
		return nil, incErr
	}
	return dbIncidents, nil
}

func (ds *incidentDataSyncer) createIncidentSyncMutations(ctx context.Context, provIncs []*ent.Incident) error {
	provIds := make([]string, len(provIncs))
	for i, p := range provIncs {
		provIds[i] = p.ProviderID
	}

	dbIncidents, dbIncErr := ds.queryDbIncidents(ctx, provIds)
	if dbIncErr != nil {
		return fmt.Errorf("querying db incidents: %w", dbIncErr)
	}

	provIdIncidentMap := make(map[string]*ent.Incident)
	syncIds := mapset.NewSet[string](provIds...)
	deletedIncidentIds := mapset.NewSet[uuid.UUID]()

	lastSyncTime := time.Time{}
	for _, dbInc := range dbIncidents {
		inc := dbInc
		needsSync := dbInc.ModifiedAt.IsZero() || inc.ModifiedAt.After(lastSyncTime)
		if !needsSync {
			syncIds.Remove(inc.ProviderID)
		} else {
			deletedIncidentIds.Add(inc.ID)
			provIdIncidentMap[inc.ProviderID] = inc
		}
	}
	if syncIds.IsEmpty() {
		return nil
	}

	for _, provInc := range provIncs {
		dbInc, incExists := provIdIncidentMap[provInc.ProviderID]
		if incExists {
			deletedIncidentIds.Remove(dbInc.ID)
		}

		incId, incSyncErr := ds.syncIncident(ctx, dbInc, provInc)
		if incSyncErr != nil {
			return fmt.Errorf("incident: %w", incSyncErr)
		}
		provInc.ID = incId

		assnsErr := ds.syncIncidentRoleAssignments(ctx, dbInc, provInc)
		if assnsErr != nil {
			return fmt.Errorf("role assignments: %w", assnsErr)
		}

		sevErr := ds.syncIncidentSeverity(ctx, dbInc, provInc)
		if sevErr != nil {
			return fmt.Errorf("severity: %w", sevErr)
		}

		typeErr := ds.syncIncidentType(ctx, dbInc, provInc)
		if typeErr != nil {
			return fmt.Errorf("type: %w", typeErr)
		}

		eventsErr := ds.syncIncidentEvents(ctx, dbInc, provInc)
		if eventsErr != nil {
			return fmt.Errorf("incident events: %w", eventsErr)
		}
	}

	if !deletedIncidentIds.IsEmpty() {
		fmt.Printf("delete incidents: %v\n", deletedIncidentIds.ToSlice())
	}

	return nil
}

func (ds *incidentDataSyncer) syncIncident(ctx context.Context, db, prov *ent.Incident) (uuid.UUID, error) {
	var m *ent.IncidentMutation
	var incId uuid.UUID
	needsSync := true
	if db == nil {
		incId = uuid.New()
		m = ds.db.Incident.Create().SetID(incId).Mutation()
	} else {
		incId = db.ID
		m = ds.db.Incident.UpdateOneID(incId).Mutation()
		// TODO: get provider mapping support for fields
		needsSync = db.Title != prov.Title
	}
	slug, slugErr := ds.slugs.generateUnique(prov.Title, func(prefix string) (int, error) {
		return ds.db.Incident.Query().Where(incident.SlugHasPrefix(prefix)).Count(ctx)
	})
	if slugErr != nil {
		return uuid.Nil, fmt.Errorf("failed to create unique incident slug: %w", slugErr)
	}
	m.SetSlug(slug)
	m.SetTitle(prov.Title)
	m.SetSummary(prov.Summary)
	m.SetProviderID(prov.ProviderID)
	m.SetChatChannelID(prov.ChatChannelID)
	m.SetOpenedAt(prov.OpenedAt)
	m.SetModifiedAt(prov.ModifiedAt)
	m.SetClosedAt(prov.ClosedAt)

	if needsSync {
		ds.mutations = append(ds.mutations, m)
	}

	return incId, nil
}

func userRoleAssignmentKey(userId, roleId uuid.UUID) string {
	return userId.String() + roleId.String()
}

func (ds *incidentDataSyncer) syncIncidentRoleAssignments(ctx context.Context, dbInc, provInc *ent.Incident) error {
	deletedIds := mapset.NewSet[uuid.UUID]()
	dbAssns := make(map[string]*ent.IncidentRoleAssignment)
	if dbInc != nil {
		assns, queryErr := dbInc.QueryRoleAssignments().All(ctx)
		if queryErr != nil {
			return fmt.Errorf("querying incident role assignments: %w", queryErr)
		}
		for _, assn := range assns {
			a := assn
			dbAssns[userRoleAssignmentKey(a.UserID, a.RoleID)] = a
			deletedIds.Add(a.ID)
		}
	}

	for _, assn := range provInc.Edges.RoleAssignments {
		a := assn
		a.IncidentID = provInc.ID

		provRole := a.Edges.Role
		provUser := a.Edges.User

		role, roleExists := ds.provIdRoleMap[provRole.ProviderID]
		if !roleExists {
			log.Warn().Str("id", provRole.ProviderID).Msg("failed to lookup incident role")
			return fmt.Errorf("missing db role for id: %s", provRole.ProviderID)
		}

		usr, usrErr := lookupProviderUser(ctx, ds.db, provUser)
		if usrErr != nil {
			log.Warn().Str("email", provUser.Email).Msg("failed to lookup incident role user by email")
			//return fmt.Errorf("role user: %w", usrErr)
			continue
		}

		dbAssn, exists := dbAssns[userRoleAssignmentKey(usr.ID, role.ID)]
		if exists {
			deletedIds.Remove(dbAssn.ID)
		} else {
			create := ds.db.IncidentRoleAssignment.Create().
				SetIncidentID(provInc.ID).
				SetUserID(usr.ID).
				SetRoleID(role.ID)
			ds.mutations = append(ds.mutations, create.Mutation())
		}
	}

	if !deletedIds.IsEmpty() {
		var mut ent.IncidentRoleAssignmentMutation
		mut.SetOp(ent.OpDelete)
		mut.Where(incidentroleassignment.IDIn(deletedIds.ToSlice()...))
		ds.mutations = append(ds.mutations, &mut)
	}

	return nil
}

func (ds *incidentDataSyncer) syncIncidentSeverity(ctx context.Context, dbInc, provInc *ent.Incident) error {
	var mutations []*ent.IncidentSeverityMutation

	// TODO: sync severity
	for _, m := range mutations {
		ds.mutations = append(ds.mutations, m)
	}
	return nil
}

func (ds *incidentDataSyncer) syncIncidentType(ctx context.Context, dbInc, provInc *ent.Incident) error {
	var mutations []*ent.IncidentTypeMutation
	// TODO: sync types
	for _, m := range mutations {
		ds.mutations = append(ds.mutations, m)
	}
	return nil
}

func (ds *incidentDataSyncer) syncIncidentEvents(ctx context.Context, dbInc, provInc *ent.Incident) error {
	var mutations []*ent.IncidentEventMutation
	// TODO: sync events
	for _, m := range mutations {
		ds.mutations = append(ds.mutations, m)
	}
	return nil
}
