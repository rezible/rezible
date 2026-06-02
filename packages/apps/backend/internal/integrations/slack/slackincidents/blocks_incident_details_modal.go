package slackincidents

import (
	"github.com/google/uuid"

	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
)

type incidentModalViewBuilder struct {
	blocks   []slack.Block
	incident *ent.Incident
	metadata *incidentDetailsModalViewMetadata
	prefs    UserSettingsIncidents
}

func newIncidentModalViewBuilder(curr *ent.Incident, meta *incidentDetailsModalViewMetadata, prefs UserSettingsIncidents) *incidentModalViewBuilder {
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
	if len(im.Types) > 0 {
		b.makeTypeSelect(im.Types)
	}
	if len(im.Tags) > 0 {
		b.makeTagsSelect(im.Tags)
	}
	if len(im.Fields) > 0 {
		b.makeCustomFieldSelect(im.Fields)
	}
	return slack.Blocks{BlockSet: b.blocks}
}

var (
	incidentModalTitleIds    = slackintegration.BlockActionIds{Block: "title", Input: "title_input"}
	incidentModalSeverityIds = slackintegration.BlockActionIds{Block: "incident_severity", Input: "severity_select"}
	incidentModalTypeIds     = slackintegration.BlockActionIds{Block: "incident_type", Input: "type_select"}
	incidentModalTagIds      = slackintegration.BlockActionIds{Block: "incident_tags", Input: "tags_select"}
)

func setIncidentDetailsModalInputMutationFields(m *ent.IncidentMutation, state *slack.ViewState) {
	m.SetTitle(incidentModalTitleIds.GetStateValue(state))

	if sevId, sevErr := uuid.Parse(incidentModalSeverityIds.GetStateSelectedValue(state)); sevErr == nil {
		m.SetSeverityID(sevId)
	}

	if typeId, typeErr := uuid.Parse(incidentModalTypeIds.GetStateSelectedValue(state)); typeErr == nil {
		m.SetTypeID(typeId)
	}

	if slackintegration.GetViewStateBlockAction(state, incidentModalTagIds) != nil {
		m.ClearTagAssignments()
		for _, selectedTagId := range incidentModalTagIds.GetStateSelectedValues(state) {
			if tagId, tagErr := uuid.Parse(selectedTagId); tagErr == nil {
				m.AddTagAssignmentIDs(tagId)
			}
		}
	}

	hasFieldInputs := false
	for _, actions := range state.Values {
		for actionID, blockAction := range actions {
			if len(actionID) < len("incident_field_select_") || actionID[:len("incident_field_select_")] != "incident_field_select_" {
				continue
			}
			if !hasFieldInputs {
				m.ClearFieldSelections()
				hasFieldInputs = true
			}
			if fieldOptionID, fieldOptionErr := uuid.Parse(blockAction.SelectedOption.Value); fieldOptionErr == nil {
				m.AddFieldSelectionIDs(fieldOptionID)
			}
		}
	}
}

func incidentModalFieldOptionIds(optId string) slackintegration.BlockActionIds {
	return slackintegration.BlockActionIds{Block: "incident_field_" + optId, Input: "incident_field_select_" + optId}
}

func (b *incidentModalViewBuilder) makeTitleInput() {
	// Title input
	titleInput := slack.NewPlainTextInputBlockElement(nil, incidentModalTitleIds.Input)
	if b.incident != nil {
		titleInput.WithInitialValue(b.incident.Title)
	}
	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentModalTitleIds.Block, slackintegration.PlainTextBlock("Title"), nil, titleInput))
}

func (b *incidentModalViewBuilder) makeOpenMilestoneModalButton() {
	milestoneButtonText := slackintegration.PlainTextBlock("Update Status")
	milestoneButton := slack.NewButtonBlockElement(actionCallbackIdIncidentMilestoneModalButton, "milestone", milestoneButtonText)
	b.blocks = append(b.blocks, slack.NewActionBlock("incident_actions", milestoneButton))
}

func (b *incidentModalViewBuilder) makeSeveritySelect(sevs ent.IncidentSeverities) {
	options := make([]*slack.OptionBlockObject, len(sevs))
	initialOptIdx := 0
	for i, sev := range sevs {
		options[i] = slack.NewOptionBlockObject(sev.ID.String(), slackintegration.PlainTextBlock(sev.Name), slackintegration.PlainTextBlock(sev.Description))
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
		slack.NewInputBlock(incidentModalSeverityIds.Block, slackintegration.PlainTextBlock("Severity"), nil, severitySelect))
}

func (b *incidentModalViewBuilder) makeTypeSelect(types ent.IncidentTypes) {
	options := make([]*slack.OptionBlockObject, len(types))
	initialOptIdx := 0
	for i, t := range types {
		options[i] = slack.NewOptionBlockObject(t.ID.String(), slackintegration.PlainTextBlock(t.Name), nil)
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
		slack.NewInputBlock(incidentModalTypeIds.Block, slackintegration.PlainTextBlock("Incident Type"), nil, typeSelect))
}

func (b *incidentModalViewBuilder) makeTagsSelect(tags ent.IncidentTags) {
	options := make([]*slack.OptionBlockObject, len(tags))
	initialOptions := make([]*slack.OptionBlockObject, 0, len(tags))
	selectedTagIds := map[uuid.UUID]struct{}{}
	if b.incident != nil {
		for _, tag := range b.incident.Edges.TagAssignments {
			selectedTagIds[tag.ID] = struct{}{}
		}
	}
	for i, tag := range tags {
		label := tag.Value
		if tag.Key != "" {
			label = tag.Key + ": " + tag.Value
		}
		options[i] = slack.NewOptionBlockObject(tag.ID.String(), slackintegration.PlainTextBlock(label), nil)
		if _, ok := selectedTagIds[tag.ID]; ok {
			initialOptions = append(initialOptions, options[i])
		}
	}

	tagSelect := slack.NewOptionsMultiSelectBlockElement(slack.OptTypeStatic, slackintegration.PlainTextBlock("Select tags"),
		incidentModalTagIds.Input, options...)
	if len(initialOptions) > 0 {
		tagSelect.WithInitialOptions(initialOptions...)
	}

	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentModalTagIds.Block, slackintegration.PlainTextBlock("Tags"), nil, tagSelect).WithOptional(true))
}

func (b *incidentModalViewBuilder) makeCustomFieldSelect(fields ent.IncidentFields) {
	b.blocks = append(b.blocks, slack.NewDividerBlock())
	for _, field := range fields {
		fieldOptions := make([]*slack.OptionBlockObject, len(field.Edges.Options))
		initialOptIdx := 0
		for i, opt := range field.Edges.Options {
			fieldOptions[i] = slack.NewOptionBlockObject(opt.ID.String(), slackintegration.PlainTextBlock(opt.Value), nil)
			if b.incident == nil || len(b.incident.Edges.FieldSelections) == 0 {
				continue
			}
			for _, s := range b.incident.Edges.FieldSelections {
				if s.IncidentFieldID != field.ID {
					continue
				}
				if s.ID == opt.ID {
					initialOptIdx = i
					break
				}
			}
		}
		ids := incidentModalFieldOptionIds(field.ID.String())
		fieldOptionSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, slackintegration.PlainTextBlock("Select an option"), ids.Input, fieldOptions...)
		if len(fieldOptions) > 0 {
			fieldOptionSelect.WithInitialOption(fieldOptions[initialOptIdx])
		}
		b.blocks = append(b.blocks,
			slack.NewInputBlock(ids.Block, slackintegration.PlainTextBlock(field.Name), nil, fieldOptionSelect).WithOptional(true))
	}
}
