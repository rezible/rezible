package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/jobs"
	"github.com/rs/zerolog/log"
)

type IntegrationsService struct {
	db     *ent.Client
	jobs   rez.JobsService
	syncer rez.IntegrationsDataSyncer
}

func NewIntegrationsService(db *ent.Client, jobs rez.JobsService, syncer rez.IntegrationsDataSyncer) (*IntegrationsService, error) {
	s := &IntegrationsService{
		db:     db,
		jobs:   jobs,
		syncer: syncer,
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

func (s *IntegrationsService) GetIntegration(ctx context.Context, name string) (*ent.Integration, error) {
	q := s.db.Integration.Query().Where(integration.Name(name))
	return q.Only(ctx)
}

type updateIntegrationMutation interface {
	Save(context.Context) (*ent.Integration, error)
	Mutation() *ent.IntegrationMutation
}

func (s *IntegrationsService) ConfigureIntegration(ctx context.Context, name string, cfg json.RawMessage) (*ent.Integration, error) {
	curr, getCurrErr := s.GetIntegration(ctx, name)
	if getCurrErr != nil && !ent.IsNotFound(getCurrErr) {
		return nil, fmt.Errorf("failed to get integration %s: %w", name, getCurrErr)
	}
	var m updateIntegrationMutation
	if curr != nil {
		m = s.db.Integration.UpdateOneID(curr.ID)
	} else {
		m = s.db.Integration.Create()
	}

	m.Mutation().SetConfig(cfg)

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

func (s *IntegrationsService) DeleteIntegration(ctx context.Context, name string) error {
	q := s.db.Integration.Delete().Where(integration.Name(name))
	_, deleteErr := q.Exec(ctx)
	return deleteErr
}

func (s *IntegrationsService) makeOAuthState(ctx context.Context, name string) (string, error) {
	// TODO
	return "TODO", nil
}

func (s *IntegrationsService) verifyOAuthState(ctx context.Context, name string, state string) error {
	// TODO
	// clear after checking
	return nil
}

func (s *IntegrationsService) getOAuth2FlowHandler(name string) (rez.IntegrationWithOAuth2SetupFlow, error) {
	intgDetail, intgErr := integrations.GetDetail(name)
	if intgErr != nil {
		return nil, intgErr
	}

	oauth2Intg, ok := intgDetail.(rez.IntegrationWithOAuth2SetupFlow)
	if !ok {
		return nil, fmt.Errorf("oauth2 flow not supported for integration %s", name)
	}

	return oauth2Intg, nil
}

func (s *IntegrationsService) StartOAuth2Flow(ctx context.Context, name string) (string, error) {
	h, hErr := s.getOAuth2FlowHandler(name)
	if hErr != nil {
		return "", hErr
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
	h, hErr := s.getOAuth2FlowHandler(name)
	if hErr != nil {
		return nil, hErr
	}

	if stateErr := s.verifyOAuthState(ctx, name, state); stateErr != nil {
		return nil, fmt.Errorf("invalid state: %w", stateErr)
	}

	token, tokenErr := h.OAuth2Config().Exchange(ctx, code)
	if tokenErr != nil {
		return nil, fmt.Errorf("exchange token: %w", tokenErr)
	}

	cfg, cfgErr := h.GetIntegrationConfigFromToken(token)
	if cfgErr != nil {
		return nil, fmt.Errorf("failed to get integration config: %w", cfgErr)
	}

	cfgJson, jsonErr := json.Marshal(cfg)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal integration config: %w", jsonErr)
	}

	intg, setErr := s.ConfigureIntegration(ctx, name, cfgJson)
	if setErr != nil {
		return nil, fmt.Errorf("failed to create integration: %w", setErr)
	}

	return intg, nil
}
