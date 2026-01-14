package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
)

const createIncidentModalViewCallbackID = "incident_modal_submit"

type incidentModalViewMetadata struct {
	UserId           string    `json:"uid"`
	CommandChannelId string    `json:"cid"`
	IncidentId       uuid.UUID `json:"iid,omitempty"`
}

func (s *ChatService) makeIncidentModalView(ctx context.Context, meta *incidentModalViewMetadata) (*slack.ModalViewRequest, error) {
	var curr *ent.Incident
	if meta.IncidentId != uuid.Nil {
		inc, incErr := s.incidents.Get(ctx, meta.IncidentId)
		if incErr != nil && !ent.IsNotFound(incErr) {
			return nil, incErr
		}
		curr = inc
	}

	incMeta, incMetaErr := s.incidents.GetIncidentMetadata(ctx)
	if incMetaErr != nil {
		return nil, fmt.Errorf("failed to get incident metadata: %w", incMetaErr)
	}

	builder := newIncidentModalViewBuilder(curr, meta)
	builder.build(incMeta)
	blockSet := builder.blockSet()

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
		CallbackID:      createIncidentModalViewCallbackID,
		Title:           plainTextBlock(titleText),
		Submit:          plainTextBlock(submitText),
		Close:           plainTextBlock("Cancel"),
		PrivateMetadata: string(jsonMetadata),
		Blocks:          blockSet,
	}

	return view, nil
}
