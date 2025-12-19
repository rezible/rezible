package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rezible/rezible/access"
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
)

type IntegrationsService struct {
	db             *ent.Client
	msgs           rez.MessageService
	syncer         rez.IntegrationsDataSyncService
	oauth2Handlers map[string]rez.OAuth2IntegrationHandler
}

const topicIntegrationUpdated = "integration_updated"

func NewIntegrationsService(db *ent.Client, msgs rez.MessageService, syncer rez.IntegrationsDataSyncService) (*IntegrationsService, error) {
	s := &IntegrationsService{
		db:             db,
		msgs:           msgs,
		syncer:         syncer,
		oauth2Handlers: make(map[string]rez.OAuth2IntegrationHandler),
	}

	msgs.AddConsumerHandler("integration_update_datasync", topicIntegrationUpdated, s.handleIntegrationUpdatedMessage)

	return s, nil
}

func (s *IntegrationsService) listQuery(p rez.ListIntegrationsParams) *ent.IntegrationQuery {
	query := s.db.Integration.Query()
	if p.Enabled {
		query.Where(integration.EnabledEQ(true))
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

func (s *IntegrationsService) CompleteOAuth2Flow(ctx context.Context, params rez.CompleteIntegrationOAuth2FlowParams) (*ent.Integration, error) {
	if stateErr := s.checkOAuthState(ctx, params.Name, params.State); stateErr != nil {
		return nil, fmt.Errorf("invalid state: %w", stateErr)
	}

	h, ok := s.getOAuth2Handler(params.Name)
	if !ok {
		return nil, fmt.Errorf("invalid integration name '%s'", params.Name)
	}

	prov, completeErr := h.CompleteOAuth2Flow(ctx, params.Code)
	if completeErr != nil {
		return nil, fmt.Errorf("failed to complete flow: %w", completeErr)
	}

	setFn := func(m *ent.IntegrationMutation) {
		m.SetName(prov.Name)
		m.SetConfig(prov.Config)
		m.SetEnabled(prov.Enabled)
	}
	intg, pcErr := s.SetIntegration(ctx, uuid.Nil, setFn)
	if pcErr != nil {
		return nil, fmt.Errorf("failed to create integration: %w", pcErr)
	}

	s.onIntegrationUpdated(intg)

	return intg, nil
}

func (s *IntegrationsService) onIntegrationUpdated(intg *ent.Integration) {
	if pubErr := s.publishIntegrationUpdatedMessage(intg); pubErr != nil {
		log.Error().Err(pubErr).Msg("failed to publish integration updated message")
	}
}

func (s *IntegrationsService) publishIntegrationUpdatedMessage(intg *ent.Integration) error {
	payloadBytes, jsonErr := json.Marshal(intg)
	if jsonErr != nil {
		return fmt.Errorf("failed to marshal integration updated message: %w", jsonErr)
	}
	return s.msgs.Publish(topicIntegrationUpdated, message.NewMessage(uuid.NewString(), payloadBytes))
}

func (s *IntegrationsService) handleIntegrationUpdatedMessage(msg *message.Message) error {
	var intg *ent.Integration
	if err := json.Unmarshal(msg.Payload, &intg); err != nil {
		return fmt.Errorf("failed to unmarshal integration updated message: %w", err)
	}
	ctx := access.TenantSystemContext(context.Background(), intg.TenantID)
	if syncErr := s.syncer.SyncIntegrationsData(ctx, ent.Integrations{intg}); syncErr != nil {
		log.Error().Err(syncErr).Msg("failed to sync integrations data")
	}
	return nil
}
