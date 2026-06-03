package db

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	iesr "github.com/rezible/rezible/ent/integrationeventsyncrun"
	iuis "github.com/rezible/rezible/ent/integrationuserinstallstate"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/riverqueue/river"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	in "github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/jobs"
)

type IntegrationsService struct {
	db   rez.Database
	jobs rez.JobService
	reg  *integrations.PackageRegistry

	oauthRedirectUrlBase *url.URL
}

func NewIntegrationsService(appCfg rez.AppConfig, db rez.Database, jobs rez.JobService, reg *integrations.PackageRegistry) (*IntegrationsService, error) {
	callbackUrl, callbackUrlErr := url.JoinPath(appCfg.FrontendUrl, "/settings/integration-callback")
	if callbackUrlErr != nil {
		return nil, fmt.Errorf("invalid oauth callback url: %w", callbackUrlErr)
	}
	redirectUrl, urlErr := url.Parse(callbackUrl)
	if urlErr != nil {
		return nil, fmt.Errorf("invalid oauth redirect url: %w", urlErr)
	}

	s := &IntegrationsService{
		db:   db,
		jobs: jobs,
		reg:  reg,

		oauthRedirectUrlBase: redirectUrl,
	}

	return s, nil
}

func (s *IntegrationsService) GetAvailable() []rez.IntegrationPackage {
	return s.reg.GetAvailable()
}

func (s *IntegrationsService) ListInstalled(ctx context.Context, params rez.ListIntegrationsParams) ([]rez.InstalledIntegration, error) {
	intgs, listErr := s.listIntegrations(ctx, params)
	if listErr != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", listErr)
	}
	cfgIs := make([]rez.InstalledIntegration, len(intgs))
	for i, intg := range intgs {
		ci, ciErr := s.asInstalledIntegration(intg)
		if ciErr != nil {
			return nil, fmt.Errorf("failed to list integrations: %w", ciErr)
		}
		cfgIs[i] = ci
	}
	return cfgIs, nil
}

func (s *IntegrationsService) LookupByRef(ctx context.Context, name string, providerRef string) (*ent.Integration, error) {
	query := s.db.Client(ctx).Integration.Query().
		Where(in.And(in.IntegrationName(name), in.ExternalProviderRef(providerRef)))
	res, resErr := query.Only(ctx)
	if resErr != nil {
		return nil, fmt.Errorf("failed to lookup integration: %w", resErr)
	}
	return res, nil
}

func (s *IntegrationsService) GetInstalled(ctx context.Context, id uuid.UUID) (rez.InstalledIntegration, error) {
	intg, getErr := s.getById(ctx, id)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get integration: %w", getErr)
	}
	return s.asInstalledIntegration(intg)
}

func (s *IntegrationsService) InstallNew(ctx context.Context, intgName string, params rez.InstallIntegrationParams) (rez.InstalledIntegration, error) {
	p, pErr := s.reg.GetPackage(intgName)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get package for integration %s: %w", intgName, pErr)
	}

	if settingsErr := p.ValidateUserSettings(params.UserSettings); settingsErr != nil {
		return nil, fmt.Errorf("failed to validate user settings: %w", settingsErr)
	}

	externalRef, cfgErr := p.ValidateInstallationConfig(params.InstallationConfig)
	if cfgErr != nil {
		return nil, fmt.Errorf("invalid config: %w", cfgErr)
	}

	setFn := func(m *ent.IntegrationMutation) {
		m.SetIntegrationName(intgName)
		m.SetDisplayName(params.DisplayName)
		m.SetExternalProviderRef(externalRef)
		m.SetInstallationConfig(params.InstallationConfig)
		m.SetUserSettings(params.UserSettings)
	}
	intg, setErr := s.set(ctx, uuid.Nil, setFn)
	if setErr != nil {
		return nil, fmt.Errorf("failed to set integration: %w", setErr)
	}
	return p.GetInstalledIntegration(intg), nil
}

func (s *IntegrationsService) UpdateInstalled(ctx context.Context, id uuid.UUID, settings map[string]any) (rez.InstalledIntegration, error) {
	curr, currErr := s.getById(ctx, id)
	if currErr != nil {
		return nil, fmt.Errorf("failed to get integration: %w", currErr)
	}
	p, pErr := s.reg.GetPackage(curr.IntegrationName)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get package for integration %s: %w", curr.IntegrationName, pErr)
	}
	if settingsErr := p.ValidateUserSettings(settings); settingsErr != nil {
		return nil, fmt.Errorf("invalid user settings: %w", settingsErr)
	}
	setFn := func(m *ent.IntegrationMutation) {
		m.SetUserSettings(settings)
	}
	intg, setErr := s.set(ctx, id, setFn)
	if setErr != nil {
		return nil, fmt.Errorf("failed to set integration: %w", setErr)
	}
	return p.GetInstalledIntegration(intg), nil
}

func (s *IntegrationsService) DeleteInstalled(ctx context.Context, id uuid.UUID) error {
	deleteErr := s.db.Client(ctx).Integration.DeleteOneID(id).Exec(ctx)
	return deleteErr
}

func (s *IntegrationsService) listQuery(ctx context.Context, p rez.ListIntegrationsParams) *ent.IntegrationQuery {
	query := s.db.Client(ctx).Integration.Query()
	if len(p.IDs) > 0 {
		query.Where(in.IDIn(p.IDs...))
	}
	if len(p.Providers) > 0 {
		if len(p.Providers) == 1 {
			query.Where(in.IntegrationName(p.Providers[0]))
		} else {
			query.Where(in.IntegrationNameIn(p.Providers...))
		}
	}
	if len(p.ExternalRefs) > 0 {
		if len(p.ExternalRefs) == 1 {
			query.Where(in.ExternalProviderRef(p.ExternalRefs[0]))
		} else {
			query.Where(in.ExternalProviderRefIn(p.ExternalRefs...))
		}
	}
	if p.ConfigValues != nil && len(p.ConfigValues) > 0 {
		for path, value := range p.ConfigValues {
			query.Where(func(s *sql.Selector) {
				s.Where(sqljson.ValueEQ(in.FieldInstallationConfig, value, sqljson.DotPath(path)))
			})
		}
	}
	return query
}

func (s *IntegrationsService) asInstalledIntegration(i *ent.Integration) (rez.InstalledIntegration, error) {
	p, pErr := s.reg.GetPackage(i.IntegrationName)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get integration package: %w", pErr)
	}
	return p.GetInstalledIntegration(i), nil
}

func (s *IntegrationsService) listIntegrations(ctx context.Context, params rez.ListIntegrationsParams) ([]*ent.Integration, error) {
	q := s.listQuery(ctx, params)
	intgs, listErr := q.All(ctx)
	if listErr != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", listErr)
	}
	return intgs, nil
}

func (s *IntegrationsService) getById(ctx context.Context, id uuid.UUID) (*ent.Integration, error) {
	return s.db.Client(ctx).Integration.Get(ctx, id)
}

func (s *IntegrationsService) getByProviderExternalRef(ctx context.Context, integrationName, externalRef string) (*ent.Integration, error) {
	q := s.db.Client(ctx).Integration.Query().
		Where(in.IntegrationName(integrationName)).
		Where(in.ExternalProviderRef(externalRef))
	intg, getErr := q.Only(ctx)
	if getErr != nil {
		if ent.IsNotFound(getErr) {
			return nil, getErr
		}
		return nil, fmt.Errorf("failed to get integration: %w", getErr)
	}
	return intg, nil
}

func (s *IntegrationsService) set(ctx context.Context, id uuid.UUID, setFn func(*ent.IntegrationMutation)) (*ent.Integration, error) {
	curr, getCurrErr := s.getById(ctx, id)
	if getCurrErr != nil && !ent.IsNotFound(getCurrErr) {
		return nil, fmt.Errorf("failed to get integration: %w", getCurrErr)
	}

	var upsert ent.EntityMutator[*ent.Integration, *ent.IntegrationMutation]
	if curr == nil {
		upsert = s.db.Client(ctx).Integration.Create()
	} else {
		upsert = s.db.Client(ctx).Integration.UpdateOneID(curr.ID)
	}

	setFn(upsert.Mutation())

	intg, saveErr := upsert.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("failed to save: %w", saveErr)
	}

	// TODO: check updated fields?
	shouldTriggerSync := s.jobs != nil
	if shouldTriggerSync {
		args := jobs.SyncIntegrationEventsArgs{
			IntegrationId: intg.ID,
			SyncReason:    "updated",
		}
		if _, jobErr := s.jobs.Insert(ctx, args, nil); jobErr != nil {
			slog.Error("failed to insert sync job", "error", jobErr)
		}
	}

	return intg, nil
}

func (s *IntegrationsService) makeUserInstallStatePredicate(ctx context.Context, intgName string) (predicate.IntegrationUserInstallState, error) {
	userId, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return nil, rez.ErrAuthSessionMissing
	}
	return iuis.And(iuis.UserIDEQ(userId), iuis.IntegrationName(intgName)), nil
}

func (s *IntegrationsService) deleteUserInstallationState(ctx context.Context, userId uuid.UUID, intgName string) error {
	del := s.db.Client(ctx).IntegrationUserInstallState.Delete().
		Where(iuis.And(iuis.UserIDEQ(userId), iuis.IntegrationName(intgName)))
	_, deleteErr := del.Exec(ctx)
	return deleteErr
}

func (s *IntegrationsService) makeUserOAuthInstallationState(ctx context.Context, userId uuid.UUID, intgName string) (string, error) {
	// TODO: replace this with something actually random
	oauthState := uuid.New().String()
	createFreshStateFn := func(ctx context.Context, client *ent.Client) error {
		if delErr := s.deleteUserInstallationState(ctx, userId, intgName); delErr != nil {
			return fmt.Errorf("delete existing user install state: %w", delErr)
		}
		state := uuid.New().String()

		create := client.IntegrationUserInstallState.Create().
			SetUserID(userId).
			SetOauthState(state).
			SetIntegrationName(intgName).
			SetExpiresAt(time.Now().Add(time.Minute * 10))

		return create.Exec(ctx)
	}
	if txErr := s.db.WithTx(ctx, createFreshStateFn); txErr != nil {
		return "", txErr
	}

	return oauthState, nil
}

func (s *IntegrationsService) updateUserInstallationStateWithOptions(ctx context.Context, id uuid.UUID, options []rez.IntegrationInstallationTarget) error {
	targets, encErr := integrations.EncodeInstallationTargetOptions(options)
	if encErr != nil {
		return fmt.Errorf("encode installation targets: %w", encErr)
	}

	update := s.db.Client(ctx).IntegrationUserInstallState.UpdateOneID(id).
		SetInstallationTargets(targets).
		SetExpiresAt(time.Now().Add(time.Minute * 10))
	if updateErr := update.Exec(ctx); updateErr != nil {
		return fmt.Errorf("update installation state: %w", updateErr)
	}
	return nil
}

func (s *IntegrationsService) lookupUserInstallationState(ctx context.Context, userId uuid.UUID, intgName string) (*ent.IntegrationUserInstallState, error) {
	query := s.db.Client(ctx).IntegrationUserInstallState.Query().
		Where(iuis.And(iuis.UserIDEQ(userId), iuis.IntegrationName(intgName)))
	state, queryErr := query.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query failed: %w", queryErr)
	}
	return state, nil
}

func (s *IntegrationsService) getOAuthIntegration(name string) (integrations.IntegrationWithOAuth2Flow, *oauth2.Config, error) {
	oi, oiErr := s.reg.GetOAuthIntegration(name)
	if oiErr != nil {
		return nil, nil, fmt.Errorf("invalid integration: %w", oiErr)
	}
	cfg := oi.OAuth2Config()
	if redirectUrl := s.oauthRedirectUrlBase.JoinPath(name); redirectUrl != nil {
		cfg.RedirectURL = redirectUrl.String()
	}
	return oi, cfg, nil
}

func (s *IntegrationsService) StartOAuth2Flow(ctx context.Context, integrationName string) (string, error) {
	userId, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return "", rez.ErrAuthSessionMissing
	}
	_, cfg, intgErr := s.getOAuthIntegration(integrationName)
	if intgErr != nil {
		return "", fmt.Errorf("failed to get integration: %w", intgErr)
	}
	state, stateErr := s.makeUserOAuthInstallationState(ctx, userId, integrationName)
	if stateErr != nil {
		return "", fmt.Errorf("failed to make oauth state: %w", stateErr)
	}
	return cfg.AuthCodeURL(state), nil
}

func (s *IntegrationsService) CompleteOAuth2Flow(ctx context.Context, integrationName string, params rez.CompleteIntegrationOAuth2Params) (*rez.CompleteIntegrationOAuth2FlowResult, error) {
	if params.State == nil && params.ClientVerifier == nil {
		return nil, fmt.Errorf("invalid params: missing state or client_verifier")
	}
	userId, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return nil, rez.ErrAuthSessionMissing
	}

	oi, cfg, intgErr := s.getOAuthIntegration(integrationName)
	if intgErr != nil {
		return nil, fmt.Errorf("failed to get integration: %w", intgErr)
	}
	installState, installStateErr := s.lookupUserInstallationState(ctx, userId, integrationName)
	if installStateErr != nil {
		return nil, fmt.Errorf("failed to get installation state: %w", installStateErr)
	}
	var opts []oauth2.AuthCodeOption
	if params.State != nil {
		if installState.OauthState != *params.State {
			return nil, fmt.Errorf("oauth state mismatch")
		}
	}
	if params.ClientVerifier != nil {
		opts = append(opts, oauth2.VerifierOption(*params.ClientVerifier))
	}
	token, tokenErr := cfg.Exchange(ctx, params.Code, opts...)
	if tokenErr != nil {
		return nil, fmt.Errorf("exchange token: %w", tokenErr)
	}
	options, optionsErr := oi.RetrieveInstallationTargetOptions(ctx, token)
	if optionsErr != nil {
		return nil, fmt.Errorf("retrieve installation targets: %w", optionsErr)
	}

	if len(options) == 0 {
		return nil, fmt.Errorf("no installation targets returned")
	}

	if len(options) == 1 {
		installed, cfgErr := s.installTargets(ctx, integrationName, options)
		if cfgErr != nil {
			return nil, fmt.Errorf("install single target options: %w", cfgErr)
		}
		return &rez.CompleteIntegrationOAuth2FlowResult{Installed: installed}, nil
	}

	if installState == nil {
		return nil, fmt.Errorf("multiple integration options require installation state")
	}

	updateErr := s.updateUserInstallationStateWithOptions(ctx, installState.ID, options)
	if updateErr != nil {
		return nil, fmt.Errorf("failed to update user installation state: %w", updateErr)
	}
	return &rez.CompleteIntegrationOAuth2FlowResult{
		InstallationTargetSelectionRequired: true,
		InstallationTargetOptions:           options,
	}, nil
}

func (s *IntegrationsService) ListUserInstallationTargets(ctx context.Context) (map[string][]rez.IntegrationInstallationTarget, error) {
	userId, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return nil, rez.ErrAuthSessionMissing
	}
	query := s.db.Client(ctx).IntegrationUserInstallState.Query().
		Where(iuis.UserID(userId)).
		Where(iuis.InstallationTargetsNotNil())
	states, queryErr := query.All(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("query failed: %w", queryErr)
	}
	targets := make(map[string][]rez.IntegrationInstallationTarget)
	for _, state := range states {
		opts, optsErr := integrations.DecodeInstallationTargetOptions(state.InstallationTargets)
		if optsErr != nil {
			return nil, fmt.Errorf("decode installation targets: %w", optsErr)
		}
		targets[state.IntegrationName] = opts
	}
	return targets, nil
}

func (s *IntegrationsService) InstallFromUserInstallationTargets(ctx context.Context, intgName string, externalRefs []string) ([]rez.InstalledIntegration, error) {
	userId, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return nil, rez.ErrAuthSessionMissing
	}
	state, stateErr := s.lookupUserInstallationState(ctx, userId, intgName)
	if stateErr != nil {
		return nil, fmt.Errorf("invalid state: %w", stateErr)
	}
	options, decodeErr := integrations.DecodeInstallationTargetOptions(state.InstallationTargets)
	if decodeErr != nil {
		return nil, fmt.Errorf("decode installation targets: %w", decodeErr)
	}

	selectedRefs := mapset.NewSet(externalRefs...)

	selected := make([]rez.IntegrationInstallationTarget, 0, selectedRefs.Cardinality())
	for _, option := range options {
		if selectedRefs.Contains(option.ExternalRef) {
			selected = append(selected, option)
		}
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("at least one integration option must be selected")
	}

	installed, installErr := s.installTargets(ctx, intgName, selected)
	if installErr != nil {
		return nil, fmt.Errorf("failed to install targets: %w", installErr)
	}
	if deleteStateErr := s.deleteUserInstallationState(ctx, userId, intgName); deleteStateErr != nil {
		slog.ErrorContext(ctx, "failed to delete user installation state",
			"error", deleteStateErr)
	}
	return installed, nil
}

func (s *IntegrationsService) installTargets(ctx context.Context, intgName string, options []rez.IntegrationInstallationTarget) ([]rez.InstalledIntegration, error) {
	installed := make([]rez.InstalledIntegration, 0, len(options))
	for _, option := range options {
		params := rez.InstallIntegrationParams{
			DisplayName:        option.DisplayName,
			InstallationConfig: option.InstallationConfig,
		}
		ci, cfgErr := s.InstallNew(ctx, intgName, params)
		if cfgErr != nil {
			return nil, fmt.Errorf("install integration %s option %s: %w", intgName, option.DisplayName, cfgErr)
		}
		installed = append(installed, ci)
	}
	return installed, nil
}

func (s *IntegrationsService) GetProviderEventProcessor(name string) (rez.ProviderEventProcessor, error) {
	procs := s.reg.GetProviderEventProcessors()
	proc, ok := procs[name]
	if !ok {
		return nil, fmt.Errorf("integration %s not found", name)
	}
	return proc, nil
}

func (s *IntegrationsService) GetProviderEventQuerier(ctx context.Context, intg *ent.Integration) (rez.IntegrationEventQuerier, error) {
	return s.reg.GetProviderEventQuerier(intg)
}

func (s *IntegrationsService) RequestIntegrationEventSync(ctx context.Context, id uuid.UUID, sources []string) error {
	args := jobs.SyncIntegrationEventsArgs{
		SyncReason:    "manual",
		IntegrationId: id,
		Sources:       sources,
	}
	opts := &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
	_, insertErr := s.jobs.Insert(ctx, args, opts)
	return insertErr
}

func (s *IntegrationsService) ListIntegrationEventSyncRuns(ctx context.Context, id uuid.UUID) (*ent.ListResult[ent.IntegrationEventSyncRun], error) {
	query := s.db.Client(ctx).IntegrationEventSyncRun.Query().
		Where(iesr.IntegrationID(id)).
		Order(iesr.ByStartedAt(sql.OrderDesc())).
		Limit(5)
	return ent.DoListQuery[ent.IntegrationEventSyncRun, *ent.IntegrationEventSyncRunQuery](ctx, query, ent.ListParams{Limit: 5})
}
