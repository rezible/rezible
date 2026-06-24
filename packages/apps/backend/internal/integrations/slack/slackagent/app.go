package slackagent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/jobs"
	"github.com/riverqueue/river"
	"github.com/slack-go/slack/slackevents"
)

type App struct {
	cfg      rez.Config
	db       rez.Database
	jobs     rez.JobService
	messages rez.MessageService
	agents   rez.AgentService
	events   rez.EventsService
}

func MakeApp(cfg rez.Config, db rez.Database, jobSvc rez.JobService, msgs rez.MessageService, agents rez.AgentService, events rez.EventsService) (*App, error) {
	h := &App{
		cfg:      cfg,
		db:       db,
		jobs:     jobSvc,
		messages: msgs,
		agents:   agents,
		events:   events,
	}
	if msgsErr := h.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}
	jobs.RegisterWorkerFunc(h.handlePostAlertInvestigationUpdate)
	return h, nil
}

func (a *App) registerMessageHandlers() error {
	return errors.Join(
		a.messages.AddEventHandlers(
			rez.NewEventHandler("slack_agent.alert_investigation_completed", a.onAgentRunCompleted),
		),
	)
}

func (a *App) IntegrationName() string {
	return integrationName
}

func (a *App) AppConfig() rez.IntegrationsConfigSlackApp {
	return a.cfg.Integrations.Slack.Agent
}

func (a *App) PublishProviderEventPipelineEventTypes() []slackevents.EventsAPIType {
	return []slackevents.EventsAPIType{}
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

var postAlertInvestigationUpdateJobOpts = &river.InsertOpts{
	UniqueOpts: river.UniqueOpts{
		ByArgs:  true,
		ByState: jobs.UniqueStateNonCompleted,
	},
}

func (a *App) onAgentRunCompleted(ctx context.Context, ev *rez.EventOnAgentRunCompleted) error {
	if ev.WorkflowKind != rez.AgentWorkflowKindAlertInvestigation {
		return nil
	}
	run, runErr := a.agents.GetRun(ctx, ev.AgentRunID)
	if runErr != nil {
		return fmt.Errorf("get completed agent run: %w", runErr)
	}
	if run.Edges.AgentTask == nil {
		return nil
	}
	alertID := agentTaskSubject(run.Edges.AgentTask.WorkflowInput, "alert")
	if alertID == uuid.Nil {
		return nil
	}
	_, insertErr := a.jobs.Insert(ctx, jobs.PostAlertInvestigationUpdate{
		AlertID:    alertID,
		AgentRunID: run.ID,
	}, postAlertInvestigationUpdateJobOpts)
	if insertErr != nil {
		return fmt.Errorf("enqueue alert investigation slack update: %w", insertErr)
	}
	return nil
}

func (a *App) handlePostAlertInvestigationUpdate(ctx context.Context, args jobs.PostAlertInvestigationUpdate) error {
	if args.AlertID == uuid.Nil || args.AgentRunID == uuid.Nil {
		return nil
	}
	msgID, findErr := a.findMessageIdForAlertEvent(ctx, args.AlertID)
	if findErr != nil {
		return fmt.Errorf("find alert slack message: %w", findErr)
	}
	if msgID == "" {
		return nil
	}
	channelID, threadTS, ok := strings.Cut(msgID.String(), "_")
	if !ok || channelID == "" || threadTS == "" {
		return fmt.Errorf("invalid slack message id %q", msgID.String())
	}
	client, clientErr := a.getEnabledIntegrationClient(ctx)
	if clientErr != nil {
		return fmt.Errorf("get slack agent client: %w", clientErr)
	}
	if client == nil {
		return nil
	}
	run, runErr := a.agents.GetRun(ctx, args.AgentRunID)
	if runErr != nil {
		return fmt.Errorf("get agent run: %w", runErr)
	}
	result, resultErr := a.agents.GetRunResult(ctx, run.ID)
	if resultErr != nil {
		return fmt.Errorf("get agent run result: %w", resultErr)
	}
	findings, findingsErr := a.agents.ListRunFindings(ctx, run.ID)
	if findingsErr != nil {
		return fmt.Errorf("list agent run findings: %w", findingsErr)
	}
	message := buildAlertInvestigationThreadMessage(result, findings)
	if message == "" {
		return nil
	}
	if _, postErr := client.SendReply(ctx, channelID, threadTS, message); postErr != nil {
		return fmt.Errorf("post alert investigation reply: %w", postErr)
	}
	return nil
}

func (a *App) getEnabledIntegrationClient(ctx context.Context) (*slackintegration.ClientWrapper, error) {
	intgs, intgErr := a.db.Client(ctx).Integration.Query().
		Where(integration.IntegrationName(integrationName)).
		All(ctx)
	if intgErr != nil && !ent.IsNotFound(intgErr) {
		return nil, fmt.Errorf("query slack agent integration: %w", intgErr)
	}
	if len(intgs) == 0 {
		return nil, nil
	}
	return slackintegration.NewClientWrapper(intgs[0])
}

func (a *App) findMessageIdForAlertEvent(ctx context.Context, alertID uuid.UUID) (slackintegration.MessageId, error) {
	_ = ctx
	_ = alertID
	return "", nil
}

func buildAlertInvestigationThreadMessage(result *ent.AgentRunResult, findings []*ent.AgentRunFinding) string {
	if result == nil {
		return ""
	}
	lines := []string{"*Rezible investigation findings*"}
	if result.Content != "" {
		lines = append(lines, result.Content)
	}
	resultFindings, _ := result.Data["findings"].(map[string]any)
	if cause, ok := resultFindings["likelyCause"].(string); ok && cause != "" {
		lines = append(lines, "*Current hypothesis:* "+cause)
	}
	if next, ok := resultFindings["recommendedNext"].(string); ok && next != "" {
		lines = append(lines, "*Recommended next:* "+next)
	}
	if checks, ok := resultFindings["suggestedChecks"].([]any); ok && len(checks) > 0 {
		lines = append(lines, "*Suggested checks:*")
		for _, check := range checks {
			if text, textOk := check.(string); textOk && text != "" {
				lines = append(lines, "- "+text)
			}
		}
	}
	if len(lines) == 2 {
		for _, finding := range findings {
			if finding.FindingKind == "recommendation" {
				lines = append(lines, "*Recommended next:* "+finding.Content)
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

func agentTaskSubject(input map[string]any, subjectType string) uuid.UUID {
	subjects, _ := input["subjects"].([]any)
	for _, item := range subjects {
		obj, _ := item.(map[string]any)
		if obj["type"] != subjectType {
			continue
		}
		idText, _ := obj["id"].(string)
		id, err := uuid.Parse(idText)
		if err == nil {
			return id
		}
	}
	return uuid.Nil
}
