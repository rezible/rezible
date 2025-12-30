package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
)

const (
	createIncidentModalViewCallbackID = "incident_modal_submit"
)

var (
	incidentModalTitleIds    = blockActionIds{Block: "title", Input: "title_input"}
	incidentModalSeverityIds = blockActionIds{Block: "incident_severity", Input: "severity_select"}
	incidentModalTypeIds     = blockActionIds{Block: "incident_type", Input: "type_select"}
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
	curr := meta.incident
	if curr == nil && meta.IncidentId != uuid.Nil {
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

	blockSet, blocksErr := s.makeIncidentModalViewBlocks(curr, incMeta)
	if blocksErr != nil {
		return nil, fmt.Errorf("failed to make incident modal view blocks: %w", blocksErr)
	}

	view := &slack.ModalViewRequest{
		Type:       "modal",
		CallbackID: createIncidentModalViewCallbackID,
		Title:      plainTextBlock("Open an Incident"),
		Submit:     plainTextBlock("Submit"),
		Close:      plainTextBlock("Cancel"),
	}
	view.Blocks = slack.Blocks{BlockSet: blockSet}

	if meta.incident != nil {
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

func (s *ChatService) makeIncidentModalViewBlocks(curr *ent.Incident, meta *rez.IncidentMetadata) ([]slack.Block, error) {
	var blocks []slack.Block

	// Title input
	titleInput := slack.NewPlainTextInputBlockElement(nil, incidentModalTitleIds.Input)
	if curr != nil {
		titleInput.WithInitialValue(curr.Title)
		log.Debug().Str("curr.Title", curr.Title).Msg("set initial title")
	}
	blocks = append(blocks,
		slack.NewInputBlock(incidentModalTitleIds.Block, plainTextBlock("Title"), nil, titleInput))

	severityOptions := make([]*slack.OptionBlockObject, len(meta.Severities))
	for i, sev := range meta.Severities {
		severityOptions[i] = slack.NewOptionBlockObject(sev.ID.String(), plainTextBlock(sev.Name), plainTextBlock(sev.Description))
	}
	severitySelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, incidentModalSeverityIds.Input, severityOptions...)
	initialSeverity := severityOptions[0]
	if curr != nil && curr.SeverityID != uuid.Nil {
		// Set initial option based on incident.Severity
		for _, opt := range severityOptions {
			if opt.Value == curr.SeverityID.String() {
				initialSeverity = opt
				log.Debug().Str("opt.Value", opt.Value).Msg("set initial severity")
				break
			}
		}
	}
	severitySelect.WithInitialOption(initialSeverity)
	blocks = append(blocks,
		slack.NewInputBlock(incidentModalSeverityIds.Block, plainTextBlock("Severity"), nil, severitySelect))

	typeOptions := make([]*slack.OptionBlockObject, len(meta.Types))
	for i, t := range meta.Types {
		typeOptions[i] = slack.NewOptionBlockObject(t.ID.String(), plainTextBlock(t.Name), nil)
	}
	initialType := typeOptions[0]
	if curr != nil && curr.TypeID != uuid.Nil {
		for _, opt := range typeOptions {
			if opt.Value == curr.TypeID.String() {
				initialType = opt
				log.Debug().Str("opt.Value", opt.Value).Msg("set initial Type")
				break
			}
		}
	}
	typeSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, incidentModalTypeIds.Input, typeOptions...)
	typeSelect.WithInitialOption(initialType)
	blocks = append(blocks,
		slack.NewInputBlock(incidentModalTypeIds.Block, plainTextBlock("Incident Type"), nil, typeSelect))

	return blocks, nil
}

func setIncidentModalStateFields(m *ent.IncidentMutation, state *slack.ViewState) {
	m.SetTitle(incidentModalTitleIds.GetStateValue(state))

	if sevId, sevErr := uuid.Parse(incidentModalSeverityIds.GetStateSelectedValue(state)); sevErr == nil {
		m.SetSeverityID(sevId)
	}

	if typeId, typeErr := uuid.Parse(incidentModalTypeIds.GetStateSelectedValue(state)); typeErr == nil {
		m.SetTypeID(typeId)
	}

	// TODO: Handle system components multi-select
	// This requires linking to IncidentFieldOption or another through table
}
