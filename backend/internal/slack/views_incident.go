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

	types, typesErr := s.incidents.ListIncidentTypes(ctx)
	if typesErr != nil {
		return nil, fmt.Errorf("failed to list incident types: %w", typesErr)
	}
	if len(types) > 0 {
		typeOptions := make([]*slack.OptionBlockObject, len(types))
		for i, t := range types {
			typeOptions[i] = slack.NewOptionBlockObject(t.ID.String(), plainTextBlock(t.Name), nil)
		}
		typeSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, "type_select", typeOptions...)
		if curr != nil && curr.TypeID != uuid.Nil {
			for _, opt := range typeOptions {
				if opt.Value == curr.TypeID.String() {
					typeSelect.WithInitialOption(opt)
					break
				}
			}
		}
		blockSet = append(blockSet,
			slack.NewInputBlock("type_block", plainTextBlock("Type"), nil, typeSelect))
	}

	comps, compsErr := s.components.ListSystemComponents(ctx, rez.ListSystemComponentsParams{})
	if compsErr != nil {
		return nil, fmt.Errorf("failed to list system components: %w", compsErr)
	}
	if comps != nil && len(comps.Data) > 0 {
		compOptions := make([]*slack.OptionBlockObject, len(comps.Data))
		for i, c := range comps.Data {
			compOptions[i] = slack.NewOptionBlockObject(c.ID.String(), plainTextBlock(c.Name), nil)
		}
		compSelect := slack.NewOptionsMultiSelectBlockElement(slack.MultiOptTypeStatic, nil, "components_select", compOptions...)
		// TODO: pre-populate with current components if editing
		blockSet = append(blockSet,
			slack.NewInputBlock("components_block", plainTextBlock("System Components"), nil, compSelect))
	}

	publicOption := slack.NewOptionBlockObject("public", plainTextBlock("Public"), nil)
	privateOption := slack.NewOptionBlockObject("public", plainTextBlock("Public"), nil)
	visibilityRadio := slack.NewRadioButtonsBlockElement("visibility_radio", publicOption, privateOption)
	// TODO: check default privacy
	visibilityRadio.InitialOption = publicOption
	if curr != nil && curr.Private {
		visibilityRadio.InitialOption = privateOption
	}
	blockSet = append(blockSet,
		slack.NewInputBlock("visibility_block", plainTextBlock("Visibility"), nil, visibilityRadio))

	return blockSet, nil
}

func setIncidentFieldsFromModal(view slack.View) func(m *ent.IncidentMutation) {
	return func(m *ent.IncidentMutation) {
		if view.State == nil {
			return
		}

		if titleInput := getViewStateBlockAction(view.State, "title_block", "title_input"); titleInput != nil {
			m.SetTitle(titleInput.Value)
		}

		if sevBlock := getViewStateBlockAction(view.State, "severity_block", "severity_select"); sevBlock != nil {
			if sevBlock.SelectedOption.Value != "" {
				if sevId, sevErr := uuid.Parse(sevBlock.SelectedOption.Value); sevErr == nil {
					//m.SetSeverityID(sevId)
					log.Debug().Str("sevId", sevId.String()).Msg("set incident severity")
				}
			}
		}

		if typeBlock := getViewStateBlockAction(view.State, "type_block", "type_select"); typeBlock != nil {
			if typeBlock.SelectedOption.Value != "" {
				if typeId, typeErr := uuid.Parse(typeBlock.SelectedOption.Value); typeErr == nil {
					//m.SetTypeID(typeId)
					log.Debug().Str("id", typeId.String()).Msg("set incident type")
				}
			}
		}

		// Handle private checkbox
		privateBlock := getViewStateBlockAction(view.State, "visibility_block", "visibility_radio")
		if privateBlock != nil && privateBlock.SelectedOption.Value == "private" {
			m.SetPrivate(true)
		}

		// TODO: Handle system components multi-select
		// This requires linking to IncidentFieldOption or another through table
	}
}
