package db

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/google/uuid"
	"github.com/rezible/rezible/access"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"

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
	if len(p.Names) > 0 {
		if len(p.Names) == 1 {
			query.Where(integration.Name(p.Names[0]))
		} else {
			query.Where(integration.NameIn(p.Names...))
		}
	}
	//if len(p.DataKinds) > 0 {
	//	hasDataKindPred := sqljson.ValueEQ(integration.FieldDataKinds, true, sqljson.DotPath(p.DataKind))
	//	query.Where(func(s *sql.Selector) {
	//		s.Where(hasDataKindPred)
	//	})
	//}
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

func (s *IntegrationsService) asConfigured(i *ent.Integration) (rez.ConfiguredIntegration, error) {
	p, pErr := integrations.GetPackage(i.Name)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get integration package: %w", pErr)
	}
	return p.GetConfiguredIntegration(i), nil
}

func (s *IntegrationsService) listIntegrations(ctx context.Context, params rez.ListIntegrationsParams) ([]*ent.Integration, error) {
	q := s.listQuery(params)
	intgs, listErr := q.All(ctx)
	if listErr != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", listErr)
	}
	return intgs, nil
}

var integrationLookupTenantIdCache = make(map[string]uuid.UUID)

var lookupTenantGroup singleflight.Group

func (s *IntegrationsService) LookupByConfigValues(ctx context.Context, name string, configValues map[string]any) (*ent.Integration, error) {
	valsJson, jsonErr := json.Marshal(configValues)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal cfg values: %w", jsonErr)
	}
	lookupKey := fmt.Sprintf("intg_tenant_%s_%s", name, string(valsJson))

	// TODO: use an actual cache
	if id, exists := integrationLookupTenantIdCache[lookupKey]; exists {
		return s.getById(ctx, id)
	}

	lookupFn := func() (any, error) {
		listParams := rez.ListIntegrationsParams{
			Names:        []string{name},
			ConfigValues: configValues,
		}
		intgs, intgsErr := s.listIntegrations(access.SystemContext(ctx), listParams)
		if intgsErr != nil {
			return nil, fmt.Errorf("failed to list integrations: %w", intgsErr)
		}
		if len(intgs) != 1 {
			return nil, fmt.Errorf("found unexpected number of matching integrations: %d", len(intgs))
		}
		return intgs[0].ID, nil
	}
	v, lookupErr, _ := lookupTenantGroup.Do(lookupKey, lookupFn)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if id, ok := v.(uuid.UUID); ok {
		return s.getById(ctx, id)
	}
	return nil, fmt.Errorf("invalid tenant id from lookup: %v", v)
}

func (s *IntegrationsService) ListConfigured(ctx context.Context, params rez.ListIntegrationsParams) ([]rez.ConfiguredIntegration, error) {
	intgs, listErr := s.listIntegrations(ctx, params)
	if listErr != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", listErr)
	}
	cfgIs := make([]rez.ConfiguredIntegration, len(intgs))
	for i, intg := range intgs {
		ci, ciErr := s.asConfigured(intg)
		if ciErr != nil {
			return nil, fmt.Errorf("failed to list integrations: %w", ciErr)
		}
		cfgIs[i] = ci
	}
	return cfgIs, nil
}

func (s *IntegrationsService) getById(ctx context.Context, id uuid.UUID) (*ent.Integration, error) {
	return s.db.Integration.Get(ctx, id)
}

func (s *IntegrationsService) Get(ctx context.Context, name string) (*ent.Integration, error) {
	q := s.db.Integration.Query().Where(integration.Name(name))
	intg, getErr := q.Only(ctx)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get integration: %w", getErr)
	}
	return intg, nil
}

func (s *IntegrationsService) GetConfigured(ctx context.Context, name string) (rez.ConfiguredIntegration, error) {
	intg, getErr := s.Get(ctx, name)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get integration: %w", getErr)
	}
	return s.asConfigured(intg)
}

func (s *IntegrationsService) SetIntegration(ctx context.Context, name string, setFn func(*ent.IntegrationMutation)) (rez.ConfiguredIntegration, error) {
	p, pErr := integrations.GetPackage(name)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get package for integration %s: %w", name, pErr)
	}
	intg, setErr := s.setIntegration(ctx, name, setFn)
	if setErr != nil {
		return nil, fmt.Errorf("failed to set integration: %w", setErr)
	}
	return p.GetConfiguredIntegration(intg), nil
}

func (s *IntegrationsService) setIntegration(ctx context.Context, name string, setFn func(*ent.IntegrationMutation)) (*ent.Integration, error) {
	curr, getCurrErr := s.Get(ctx, name)
	if getCurrErr != nil && !ent.IsNotFound(getCurrErr) {
		return nil, fmt.Errorf("failed to get integration %s: %w", name, getCurrErr)
	}

	var upsert interface {
		Save(context.Context) (*ent.Integration, error)
		Mutation() *ent.IntegrationMutation
	}
	if curr == nil {
		upsert = s.db.Integration.Create().SetName(name)
	} else {
		upsert = s.db.Integration.UpdateOneID(curr.ID)
	}

	setFn(upsert.Mutation())

	intg, saveErr := upsert.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("failed to save: %w", saveErr)
	}

	args := jobs.SyncIntegrationsData{
		IntegrationId: intg.ID,
	}
	if jobErr := s.jobs.Insert(ctx, args, nil); jobErr != nil {
		log.Error().Err(jobErr).Msg("failed to insert sync job")
	}

	return intg, nil
}

func (s *IntegrationsService) DeleteConfigured(ctx context.Context, name string) error {
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

func (s *IntegrationsService) StartOAuth2Flow(ctx context.Context, name string) (string, error) {
	oi, oiErr := integrations.GetOAuthIntegration(name)
	if oiErr != nil {
		return "", fmt.Errorf("invalid oauth2 integration: %w", oiErr)
	}
	state, stateErr := s.makeOAuthState(ctx, name)
	if stateErr != nil {
		return "", fmt.Errorf("failed to make oauth state: %w", stateErr)
	}
	return oi.OAuth2Config().AuthCodeURL(state), nil
}

func (s *IntegrationsService) CompleteOAuth2Flow(ctx context.Context, name, state, code string) (rez.ConfiguredIntegration, error) {
	oi, oiErr := integrations.GetOAuthIntegration(name)
	if oiErr != nil {
		return nil, fmt.Errorf("invalid oauth2 integration: %w", oiErr)
	}
	if stateErr := s.verifyOAuthState(ctx, name, state); stateErr != nil {
		return nil, fmt.Errorf("invalid state: %w", stateErr)
	}
	token, tokenErr := oi.OAuth2Config().Exchange(ctx, code)
	if tokenErr != nil {
		return nil, fmt.Errorf("exchange token: %w", tokenErr)
	}
	cfg, cfgErr := oi.ExtractIntegrationConfigFromToken(token)
	if cfgErr != nil {
		return nil, fmt.Errorf("extract integration config: %w", cfgErr)
	}
	intg, setErr := s.SetIntegration(ctx, name, func(m *ent.IntegrationMutation) {
		m.SetConfig(cfg)
	})
	if setErr != nil {
		return nil, fmt.Errorf("set integration: %w", setErr)
	}
	return intg, nil
}

func (s *IntegrationsService) getConfiguredIntegrationForDataKind(ctx context.Context, dataKind string) (rez.ConfiguredIntegration, error) {
	var names []string
	for _, p := range integrations.GetAvailable() {
		if slices.Contains(p.SupportedDataKinds(), dataKind) {
			names = append(names, p.Name())
		}
	}
	intgs, listErr := s.listIntegrations(ctx, rez.ListIntegrationsParams{Names: names})
	if listErr != nil {
		return nil, listErr
	}
	for _, intg := range intgs {
		p, pErr := integrations.GetPackage(intg.Name)
		if pErr != nil {
			return nil, fmt.Errorf("get package %s: %w", intg.Name, pErr)
		}
		ci := p.GetConfiguredIntegration(intg)
		if slices.Contains(ci.EnabledDataKinds(), dataKind) {
			// TODO: return multiple?
			return ci, nil
		}
	}
	return nil, rez.ErrNoConfiguredIntegrations
}

func (s *IntegrationsService) GetChatIntegration(ctx context.Context) (rez.ChatService, error) {
	p, pErr := s.getConfiguredIntegrationForDataKind(ctx, "chat")
	if pErr != nil {
		return nil, pErr
	}
	if chatPackage, ok := p.(rez.IntegrationWithChatService); ok {
		return chatPackage.ChatService(ctx)
	}
	return nil, rez.ErrNoConfiguredIntegrations
}

func (s *IntegrationsService) GetVideoConferenceIntegration(ctx context.Context) (rez.VideoConferenceIntegration, error) {
	p, pErr := s.getConfiguredIntegrationForDataKind(ctx, "video_conference")
	if pErr != nil {
		return nil, pErr
	}
	if vcPackage, ok := p.(rez.IntegrationWithVideoConference); ok {
		return vcPackage.VideoConferenceIntegration(ctx)
	}
	return nil, rez.ErrNoConfiguredIntegrations
}
