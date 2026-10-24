package slackincidents

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/errs"
)

const (
	textIncidentManagementOff = "Incident management is off for this organization."
	textAgentMissing          = "Add the Rezible Slack agent to this workspace first. `/incident` uses it to know who you are."
)

var (
	errIncidentResolved = errs.New(errs.CodeConflict, "This incident is resolved. Its response can't change.")
	errUnknownUser      = errs.New(errs.CodeForbidden, slackintegration.UnknownUserText)
)

// handleIncidentCommand answers /incident. The shared service has already replied to Slack users Rezible
// hasn't seen.
func (a *App) handleIncidentCommand(ctx context.Context, ii *ent.Integration, cmd *slack.SlashCommand) (*slack.Blocks, error) {
	enabled, enabledErr := a.orgs.IsIncidentManagementEnabled(ctx)
	if enabledErr != nil {
		return nil, fmt.Errorf("check incident management: %w", enabledErr)
	}
	if !enabled {
		return slackintegration.CommandReply(textIncidentManagementOff), nil
	}
	hasAgent, agentErr := a.hasAgent(ctx, ii)
	if agentErr != nil {
		return nil, fmt.Errorf("check slack agent installation: %w", agentErr)
	}
	if !hasAgent {
		return slackintegration.CommandReply(textAgentMissing), nil
	}

	channelInc, channelErr := a.channelIncident(ctx, cmd.ChannelID)
	if channelErr != nil {
		return nil, channelErr
	}

	text := strings.TrimSpace(cmd.Text)
	switch text {
	case "":
		if channelInc != nil {
			return nil, a.openIncidentDetailsModal(ctx, ii, cmd, channelInc.ID)
		}
		return nil, a.openIncidentDetailsModal(ctx, ii, cmd, uuid.Nil)
	case "new":
		return nil, a.openIncidentDetailsModal(ctx, ii, cmd, uuid.Nil)
	case "update":
		if channelInc == nil {
			return slackintegration.CommandReply("Run `/incident update` in an incident's channel."), nil
		}
		return nil, a.openIncidentDetailsModal(ctx, ii, cmd, channelInc.ID)
	case "status":
		if channelInc == nil {
			return slackintegration.CommandReply("Run `/incident status` in an incident's channel."), nil
		}
		view, viewErr := makeIncidentMilestoneModalView(channelInc)
		if errors.Is(viewErr, errIncidentResolved) {
			return slackintegration.CommandReply(errs.PublicMessage(viewErr)), nil
		}
		if viewErr != nil {
			return nil, viewErr
		}
		return nil, a.openModal(ctx, ii, cmd.TriggerID, view)
	case "help":
		return slackintegration.CommandReply(a.helpText()), nil
	}
	unknown := fmt.Sprintf("Unknown command `%s`.\n%s", slackintegration.EscapeText(text), a.helpText())
	return slackintegration.CommandReply(unknown), nil
}

// channelIncident is the incident whose response channel this is, or nil when it is not a response channel.
func (a *App) channelIncident(ctx context.Context, channelID string) (*ent.Incident, error) {
	inc, incErr := a.incidents.GetIncident(ctx, incident.ChatChannelID(channelID))
	if incErr != nil {
		if ent.IsNotFound(incErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("get incident by channel: %w", incErr)
	}
	return inc, nil
}

func (a *App) openIncidentDetailsModal(ctx context.Context, ii *ent.Integration, cmd *slack.SlashCommand, incidentID uuid.UUID) error {
	meta := &incidentDetailsModalViewMetadata{CommandChannelId: cmd.ChannelID, IncidentId: incidentID}
	view, viewErr := a.makeIncidentDetailsModalView(ctx, meta)
	if viewErr != nil {
		return fmt.Errorf("make incident details modal: %w", viewErr)
	}
	return a.openModal(ctx, ii, cmd.TriggerID, view)
}

func (a *App) openModal(ctx context.Context, ii *ent.Integration, triggerID string, view *slack.ModalViewRequest) error {
	cw, cwErr := a.newClientWrapper(ii)
	if cwErr != nil {
		return fmt.Errorf("create client wrapper: %w", cwErr)
	}
	if openErr := cw.OpenModalView(ctx, triggerID, *view); openErr != nil {
		return fmt.Errorf("open modal: %w", openErr)
	}
	return nil
}

func (a *App) helpText() string {
	lines := []string{
		"*Rezible incidents*",
		"`/incident new` · declare an incident",
		"`/incident update` · edit this incident's details (in its channel)",
		"`/incident status` · record that it is still impacted, mitigated or resolved (in its channel)",
		"`/incident help` · show this",
	}
	webLine := "Roles and linked situations are edited in Rezible."
	if incidentsUrl, urlErr := a.cfg.App.GetFrontendUrl("incidents"); urlErr == nil {
		webLine = fmt.Sprintf("Roles and linked situations are edited in Rezible: <%s|Open incidents>", incidentsUrl)
	}
	return strings.Join(append(lines, webLine), "\n")
}
