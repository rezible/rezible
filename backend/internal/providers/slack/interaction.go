package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallusershift"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"net/http"
	"time"
)

var (
	createAnnotationCallbackID        = "create_annotation"
	createAnnotationConfirmCallbackID = "create_annotation_confirm"
)

func (p *ChatProvider) handleInteractionWebhook(w http.ResponseWriter, r *http.Request) {
	if verifyErr := p.verifyWebhook(w, r); verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}

	payload := r.PostFormValue("payload")
	if payload == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var ic slack.InteractionCallback
	if jsonErr := json.Unmarshal([]byte(payload), &ic); jsonErr != nil {
		log.Debug().Err(jsonErr).Msg("failed to unmarshal interaction payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	var handlerErr error
	switch ic.Type {
	case slack.InteractionTypeMessageAction:
		handlerErr = p.handleMessageAction(ctx, &ic)
	case slack.InteractionTypeViewSubmission:
		handlerErr = p.handleViewSubmission(ctx, &ic)
	default:
		log.Warn().Str("type", string(ic.Type)).Msg("unknown interaction type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if handlerErr != nil {
		log.Error().Err(handlerErr).Str("type", string(ic.Type)).Msg("failed to handle interaction")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (p *ChatProvider) handleMessageAction(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.CallbackID {
	case createAnnotationCallbackID:
		return p.handleCreateAnnotationAction(ctx, ic)
	}
	return fmt.Errorf("unknown message action: %s", ic.CallbackID)
}

func (p *ChatProvider) handleCreateAnnotationAction(ctx context.Context, ic *slack.InteractionCallback) error {
	view, viewErr := p.createAnnotationModalView(ctx, ic)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}

	resp, respErr := p.client.OpenViewContext(ctx, ic.TriggerID, *view)
	if resp != nil && !resp.Ok && len(resp.ResponseMetadata.Messages) > 0 {
		log.Debug().
			Strs("messages", resp.ResponseMetadata.Messages).
			Msg("message action: open view failed")
	}

	if respErr != nil {
		return fmt.Errorf("open view: %w", respErr)
	}
	return nil
}

func (p *ChatProvider) createAnnotationModalView(ctx context.Context, ic *slack.InteractionCallback) (*slack.ModalViewRequest, error) {
	user, userErr := p.lookupUser(ctx, ic.User.ID)
	if userErr != nil {
		return nil, fmt.Errorf("failed to get user: %w", userErr)
	}

	shiftIsActive := oncallusershift.And(oncallusershift.EndAtGT(time.Now()), oncallusershift.StartAtLT(time.Now()))
	shifts, shiftsErr := user.QueryOncallShifts().Where(shiftIsActive).All(ctx)
	if shiftsErr != nil && !ent.IsNotFound(shiftsErr) {
		return nil, fmt.Errorf("failed to get active oncall shifts: %w", shiftsErr)
	}

	if len(shifts) == 0 {
		return &slack.ModalViewRequest{
			Type:  "modal",
			Title: plainText("No Active Shift"),
			Blocks: slack.Blocks{BlockSet: []slack.Block{
				slack.NewSectionBlock(plainText("You do not have a current oncall shift to annotate"), nil, nil),
			}},
			Close:      plainText("Cancel"),
			CallbackID: createAnnotationConfirmCallbackID,
		}, nil
	}

	view := &slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      createAnnotationConfirmCallbackID,
		PrivateMetadata: user.ID.String(),
		Title:           plainText("Create Annotation"),
		Close:           plainText("Cancel"),
		Submit:          plainText("Create"),
	}

	italicStyle := &slack.RichTextSectionTextStyle{Italic: true}
	messageDetailsBlock := slack.NewRichTextBlock("anno_msg",
		slack.NewRichTextSection(
			slack.NewRichTextSectionTextElement(ic.Message.Text, italicStyle)))

	var shiftBlock slack.Block
	if len(shifts) == 1 {
		view.PrivateMetadata = fmt.Sprintf("%s,%s", user.ID, shifts[0].ID)
		shiftBlock = plainText("Annotating your active shift in " + shifts[0].Edges.Roster.Name)
	} else {
		shiftOptions := make([]*slack.OptionBlockObject, len(shifts))
		for i, sh := range shifts {
			var desc *slack.TextBlockObject
			shiftRoster := sh.Edges.Roster
			if shiftRoster.ChatChannelID != "" {
				desc = plainText(shiftRoster.ChatChannelID)
			}
			shiftOptions[i] = slack.NewOptionBlockObject(sh.ID.String(), plainText(shiftRoster.Name), desc)
		}

		shiftSelectElement := slack.NewOptionsSelectBlockElement(
			slack.OptTypeStatic,
			plainText("Select the roster shift to annotate"),
			"select_shift",
			shiftOptions...)

		shiftBlock = slack.NewSectionBlock(plainText("Oncall Shift Rosters"), nil, slack.NewAccessory(shiftSelectElement))
	}

	notesInputBlock := slack.NewInputBlock(
		"notes_input",
		plainText("Notes"),
		plainText("You can edit this later"),
		slack.NewPlainTextInputBlockElement(nil, "notes_input_text"))

	view.Blocks = slack.Blocks{BlockSet: []slack.Block{
		messageDetailsBlock,
		slack.NewDividerBlock(),
		shiftBlock,
		slack.NewDividerBlock(),
		notesInputBlock,
	}}

	return view, nil
}

func (p *ChatProvider) handleViewSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	fmt.Printf("view submission: %+v\n", ic)
	// ic.View.PrivateMetadata
	return nil
}
