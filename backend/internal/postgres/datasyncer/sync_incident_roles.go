package datasyncer

import (
	"context"
	"fmt"
	"iter"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentrole"
)

func syncIncidentRoles(ctx context.Context, db *ent.Client, prov rez.IncidentDataProvider) error {
	b := &incidentRolesBatcher{db: db, provider: prov}
	s := newBatchedDataSyncer[*ent.IncidentRole](db, "incident_roles", b)
	return s.Sync(ctx)
}

type incidentRolesBatcher struct {
	db       *ent.Client
	provider rez.IncidentDataProvider

	dbProviderIdMap map[string]*ent.IncidentRole
	deletedIds      mapset.Set[uuid.UUID]
}

func newIncidentRolesSyncer(db *ent.Client, prov rez.IncidentDataProvider) *batchedDataSyncer[*ent.IncidentRole] {
	return newBatchedDataSyncer[*ent.IncidentRole](db, "incident_role", &incidentRolesBatcher{db: db, provider: prov})
}

func (b *incidentRolesBatcher) setup(ctx context.Context) error {
	dbRoles, dbRolesErr := b.db.IncidentRole.Query().All(ctx)
	if dbRolesErr != nil {
		return fmt.Errorf("querying db roles: %w", dbRolesErr)
	}

	b.dbProviderIdMap = make(map[string]*ent.IncidentRole)
	b.deletedIds = mapset.NewSet[uuid.UUID]()

	for _, role := range dbRoles {
		r := role
		b.dbProviderIdMap[r.ProviderID] = r
		b.deletedIds.Add(role.ID)
	}

	return nil
}

func (b *incidentRolesBatcher) pullData(ctx context.Context) iter.Seq2[*ent.IncidentRole, error] {
	roles, rolesErr := b.provider.GetRoles(ctx)
	return func(yield func(*ent.IncidentRole, error) bool) {
		if rolesErr != nil {
			yield(nil, rolesErr)
		} else {
			for _, role := range roles {
				if !yield(role, nil) {
					break
				}
			}
		}
	}
}

func (b *incidentRolesBatcher) createBatchMutations(ctx context.Context, batch []*ent.IncidentRole) ([]ent.Mutation, error) {
	var mutations []ent.Mutation

	for _, r := range batch {
		prov := r
		db, exists := b.dbProviderIdMap[prov.ProviderID]
		if exists {
			b.deletedIds.Remove(db.ID)
		}

		var mut *ent.IncidentRoleMutation
		needsSync := true
		if db == nil {
			mut = b.db.IncidentRole.Create().SetID(uuid.New()).Mutation()
		} else {
			mut = b.db.IncidentRole.UpdateOneID(db.ID).Mutation()
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

	return mutations, nil
}

func (b *incidentRolesBatcher) getDeletionMutations() []ent.Mutation {
	var mutations []ent.Mutation

	if !b.deletedIds.IsEmpty() {
		var mut ent.IncidentRoleMutation
		mut.SetOp(ent.OpDelete)
		mut.Where(incidentrole.IDIn(b.deletedIds.ToSlice()...))
		mutations = append(mutations, &mut)
	}

	return mutations
}
