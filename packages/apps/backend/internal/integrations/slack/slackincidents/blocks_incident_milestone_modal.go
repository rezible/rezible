package slackincidents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incident"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/internal/integrations/slack"
)

const viewCallbackIdIncidentMilestoneModal = "incident_milestone_modal"

type incidentMilestoneModalViewMetadata struct {
	UserId     string    `json:"uid"`
	IncidentId uuid.UUID `json:"iid"`
}

func (a *App) makeIncidentMilestoneModalView(ctx context.Context, meta *incidentMilestoneModalViewMetadata) (*slack.ModalViewRequest, error) {
	inc, incErr := a.incidents.Get(ctx, incident.ID(meta.IncidentId))
	if incErr != nil && !ent.IsNotFound(incErr) {
		return nil, incErr
	}

	builder := &incidentMilestoneModalViewBuilder{
		incident: inc,
		metadata: meta,
	}
	blockSet := builder.Build()

	jsonMetadata, jsonErr := json.Marshal(meta)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", jsonErr)
	}

	view := &slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      viewCallbackIdIncidentMilestoneModal,
		Title:           slackintegration.PlainTextBlock("Update Incident Status"),
		Submit:          slackintegration.PlainTextBlock("Save"),
		Close:           slackintegration.PlainTextBlock("Cancel"),
		PrivateMetadata: string(jsonMetadata),
		Blocks:          blockSet,
	}

	return view, nil
}

type incidentMilestoneModalViewBuilder struct {
	blocks   []slack.Block
	incident *ent.Incident
	metadata *incidentMilestoneModalViewMetadata
}

func (b *incidentMilestoneModalViewBuilder) Build() slack.Blocks {
	b.blocks = make([]slack.Block, 0)
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
