package db

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	pesr "github.com/rezible/rezible/ent/providereventsyncrun"
	"github.com/riverqueue/river"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	ioas "github.com/rezible/rezible/ent/integrationoauthstate"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/jobs"
)

type IntegrationsService struct {
	db   *ent.Client
	jobs rez.JobsService
}

func NewIntegrationsService(db *ent.Client, jobSvc rez.JobsService) (*IntegrationsService, error) {
	s := &IntegrationsService{
		db:   db,
		jobs: jobSvc,
	}

	s.registerJobs()

	return s, nil
}

func (s *IntegrationsService) registerJobs() {
	/*
		syncAllTenantIntegrationsDataPeriodicJob := jobs.NewPeriodicJob(
			jobs.PeriodicInterval(time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return &jobs.SyncIntegrationsData{}, &river.InsertOpts{
					UniqueOpts: river.UniqueOpts{
						ByState: jobs.NonCompletedJobStates,
					},
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		)
		jobs.RegisterPeriodicJob(syncAllTenantIntegrationsDataPeriodicJob)
		jobs.RegisterWorkerFunc(s.syncer.SyncIntegrationsData)
	*/
}

func (s *IntegrationsService) GetAvailable() []rez.IntegrationPackage {
	return integrations.GetAvailable()
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

func (s *IntegrationsService) GetConfigured(ctx context.Context, id uuid.UUID) (rez.ConfiguredIntegration, error) {
	intg, getErr := s.getById(ctx, id)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get integration: %w", getErr)
	}
	return s.asConfigured(intg)
}

func (s *IntegrationsService) Configure(ctx context.Context, params rez.ConfigureIntegrationParams) (rez.ConfiguredIntegration, error) {
	if params.ExternalRef == "" {
		return nil, fmt.Errorf("external_ref is required")
	}
	if params.DisplayName == "" {
		params.DisplayName = params.ExternalRef
	}
	if params.Config == nil {
		params.Config = map[string]any{}
	}
	p, pErr := integrations.GetPackage(params.Provider)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get package for integration %s: %w", params.Provider, pErr)
	}
	if prefsErr := p.ValidateConfig(params.Config); prefsErr != nil {
		return nil, fmt.Errorf("invalid config: %w", prefsErr)
	}
	intg, setErr := s.set(ctx, params, func(m *ent.IntegrationMutation) {
		m.SetConfig(params.Config)
		if params.Preferences != nil {
			m.SetUserPreferences(params.Preferences)
		}
	})
	if setErr != nil {
		return nil, fmt.Errorf("failed to set integration: %w", setErr)
	}
	return p.GetConfiguredIntegration(intg), nil
}

func (s *IntegrationsService) UpdateConfiguredPreferences(ctx context.Context, id uuid.UUID, prefs map[string]any) (rez.ConfiguredIntegration, error) {
	curr, currErr := s.getById(ctx, id)
	if currErr != nil {
		return nil, fmt.Errorf("failed to get integration: %w", currErr)
	}
	p, pErr := integrations.GetPackage(curr.Provider)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get package for integration %s: %w", curr.Provider, pErr)
	}
	if prefsErr := p.ValidateUserPreferences(prefs); prefsErr != nil {
		return nil, fmt.Errorf("invalid user preferences: %w", prefsErr)
	}
	intg, saveErr := s.db.Integration.UpdateOneID(id).SetUserPreferences(prefs).Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("failed to set integration: %w", saveErr)
	}
	return p.GetConfiguredIntegration(intg), nil
}

func (s *IntegrationsService) DeleteConfigured(ctx context.Context, id uuid.UUID) error {
	deleteErr := s.db.Integration.DeleteOneID(id).Exec(ctx)
	return deleteErr
}

func (s *IntegrationsService) GetProviderEventQueriers(ctx context.Context, provider string) ([]rez.ProviderEventQuerier, error) {
	intgs, queryErr := s.db.Integration.Query().Where(integration.ProviderEQ(provider)).All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", queryErr)
	}
	var queriers []rez.ProviderEventQuerier
	for _, intg := range intgs {
		q, qErr := integrations.GetProviderSourceEventQueriers(ctx, intg)
		if qErr != nil {
			return nil, fmt.Errorf("failed to get integration event queriers: %w", qErr)
		}
		queriers = append(queriers, q...)
	}
	return queriers, nil
}

func (s *IntegrationsService) listQuery(p rez.ListIntegrationsParams) *ent.IntegrationQuery {
	query := s.db.Integration.Query()
	if len(p.IDs) > 0 {
		query.Where(integration.IDIn(p.IDs...))
	}
	if len(p.Providers) > 0 {
		if len(p.Providers) == 1 {
			query.Where(integration.Provider(p.Providers[0]))
		} else {
			query.Where(integration.ProviderIn(p.Providers...))
		}
	}
	if len(p.ExternalRefs) > 0 {
		if len(p.ExternalRefs) == 1 {
			query.Where(integration.ExternalRef(p.ExternalRefs[0]))
		} else {
			query.Where(integration.ExternalRefIn(p.ExternalRefs...))
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
	p, pErr := integrations.GetPackage(i.Provider)
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

func (s *IntegrationsService) getById(ctx context.Context, id uuid.UUID) (*ent.Integration, error) {
	return s.db.Integration.Get(ctx, id)
}

func (s *IntegrationsService) getByProviderExternalRef(ctx context.Context, provider, externalRef string) (*ent.Integration, error) {
	q := s.db.Integration.Query().Where(integration.Provider(provider), integration.ExternalRef(externalRef))
	intg, getErr := q.Only(ctx)
	if getErr != nil {
		if ent.IsNotFound(getErr) {
			return nil, getErr
		}
		return nil, fmt.Errorf("failed to get integration: %w", getErr)
	}
	return intg, nil
}

func (s *IntegrationsService) set(ctx context.Context, params rez.ConfigureIntegrationParams, setFn func(*ent.IntegrationMutation)) (*ent.Integration, error) {
	curr, getCurrErr := s.getByProviderExternalRef(ctx, params.Provider, params.ExternalRef)
	if getCurrErr != nil && !ent.IsNotFound(getCurrErr) {
		return nil, fmt.Errorf("failed to get integration %s/%s: %w", params.Provider, params.ExternalRef, getCurrErr)
	}

	var upsert ent.EntityMutator[*ent.Integration, *ent.IntegrationMutation]
	if curr == nil {
		upsert = s.db.Integration.Create().
			SetProvider(params.Provider).
			SetExternalRef(params.ExternalRef)
	} else {
		upsert = s.db.Integration.UpdateOneID(curr.ID)
	}

	upsert.Mutation().SetDisplayName(params.DisplayName)
	setFn(upsert.Mutation())

	intg, saveErr := upsert.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("failed to save: %w", saveErr)
	}

	if s.jobs != nil {
		args := jobs.ProviderEventSyncJob{
			ProviderSources: map[string][]string{intg.Provider: {}},
			SyncReason:      "configured",
		}
		if _, jobErr := s.jobs.Insert(ctx, args, nil); jobErr != nil {
			slog.Error("failed to insert sync job", "error", jobErr)
		}
	}

	return intg, nil
}

func (s *IntegrationsService) makeOAuthState(ctx context.Context, provider string) (string, error) {
	userId, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return "", rez.ErrAuthSessionMissing
	}
	// TODO: replace this with something actually random
	state := uuid.New().String()
	create := s.db.IntegrationOAuthState.Create().
		SetUserID(userId).
		SetState(state).
		SetProvider(provider)
	return state, create.Exec(ctx)
}

func (s *IntegrationsService) getOAuthState(ctx context.Context, provider string, state string) (*ent.IntegrationOAuthState, error) {
	userId, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return nil, rez.ErrAuthSessionMissing
	}
	userIntegrationStates := ioas.And(ioas.UserIDEQ(userId), ioas.ProviderEQ(provider))
	query := s.db.IntegrationOAuthState.Query().
		Where(userIntegrationStates, ioas.ExpiresAtGT(time.Now()), ioas.StateEQ(state))
	stateMatch, queryErr := query.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query failed: %w", queryErr)
	}
	cleanup := s.db.IntegrationOAuthState.Delete().
		Where(userIntegrationStates, ioas.ExpiresAtLT(time.Now()))
	if _, cleanupErr := cleanup.Exec(ctx); cleanupErr != nil {
		slog.Error("failed to cleanup old integration user oauth states",
			"error", cleanupErr,
			"provider", provider,
			"userId", userId.String(),
		)
	}
	return stateMatch, nil
}

func (s *IntegrationsService) StartOAuth2Flow(ctx context.Context, provider string, redirect *url.URL) (string, error) {
	oi, oiErr := integrations.GetOAuthIntegration(provider)
	if oiErr != nil {
		return "", fmt.Errorf("invalid oauth2 integration: %w", oiErr)
	}
	state, stateErr := s.makeOAuthState(ctx, provider)
	if stateErr != nil {
		return "", fmt.Errorf("failed to make oauth state: %w", stateErr)
	}
	cfg := oi.OAuth2Config()
	if redirect != nil {
		cfg.RedirectURL = redirect.String()
	}
	return cfg.AuthCodeURL(state), nil
}

func (s *IntegrationsService) CompleteOAuth2Flow(ctx context.Context, provider string, params rez.CompleteIntegrationOAuth2Params) (*rez.CompleteIntegrationOAuth2Result, error) {
	oi, oiErr := integrations.GetOAuthIntegration(provider)
	if oiErr != nil {
		return nil, fmt.Errorf("invalid oauth2 integration: %w", oiErr)
	}

	var oauthState *ent.IntegrationOAuthState
	var opts []oauth2.AuthCodeOption
	if params.State != nil {
		state, stateErr := s.getOAuthState(ctx, provider, *params.State)
		if stateErr != nil {
			return nil, fmt.Errorf("invalid state: %w", stateErr)
		}
		oauthState = state
	} else if params.ClientVerifier != nil {
		opts = append(opts, oauth2.VerifierOption(*params.ClientVerifier))
	} else {
		return nil, fmt.Errorf("invalid oauth2 integration: missing state or client_verifier")
	}

	token, tokenErr := oi.OAuth2Config().Exchange(ctx, params.Code, opts...)
	if tokenErr != nil {
		return nil, fmt.Errorf("exchange token: %w", tokenErr)
	}
	options, optionsErr := oi.ExtractIntegrationOptionsFromToken(token)
	if optionsErr != nil {
		return nil, fmt.Errorf("extract integration options: %w", optionsErr)
	}
	if len(options) == 0 {
		return nil, fmt.Errorf("no integration options returned")
	}
	if len(options) == 1 {
		configured, cfgErr := s.configureOptions(ctx, provider, options)
		if cfgErr != nil {
			return nil, cfgErr
		}
		res := &rez.CompleteIntegrationOAuth2Result{
			Status:     "configured",
			Configured: configured,
		}
		return res, nil
	}
	if oauthState == nil {
		return nil, fmt.Errorf("multiple integration options require oauth state")
	}
	stateUpdate := oauthState.Update().
		SetSelectionOptions(s.externalOptionsToMaps(options))
	if updateErr := stateUpdate.Exec(ctx); updateErr != nil {
		return nil, fmt.Errorf("save oauth selection options: %w", updateErr)
	}
	res := &rez.CompleteIntegrationOAuth2Result{
		Status:         "selection_required",
		SelectionToken: oauthState.State,
		Options:        options,
	}
	return res, nil
}

func (s *IntegrationsService) SelectOAuth2Flow(ctx context.Context, provider string, params rez.SelectIntegrationOAuth2Params) (*rez.CompleteIntegrationOAuth2Result, error) {
	state, stateErr := s.getOAuthState(ctx, provider, params.SelectionToken)
	if stateErr != nil {
		return nil, fmt.Errorf("invalid state: %w", stateErr)
	}
	options := s.externalOptionsFromMaps(state.SelectionOptions)
	allowed := make(map[string]rez.ExternalIntegrationOption, len(options))
	for _, option := range options {
		allowed[option.ExternalRef] = option
	}
	selected := make([]rez.ExternalIntegrationOption, 0, len(params.ExternalRefs))
	for _, ref := range params.ExternalRefs {
		option, ok := allowed[ref]
		if !ok {
			return nil, fmt.Errorf("invalid integration option selected: %s", ref)
		}
		selected = append(selected, option)
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("at least one integration option must be selected")
	}
	configured, cfgErr := s.configureOptions(ctx, provider, selected)
	if cfgErr != nil {
		return nil, cfgErr
	}
	if deleteErr := s.db.IntegrationOAuthState.DeleteOneID(state.ID).Exec(ctx); deleteErr != nil {
		slog.Error("failed to delete oauth selection state", "error", deleteErr)
	}
	return &rez.CompleteIntegrationOAuth2Result{Status: "configured", Configured: configured}, nil
}

func (s *IntegrationsService) configureOptions(ctx context.Context, provider string, options []rez.ExternalIntegrationOption) ([]rez.ConfiguredIntegration, error) {
	configured := make([]rez.ConfiguredIntegration, 0, len(options))
	for _, option := range options {
		params := rez.ConfigureIntegrationParams{
			Provider:    provider,
			DisplayName: option.DisplayName,
			ExternalRef: option.ExternalRef,
			Config:      option.Config,
		}
		ci, cfgErr := s.Configure(ctx, params)
		if cfgErr != nil {
			return nil, fmt.Errorf("set integration %s/%s: %w", provider, option.ExternalRef, cfgErr)
		}
		configured = append(configured, ci)
	}
	return configured, nil
}

func (s *IntegrationsService) externalOptionsToMaps(options []rez.ExternalIntegrationOption) []map[string]any {
	result := make([]map[string]any, 0, len(options))
	for _, option := range options {
		result = append(result, map[string]any{
			"externalRef": option.ExternalRef,
			"displayName": option.DisplayName,
			"config":      option.Config,
		})
	}
	return result
}

func (s *IntegrationsService) externalOptionsFromMaps(raw []map[string]any) []rez.ExternalIntegrationOption {
	result := make([]rez.ExternalIntegrationOption, 0, len(raw))
	for _, item := range raw {
		option := rez.ExternalIntegrationOption{}
		if ref, ok := item["externalRef"].(string); ok {
			option.ExternalRef = ref
		}
		if name, ok := item["displayName"].(string); ok {
			option.DisplayName = name
		}
		if cfg, ok := item["config"].(map[string]any); ok {
			option.Config = cfg
		}
		if option.ExternalRef != "" {
			result = append(result, option)
		}
	}
	return result
}

func (s *IntegrationsService) getConfiguredIntegrationForDataKind(ctx context.Context, dataKind string) (rez.ConfiguredIntegration, error) {
	var providers []string
	for _, p := range integrations.GetAvailable() {
		if slices.Contains(p.SupportedDataKinds(), dataKind) {
			providers = append(providers, p.Name())
		}
	}
	intgs, listErr := s.listIntegrations(ctx, rez.ListIntegrationsParams{Providers: providers})
	if listErr != nil {
		return nil, listErr
	}
	for _, intg := range intgs {
		p, pErr := integrations.GetPackage(intg.Provider)
		if pErr != nil {
			return nil, fmt.Errorf("get package %s: %w", intg.Provider, pErr)
		}
		ci := p.GetConfiguredIntegration(intg)
		if ci.GetAvailableDataKinds()[dataKind] {
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
	MakeVideoConferenceService(context.Context) (rez.VideoConferenceService, error)
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

func (s *IntegrationsService) RequestDataSync(ctx context.Context, providerSources map[string][]string) error {
	args := jobs.ProviderEventSyncJob{
		ProviderSources: providerSources,
		SyncReason:      "manual",
	}
	opts := &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
	_, insertErr := s.jobs.Insert(ctx, args, opts)
	if insertErr != nil {
		return fmt.Errorf("insert job: %w", insertErr)
	}
	return nil
}

func (s *IntegrationsService) GetDataSyncStatus(ctx context.Context, provider string) (*ent.ListResult[ent.ProviderEventSyncRun], error) {
	query := s.db.ProviderEventSyncRun.Query().
		Where(pesr.Provider(provider)).
		Order(pesr.ByStartedAt(sql.OrderDesc())).
		Limit(5)
	return ent.DoListQuery[ent.ProviderEventSyncRun, *ent.ProviderEventSyncRunQuery](ctx, query, ent.ListParams{Limit: 5})
}
