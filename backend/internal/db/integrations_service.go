package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/jobs"
	"github.com/rs/zerolog/log"
)

type IntegrationsService struct {
	db             *ent.Client
	jobs           rez.JobsService
	syncer         rez.IntegrationsDataSyncer
	oauth2Handlers map[string]rez.OAuth2IntegrationHandler
}

func NewIntegrationsService(db *ent.Client, jobs rez.JobsService, syncer rez.IntegrationsDataSyncer) (*IntegrationsService, error) {
	s := &IntegrationsService{
		db:             db,
		jobs:           jobs,
		syncer:         syncer,
		oauth2Handlers: make(map[string]rez.OAuth2IntegrationHandler),
	}

	return s, nil
}

func (s *IntegrationsService) listQuery(p rez.ListIntegrationsParams) *ent.IntegrationQuery {
	query := s.db.Integration.Query()
	if p.Name != "" {
		query.Where(integration.Name(p.Name))
	}
	return query
}

func (s *IntegrationsService) ListIntegrations(ctx context.Context, params rez.ListIntegrationsParams) ([]*ent.Integration, error) {
	query := s.listQuery(params)
	return query.All(ctx)
}

func (s *IntegrationsService) GetIntegration(ctx context.Context, id uuid.UUID) (*ent.Integration, error) {
	return s.db.Integration.Get(ctx, id)
}

type updateIntegrationMutation interface {
	Save(context.Context) (*ent.Integration, error)
	Mutation() *ent.IntegrationMutation
}

func (s *IntegrationsService) SetIntegration(ctx context.Context, id uuid.UUID, setFn func(*ent.IntegrationMutation)) (*ent.Integration, error) {

	var m updateIntegrationMutation
	if id != uuid.Nil {
		m = s.db.Integration.UpdateOneID(id)
	} else {
		m = s.db.Integration.Create()
	}
	setFn(m.Mutation())

	updated, saveErr := m.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("failed to save: %w", saveErr)
	}

	args := jobs.SyncIntegrationsData{
		IntegrationId: updated.ID,
	}
	if jobErr := s.jobs.Insert(ctx, args, nil); jobErr != nil {
		log.Error().Err(jobErr).Msg("failed to insert sync job")
	}

	return updated, nil
}

func (s *IntegrationsService) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	return s.db.Integration.DeleteOneID(id).Exec(ctx)
}

func (s *IntegrationsService) RegisterOAuth2Handler(name string, h rez.OAuth2IntegrationHandler) {
	s.oauth2Handlers[name] = h
}

func (s *IntegrationsService) makeOAuthState(ctx context.Context, name string) (string, error) {
	// TODO
	return "TODO", nil
}

func (s *IntegrationsService) checkOAuthState(ctx context.Context, name string, state string) error {
	// TODO
	// clear after checking
	return nil
}

func (s *IntegrationsService) getOAuth2Handler(name string) (rez.OAuth2IntegrationHandler, bool) {
	h, ok := s.oauth2Handlers[name]
	return h, ok
}

func (s *IntegrationsService) StartOAuth2Flow(ctx context.Context, name string) (string, error) {
	h, ok := s.getOAuth2Handler(name)
	if !ok {
		return "", fmt.Errorf("invalid integration '%s'", name)
	}

	cfg := h.OAuth2Config()
	if cfg == nil {
		return "", errors.New("invalid integration configuration")
	}

	state, stateErr := s.makeOAuthState(ctx, name)
	if stateErr != nil {
		return "", fmt.Errorf("failed to make oauth state: %w", stateErr)
	}

	return cfg.AuthCodeURL(state), nil
}

func (s *IntegrationsService) CompleteOAuth2Flow(ctx context.Context, name, state, code string) (*ent.Integration, error) {
	if stateErr := s.checkOAuthState(ctx, name, state); stateErr != nil {
		return nil, fmt.Errorf("invalid state: %w", stateErr)
	}

	h, ok := s.getOAuth2Handler(name)
	if !ok {
		return nil, fmt.Errorf("invalid integration name '%s'", name)
	}

	token, tokenErr := h.OAuth2Config().Exchange(ctx, code)
	if tokenErr != nil {
		return nil, fmt.Errorf("exchange token: %w", tokenErr)
	}

	prov, intgErr := h.GetIntegrationFromToken(token)
	if intgErr != nil {
		return nil, fmt.Errorf("failed to get integration: %w", intgErr)
	}

	setFn := func(m *ent.IntegrationMutation) {
		m.SetName(prov.Name)
		m.SetConfig(prov.Config)
	}
	intg, setErr := s.SetIntegration(ctx, uuid.Nil, setFn)
	if setErr != nil {
		return nil, fmt.Errorf("failed to create integration: %w", setErr)
	}

	return intg, nil
}
