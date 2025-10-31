package db

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/privacy"
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

type userCtxKey struct{}

func (s *UserService) CreateUserContext(ctx context.Context, userId uuid.UUID) (context.Context, error) {
	// TODO: revise usage of this
	userLookupCtx := privacy.DecisionContext(ctx, privacy.Allow)
	usr, userErr := s.GetById(userLookupCtx, userId)
	if userErr != nil {
		if ent.IsNotFound(userErr) {
			return nil, rez.ErrInvalidUser
		}
		return nil, fmt.Errorf("get user by id: %w", userErr)
	}
	return context.WithValue(access.TenantUserContext(ctx, usr.TenantID), userCtxKey{}, usr), nil
}

func (s *UserService) GetUserContext(ctx context.Context) *ent.User {
	return ctx.Value(userCtxKey{}).(*ent.User)
}

func nilEmptyString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *UserService) getUserByProviderID(ctx context.Context, providerID string) (*ent.User, error) {
	userQuery := s.db.User.Query().
		Where(user.ProviderID(providerID))
	return userQuery.Only(ctx)
}

func (s *UserService) FindOrCreateAuthProviderUser(ctx context.Context, pu ent.User) (*ent.User, error) {
	usr, usrErr := s.getUserByProviderID(ctx, pu.ProviderID)
	if usrErr != nil && !ent.IsNotFound(usrErr) {
		return nil, fmt.Errorf("failed to query user: %w", usrErr)
	}
	if usr != nil {
		return usr, nil
	} else if !rez.Config.AllowUserCreation() {
		return nil, rez.ErrInvalidUser
	}

	// tenant exists, user does not exist

	createUser := s.db.User.Create().
		SetProviderID(pu.ProviderID).
		SetEmail(pu.Email).
		SetConfirmed(pu.Confirmed).
		SetNillableName(nilEmptyString(pu.Name)).
		SetNillableTimezone(nilEmptyString(pu.Timezone))

	created, createErr := createUser.Save(ctx)
	if createErr != nil {
		return nil, fmt.Errorf("create user: %w", createErr)
	}
	return created, nil
}

func (s *UserService) Create(ctx context.Context, user ent.User) (*ent.User, error) {
	created, createErr := s.db.User.Create().
		SetEmail(user.Email).
		SetName(user.Name).
		Save(ctx)
	if createErr != nil {
		return nil, fmt.Errorf("failed to create user: %w", createErr)
	}
	return created, nil
}

func (s *UserService) GetById(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	return s.db.User.Get(ctx, id)
}

func (s *UserService) getOneWhere(ctx context.Context, p predicate.User) (*ent.User, error) {
	return s.db.User.Query().Where(p).Only(ctx)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	return s.getOneWhere(ctx, user.Email(email))
}

func (s *UserService) GetByChatId(ctx context.Context, chatId string) (*ent.User, error) {
	return s.getOneWhere(ctx, user.ChatID(chatId))
}

func (s *UserService) GetTenantById(ctx context.Context, id int) (*ent.Tenant, error) {
	return s.db.Tenant.Get(ctx, id)
}

func (s *UserService) LookupProviderUser(ctx context.Context, provUser *ent.User) (*ent.User, error) {
	// TODO: use provider mapping to match user details, not just by email
	email := provUser.Email
	if rez.Config.DebugMode() && os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL") != "" {
		email = os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL")
		//log.Debug().Str("email", email).Msg("using debug auth email")
	}

	allowQueryCtx := privacy.DecisionContext(ctx, privacy.Allow)
	u, lookupErr := s.GetByEmail(allowQueryCtx, email)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("users.GetByEmail: %w", lookupErr)
	}
	return u, nil
}

func (s *UserService) ListUsers(ctx context.Context, params rez.ListUsersParams) ([]*ent.User, error) {
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
