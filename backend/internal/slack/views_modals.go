package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"
)

const (
	viewCallbackIdAnnotationModal        = "annotation_modal"
	viewCallbackIdIncidentDetailsModal   = "incident_details_modal"
	viewCallbackIdIncidentMilestoneModal = "incident_milestone_modal"
	viewCallbackIdUserHome               = "user_home"
)

func makeUserHomeView(ctx context.Context, user *ent.User) (*slack.HomeTabViewRequest, error) {
	var blocks []slack.Block
	blocks = append(blocks, slack.NewSectionBlock(plainTextBlock("Home Tab"), nil, nil))
	homeView := slack.HomeTabViewRequest{
		Type:            slack.VTHomeTab,
		CallbackID:      viewCallbackIdUserHome,
		PrivateMetadata: "foo",
		Blocks:          slack.Blocks{BlockSet: blocks},
		ExternalID:      user.ID.String(),
	}
	return &homeView, nil
}

type annotationModalMetadata struct {
	UserId       string    `json:"uid"`
	MsgId        messageId `json:"mid"`
	MsgText      string    `json:"mtx"`
	AnnotationId uuid.UUID `json:"aid,omitempty"`
}

func (s *ChatService) makeAnnotationModalView(ctx context.Context, meta *annotationModalMetadata) (*slack.ModalViewRequest, error) {
	usr, usrCtx, userErr := s.lookupChatUser(ctx, meta.UserId)
	if userErr != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", userErr)
	}

	ev := &ent.Event{ExternalID: meta.MsgId.String()}

	curr, currErr := s.annos.LookupByUserEvent(usrCtx, usr.ID, ev)
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

	return &slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      viewCallbackIdAnnotationModal,
		Title:           plainTextBlock(titleText),
		Submit:          plainTextBlock(submitText),
		Close:           plainTextBlock("Cancel"),
		Blocks:          blockSet,
		PrivateMetadata: string(jsonMetadata),
	}, nil
}

func (s *ChatService) getAnnotationModalAnnotation(ctx context.Context, view slack.View) (*ent.EventAnnotation, error) {
	var meta annotationModalMetadata
	if jsonErr := json.Unmarshal([]byte(view.PrivateMetadata), &meta); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", jsonErr)
	}

	usr, _, userErr := s.lookupChatUser(ctx, meta.UserId)
	if userErr != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", userErr)
	}

	var notes string
	if view.State != nil {
		if notesInput, inputOk := view.State.Values["notes_input"]; inputOk {
			if noteBlock, noteOk := notesInput["notes_input_text"]; noteOk {
				notes = noteBlock.Value
			}
		}
	}

	ev := &ent.Event{
		ExternalID:  meta.MsgId.String(),
		Kind:        "message",
		Timestamp:   meta.MsgId.getTimestamp(),
		Source:      "slack",
		Title:       "Slack Message",
		Description: meta.MsgText,
	}

	anno := &ent.EventAnnotation{
		ID:        meta.AnnotationId,
		CreatorID: usr.ID,
		Notes:     notes,
		Edges: ent.EventAnnotationEdges{
			Event: ev,
		},
	}

	return anno, nil
}

type incidentDetailsModalViewMetadata struct {
	UserId           string    `json:"uid"`
	CommandChannelId string    `json:"cid"`
	IncidentId       uuid.UUID `json:"iid,omitempty"`
}

func (s *ChatService) makeIncidentDetailsModalView(ctx context.Context, meta *incidentDetailsModalViewMetadata) (*slack.ModalViewRequest, error) {
	var curr *ent.Incident
	if meta.IncidentId != uuid.Nil {
		inc, incErr := s.incidents.Get(ctx, meta.IncidentId)
		if incErr != nil && !ent.IsNotFound(incErr) {
			return nil, incErr
		}
		curr = inc
	}

	incMeta, incMetaErr := s.incidents.GetIncidentMetadata(ctx)
	if incMetaErr != nil {
		return nil, fmt.Errorf("failed to get incident metadata: %w", incMetaErr)
	}

	builder := newIncidentModalViewBuilder(curr, meta)
	blockSet := builder.build(incMeta)

	jsonMetadata, jsonErr := json.Marshal(meta)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", jsonErr)
	}

	titleText := "Open Incident"
	submitText := "Submit"
	if curr != nil {
		titleText = "Update Incident"
		submitText = "Update"
	}

	view := &slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      viewCallbackIdIncidentDetailsModal,
		Title:           plainTextBlock(titleText),
		Submit:          plainTextBlock(submitText),
		Close:           plainTextBlock("Cancel"),
		PrivateMetadata: string(jsonMetadata),
		Blocks:          blockSet,
	}

	return view, nil
}

type incidentMilestoneModalViewMetadata struct {
	UserId     string    `json:"uid"`
	IncidentId uuid.UUID `json:"iid"`
}

func (s *ChatService) makeIncidentMilestoneModalView(ctx context.Context, meta *incidentMilestoneModalViewMetadata) (*slack.ModalViewRequest, error) {
	inc, incErr := s.incidents.Get(ctx, meta.IncidentId)
	if incErr != nil && !ent.IsNotFound(incErr) {
		return nil, incErr
	}

	builder := newIncidentMilestoneModalViewBuilder(inc, meta)
	blockSet := builder.build()

	jsonMetadata, jsonErr := json.Marshal(meta)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", jsonErr)
	}

	view := &slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      viewCallbackIdIncidentMilestoneModal,
		Title:           plainTextBlock("Update Incident Status"),
		Submit:          plainTextBlock("Save"),
		Close:           plainTextBlock("Cancel"),
		PrivateMetadata: string(jsonMetadata),
		Blocks:          blockSet,
	}

	return view, nil
}
