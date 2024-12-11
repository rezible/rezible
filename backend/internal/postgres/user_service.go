package postgres

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	rez "github.com/twohundreds/rezible"
	"github.com/twohundreds/rezible/ent"
	"github.com/twohundreds/rezible/ent/predicate"
	"github.com/twohundreds/rezible/ent/team"
	"github.com/twohundreds/rezible/ent/user"
)

type UserService struct {
	db     *ent.Client
	loader rez.ProviderLoader
}

func NewUserService(db *ent.Client, pl rez.ProviderLoader) (*UserService, error) {
	s := &UserService{
		db:     db,
		loader: pl,
	}

	return s, nil
}

func (s *UserService) SyncData(ctx context.Context) error {
	prov, provErr := s.loader.LoadUserDataProvider(ctx)
	if provErr != nil {
		return fmt.Errorf("loading user data provider: %w", provErr)
	}
	syncer := newUserDataSyncer(s.db, prov)
	return syncer.syncProviderData(ctx)
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
