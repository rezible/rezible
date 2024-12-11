package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rezible/rezible/ent"
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
	user, userErr := p.lookupUser(ctx, ic.User.ID)
	if userErr != nil {
		return fmt.Errorf("failed to get user: %w", userErr)
	}

	rosters, rostersErr := user.QueryTeams().QueryOncallRosters().All(ctx)
	if rostersErr != nil && !ent.IsNotFound(rostersErr) {
		return fmt.Errorf("failed to get user oncall rosters: %w", rostersErr)
	}

	view := p.createAnnotationModalView(ic, rosters)

	resp, respErr := p.client.OpenViewContext(ctx, ic.TriggerID, view)
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

func (p *ChatProvider) createAnnotationModalView(ic *slack.InteractionCallback, rosters []*ent.OncallRoster) slack.ModalViewRequest {
	if len(rosters) == 0 {
		return slack.ModalViewRequest{
			Type:  "modal",
			Title: plainText("No Oncall Rosters"),
			Blocks: slack.Blocks{BlockSet: []slack.Block{
				slack.NewSectionBlock(plainText("Didn't find any oncall rosters for you"), nil, nil),
			}},
			Close:      plainText("Cancel"),
			CallbackID: createAnnotationConfirmCallbackID,
		}
	}
	italicStyle := &slack.RichTextSectionTextStyle{Italic: true}
	messageDetails := slack.NewRichTextBlock("anno_msg",
		slack.NewRichTextSection(
			slack.NewRichTextSectionTextElement(ic.Message.Text, italicStyle)))

	rosterOptions := make([]*slack.OptionBlockObject, len(rosters))
	for i, r := range rosters {
		var desc *slack.TextBlockObject
		if r.ChatChannelID != "" {
			desc = plainText(r.ChatChannelID)
		}
		rosterOptions[i] = slack.NewOptionBlockObject(r.ID.String(), plainText(r.Name), desc)
	}

	rosterSelectElement := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		plainText("Select from your rosters"),
		"select_roster",
		rosterOptions...)

	rosterSelect := slack.NewSectionBlock(plainText("Oncall Roster"), nil, slack.NewAccessory(rosterSelectElement))

	notesInput := slack.NewInputBlock(
		"notes_input",
		plainText("Notes"),
		plainText("You can edit this later"),
		slack.NewPlainTextInputBlockElement(nil, "notes_input_text"))

	return slack.ModalViewRequest{
		Type:       "modal",
		CallbackID: createAnnotationConfirmCallbackID,
		Title:      plainText("Create Oncall Annotation"),
		Close:      plainText("Cancel"),
		Submit:     plainText("Create"),
		Blocks: slack.Blocks{BlockSet: []slack.Block{
			messageDetails,
			slack.NewDividerBlock(),
			rosterSelect,
			slack.NewDividerBlock(),
			notesInput,
		}},
	}
}

func (p *ChatProvider) handleViewSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	fmt.Printf("view submission: %+v\n", ic)
	return nil
}
