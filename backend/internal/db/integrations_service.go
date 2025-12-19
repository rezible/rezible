package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/internal/db/datasync"
)

type IntegrationsService struct {
	db             *ent.Client
	oauth2Handlers map[integration.IntegrationType]map[string]rez.OAuth2IntegrationHandler
}

func NewIntegrationsService(db *ent.Client) (*IntegrationsService, error) {
	s := &IntegrationsService{
		db:             db,
		oauth2Handlers: make(map[integration.IntegrationType]map[string]rez.OAuth2IntegrationHandler),
	}

	return s, nil
}

func (s *IntegrationsService) listQuery(p rez.ListIntegrationsParams) *ent.IntegrationQuery {
	query := s.db.Integration.Query()
	if p.Enabled {
		query.Where(integration.EnabledEQ(true))
	}
	if p.Type != "" {
		query.Where(integration.IntegrationTypeEQ(p.Type))
	}
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
	return m.Save(ctx)
}

func (s *IntegrationsService) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	return s.db.Integration.DeleteOneID(id).Exec(ctx)
}

func (s *IntegrationsService) RegisterOAuth2Handler(t integration.IntegrationType, name string, h rez.OAuth2IntegrationHandler) {
	_, ok := s.oauth2Handlers[t]
	if !ok {
		s.oauth2Handlers[t] = make(map[string]rez.OAuth2IntegrationHandler)
	}
	s.oauth2Handlers[t][name] = h
}

func (s *IntegrationsService) makeOAuthState(ctx context.Context, t integration.IntegrationType, name string) (string, error) {
	// TODO
	return "TODO", nil
}

func (s *IntegrationsService) checkOAuthState(ctx context.Context, t integration.IntegrationType, name string, state string) error {
	// TODO
	// clear after checking
	return nil
}

func (s *IntegrationsService) getOAuth2Handler(t integration.IntegrationType, name string) (rez.OAuth2IntegrationHandler, bool) {
	if byName, typeOk := s.oauth2Handlers[t]; typeOk {
		h, nameOk := byName[name]
		return h, nameOk
	}
	return nil, false
}

func (s *IntegrationsService) StartOAuth2Flow(ctx context.Context, t integration.IntegrationType, name string) (string, error) {
	h, ok := s.getOAuth2Handler(t, name)
	if !ok {
		return "", fmt.Errorf("invalid integration type '%s'", t)
	}

	cfg := h.OAuth2Config()
	if cfg == nil {
		return "", errors.New("invalid integration configuration")
	}

	state, stateErr := s.makeOAuthState(ctx, t, name)
	if stateErr != nil {
		return "", fmt.Errorf("failed to make oauth state: %w", stateErr)
	}

	return cfg.AuthCodeURL(state), nil
}

func (s *IntegrationsService) CompleteOAuth2Flow(ctx context.Context, params rez.CompleteIntegrationOAuth2FlowParams) (*ent.Integration, error) {
	if stateErr := s.checkOAuthState(ctx, params.Type, params.Name, params.State); stateErr != nil {
		return nil, fmt.Errorf("invalid state: %w", stateErr)
	}

	h, ok := s.getOAuth2Handler(params.Type, params.Name)
	if !ok {
		return nil, fmt.Errorf("invalid integration type '%s'", params.Type)
	}

	prov, completeErr := h.CompleteOAuth2Flow(ctx, params.Code)
	if completeErr != nil {
		return nil, fmt.Errorf("failed to complete flow: %w", completeErr)
	}

	setFn := func(m *ent.IntegrationMutation) {
		m.SetIntegrationType(prov.IntegrationType)
		m.SetName(prov.Name)
		m.SetConfig(prov.Config)
		m.SetEnabled(prov.Enabled)
	}
	intg, pcErr := s.SetIntegration(ctx, uuid.Nil, setFn)
	if pcErr != nil {
		return nil, fmt.Errorf("failed to create integration: %w", pcErr)
	}

	return intg, nil
}

func (s *IntegrationsService) MakeDataSyncer(pl rez.DataProviderLoader) rez.IntegrationsDataSyncService {
	return datasync.NewIntegrationsSyncer(s.db, pl)
}
