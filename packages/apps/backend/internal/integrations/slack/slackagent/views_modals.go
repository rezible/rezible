package slackagent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	ea "github.com/rezible/rezible/ent/eventannotation"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/execution"

	"github.com/rezible/rezible/internal/integrations/slack"
	goslack "github.com/slack-go/slack"
)

const (
	viewCallbackIdAnnotationModal = "annotation_modal"
	viewCallbackIdUserHome        = "user_home"
)

func makeUserHomeView(ctx context.Context) (*goslack.HomeTabViewRequest, error) {
	var blocks []goslack.Block
	blocks = append(blocks, goslack.NewSectionBlock(slack.PlainTextBlock("Home Tab"), nil, nil))
	userId, userOk := execution.GetContext(ctx).UserID()
	if !userOk {
		return nil, fmt.Errorf("no user context")
	}
	homeView := goslack.HomeTabViewRequest{
		Type:            goslack.VTHomeTab,
		CallbackID:      viewCallbackIdUserHome,
		PrivateMetadata: "foo",
		Blocks:          goslack.Blocks{BlockSet: blocks},
		ExternalID:      userId.String(),
	}
	return &homeView, nil
}

type annotationModalMetadata struct {
	UserId       string          `json:"uid"`
	MsgId        slack.MessageId `json:"mid"`
	MsgText      string          `json:"mtx"`
	AnnotationId uuid.UUID       `json:"aid,omitempty"`
}

func (i *Integration) makeAnnotationModalView(ctx context.Context, meta *annotationModalMetadata) (*goslack.ModalViewRequest, error) {
	userId, userOk := execution.GetContext(ctx).UserID()
	if !userOk {
		return nil, fmt.Errorf("no user context")
	}

	lookupAnno := ea.And(
		ea.HasEventWith(ne.ProviderSubjectRef(meta.MsgId.String())),
		ea.CreatorID(userId))

	curr, currErr := i.eventAnnos.Lookup(ctx, lookupAnno)
	if currErr != nil && !ent.IsNotFound(currErr) {
		return nil, fmt.Errorf("failed to lookup existing event annotation: %w", currErr)
	}
	if curr != nil {
		meta.AnnotationId = curr.ID
	}

	builder := newAnnotationModalBuilder(curr, meta)
	blockSet := builder.build()

	titleText := "Create Annotation"
	submitText := "Create"

	if curr != nil {
		titleText = "Update Annotation"
		submitText = "Update"
	}

	jsonMetadata, jsonErr := json.Marshal(meta)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", jsonErr)
	}

	return &goslack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      viewCallbackIdAnnotationModal,
		Title:           slack.PlainTextBlock(titleText),
		Submit:          slack.PlainTextBlock(submitText),
		Close:           slack.PlainTextBlock("Cancel"),
		Blocks:          blockSet,
		PrivateMetadata: string(jsonMetadata),
	}, nil
}

func (i *Integration) getAnnotationModalAnnotation(ctx context.Context, view goslack.View) (*ent.EventAnnotation, error) {
	var meta annotationModalMetadata
	if jsonErr := json.Unmarshal([]byte(view.PrivateMetadata), &meta); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", jsonErr)
	}

	userId, userOk := execution.GetContext(ctx).UserID()
	if !userOk {
		return nil, fmt.Errorf("no user context")
	}

	var notes string
	if view.State != nil {
		if notesInput, inputOk := view.State.Values["notes_input"]; inputOk {
			if noteBlock, noteOk := notesInput["notes_input_text"]; noteOk {
				notes = noteBlock.Value
			}
		}
	}

	// TODO: convert this from event processor?
	ev := &ent.NormalizedEvent{
		ProviderSubjectRef: meta.MsgId.String(),
		//Kind:        "message",
		//Timestamp:   meta.MsgId.getTimestamp(),
		//Source:      "slack",
		//Title:       "Slack Message",
		//Description: meta.MsgText,
	}

	anno := &ent.EventAnnotation{
		ID:        meta.AnnotationId,
		CreatorID: userId,
		Notes:     notes,
		Edges: ent.EventAnnotationEdges{
			Event: ev,
		},
	}

	return anno, nil
}
