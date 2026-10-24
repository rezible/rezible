package slackincidents

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/internal/integrations/slack"
)

const viewCallbackIdIncidentMilestoneModal = "incident_milestone_modal"

type incidentMilestoneModalViewMetadata struct {
	IncidentId uuid.UUID `json:"iid"`
}

type statusOption struct {
	label string
	kind  im.Kind
}

// statusOptions are the changes that can be recorded from each response state. A resolved incident has none.
var statusOptions = map[incident.ResponseState][]statusOption{
	incident.ResponseStateStarted: {
		{label: "Still impacted", kind: im.KindImpact},
		{label: "Mitigated", kind: im.KindMitigation},
		{label: "Resolved", kind: im.KindResolution},
	},
	incident.ResponseStateMitigated: {
		{label: "Mitigation failed", kind: im.KindImpact},
		{label: "Resolved", kind: im.KindResolution},
	},
}

var (
	incidentMilestoneModalKindIds  = slackintegration.BlockActionIds{Block: "incident_milestone_kind", Input: "kind_select"}
	incidentMilestoneModalNotesIds = slackintegration.BlockActionIds{Block: "incident_milestone_notes", Input: "notes_input"}
)

// makeIncidentMilestoneModalView builds the status modal for a started or mitigated incident.
func makeIncidentMilestoneModalView(inc *ent.Incident) (*slack.ModalViewRequest, error) {
	options := statusOptions[inc.ResponseState]
	if len(options) == 0 {
		return nil, errIncidentResolved
	}

	radioOptions := make([]*slack.OptionBlockObject, len(options))
	for i, option := range options {
		radioOptions[i] = slack.NewOptionBlockObject(option.kind.String(), slackintegration.PlainTextBlock(option.label), nil)
	}
	kindInput := slack.NewRadioButtonsBlockElement(incidentMilestoneModalKindIds.Input, radioOptions...)
	notesInput := slack.NewPlainTextInputBlockElement(nil, incidentMilestoneModalNotesIds.Input).WithMultiline(true)
	notesBlock := slack.NewInputBlock(incidentMilestoneModalNotesIds.Block, slackintegration.PlainTextBlock("Notes"), nil, notesInput)
	blocks := []slack.Block{
		slack.NewInputBlock(incidentMilestoneModalKindIds.Block, slackintegration.PlainTextBlock("What changed?"), nil, kindInput),
		notesBlock.WithOptional(true),
	}

	jsonMetadata, jsonErr := json.Marshal(incidentMilestoneModalViewMetadata{IncidentId: inc.ID})
	if jsonErr != nil {
		return nil, fmt.Errorf("marshal metadata: %w", jsonErr)
	}
	view := &slack.ModalViewRequest{
		Type:            slack.VTModal,
		CallbackID:      viewCallbackIdIncidentMilestoneModal,
		Title:           slackintegration.PlainTextBlock("Record status"),
		Submit:          slackintegration.PlainTextBlock("Save"),
		Close:           slackintegration.PlainTextBlock("Cancel"),
		PrivateMetadata: string(jsonMetadata),
		Blocks:          slack.Blocks{BlockSet: blocks},
	}
	return view, nil
}

func getIncidentMilestoneModalViewMetadata(view slack.View) (*incidentMilestoneModalViewMetadata, error) {
	var meta incidentMilestoneModalViewMetadata
	if jsonErr := json.Unmarshal([]byte(view.PrivateMetadata), &meta); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal incident milestone modal metadata: %w", jsonErr)
	}
	return &meta, nil
}

func incidentMilestoneModalInput(state *slack.ViewState) rez.AddIncidentMilestoneParams {
	return rez.AddIncidentMilestoneParams{
		Kind:        im.Kind(incidentMilestoneModalKindIds.GetStateSelectedValue(state)),
		Description: incidentMilestoneModalNotesIds.GetStateValue(state),
	}
}
