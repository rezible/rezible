package slackincidents

import (
	"github.com/google/uuid"
	goslack "github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
)

type incidentModalViewBuilder struct {
	blocks   []goslack.Block
	incident *ent.Incident
	metadata *incidentDetailsModalViewMetadata
	prefs    incidentPreferences
}

func newIncidentModalViewBuilder(curr *ent.Incident, meta *incidentDetailsModalViewMetadata, prefs incidentPreferences) *incidentModalViewBuilder {
	return &incidentModalViewBuilder{
		blocks:   []goslack.Block{},
		incident: curr,
		metadata: meta,
		prefs:    prefs,
	}
}

func (b *incidentModalViewBuilder) build(im *rez.IncidentMetadata) goslack.Blocks {
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
	return goslack.Blocks{BlockSet: b.blocks}
}

var (
	incidentModalTitleIds    = slack.BlockActionIds{Block: "title", Input: "title_input"}
	incidentModalSeverityIds = slack.BlockActionIds{Block: "incident_severity", Input: "severity_select"}
	incidentModalTypeIds     = slack.BlockActionIds{Block: "incident_type", Input: "type_select"}
	incidentModalTagIds      = slack.BlockActionIds{Block: "incident_tags", Input: "tags_select"}
)

func setIncidentDetailsModalInputMutationFields(m *ent.IncidentMutation, state *goslack.ViewState) {
	m.SetTitle(incidentModalTitleIds.GetStateValue(state))

	if sevId, sevErr := uuid.Parse(incidentModalSeverityIds.GetStateSelectedValue(state)); sevErr == nil {
		m.SetSeverityID(sevId)
	}

	if typeId, typeErr := uuid.Parse(incidentModalTypeIds.GetStateSelectedValue(state)); typeErr == nil {
		m.SetTypeID(typeId)
	}

	if slack.GetViewStateBlockAction(state, incidentModalTagIds) != nil {
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

func incidentModalFieldOptionIds(optId string) slack.BlockActionIds {
	return slack.BlockActionIds{Block: "incident_field_" + optId, Input: "incident_field_select_" + optId}
}

func (b *incidentModalViewBuilder) makeTitleInput() {
	// Title input
	titleInput := goslack.NewPlainTextInputBlockElement(nil, incidentModalTitleIds.Input)
	if b.incident != nil {
		titleInput.WithInitialValue(b.incident.Title)
	}
	b.blocks = append(b.blocks,
		goslack.NewInputBlock(incidentModalTitleIds.Block, slack.PlainTextBlock("Title"), nil, titleInput))
}

func (b *incidentModalViewBuilder) makeOpenMilestoneModalButton() {
	milestoneButtonText := slack.PlainTextBlock("Update Status")
	milestoneButton := goslack.NewButtonBlockElement(actionCallbackIdIncidentMilestoneModalButton, "milestone", milestoneButtonText)
	b.blocks = append(b.blocks, goslack.NewActionBlock("incident_actions", milestoneButton))
}

func (b *incidentModalViewBuilder) makeSeveritySelect(sevs ent.IncidentSeverities) {
	options := make([]*goslack.OptionBlockObject, len(sevs))
	initialOptIdx := 0
	for i, sev := range sevs {
		options[i] = goslack.NewOptionBlockObject(sev.ID.String(), slack.PlainTextBlock(sev.Name), slack.PlainTextBlock(sev.Description))
		if b.incident != nil && b.incident.SeverityID == sev.ID {
			initialOptIdx = i
		}
	}
	severitySelect := goslack.NewOptionsSelectBlockElement(goslack.OptTypeStatic, nil,
		incidentModalSeverityIds.Input, options...)
	if len(options) > 0 {
		severitySelect.WithInitialOption(options[initialOptIdx])
	}
	b.blocks = append(b.blocks,
		goslack.NewInputBlock(incidentModalSeverityIds.Block, slack.PlainTextBlock("Severity"), nil, severitySelect))
}

func (b *incidentModalViewBuilder) makeTypeSelect(types ent.IncidentTypes) {
	options := make([]*goslack.OptionBlockObject, len(types))
	initialOptIdx := 0
	for i, t := range types {
		options[i] = goslack.NewOptionBlockObject(t.ID.String(), slack.PlainTextBlock(t.Name), nil)
		if b.incident != nil && b.incident.TypeID == t.ID {
			initialOptIdx = i
		}
	}
	typeSelect := goslack.NewOptionsSelectBlockElement(goslack.OptTypeStatic, nil,
		incidentModalTypeIds.Input, options...)
	if len(options) > 0 {
		typeSelect.WithInitialOption(options[initialOptIdx])
	}
	b.blocks = append(b.blocks,
		goslack.NewInputBlock(incidentModalTypeIds.Block, slack.PlainTextBlock("Incident Type"), nil, typeSelect))
}

func (b *incidentModalViewBuilder) makeTagsSelect(tags ent.IncidentTags) {
	options := make([]*goslack.OptionBlockObject, len(tags))
	initialOptions := make([]*goslack.OptionBlockObject, 0, len(tags))
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
		options[i] = goslack.NewOptionBlockObject(tag.ID.String(), slack.PlainTextBlock(label), nil)
		if _, ok := selectedTagIds[tag.ID]; ok {
			initialOptions = append(initialOptions, options[i])
		}
	}

	tagSelect := goslack.NewOptionsMultiSelectBlockElement(goslack.OptTypeStatic, slack.PlainTextBlock("Select tags"),
		incidentModalTagIds.Input, options...)
	if len(initialOptions) > 0 {
		tagSelect.WithInitialOptions(initialOptions...)
	}

	b.blocks = append(b.blocks,
		goslack.NewInputBlock(incidentModalTagIds.Block, slack.PlainTextBlock("Tags"), nil, tagSelect).WithOptional(true))
}

func (b *incidentModalViewBuilder) makeCustomFieldSelect(fields ent.IncidentFields) {
	b.blocks = append(b.blocks, goslack.NewDividerBlock())
	for _, field := range fields {
		fieldOptions := make([]*goslack.OptionBlockObject, len(field.Edges.Options))
		initialOptIdx := 0
		for i, opt := range field.Edges.Options {
			fieldOptions[i] = goslack.NewOptionBlockObject(opt.ID.String(), slack.PlainTextBlock(opt.Value), nil)
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
		fieldOptionSelect := goslack.NewOptionsSelectBlockElement(goslack.OptTypeStatic, slack.PlainTextBlock("Select an option"), ids.Input, fieldOptions...)
		if len(fieldOptions) > 0 {
			fieldOptionSelect.WithInitialOption(fieldOptions[initialOptIdx])
		}
		b.blocks = append(b.blocks,
			goslack.NewInputBlock(ids.Block, slack.PlainTextBlock(field.Name), nil, fieldOptionSelect).WithOptional(true))
	}
}
