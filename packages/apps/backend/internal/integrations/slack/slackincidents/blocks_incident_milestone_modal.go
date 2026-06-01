package slackincidents

import (
	"log/slog"

	"github.com/rezible/rezible/internal/integrations/slack"
	slackgo "github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
)

type incidentMilestoneModalViewBuilder struct {
	blocks   []slackgo.Block
	incident *ent.Incident
	metadata *incidentMilestoneModalViewMetadata
}

func newIncidentMilestoneModalViewBuilder(curr *ent.Incident, meta *incidentMilestoneModalViewMetadata) *incidentMilestoneModalViewBuilder {
	return &incidentMilestoneModalViewBuilder{
		blocks:   []slackgo.Block{},
		incident: curr,
		metadata: meta,
	}
}

func (b *incidentMilestoneModalViewBuilder) build() slackgo.Blocks {
	b.makeMilestoneSelect()
	b.makeNotesInput()
	return slackgo.Blocks{BlockSet: b.blocks}
}

var (
	incidentMilestoneModalKindIds  = slack.BlockActionIds{Block: "incident_milestone_kind", Input: "kind_select"}
	incidentMilestoneModalNotesIds = slack.BlockActionIds{Block: "incident_milestone_notes", Input: "notes_input"}
)

func setIncidentMilestoneModalInputMutationFields(m *ent.IncidentMilestoneMutation, state *slackgo.ViewState) {
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
	kindsOptions := []*slackgo.OptionBlockObject{
		slackgo.NewOptionBlockObject(im.KindImpact.String(), slack.PlainTextBlock("Impact"), nil),
		slackgo.NewOptionBlockObject(im.KindMitigation.String(), slack.PlainTextBlock("Mitigated"), nil),
		slackgo.NewOptionBlockObject(im.KindResolution.String(), slack.PlainTextBlock("Resolved"), nil),
	}

	kindsSelect := slackgo.NewOptionsSelectBlockElement(slackgo.OptTypeStatic, nil, incidentMilestoneModalKindIds.Input, kindsOptions...)

	initialOpt := kindsOptions[0]
	kindsSelect.WithInitialOption(initialOpt)

	b.blocks = append(b.blocks,
		slackgo.NewInputBlock(incidentMilestoneModalKindIds.Block, slack.PlainTextBlock("Incident Status"), nil, kindsSelect))
}

func (b *incidentMilestoneModalViewBuilder) makeNotesInput() {
	notesInput := slackgo.NewPlainTextInputBlockElement(nil, incidentMilestoneModalNotesIds.Input)
	b.blocks = append(b.blocks,
		slackgo.NewInputBlock(incidentMilestoneModalNotesIds.Block, slack.PlainTextBlock("Notes"), nil, notesInput))
}
