package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/internal/db/datasync"
	"github.com/rezible/rezible/jobs"
)

type IntegrationsService struct {
	db     *ent.Client
	jobs   rez.JobsService
	syncer *datasync.Syncer
}

func NewIntegrationsService(db *ent.Client, jobSvc rez.JobsService) (*IntegrationsService, error) {
	syncer := datasync.NewSyncerService(db)
	jobs.RegisterPeriodicJob(jobs.SyncAllTenantIntegrationsDataPeriodicJob)
	jobs.RegisterWorkerFunc(syncer.SyncIntegrationsData)

	s := &IntegrationsService{
		db:     db,
		jobs:   jobSvc,
		syncer: syncer,
	}

	return s, nil
}

func (s *IntegrationsService) listQuery(p rez.ListIntegrationsParams) *ent.IntegrationQuery {
	query := s.db.Integration.Query()
	if p.Name != "" {
		query.Where(integration.Name(p.Name))
	}
	if p.DataKind != "" {
		hasDataKindPred := sqljson.ValueEQ(integration.FieldDataKinds, true, sqljson.DotPath(p.DataKind))
		query.Where(func(s *sql.Selector) {
			s.Where(hasDataKindPred)
		})
	}
	if p.ConfigValues != nil && len(p.ConfigValues) > 0 {
		for path, value := range p.ConfigValues {
			query.Where(func(s *sql.Selector) {
				s.Where(sqljson.ValueEQ(integration.FieldConfig, value, sqljson.DotPath(path)))
			})
		}
	}
	if p.Filter != nil {
		p.Filter(query)
	}
	return query
}

func (s *IntegrationsService) ListIntegrations(ctx context.Context, params rez.ListIntegrationsParams) ([]*ent.Integration, error) {
	return s.listQuery(params).All(ctx)
}

func (s *IntegrationsService) GetIntegration(ctx context.Context, name string) (*ent.Integration, error) {
	q := s.db.Integration.Query().Where(integration.Name(name))
	return q.Only(ctx)
}

type configureIntegrationMutation interface {
	Save(context.Context) (*ent.Integration, error)
	Mutation() *ent.IntegrationMutation
}

func (s *IntegrationsService) ConfigureIntegration(ctx context.Context, name string, user bool, rawCfg json.RawMessage, dataKinds map[string]bool) (*ent.Integration, error) {
	curr, getCurrErr := s.GetIntegration(ctx, name)
	if getCurrErr != nil && !ent.IsNotFound(getCurrErr) {
		return nil, fmt.Errorf("failed to get integration %s: %w", name, getCurrErr)
	}

	p, pErr := integrations.GetPackage(name)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get package %s: %w", name, pErr)
	}

	var m configureIntegrationMutation
	if curr == nil {
		m = s.db.Integration.Create().SetName(name)
	} else {
		m = s.db.Integration.UpdateOneID(curr.ID)
	}

	newCfg := rawCfg
	if user {
		currCfg := []byte("{}")
		if curr != nil {
			currCfg = curr.Config
		}
		merged, mergeErr := p.MergeUserConfig(currCfg, rawCfg)
		if mergeErr != nil {
			return nil, fmt.Errorf("failed to merge user config: %w", mergeErr)
		}
		newCfg = merged
	}

	valid, validErr := p.ValidateConfig(newCfg)
	if !valid || validErr != nil {
		return nil, fmt.Errorf("failed to validate config: %w", validErr)
	}

	mut := m.Mutation()
	mut.SetConfig(newCfg)
	mut.SetDataKinds(dataKinds)

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

func (s *IntegrationsService) getIntegrationOAuthConfig(d rez.IntegrationPackage) (rez.IntegrationWithOAuth2SetupFlow, *oauth2.Config, error) {
	oauth2Intg, ok := d.(rez.IntegrationWithOAuth2SetupFlow)
	if !ok {
		return nil, nil, fmt.Errorf("oauth2 flow not supported for integration %s", d.Name())
	}

	cfg := oauth2Intg.OAuth2Config()
	if cfg == nil {
		return nil, nil, errors.New("invalid integration configuration")
	}
	return oauth2Intg, cfg, nil
}

func (s *IntegrationsService) StartOAuth2Flow(ctx context.Context, name string) (string, error) {
	intgDetail, intgErr := integrations.GetPackage(name)
	if intgErr != nil {
		return "", intgErr
	}

	_, cfg, cfgErr := s.getIntegrationOAuthConfig(intgDetail)
	if cfgErr != nil {
		return "", errors.New("invalid integration configuration")
	}

	state, stateErr := s.makeOAuthState(ctx, name)
	if stateErr != nil {
		return "", fmt.Errorf("failed to make oauth state: %w", stateErr)
	}

	return cfg.AuthCodeURL(state), nil
}

func (s *IntegrationsService) CompleteOAuth2Flow(ctx context.Context, name, state, code string) (*ent.Integration, error) {
	intgDetail, intgErr := integrations.GetPackage(name)
	if intgErr != nil {
		return nil, intgErr
	}

	oauth2Intg, oauthCfg, oauthCfgErr := s.getIntegrationOAuthConfig(intgDetail)
	if oauthCfgErr != nil {
		return nil, errors.New("invalid integration configuration")
	}

	if stateErr := s.verifyOAuthState(ctx, name, state); stateErr != nil {
		return nil, fmt.Errorf("invalid state: %w", stateErr)
	}

	token, tokenErr := oauthCfg.Exchange(ctx, code)
	if tokenErr != nil {
		return nil, fmt.Errorf("exchange token: %w", tokenErr)
	}

	cfg, cfgErr := oauth2Intg.GetIntegrationConfigFromToken(token)
	if cfgErr != nil {
		return nil, fmt.Errorf("failed to get integration config: %w", cfgErr)
	}

	cfgJson, jsonErr := json.Marshal(cfg)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal integration config: %w", jsonErr)
	}

	enabledKinds := map[string]bool{}
	kinds := intgDetail.SupportedDataKinds()
	if len(kinds) == 1 {
		enabledKinds[kinds[0]] = true
	} else {
		for _, kind := range kinds {
			enabledKinds[kind] = false
		}
	}

	intg, setErr := s.ConfigureIntegration(ctx, name, false, cfgJson, enabledKinds)
	if setErr != nil {
		return nil, fmt.Errorf("failed to create integration: %w", setErr)
	}

	return intg, nil
}

func (s *IntegrationsService) getPackageSupportingDataKind(ctx context.Context, dataKind string) (rez.IntegrationPackage, error) {
	intgs, listErr := s.ListIntegrations(ctx, rez.ListIntegrationsParams{
		DataKind: dataKind,
	})
	if listErr != nil {
		return nil, listErr
	}
	if len(intgs) != 1 {
		return nil, fmt.Errorf("expected 1 integration, got %d", len(intgs))
	}
	return integrations.GetPackage(intgs[0].Name)
}

func (s *IntegrationsService) GetChatService(ctx context.Context) (rez.ChatService, error) {
	p, pErr := s.getPackageSupportingDataKind(ctx, "chat")
	if pErr != nil {
		return nil, pErr
	}
	if chatPackage, ok := p.(rez.IntegrationWithChatService); ok {
		return chatPackage.GetChatService(), nil
	}
	return nil, rez.ErrNoConfiguredIntegrations
}

func (s *IntegrationsService) GetVideoConferenceService(ctx context.Context) (rez.VideoConferenceService, error) {
	p, pErr := s.getPackageSupportingDataKind(ctx, "video_conference")
	if pErr != nil {
		return nil, pErr
	}
	if vcPackage, ok := p.(rez.IntegrationWithVideoConferenceService); ok {
		return vcPackage.GetVideoConferenceService(), nil
	}
	return nil, rez.ErrNoConfiguredIntegrations
}
