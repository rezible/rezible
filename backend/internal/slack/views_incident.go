package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
)

const (
	createIncidentModalViewCallbackID = "create_incident_confirm"
)

type (
	incidentViewMetadata struct {
		UserId     string    `json:"uid"`
		ChannelId  string    `json:"cid"`
		IncidentId uuid.UUID `json:"iid,omitempty"`
		incident   *ent.Incident
	}
)

func (s *ChatService) fetchIncidentViewMetadata(ctx context.Context, ic *slack.InteractionCallback) (*incidentViewMetadata, error) {
	var meta incidentViewMetadata
	if ic.View.PrivateMetadata != "" {
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
			return nil, jsonErr
		}
		if meta.IncidentId != uuid.Nil {
			inc, incErr := s.incidents.Get(ctx, meta.IncidentId)
			if incErr != nil {
				return nil, fmt.Errorf("failed to fetch incident: %w", incErr)
			}
			meta.incident = inc
		}
	}
	return &meta, nil
}

func (s *ChatService) makeIncidentModalView(ctx context.Context, meta *incidentViewMetadata) (*slack.ModalViewRequest, error) {
	view := &slack.ModalViewRequest{
		Type:       "modal",
		CallbackID: createIncidentModalViewCallbackID,
		Title:      plainTextBlock("Open an Incident"),
		Submit:     plainTextBlock("Submit"),
		Close:      plainTextBlock("Cancel"),
	}

	blockSet, blocksErr := s.makeIncidentModalViewBlocks(ctx, meta)
	if blocksErr != nil {
		return nil, fmt.Errorf("failed to make incident modal view blocks: %w", blocksErr)
	}
	view.Blocks = slack.Blocks{BlockSet: blockSet}

	if meta.IncidentId != uuid.Nil {
		view.Title = plainTextBlock("Update Incident")
		view.Submit = plainTextBlock("Update")
	}

	jsonMetadata, jsonErr := json.Marshal(meta)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", jsonErr)
	}
	view.PrivateMetadata = string(jsonMetadata)

	return view, nil
}

func (s *ChatService) makeIncidentModalViewBlocks(ctx context.Context, meta *incidentViewMetadata) ([]slack.Block, error) {
	var blockSet []slack.Block

	var curr *ent.Incident
	if meta.IncidentId != uuid.Nil && meta.incident == nil {
		inc, incErr := s.incidents.Get(ctx, meta.IncidentId)
		if incErr != nil && !ent.IsNotFound(incErr) {
			return nil, incErr
		}
		curr = inc
	}

	// Title input
	titleInput := slack.NewPlainTextInputBlockElement(nil, "title_input")
	if curr != nil {
		titleInput.WithInitialValue(curr.Title)
	}
	blockSet = append(blockSet,
		slack.NewInputBlock("title_block", plainTextBlock("Title"), nil, titleInput))

	sevs, sevsErr := s.incidents.ListIncidentSeverities(ctx)
	if sevsErr != nil {
		return nil, fmt.Errorf("failed to list severities: %w", sevsErr)
	}

	if len(sevs) == 0 {
		sevs = append(sevs, &ent.IncidentSeverity{ID: uuid.New(), Name: "test", Description: "foo bar"})
	}

	// Severity dropdown
	severityOptions := make([]*slack.OptionBlockObject, len(sevs))
	for i, sev := range sevs {
		severityOptions[i] = slack.NewOptionBlockObject(sev.ID.String(), plainTextBlock(sev.Name), plainTextBlock(sev.Description))
	}
	severitySelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, "severity_select", severityOptions...)
	initialSeverity := severityOptions[0]
	if curr != nil && curr.SeverityID != uuid.Nil {
		// Set initial option based on incident.Severity
		for _, opt := range severityOptions {
			if opt.Value == curr.SeverityID.String() {
				initialSeverity = opt
				break
			}
		}
	}
	severitySelect.WithInitialOption(initialSeverity)

	blockSet = append(blockSet,
		slack.NewInputBlock("severity_block", plainTextBlock("Severity"), nil, severitySelect))

	return blockSet, nil
}

func getIncidentModalMetadata(view slack.View) (*incidentViewMetadata, error) {
	var meta incidentViewMetadata
	if jsonErr := json.Unmarshal([]byte(view.PrivateMetadata), &meta); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", jsonErr)
	}

	var title string
	var severityId uuid.UUID
	if view.State != nil {
		if titleInput := getViewStateBlockAction(view.State, "title_block", "title_input"); titleInput != nil {
			title = titleInput.Value
		}
		if sevBlock := getViewStateBlockAction(view.State, "severity_block", "severity_select"); sevBlock != nil {
			if sevBlock.SelectedOption.Value != "" {
				if sevId, sevErr := uuid.Parse(sevBlock.SelectedOption.Value); sevErr == nil {
					severityId = sevId
				}
			}
		}
	}

	incident := &ent.Incident{
		ID:         meta.IncidentId,
		Title:      title,
		SeverityID: severityId,
	}
	meta.incident = incident

	return &meta, nil
}

func (s *ChatService) sendIncidentCreatedMessage(ctx context.Context, md *incidentViewMetadata) error {
	inc := md.incident
	sendErr := s.sendMessage(ctx, md.ChannelId,
		slack.MsgOptionText(fmt.Sprintf("Incident created: *%s* #%s", inc.Title, inc.ChatChannelID), false),
		slack.MsgOptionBroadcast())
	if sendErr != nil {
		//log.Error().Interface("meta", meta).Err(sendErr).Msg("failed to send incident created message")
		return sendErr
	}
	return nil
}
