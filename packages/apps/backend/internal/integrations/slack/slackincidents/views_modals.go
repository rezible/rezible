package slackincidents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/execution"

	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/slack-go/slack"
)

const (
	viewCallbackIdAnnotationModal        = "annotation_modal"
	viewCallbackIdIncidentDetailsModal   = "incident_details_modal"
	viewCallbackIdIncidentMilestoneModal = "incident_milestone_modal"
	viewCallbackIdUserHome               = "user_home"
)

func makeUserHomeView(ctx context.Context) (*slack.HomeTabViewRequest, error) {
	var blocks []slack.Block
	blocks = append(blocks, slack.NewSectionBlock(slackintegration.PlainTextBlock("Home Tab"), nil, nil))
	userId, userOk := execution.GetContext(ctx).UserID()
	if !userOk {
		return nil, fmt.Errorf("no user context")
	}
	homeView := slack.HomeTabViewRequest{
		Type:            slack.VTHomeTab,
		CallbackID:      viewCallbackIdUserHome,
		PrivateMetadata: "foo",
		Blocks:          slack.Blocks{BlockSet: blocks},
		ExternalID:      userId.String(),
	}
	return &homeView, nil
}

type incidentDetailsModalViewMetadata struct {
	UserId           string    `json:"uid"`
	CommandChannelId string    `json:"cid"`
	IncidentId       uuid.UUID `json:"iid,omitempty"`
}

func (i *Integration) makeIncidentDetailsModalView(ctx context.Context, prefs incidentPreferences, meta *incidentDetailsModalViewMetadata) (*slack.ModalViewRequest, error) {
	var curr *ent.Incident
	if meta.IncidentId != uuid.Nil {
		inc, incErr := i.incidents.Get(ctx, incident.ID(meta.IncidentId))
		if incErr != nil && !ent.IsNotFound(incErr) {
			return nil, incErr
		}
		curr = inc
	}

	incMeta, incMetaErr := i.incidents.GetIncidentMetadata(ctx)
	if incMetaErr != nil {
		return nil, fmt.Errorf("failed to get incident metadata: %w", incMetaErr)
	}

	builder := newIncidentModalViewBuilder(curr, meta, prefs)
	blockSet := builder.build(incMeta)

	jsonMetadata, jsonErr := json.Marshal(meta)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", jsonErr)
	}

	titleText := "Open Incident"
	submitText := "Submit"
	if curr != nil {
		titleText = "Update Incident"
		submitText = "Update"
	}

	view := &slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      viewCallbackIdIncidentDetailsModal,
		Title:           slackintegration.PlainTextBlock(titleText),
		Submit:          slackintegration.PlainTextBlock(submitText),
		Close:           slackintegration.PlainTextBlock("Cancel"),
		PrivateMetadata: string(jsonMetadata),
		Blocks:          blockSet,
	}

	return view, nil
}

type incidentMilestoneModalViewMetadata struct {
	UserId     string    `json:"uid"`
	IncidentId uuid.UUID `json:"iid"`
}

func (i *Integration) makeIncidentMilestoneModalView(ctx context.Context, meta *incidentMilestoneModalViewMetadata) (*slack.ModalViewRequest, error) {
	inc, incErr := i.incidents.Get(ctx, incident.ID(meta.IncidentId))
	if incErr != nil && !ent.IsNotFound(incErr) {
		return nil, incErr
	}

	builder := newIncidentMilestoneModalViewBuilder(inc, meta)
	blockSet := builder.build()

	jsonMetadata, jsonErr := json.Marshal(meta)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", jsonErr)
	}

	view := &slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      viewCallbackIdIncidentMilestoneModal,
		Title:           slackintegration.PlainTextBlock("Update Incident Status"),
		Submit:          slackintegration.PlainTextBlock("Save"),
		Close:           slackintegration.PlainTextBlock("Cancel"),
		PrivateMetadata: string(jsonMetadata),
		Blocks:          blockSet,
	}

	return view, nil
}
