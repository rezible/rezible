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
	goslack "github.com/slack-go/slack"
)

const (
	viewCallbackIdAnnotationModal        = "annotation_modal"
	viewCallbackIdIncidentDetailsModal   = "incident_details_modal"
	viewCallbackIdIncidentMilestoneModal = "incident_milestone_modal"
	viewCallbackIdUserHome               = "user_home"
)

func makeUserHomeView(ctx context.Context) (*goslack.HomeTabViewRequest, error) {
	var blocks []goslack.Block
	blocks = append(blocks, goslack.NewSectionBlock(slack.PlainTextBlock("Home Tab"), nil, nil))
	userId, userOk := execution.GetContext(ctx).UserID()
	if !userOk {
		return nil, fmt.Errorf("no user context")
	}
	homeView := goslack.HomeTabViewRequest{
		Type:            goslack.VTHomeTab,
		CallbackID:      viewCallbackIdUserHome,
		PrivateMetadata: "foo",
		Blocks:          goslack.Blocks{BlockSet: blocks},
		ExternalID:      userId.String(),
	}
	return &homeView, nil
}

type incidentDetailsModalViewMetadata struct {
	UserId           string    `json:"uid"`
	CommandChannelId string    `json:"cid"`
	IncidentId       uuid.UUID `json:"iid,omitempty"`
}

func (i *Integration) makeIncidentDetailsModalView(ctx context.Context, prefs incidentPreferences, meta *incidentDetailsModalViewMetadata) (*goslack.ModalViewRequest, error) {
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

	view := &goslack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      viewCallbackIdIncidentDetailsModal,
		Title:           slack.PlainTextBlock(titleText),
		Submit:          slack.PlainTextBlock(submitText),
		Close:           slack.PlainTextBlock("Cancel"),
		PrivateMetadata: string(jsonMetadata),
		Blocks:          blockSet,
	}

	return view, nil
}

type incidentMilestoneModalViewMetadata struct {
	UserId     string    `json:"uid"`
	IncidentId uuid.UUID `json:"iid"`
}

func (i *Integration) makeIncidentMilestoneModalView(ctx context.Context, meta *incidentMilestoneModalViewMetadata) (*goslack.ModalViewRequest, error) {
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

	view := &goslack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      viewCallbackIdIncidentMilestoneModal,
		Title:           slack.PlainTextBlock("Update Incident Status"),
		Submit:          slack.PlainTextBlock("Save"),
		Close:           slack.PlainTextBlock("Cancel"),
		PrivateMetadata: string(jsonMetadata),
		Blocks:          blockSet,
	}

	return view, nil
}
