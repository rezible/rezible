package slackincidents

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/messages"
	"github.com/riverqueue/river"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type App struct {
	cfg        rez.Config
	db         rez.Database
	jobs       rez.JobService
	orgs       rez.OrganizationService
	incidents  rez.IncidentDetailsService
	management rez.IncidentManagementService
}

func MakeApp(cfg rez.Config, db rez.Database, js rez.JobService, orgs rez.OrganizationService, incidents rez.IncidentDetailsService, management rez.IncidentManagementService) *App {
	return &App{
		cfg:        cfg,
		db:         db,
		jobs:       js,
		orgs:       orgs,
		incidents:  incidents,
		management: management,
	}
}

func (a *App) MakeIntegration(deps *slackintegration.AppServiceDependencies) (*Integration, error) {
	svc, svcErr := slackintegration.NewAppService(a, deps)
	if svcErr != nil {
		return nil, fmt.Errorf("make app service: %w", svcErr)
	}
	return MakeIntegration(svc), nil
}

func (a *App) MakeMessageHandlers() []rez.MessageEventHandler {
	return []rez.MessageEventHandler{
		messages.NewEventHandler("slack.incidents.updated", a.onIncidentUpdated),
		messages.NewEventHandler("slack.incidents.milestone_updated", a.onIncidentMilestoneUpdated),
	}
}

func (a *App) IntegrationName() string {
	return integrationName
}

func (a *App) Config() rez.IntegrationsConfigSlackApp {
	return a.cfg.Integrations.Slack.Incidents
}

func (a *App) OAuthScopes() []string {
	return []string{
		"app_mentions:read",
		"assistant:write",
		"channels:history",
		"channels:join",
		"channels:read",
		"chat:write",
		"chat:write.customize",
		"chat:write.public",
		"commands",
		"groups:history",
		"groups:read",
		"im:history",
		"im:read",
		"im:write",
		"im:write.topic",
		"incoming-webhook",
		"metadata.message:read",
		"mpim:history",
		"pins:read",
		"reactions:read",
		"usergroups:read",
		"users.profile:read",
		"users:read",
		"users:read.email",
		"channels:write.topic",
		"channels:manage",
		"channels:write.invites",
	}
}

func (a *App) PublishEventTypes() []slackevents.EventsAPIType {
	return []slackevents.EventsAPIType{}
}

func (a *App) SlashCommandHandlers() map[string]slackintegration.SlashCommandHandler {
	return map[string]slackintegration.SlashCommandHandler{
		"/incident": a.handleIncidentCommand,
	}
}

func (a *App) InteractionCallbackHandlers() map[slack.InteractionType]slackintegration.InteractionCallbackHandler {
	return map[slack.InteractionType]slackintegration.InteractionCallbackHandler{
		slack.InteractionTypeMessageAction:  a.handleMessageActionInteraction,
		slack.InteractionTypeBlockActions:   a.handleBlockActionsInteraction,
		slack.InteractionTypeViewSubmission: a.handleViewSubmissionInteraction,
	}
}

// getEnabledIntegrationClient returns a client for the organization's installation, or nil when there is none.
// An organization has at most one incident integration, so the first installation is the only one.
func (a *App) getEnabledIntegrationClient(ctx context.Context) (*slackintegration.ClientWrapper, error) {
	queryInstall := a.db.Client(ctx).Integration.Query().
		Where(integration.Name(integrationName))
	intg, intgErr := queryInstall.First(ctx)
	if intgErr != nil {
		if ent.IsNotFound(intgErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("query slack incidents installation: %w", intgErr)
	}
	return a.newClientWrapper(intg)
}

func (a *App) clientOptions() []slack.Option {
	return slackintegration.ClientOptions(a.cfg.Integrations.Slack)
}

func (a *App) newClientWrapper(intg *ent.Integration) (*slackintegration.ClientWrapper, error) {
	return slackintegration.NewClientWrapper(intg, a.clientOptions()...)
}

// hasAgent is whether the Slack agent is installed in the same workspace as the installation.
func (a *App) hasAgent(ctx context.Context, intg *ent.Integration) (bool, error) {
	return a.db.Client(ctx).Integration.Query().
		Where(integration.Name(agentIntegrationName), integration.ProviderInstallationRef(intg.ProviderInstallationRef)).
		Exist(ctx)
}

// onIncidentUpdated decides nothing: it schedules a sync, which reads the incident when it runs.
func (a *App) onIncidentUpdated(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	return a.insertChannelSync(ctx, ev.IncidentId)
}

func (a *App) onIncidentMilestoneUpdated(ctx context.Context, ev *rez.EventOnIncidentMilestoneUpdated) error {
	args := postIncidentMilestoneJobArgs{IncidentID: ev.IncidentId, MilestoneID: ev.MilestoneId}
	if _, insertErr := a.jobs.Insert(ctx, args, nil); insertErr != nil {
		return fmt.Errorf("insert incident milestone message job: %w", insertErr)
	}
	return a.insertChannelSync(ctx, ev.IncidentId)
}

func (a *App) insertChannelSync(ctx context.Context, incidentID uuid.UUID) error {
	dueAt := time.Now().Truncate(time.Second).Add(time.Second)
	args := syncIncidentChannelJobArgs{IncidentID: incidentID, DueAt: dueAt}
	if _, insertErr := a.jobs.Insert(ctx, args, &river.InsertOpts{ScheduledAt: dueAt}); insertErr != nil {
		return fmt.Errorf("insert incident channel sync job: %w", insertErr)
	}
	return nil
}
