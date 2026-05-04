package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/execution"
)

type UserService struct {
	db   *ent.Client
	orgs rez.OrganizationService
}

func NewUserService(db *ent.Client, orgs rez.OrganizationService) (*UserService, error) {
	s := &UserService{
		db:   db,
		orgs: orgs,
	}

	return s, nil
}

func (s *UserService) SyncFromAuthProvider(ctx context.Context, po ent.Organization, pu ent.User) (*ent.User, error) {
	org, orgErr := s.orgs.SyncFromAuthProvider(ctx, po)
	if orgErr != nil {
		return nil, fmt.Errorf("sync organization: %w", orgErr)
	}
	ctx = execution.SystemTenantContext(ctx, org.TenantID)

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
	return s.Set(ctx, userId, syncAuthDetailsFn)
}

func (s *UserService) Get(ctx context.Context, p predicate.User) (*ent.User, error) {
	return s.db.User.Query().Where(p).Only(ctx)
}

func (s *UserService) Set(ctx context.Context, id uuid.UUID, setFn func(*ent.UserMutation)) (*ent.User, error) {
	var savedUser *ent.User
	setTxFn := func(tx *ent.Tx) error {
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
	if txErr := ent.WithTx(ctx, s.db, setTxFn); txErr != nil {
		return nil, txErr
	}
	return savedUser, nil
}

func (s *UserService) List(ctx context.Context, params rez.ListUsersParams) ([]*ent.User, error) {
	query := s.db.User.Query().
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
