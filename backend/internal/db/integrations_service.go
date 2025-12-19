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
	oauth2Handlers map[integration.IntegrationType]map[string]rez.OAuth2IntegrationHandler
}

const topicIntegrationUpdated = "integration_updated"

func NewIntegrationsService(db *ent.Client, msgs rez.MessageService, syncer rez.IntegrationsDataSyncService) (*IntegrationsService, error) {
	s := &IntegrationsService{
		db:             db,
		msgs:           msgs,
		syncer:         syncer,
		oauth2Handlers: make(map[integration.IntegrationType]map[string]rez.OAuth2IntegrationHandler),
	}

	msgs.AddConsumerHandler("integration_update_datasync", topicIntegrationUpdated, s.handleIntegrationUpdatedMessage)

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

	s.onIntegrationUpdated(intg)

	return intg, nil
}

type integrationUpdatedPayload struct {
	TenantId      int       `json:"tenantId"`
	IntegrationId uuid.UUID `json:"id"`
}

func (s *IntegrationsService) onIntegrationUpdated(intg *ent.Integration) {
	if pubErr := s.publishIntegrationUpdatedMessage(intg); pubErr != nil {
		log.Error().Err(pubErr).Msg("failed to publish integration updated message")
	}
}

func (s *IntegrationsService) publishIntegrationUpdatedMessage(intg *ent.Integration) error {
	payload := integrationUpdatedPayload{
		IntegrationId: intg.ID,
		TenantId:      intg.TenantID,
	}
	payloadBytes, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return fmt.Errorf("failed to marshal integration updated message: %w", jsonErr)
	}
	return s.msgs.Publish(topicIntegrationUpdated, message.NewMessage(uuid.NewString(), payloadBytes))
}

func (s *IntegrationsService) handleIntegrationUpdatedMessage(msg *message.Message) error {
	var payload integrationUpdatedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal integration updated message: %w", err)
	}
	ctx := access.TenantSystemContext(context.Background(), payload.TenantId)
	// TODO
	if syncErr := s.syncer.SyncUserData(ctx); syncErr != nil {
		log.Error().Err(syncErr).Msg("failed to sync user data")
	}
	return nil
}
