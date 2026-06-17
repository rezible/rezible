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

func (s *UserService) SyncFromAuthProvider(ctx context.Context, pu ent.User) (*ent.User, error) {
	var usr *ent.User
	return usr, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var userId uuid.UUID
		existing, getErr := s.Get(ctx, user.AuthProviderID(pu.AuthProviderID))
		if getErr != nil && !ent.IsNotFound(getErr) {
			return fmt.Errorf("query existing: %w", getErr)
		} else if existing != nil {
			if existing.Email == pu.Email && existing.Name == pu.Name {
				usr = existing.Unwrap()
				return nil
			}
			userId = existing.ID
		}

		syncAuthDetailsFn := func(m *ent.UserMutation) {
			m.SetAuthProviderID(pu.AuthProviderID)
			m.SetEmail(pu.Email)
			m.SetName(pu.Name)
			m.SetTimezone(pu.Timezone)
		}
		saved, setErr := s.Set(ctx, userId, syncAuthDetailsFn)
		if setErr != nil {
			return fmt.Errorf("set user: %w", setErr)
		}

		//if existing == nil {
		//	createAdminRole := tx.OrganizationRole.Create().
		//		SetRole(organizationrole.RoleAdmin).
		//		SetUserID(saved.ID).
		//		SetOrganizationID(org.ID)
		//	if createRoleErr := createAdminRole.Exec(ctx); createRoleErr != nil {
		//		return fmt.Errorf("create admin role: %w", createRoleErr)
		//	}
		//}

		usr = saved.Unwrap()

		return nil
	})
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
