package postgres

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
	users    rez.UserService
	provider rez.IncidentDataProvider

	slugs         *slugTracker
	roleProvIdMap map[string]uuid.UUID

	mutations []ent.Mutation
}

func newIncidentDataSyncer(db *ent.Client, users rez.UserService, prov rez.IncidentDataProvider) *incidentDataSyncer {
	ds := &incidentDataSyncer{db: db, users: users, provider: prov, slugs: newSlugTracker()}
	ds.resetState()
	return ds
}

func (ds *incidentDataSyncer) resetState() {
	ds.mutations = make([]ent.Mutation, 0)
	ds.slugs.reset()
	ds.roleProvIdMap = make(map[string]uuid.UUID)
}

func (ds *incidentDataSyncer) saveSyncHistory(ctx context.Context, start time.Time, num int, dataType string) {
	historyErr := ds.db.ProviderSyncHistory.Create().
		SetStartedAt(start).
		SetFinishedAt(time.Now()).
		SetNumMutations(num).
		SetDataType(dataType).
		Exec(ctx)
	if historyErr != nil {
		log.Error().Err(historyErr).Msg("failed to save sync history")
	}
}

func (ds *incidentDataSyncer) syncProviderData(ctx context.Context) error {
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

		mut, needsUpdate := ds.syncIncidentRole(db, prov)
		if needsUpdate {
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

	ds.saveSyncHistory(ctx, start, numMutations, "incident_roles")

	return nil
}

func (ds *incidentDataSyncer) syncIncidentRole(db, prov *ent.IncidentRole) (*ent.IncidentRoleMutation, bool) {
	var m *ent.IncidentRoleMutation

	needsSync := true
	if db == nil {
		m = ds.db.IncidentRole.Create().SetID(uuid.New()).Mutation()
	} else {
		m = ds.db.IncidentRole.UpdateOneID(db.ID).Mutation()
		// TODO: get provider mapping support for fields
		needsSync = db.Name != prov.Name || db.Required != prov.Required
	}
	m.SetProviderID(prov.ProviderID)
	m.SetName(prov.Name)
	m.SetRequired(prov.Required)

	return m, needsSync
}

func (ds *incidentDataSyncer) syncAllProviderIncidents(ctx context.Context) error {
	var providerIdsBatch []string

	start := time.Now()
	var numMutations int

	batchSize := 10
	for inc, pullErr := range ds.provider.PullIncidents(ctx) {
		if pullErr != nil {
			return fmt.Errorf("pull: %w", pullErr)
		}

		providerIdsBatch = append(providerIdsBatch, inc.ProviderID)

		if len(providerIdsBatch) >= batchSize {
			batchMuts, syncErr := ds.syncProviderIncidentIds(ctx, providerIdsBatch)
			if syncErr != nil {
				return syncErr
			}
			numMutations += batchMuts
			providerIdsBatch = make([]string, 0)
		}
	}

	lastBatchMuts, batchErr := ds.syncProviderIncidentIds(ctx, providerIdsBatch)
	if batchErr != nil {
		return batchErr
	}
	numMutations += lastBatchMuts

	ds.saveSyncHistory(ctx, start, numMutations, "incidents")

	return nil
}

func (ds *incidentDataSyncer) syncProviderIncidentIds(ctx context.Context, provIds []string) (int, error) {
	if len(provIds) == 0 {
		return 0, nil
	}

	ds.resetState()
	syncErr := ds.createIncidentBatchSyncMutations(ctx, provIds)
	if syncErr != nil {
		return 0, fmt.Errorf("building mutations: %w", syncErr)
	}

	if applyErr := applySyncMutations(ctx, ds.db, ds.mutations); applyErr != nil {
		return 0, fmt.Errorf("applying mutations: %w", applyErr)
	}
	numMutations := len(ds.mutations)
	ds.resetState()

	return numMutations, nil
}

func (ds *incidentDataSyncer) createIncidentBatchSyncMutations(ctx context.Context, provIds []string) error {
	dbIncidents, incErr := ds.db.Incident.Query().Where(incident.ProviderIDIn(provIds...)).All(ctx)
	if incErr != nil && !ent.IsNotFound(incErr) {
		return fmt.Errorf("querying db incidents: %w", incErr)
	}

	syncIds := mapset.NewSet(provIds...)
	incidentProvIdMap := make(map[string]*ent.Incident)
	deletedIncidentIds := mapset.NewSet[uuid.UUID]()

	lastSyncTime := time.Time{}
	for _, dbInc := range dbIncidents {
		inc := dbInc
		needsSync := dbInc.ModifiedAt.IsZero() || inc.ModifiedAt.After(lastSyncTime)
		if !needsSync {
			syncIds.Remove(inc.ProviderID)
		} else {
			deletedIncidentIds.Add(inc.ID)
			incidentProvIdMap[inc.ProviderID] = inc
		}
	}

	if syncIds.IsEmpty() {
		return nil
	}

	dbRoles, rolesErr := ds.db.IncidentRole.Query().All(ctx)
	if rolesErr != nil {
		return fmt.Errorf("db incident roles: %w", rolesErr)
	}
	for _, dbRole := range dbRoles {
		ds.roleProvIdMap[dbRole.ProviderID] = dbRole.ID
	}

	for _, provId := range provIds {
		provInc, fetchIncErr := ds.provider.GetIncidentByID(ctx, provId)
		if fetchIncErr != nil {
			return fmt.Errorf("fetching provider incident: %w", fetchIncErr)
		}

		dbInc, incExists := incidentProvIdMap[provId]
		if incExists {
			deletedIncidentIds.Remove(dbInc.ID)
		}

		incId, incSyncErr := ds.syncIncident(ctx, dbInc, provInc)
		if incSyncErr != nil {
			return fmt.Errorf("syncing incident: %w", incSyncErr)
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
			return fmt.Errorf("events: %w", eventsErr)
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

		roleId, roleExists := ds.roleProvIdMap[provRole.ProviderID]
		if !roleExists {
			log.Warn().Str("id", provRole.ProviderID).Msg("failed to lookup incident role")
			return fmt.Errorf("missing db role for id: %s", provRole.ProviderID)
		}

		usr, usrErr := ds.users.GetByEmail(ctx, provUser.Email)
		if usrErr != nil {
			log.Warn().Str("email", provUser.Email).Msg("failed to lookup incident role user by email")
			//return fmt.Errorf("role user: %w", usrErr)
			continue
		}

		dbAssn, exists := dbAssns[userRoleAssignmentKey(usr.ID, roleId)]
		if exists {
			deletedIds.Remove(dbAssn.ID)
		} else {
			create := ds.db.IncidentRoleAssignment.Create().
				SetIncidentID(provInc.ID).
				SetUserID(usr.ID).
				SetRoleID(roleId)
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
