package slack

import (
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"
)

type incidentModalViewBuilder struct {
	blocks   []slack.Block
	incident *ent.Incident
	metadata *incidentDetailsModalViewMetadata
	prefs    incidentPreferences
}

func newIncidentModalViewBuilder(curr *ent.Incident, meta *incidentDetailsModalViewMetadata, prefs incidentPreferences) *incidentModalViewBuilder {
	return &incidentModalViewBuilder{
		blocks:   []slack.Block{},
		incident: curr,
		metadata: meta,
		prefs:    prefs,
	}
}

func (b *incidentModalViewBuilder) build(im *rez.IncidentMetadata) slack.Blocks {
	b.makeTitleInput()
	b.makeSeveritySelect(im.Severities)
	if b.incident != nil {
		b.makeOpenMilestoneModalButton()
	}
	// only allow setting incident type on creation
	if b.incident == nil && len(im.Types) > 0 {
		b.makeTypeSelect(im.Types)
	}
	if len(im.Fields) > 0 {
		b.makeCustomFieldSelect(im.Fields)
	}
	return slack.Blocks{BlockSet: b.blocks}
}

var (
	incidentModalTitleIds    = blockActionIds{Block: "title", Input: "title_input"}
	incidentModalSeverityIds = blockActionIds{Block: "incident_severity", Input: "severity_select"}
	incidentModalTypeIds     = blockActionIds{Block: "incident_type", Input: "type_select"}
)

func setIncidentDetailsModalInputMutationFields(m *ent.IncidentMutation, state *slack.ViewState) {
	m.SetTitle(incidentModalTitleIds.GetStateValue(state))

	if sevId, sevErr := uuid.Parse(incidentModalSeverityIds.GetStateSelectedValue(state)); sevErr == nil {
		m.SetSeverityID(sevId)
	}

	if typeId, typeErr := uuid.Parse(incidentModalTypeIds.GetStateSelectedValue(state)); typeErr == nil {
		m.SetTypeID(typeId)
	}

	m.ClearFieldSelections()
	for _, actions := range state.Values {
		for actionID, blockAction := range actions {
			if len(actionID) < len("incident_field_select_") || actionID[:len("incident_field_select_")] != "incident_field_select_" {
				continue
			}
			if fieldOptionID, fieldOptionErr := uuid.Parse(blockAction.SelectedOption.Value); fieldOptionErr == nil {
				m.AddFieldSelectionIDs(fieldOptionID)
			}
		}
	}
}

func incidentModalFieldOptionIds(optId string) blockActionIds {
	return blockActionIds{Block: "incident_field_" + optId, Input: "incident_field_select_" + optId}
}

func (b *incidentModalViewBuilder) makeTitleInput() {
	// Title input
	titleInput := slack.NewPlainTextInputBlockElement(nil, incidentModalTitleIds.Input)
	if b.incident != nil {
		titleInput.WithInitialValue(b.incident.Title)
	}
	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentModalTitleIds.Block, plainText("Title"), nil, titleInput))
}

func (b *incidentModalViewBuilder) makeOpenMilestoneModalButton() {
	milestoneButtonText := plainText("Update Status")
	milestoneButton := slack.NewButtonBlockElement(actionCallbackIdIncidentMilestoneModalButton, "milestone", milestoneButtonText)
	b.blocks = append(b.blocks, slack.NewActionBlock("incident_actions", milestoneButton))
}

func (b *incidentModalViewBuilder) makeSeveritySelect(sevs ent.IncidentSeverities) {
	options := make([]*slack.OptionBlockObject, len(sevs))
	initialOptIdx := 0
	for i, sev := range sevs {
		options[i] = slack.NewOptionBlockObject(sev.ID.String(), plainText(sev.Name), plainText(sev.Description))
		if b.incident != nil && b.incident.SeverityID == sev.ID {
			initialOptIdx = i
		}
	}
	severitySelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil,
		incidentModalSeverityIds.Input, options...)
	if len(options) > 0 {
		severitySelect.WithInitialOption(options[initialOptIdx])
	}
	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentModalSeverityIds.Block, plainText("Severity"), nil, severitySelect))
}

func (b *incidentModalViewBuilder) makeTypeSelect(types ent.IncidentTypes) {
	options := make([]*slack.OptionBlockObject, len(types))
	initialOptIdx := 0
	for i, t := range types {
		options[i] = slack.NewOptionBlockObject(t.ID.String(), plainText(t.Name), nil)
		if b.incident != nil && b.incident.TypeID == t.ID {
			initialOptIdx = i
		}
	}
	typeSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil,
		incidentModalTypeIds.Input, options...)
	if len(options) > 0 {
		typeSelect.WithInitialOption(options[initialOptIdx])
	}
	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentModalTypeIds.Block, plainText("Incident Type"), nil, typeSelect))
}

func (b *incidentModalViewBuilder) makeCustomFieldSelect(fields ent.IncidentFields) {
	b.blocks = append(b.blocks, slack.NewDividerBlock())
	for _, field := range fields {
		fieldOptions := make([]*slack.OptionBlockObject, len(field.Edges.Options))
		initialOptIdx := 0
		for i, opt := range field.Edges.Options {
			fieldOptions[i] = slack.NewOptionBlockObject(opt.ID.String(), plainText(opt.Value), nil)
			if b.incident == nil || len(b.incident.Edges.FieldSelections) == 0 {
				continue
			}
			for _, s := range b.incident.Edges.FieldSelections {
				if s.IncidentFieldID != field.ID {
					continue
				}
				for _, selected := range fieldOptions {
					if selected.Value == opt.Value {
						initialOptIdx = i
						break
					}
				}
			}
		}
		ids := incidentModalFieldOptionIds(field.ID.String())
		fieldOptionSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, ids.Input, fieldOptions...)
		if len(fieldOptions) > 0 {
			fieldOptionSelect.WithInitialOption(fieldOptions[initialOptIdx])
		}
		b.blocks = append(b.blocks,
			slack.NewInputBlock(ids.Block, plainText(field.Name), nil, fieldOptionSelect))
	}
}
