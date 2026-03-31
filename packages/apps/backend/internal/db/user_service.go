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

func (s *UserService) FindOrCreateAuthProviderUser(ctx context.Context, pu ent.User) (*ent.User, error) {
	usr, usrErr := s.Get(ctx, user.AuthProviderID(pu.AuthProviderID))
	if usrErr != nil && !ent.IsNotFound(usrErr) {
		return nil, fmt.Errorf("failed to query user: %w", usrErr)
	}
	if usr != nil {
		return usr, nil
	}

	// tenant exists, user does not exist
	return s.Set(ctx, uuid.Nil, func(m *ent.UserMutation) {
		m.SetAuthProviderID(pu.AuthProviderID)
		m.SetEmail(pu.Email)
		m.SetConfirmed(pu.Confirmed)
		m.SetName(pu.Name)
		m.SetTimezone(pu.Timezone)
	})
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
			return fmt.Errorf("mutate user: %w", saveErr)
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
