package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/google/uuid"
	"github.com/rezible/rezible/access"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	ioas "github.com/rezible/rezible/ent/integrationoauthstate"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/internal/db/datasync"
	"github.com/rezible/rezible/jobs"
)

type IntegrationsService struct {
	db     *ent.Client
	jobs   rez.JobsService
	auth   rez.AuthService
	syncer *datasync.Syncer
}

func NewIntegrationsService(db *ent.Client, jobSvc rez.JobsService, auth rez.AuthService) (*IntegrationsService, error) {
	s := &IntegrationsService{
		db:     db,
		jobs:   jobSvc,
		auth:   auth,
		syncer: datasync.NewSyncerService(db),
	}

	s.registerJobs()

	return s, nil
}

func (s *IntegrationsService) registerJobs() {
	syncAllTenantIntegrationsDataPeriodicJob := jobs.NewPeriodicJob(
		jobs.PeriodicInterval(time.Hour),
		func() (jobs.JobArgs, *jobs.InsertOpts) {
			return &jobs.SyncIntegrationsData{}, &jobs.InsertOpts{
				UniqueOpts: jobs.UniqueOpts{
					ByState: jobs.NonCompletedJobStates,
				},
			}
		},
		&jobs.PeriodicJobOpts{RunOnStart: true},
	)
	jobs.RegisterPeriodicJob(syncAllTenantIntegrationsDataPeriodicJob)
	jobs.RegisterWorkerFunc(s.syncer.SyncIntegrationsData)
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

func (s *IntegrationsService) GetConfigured(ctx context.Context, name string) (rez.ConfiguredIntegration, error) {
	intg, getErr := s.Get(ctx, name)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get integration: %w", getErr)
	}
	return s.asConfigured(intg)
}

func (s *IntegrationsService) Configure(ctx context.Context, name string, cfg map[string]any) (rez.ConfiguredIntegration, error) {
	p, pErr := integrations.GetPackage(name)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get package for integration %s: %w", name, pErr)
	}
	if prefsErr := p.ValidateConfig(cfg); prefsErr != nil {
		return nil, fmt.Errorf("invalid config: %w", prefsErr)
	}
	intg, setErr := s.set(ctx, name, func(m *ent.IntegrationMutation) { m.SetConfig(cfg) })
	if setErr != nil {
		return nil, fmt.Errorf("failed to set integration: %w", setErr)
	}
	return p.GetConfiguredIntegration(intg), nil
}

func (s *IntegrationsService) UpdateConfiguredPreferences(ctx context.Context, name string, prefs map[string]any) (rez.ConfiguredIntegration, error) {
	p, pErr := integrations.GetPackage(name)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get package for integration %s: %w", name, pErr)
	}
	if prefsErr := p.ValidateUserPreferences(prefs); prefsErr != nil {
		return nil, fmt.Errorf("invalid user preferences: %w", prefsErr)
	}
	intg, setErr := s.set(ctx, name, func(m *ent.IntegrationMutation) { m.SetUserPreferences(prefs) })
	if setErr != nil {
		return nil, fmt.Errorf("failed to set integration: %w", setErr)
	}
	return p.GetConfiguredIntegration(intg), nil
}

func (s *IntegrationsService) DeleteConfigured(ctx context.Context, name string) error {
	q := s.db.Integration.Delete().Where(integration.Name(name))
	_, deleteErr := q.Exec(ctx)
	return deleteErr
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
	if p.ConfigValues != nil && len(p.ConfigValues) > 0 {
		for path, value := range p.ConfigValues {
			query.Where(func(s *sql.Selector) {
				s.Where(sqljson.ValueEQ(integration.FieldConfig, value, sqljson.DotPath(path)))
			})
		}
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
	if len(params.Names) == 1 && params.ConfigValues != nil {
		// s.lookupByConfigValues(ctx, params.Names[0], params.ConfigValues)
	}
	q := s.listQuery(params)
	intgs, listErr := q.All(ctx)
	if listErr != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", listErr)
	}
	return intgs, nil
}

var integrationLookupTenantIdCache = make(map[string]uuid.UUID)

var lookupTenantGroup singleflight.Group

func (s *IntegrationsService) lookupByConfigValues(ctx context.Context, name string, configValues map[string]any) (*ent.Integration, error) {
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

func (s *IntegrationsService) set(ctx context.Context, name string, setFn func(*ent.IntegrationMutation)) (*ent.Integration, error) {
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

func (s *IntegrationsService) makeOAuthState(ctx context.Context, name string) (string, error) {
	userId := s.auth.GetAuthSession(ctx).UserId()
	state := uuid.New().String()
	create := s.db.IntegrationOAuthState.Create().
		SetUserID(userId).
		SetState(state).
		SetIntegrationName(name)
	return state, create.Exec(ctx)
}

func (s *IntegrationsService) verifyOAuthState(ctx context.Context, name string, state string) error {
	userId := s.auth.GetAuthSession(ctx).UserId()
	userIntegrationStates := ioas.And(ioas.UserIDEQ(userId), ioas.IntegrationNameEQ(name))
	query := s.db.IntegrationOAuthState.Query().
		Where(userIntegrationStates, ioas.ExpiresAtGT(time.Now()), ioas.StateEQ(state))
	stateMatch, queryErr := query.Exist(ctx)
	if queryErr != nil {
		return fmt.Errorf("query failed: %w", queryErr)
	}
	cleanup := s.db.IntegrationOAuthState.Delete().Where(userIntegrationStates, ioas.ExpiresAtLT(time.Now()))
	if _, cleanupErr := cleanup.Exec(ctx); cleanupErr != nil {
		log.Error().Err(cleanupErr).
			Str("name", name).
			Str("userId", userId.String()).
			Msg("failed to cleanup old integration user oauth states")
	}
	if !stateMatch {
		return fmt.Errorf("no match found")
	}
	return nil
}

func (s *IntegrationsService) StartOAuth2Flow(ctx context.Context, name string, redirect *url.URL) (string, error) {
	oi, oiErr := integrations.GetOAuthIntegration(name)
	if oiErr != nil {
		return "", fmt.Errorf("invalid oauth2 integration: %w", oiErr)
	}
	state, stateErr := s.makeOAuthState(ctx, name)
	if stateErr != nil {
		return "", fmt.Errorf("failed to make oauth state: %w", stateErr)
	}
	cfg := oi.OAuth2Config()
	if redirect != nil {
		cfg.RedirectURL = redirect.String()
	}
	return cfg.AuthCodeURL(state), nil
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
	intg, setErr := s.Configure(ctx, name, cfg)
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
		if ci.GetDataKinds()[dataKind] {
			// TODO: return multiple?
			return ci, nil
		}
	}
	return nil, rez.ErrNoConfiguredIntegrations
}

type IntegrationWithChatService interface {
	MakeChatService(context.Context) (rez.ChatService, error)
}

func (s *IntegrationsService) GetChatService(ctx context.Context) (rez.ChatService, error) {
	p, pErr := s.getConfiguredIntegrationForDataKind(ctx, "chat")
	if pErr != nil {
		return nil, pErr
	}
	if chatPackage, ok := p.(IntegrationWithChatService); ok {
		return chatPackage.MakeChatService(ctx)
	}
	return nil, rez.ErrNoConfiguredIntegrations
}

type IntegrationWithVideoConference interface {
	MakeVideoConferenceService(ctx context.Context) (rez.VideoConferenceService, error)
}

func (s *IntegrationsService) GetVideoConferenceService(ctx context.Context) (rez.VideoConferenceService, error) {
	p, pErr := s.getConfiguredIntegrationForDataKind(ctx, "video_conference")
	if pErr != nil {
		return nil, pErr
	}
	if vci, ok := p.(IntegrationWithVideoConference); ok {
		return vci.MakeVideoConferenceService(ctx)
	}
	return nil, rez.ErrNoConfiguredIntegrations
}
