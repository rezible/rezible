package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	iesr "github.com/rezible/rezible/ent/integrationeventsyncrun"
	iuis "github.com/rezible/rezible/ent/integrationuserinstallstate"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/riverqueue/river"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	in "github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/jobs"
)

type IntegrationInstallationsService struct {
	db rez.Database
}

func NewIntegrationInstallationsService(db rez.Database) (*IntegrationInstallationsService, error) {
	return &IntegrationInstallationsService{db: db}, nil
}

func (s *IntegrationInstallationsService) LookupInstallation(ctx context.Context, pred predicate.Integration) (*ent.Integration, error) {
	return s.db.Client(ctx).Integration.Query().Where(pred).Only(ctx)
}

func (s *IntegrationInstallationsService) ListInstallations(ctx context.Context, preds ...predicate.Integration) ([]*ent.Integration, error) {
	return s.db.Client(ctx).Integration.Query().Where(preds...).All(ctx)
}

type IntegrationsService struct {
	db            rez.Database
	jobs          rez.JobService
	installations rez.IntegrationInstallationLookup
	registry      *integrations.Registry

	oauthRedirectUrlBase *url.URL
	webhookUrlBase       *url.URL
}

func NewIntegrationsService(cfg rez.Config, db rez.Database, jobSvc rez.JobService, installations rez.IntegrationInstallationLookup, reg *integrations.Registry) (*IntegrationsService, error) {
	redirectUrl, redirectUrlErr := cfg.App.GetFrontendUrl("/connect")
	if redirectUrlErr != nil {
		return nil, fmt.Errorf("invalid oauth callback url: %w", redirectUrlErr)
	}

	webhookUrl, webhookUrlErr := url.Parse("https://" + cfg.App.ApiDomain)
	if webhookUrlErr != nil {
		return nil, fmt.Errorf("invalid api domain: %w", webhookUrlErr)
	}

	s := &IntegrationsService{
		db:                   db,
		jobs:                 jobSvc,
		installations:        installations,
		registry:             reg,
		oauthRedirectUrlBase: redirectUrl,
		webhookUrlBase:       webhookUrl.JoinPath(cfg.HttpServer.BasePath, "webhooks"),
	}

	return s, nil
}

func (s *IntegrationsService) ListInstallable(ctx context.Context) ([]rez.IntegrationDefinition, error) {
	state, stateErr := s.getInstallState(ctx)
	if stateErr != nil {
		return nil, fmt.Errorf("get install state: %w", stateErr)
	}

	available := s.registry.GetAvailable()
	installable := make([]rez.IntegrationDefinition, 0, len(available))
	for _, p := range available {
		if requirementsErr := s.checkInstallRequirements(p, state); requirementsErr == nil {
			installable = append(installable, p)
		}
	}
	return installable, nil
}

func (s *IntegrationsService) getInstallState(ctx context.Context) (*integrations.IntegrationInstallState, error) {
	installed, listErr := s.ListAllInstalled(ctx)
	if listErr != nil {
		return nil, fmt.Errorf("list installed: %w", listErr)
	}

	prefs, prefsErr := s.db.Client(ctx).OrganizationPreferences.Query().Only(ctx)
	if prefsErr != nil && !ent.IsNotFound(prefsErr) {
		return nil, fmt.Errorf("query organization preferences: %w", prefsErr)
	}

	return &integrations.IntegrationInstallState{
		Installed:   installed,
		Preferences: prefs,
	}, nil
}

func (s *IntegrationsService) checkInstallRequirements(p rez.IntegrationDefinition, state *integrations.IntegrationInstallState) error {
	ri, hasRequirements := p.(integrations.IntegrationWithInstallRequirements)
	if !hasRequirements {
		return nil
	}
	if requirementsErr := ri.CheckInstallRequirements(state); requirementsErr != nil {
		return fmt.Errorf("%w: %w", rez.ErrConflict, requirementsErr)
	}
	return nil
}

// checkCanInstallNew applies installation limits and requirements to a new installation.
// Updating an existing installation for the same provider target is always allowed.
func (s *IntegrationsService) checkCanInstallNew(ctx context.Context, p rez.IntegrationDefinition) error {
	state, stateErr := s.getInstallState(ctx)
	if stateErr != nil {
		return fmt.Errorf("get install state: %w", stateErr)
	}

	if maxInstalls := p.MaxInstalls(); maxInstalls != nil {
		installCount := 0
		for _, ii := range state.Installed {
			if ii.Integration().Name == p.Name() {
				installCount++
			}
		}
		if installCount >= *maxInstalls {
			return fmt.Errorf("%w: %s allows at most %d installation(s)", rez.ErrConflict, p.DisplayName(), *maxInstalls)
		}
	}

	return s.checkInstallRequirements(p, state)
}

func (s *IntegrationsService) ListInstalled(ctx context.Context, params rez.ListIntegrationsParams) (*ent.ListResult[ent.Integration], error) {
	query := s.listQuery(ctx, params).Order(in.ByID(params.GetOrder()))
	return ent.DoListQuery[ent.Integration, *ent.IntegrationQuery](ctx, query, params.ListParams)
}

func (s *IntegrationsService) ListAllInstalled(ctx context.Context, predicates ...predicate.Integration) ([]rez.InstalledIntegration, error) {
	intgs, listErr := s.ListInstallations(ctx, predicates...)
	if listErr != nil {
		return nil, fmt.Errorf("query: %w", listErr)
	}
	cfgIs := make([]rez.InstalledIntegration, len(intgs))
	for i, intg := range intgs {
		ci, ciErr := s.AsInstalledIntegration(intg)
		if ciErr != nil {
			return nil, fmt.Errorf("as installed integration: %w", ciErr)
		}
		cfgIs[i] = ci
	}
	return cfgIs, nil
}

func (s *IntegrationsService) GetAvailableAgentTools(ctx context.Context, params rez.GetAvailableAiAgentToolsParams) ([]ai.Tool, error) {
	installed, listErr := s.ListAllInstalled(ctx)
	if listErr != nil {
		return nil, fmt.Errorf("list installed: %w", listErr)
	}

	pkgToolsMap, toolsErr := s.registry.GetAvailableAgentTools(ctx, installed, params)
	if toolsErr != nil {
		return nil, fmt.Errorf("get available tools: %w", toolsErr)
	}

	toolNames := mapset.NewSet[string]()
	var tools []ai.Tool
	for pkg, pkgTools := range pkgToolsMap {
		for _, tool := range pkgTools {
			toolName := tool.Name()
			if !toolNames.Add(toolName) {
				return nil, fmt.Errorf("duplicate agent tool %q from integration %s", toolName, pkg.Name())
			}
			tools = append(tools, tool)
		}
	}
	return tools, nil
}

func (s *IntegrationsService) GetInstalledIntegration(ctx context.Context, id uuid.UUID) (rez.InstalledIntegration, error) {
	intg, getErr := s.LookupInstallation(ctx, in.ID(id))
	if getErr != nil {
		if ent.IsNotFound(getErr) {
			return nil, fmt.Errorf("%w: integration installation %s", rez.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get integration: %w", getErr)
	}
	return s.AsInstalledIntegration(intg)
}

func (s *IntegrationsService) LookupInstallation(ctx context.Context, pred predicate.Integration) (*ent.Integration, error) {
	return s.installations.LookupInstallation(ctx, pred)
}

func (s *IntegrationsService) ListInstallations(ctx context.Context, preds ...predicate.Integration) ([]*ent.Integration, error) {
	return s.installations.ListInstallations(ctx, preds...)
}

func (s *IntegrationsService) InstallNew(ctx context.Context, name string, rawCfg []byte) (rez.InstalledIntegration, error) {
	p, pErr := s.registry.Get(name)
	if pErr != nil {
		return nil, fmt.Errorf("get integration package %s: %w", name, pErr)
	}
	cfg, cfgErr := p.ValidateInstallationConfig(rawCfg)
	if cfgErr != nil {
		return nil, fmt.Errorf("invalid installation config: %w", cfgErr)
	}
	return s.InstallFromTarget(ctx, rez.IntegrationInstallationTarget{
		DisplayName: p.DisplayName(),
		Config:      cfg,
	})
}

func (s *IntegrationsService) InstallFromTarget(ctx context.Context, target rez.IntegrationInstallationTarget) (rez.InstalledIntegration, error) {
	ref := target.Config.InstallationTargetRef()
	if validateErr := ref.Validate(); validateErr != nil {
		return nil, fmt.Errorf("invalid installation target: %w", validateErr)
	}
	p, pErr := s.registry.Get(ref.ProviderNamespace)
	if pErr != nil {
		return nil, fmt.Errorf("get integration package %s: %w", ref.ProviderNamespace, pErr)
	}
	if p.Provider() != ref.Provider {
		return nil, fmt.Errorf("installation target provider %q does not match integration provider %q", ref.Provider, p.Provider())
	}
	cfg, encErr := target.Config.Encode()
	if encErr != nil {
		return nil, fmt.Errorf("failed to encode config: %w", encErr)
	}
	_, validateErr := p.ValidateInstallationConfig(cfg)
	if validateErr != nil {
		return nil, fmt.Errorf("invalid config: %w", validateErr)
	}
	displayName := target.DisplayName
	if displayName == "" {
		displayName = p.DisplayName()
	}
	setFn := func(m *ent.IntegrationMutation) {
		m.SetProvider(p.Provider())
		m.SetName(ref.ProviderNamespace)
		m.SetDisplayName(displayName)
		m.SetProviderInstallationRef(ref.ResourceRef)
		m.SetInstallationConfig(cfg)
	}
	existingQuery := s.db.Client(ctx).Integration.Query()
	existingQuery.Where(
		in.Provider(p.Provider()),
		in.Name(ref.ProviderNamespace),
		in.ProviderInstallationRef(ref.ResourceRef),
	)
	installationID, queryExistingErr := existingQuery.OnlyID(ctx)
	if queryExistingErr != nil && !ent.IsNotFound(queryExistingErr) {
		return nil, fmt.Errorf("query existing installation: %w", queryExistingErr)
	}
	if ent.IsNotFound(queryExistingErr) {
		if installErr := s.checkCanInstallNew(ctx, p); installErr != nil {
			return nil, fmt.Errorf("check can install: %w", installErr)
		}
	}
	intg, setErr := s.set(ctx, installationID, setFn)
	if setErr != nil {
		return nil, fmt.Errorf("set integration: %w", setErr)
	}
	return p.GetInstalledIntegration(intg)
}

func (s *IntegrationsService) UpdateInstallation(ctx context.Context, id uuid.UUID, setFn func(*ent.IntegrationMutation)) (rez.InstalledIntegration, error) {
	curr, currErr := s.LookupInstallation(ctx, in.ID(id))
	if currErr != nil {
		return nil, fmt.Errorf("failed to get integration: %w", currErr)
	}
	p, pErr := s.registry.Get(curr.Name)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get package for integration %s: %w", curr.Name, pErr)
	}

	m := &ent.IntegrationMutation{}
	setFn(m)

	if userSettings, updatedSettings := m.UserSettings(); updatedSettings {
		if settingsErr := p.ValidateUserSettings(userSettings); settingsErr != nil {
			return nil, fmt.Errorf("invalid user settings: %w", settingsErr)
		}
	}

	intg, setErr := s.set(ctx, id, setFn)
	if setErr != nil {
		return nil, fmt.Errorf("failed to set integration: %w", setErr)
	}
	return p.GetInstalledIntegration(intg)
}

func (s *IntegrationsService) DeleteInstalled(ctx context.Context, id uuid.UUID) error {
	deleteErr := s.db.Client(ctx).Integration.DeleteOneID(id).Exec(ctx)
	return deleteErr
}

func (s *IntegrationsService) IssueWebhookToken(ctx context.Context, id uuid.UUID) (string, error) {
	curr, currErr := s.LookupInstallation(ctx, in.ID(id))
	if currErr != nil {
		if ent.IsNotFound(currErr) {
			return "", fmt.Errorf("%w: integration installation %s", rez.ErrNotFound, id)
		}
		return "", fmt.Errorf("get integration: %w", currErr)
	}
	token, tokenHash, tokenErr := integrations.NewWebhookToken()
	if tokenErr != nil {
		return "", fmt.Errorf("new webhook token: %w", tokenErr)
	}
	setToken := s.db.Client(ctx).Integration.UpdateOneID(curr.ID).
		SetWebhookTokenHash(tokenHash)
	if updateErr := setToken.Exec(ctx); updateErr != nil {
		return "", fmt.Errorf("store webhook token: %w", updateErr)
	}
	return s.webhookUrlBase.JoinPath(curr.Name, token).String(), nil
}

func (s *IntegrationsService) listQuery(ctx context.Context, p rez.ListIntegrationsParams) *ent.IntegrationQuery {
	query := s.db.Client(ctx).Integration.Query()
	if len(p.Predicates) > 0 {
		return query.Where(p.Predicates...)
	}
	return query
}

func (s *IntegrationsService) AsInstalledIntegration(i *ent.Integration) (rez.InstalledIntegration, error) {
	p, pErr := s.registry.Get(i.Name)
	if pErr != nil {
		return nil, fmt.Errorf("failed to get integration package: %w", pErr)
	}
	return p.GetInstalledIntegration(i)
}

func (s *IntegrationsService) LookupByProviderInstallationRef(ctx context.Context, name, resourceRef string) (*ent.Integration, error) {
	q := s.db.Client(ctx).Integration.Query().
		Where(in.Name(name)).
		Where(in.ProviderInstallationRef(resourceRef))
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
	curr, getCurrErr := s.LookupInstallation(ctx, in.ID(id))
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
		args := jobs.SyncIntegrationSourceEvents{
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
	state := uuid.New().String()
	createFreshStateFn := func(ctx context.Context, client *ent.Client) (string, error) {
		if delErr := s.deleteUserInstallationState(ctx, userId, intgName); delErr != nil {
			return "", fmt.Errorf("delete existing user install state: %w", delErr)
		}

		create := client.IntegrationUserInstallState.Create().
			SetUserID(userId).
			SetIntegrationName(intgName).
			SetOauthState(state).
			SetExpiresAt(time.Now().Add(time.Minute * 10))

		if createErr := create.Exec(ctx); createErr != nil {
			return "", createErr
		}
		return state, nil
	}
	return ent.WithTxReturning(ctx, s.db, createFreshStateFn)
}

func (s *IntegrationsService) updateUserInstallationStateWithOptions(ctx context.Context, id uuid.UUID, options []rez.IntegrationInstallationTarget) error {
	targets := map[string]json.RawMessage{}
	for _, opt := range options {
		if _, exists := targets[opt.DisplayName]; exists {
			return fmt.Errorf("multiple installation targets for display name %q", opt.DisplayName)
		}
		cfg, cfgErr := opt.Config.Encode()
		if cfgErr != nil {
			return fmt.Errorf("failed to encode integration option: %w", cfgErr)
		}
		targets[opt.DisplayName] = cfg
	}

	update := s.db.Client(ctx).IntegrationUserInstallState.UpdateOneID(id).
		SetInstallationTargetConfigs(targets).
		SetExpiresAt(time.Now().Add(time.Minute * 10))
	if updateErr := update.Exec(ctx); updateErr != nil {
		return fmt.Errorf("update installation state: %w", updateErr)
	}
	return nil
}

func (s *IntegrationsService) lookupUserInstallationState(ctx context.Context, userId uuid.UUID, intgName string) (*ent.IntegrationUserInstallState, error) {
	return s.db.Client(ctx).IntegrationUserInstallState.Query().
		Where(iuis.And(iuis.UserIDEQ(userId), iuis.IntegrationName(intgName))).
		Only(ctx)
}

func (s *IntegrationsService) getOAuthIntegration(name string) (rez.OAuth2FlowIntegration, *oauth2.Config, error) {
	oi, oiErr := s.registry.GetOAuth2FlowIntegration(name)
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
	p, pErr := s.registry.Get(integrationName)
	if pErr != nil {
		return "", fmt.Errorf("get integration package %s: %w", integrationName, pErr)
	}
	installState, installStateErr := s.getInstallState(ctx)
	if installStateErr != nil {
		return "", fmt.Errorf("get install state: %w", installStateErr)
	}
	if requirementsErr := s.checkInstallRequirements(p, installState); requirementsErr != nil {
		return "", fmt.Errorf("check install requirements: %w", requirementsErr)
	}
	state, stateErr := s.makeUserOAuthInstallationState(ctx, userId, integrationName)
	if stateErr != nil {
		return "", fmt.Errorf("failed to make oauth state: %w", stateErr)
	}
	return cfg.AuthCodeURL(state), nil
}

func (s *IntegrationsService) CompleteOAuth2Flow(ctx context.Context, integrationName string, params rez.CompleteIntegrationOAuth2FlowParams) (*rez.CompleteIntegrationOAuth2FlowResult, error) {
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
		installed, installErr := s.InstallTargets(ctx, options)
		if installErr != nil {
			return nil, fmt.Errorf("install single target options: %w", installErr)
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

func (s *IntegrationsService) decodeStateInstallationTargets(state *ent.IntegrationUserInstallState) ([]rez.IntegrationInstallationTarget, error) {
	targets := make([]rez.IntegrationInstallationTarget, 0, len(state.InstallationTargetConfigs))
	for displayName, rawCfg := range state.InstallationTargetConfigs {
		p, pErr := s.registry.Get(state.IntegrationName)
		if pErr != nil {
			return nil, fmt.Errorf("get integration: %w", pErr)
		}
		cfg, cfgErr := p.ValidateInstallationConfig(rawCfg)
		if cfgErr != nil {
			return nil, fmt.Errorf("invalid installation config: %w", cfgErr)
		}
		targets = append(targets, rez.IntegrationInstallationTarget{
			DisplayName: displayName,
			Config:      cfg,
		})
	}
	return targets, nil
}

func (s *IntegrationsService) ListUserInstallationTargets(ctx context.Context) ([]rez.IntegrationInstallationTarget, error) {
	userId, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return nil, rez.ErrAuthSessionMissing
	}
	query := s.db.Client(ctx).IntegrationUserInstallState.Query().
		Where(iuis.UserID(userId)).
		Where(iuis.InstallationTargetConfigsNotNil())
	states, queryErr := query.All(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("query failed: %w", queryErr)
	}
	targets := make([]rez.IntegrationInstallationTarget, 0, len(states))
	for _, state := range states {
		stateTargets, targetErr := s.decodeStateInstallationTargets(state)
		if targetErr != nil {
			return nil, fmt.Errorf("failed to decode installation targets: %w", targetErr)
		}
		targets = append(targets, stateTargets...)
	}
	return targets, nil
}

func (s *IntegrationsService) InstallFromUserInstallationTargets(ctx context.Context, intgName string, resourceRefs []string) ([]rez.InstalledIntegration, error) {
	userId, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return nil, rez.ErrAuthSessionMissing
	}
	state, stateErr := s.lookupUserInstallationState(ctx, userId, intgName)
	if stateErr != nil {
		return nil, fmt.Errorf("invalid state: %w", stateErr)
	}
	options, decodeErr := s.decodeStateInstallationTargets(state)
	if decodeErr != nil {
		return nil, fmt.Errorf("decode installation targets: %w", decodeErr)
	}

	selectedRefs := mapset.NewSet(resourceRefs...)

	selected := make([]rez.IntegrationInstallationTarget, 0, selectedRefs.Cardinality())
	for _, option := range options {
		ref := option.Config.InstallationTargetRef()
		if ref.ProviderNamespace == intgName && selectedRefs.Contains(ref.ResourceRef) {
			selected = append(selected, option)
		}
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("at least one integration option must be selected")
	}
	installed, installErr := s.InstallTargets(ctx, selected)
	if installErr != nil {
		return nil, fmt.Errorf("failed to install targets: %w", installErr)
	}
	if deleteStateErr := s.deleteUserInstallationState(ctx, userId, intgName); deleteStateErr != nil {
		slog.ErrorContext(ctx, "failed to delete user installation state",
			"error", deleteStateErr)
	}
	return installed, nil
}

func (s *IntegrationsService) InstallTargets(ctx context.Context, targets []rez.IntegrationInstallationTarget) ([]rez.InstalledIntegration, error) {
	installed := make([]rez.InstalledIntegration, 0, len(targets))
	for _, target := range targets {
		ii, installErr := s.InstallFromTarget(ctx, target)
		if installErr != nil {
			return nil, fmt.Errorf("install integration target %s: %w", target.DisplayName, installErr)
		}
		installed = append(installed, ii)
	}
	return installed, nil
}

func (s *IntegrationsService) RequestIntegrationEventSync(ctx context.Context, id uuid.UUID, sources []string) error {
	args := jobs.SyncIntegrationSourceEvents{
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

func (s *IntegrationsService) ListIntegrationEventSyncRuns(ctx context.Context, id uuid.UUID) ([]*ent.IntegrationEventSyncRun, error) {
	query := s.db.Client(ctx).IntegrationEventSyncRun.Query().
		Where(iesr.IntegrationID(id)).
		Order(iesr.ByStartedAt(sql.OrderDesc()), iesr.ByID(sql.OrderDesc())).
		Limit(5)
	return query.All(ctx)
}
