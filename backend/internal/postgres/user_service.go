package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/privacy"
	"github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/ent/tenant"
	"github.com/rezible/rezible/ent/user"
)

type UserService struct {
	db *ent.Client
}

func NewUserService(db *ent.Client) (*UserService, error) {
	s := &UserService{
		db: db,
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
			return nil, rez.ErrAuthSessionUserMissing
		}
		return nil, fmt.Errorf("get user by id: %w", userErr)
	}
	return context.WithValue(access.TenantUserContext(ctx, usr.TenantID), userCtxKey{}, usr), nil
}

func (s *UserService) GetUserContext(ctx context.Context) *ent.User {
	return ctx.Value(userCtxKey{}).(*ent.User)
}

func getEmailDomain(email string) (string, error) {
	emailParts := strings.SplitN(email, "@", 2)
	if len(emailParts) != 2 {
		return "", fmt.Errorf("invalid user email domain")
	}
	return emailParts[1], nil
}

func nilEmptyString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *UserService) getUserByAuthProviderID(ctx context.Context, authProviderID string) (*ent.User, error) {
	userQuery := s.db.User.Query().Where(user.AuthProviderID(authProviderID))
	return userQuery.Only(ctx)
}

func (s *UserService) FindOrCreateAuthProviderUser(ctx context.Context, provUser *ent.User, tenantAuthId string) (*ent.User, error) {
	tnt, tenantErr := s.db.Tenant.Query().Where(tenant.AuthID(tenantAuthId)).Only(ctx)
	if tenantErr != nil {
		if !ent.IsNotFound(tenantErr) {
			return nil, fmt.Errorf("failed to query tenant: %w", tenantErr)
		}
		// TODO: allow returning error here (tenant does not exist, don't create new tenant)
	}

	if tnt != nil {
		ctx = access.TenantSystemContext(ctx, tnt.ID)
		usr, usrErr := s.getUserByAuthProviderID(ctx, provUser.AuthProviderID)
		if usrErr != nil {
			if !ent.IsNotFound(usrErr) {
				return nil, fmt.Errorf("failed to query user: %w", usrErr)
			}
			// TODO: allow returning error here (tenant exists, don't create new user)
		} else if usr != nil {
			return usr, nil
		}
	}

	// tenant and user do not exist

	ctx = access.SystemContext(ctx)

	var createdUser *ent.User

	createUserTenantFn := func(tx *ent.Tx) error {
		if tnt == nil {
			createTenant := tx.Tenant.Create().
				SetAuthID(tenantAuthId).
				SetName(tenantAuthId)

			tnt, tenantErr = createTenant.Save(ctx)
			if tenantErr != nil {
				return fmt.Errorf("create tenant: %w", tenantErr)
			}
			ctx = access.TenantSystemContext(ctx, tnt.ID)
		}

		createUser := tx.User.Create().
			SetTenantID(tnt.ID).
			SetAuthProviderID(provUser.AuthProviderID).
			SetEmail(provUser.Email).
			SetConfirmed(provUser.Confirmed).
			SetNillableName(nilEmptyString(provUser.Name)).
			SetNillableTimezone(nilEmptyString(provUser.Timezone))

		created, createErr := createUser.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create user: %w", createErr)
		}
		createdUser = created

		return nil
	}

	if txErr := ent.WithTx(ctx, s.db, createUserTenantFn); txErr != nil {
		return nil, txErr
	}

	return createdUser, nil
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

func (s *UserService) LookupProviderUser(ctx context.Context, provUser *ent.User) (*ent.User, error) {
	// TODO: use provider mapping to match user details, not just by email
	email := provUser.Email
	if rez.DebugMode && os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL") != "" {
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
