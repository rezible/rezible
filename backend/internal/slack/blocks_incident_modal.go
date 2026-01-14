package slack

import (
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

type incidentModalViewBuilder struct {
	blocks   []slack.Block
	incident *ent.Incident
	metadata *incidentModalViewMetadata
}

func newIncidentModalViewBuilder(curr *ent.Incident, meta *incidentModalViewMetadata) *incidentModalViewBuilder {
	return &incidentModalViewBuilder{
		blocks:   []slack.Block{},
		incident: curr,
		metadata: meta,
	}
}

func (b *incidentModalViewBuilder) build(im *rez.IncidentMetadata) {
	b.makeTitleInput()
	b.makeSeveritySelect(im.Severities)
	b.makeTypeSelect(im.Types)
	b.makeCustomFieldSelect(im.Fields)
}

func (b *incidentModalViewBuilder) blockSet() slack.Blocks {
	return slack.Blocks{BlockSet: b.blocks}
}

var (
	incidentModalTitleIds    = blockActionIds{Block: "title", Input: "title_input"}
	incidentModalSeverityIds = blockActionIds{Block: "incident_severity", Input: "severity_select"}
	incidentModalTypeIds     = blockActionIds{Block: "incident_type", Input: "type_select"}
)

func incidentModalFieldOptionIds(optId string) blockActionIds {
	return blockActionIds{Block: "incident_field_" + optId, Input: "incident_field_select_" + optId}
}

func (b *incidentModalViewBuilder) makeTitleInput() {
	// Title input
	titleInput := slack.NewPlainTextInputBlockElement(nil, incidentModalTitleIds.Input)
	if b.incident != nil {
		titleInput.WithInitialValue(b.incident.Title)
		log.Debug().Str("curr.Title", b.incident.Title).Msg("set initial title")
	}
	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentModalTitleIds.Block, plainTextBlock("Title"), nil, titleInput))

}

func (b *incidentModalViewBuilder) makeSeveritySelect(sevs ent.IncidentSeverities) {
	severityOptions := make([]*slack.OptionBlockObject, len(sevs))
	for i, sev := range sevs {
		severityOptions[i] = slack.NewOptionBlockObject(sev.ID.String(),
			plainTextBlock(sev.Name), plainTextBlock(sev.Description))
	}
	severitySelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, incidentModalSeverityIds.Input, severityOptions...)
	initialSeverity := severityOptions[0]
	if b.incident != nil && b.incident.SeverityID != uuid.Nil {
		// Set initial option based on incident.Severity
		for _, opt := range severityOptions {
			if opt.Value == b.incident.SeverityID.String() {
				initialSeverity = opt
				log.Debug().Str("opt.Value", opt.Value).Msg("set initial severity")
				break
			}
		}
	}
	severitySelect.WithInitialOption(initialSeverity)
	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentModalSeverityIds.Block, plainTextBlock("Severity"), nil, severitySelect))

}

func (b *incidentModalViewBuilder) makeTypeSelect(types ent.IncidentTypes) {
	// only allow setting incident type on creation
	if b.incident != nil || len(types) == 0 {
		return
	}
	typeOptions := make([]*slack.OptionBlockObject, len(types))
	for i, t := range types {
		typeOptions[i] = slack.NewOptionBlockObject(t.ID.String(), plainTextBlock(t.Name), nil)
	}
	typeSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, incidentModalTypeIds.Input, typeOptions...)

	initialType := typeOptions[0]
	if b.incident != nil && b.incident.TypeID != uuid.Nil {
		for _, opt := range typeOptions {
			if opt.Value == b.incident.TypeID.String() {
				initialType = opt
				break
			}
		}
	}
	typeSelect.WithInitialOption(initialType)

	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentModalTypeIds.Block, plainTextBlock("Incident Type"), nil, typeSelect))
}

func (b *incidentModalViewBuilder) makeCustomFieldSelect(fields ent.IncidentFields) {
	if len(fields) == 0 {
		return
	}
	b.blocks = append(b.blocks, slack.NewDividerBlock())
	for _, field := range fields {
		fieldOptions := make([]*slack.OptionBlockObject, len(field.Edges.Options))
		for i, opt := range field.Edges.Options {
			fieldOptions[i] = slack.NewOptionBlockObject(opt.ID.String(), plainTextBlock(opt.Value), nil)
		}
		initialOption := fieldOptions[0]
		if b.incident != nil && len(b.incident.Edges.FieldSelections) > 0 {
			for _, s := range b.incident.Edges.FieldSelections {
				if s.IncidentFieldID == field.ID {
					for _, opt := range fieldOptions {
						if opt.Value == s.Value {
							initialOption = opt
							break
						}
					}
					break
				}
			}
		}
		ids := incidentModalFieldOptionIds(field.ID.String())
		fieldOptionSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, ids.Input, fieldOptions...)
		fieldOptionSelect.WithInitialOption(initialOption)
		b.blocks = append(b.blocks,
			slack.NewInputBlock(ids.Block, plainTextBlock(field.Name), nil, fieldOptionSelect))
	}
}
