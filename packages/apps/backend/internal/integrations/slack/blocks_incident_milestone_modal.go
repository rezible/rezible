package slack

import (
	"log/slog"

	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
)

type incidentMilestoneModalViewBuilder struct {
	blocks   []slack.Block
	incident *ent.Incident
	metadata *incidentMilestoneModalViewMetadata
}

func newIncidentMilestoneModalViewBuilder(curr *ent.Incident, meta *incidentMilestoneModalViewMetadata) *incidentMilestoneModalViewBuilder {
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
		slog.Debug("invalid kind", "kindOpt", kindOpt)
	} else {
		m.SetKind(kind)
	}
}

func (b *incidentMilestoneModalViewBuilder) makeMilestoneSelect() {
	kindsOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject(im.KindImpact.String(), plainText("Impact"), nil),
		slack.NewOptionBlockObject(im.KindMitigation.String(), plainText("Mitigated"), nil),
		slack.NewOptionBlockObject(im.KindResolution.String(), plainText("Resolved"), nil),
	}

	kindsSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, incidentMilestoneModalKindIds.Input, kindsOptions...)

	initialOpt := kindsOptions[0]
	kindsSelect.WithInitialOption(initialOpt)

	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentMilestoneModalKindIds.Block, plainText("Incident Status"), nil, kindsSelect))
}

func (b *incidentMilestoneModalViewBuilder) makeNotesInput() {
	notesInput := slack.NewPlainTextInputBlockElement(nil, incidentMilestoneModalNotesIds.Input)
	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentMilestoneModalNotesIds.Block, plainText("Notes"), nil, notesInput))
}
