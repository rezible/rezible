package slack

import (
	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

type incidentMilestoneModalViewBuilder struct {
	blocks   []slack.Block
	incident *ent.Incident
	metadata *incidentModalViewMetadata
}

func newIncidentMilestoneModalViewBuilder(curr *ent.Incident, meta *incidentModalViewMetadata) *incidentMilestoneModalViewBuilder {
	return &incidentMilestoneModalViewBuilder{
		blocks:   []slack.Block{},
		incident: curr,
		metadata: meta,
	}
}

func (b *incidentMilestoneModalViewBuilder) build() slack.Blocks {
	b.makeMilestoneSelect()
	b.makeNotesInput()
	return slack.Blocks{BlockSet: b.blocks}
}

var (
	incidentMilestoneModalKindIds  = blockActionIds{Block: "incident_milestone_kind", Input: "kind_select"}
	incidentMilestoneModalNotesIds = blockActionIds{Block: "incident_milestone_notes", Input: "notes_input"}
)

func setIncidentMilestoneModalInputMutationFields(m *ent.IncidentMilestoneMutation, state *slack.ViewState) {
	m.SetDescription(incidentMilestoneModalNotesIds.GetStateValue(state))

	kindOpt := incidentMilestoneModalKindIds.GetStateSelectedValue(state)
	kind := im.Kind(kindOpt)
	if kindErr := im.KindValidator(kind); kindErr != nil {
		log.Debug().Str("kindOpt", kindOpt).Msg("invalid kind")
	} else {
		m.SetKind(kind)
	}
}

func (b *incidentMilestoneModalViewBuilder) makeMilestoneSelect() {
	kindsOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject(im.KindImpact.String(), plainTextBlock("Impact"), nil),
		slack.NewOptionBlockObject(im.KindMitigation.String(), plainTextBlock("Mitigated"), nil),
		slack.NewOptionBlockObject(im.KindResolution.String(), plainTextBlock("Resolved"), nil),
	}

	kindsSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, incidentMilestoneModalKindIds.Input, kindsOptions...)

	initialOpt := kindsOptions[0]
	if b.incident != nil {
		for _, opt := range kindsOptions {
			if opt.Value == b.incident.TypeID.String() {
				initialOpt = opt
				break
			}
		}
	}
	kindsSelect.WithInitialOption(initialOpt)

	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentMilestoneModalKindIds.Block, plainTextBlock("Incident Status"), nil, kindsSelect))
}

func (b *incidentMilestoneModalViewBuilder) makeNotesInput() {
	notesInput := slack.NewPlainTextInputBlockElement(nil, incidentMilestoneModalNotesIds.Input)
	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentMilestoneModalNotesIds.Block, plainTextBlock("Notes"), nil, notesInput))
}
