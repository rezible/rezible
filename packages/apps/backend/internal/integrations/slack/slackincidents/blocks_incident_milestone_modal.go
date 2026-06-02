package slackincidents

import (
	"log/slog"

	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/internal/integrations/slack"
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
	incidentMilestoneModalKindIds  = slackintegration.BlockActionIds{Block: "incident_milestone_kind", Input: "kind_select"}
	incidentMilestoneModalNotesIds = slackintegration.BlockActionIds{Block: "incident_milestone_notes", Input: "notes_input"}
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
		slack.NewOptionBlockObject(im.KindImpact.String(), slackintegration.PlainTextBlock("Impact"), nil),
		slack.NewOptionBlockObject(im.KindMitigation.String(), slackintegration.PlainTextBlock("Mitigated"), nil),
		slack.NewOptionBlockObject(im.KindResolution.String(), slackintegration.PlainTextBlock("Resolved"), nil),
	}

	kindsSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, incidentMilestoneModalKindIds.Input, kindsOptions...)

	initialOpt := kindsOptions[0]
	kindsSelect.WithInitialOption(initialOpt)

	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentMilestoneModalKindIds.Block, slackintegration.PlainTextBlock("Incident Status"), nil, kindsSelect))
}

func (b *incidentMilestoneModalViewBuilder) makeNotesInput() {
	notesInput := slack.NewPlainTextInputBlockElement(nil, incidentMilestoneModalNotesIds.Input)
	b.blocks = append(b.blocks,
		slack.NewInputBlock(incidentMilestoneModalNotesIds.Block, slackintegration.PlainTextBlock("Notes"), nil, notesInput))
}
