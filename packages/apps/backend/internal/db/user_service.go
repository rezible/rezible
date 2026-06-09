package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organizationrole"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/execution"
)

type UserService struct {
	db        rez.Database
	orgs      rez.OrganizationService
	knowledge rez.KnowledgeService
}

func NewUserService(db rez.Database, orgs rez.OrganizationService, knowledge rez.KnowledgeService) (*UserService, error) {
	s := &UserService{
		db:        db,
		orgs:      orgs,
		knowledge: knowledge,
	}

	return s, nil
}

func (s *UserService) SyncFromAuthProvider(ctx context.Context, po ent.Organization, pu ent.User) (*ent.User, error) {
	// TODO: wrap this all in a transaction

	org, orgErr := s.orgs.SyncFromAuthProvider(ctx, po)
	if orgErr != nil {
		return nil, fmt.Errorf("sync organization: %w", orgErr)
	}
	ctx = execution.NewTenantContext(ctx, org.TenantID)

	existing, getErr := s.Get(ctx, user.AuthProviderID(pu.AuthProviderID))
	if getErr != nil && !ent.IsNotFound(getErr) {
		return nil, fmt.Errorf("query existing: %w", getErr)
	}

	var userId uuid.UUID
	if existing != nil {
		// TODO: check if should sync every time
		// if !AlwaysSyncAuthDetails { return existing, nil }
		userId = existing.ID
	}
	syncAuthDetailsFn := func(m *ent.UserMutation) {
		m.SetAuthProviderID(pu.AuthProviderID)
		m.SetEmail(pu.Email)
		m.SetName(pu.Name)
		m.SetTimezone(pu.Timezone)
	}
	usr, setErr := s.Set(ctx, userId, syncAuthDetailsFn)
	if setErr != nil {
		return nil, fmt.Errorf("set user: %w", setErr)
	}

	if org.InitialSetupAt.IsZero() {
		roles, rolesErr := org.QueryRoles().All(ctx)
		if rolesErr != nil {
			return nil, fmt.Errorf("query roles: %w", rolesErr)
		}

		var hasAdmin bool
		for _, role := range roles {
			if role.Role == organizationrole.RoleAdmin {
				hasAdmin = true
				break
			}
		}

		if !hasAdmin {
			// TODO: create initial admin user role
		}
	}

	return usr, nil
}

func (s *UserService) Get(ctx context.Context, p predicate.User) (*ent.User, error) {
	return s.db.Client(ctx).User.Query().Where(p).Only(ctx)
}

func (s *UserService) Set(ctx context.Context, id uuid.UUID, setFn func(*ent.UserMutation)) (*ent.User, error) {
	var savedUser *ent.User
	setTxFn := func(txCtx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.User, *ent.UserMutation]
		if id == uuid.Nil {
			mutator = tx.User.Create().SetID(uuid.New())
		} else {
			mutator = tx.User.UpdateOneID(id)
		}

		setFn(mutator.Mutation())

		var saveErr error
		savedUser, saveErr = mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save: %w", saveErr)
		}
		return nil
	}
	if txErr := s.db.WithTx(ctx, setTxFn); txErr != nil {
		return nil, txErr
	}
	return savedUser, nil
}

func (s *UserService) List(ctx context.Context, params rez.ListUsersParams) ([]*ent.User, error) {
	query := s.db.Client(ctx).User.Query().
		Order(user.ByID()).
		Limit(params.GetLimit()).
		Offset(params.Offset)

	if len(params.Search) > 0 {
		query = query.Where(user.NameContainsFold(params.Search))
	}
	if params.TeamID != uuid.Nil {
		query = query.Where(user.HasTeamsWith(team.ID(params.TeamID)))
	}

	res, queryErr := query.All(params.GetQueryContext(ctx))
	if queryErr != nil {
		return nil, fmt.Errorf("query users: %w", queryErr)
	}
	return res, nil
}
